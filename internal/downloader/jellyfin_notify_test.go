package downloader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/stretchr/testify/require"
)

// 4.1 processResults collects the destination folder of every successfully
// resumed item, excluding failed and unrecognized results.
func TestResumeHelper_ProcessResults_CollectsOnlySuccessfulPaths(t *testing.T) {
	rh := &ResumeHelper{}
	stats := &ResumeStats{}
	jobInfo := map[int]resumeJobInfo{
		1: {downloadID: 1, displayName: "Movie A"},
		2: {downloadID: 2, displayName: "Movie B"},
	}

	results := make(chan DownloadJobResult, 3)
	results <- DownloadJobResult{JobID: 1, Result: &DownloadResult{FilePath: "/media/movies/Movie A (2020)/Movie A (2020).mkv"}}
	results <- DownloadJobResult{JobID: 2, Error: fmt.Errorf("boom")}
	results <- DownloadJobResult{JobID: 99} // unrecognized job id
	close(results)

	changedPaths := rh.processResults(results, jobInfo, stats)

	require.Equal(t, 1, stats.Resumed)
	require.Equal(t, 2, stats.Failed, "both the errored result and the unrecognized job must count as failed")
	require.Equal(t, []string{filepath.Clean("/media/movies/Movie A (2020)")}, changedPaths)
}

func captureAppLoggerWarnings(t *testing.T) *bytes.Buffer {
	t.Helper()
	original := logger.AppLogger()
	t.Cleanup(func() { logger.SetAppLogger(original) })

	var buf bytes.Buffer
	logger.SetAppLogger(logger.New(logger.Config{
		Output:   &buf,
		MinLevel: logger.LevelWarn,
		Format:   logger.FormatText,
	}))
	return &buf
}

// 4.2 Jellyfin disabled: no network call is attempted.
func TestResumeHelper_NotifyJellyfin_DisabledSkipsNetworkCall(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer server.Close()

	rh := &ResumeHelper{}
	cfg := &config.Config{Jellyfin: config.JellyfinConfig{Enabled: false, URL: server.URL}}

	rh.notifyJellyfin(context.Background(), cfg, []string{"/media/movies/Movie (2020)"})

	require.False(t, called, "jellyfin must not be contacted when the integration is disabled")
}

// Jellyfin enabled but no base URL configured: warns, no network call.
func TestResumeHelper_NotifyJellyfin_EnabledButNoURLLogsWarningSkipsNetworkCall(t *testing.T) {
	buf := captureAppLoggerWarnings(t)

	rh := &ResumeHelper{}
	cfg := &config.Config{Jellyfin: config.JellyfinConfig{Enabled: true, URL: ""}}

	rh.notifyJellyfin(context.Background(), cfg, []string{"/media/movies/Movie (2020)"})

	require.Contains(t, buf.String(), "no url is configured", "expected a warning identifying the missing configuration")
}

// 4.2 Zero changed paths (e.g. --dry-run, or a run that resumed nothing):
// no notification is sent.
func TestResumeHelper_NotifyJellyfin_NoChangedPathsSkipsNetworkCall(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer server.Close()

	rh := &ResumeHelper{}
	cfg := &config.Config{Jellyfin: config.JellyfinConfig{Enabled: true, URL: server.URL}}

	rh.notifyJellyfin(context.Background(), cfg, nil)

	require.False(t, called, "jellyfin must not be contacted when no paths were collected")
}

// 4.2 / 5.1 A single grouped notification is sent per run, with duplicate
// paths (e.g. multiple episodes of the same season) collapsed into one, and
// a failure to notify is logged as a warning rather than returned as an
// error.
func TestResumeHelper_NotifyJellyfin_SendsSingleNotificationWithDedupedPaths(t *testing.T) {
	var requestCount int
	var gotBody struct {
		Updates []struct{ Path, UpdateType string }
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/Library/Media/Updated", r.URL.Path)
		require.Equal(t, "test-key", r.Header.Get("X-Emby-Token"))
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	rh := &ResumeHelper{}
	cfg := &config.Config{Jellyfin: config.JellyfinConfig{Enabled: true, URL: server.URL, APIKey: "test-key"}}

	seasonDir := filepath.Join("/media/tvshows", "Show (2021)", "Season 01")
	movieDir := filepath.Join("/media/movies", "Movie (2020)")
	changedPaths := []string{seasonDir, seasonDir, movieDir}

	rh.notifyJellyfin(context.Background(), cfg, changedPaths)

	require.Equal(t, 1, requestCount, "exactly one notification must be sent per run")
	require.Len(t, gotBody.Updates, 2, "the repeated season path must appear only once")

	var paths []string
	for _, u := range gotBody.Updates {
		require.Equal(t, "Created", u.UpdateType)
		paths = append(paths, u.Path)
	}
	require.Contains(t, paths, seasonDir)
	require.Contains(t, paths, movieDir)
}

// 4.3 stalkeer resume-downloads --dry-run completes no items, so
// ResumeDownloads must return (from its dry-run branch) before ever reaching
// the notify call site, even when there are incomplete downloads to plan.
func TestResumeDownloads_DryRun_NeverAttemptsJellyfinCall(t *testing.T) {
	db := setupTestDB(t)

	pending := &models.DownloadInfo{Status: string(models.DownloadStatusPending)}
	require.NoError(t, db.Create(pending).Error)

	called := false
	jellyfinServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer jellyfinServer.Close()

	stateManager := NewStateManager(DefaultStateManagerConfig())
	dl := New(5, 1, 0)
	rh := NewResumeHelper(stateManager, dl)

	config.SetConfig(&config.Config{Jellyfin: config.JellyfinConfig{Enabled: true, URL: jellyfinServer.URL}})
	t.Cleanup(func() { config.SetConfig(nil) })

	stats, err := rh.ResumeDownloads(context.Background(), ResumeOptions{DryRun: true})
	require.NoError(t, err)
	require.Equal(t, 1, stats.Total, "the pending download should have been found as an incomplete download")
	require.False(t, called, "dry-run must never contact jellyfin, since it completes no items")
}
