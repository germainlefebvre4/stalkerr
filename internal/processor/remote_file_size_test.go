package processor

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/models"
)

// newRemoteFileSizeTestLine creates and persists a ProcessedLine eligible (by
// default) for the remote file size backfill, with a unique line_hash.
func newRemoteFileSizeTestLine(t *testing.T, contentType models.ContentType, url string, checkedAt *time.Time) models.ProcessedLine {
	t.Helper()

	now := time.Now()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s-%d", t.Name(), time.Now().UnixNano())))
	line := models.ProcessedLine{
		LineContent:             fmt.Sprintf("#EXTINF:-1,%s\n%s", t.Name(), url),
		LineURL:                 &url,
		LineHash:                hex.EncodeToString(hash[:]),
		TvgName:                 t.Name(),
		GroupTitle:              "Test Group",
		ProcessedAt:             now,
		ContentType:             contentType,
		State:                   models.StateProcessed,
		RemoteFileSizeCheckedAt: checkedAt,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	db := database.Get()
	if err := db.Create(&line).Error; err != nil {
		t.Fatalf("failed to create test processed line: %v", err)
	}
	return line
}

func reloadProcessedLine(t *testing.T, id uint) models.ProcessedLine {
	t.Helper()
	var line models.ProcessedLine
	if err := database.Get().First(&line, id).Error; err != nil {
		t.Fatalf("failed to reload processed line %d: %v", id, err)
	}
	return line
}

// TestBackfillRemoteFileSize_HeadSuccess verifies that a HEAD request returning
// a usable Content-Length is persisted as remote_file_size without a fallback GET.
func TestBackfillRemoteFileSize_HeadSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	getCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getCalls++
		}
		w.Header().Set("Content-Length", "123456")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	line := newRemoteFileSizeTestLine(t, models.ContentTypeMovies, srv.URL, nil)

	stats, err := BackfillRemoteFileSize(database.Get(), logger.AppLogger(), 5*time.Second, 200)
	if err != nil {
		t.Fatalf("BackfillRemoteFileSize error: %v", err)
	}

	if stats.Found != 1 {
		t.Errorf("expected Found=1, got %d", stats.Found)
	}
	if stats.Errors != 0 {
		t.Errorf("expected Errors=0, got %d", stats.Errors)
	}
	if getCalls != 0 {
		t.Errorf("expected no GET fallback calls, got %d", getCalls)
	}

	updated := reloadProcessedLine(t, line.ID)
	if updated.RemoteFileSize == nil || *updated.RemoteFileSize != 123456 {
		t.Errorf("expected remote_file_size=123456, got %v", updated.RemoteFileSize)
	}
	if updated.RemoteFileSizeCheckedAt == nil {
		t.Error("expected remote_file_size_checked_at to be set")
	}
}

// TestBackfillRemoteFileSize_HeadFailureFallsBackToRangeGet verifies that when
// HEAD fails, a ranged GET is used to obtain the size from Content-Range.
func TestBackfillRemoteFileSize_HeadFailureFallsBackToRangeGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Range") != "bytes=0-0" {
			http.Error(w, "range header missing", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Range", "bytes 0-0/987654")
		w.WriteHeader(http.StatusPartialContent)
		w.Write([]byte{0})
	}))
	defer srv.Close()

	line := newRemoteFileSizeTestLine(t, models.ContentTypeTVShows, srv.URL, nil)

	stats, err := BackfillRemoteFileSize(database.Get(), logger.AppLogger(), 5*time.Second, 200)
	if err != nil {
		t.Fatalf("BackfillRemoteFileSize error: %v", err)
	}

	if stats.Found != 1 {
		t.Errorf("expected Found=1, got %d", stats.Found)
	}

	updated := reloadProcessedLine(t, line.ID)
	if updated.RemoteFileSize == nil || *updated.RemoteFileSize != 987654 {
		t.Errorf("expected remote_file_size=987654, got %v", updated.RemoteFileSize)
	}
}

// TestBackfillRemoteFileSize_BothProbesFail verifies that a line whose HEAD and
// ranged GET probes both fail is marked as checked, with no size and no abort.
func TestBackfillRemoteFileSize_BothProbesFail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	line := newRemoteFileSizeTestLine(t, models.ContentTypeUncategorized, srv.URL, nil)

	stats, err := BackfillRemoteFileSize(database.Get(), logger.AppLogger(), 5*time.Second, 200)
	if err != nil {
		t.Fatalf("BackfillRemoteFileSize error: %v", err)
	}

	if stats.Errors != 1 {
		t.Errorf("expected Errors=1, got %d", stats.Errors)
	}
	if stats.Found != 0 {
		t.Errorf("expected Found=0, got %d", stats.Found)
	}

	updated := reloadProcessedLine(t, line.ID)
	if updated.RemoteFileSize != nil {
		t.Errorf("expected remote_file_size to remain nil, got %v", *updated.RemoteFileSize)
	}
	if updated.RemoteFileSizeCheckedAt == nil {
		t.Error("expected remote_file_size_checked_at to be set even on failure, so the line is never retried")
	}
}

// TestBackfillRemoteFileSize_ChannelsExcluded verifies that channel entries are
// never selected by the backfill query and never probed.
func TestBackfillRemoteFileSize_ChannelsExcluded(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Length", "111")
	}))
	defer srv.Close()

	line := newRemoteFileSizeTestLine(t, models.ContentTypeChannels, srv.URL, nil)

	stats, err := BackfillRemoteFileSize(database.Get(), logger.AppLogger(), 5*time.Second, 200)
	if err != nil {
		t.Fatalf("BackfillRemoteFileSize error: %v", err)
	}

	if stats.Checked != 0 {
		t.Errorf("expected Checked=0 for a channel-only backlog, got %d", stats.Checked)
	}
	if calls != 0 {
		t.Errorf("expected no HTTP calls for a channel entry, got %d", calls)
	}

	updated := reloadProcessedLine(t, line.ID)
	if updated.RemoteFileSizeCheckedAt != nil {
		t.Error("expected remote_file_size_checked_at to remain unset for a channel entry")
	}
}

// TestBackfillRemoteFileSize_PerRunCapRespected verifies that only perRunCap
// lines are probed even when more eligible lines exist.
func TestBackfillRemoteFileSize_PerRunCapRespected(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "42")
	}))
	defer srv.Close()

	const total = 5
	const perRunCap = 2
	for i := 0; i < total; i++ {
		newRemoteFileSizeTestLine(t, models.ContentTypeMovies, srv.URL, nil)
	}

	stats, err := BackfillRemoteFileSize(database.Get(), logger.AppLogger(), 5*time.Second, perRunCap)
	if err != nil {
		t.Fatalf("BackfillRemoteFileSize error: %v", err)
	}

	if stats.Checked != perRunCap {
		t.Errorf("expected Checked=%d, got %d", perRunCap, stats.Checked)
	}

	var remainingUnchecked int64
	database.Get().Model(&models.ProcessedLine{}).
		Where("remote_file_size_checked_at IS NULL").
		Count(&remainingUnchecked)
	if remainingUnchecked != total-perRunCap {
		t.Errorf("expected %d lines still unchecked, got %d", total-perRunCap, remainingUnchecked)
	}
}

// TestBackfillRemoteFileSize_AlreadyCheckedRowsSkipped verifies that a line with
// remote_file_size_checked_at already set is never re-probed.
func TestBackfillRemoteFileSize_AlreadyCheckedRowsSkipped(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Length", "999")
	}))
	defer srv.Close()

	alreadyChecked := time.Now().Add(-24 * time.Hour)
	line := newRemoteFileSizeTestLine(t, models.ContentTypeMovies, srv.URL, &alreadyChecked)

	stats, err := BackfillRemoteFileSize(database.Get(), logger.AppLogger(), 5*time.Second, 200)
	if err != nil {
		t.Fatalf("BackfillRemoteFileSize error: %v", err)
	}

	if stats.Checked != 0 {
		t.Errorf("expected Checked=0 for an already-checked line, got %d", stats.Checked)
	}
	if calls != 0 {
		t.Errorf("expected no HTTP calls for an already-checked line, got %d", calls)
	}

	updated := reloadProcessedLine(t, line.ID)
	if updated.RemoteFileSize != nil {
		t.Errorf("expected remote_file_size to remain nil, got %v", *updated.RemoteFileSize)
	}
}
