package processor

import (
	"fmt"
	"net/http"
	"time"

	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/gorm"
)

// RemoteFileSizeBackfillStats holds the results of a remote file size backfill run.
type RemoteFileSizeBackfillStats struct {
	Checked int // lines probed (remote_file_size_checked_at set on completion)
	Found   int // probes that yielded a usable size
	Errors  int // probes that failed on both HEAD and range-GET
}

// remoteFileSizeEligibleContentTypes are the content types probed for a remote
// file size; live channels have no fixed file size and are never probed.
var remoteFileSizeEligibleContentTypes = []models.ContentType{
	models.ContentTypeMovies,
	models.ContentTypeTVShows,
	models.ContentTypeUncategorized,
}

// BackfillRemoteFileSize queries ProcessedLine rows eligible for a remote file
// size probe (movies/tvshows/uncategorized, line_url set, never checked before),
// probes each one's line_url (HEAD, falling back to a ranged GET), and persists
// the result. Every probed row has remote_file_size_checked_at set regardless of
// outcome, so it is never retried by a later run. The number of rows probed in a
// single call is bounded by perRunCap. A single row's probe failure is logged and
// does not abort the run.
func BackfillRemoteFileSize(db *gorm.DB, log *logger.Logger, timeout time.Duration, perRunCap int) (*RemoteFileSizeBackfillStats, error) {
	stats := &RemoteFileSizeBackfillStats{}

	if perRunCap <= 0 {
		return stats, nil
	}

	var lines []models.ProcessedLine
	if err := db.Where("content_type IN ? AND line_url IS NOT NULL AND remote_file_size_checked_at IS NULL", remoteFileSizeEligibleContentTypes).
		Limit(perRunCap).
		Find(&lines).Error; err != nil {
		return stats, fmt.Errorf("failed to query lines for remote file size backfill: %w", err)
	}

	if len(lines) == 0 {
		return stats, nil
	}

	client := &http.Client{Timeout: timeout}

	for i := range lines {
		line := &lines[i]
		stats.Checked++

		size, err := probeRemoteFileSize(client, *line.LineURL)
		now := time.Now()

		updates := map[string]interface{}{
			"remote_file_size_checked_at": now,
		}
		if err != nil {
			stats.Errors++
			log.WithFields(map[string]interface{}{
				"line_id": line.ID,
				"url":     *line.LineURL,
				"error":   err,
			}).Warn("failed to determine remote file size")
		} else {
			stats.Found++
			updates["remote_file_size"] = size
		}

		if err := db.Model(&models.ProcessedLine{}).Where("id = ?", line.ID).Updates(updates).Error; err != nil {
			log.WithFields(map[string]interface{}{
				"line_id": line.ID,
				"error":   err,
			}).Warn("failed to persist remote file size probe result")
		}
	}

	return stats, nil
}

// probeRemoteFileSize determines the byte size of the resource at url via a HEAD
// request, falling back to a ranged GET (bytes=0-0) when HEAD fails or does not
// return a usable Content-Length.
func probeRemoteFileSize(client *http.Client, url string) (int64, error) {
	if size, err := probeHead(client, url); err == nil {
		return size, nil
	}

	return probeRangeGet(client, url)
}

func probeHead(client *http.Client, url string) (int64, error) {
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create HEAD request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("HEAD request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("HEAD returned status %d", resp.StatusCode)
	}

	if resp.ContentLength > 0 {
		return resp.ContentLength, nil
	}

	return 0, fmt.Errorf("HEAD response did not include a usable Content-Length")
}

func probeRangeGet(client *http.Client, url string) (int64, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create range GET request: %w", err)
	}
	req.Header.Set("Range", "bytes=0-0")

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("range GET request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("range GET returned status %d", resp.StatusCode)
	}

	contentRange := resp.Header.Get("Content-Range")
	if contentRange == "" {
		return 0, fmt.Errorf("range GET response did not include a Content-Range header")
	}

	var start, end, total int64
	if _, err := fmt.Sscanf(contentRange, "bytes %d-%d/%d", &start, &end, &total); err != nil {
		return 0, fmt.Errorf("failed to parse Content-Range header %q: %w", contentRange, err)
	}
	if total <= 0 {
		return 0, fmt.Errorf("invalid total size in Content-Range header %q", contentRange)
	}

	return total, nil
}
