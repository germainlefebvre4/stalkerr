package processor

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/classifier"
	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/external/tmdb"
	"github.com/glefebvre/stalkeer/internal/models"
)

func setupTestDB(t *testing.T) {
	t.Helper()

	// Load config. Database connection settings come from whatever STALKEER_DATABASE_*
	// (or DB_*) environment variables are already set (see .github/workflows/ci.yml
	// for CI, or export them locally to match your own Postgres instance).
	if err := config.Load(); err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Initialize database
	if err := database.Initialize(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}

	// Clean up tables
	db := database.Get()
	db.Exec("TRUNCATE TABLE processed_lines, processing_logs, movies, tvshows, manual_mappings CASCADE")
}

func teardownTestDB(t *testing.T) {
	t.Helper()
	if err := database.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}
}

func createTestM3U(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.m3u")

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	return tmpFile
}

func TestNewProcessor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	tmpFile := createTestM3U(t, "#EXTM3U\n#EXTINF:-1,Test\nhttp://example.com/test.mkv")

	proc, err := NewProcessor(tmpFile, "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	if proc == nil {
		t.Fatal("processor should not be nil")
	}
	if proc.parser == nil {
		t.Error("parser should not be nil")
	}
	if proc.classifier == nil {
		t.Error("classifier should not be nil")
	}
	if proc.filter == nil {
		t.Error("filter should not be nil")
	}
}

func TestProcessBasic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	content := `#EXTM3U
#EXTINF:-1 tvg-name="Test Movie" group-title="Movies",Test Movie
http://example.com/movie.mkv
#EXTINF:-1 tvg-name="Another Movie" group-title="Movies",Another Movie
http://example.com/movie2.mp4`

	tmpFile := createTestM3U(t, content)

	proc, err := NewProcessor(tmpFile, "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	opts := ProcessOptions{
		Force:            false,
		Limit:            0,
		BatchSize:        10,
		ProgressInterval: 100,
	}

	stats, err := proc.Process(opts)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if stats == nil {
		t.Fatal("stats should not be nil")
	}

	// Verify stats (may be filtered depending on config)
	if stats.TotalLines <= 0 {
		t.Errorf("expected TotalLines > 0, got %d", stats.TotalLines)
	}
}

func TestProcessWithLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	content := `#EXTM3U
#EXTINF:-1 tvg-name="Movie 1" group-title="Movies",Movie 1
http://example.com/1.mkv
#EXTINF:-1 tvg-name="Movie 2" group-title="Movies",Movie 2
http://example.com/2.mkv
#EXTINF:-1 tvg-name="Movie 3" group-title="Movies",Movie 3
http://example.com/3.mkv`

	tmpFile := createTestM3U(t, content)

	proc, err := NewProcessor(tmpFile, "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	opts := ProcessOptions{
		Force:            false,
		Limit:            2,
		BatchSize:        10,
		ProgressInterval: 100,
	}

	stats, err := proc.Process(opts)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Processed count should not exceed limit
	if stats.Processed > opts.Limit {
		t.Errorf("expected Processed <= %d, got %d", opts.Limit, stats.Processed)
	}
}

func TestProcessDuplicates(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	content := `#EXTM3U
#EXTINF:-1 tvg-name="Test Movie" group-title="Movies",Test Movie
http://example.com/movie.mkv`

	tmpFile := createTestM3U(t, content)

	proc, err := NewProcessor(tmpFile, "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	opts := ProcessOptions{
		Force:            false,
		Limit:            0,
		BatchSize:        10,
		ProgressInterval: 100,
	}

	// First processing
	stats1, err := proc.Process(opts)
	if err != nil {
		t.Fatalf("First Process failed: %v", err)
	}

	// Second processing with fresh processor instance (should detect duplicate in DB)
	proc2, err := NewProcessor(tmpFile, "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	stats2, err := proc2.Process(opts)
	if err != nil {
		t.Fatalf("Second Process failed: %v", err)
	}

	// Second run should have duplicates (if not filtered)
	if stats1.Processed > 0 && stats2.DuplicatesFound == 0 && stats2.FilteredOut == 0 {
		t.Error("expected duplicates to be detected in second run")
	}
}

func TestProcessWithForce(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	content := `#EXTM3U
#EXTINF:-1 tvg-name="Test Movie" group-title="Movies",Test Movie
http://example.com/movie.mkv`

	tmpFile := createTestM3U(t, content)

	proc, err := NewProcessor(tmpFile, "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	opts := ProcessOptions{
		Force:            false, // set to false first to process normally
		Limit:            0,
		BatchSize:        10,
		ProgressInterval: 100,
	}

	// First processing
	stats1, err := proc.Process(opts)
	if err != nil {
		t.Fatalf("First Process failed: %v", err)
	}

	// Second processing with force on a fresh processor instance (should process again)
	proc2, err := NewProcessor(tmpFile, "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	optsForce := opts
	optsForce.Force = true
	stats2, err := proc2.Process(optsForce)
	if err != nil {
		t.Fatalf("Second Process failed: %v", err)
	}

	// With force, duplicates should not be detected
	if stats2.DuplicatesFound > 0 {
		t.Errorf("expected no duplicates with force flag, got %d", stats2.DuplicatesFound)
	}

	// Both runs should have same processed count (if not filtered)
	if stats1.Processed > 0 && stats2.Processed != stats1.Processed && stats2.FilteredOut == 0 {
		t.Errorf("expected same processed count, got %d and %d", stats1.Processed, stats2.Processed)
	}
}

func TestProcessingLogCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	content := `#EXTM3U
#EXTINF:-1 tvg-name="Test Movie" group-title="Movies",Test Movie
http://example.com/movie.mkv`

	tmpFile := createTestM3U(t, content)

	proc, err := NewProcessor(tmpFile, "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	opts := ProcessOptions{
		Force:            false,
		Limit:            0,
		BatchSize:        10,
		ProgressInterval: 100,
	}

	_, err = proc.Process(opts)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Check processing log was created
	db := database.Get()
	var count int64
	db.Model(&models.ProcessingLog{}).Where("action = ?", "process_m3u").Count(&count)
	if count == 0 {
		t.Error("expected processing log to be created")
	}

	// Check log has completed status
	var log models.ProcessingLog
	db.Where("action = ?", "process_m3u").Order("created_at DESC").First(&log)
	if log.Status != "success" && log.Status != "completed_with_errors" {
		t.Errorf("expected status 'success' or 'completed_with_errors', got '%s'", log.Status)
	}
	if log.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
}

func TestProcessAttributesLineToRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	content := `#EXTM3U
#EXTINF:-1 tvg-name="Test Movie" group-title="Movies",Test Movie
http://example.com/movie.mkv`

	tmpFile := createTestM3U(t, content)

	proc, err := NewProcessor(tmpFile, "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	opts := ProcessOptions{
		Force:            false,
		Limit:            0,
		BatchSize:        10,
		ProgressInterval: 100,
	}

	if _, err := proc.Process(opts); err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	db := database.Get()

	var log models.ProcessingLog
	if err := db.Where("action = ?", "process_m3u").Order("created_at DESC").First(&log).Error; err != nil {
		t.Fatalf("failed to load processing log: %v", err)
	}

	var line models.ProcessedLine
	if err := db.Where("tvg_name = ?", "Test Movie").First(&line).Error; err != nil {
		t.Fatalf("failed to load processed line: %v", err)
	}

	if line.ProcessingLogID == nil || *line.ProcessingLogID != log.ID {
		t.Errorf("expected ProcessingLogID %d, got %v", log.ID, line.ProcessingLogID)
	}
}

func TestProcessWithForceReattributesLineToNewerRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	content := `#EXTM3U
#EXTINF:-1 tvg-name="Test Movie" group-title="Movies",Test Movie
http://example.com/movie.mkv`

	tmpFile := createTestM3U(t, content)

	proc, err := NewProcessor(tmpFile, "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	opts := ProcessOptions{
		Force:            false,
		Limit:            0,
		BatchSize:        10,
		ProgressInterval: 100,
	}

	if _, err := proc.Process(opts); err != nil {
		t.Fatalf("First Process failed: %v", err)
	}

	db := database.Get()

	var firstLog models.ProcessingLog
	if err := db.Where("action = ?", "process_m3u").Order("created_at DESC").First(&firstLog).Error; err != nil {
		t.Fatalf("failed to load first processing log: %v", err)
	}

	proc2, err := NewProcessor(tmpFile, "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	optsForce := opts
	optsForce.Force = true
	if _, err := proc2.Process(optsForce); err != nil {
		t.Fatalf("Second Process failed: %v", err)
	}

	var secondLog models.ProcessingLog
	if err := db.Where("action = ?", "process_m3u").Order("created_at DESC").First(&secondLog).Error; err != nil {
		t.Fatalf("failed to load second processing log: %v", err)
	}

	if secondLog.ID == firstLog.ID {
		t.Fatal("expected a new processing log for the forced re-process")
	}

	var line models.ProcessedLine
	if err := db.Where("tvg_name = ?", "Test Movie").First(&line).Error; err != nil {
		t.Fatalf("failed to load processed line: %v", err)
	}

	if line.ProcessingLogID == nil || *line.ProcessingLogID != secondLog.ID {
		t.Errorf("expected ProcessingLogID %d (newer run), got %v", secondLog.ID, line.ProcessingLogID)
	}
}

func TestProcessSkippedDuplicateKeepsOriginalAttribution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	content := `#EXTM3U
#EXTINF:-1 tvg-name="Test Movie" group-title="Movies",Test Movie
http://example.com/movie.mkv`

	tmpFile := createTestM3U(t, content)

	proc, err := NewProcessor(tmpFile, "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	opts := ProcessOptions{
		Force:            false,
		Limit:            0,
		BatchSize:        10,
		ProgressInterval: 100,
	}

	if _, err := proc.Process(opts); err != nil {
		t.Fatalf("First Process failed: %v", err)
	}

	db := database.Get()

	var firstLog models.ProcessingLog
	if err := db.Where("action = ?", "process_m3u").Order("created_at DESC").First(&firstLog).Error; err != nil {
		t.Fatalf("failed to load first processing log: %v", err)
	}

	// Second, non-forced run should skip the line as a duplicate.
	proc2, err := NewProcessor(tmpFile, "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	stats2, err := proc2.Process(opts)
	if err != nil {
		t.Fatalf("Second Process failed: %v", err)
	}
	if stats2.DuplicatesFound == 0 {
		t.Fatal("expected the second run to detect a duplicate")
	}

	var line models.ProcessedLine
	if err := db.Where("tvg_name = ?", "Test Movie").First(&line).Error; err != nil {
		t.Fatalf("failed to load processed line: %v", err)
	}

	if line.ProcessingLogID == nil || *line.ProcessingLogID != firstLog.ID {
		t.Errorf("expected ProcessingLogID to remain %d, got %v", firstLog.ID, line.ProcessingLogID)
	}
}

func TestProcessNewItemsCountOnlyCountsCreatedLines(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	firstContent := `#EXTM3U
#EXTINF:-1 tvg-name="Existing Movie" group-title="Movies",Existing Movie
http://example.com/existing.mkv`

	proc, err := NewProcessor(createTestM3U(t, firstContent), "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	opts := ProcessOptions{BatchSize: 10, ProgressInterval: 100}
	if _, err := proc.Process(opts); err != nil {
		t.Fatalf("First Process failed: %v", err)
	}

	// Second run: force re-processing of the existing line (update), plus one brand-new line (create).
	secondContent := `#EXTM3U
#EXTINF:-1 tvg-name="Existing Movie" group-title="Movies",Existing Movie
http://example.com/existing.mkv
#EXTINF:-1 tvg-name="New Movie" group-title="Movies",New Movie
http://example.com/new.mkv`

	proc2, err := NewProcessor(createTestM3U(t, secondContent), "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	optsForce := opts
	optsForce.Force = true
	stats, err := proc2.Process(optsForce)
	if err != nil {
		t.Fatalf("Second Process failed: %v", err)
	}

	if stats.Processed != 2 {
		t.Fatalf("expected 2 processed items, got %d", stats.Processed)
	}
	if stats.NewItems != 1 {
		t.Errorf("expected NewItems to count only the newly-created line, got %d", stats.NewItems)
	}
}

func TestUpdateProcessingLogPersistsRunStatistics(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	content := `#EXTM3U
#EXTINF:-1 tvg-name="Movie One (2020)" group-title="ACTION-FR",Movie One (2020)
http://example.com/movie1.mkv
#EXTINF:-1 tvg-name="Show One S01E01" group-title="ANIMATION",Show One S01E01
http://example.com/show1.mkv`

	proc, err := NewProcessor(createTestM3U(t, content), "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	opts := ProcessOptions{BatchSize: 10, ProgressInterval: 100, SkipTMDB: true}
	stats, err := proc.Process(opts)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if stats.Movies != 1 {
		t.Fatalf("expected 1 movie classified, got %d", stats.Movies)
	}
	if stats.TVShows != 1 {
		t.Fatalf("expected 1 TV show classified, got %d", stats.TVShows)
	}

	db := database.Get()
	var log models.ProcessingLog
	if err := db.Where("action = ?", "process_m3u").Order("created_at DESC").First(&log).Error; err != nil {
		t.Fatalf("failed to load processing log: %v", err)
	}

	if log.MoviesCount == nil || *log.MoviesCount != stats.Movies {
		t.Errorf("expected MoviesCount %d, got %v", stats.Movies, log.MoviesCount)
	}
	if log.TVShowsCount == nil || *log.TVShowsCount != stats.TVShows {
		t.Errorf("expected TVShowsCount %d, got %v", stats.TVShows, log.TVShowsCount)
	}
	if log.NewItemsCount == nil || *log.NewItemsCount != stats.NewItems {
		t.Errorf("expected NewItemsCount %d, got %v", stats.NewItems, log.NewItemsCount)
	}
	if log.TMDBMatchedCount == nil || *log.TMDBMatchedCount != stats.TMDBMatched {
		t.Errorf("expected TMDBMatchedCount %d, got %v", stats.TMDBMatched, log.TMDBMatchedCount)
	}
	if log.TMDBUnmatchedCount == nil || *log.TMDBUnmatchedCount != stats.TMDBNotFound {
		t.Errorf("expected TMDBUnmatchedCount %d, got %v", stats.TMDBNotFound, log.TMDBUnmatchedCount)
	}

	expectedTitles := []string{"ACTION-FR", "ANIMATION"}
	if len(log.GroupTitles) != len(expectedTitles) {
		t.Fatalf("expected group titles %v, got %v", expectedTitles, log.GroupTitles)
	}
	for i, title := range expectedTitles {
		if log.GroupTitles[i] != title {
			t.Errorf("expected group titles %v, got %v", expectedTitles, log.GroupTitles)
			break
		}
	}
}

// TestUpdateProcessingLogPersistsMetadataBackfillCounts verifies that a run's
// rich-metadata backfill outcome (BackfillStats.Updated/Errors, folded into
// Statistics.MetadataBackfilled/MetadataBackfillErrors by Process) is
// persisted onto that run's processing_logs entry as non-null counts.
func TestUpdateProcessingLogPersistsMetadataBackfillCounts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	content := `#EXTM3U
#EXTINF:-1 tvg-name="Movie One" group-title="ACTION-FR",Movie One
http://example.com/movie1.mkv`

	proc, err := NewProcessor(createTestM3U(t, content), "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	logEntry := &models.ProcessingLog{
		Action:    "process_m3u",
		Status:    "in_progress",
		StartedAt: time.Now(),
	}
	if err := proc.db.Create(logEntry).Error; err != nil {
		t.Fatalf("failed to create processing log: %v", err)
	}

	stats := &Statistics{
		Processed:              1,
		Movies:                 1,
		MetadataBackfilled:     5,
		MetadataBackfillErrors: 1,
		GroupTitles:            map[string]struct{}{"ACTION-FR": {}},
	}

	proc.updateProcessingLog(logEntry, "success", stats, "")

	db := database.Get()
	var log models.ProcessingLog
	if err := db.First(&log, logEntry.ID).Error; err != nil {
		t.Fatalf("failed to reload processing log: %v", err)
	}

	if log.MetadataBackfilledCount == nil || *log.MetadataBackfilledCount != 5 {
		t.Errorf("expected MetadataBackfilledCount 5, got %v", log.MetadataBackfilledCount)
	}
	if log.MetadataBackfillErrorsCount == nil || *log.MetadataBackfillErrorsCount != 1 {
		t.Errorf("expected MetadataBackfillErrorsCount 1, got %v", log.MetadataBackfillErrorsCount)
	}
}

func TestProcessingLogPersistsPartialStatisticsOnFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	content := `#EXTM3U
#EXTINF:-1 tvg-name="Movie One" group-title="ACTION-FR",Movie One
http://example.com/movie1.mkv`

	proc, err := NewProcessor(createTestM3U(t, content), "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	// Simulate a run that accumulated statistics from items processed before a
	// fatal error, then got marked failed - mirroring the call `Process` makes
	// when `parser.Parse` errors out, but with non-zero accumulated statistics
	// to verify updateProcessingLog persists them rather than dropping them.
	logEntry := &models.ProcessingLog{
		Action:    "process_m3u",
		Status:    "in_progress",
		StartedAt: time.Now(),
	}
	if err := proc.db.Create(logEntry).Error; err != nil {
		t.Fatalf("failed to create processing log: %v", err)
	}

	stats := &Statistics{
		Processed:   6,
		Movies:      4,
		TVShows:     2,
		NewItems:    6,
		TMDBMatched: 5,
		GroupTitles: map[string]struct{}{"ACTION-FR": {}},
	}

	proc.updateProcessingLog(logEntry, "failed", stats, "fatal parse error")

	db := database.Get()
	var log models.ProcessingLog
	if err := db.First(&log, logEntry.ID).Error; err != nil {
		t.Fatalf("failed to reload processing log: %v", err)
	}

	if log.Status != "failed" {
		t.Errorf("expected status 'failed', got %q", log.Status)
	}
	if log.MoviesCount == nil || *log.MoviesCount != 4 {
		t.Errorf("expected MoviesCount 4, got %v", log.MoviesCount)
	}
	if log.TVShowsCount == nil || *log.TVShowsCount != 2 {
		t.Errorf("expected TVShowsCount 2, got %v", log.TVShowsCount)
	}
	if log.NewItemsCount == nil || *log.NewItemsCount != 6 {
		t.Errorf("expected NewItemsCount 6, got %v", log.NewItemsCount)
	}
	if len(log.GroupTitles) != 1 || log.GroupTitles[0] != "ACTION-FR" {
		t.Errorf("expected GroupTitles [ACTION-FR], got %v", log.GroupTitles)
	}
}

func TestProcessNoItemsProcessedRecordsEmptyGroupTitlesAndZeroCounts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	content := `#EXTM3U
#EXTINF:-1 tvg-name="Movie One" group-title="ACTION-FR",Movie One
http://example.com/movie1.mkv`

	proc, err := NewProcessor(createTestM3U(t, content), "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	opts := ProcessOptions{BatchSize: 10, ProgressInterval: 100}
	if _, err := proc.Process(opts); err != nil {
		t.Fatalf("First Process failed: %v", err)
	}

	// Second, non-forced run over the same content: every line is a duplicate and skipped.
	proc2, err := NewProcessor(createTestM3U(t, content), "default")
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	if _, err := proc2.Process(opts); err != nil {
		t.Fatalf("Second Process failed: %v", err)
	}

	db := database.Get()
	var log models.ProcessingLog
	if err := db.Where("action = ?", "process_m3u").Order("created_at DESC").First(&log).Error; err != nil {
		t.Fatalf("failed to load second processing log: %v", err)
	}

	if log.MoviesCount == nil || *log.MoviesCount != 0 {
		t.Errorf("expected MoviesCount 0, got %v", log.MoviesCount)
	}
	if log.NewItemsCount == nil || *log.NewItemsCount != 0 {
		t.Errorf("expected NewItemsCount 0, got %v", log.NewItemsCount)
	}
	if log.GroupTitles == nil {
		t.Error("expected GroupTitles to be an empty list, got nil")
	}
	if len(log.GroupTitles) != 0 {
		t.Errorf("expected GroupTitles to be empty, got %v", log.GroupTitles)
	}
}

func TestExtractTitleAndYear(t *testing.T) {
	p := &Processor{}

	tests := []struct {
		name      string
		input     string
		wantTitle string
		wantYear  *int
	}{
		{
			name:      "trailing SD suffix stripped",
			input:     "Wonder Woman SD",
			wantTitle: "Wonder Woman",
			wantYear:  nil,
		},
		{
			name:      "trailing SD with accented characters",
			input:     "Jumanji : Bienvenue dans la jungle SD",
			wantTitle: "Jumanji : Bienvenue dans la jungle",
			wantYear:  nil,
		},
		{
			name:      "FHD MULTI suffix stripped with year in parentheses",
			input:     "Die Hart 2 (2024) FHD MULTI",
			wantTitle: "Die Hart 2",
			wantYear:  intPtr(2024),
		},
		{
			name:      "HD MULTI suffix stripped with year in parentheses",
			input:     "Heist 88 (2024) HD MULTI",
			wantTitle: "Heist 88",
			wantYear:  intPtr(2024),
		},
		{
			name:      "year in parentheses without suffix",
			input:     "Inception (2010)",
			wantTitle: "Inception",
			wantYear:  intPtr(2010),
		},
		{
			name:      "dash year format",
			input:     "Super Dark Times - 2017",
			wantTitle: "Super Dark Times",
			wantYear:  intPtr(2017),
		},
		{
			name:      "dash year format with accents",
			input:     "Une Couronne pour Noël - 2015",
			wantTitle: "Une Couronne pour Noël",
			wantYear:  intPtr(2015),
		},
		{
			name:      "hyphen in title without year is preserved",
			input:     "Spider-Man : No Way Home",
			wantTitle: "Spider-Man : No Way Home",
			wantYear:  nil,
		},
		{
			name:      "dash in title not followed by valid year is preserved",
			input:     "Mission : Impossible - Fallout",
			wantTitle: "Mission : Impossible - Fallout",
			wantYear:  nil,
		},
		{
			name:      "plain title without suffix or year",
			input:     "Venom",
			wantTitle: "Venom",
			wantYear:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTitle, gotYear := p.extractTitleAndYear(tt.input)
			if gotTitle != tt.wantTitle {
				t.Errorf("title: got %q, want %q", gotTitle, tt.wantTitle)
			}
			if tt.wantYear == nil && gotYear != nil {
				t.Errorf("year: got %d, want nil", *gotYear)
			} else if tt.wantYear != nil && gotYear == nil {
				t.Errorf("year: got nil, want %d", *tt.wantYear)
			} else if tt.wantYear != nil && gotYear != nil && *gotYear != *tt.wantYear {
				t.Errorf("year: got %d, want %d", *gotYear, *tt.wantYear)
			}
		})
	}
}

func intPtr(i int) *int { return &i }

func TestSetContentTypeResolution(t *testing.T) {
	// Unit test: verifies that setContentType persists the resolution from the classifier.
	// Uses SkipTMDB=true and TMDBLanguage set to avoid config/DB dependencies.
	p := &Processor{
		classifier: classifier.New(),
	}

	res := "1080p"
	cl := classifier.Classification{
		ContentType: classifier.ContentTypeMovie,
		Resolution:  &res,
	}

	line := &models.ProcessedLine{TvgName: "Inception 1080p"}
	opts := &ProcessOptions{SkipTMDB: true, TMDBLanguage: "en-US"}
	stats := &Statistics{}

	if err := p.setContentType(line, cl, opts, stats); err != nil {
		t.Fatalf("setContentType returned error: %v", err)
	}

	if line.Resolution == nil {
		t.Fatal("expected Resolution to be set, got nil")
	}
	if *line.Resolution != "1080p" {
		t.Errorf("expected Resolution = '1080p', got '%s'", *line.Resolution)
	}
}

func TestSetContentTypeResolutionNil(t *testing.T) {
	// Verifies that nil resolution from classifier results in nil on ProcessedLine.
	p := &Processor{
		classifier: classifier.New(),
	}

	cl := classifier.Classification{
		ContentType: classifier.ContentTypeMovie,
		Resolution:  nil,
	}

	line := &models.ProcessedLine{TvgName: "Inception"}
	opts := &ProcessOptions{SkipTMDB: true, TMDBLanguage: "en-US"}
	stats := &Statistics{}

	if err := p.setContentType(line, cl, opts, stats); err != nil {
		t.Fatalf("setContentType returned error: %v", err)
	}

	if line.Resolution != nil {
		t.Errorf("expected Resolution to be nil, got '%s'", *line.Resolution)
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name       string
		tvgName    string
		groupTitle string
		want       *string
	}{
		{
			name:    "VF entry",
			tvgName: "Die Hart 2 (2024) FHD VF",
			want:    strPtr("VF"),
		},
		{
			name:    "MULTI entry",
			tvgName: "Heist 88 (2024) HD MULTI",
			want:    strPtr("MULTI"),
		},
		{
			name:    "VOSTFR entry",
			tvgName: "Clifford (FHD VOSTFR)",
			want:    strPtr("VOSTFR"),
		},
		{
			name:    "no language marker",
			tvgName: "Venom (2018) HD",
			want:    nil,
		},
		{
			name:       "language marker only in group title",
			tvgName:    "Show S01E01",
			groupTitle: "Séries VF",
			want:       strPtr("VF"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectLanguage(tt.tvgName, tt.groupTitle)
			if tt.want == nil && got != nil {
				t.Errorf("got %q, want nil", *got)
			} else if tt.want != nil && got == nil {
				t.Errorf("got nil, want %q", *tt.want)
			} else if tt.want != nil && got != nil && *got != *tt.want {
				t.Errorf("got %q, want %q", *got, *tt.want)
			}
		})
	}
}

func TestSetContentTypeLanguage(t *testing.T) {
	p := &Processor{
		classifier: classifier.New(),
	}

	cl := classifier.Classification{ContentType: classifier.ContentTypeMovie}
	opts := &ProcessOptions{SkipTMDB: true, TMDBLanguage: "en-US"}
	stats := &Statistics{}

	vfLine := &models.ProcessedLine{TvgName: "Die Hart 2 (2024) FHD VF"}
	if err := p.setContentType(vfLine, cl, opts, stats); err != nil {
		t.Fatalf("setContentType returned error: %v", err)
	}
	if vfLine.Language == nil || *vfLine.Language != "VF" {
		t.Errorf("expected Language = 'VF', got %v", vfLine.Language)
	}

	unmarkedLine := &models.ProcessedLine{TvgName: "Venom (2018) HD"}
	if err := p.setContentType(unmarkedLine, cl, opts, stats); err != nil {
		t.Fatalf("setContentType returned error: %v", err)
	}
	if unmarkedLine.Language != nil {
		t.Errorf("expected Language to be nil, got %q", *unmarkedLine.Language)
	}
}

func TestDetectFrenchVariant(t *testing.T) {
	tests := []struct {
		name       string
		tvgName    string
		groupTitle string
		want       *string
	}{
		{
			name:    "VFQ alone",
			tvgName: "Movie (2024) 720P VFQ",
			want:    strPtr("VFQ"),
		},
		{
			name:    "MULTI and VFQ together",
			tvgName: "Movie (2024) Multi.Vfq.720P",
			want:    strPtr("VFQ"),
		},
		{
			name:    "VF and VFQ together",
			tvgName: "Movie (2024) VF VFQ",
			want:    strPtr("VFQ"),
		},
		{
			name:    "no VFQ marker",
			tvgName: "Movie (2024) FHD VF",
			want:    nil,
		},
		{
			name:       "VFQ marker only in group title",
			tvgName:    "Show S01E01",
			groupTitle: "Séries VFQ",
			want:       strPtr("VFQ"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectFrenchVariant(tt.tvgName, tt.groupTitle)
			if tt.want == nil && got != nil {
				t.Errorf("got %q, want nil", *got)
			} else if tt.want != nil && got == nil {
				t.Errorf("got nil, want %q", *tt.want)
			} else if tt.want != nil && got != nil && *got != *tt.want {
				t.Errorf("got %q, want %q", *got, *tt.want)
			}
		})
	}
}

func TestSetContentTypeFrenchVariant(t *testing.T) {
	p := &Processor{
		classifier: classifier.New(),
	}

	cl := classifier.Classification{ContentType: classifier.ContentTypeMovie}
	opts := &ProcessOptions{SkipTMDB: true, TMDBLanguage: "en-US"}
	stats := &Statistics{}

	multiVfqLine := &models.ProcessedLine{TvgName: "Movie (2024) Multi.Vfq.720P"}
	if err := p.setContentType(multiVfqLine, cl, opts, stats); err != nil {
		t.Fatalf("setContentType returned error: %v", err)
	}
	if multiVfqLine.Language == nil || *multiVfqLine.Language != "MULTI" {
		t.Errorf("expected Language = 'MULTI', got %v", multiVfqLine.Language)
	}
	if multiVfqLine.FrenchVariant == nil || *multiVfqLine.FrenchVariant != "VFQ" {
		t.Errorf("expected FrenchVariant = 'VFQ', got %v", multiVfqLine.FrenchVariant)
	}

	vfVfqLine := &models.ProcessedLine{TvgName: "Movie (2024) VF VFQ"}
	if err := p.setContentType(vfVfqLine, cl, opts, stats); err != nil {
		t.Fatalf("setContentType returned error: %v", err)
	}
	if vfVfqLine.Language == nil || *vfVfqLine.Language != "VF" {
		t.Errorf("expected Language = 'VF', got %v", vfVfqLine.Language)
	}
	if vfVfqLine.FrenchVariant == nil || *vfVfqLine.FrenchVariant != "VFQ" {
		t.Errorf("expected FrenchVariant = 'VFQ', got %v", vfVfqLine.FrenchVariant)
	}

	noVfqLine := &models.ProcessedLine{TvgName: "Movie (2024) FHD VF"}
	if err := p.setContentType(noVfqLine, cl, opts, stats); err != nil {
		t.Fatalf("setContentType returned error: %v", err)
	}
	if noVfqLine.FrenchVariant != nil {
		t.Errorf("expected FrenchVariant to be nil, got %q", *noVfqLine.FrenchVariant)
	}
}

func strPtr(s string) *string { return &s }

func TestComputeLineHash(t *testing.T) {
	hash1 := computeLineHash("Test Movie http://example.com/movie.mkv")
	hash2 := computeLineHash("Test Movie http://example.com/movie.mkv")
	hash3 := computeLineHash("Different Movie http://example.com/movie.mkv")

	if hash1 != hash2 {
		t.Error("same content should produce same hash")
	}

	if hash1 == hash3 {
		t.Error("different content should produce different hash")
	}

	if len(hash1) != 64 {
		t.Errorf("expected hash length 64, got %d", len(hash1))
	}
}

func TestProcessWithManualMapping(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	db := database.Get()

	// Seed ManualMapping in DB
	mapping := models.ManualMapping{
		TvgName:     "FR: INCEPTION (2010)",
		GroupTitle:  "FR: FILMS ACTION",
		ContentType: models.ContentTypeMovies,
		TMDBID:      27205,
	}
	db.Create(&mapping)

	// Mock TMDB Server
	mockTMDB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasPrefix(r.URL.Path, "/movie/27205") {
			if strings.HasSuffix(r.URL.Path, "/external_ids") {
				w.Write([]byte(`{"tvdb_id":12345}`))
			} else {
				w.Write([]byte(`{"id":27205,"title":"Inception","release_date":"2010-07-16","genres":[{"id":28,"name":"Action"}],"runtime":148}`))
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockTMDB.Close()

	// Override TMDB baseURL
	tmdb.SetBaseURL(mockTMDB.URL)
	defer tmdb.SetBaseURL("https://api.themoviedb.org/3")

	// Set configuration so TMDB is enabled
	config.SetConfig(&config.Config{
		TMDB: config.TMDBConfig{
			Enabled:           true,
			APIKey:            "mock-key",
			Language:          "en-US",
			RequestsPerSecond: 0,
		},
	})

	tmpFile := createTestM3U(t, "#EXTM3U\n#EXTINF:-1 tvg-name=\"FR: INCEPTION (2010)\" group-title=\"FR: FILMS ACTION\",FR: INCEPTION (2010)\nhttp://example.com/inception.mkv")

	p, err := NewProcessor(tmpFile, "default")
	if err != nil {
		t.Fatalf("failed to create processor: %v", err)
	}

	line := &models.ProcessedLine{
		TvgName:    "FR: INCEPTION (2010)",
		GroupTitle: "FR: FILMS ACTION",
	}

	cl := classifier.Classification{
		ContentType: classifier.ContentTypeSeries, // Note: classifier says "Series" but manual mapping says "Movies"! This tests that mapping takes precedence!
		Resolution:  nil,
	}

	stats := &Statistics{}
	opts := &ProcessOptions{
		SkipTMDB:     false,
		TMDBLanguage: "en-US",
	}

	if err := p.setContentType(line, cl, opts, stats); err != nil {
		t.Fatalf("setContentType failed: %v", err)
	}

	if line.ContentType != models.ContentTypeMovies {
		t.Errorf("expected ContentType 'movies' (from manual mapping), got '%s'", line.ContentType)
	}

	if line.MovieID == nil {
		t.Fatal("expected MovieID to be populated")
	}

	var movie models.Movie
	if err := db.First(&movie, *line.MovieID).Error; err != nil {
		t.Fatalf("failed to find associated movie in DB: %v", err)
	}

	if movie.TMDBID != 27205 || movie.TMDBTitle != "Inception" {
		t.Errorf("associated movie details are incorrect: %+v", movie)
	}
}
