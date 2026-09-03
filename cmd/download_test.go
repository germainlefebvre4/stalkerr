package main

import (
	"context"
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
	"github.com/glefebvre/stalkeer/internal/models"
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
