package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/downloader"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/scheduler"
	"github.com/stretchr/testify/require"
)

// 3.1 downloadStats.recordItem is safe under concurrent successful
// completions and collects one changed path per success. Run with -race.
func TestDownloadStats_RecordItem_ConcurrentSuccessesCollectChangedPaths(t *testing.T) {
	stats := &downloadStats{}

	const workers = 20
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			stats.recordItem(true, filepath.Join("/media/movies", "Movie", "dir"))
		}(i)
	}
	wg.Wait()

	require.Equal(t, workers, stats.Total)
	require.Equal(t, workers, stats.Downloaded)
	require.Equal(t, 0, stats.Failed)
	require.Len(t, stats.changedPaths, workers)
}

func TestDedupePaths(t *testing.T) {
	paths := []string{
		"/media/tvshows/Show (2021)/Season 01/",
		"/media/tvshows/Show (2021)/Season 01",
		"/media/movies/Movie (2020)",
	}

	deduped := dedupePaths(paths)

	require.Len(t, deduped, 2, "the two equivalent season paths must collapse into one")
	require.Contains(t, deduped, filepath.Clean("/media/tvshows/Show (2021)/Season 01"))
	require.Contains(t, deduped, filepath.Clean("/media/movies/Movie (2020)"))
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

// 3.3 Jellyfin disabled: no network call is attempted.
func TestNotifyJellyfin_DisabledSkipsNetworkCall(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer server.Close()

	cfg := &config.Config{Jellyfin: config.JellyfinConfig{Enabled: false, URL: server.URL}}

	notifyJellyfin(context.Background(), cfg, []string{"/media/movies/Movie (2020)"})

	require.False(t, called, "jellyfin must not be contacted when the integration is disabled")
}

// Jellyfin enabled but no base URL configured: warns, no network call.
func TestNotifyJellyfin_EnabledButNoURLLogsWarningSkipsNetworkCall(t *testing.T) {
	buf := captureAppLoggerWarnings(t)

	cfg := &config.Config{Jellyfin: config.JellyfinConfig{Enabled: true, URL: ""}}

	notifyJellyfin(context.Background(), cfg, []string{"/media/movies/Movie (2020)"})

	require.Contains(t, buf.String(), "no url is configured", "expected a warning identifying the missing configuration")
}

// 3.3 Zero successful downloads: no notification is sent.
func TestNotifyJellyfin_NoChangedPathsSkipsNetworkCall(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer server.Close()

	cfg := &config.Config{Jellyfin: config.JellyfinConfig{Enabled: true, URL: server.URL}}

	notifyJellyfin(context.Background(), cfg, nil)

	require.False(t, called, "jellyfin must not be contacted when no paths were collected")
}

// 3.2 / 5.1 A single grouped notification is sent per run, with duplicate
// paths (e.g. multiple episodes of the same season) collapsed into one.
func TestNotifyJellyfin_SendsSingleNotificationWithDedupedPaths(t *testing.T) {
	var requestCount int
	var gotBody struct {
		Updates []struct {
			Path       string
			UpdateType string
		}
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

	cfg := &config.Config{Jellyfin: config.JellyfinConfig{Enabled: true, URL: server.URL, APIKey: "test-key"}}

	seasonDir := filepath.Join("/media/tvshows", "Show (2021)", "Season 01")
	movieDir := filepath.Join("/media/movies", "Movie (2020)")
	changedPaths := []string{
		seasonDir, // episode 1
		seasonDir, // episode 2, same season
		movieDir,
	}

	notifyJellyfin(context.Background(), cfg, changedPaths)

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

// 5.1 End-to-end: a run downloading multiple episodes of the same season
// sends a single notification in which that season's folder appears once.
func TestRunDownloadWorkerPool_MultipleEpisodesSameSeason_SingleGroupedNotification(t *testing.T) {
	db := setupDownloadTestDB(t)
	tempDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		w.Write([]byte("fake-video-content"))
	}))
	defer server.Close()

	seasonDir := filepath.Join(tempDir, "Show (2021)", "Season 01")
	ep1Line := createTestProcessedLine(t, db, server.URL+"/ep1.mp4")
	ep2Line := createTestProcessedLine(t, db, server.URL+"/ep2.mp4")

	streams := []*scheduler.Stream{
		{
			Tier:      scheduler.Tier1,
			SourceKey: "series:1",
			SeriesID:  1,
			Season:    1,
			Items: []scheduler.Item{
				{DisplayName: "Show S01E01", BaseDestDir: filepath.Join(seasonDir, "Show S01E01"), Candidates: []models.ProcessedLine{ep1Line}, Episode: 1},
				{DisplayName: "Show S01E02", BaseDestDir: filepath.Join(seasonDir, "Show S01E02"), Candidates: []models.ProcessedLine{ep2Line}, Episode: 2},
			},
		},
	}

	sched := scheduler.NewScheduler(streams, 0)
	dl := downloader.New(5*time.Second, 1, 0)
	cfg := &config.Config{Downloads: config.DownloadsConfig{TempDir: tempDir}}

	stats := runDownloadWorkerPool(context.Background(), sched, dl, cfg, 2, false)
	require.Equal(t, 2, stats.Downloaded)

	var requestCount int
	var gotBody struct {
		Updates []struct{ Path, UpdateType string }
	}
	jellyfinServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer jellyfinServer.Close()

	cfg.Jellyfin = config.JellyfinConfig{Enabled: true, URL: jellyfinServer.URL}
	notifyJellyfin(context.Background(), cfg, stats.changedPaths)

	require.Equal(t, 1, requestCount, "exactly one notification must be sent for the run")
	require.Len(t, gotBody.Updates, 1, "the season folder must appear only once despite two episodes completing")
	require.Equal(t, seasonDir, gotBody.Updates[0].Path)
}
