package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.ProcessedLine{},
		&models.Movie{},
		&models.TVShow{},
		&models.DownloadInfo{},
	)
	require.NoError(t, err)

	// Set global database instance
	database.SetDB(db)
	return db
}

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		timeout       time.Duration
		retryAttempts int
		wantTimeout   time.Duration
		wantRetries   int
	}{
		{
			name:          "with custom values",
			timeout:       60 * time.Second,
			retryAttempts: 5,
			wantTimeout:   60 * time.Second,
			wantRetries:   5,
		},
		{
			name:          "with zero values uses defaults",
			timeout:       0,
			retryAttempts: 0,
			wantTimeout:   600 * time.Second,
			wantRetries:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := New(tt.timeout, tt.retryAttempts)
			assert.NotNil(t, d)
			assert.Equal(t, tt.wantTimeout, d.httpClient.Timeout)
			assert.Equal(t, tt.wantRetries, d.retryConfig.MaxAttempts)
		})
	}
}

func TestDownload_Success(t *testing.T) {
	_ = setupTestDB(t)

	// Create test server
	content := []byte("test content for download")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	// Create temp directory for test
	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "testfile.txt")

	// Create downloader
	d := New(10*time.Second, 3)

	// Track progress calls
	var progressCalls int
	var lastDownloaded, lastTotal int64

	// Perform download
	result, err := d.Download(context.Background(), DownloadOptions{
		URL:             server.URL,
		BaseDestPath: destPath,
		OnProgress: func(downloaded, total int64) {
			progressCalls++
			lastDownloaded = downloaded
			lastTotal = total
		},
	})

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, result)
	// The server URL has no path and the response has no recognized Content-Type,
	// so detectFileExtension falls back to its ".mkv" default, which Download()
	// appends to BaseDestPath.
	wantPath := destPath + ".mkv"
	assert.Equal(t, wantPath, result.FilePath)
	assert.Equal(t, int64(len(content)), result.FileSize)
	assert.Greater(t, progressCalls, 0)
	assert.Equal(t, int64(len(content)), lastDownloaded)
	assert.Equal(t, int64(len(content)), lastTotal)

	// Verify file exists and content matches
	fileContent, err := os.ReadFile(wantPath)
	require.NoError(t, err)
	assert.Equal(t, content, fileContent)

	// Verify no temp file left
	_, err = os.Stat(wantPath + ".tmp")
	assert.True(t, os.IsNotExist(err))
}

func TestDownload_WithDatabaseTracking(t *testing.T) {
	db := setupTestDB(t)

	// Create test processed line
	lineURL := "http://example.com/stream.mkv"
	processedLine := &models.ProcessedLine{
		LineURL:     &lineURL,
		LineContent: "#EXTINF:-1,Test Movie",
		LineHash:    "testhash123",
		TvgName:     "Test Movie",
		GroupTitle:  "Movies",
		ContentType: models.ContentTypeMovies,
		State:       models.StateProcessed,
	}
	err := db.Create(processedLine).Error
	require.NoError(t, err)

	// Create test server
	content := []byte("test movie content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	// Create temp directory
	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "movie.mkv")

	// Create downloader
	d := New(10*time.Second, 3)

	// Perform download with database tracking
	result, err := d.Download(context.Background(), DownloadOptions{
		URL:             server.URL,
		BaseDestPath: destPath,
		ProcessedLineID: processedLine.ID,
	})

	// Assertions
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify database state updated to downloaded
	var updated models.ProcessedLine
	err = db.First(&updated, processedLine.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.StateDownloaded, updated.State)
	require.NotNil(t, updated.DownloadedAt)
	assert.WithinDuration(t, time.Now(), *updated.DownloadedAt, 5*time.Second)
}

func TestDownload_ValidationErrors(t *testing.T) {
	d := New(10*time.Second, 3)

	tests := []struct {
		name string
		opts DownloadOptions
	}{
		{
			name: "empty URL",
			opts: DownloadOptions{
				URL:             "",
				BaseDestPath: "/tmp/file.txt",
			},
		},
		{
			name: "empty destination",
			opts: DownloadOptions{
				URL:             "http://example.com/file",
				BaseDestPath: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := d.Download(context.Background(), tt.opts)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestDownload_HTTPErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"not found", http.StatusNotFound},
		{"forbidden", http.StatusForbidden},
		{"internal server error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server with error status
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			// Create temp directory
			tempDir := t.TempDir()
			destPath := filepath.Join(tempDir, "file.txt")

			// Create downloader
			d := New(10*time.Second, 3)

			// Perform download
			result, err := d.Download(context.Background(), DownloadOptions{
				URL:             server.URL,
				BaseDestPath: destPath,
			})

			// Should fail
			assert.Error(t, err)
			assert.Nil(t, result)

			// File should not exist
			_, err = os.Stat(destPath)
			assert.True(t, os.IsNotExist(err))
		})
	}
}

func TestDownload_Retry(t *testing.T) {
	attemptCount := 0
	content := []byte("test content after retry")

	// Create test server that fails first 2 attempts, then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	// Create temp directory
	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "file.txt")

	// Create downloader with 5 retry attempts
	d := New(10*time.Second, 5)

	// Perform download
	result, err := d.Download(context.Background(), DownloadOptions{
		URL:             server.URL,
		BaseDestPath: destPath,
	})

	// Should succeed after retries
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, attemptCount) // Should have tried 3 times
	assert.Equal(t, int64(len(content)), result.FileSize)

	// Verify file content. The server URL has no path and the response has no
	// recognized Content-Type, so detectFileExtension falls back to its ".mkv"
	// default, which Download() appends to BaseDestPath.
	fileContent, err := os.ReadFile(destPath + ".mkv")
	require.NoError(t, err)
	assert.Equal(t, content, fileContent)
}

func TestDownload_ContextCancellation(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // Slow response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("content"))
	}))
	defer server.Close()

	// Create temp directory
	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "file.txt")

	// Create downloader
	d := New(10*time.Second, 1)

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Perform download - should be cancelled
	result, err := d.Download(ctx, DownloadOptions{
		URL:             server.URL,
		BaseDestPath: destPath,
	})

	// Should fail due to context cancellation
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestDownload_DatabaseStateOnFailure(t *testing.T) {
	db := setupTestDB(t)

	// Create test processed line
	lineURL := "http://example.com/stream.mkv"
	processedLine := &models.ProcessedLine{
		LineURL:     &lineURL,
		LineContent: "#EXTINF:-1,Test Movie",
		LineHash:    "testhash456",
		TvgName:     "Test Movie",
		GroupTitle:  "Movies",
		ContentType: models.ContentTypeMovies,
		State:       models.StateProcessed,
	}
	err := db.Create(processedLine).Error
	require.NoError(t, err)

	// Create test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create temp directory
	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "movie.mkv")

	// Create downloader
	d := New(10*time.Second, 2)

	// Perform download with database tracking
	result, err := d.Download(context.Background(), DownloadOptions{
		URL:             server.URL,
		BaseDestPath: destPath,
		ProcessedLineID: processedLine.ID,
	})

	// Should fail
	assert.Error(t, err)
	assert.Nil(t, result)

	// Verify database state updated to failed
	var updated models.ProcessedLine
	err = db.First(&updated, processedLine.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.StateFailed, updated.State)
	assert.Nil(t, updated.DownloadedAt)
}

func TestProgressReader(t *testing.T) {
	content := []byte("test content for progress tracking")
	reader := &progressReader{
		reader: io.NopCloser(io.LimitReader(http.NoBody, int64(len(content)))),
		total:  int64(len(content)),
	}

	var progressUpdates int
	var lastDownloaded, lastTotal int64

	reader.onProgress = func(downloaded, total int64) {
		progressUpdates++
		lastDownloaded = downloaded
		lastTotal = total
	}

	// Simulate reading in chunks
	buf := make([]byte, 10)
	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		assert.Greater(t, n, 0)
	}

	assert.Greater(t, progressUpdates, 0)
	assert.Equal(t, int64(len(content)), lastTotal)
	// Note: lastDownloaded won't equal content length because we're using a limited reader
	_ = lastDownloaded
}

func TestDownload_CreatesDestinationDirectory(t *testing.T) {
	content := []byte("test content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	// Create temp directory
	tempDir := t.TempDir()
	// Use nested path that doesn't exist
	destPath := filepath.Join(tempDir, "movies", "test", "file.mkv")

	// Create downloader
	d := New(10*time.Second, 3)

	// Perform download
	result, err := d.Download(context.Background(), DownloadOptions{
		URL:             server.URL,
		BaseDestPath: destPath,
	})

	// Should succeed and create nested directories
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify file exists. The server URL has no path and the response has no
	// recognized Content-Type, so detectFileExtension falls back to its ".mkv"
	// default, which Download() appends to BaseDestPath (already ".mkv" here,
	// so the final file is "file.mkv.mkv").
	_, err = os.Stat(destPath + ".mkv")
	assert.NoError(t, err)

	// Verify directory structure created
	_, err = os.Stat(filepath.Join(tempDir, "movies", "test"))
	assert.NoError(t, err)
}

func TestDownload_URLStoredInDownloadInfo(t *testing.T) {
	setupTestDB(t)
	gdb := database.Get()
	if gdb == nil {
		t.Skip("skipping: database not available")
	}
	if sqlDB, err := gdb.DB(); err != nil || sqlDB.Ping() != nil {
		t.Skip("skipping: database not reachable")
	}

	// Create ProcessedLine directly in the global DB (same DB the downloader uses)
	lineURL := "http://example.com/url-tracking-test.mkv"
	processedLine := &models.ProcessedLine{
		LineURL:     &lineURL,
		LineContent: "#EXTINF:-1,URL Tracking Test",
		LineHash:    "urltracking001",
		TvgName:     "URL Tracking Test",
		GroupTitle:  "Movies",
		ContentType: models.ContentTypeMovies,
		State:       models.StateProcessed,
	}
	err := gdb.Create(processedLine).Error
	require.NoError(t, err)
	t.Cleanup(func() { gdb.Delete(processedLine) })

	// Create test server
	content := []byte("url tracking test content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "url-tracking-test.mkv")

	d := New(10*time.Second, 3)
	_, err = d.Download(context.Background(), DownloadOptions{
		URL:             server.URL,
		BaseDestPath:    destPath,
		ProcessedLineID: processedLine.ID,
	})
	require.NoError(t, err)

	// Reload ProcessedLine to get DownloadInfoID
	var updated models.ProcessedLine
	err = gdb.First(&updated, processedLine.ID).Error
	require.NoError(t, err)
	require.NotNil(t, updated.DownloadInfoID, "DownloadInfoID should be set after download")

	// Check DownloadInfo has URL stored
	var dlInfo models.DownloadInfo
	err = gdb.First(&dlInfo, *updated.DownloadInfoID).Error
	require.NoError(t, err)
	assert.Equal(t, server.URL, dlInfo.URL, "DownloadInfo.URL should match the download URL")
	assert.Equal(t, string(models.DownloadStatusCompleted), dlInfo.Status)

	t.Cleanup(func() { gdb.Delete(&dlInfo) })
}

func TestDownload_RetryCountIncrements(t *testing.T) {
	setupTestDB(t)
	gdb := database.Get()
	if gdb == nil {
		t.Skip("skipping: database not available")
	}
	if sqlDB, err := gdb.DB(); err != nil || sqlDB.Ping() != nil {
		t.Skip("skipping: database not reachable")
	}

	lineURL := "http://example.com/retry-count-test.mkv"
	processedLine := &models.ProcessedLine{
		LineURL:     &lineURL,
		LineContent: "#EXTINF:-1,Retry Count Test",
		LineHash:    "retrycount001",
		TvgName:     "Retry Count Test",
		GroupTitle:  "Movies",
		ContentType: models.ContentTypeMovies,
		State:       models.StateProcessed,
	}
	err := gdb.Create(processedLine).Error
	require.NoError(t, err)
	t.Cleanup(func() { gdb.Delete(processedLine) })

	// Server fails first 2 attempts, succeeds on 3rd
	attemptCount := 0
	content := []byte("retry count test content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "retry-count-test.mkv")

	d := New(10*time.Second, 5)
	_, err = d.Download(context.Background(), DownloadOptions{
		URL:             server.URL,
		BaseDestPath:    destPath,
		ProcessedLineID: processedLine.ID,
	})
	require.NoError(t, err)

	// Reload ProcessedLine to get DownloadInfoID
	var updated models.ProcessedLine
	err = gdb.First(&updated, processedLine.ID).Error
	require.NoError(t, err)
	require.NotNil(t, updated.DownloadInfoID)

	// retry_count should reflect 2 retries (attempts 1 and 2 failed, attempt 3 succeeded)
	var dlInfo models.DownloadInfo
	err = gdb.First(&dlInfo, *updated.DownloadInfoID).Error
	require.NoError(t, err)
	assert.Equal(t, 2, dlInfo.RetryCount, "retry_count should be 2 after 2 failed attempts")
	assert.Equal(t, server.URL, dlInfo.URL)

	t.Cleanup(func() { gdb.Delete(&dlInfo) })
}

func TestDownload_SetsTargetPathBeforeTransfer(t *testing.T) {
	setupTestDB(t)
	gdb := database.Get()
	if gdb == nil {
		t.Skip("skipping: database not available")
	}
	if sqlDB, err := gdb.DB(); err != nil || sqlDB.Ping() != nil {
		t.Skip("skipping: database not reachable")
	}

	lineURL := "http://example.com/target-path-test.mkv"
	processedLine := &models.ProcessedLine{
		LineURL:     &lineURL,
		LineContent: "#EXTINF:-1,Target Path Test",
		LineHash:    "targetpath001",
		TvgName:     "Target Path Test",
		GroupTitle:  "Movies",
		ContentType: models.ContentTypeMovies,
		State:       models.StateProcessed,
	}
	require.NoError(t, gdb.Create(processedLine).Error)
	t.Cleanup(func() { gdb.Delete(processedLine) })

	// Capture target_path from inside the HTTP handler, i.e. before the
	// transfer (and any response) has completed.
	var capturedTargetPath *string
	content := []byte("target path test content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pl models.ProcessedLine
		if gdb.First(&pl, processedLine.ID).Error == nil && pl.DownloadInfoID != nil {
			var dl models.DownloadInfo
			if gdb.First(&dl, *pl.DownloadInfoID).Error == nil {
				capturedTargetPath = dl.TargetPath
			}
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "target-path-test")

	d := New(10*time.Second, 3)
	_, err := d.Download(context.Background(), DownloadOptions{
		URL:             server.URL,
		BaseDestPath:    destPath,
		ProcessedLineID: processedLine.ID,
	})
	require.NoError(t, err)

	// The server URL has no path and the response has no recognized
	// Content-Type, so detectFileExtension falls back to its ".mkv" default.
	require.NotNil(t, capturedTargetPath, "target_path should be set before the HTTP transfer starts")
	assert.Equal(t, destPath+".mkv", *capturedTargetPath)
}

func TestDownload_StagingPathStableAcrossRetries(t *testing.T) {
	setupTestDB(t)
	gdb := database.Get()
	if gdb == nil {
		t.Skip("skipping: database not available")
	}
	if sqlDB, err := gdb.DB(); err != nil || sqlDB.Ping() != nil {
		t.Skip("skipping: database not reachable")
	}

	lineURL := "http://example.com/staging-stable-test.mkv"
	processedLine := &models.ProcessedLine{
		LineURL:     &lineURL,
		LineContent: "#EXTINF:-1,Staging Stable Test",
		LineHash:    "stagingstable001",
		TvgName:     "Staging Stable Test",
		GroupTitle:  "Movies",
		ContentType: models.ContentTypeMovies,
		State:       models.StateProcessed,
	}
	require.NoError(t, gdb.Create(processedLine).Error)
	t.Cleanup(func() { gdb.Delete(processedLine) })

	content := []byte("staging stable test content")
	attemptCount := 0
	var mu sync.Mutex
	var observedStagingPaths []string

	// Fail the first 2 attempts, succeed on the 3rd, recording staging_path
	// as observed at the start of each attempt (requests are sequential, so
	// no concurrent access to observedStagingPaths).
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		var pl models.ProcessedLine
		if gdb.First(&pl, processedLine.ID).Error == nil && pl.DownloadInfoID != nil {
			var dl models.DownloadInfo
			if gdb.First(&dl, *pl.DownloadInfoID).Error == nil && dl.StagingPath != nil {
				mu.Lock()
				observedStagingPaths = append(observedStagingPaths, *dl.StagingPath)
				mu.Unlock()
			}
		}

		if attemptCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "staging-stable-test")

	d := New(10*time.Second, 5)
	_, err := d.Download(context.Background(), DownloadOptions{
		URL:             server.URL,
		BaseDestPath:    destPath,
		ProcessedLineID: processedLine.ID,
	})
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	require.GreaterOrEqual(t, len(observedStagingPaths), 2, "expected staging_path to be observed across at least 2 attempts")
	for _, p := range observedStagingPaths {
		assert.Equal(t, observedStagingPaths[0], p, "staging_path must stay stable across retries within one Download() call")
	}
	assert.Contains(t, observedStagingPaths[0], "download.tmp")
}

func TestDownload_ClearsTargetAndStagingPathOnCompletion(t *testing.T) {
	setupTestDB(t)
	gdb := database.Get()
	if gdb == nil {
		t.Skip("skipping: database not available")
	}
	if sqlDB, err := gdb.DB(); err != nil || sqlDB.Ping() != nil {
		t.Skip("skipping: database not reachable")
	}

	lineURL := "http://example.com/clear-paths-test.mkv"
	processedLine := &models.ProcessedLine{
		LineURL:     &lineURL,
		LineContent: "#EXTINF:-1,Clear Paths Test",
		LineHash:    "clearpaths001",
		TvgName:     "Clear Paths Test",
		GroupTitle:  "Movies",
		ContentType: models.ContentTypeMovies,
		State:       models.StateProcessed,
	}
	require.NoError(t, gdb.Create(processedLine).Error)
	t.Cleanup(func() { gdb.Delete(processedLine) })

	content := []byte("clear paths test content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "clear-paths-test")

	d := New(10*time.Second, 3)
	_, err := d.Download(context.Background(), DownloadOptions{
		URL:             server.URL,
		BaseDestPath:    destPath,
		ProcessedLineID: processedLine.ID,
	})
	require.NoError(t, err)

	var pl models.ProcessedLine
	require.NoError(t, gdb.First(&pl, processedLine.ID).Error)
	require.NotNil(t, pl.DownloadInfoID)

	var dl models.DownloadInfo
	require.NoError(t, gdb.First(&dl, *pl.DownloadInfoID).Error)
	assert.Equal(t, string(models.DownloadStatusCompleted), dl.Status)
	assert.NotNil(t, dl.DownloadPath)
	assert.Nil(t, dl.TargetPath, "target_path should be cleared once the download completes")
	assert.Nil(t, dl.StagingPath, "staging_path should be cleared once the download completes")

	t.Cleanup(func() { gdb.Delete(&dl) })
}

func TestDownload_FailureMidTransferLeavesTargetAndStagingPathPopulated(t *testing.T) {
	setupTestDB(t)
	gdb := database.Get()
	if gdb == nil {
		t.Skip("skipping: database not available")
	}
	if sqlDB, err := gdb.DB(); err != nil || sqlDB.Ping() != nil {
		t.Skip("skipping: database not reachable")
	}

	lineURL := "http://example.com/mid-transfer-failure-test.mkv"
	processedLine := &models.ProcessedLine{
		LineURL:     &lineURL,
		LineContent: "#EXTINF:-1,Mid Transfer Failure Test",
		LineHash:    "midtransferfail001",
		TvgName:     "Mid Transfer Failure Test",
		GroupTitle:  "Movies",
		ContentType: models.ContentTypeMovies,
		State:       models.StateProcessed,
	}
	require.NoError(t, gdb.Create(processedLine).Error)
	t.Cleanup(func() { gdb.Delete(processedLine) })

	// Claim more content than is actually sent, then hijack and close the
	// connection early to trigger an "unexpected EOF" while io.Copy reads
	// the body (internal/downloader/downloader.go:394-396).
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("short body"))
		if hj, ok := w.(http.Hijacker); ok {
			if conn, _, err := hj.Hijack(); err == nil {
				conn.Close()
			}
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "mid-transfer-failure-test")

	// A single attempt (no retries) so the failed state is not overwritten.
	d := New(10*time.Second, 1)
	_, err := d.Download(context.Background(), DownloadOptions{
		URL:             server.URL,
		BaseDestPath:    destPath,
		ProcessedLineID: processedLine.ID,
	})
	require.Error(t, err)

	var pl models.ProcessedLine
	require.NoError(t, gdb.First(&pl, processedLine.ID).Error)
	require.NotNil(t, pl.DownloadInfoID)

	var dl models.DownloadInfo
	require.NoError(t, gdb.First(&dl, *pl.DownloadInfoID).Error)
	assert.Equal(t, string(models.DownloadStatusFailed), dl.Status)
	require.NotNil(t, dl.TargetPath)
	assert.Equal(t, destPath+".mkv", *dl.TargetPath)
	require.NotNil(t, dl.StagingPath)
	assert.Contains(t, *dl.StagingPath, "download.tmp")
	assert.Nil(t, dl.DownloadPath)

	t.Cleanup(func() { gdb.Delete(&dl) })
}

func TestDownload_FailureDuringMoveLeavesTargetAndStagingPathPopulated(t *testing.T) {
	setupTestDB(t)
	gdb := database.Get()
	if gdb == nil {
		t.Skip("skipping: database not available")
	}
	if sqlDB, err := gdb.DB(); err != nil || sqlDB.Ping() != nil {
		t.Skip("skipping: database not reachable")
	}

	lineURL := "http://example.com/move-failure-test.mkv"
	processedLine := &models.ProcessedLine{
		LineURL:     &lineURL,
		LineContent: "#EXTINF:-1,Move Failure Test",
		LineHash:    "movefailure001",
		TvgName:     "Move Failure Test",
		GroupTitle:  "Movies",
		ContentType: models.ContentTypeMovies,
		State:       models.StateProcessed,
	}
	require.NoError(t, gdb.Create(processedLine).Error)
	t.Cleanup(func() { gdb.Delete(processedLine) })

	content := []byte("move failure test content")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "move-failure-test")
	// Pre-create a (non-empty-capable) directory at the exact path the
	// downloaded file must land on, so the move-to-destination step
	// (internal/downloader/downloader.go:254-273) fails deterministically:
	// os.Rename refuses to replace a directory with a file, and the
	// copy fallback's os.Create refuses to open a directory for writing.
	finalDestPath := destPath + ".mkv"
	require.NoError(t, os.MkdirAll(finalDestPath, 0755))

	d := New(10*time.Second, 1)
	_, err := d.Download(context.Background(), DownloadOptions{
		URL:             server.URL,
		BaseDestPath:    destPath,
		ProcessedLineID: processedLine.ID,
	})
	require.Error(t, err)

	var pl models.ProcessedLine
	require.NoError(t, gdb.First(&pl, processedLine.ID).Error)
	require.NotNil(t, pl.DownloadInfoID)

	var dl models.DownloadInfo
	require.NoError(t, gdb.First(&dl, *pl.DownloadInfoID).Error)
	assert.Equal(t, string(models.DownloadStatusFailed), dl.Status)
	require.NotNil(t, dl.TargetPath)
	assert.Equal(t, finalDestPath, *dl.TargetPath)
	require.NotNil(t, dl.StagingPath)
	assert.Contains(t, *dl.StagingPath, "download.tmp")
	assert.Nil(t, dl.DownloadPath)

	t.Cleanup(func() { gdb.Delete(&dl) })
}
