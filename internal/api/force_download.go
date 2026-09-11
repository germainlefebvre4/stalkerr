package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/downloader"
	"github.com/glefebvre/stalkeer/internal/external/radarr"
	"github.com/glefebvre/stalkeer/internal/external/sonarr"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/retry"
	"gorm.io/gorm"
)

// existenceCheckTimeout bounds the live Radarr/Sonarr lookup performed synchronously
// when a forced download is requested, keeping the endpoint responsive.
const existenceCheckTimeout = 10 * time.Second

// ForceDownloadResponse is returned when a forced-download request is accepted.
type ForceDownloadResponse struct {
	Status          string `json:"status"`
	ProcessedLineID uint   `json:"processed_line_id"`
}

// forceDownloadItem triggers an asynchronous, on-demand download of exactly one
// ProcessedLine occurrence, independent of its sibling occurrences and of
// Radarr/Sonarr's "missing" status. See openspec/changes/force-download-media-occurrence.
func (s *Server) forceDownloadItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "invalid item id")
		return
	}

	db := database.Get()
	var item models.ProcessedLine
	if err := db.Preload("Movie").Preload("TVShow").First(&item, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondError(c, http.StatusNotFound, "not_found", fmt.Sprintf("item with id %d not found", id))
			return
		}
		respondError(c, http.StatusInternalServerError, "database_error", "failed to fetch item")
		return
	}

	// Local eligibility checks first (cheapest), so an obviously-wrong request
	// never reaches Radarr/Sonarr.
	isMovie := item.ContentType == models.ContentTypeMovies && item.Movie != nil
	isEpisode := item.ContentType == models.ContentTypeTVShows && item.TVShow != nil && item.TVShow.Season != nil && item.TVShow.Episode != nil
	if !isMovie && !isEpisode {
		respondError(c, http.StatusUnprocessableEntity, "not_matched", "item is not matched to a movie or TV show")
		return
	}

	if item.State == models.StateDownloaded {
		respondError(c, http.StatusConflict, "already_downloaded", "item is already downloaded")
		return
	}
	if item.State == models.StateDownloading {
		respondError(c, http.StatusConflict, "already_downloading", "a download for this item is already in progress")
		return
	}

	if item.LineURL == nil || *item.LineURL == "" {
		respondError(c, http.StatusUnprocessableEntity, "no_stream_url", "item has no stream URL to download")
		return
	}

	cfg := config.Get()
	ctx, cancel := context.WithTimeout(c.Request.Context(), existenceCheckTimeout)
	defer cancel()

	fallbackMarker := fmt.Sprintf("occurrence-%d", item.ID)

	var basePath string
	if isMovie {
		path, errResp := s.resolveForceDownloadMoviePath(ctx, cfg, &item, fallbackMarker)
		if errResp != nil {
			respondError(c, errResp.status, errResp.code, errResp.message)
			return
		}
		basePath = path
	} else {
		path, errResp := s.resolveForceDownloadEpisodePath(ctx, cfg, &item, fallbackMarker)
		if errResp != nil {
			respondError(c, errResp.status, errResp.code, errResp.message)
			return
		}
		basePath = path
	}

	if err := s.persistForceDownloadPath(db, &item, basePath); err != nil {
		respondError(c, http.StatusInternalServerError, "database_error", "failed to persist download record")
		return
	}

	c.JSON(http.StatusAccepted, ForceDownloadResponse{
		Status:          "queued",
		ProcessedLineID: item.ID,
	})
}

// forceDownloadErrorResponse pairs an HTTP status with the error code/message to send for it.
type forceDownloadErrorResponse struct {
	status  int
	code    string
	message string
}

// resolveForceDownloadMoviePath runs the live Radarr existence check for a movie
// occurrence and, on success, computes its resolution-suffixed destination base path.
func (s *Server) resolveForceDownloadMoviePath(ctx context.Context, cfg *config.Config, item *models.ProcessedLine, fallbackMarker string) (string, *forceDownloadErrorResponse) {
	if cfg.Radarr.URL == "" || cfg.Radarr.APIKey == "" {
		return "", &forceDownloadErrorResponse{http.StatusServiceUnavailable, "radarr_not_configured", "Radarr is not configured"}
	}

	client := radarr.New(radarr.Config{
		BaseURL:     cfg.Radarr.URL,
		APIKey:      cfg.Radarr.APIKey,
		Timeout:     existenceCheckTimeout,
		RetryConfig: retry.Config{MaxAttempts: 1},
	})

	movie, err := client.GetMovieByTMDBID(ctx, item.Movie.TMDBID)
	if err != nil {
		return "", &forceDownloadErrorResponse{http.StatusBadGateway, "existence_check_failed", "failed to verify movie existence in Radarr"}
	}
	if movie == nil {
		return "", &forceDownloadErrorResponse{http.StatusNotFound, "media_not_found", "movie not found in Radarr"}
	}
	if !movie.Monitored {
		return "", &forceDownloadErrorResponse{http.StatusUnprocessableEntity, "not_monitored", "movie is not monitored in Radarr"}
	}

	basePath, _ := downloader.BuildRadarrDestPathWithResolution(
		movie.Path, cfg.Downloads.MoviesPath, item.Movie.TMDBTitle, item.Movie.TMDBYear,
		item.Resolution, item.Language, item.FrenchVariant, fallbackMarker,
	)
	return basePath, nil
}

// resolveForceDownloadEpisodePath runs the live Sonarr existence check for a TV
// episode occurrence and, on success, computes its resolution-suffixed destination
// base path. It fails closed when the TV show's TVDBID hasn't been backfilled yet.
func (s *Server) resolveForceDownloadEpisodePath(ctx context.Context, cfg *config.Config, item *models.ProcessedLine, fallbackMarker string) (string, *forceDownloadErrorResponse) {
	if item.TVShow.TVDBID == nil {
		return "", &forceDownloadErrorResponse{http.StatusNotFound, "media_not_found", "TV show has no known TVDB ID yet"}
	}

	if cfg.Sonarr.URL == "" || cfg.Sonarr.APIKey == "" {
		return "", &forceDownloadErrorResponse{http.StatusServiceUnavailable, "sonarr_not_configured", "Sonarr is not configured"}
	}

	client := sonarr.New(sonarr.Config{
		BaseURL:     cfg.Sonarr.URL,
		APIKey:      cfg.Sonarr.APIKey,
		Timeout:     existenceCheckTimeout,
		RetryConfig: retry.Config{MaxAttempts: 1},
	})

	series, episode, err := client.FindEpisodeByTVDBID(ctx, *item.TVShow.TVDBID, *item.TVShow.Season, *item.TVShow.Episode)
	if err != nil {
		return "", &forceDownloadErrorResponse{http.StatusBadGateway, "existence_check_failed", "failed to verify episode existence in Sonarr"}
	}
	if series == nil || episode == nil {
		return "", &forceDownloadErrorResponse{http.StatusNotFound, "media_not_found", "episode not found in Sonarr"}
	}
	if !series.Monitored {
		return "", &forceDownloadErrorResponse{http.StatusUnprocessableEntity, "not_monitored", "series is not monitored in Sonarr"}
	}

	basePath, _ := downloader.BuildSonarrDestPathWithResolution(
		series.Path, cfg.Downloads.TVShowsPath, item.TVShow.TMDBTitle, item.TVShow.TMDBYear,
		*item.TVShow.Season, *item.TVShow.Episode, item.Resolution, item.Language, item.FrenchVariant, fallbackMarker,
	)
	return basePath, nil
}

// persistForceDownloadPath creates (or reuses) the DownloadInfo record for item and
// persists baseDestPath onto it before the transfer starts, so a mid-transfer server
// restart resumes into the same resolution-suffixed path (see resume_helper.go
// buildBaseDestPath and design Decision 4).
func (s *Server) persistForceDownloadPath(db *gorm.DB, item *models.ProcessedLine, baseDestPath string) error {
	if item.DownloadInfoID != nil {
		return db.Model(&models.DownloadInfo{}).
			Where("id = ?", *item.DownloadInfoID).
			Update("download_path", baseDestPath).Error
	}

	downloadInfo := models.DownloadInfo{
		URL:          *item.LineURL,
		Status:       string(models.DownloadStatusPending),
		DownloadPath: &baseDestPath,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&downloadInfo).Error; err != nil {
			return err
		}
		downloadInfoID := downloadInfo.ID
		item.DownloadInfoID = &downloadInfoID
		return tx.Model(&models.ProcessedLine{}).
			Where("id = ?", item.ID).
			Update("download_info_id", &downloadInfoID).Error
	})
}
