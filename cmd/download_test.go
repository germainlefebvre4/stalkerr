package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/downloader"
	"github.com/glefebvre/stalkeer/internal/external/radarr"
	"github.com/glefebvre/stalkeer/internal/external/sonarr"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/retry"
	"github.com/glefebvre/stalkeer/internal/scheduler"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDownloadTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.ProcessedLine{},
		&models.Movie{},
		&models.TVShow{},
		&models.DownloadInfo{},
	))

	// A ":memory:" sqlite database is private per connection; cap the pool at
	// one connection so concurrent workers share the same in-memory database
	// instead of each seeing an empty one.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	database.SetDB(db)
	return db
}

func createTestProcessedLine(t *testing.T, db *gorm.DB, url string) models.ProcessedLine {
	line := models.ProcessedLine{
		LineContent: "line", LineHash: url, TvgName: "x", GroupTitle: "g",
		ProcessedAt: time.Now(), ContentType: models.ContentTypeMovies, State: models.StateProcessed,
		LineURL: &url, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&line).Error)
	return line
}

// 3.2 Worker-pool run loop: all items complete and summary stats are correct.
func TestRunDownloadWorkerPool_AllItemsComplete(t *testing.T) {
	db := setupDownloadTestDB(t)
	tempDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		w.Write([]byte("fake-video-content"))
	}))
	defer server.Close()

	movieLine := createTestProcessedLine(t, db, server.URL+"/movie.mp4")
	ep1Line := createTestProcessedLine(t, db, server.URL+"/ep1.mp4")
	ep2Line := createTestProcessedLine(t, db, server.URL+"/ep2.mp4")

	streams := []*scheduler.Stream{
		{
			Tier:      scheduler.Tier1,
			SourceKey: "movie:1",
			Items: []scheduler.Item{{
				DisplayName:  "Test Movie",
				BaseDestPath: filepath.Join(tempDir, "Test Movie"),
				Candidates:   []models.ProcessedLine{movieLine},
			}},
		},
		{
			Tier:      scheduler.Tier1,
			SourceKey: "series:1",
			SeriesID:  1,
			Season:    1,
			Items: []scheduler.Item{
				{DisplayName: "Show S01E01", BaseDestPath: filepath.Join(tempDir, "Show S01E01"), Candidates: []models.ProcessedLine{ep1Line}, Episode: 1},
				{DisplayName: "Show S01E02", BaseDestPath: filepath.Join(tempDir, "Show S01E02"), Candidates: []models.ProcessedLine{ep2Line}, Episode: 2},
			},
		},
	}

	sched := scheduler.NewScheduler(streams, 0)
	dl := downloader.New(5*time.Second, 1, 0)
	cfg := &config.Config{Downloads: config.DownloadsConfig{TempDir: tempDir}}

	stats := runDownloadWorkerPool(context.Background(), sched, dl, cfg, 2, false)

	require.Equal(t, 3, stats.Total)
	require.Equal(t, 3, stats.Downloaded)
	require.Equal(t, 0, stats.Failed)

	require.FileExists(t, filepath.Join(tempDir, "Test Movie.mp4"))
	require.FileExists(t, filepath.Join(tempDir, "Show S01E01.mp4"))
	require.FileExists(t, filepath.Join(tempDir, "Show S01E02.mp4"))
}

// media-download-scheduling spec, "Configurable shared concurrency limit":
// the number of concurrently claimed streams (movies and series combined)
// never exceeds the configured limit.
func TestRunDownloadWorkerPool_RespectsSharedConcurrencyLimit(t *testing.T) {
	db := setupDownloadTestDB(t)
	tempDir := t.TempDir()

	var inFlight int32
	var maxInFlight int32
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&inFlight, 1)
		mu.Lock()
		if current > maxInFlight {
			maxInFlight = current
		}
		mu.Unlock()
		time.Sleep(30 * time.Millisecond)
		w.Header().Set("Content-Type", "video/mp4")
		w.Write([]byte("fake-content"))
		atomic.AddInt32(&inFlight, -1)
	}))
	defer server.Close()

	const limit = 2
	var streams []*scheduler.Stream
	// Mix of movie streams and series-season streams, drawn from the same
	// shared pool, to assert the limit applies across both combined.
	for i := 0; i < 3; i++ {
		line := createTestProcessedLine(t, db, fmt.Sprintf("%s/movie%d.mp4", server.URL, i))
		streams = append(streams, &scheduler.Stream{
			Tier:      scheduler.Tier1,
			SourceKey: fmt.Sprintf("movie:%d", i),
			Items:     []scheduler.Item{{DisplayName: fmt.Sprintf("Movie %d", i), BaseDestPath: filepath.Join(tempDir, fmt.Sprintf("movie%d", i)), Candidates: []models.ProcessedLine{line}}},
		})
	}
	for i := 0; i < 3; i++ {
		line := createTestProcessedLine(t, db, fmt.Sprintf("%s/series%d.mp4", server.URL, i))
		streams = append(streams, &scheduler.Stream{
			Tier:      scheduler.Tier1,
			SourceKey: fmt.Sprintf("series:%d", i+100),
			SeriesID:  i + 100,
			Season:    1,
			Items:     []scheduler.Item{{DisplayName: fmt.Sprintf("Series %d", i), BaseDestPath: filepath.Join(tempDir, fmt.Sprintf("series%d", i)), Candidates: []models.ProcessedLine{line}, Episode: 1}},
		})
	}

	sched := scheduler.NewScheduler(streams, 0)
	dl := downloader.New(5*time.Second, 1, 0)
	cfg := &config.Config{Downloads: config.DownloadsConfig{TempDir: tempDir}}

	stats := runDownloadWorkerPool(context.Background(), sched, dl, cfg, limit, false)

	require.Equal(t, 6, stats.Total)
	require.Equal(t, 6, stats.Downloaded)
	require.LessOrEqual(t, int(maxInFlight), limit, "concurrent downloads across movies and series combined must not exceed the configured limit")
}

// 3.3 --dry-run and --limit flags.
func TestPrintDryRunPlan_DoesNotDownload(t *testing.T) {
	tempDir := t.TempDir()
	streams := []*scheduler.Stream{{
		Tier:      scheduler.Tier1,
		SourceKey: "movie:1",
		Items: []scheduler.Item{{
			DisplayName:  "Test Movie",
			BaseDestPath: filepath.Join(tempDir, "Test Movie"),
			Candidates:   []models.ProcessedLine{{LineURL: strPtrDL("http://example.com/x.mp4")}},
		}},
	}}

	printDryRunPlan(streams)

	entries, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	require.Empty(t, entries, "dry-run must not write any files")
}

func TestApplyLimit_CLIIntegration(t *testing.T) {
	streams := []*scheduler.Stream{
		{SourceKey: "movie:1", Items: []scheduler.Item{{}}},
		{SourceKey: "movie:2", Items: []scheduler.Item{{}}},
		{SourceKey: "movie:3", Items: []scheduler.Item{{}}},
	}

	limited := scheduler.ApplyLimit(streams, 2)
	require.Len(t, limited, 2)
}

// m3u-quality-selection spec: the unified download command attempts each
// candidate URL in quality-preference order, marks a failed candidate's
// ProcessedLine "failed" and moves to the next, and stops on first success.
func TestDownloadItem_QualityFallbackLoop(t *testing.T) {
	db := setupDownloadTestDB(t)
	tempDir := t.TempDir()

	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		if r.URL.Path == "/720p.mp4" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "video/mp4")
		w.Write([]byte("fake-content"))
	}))
	defer server.Close()

	failingLine := createTestProcessedLine(t, db, server.URL+"/720p.mp4")
	succeedingLine := createTestProcessedLine(t, db, server.URL+"/1080p.mp4")

	item := &scheduler.Item{
		DisplayName:  "Test Movie",
		BaseDestPath: filepath.Join(tempDir, "Test Movie"),
		Candidates:   []models.ProcessedLine{failingLine, succeedingLine},
	}

	dl := downloader.New(5*time.Second, 1, 0)
	cfg := &config.Config{Downloads: config.DownloadsConfig{TempDir: tempDir}}

	success := downloadItem(context.Background(), dl, cfg, item, false)

	require.True(t, success, "expected the second candidate to succeed")
	require.EqualValues(t, 2, hits, "both candidates should have been attempted, no more")

	var failedLine models.ProcessedLine
	require.NoError(t, db.First(&failedLine, failingLine.ID).Error)
	require.Equal(t, models.StateFailed, failedLine.State, "the failed candidate's ProcessedLine should be marked failed")

	require.FileExists(t, filepath.Join(tempDir, "Test Movie.mp4"))
}

func TestDownloadItem_AllCandidatesFail(t *testing.T) {
	db := setupDownloadTestDB(t)
	tempDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	line1 := createTestProcessedLine(t, db, server.URL+"/a.mp4")
	line2 := createTestProcessedLine(t, db, server.URL+"/b.mp4")

	item := &scheduler.Item{
		DisplayName:  "Test Movie",
		BaseDestPath: filepath.Join(tempDir, "Test Movie"),
		Candidates:   []models.ProcessedLine{line1, line2},
	}

	dl := downloader.New(5*time.Second, 1, 0)
	cfg := &config.Config{Downloads: config.DownloadsConfig{TempDir: tempDir}}

	success := downloadItem(context.Background(), dl, cfg, item, false)

	require.False(t, success, "expected the item to be counted as failed, not crash")
}

func TestDownloadItem_SuccessStopsLoop(t *testing.T) {
	db := setupDownloadTestDB(t)
	tempDir := t.TempDir()

	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Content-Type", "video/mp4")
		w.Write([]byte("fake-content"))
	}))
	defer server.Close()

	firstLine := createTestProcessedLine(t, db, server.URL+"/720p.mp4")
	secondLine := createTestProcessedLine(t, db, server.URL+"/1080p.mp4")

	item := &scheduler.Item{
		DisplayName:  "Test Movie",
		BaseDestPath: filepath.Join(tempDir, "Test Movie"),
		Candidates:   []models.ProcessedLine{firstLine, secondLine},
	}

	dl := downloader.New(5*time.Second, 1, 0)
	cfg := &config.Config{Downloads: config.DownloadsConfig{TempDir: tempDir}}

	success := downloadItem(context.Background(), dl, cfg, item, false)

	require.True(t, success)
	require.EqualValues(t, 1, hits, "no further candidates should be attempted once one succeeds")
}

func strPtrDL(s string) *string { return &s }

func newTestRadarrClient(baseURL string) *radarr.Client {
	return radarr.New(radarr.Config{
		BaseURL:     baseURL,
		APIKey:      "test-key",
		Timeout:     5 * time.Second,
		RetryConfig: retry.Config{MaxAttempts: 1},
	})
}

func newTestSonarrClient(baseURL string) *sonarr.Client {
	return sonarr.New(sonarr.Config{
		BaseURL:     baseURL,
		APIKey:      "test-key",
		Timeout:     5 * time.Second,
		RetryConfig: retry.Config{MaxAttempts: 1},
	})
}

// 2.1 reconcileDownloadPaths also returns a monitored-status snapshot built
// from the same already-fetched Radarr/Sonarr libraries (including an
// explicit unmonitored series, which GetAllSeries - unlike
// GetAllMonitoredSeries - must still surface).
func TestReconcileDownloadPaths_ReturnsMonitoredMaps(t *testing.T) {
	db := setupDownloadTestDB(t)

	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 1, "tmdbId": 42, "path": "/downloads/radarr/Monitored Movie", "monitored": true},
			{"id": 2, "tmdbId": 43, "path": "/downloads/radarr/Unmonitored Movie", "monitored": false},
		})
	}))
	defer radarrServer.Close()

	sonarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 1, "tvdbId": 200, "path": "/downloads/sonarr/Monitored Show", "monitored": true},
			{"id": 2, "tvdbId": 201, "path": "/downloads/sonarr/Unmonitored Show", "monitored": false},
		})
	}))
	defer sonarrServer.Close()

	movieMonitored, seriesMonitored := reconcileDownloadPaths(
		context.Background(), db,
		newTestRadarrClient(radarrServer.URL), newTestSonarrClient(sonarrServer.URL),
		false,
	)

	require.Equal(t, true, movieMonitored[42])
	require.Equal(t, false, movieMonitored[43])
	require.Equal(t, true, seriesMonitored[200])
	require.Equal(t, false, seriesMonitored[201], "an unmonitored series must be explicitly recorded false, not merely absent")
}

// 2.2 End-to-end: a movie confirmed unmonitored via the live Radarr fetch
// results in its incomplete download not appearing in the built stream set,
// while a monitored movie's incomplete download still resumes.
func TestReconcileDownloadPaths_UnmonitoredMovieExcludedFromBuiltStreams(t *testing.T) {
	db := setupDownloadTestDB(t)
	tempDir := t.TempDir()

	unmonitoredMovie := models.Movie{TMDBID: 43, TMDBTitle: "Unmonitored Movie", TMDBYear: 2019}
	require.NoError(t, db.Create(&unmonitoredMovie).Error)
	unmonitoredLine := models.ProcessedLine{
		LineContent: "unmonitored", LineHash: "hash-unmonitored", TvgName: "Unmonitored Movie",
		ContentType: models.ContentTypeMovies, MovieID: &unmonitoredMovie.ID,
		State: models.StateDownloading, ProcessedAt: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&unmonitoredLine).Error)
	unmonitoredDl := models.DownloadInfo{Status: string(models.DownloadStatusDownloading)}
	require.NoError(t, db.Create(&unmonitoredDl).Error)
	require.NoError(t, db.Model(&unmonitoredLine).Update("download_info_id", unmonitoredDl.ID).Error)

	monitoredMovie := models.Movie{TMDBID: 44, TMDBTitle: "Monitored Movie", TMDBYear: 2020}
	require.NoError(t, db.Create(&monitoredMovie).Error)
	monitoredLine := models.ProcessedLine{
		LineContent: "monitored", LineHash: "hash-monitored", TvgName: "Monitored Movie",
		ContentType: models.ContentTypeMovies, MovieID: &monitoredMovie.ID,
		State: models.StateDownloading, ProcessedAt: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	require.NoError(t, db.Create(&monitoredLine).Error)
	monitoredDl := models.DownloadInfo{Status: string(models.DownloadStatusDownloading)}
	require.NoError(t, db.Create(&monitoredDl).Error)
	require.NoError(t, db.Model(&monitoredLine).Update("download_info_id", monitoredDl.ID).Error)

	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 1, "tmdbId": 43, "path": filepath.Join(tempDir, "Unmonitored Movie"), "monitored": false},
			{"id": 2, "tmdbId": 44, "path": filepath.Join(tempDir, "Monitored Movie"), "monitored": true},
		})
	}))
	defer radarrServer.Close()

	movieMonitored, seriesMonitored := reconcileDownloadPaths(
		context.Background(), db, newTestRadarrClient(radarrServer.URL), nil, false,
	)

	streams, err := scheduler.BuildStreams(context.Background(), scheduler.BuildDeps{
		Config: &config.Config{
			Downloads: config.DownloadsConfig{MoviesPath: filepath.Join(tempDir, "movies"), MaxRetryAttempts: 5},
		},
		DB:                     db,
		StateManager:           downloader.NewStateManager(downloader.DefaultStateManagerConfig()),
		MonitoredMovieTMDBIDs:  movieMonitored,
		MonitoredSeriesTVDBIDs: seriesMonitored,
	})
	require.NoError(t, err)

	require.Len(t, streams, 1, "only the monitored movie's incomplete download should be resumed")
	require.Equal(t, fmt.Sprintf("movie:%d", monitoredMovie.ID), streams[0].SourceKey)
}

// 2.2 Scheduled reconciliation: a completed download whose stored root
// doesn't match the fake Radarr library fixture is corrected, and a download
// with no library match is left untouched.
func TestReconcileDownloadPaths_CorrectsMatchedRow_LeavesUnmatchedUntouched(t *testing.T) {
	db := setupDownloadTestDB(t)
	tempDir := t.TempDir()

	newRoot := filepath.Join(tempDir, "movies", "Dune Part Two (2021)")
	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 1, "tmdbId": 42, "path": newRoot},
		})
	}))
	defer radarrServer.Close()

	movie := models.Movie{TMDBID: 42, TMDBTitle: "Dune", TMDBYear: 2021}
	require.NoError(t, db.Create(&movie).Error)

	oldPath := filepath.Join(tempDir, "movies", "Dune (2021)", "Dune (2021).mkv")
	require.NoError(t, os.MkdirAll(filepath.Dir(oldPath), 0755))
	require.NoError(t, os.WriteFile(oldPath, []byte("content"), 0644))

	dl := models.DownloadInfo{URL: "http://example.com/dune", Status: "completed", DownloadPath: &oldPath}
	require.NoError(t, db.Create(&dl).Error)
	line := models.ProcessedLine{
		LineContent: "dune", LineHash: "hash-dune", TvgName: "Dune",
		ContentType: models.ContentTypeMovies, MovieID: &movie.ID, DownloadInfoID: &dl.ID,
		State: models.StateDownloaded, ProcessedAt: time.Now(),
	}
	require.NoError(t, db.Create(&line).Error)

	unmatchedMovie := models.Movie{TMDBID: 99, TMDBTitle: "Unmatched", TMDBYear: 2020}
	require.NoError(t, db.Create(&unmatchedMovie).Error)
	unmatchedPath := filepath.Join(tempDir, "movies", "Unmatched (2020)", "Unmatched (2020).mkv")
	require.NoError(t, os.MkdirAll(filepath.Dir(unmatchedPath), 0755))
	require.NoError(t, os.WriteFile(unmatchedPath, []byte("content"), 0644))
	unmatchedDl := models.DownloadInfo{URL: "http://example.com/unmatched", Status: "completed", DownloadPath: &unmatchedPath}
	require.NoError(t, db.Create(&unmatchedDl).Error)
	unmatchedLine := models.ProcessedLine{
		LineContent: "unmatched", LineHash: "hash-unmatched", TvgName: "Unmatched",
		ContentType: models.ContentTypeMovies, MovieID: &unmatchedMovie.ID, DownloadInfoID: &unmatchedDl.ID,
		State: models.StateDownloaded, ProcessedAt: time.Now(),
	}
	require.NoError(t, db.Create(&unmatchedLine).Error)

	reconcileDownloadPaths(context.Background(), db, newTestRadarrClient(radarrServer.URL), nil, false)

	var updated models.DownloadInfo
	require.NoError(t, db.First(&updated, dl.ID).Error)
	expectedNewPath := filepath.Join(newRoot, "Dune (2021).mkv")
	require.NotNil(t, updated.DownloadPath)
	require.Equal(t, expectedNewPath, *updated.DownloadPath)

	var untouched models.DownloadInfo
	require.NoError(t, db.First(&untouched, unmatchedDl.ID).Error)
	require.NotNil(t, untouched.DownloadPath)
	require.Equal(t, unmatchedPath, *untouched.DownloadPath)
}

// 2.3 The run tolerates a Radarr fetch failure: reconciliation is skipped
// (no panic, no row mutated) rather than failing the run's primary work.
func TestReconcileDownloadPaths_ToleratesRadarrFetchFailure(t *testing.T) {
	db := setupDownloadTestDB(t)
	tempDir := t.TempDir()

	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer radarrServer.Close()

	movie := models.Movie{TMDBID: 42, TMDBTitle: "Dune", TMDBYear: 2021}
	require.NoError(t, db.Create(&movie).Error)

	oldPath := filepath.Join(tempDir, "movies", "Dune (2021)", "Dune (2021).mkv")
	require.NoError(t, os.MkdirAll(filepath.Dir(oldPath), 0755))
	require.NoError(t, os.WriteFile(oldPath, []byte("content"), 0644))

	dl := models.DownloadInfo{URL: "http://example.com/dune", Status: "completed", DownloadPath: &oldPath}
	require.NoError(t, db.Create(&dl).Error)
	line := models.ProcessedLine{
		LineContent: "dune", LineHash: "hash-dune", TvgName: "Dune",
		ContentType: models.ContentTypeMovies, MovieID: &movie.ID, DownloadInfoID: &dl.ID,
		State: models.StateDownloaded, ProcessedAt: time.Now(),
	}
	require.NoError(t, db.Create(&line).Error)

	require.NotPanics(t, func() {
		reconcileDownloadPaths(context.Background(), db, newTestRadarrClient(radarrServer.URL), nil, false)
	})

	var unchanged models.DownloadInfo
	require.NoError(t, db.First(&unchanged, dl.ID).Error)
	require.NotNil(t, unchanged.DownloadPath)
	require.Equal(t, oldPath, *unchanged.DownloadPath)
}
