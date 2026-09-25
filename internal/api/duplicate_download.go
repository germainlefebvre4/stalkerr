package api

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/gorm"
)

const (
	// duplicateOutcomeCancelled/duplicateOutcomeUnchanged are the per-occurrence
	// outcomes reported by the resolve-duplicate endpoint (design D3).
	duplicateOutcomeCancelled = "cancelled"
	duplicateOutcomeUnchanged = "unchanged"

	// duplicateResolveErrorMessage mirrors cancelDownloadErrorMessage's shape
	// but distinguishes this trigger from a plain manual cancel.
	duplicateResolveErrorMessage = "Cancelled by duplicate-download resolve action: file no longer exists on disk"
)

// DuplicateOccurrenceOutcome reports what happened to one occurrence of a
// resolved duplicate group.
type DuplicateOccurrenceOutcome struct {
	DownloadInfoID uint   `json:"download_info_id"`
	DownloadPath   string `json:"download_path"`
	Outcome        string `json:"outcome"`
}

// ResolveDuplicateResponse is returned when a resolve-duplicate request
// succeeds: the outcome of every occurrence in the resolved group.
type ResolveDuplicateResponse struct {
	Occurrences []DuplicateOccurrenceOutcome `json:"occurrences"`
}

// resolveDuplicateDownload applies the Manual Resolve Action (capability
// duplicate-download-detection) to the duplicate group containing the given
// download_info id: for every occurrence in that group, it checks whether the
// file still exists on disk and transitions any occurrence whose file is gone
// to the cancelled terminal status, leaving occurrences whose file is present
// unchanged. This only ever runs when explicitly requested through this
// endpoint - it is never invoked by the scheduled reconciliation pass or any
// other background process.
func (s *Server) resolveDuplicateDownload(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "invalid download id")
		return
	}

	db := database.Get()
	groups, err := database.FindDuplicateDownloadGroups(db)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "database_error", "failed to evaluate duplicate download groups")
		return
	}

	group := findGroupContaining(groups, uint(id))
	if group == nil {
		respondError(c, http.StatusNotFound, "not_a_duplicate", fmt.Sprintf("download %d does not belong to any duplicate group", id))
		return
	}

	outcomes := make([]DuplicateOccurrenceOutcome, 0, len(group.Occurrences))
	for _, occ := range group.Occurrences {
		outcome := duplicateOutcomeUnchanged
		if _, statErr := os.Stat(occ.DownloadPath); os.IsNotExist(statErr) {
			if err := cancelDuplicateOccurrence(db, occ.DownloadInfoID); err != nil {
				respondError(c, http.StatusInternalServerError, "database_error", "failed to cancel duplicate occurrence")
				return
			}
			outcome = duplicateOutcomeCancelled
		}
		outcomes = append(outcomes, DuplicateOccurrenceOutcome{
			DownloadInfoID: occ.DownloadInfoID,
			DownloadPath:   occ.DownloadPath,
			Outcome:        outcome,
		})
	}

	c.JSON(http.StatusOK, ResolveDuplicateResponse{Occurrences: outcomes})
}

// findGroupContaining returns a pointer to the group in groups that contains
// an occurrence with the given download_info id, or nil if none does.
func findGroupContaining(groups []database.DuplicateGroup, downloadInfoID uint) *database.DuplicateGroup {
	for i := range groups {
		for _, occ := range groups[i].Occurrences {
			if occ.DownloadInfoID == downloadInfoID {
				return &groups[i]
			}
		}
	}
	return nil
}

// cancelDuplicateOccurrence transitions one occurrence to the cancelled
// terminal status, reusing the same transaction shape cancelDownload
// (internal/api/cancel_download.go) uses to reach that status.
func cancelDuplicateOccurrence(db *gorm.DB, downloadInfoID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.DownloadInfo{}).
			Where("id = ?", downloadInfoID).
			Updates(map[string]interface{}{
				"status":        string(models.DownloadStatusCancelled),
				"error_message": duplicateResolveErrorMessage,
			}).Error; err != nil {
			return err
		}
		return tx.Model(&models.ProcessedLine{}).
			Where("download_info_id = ?", downloadInfoID).
			Update("state", string(models.StateCancelled)).Error
	})
}
