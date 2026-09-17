package downloader

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// alwaysStopWindows returns a schedule window active on every day, all day,
// with action "stop" - used so a policy.Engine's effective policy is "stop"
// from the very first check, without depending on wall-clock minute
// granularity.
func alwaysStopWindows() []models.BandwidthScheduleWindow {
	return []models.BandwidthScheduleWindow{
		{
			DaysOfWeek: models.StringList{
				"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday",
			},
			StartTime: "00:00",
			EndTime:   "23:59",
			Action:    "stop",
		},
	}
}

// TestDownloadFileWithResume_PolicyStopAbortsWithoutWaitingForFullTransfer
// covers "Aborting an In-Progress Transfer on Stop": the abort must happen
// within a short, bounded delay independent of the file's remaining size.
// The test server would take several seconds to fully stream its body if
// drained to completion; a stopped policy must abort long before that.
func TestDownloadFileWithResume_PolicyStopAbortsWithoutWaitingForFullTransfer(t *testing.T) {
	_ = setupTestDB(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		require.True(t, ok)
		w.WriteHeader(http.StatusOK)
		chunk := make([]byte, 64*1024)
		for i := 0; i < 50; i++ { // 50 * 100ms = 5s if fully drained
			if _, err := w.Write(chunk); err != nil {
				return
			}
			flusher.Flush()
			time.Sleep(100 * time.Millisecond)
		}
	}))
	defer server.Close()

	d := New(10*time.Second, 3, 0)
	engine := policy.New(&config.Config{}, alwaysStopWindows(), nil)
	d.SetPolicyEngine(engine)

	destPath := filepath.Join(t.TempDir(), "file")

	start := time.Now()
	result, err := d.Download(context.Background(), DownloadOptions{
		URL:          server.URL,
		BaseDestPath: destPath,
	})
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.True(t, errors.Is(err, policy.ErrStoppedByPolicy), "expected error to wrap policy.ErrStoppedByPolicy, got %v", err)
	assert.Nil(t, result)
	assert.Less(t, elapsed, 2*time.Second, "expected the abort to happen well before the full 5s transfer would complete")
}

// TestDownload_PolicyStopSetsDistinctStatusWithoutConsumingRetryBudget covers
// "Policy Abort Is Not a Failure": the resulting DownloadInfo status must be
// DownloadStatusPolicyStopped, and retry_count must be left untouched.
func TestDownload_PolicyStopSetsDistinctStatusWithoutConsumingRetryBudget(t *testing.T) {
	db := setupTestDB(t)

	lineURL := "http://example.com/stream.mkv"
	processedLine := &models.ProcessedLine{
		LineURL:     &lineURL,
		LineContent: "#EXTINF:-1,Test Movie",
		LineHash:    "policytest123",
		TvgName:     "Test Movie",
		GroupTitle:  "Movies",
		ContentType: models.ContentTypeMovies,
		State:       models.StateProcessed,
	}
	require.NoError(t, db.Create(processedLine).Error)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", 1024))
		w.WriteHeader(http.StatusOK)
		w.Write(make([]byte, 1024))
	}))
	defer server.Close()

	d := New(10*time.Second, 3, 0)
	engine := policy.New(&config.Config{}, alwaysStopWindows(), nil)
	d.SetPolicyEngine(engine)

	destPath := filepath.Join(t.TempDir(), "file")
	result, err := d.Download(context.Background(), DownloadOptions{
		URL:             server.URL,
		BaseDestPath:    destPath,
		ProcessedLineID: processedLine.ID,
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, policy.ErrStoppedByPolicy))
	assert.Nil(t, result)

	var downloadInfo models.DownloadInfo
	require.NoError(t, db.Where("url = ?", server.URL).First(&downloadInfo).Error)
	assert.Equal(t, string(models.DownloadStatusPolicyStopped), downloadInfo.Status)
	assert.Equal(t, 0, downloadInfo.RetryCount, "policy-stop must not consume the retry budget")
	assert.Nil(t, downloadInfo.LockedAt, "lock must be released on policy-stop")
}

// TestDownload_NoPolicyEngineIsUnrestricted confirms that a Downloader with
// no policy engine attached (the default, as in every pre-existing test in
// this package) behaves exactly as before this feature existed.
func TestDownload_NoPolicyEngineIsUnrestricted(t *testing.T) {
	_ = setupTestDB(t)

	content := []byte("unrestricted content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	d := New(10*time.Second, 3, 0)
	assert.False(t, d.IsPolicyStopped())

	destPath := filepath.Join(t.TempDir(), "file")
	result, err := d.Download(context.Background(), DownloadOptions{
		URL:          server.URL,
		BaseDestPath: destPath,
	})

	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), result.FileSize)
}
