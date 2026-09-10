package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/m3udownloader"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/processor"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupProcessTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.Movie{},
		&models.TVShow{},
		&models.Channel{},
		&models.Uncategorized{},
		&models.FilterConfig{},
		&models.ProcessingLog{},
		&models.DownloadInfo{},
		&models.ProcessedLine{},
		&models.ManualMapping{},
	))

	// A ":memory:" sqlite database is private per connection; cap the pool at
	// one connection so the processor's queries see the same in-memory database
	// this test seeded, instead of a fresh empty one from another connection.
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	database.SetDB(db)
	config.SetConfig(&config.Config{})
	return db
}

func writeTestM3U(t *testing.T, dir, name, content string) string {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0755))
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	return path
}

func TestProcessConfiguredSources_MissingFileSkippedPresentFileProcessed(t *testing.T) {
	db := setupProcessTestDB(t)

	tmpDir := t.TempDir()
	// Both sources share the same configured base file_path/archive_dir (as a
	// user copying the singular block per source would); the effective
	// per-source file lives under the SourcePaths-resolved subdirectory, the
	// same path m3u-download would have written to.
	baseFilePath := filepath.Join(tmpDir, "m3u", "playlist.m3u")
	archiveDir := filepath.Join(tmpDir, "archives")

	sources := []config.M3USourceConfig{
		{Name: "provider-missing", FilePath: baseFilePath, Download: config.M3UDownloadConfig{ArchiveDir: archiveDir}},
		{Name: "provider-present", FilePath: baseFilePath, Download: config.M3UDownloadConfig{ArchiveDir: archiveDir}},
	}

	presentDest, _ := m3udownloader.SourcePaths(baseFilePath, archiveDir, "provider-present", false)
	writeTestM3U(t, filepath.Dir(presentDest), filepath.Base(presentDest), "#EXTM3U\n#EXTINF:-1,Present Movie\nhttp://example.com/present.mkv")
	// provider-missing's resolved destination is intentionally never written.

	log := logger.NewWithLevelAndFormat("info", "text")
	opts := processor.ProcessOptions{SkipTMDB: true, BatchSize: 100, ProgressInterval: 1000}

	if hadError := processConfiguredSources(sources, false, opts, log); hadError {
		t.Error("expected no hard failure - a missing source file should only be skipped with a warning")
	}

	var presentCount int64
	if err := db.Model(&models.ProcessedLine{}).Where("source_name = ?", "provider-present").Count(&presentCount).Error; err != nil {
		t.Fatalf("failed to count processed lines: %v", err)
	}
	if presentCount != 1 {
		t.Errorf("expected 1 processed line for provider-present, got %d", presentCount)
	}

	var missingCount int64
	if err := db.Model(&models.ProcessedLine{}).Where("source_name = ?", "provider-missing").Count(&missingCount).Error; err != nil {
		t.Fatalf("failed to count processed lines: %v", err)
	}
	if missingCount != 0 {
		t.Errorf("expected 0 processed lines for provider-missing (its file was missing), got %d", missingCount)
	}
}

func TestRunProcessTarget_PositionalArgUsesDefaultSourceName(t *testing.T) {
	db := setupProcessTestDB(t)

	tmpDir := t.TempDir()
	explicitFile := writeTestM3U(t, tmpDir, "explicit.m3u", "#EXTM3U\n#EXTINF:-1,Explicit Movie\nhttp://example.com/explicit.mkv")

	// The command's positional-argument path (process <m3u-file>) bypasses the
	// configured source list entirely and calls runProcessTarget directly with
	// the legacy "default" source name - exercised here the same way.
	stats, err := runProcessTarget(explicitFile, "default", processor.ProcessOptions{SkipTMDB: true, BatchSize: 100, ProgressInterval: 1000})
	if err != nil {
		t.Fatalf("runProcessTarget failed: %v", err)
	}
	if stats.Processed != 1 {
		t.Errorf("expected 1 processed entry, got %d", stats.Processed)
	}

	var line models.ProcessedLine
	if err := db.Where("tvg_name = ?", "Explicit Movie").First(&line).Error; err != nil {
		t.Fatalf("failed to find processed line: %v", err)
	}
	if line.SourceName != "default" {
		t.Errorf("expected SourceName 'default', got %q", line.SourceName)
	}
}
