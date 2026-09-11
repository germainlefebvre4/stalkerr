package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/gorm"
)

// cancelDownloadErrorMessage is the error_message value applied to a
// manually cancelled occurrence. A retry-exhaustion transition (handled in
// internal/downloader/state_manager.go) leaves the last real failure message
// in place instead, so the two triggers stay distinguishable without a
// second status value.
const cancelDownloadErrorMessage = "Cancelled by user"

// CancelDownloadResponse is returned when a cancel request succeeds.
type CancelDownloadResponse struct {
	Status     string `json:"status"`
	DownloadID uint   `json:"download_id"`
}

// cancelDownload cancels exactly one download occurrence (capability
// cancel-download-occurrence): a thin, synchronous status transition with no
// external I/O, unlike forceDownloadItem. It moves the occurrence into the
// same terminal excluded status a retry-exhausted occurrence reaches,
// without affecting any sibling occurrence of the same movie/episode.
func (s *Server) cancelDownload(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid_request", "invalid download id")
		return
	}

	db := database.Get()
	var dl models.DownloadInfo
	if err := db.First(&dl, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respondError(c, http.StatusNotFound, "not_found", fmt.Sprintf("download with id %d not found", id))
			return
		}
		respondError(c, http.StatusInternalServerError, "database_error", "failed to fetch download")
		return
	}

	eligible := dl.Status == string(models.DownloadStatusPending) ||
		dl.Status == string(models.DownloadStatusFailed) ||
		dl.Status == string(models.DownloadStatusRetrying)

	if !eligible {
		errCode := "not_cancellable"
		message := fmt.Sprintf("download with status %q cannot be cancelled", dl.Status)
		switch dl.Status {
		case string(models.DownloadStatusCompleted):
			errCode = "already_completed"
			message = "download is already completed"
		case string(models.DownloadStatusDownloading):
			errCode = "downloading_in_progress"
			message = "download is actively in progress and cannot be cancelled"
		case string(models.DownloadStatusCancelled):
			errCode = "already_cancelled"
			message = "download is already cancelled"
		}
		respondError(c, http.StatusConflict, errCode, message)
		return
	}

	errMsg := cancelDownloadErrorMessage
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.DownloadInfo{}).
			Where("id = ?", dl.ID).
			Updates(map[string]interface{}{
				"status":        string(models.DownloadStatusCancelled),
				"error_message": errMsg,
			}).Error; err != nil {
			return err
		}
		return tx.Model(&models.ProcessedLine{}).
			Where("download_info_id = ?", dl.ID).
			Update("state", string(models.StateCancelled)).Error
	}); err != nil {
		respondError(c, http.StatusInternalServerError, "database_error", "failed to cancel download")
		return
	}

	c.JSON(http.StatusOK, CancelDownloadResponse{
		Status:     string(models.DownloadStatusCancelled),
		DownloadID: dl.ID,
	})
}
