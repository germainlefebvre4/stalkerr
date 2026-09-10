package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/logger"
)

func TestDownloadAllSources_OneFailsOthersStillRun(t *testing.T) {
	tmpDir := t.TempDir()

	goodServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("#EXTM3U\n#EXTINF:-1,Channel\nhttp://example.com/stream\n"))
	}))
	defer goodServer.Close()

	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer badServer.Close()

	cfg := &config.Config{
		M3U: config.M3UConfig{
			Sources: []config.M3USourceConfig{
				{
					Name:     "provider-a",
					FilePath: filepath.Join(tmpDir, "provider-a.m3u"),
					Download: config.M3UDownloadConfig{
						URL:            badServer.URL,
						ArchiveDir:     filepath.Join(tmpDir, "archives"),
						RetentionCount: 5,
						MaxFileSizeMB:  10,
						TimeoutSeconds: 5,
						RetryAttempts:  1,
					},
				},
				{
					Name:     "provider-b",
					FilePath: filepath.Join(tmpDir, "provider-b.m3u"),
					Download: config.M3UDownloadConfig{
						URL:            goodServer.URL,
						ArchiveDir:     filepath.Join(tmpDir, "archives"),
						RetentionCount: 5,
						MaxFileSizeMB:  10,
						TimeoutSeconds: 5,
						RetryAttempts:  1,
					},
				},
			},
		},
	}

	log := logger.NewWithLevelAndFormat("info", "text")
	failures := downloadAllSources(cfg, log, "", true)

	if failures != 1 {
		t.Errorf("expected exactly 1 failure, got %d", failures)
	}

	// Non-implicit sources download under a per-source subdirectory (see
	// SourcePaths), so provider-a's file lives at <tmpDir>/provider-a/provider-a.m3u.
	if _, err := os.Stat(filepath.Join(tmpDir, "provider-a", "provider-a.m3u")); !os.IsNotExist(err) {
		t.Error("expected provider-a's file to not exist after a failed download")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "provider-b", "provider-b.m3u")); err != nil {
		t.Errorf("expected provider-b's file to exist after a successful download, got: %v", err)
	}
}

func TestSourceArchives_MultipleSourcesUseSeparateDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		M3U: config.M3UConfig{
			Sources: []config.M3USourceConfig{
				{
					Name:     "provider-a",
					FilePath: filepath.Join(tmpDir, "provider-a.m3u"),
					Download: config.M3UDownloadConfig{ArchiveDir: filepath.Join(tmpDir, "archives")},
				},
				{
					Name:     "provider-b",
					FilePath: filepath.Join(tmpDir, "provider-b.m3u"),
					Download: config.M3UDownloadConfig{ArchiveDir: filepath.Join(tmpDir, "archives")},
				},
			},
		},
	}

	// Seed one archive file per source's expected archive subdirectory.
	for _, name := range []string{"provider-a", "provider-b"} {
		dir := filepath.Join(tmpDir, "archives", name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create archive dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "playlist_20260101_000000.000000.m3u"), []byte("#EXTM3U\n"), 0644); err != nil {
			t.Fatalf("failed to write archive file: %v", err)
		}
	}

	log := logger.NewWithLevelAndFormat("info", "text")
	results := listAllSourceArchives(cfg, log)

	if len(results) != 2 {
		t.Fatalf("expected 2 source archive results, got %d", len(results))
	}

	seenDirs := make(map[string]bool)
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error listing archives for source %q: %v", r.SourceName, r.Err)
		}
		if len(r.Archives) != 1 {
			t.Errorf("expected 1 archive for source %q, got %d", r.SourceName, len(r.Archives))
		}
		if seenDirs[r.ArchiveDir] {
			t.Errorf("expected distinct archive directories per source, got duplicate %q", r.ArchiveDir)
		}
		seenDirs[r.ArchiveDir] = true
	}
}

func TestDownloadAllSources_LegacyImplicitSourceKeepsUnchangedPath(t *testing.T) {
	tmpDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("#EXTM3U\n#EXTINF:-1,Channel\nhttp://example.com/stream\n"))
	}))
	defer server.Close()

	destPath := filepath.Join(tmpDir, "playlist.m3u")

	cfg := &config.Config{
		M3U: config.M3UConfig{
			FilePath: destPath,
			Download: config.M3UDownloadConfig{
				URL:            server.URL,
				ArchiveDir:     filepath.Join(tmpDir, "archives"),
				RetentionCount: 5,
				MaxFileSizeMB:  10,
				TimeoutSeconds: 5,
				RetryAttempts:  1,
			},
		},
	}

	log := logger.NewWithLevelAndFormat("info", "text")
	if failures := downloadAllSources(cfg, log, "", true); failures != 0 {
		t.Fatalf("expected no failures, got %d", failures)
	}

	if _, err := os.Stat(destPath); err != nil {
		t.Errorf("expected legacy destination path unchanged and present, got: %v", err)
	}
}
