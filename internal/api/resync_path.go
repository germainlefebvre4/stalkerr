package api

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/external/radarr"
	"github.com/glefebvre/stalkeer/internal/external/sonarr"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/retry"
	"gorm.io/gorm"
)

// ReconcileOutcome is the machine-readable result of a path reconciliation
// attempt, shared by the scheduled reconciliation pass and the on-demand
// resync endpoint.
type ReconcileOutcome string

const (
	OutcomeCorrected          ReconcileOutcome = "corrected"
	OutcomeAlreadyUpToDate    ReconcileOutcome = "already_up_to_date"
	OutcomeNotManaged         ReconcileOutcome = "not_managed_by_radarr_sonarr"
	OutcomeRenameTargetExists ReconcileOutcome = "rename_target_exists"
	OutcomeRenameFailed       ReconcileOutcome = "rename_failed"
	OutcomeDBUpdateFailed     ReconcileOutcome = "database_update_failed"
)

// downloadRoot splits a completed download's stored path into the root
// directory Radarr/Sonarr owns (the movie's parent folder, or the series'
// parent-of-Season-NN folder) and the sub-path below it that stalkeer's own
// naming convention already produced.
type downloadRoot struct {
	Root    string
	SubPath string
}

// extractDownloadRoot recovers downloadRoot from downloadPath using the same
// logic already used by the rename handler: filepath.Dir for a movie, and
// detectTVSeasonPath for a TV episode. ok is false when contentType is TV but
// downloadPath's parent directory doesn't follow the "Season NN" convention,
// since the root then can't be reliably recovered.
func extractDownloadRoot(downloadPath string, contentType models.ContentType) (downloadRoot, bool) {
	if contentType == models.ContentTypeTVShows {
		info, ok := detectTVSeasonPath(downloadPath)
		if !ok {
			return downloadRoot{}, false
		}
		return downloadRoot{
			Root:    info.SeriesRoot,
			SubPath: filepath.Join(filepath.Base(info.SeasonDir), filepath.Base(downloadPath)),
		}, true
	}

	return downloadRoot{
		Root:    filepath.Dir(downloadPath),
		SubPath: filepath.Base(downloadPath),
	}, true
}

// computeReconciledPath compares root against targetRoot (Radarr's/Sonarr's
// current movie.Path/series.Path) and, when they differ, returns the
// corrected full file path with root.SubPath preserved below targetRoot.
func computeReconciledPath(root downloadRoot, targetRoot string) (newPath string, changed bool) {
	if targetRoot == "" || targetRoot == root.Root {
		return "", false
	}
	return filepath.Join(targetRoot, root.SubPath), true
}

// applyPathCorrection moves the download's file from oldPath to newPath and
// updates download_info.download_path in a DB transaction, reusing the same
// move/collision-check/cleanup primitives as the manual rename endpoint
// (moveSingleFile, detectTVSeasonPath, removeDirIfEmpty).
//
// oldPath itself may no longer exist: Radarr/Sonarr commonly rename/move a
// whole series or movie folder as one filesystem operation (its own
// organizer, or a user's manual `mv`), which relocates every file beneath it
// in one shot — including ones stalkeer already recorded as completed. When
// that has happened, oldPath is gone and newPath already holds the real
// file; that is not a destination collision to block on, it is confirmation
// the correction already happened on disk, so this only needs to catch the
// DB record up (no move, nothing to clean up).
func applyPathCorrection(db *gorm.DB, downloadID uint, oldPath, newPath string) ReconcileOutcome {
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		if _, err := os.Stat(newPath); err != nil {
			// Neither the old nor the new path has the file: it's genuinely
			// missing, not just relocated. Report failure rather than
			// pointing download_path at a file that isn't there.
			return OutcomeRenameFailed
		}
		return updateDownloadPathColumn(db, downloadID, newPath)
	}

	if _, err := os.Stat(newPath); err == nil {
		return OutcomeRenameTargetExists
	} else if !os.IsNotExist(err) {
		return OutcomeRenameFailed
	}

	if err := moveSingleFile(oldPath, newPath); err != nil {
		return OutcomeRenameFailed
	}

	if outcome := updateDownloadPathColumn(db, downloadID, newPath); outcome != OutcomeCorrected {
		return outcome
	}

	if info, ok := detectTVSeasonPath(oldPath); ok {
		if removeDirIfEmpty(info.SeasonDir) {
			removeDirIfEmpty(info.SeriesRoot)
		}
	} else {
		removeDirIfEmpty(filepath.Dir(oldPath))
	}

	return OutcomeCorrected
}

// updateDownloadPathColumn persists newPath as download_info.download_path
// in a transaction, translating a failure into the outcome the rename
// endpoint has always used to signal "file moved, DB didn't catch up".
func updateDownloadPathColumn(db *gorm.DB, downloadID uint, newPath string) ReconcileOutcome {
	if err := db.Transaction(func(tx *gorm.DB) error {
		return tx.Model(&models.DownloadInfo{}).Where("id = ?", downloadID).Update("download_path", newPath).Error
	}); err != nil {
		return OutcomeDBUpdateFailed
	}
	return OutcomeCorrected
}

// reconcileDownloadPath is the shared entry point used by both the scheduled
// reconciliation pass and the on-demand resync endpoint: given a completed
// download's stored path/content type and the current Radarr/Sonarr root for
// its movie/series, it corrects download_path when it has drifted from that
// root, leaving it untouched otherwise.
func reconcileDownloadPath(db *gorm.DB, dl models.DownloadInfo, contentType models.ContentType, targetRoot string) (outcome ReconcileOutcome, oldPath, newPath string) {
	if dl.DownloadPath != nil {
		oldPath = *dl.DownloadPath
	}

	root, ok := extractDownloadRoot(oldPath, contentType)
	if !ok {
		return OutcomeRenameFailed, oldPath, ""
	}

	corrected, changed := computeReconciledPath(root, targetRoot)
	if !changed {
		return OutcomeAlreadyUpToDate, oldPath, oldPath
	}

	outcome = applyPathCorrection(db, dl.ID, oldPath, corrected)
	if outcome == OutcomeCorrected {
		return outcome, oldPath, corrected
	}
	return outcome, oldPath, ""
}

// ReconcileScheduledDownloadPaths corrects the stored download_path of every
// completed download whose movie/series is present in the given Radarr/Sonarr
// library snapshot (movieTMDBPaths keyed by Movie.TMDBID, seriesTVDBPaths
// keyed by TVShow.TVDBID). A download whose movie/series has no entry in
// either map is left entirely untouched (population-2 downloads, per
// proposal.md). This function makes no Radarr/Sonarr API calls itself:
// callers pass in library data already fetched for the run.
func ReconcileScheduledDownloadPaths(db *gorm.DB, movieTMDBPaths map[int]string, seriesTVDBPaths map[int]string) {
	var downloads []models.DownloadInfo
	if err := db.Where("status = ?", string(models.DownloadStatusCompleted)).
		Preload("ProcessedLines.Movie").
		Preload("ProcessedLines.TVShow").
		Find(&downloads).Error; err != nil {
		logger.AppLogger().WithFields(map[string]interface{}{
			"error": err,
		}).Warn("path reconciliation: failed to load completed downloads, skipping this run")
		return
	}

	for _, dl := range downloads {
		if dl.DownloadPath == nil || *dl.DownloadPath == "" || len(dl.ProcessedLines) == 0 {
			continue
		}
		line := dl.ProcessedLines[0]

		var targetRoot string
		var contentType models.ContentType
		switch {
		case line.Movie != nil:
			path, ok := movieTMDBPaths[line.Movie.TMDBID]
			if !ok {
				continue
			}
			targetRoot, contentType = path, models.ContentTypeMovies
		case line.TVShow != nil && line.TVShow.TVDBID != nil:
			path, ok := seriesTVDBPaths[*line.TVShow.TVDBID]
			if !ok {
				continue
			}
			targetRoot, contentType = path, models.ContentTypeTVShows
		default:
			continue
		}

		outcome, oldPath, newPath := reconcileDownloadPath(db, dl, contentType, targetRoot)
		switch outcome {
		case OutcomeCorrected:
			logger.AppLogger().WithFields(map[string]interface{}{
				"download_id": dl.ID,
				"old_path":    oldPath,
				"new_path":    newPath,
			}).Info("path reconciliation: corrected stale download_path")
		case OutcomeAlreadyUpToDate:
			// Nothing to do.
		default:
			logger.AppLogger().WithFields(map[string]interface{}{
				"download_id": dl.ID,
				"outcome":     outcome,
			}).Warn("path reconciliation: could not correct download_path")
		}
	}
}

// resyncDownloadPath performs an on-demand correction of exactly one
// completed download's stored path against a live Radarr/Sonarr lookup.
func (s *Server) resyncDownloadPath(c *gin.Context) {
	db := database.Get()
	id := c.Param("id")

	var dl models.DownloadInfo
	if err := db.Preload("ProcessedLines.Movie").Preload("ProcessedLines.TVShow").First(&dl, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "Download not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch download",
		})
		return
	}

	if dl.DownloadPath == nil || *dl.DownloadPath == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "Download has no completed file to resync",
		})
		return
	}

	if len(dl.ProcessedLines) == 0 {
		c.JSON(http.StatusOK, gin.H{"status": OutcomeNotManaged})
		return
	}
	line := dl.ProcessedLines[0]

	cfg := config.Get()
	ctx, cancel := context.WithTimeout(c.Request.Context(), existenceCheckTimeout)
	defer cancel()

	var targetRoot string
	var contentType models.ContentType
	found := false

	switch {
	case line.Movie != nil:
		contentType = models.ContentTypeMovies
		if cfg.Radarr.URL == "" || cfg.Radarr.APIKey == "" {
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{
				Error:   "radarr_not_configured",
				Message: "Radarr is not configured",
			})
			return
		}

		client := radarr.New(radarr.Config{
			BaseURL:     cfg.Radarr.URL,
			APIKey:      cfg.Radarr.APIKey,
			Timeout:     existenceCheckTimeout,
			RetryConfig: retry.Config{MaxAttempts: 1},
		})

		movie, err := client.GetMovieByTMDBID(ctx, line.Movie.TMDBID)
		if err != nil {
			c.JSON(http.StatusBadGateway, ErrorResponse{
				Error:   "existence_check_failed",
				Message: "failed to verify movie existence in Radarr",
			})
			return
		}
		if movie != nil {
			targetRoot = movie.Path
			found = true
		}

	case line.TVShow != nil && line.TVShow.TVDBID != nil:
		contentType = models.ContentTypeTVShows
		if cfg.Sonarr.URL == "" || cfg.Sonarr.APIKey == "" {
			c.JSON(http.StatusServiceUnavailable, ErrorResponse{
				Error:   "sonarr_not_configured",
				Message: "Sonarr is not configured",
			})
			return
		}

		client := sonarr.New(sonarr.Config{
			BaseURL:     cfg.Sonarr.URL,
			APIKey:      cfg.Sonarr.APIKey,
			Timeout:     existenceCheckTimeout,
			RetryConfig: retry.Config{MaxAttempts: 1},
		})

		series, err := client.GetSeriesByTVDBID(ctx, *line.TVShow.TVDBID)
		if err != nil {
			c.JSON(http.StatusBadGateway, ErrorResponse{
				Error:   "existence_check_failed",
				Message: "failed to verify series existence in Sonarr",
			})
			return
		}
		if series != nil {
			targetRoot = series.Path
			found = true
		}
	}

	if !found {
		c.JSON(http.StatusOK, gin.H{"status": OutcomeNotManaged})
		return
	}

	outcome, oldPath, newPath := reconcileDownloadPath(db, dl, contentType, targetRoot)

	switch outcome {
	case OutcomeCorrected:
		c.JSON(http.StatusOK, gin.H{
			"status":   outcome,
			"old_path": oldPath,
			"new_path": newPath,
		})
	case OutcomeAlreadyUpToDate:
		c.JSON(http.StatusOK, gin.H{"status": outcome})
	case OutcomeRenameTargetExists:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   string(outcome),
			Message: "A file already exists at the destination path",
		})
	case OutcomeDBUpdateFailed:
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   string(outcome),
			Message: "File moved successfully on disk, but database path update failed",
		})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   string(OutcomeRenameFailed),
			Message: "Failed to move physical file",
		})
	}
}
