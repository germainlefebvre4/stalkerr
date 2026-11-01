package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/gorm"
)

// seedCancelledOccurrence creates a movie with one cancelled, forced-or-not
// occurrence linked to a cancelled DownloadInfo, mirroring what the scheduler
// leaves behind after the "sibling already downloaded" cancellation.
func seedCancelledOccurrence(t *testing.T, db *gorm.DB, movie *models.Movie, hash string, url string) (models.ProcessedLine, models.DownloadInfo) {
	t.Helper()
	errMsg := "previous failure"
	lockedAt := time.Now()
	lockedBy := "worker-1"
	dl := models.DownloadInfo{
		URL:          url,
		Status:       string(models.DownloadStatusCancelled),
		RetryCount:   3,
		ErrorMessage: &errMsg,
		LockedAt:     &lockedAt,
		LockedBy:     &lockedBy,
		IsForced:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := db.Create(&dl).Error; err != nil {
		t.Fatalf("create download: %v", err)
	}
	line := models.ProcessedLine{
		LineContent:    hash,
		LineHash:       hash,
		LineURL:        strPtr2(url),
		TvgName:        "Dune",
		ContentType:    models.ContentTypeMovies,
		MovieID:        &movie.ID,
		DownloadInfoID: &dl.ID,
		State:          models.StateCancelled,
		ProcessedAt:    time.Now(),
	}
	if err := db.Create(&line).Error; err != nil {
		t.Fatalf("create line: %v", err)
	}
	return line, dl
}

func newRadarrDuneServer(t *testing.T) *httptest.Server {
	t.Helper()
	moviePath := t.TempDir()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 5, "title": "Dune", "year": 2021, "tmdbId": 42, "path": moviePath, "monitored": true},
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func postForce(server *Server, id uint) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/items/%d/force-download", id), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)
	return w
}

func TestForceDownloadItem_ReactivatesCancelledOccurrence(t *testing.T) {
	db := setupTestDB(t)
	radarr := newRadarrDuneServer(t)
	newForceDownloadTestConfig(t, radarr.URL, "")
	server := NewServer()

	movie := models.Movie{TMDBID: 42, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)
	target, targetDL := seedCancelledOccurrence(t, db, &movie, "hash-target", "http://example.com/target")
	other, otherDL := seedCancelledOccurrence(t, db, &movie, "hash-other", "http://example.com/other")

	w := postForce(server, target.ID)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", w.Code, w.Body.String())
	}

	var dl models.DownloadInfo
	db.First(&dl, targetDL.ID)
	if dl.Status != string(models.DownloadStatusPending) {
		t.Errorf("expected pending, got %q", dl.Status)
	}
	if dl.RetryCount != 0 {
		t.Errorf("expected retry_count 0, got %d", dl.RetryCount)
	}
	if dl.ErrorMessage != nil || dl.LockedAt != nil || dl.LockedBy != nil {
		t.Errorf("expected error and lock cleared, got %v %v %v", dl.ErrorMessage, dl.LockedAt, dl.LockedBy)
	}
	if !dl.IsForced || dl.DownloadPath == nil || *dl.DownloadPath == "" {
		t.Errorf("expected forced with a recorded download path, got forced=%v path=%v", dl.IsForced, dl.DownloadPath)
	}

	var line models.ProcessedLine
	db.First(&line, target.ID)
	if line.State != models.StateProcessed {
		t.Errorf("expected line processed, got %q", line.State)
	}

	// Other cancelled occurrences of the same movie stay cancelled.
	var otherDLAfter models.DownloadInfo
	db.First(&otherDLAfter, otherDL.ID)
	if otherDLAfter.Status != string(models.DownloadStatusCancelled) || otherDLAfter.RetryCount != 3 {
		t.Errorf("expected sibling download untouched, got status=%q retry=%d", otherDLAfter.Status, otherDLAfter.RetryCount)
	}
	var otherLine models.ProcessedLine
	db.First(&otherLine, other.ID)
	if otherLine.State != models.StateCancelled {
		t.Errorf("expected sibling line still cancelled, got %q", otherLine.State)
	}

	// The reactivated record is visible in the default (active) listing.
	req, _ := http.NewRequest("GET", "/api/v1/downloads?status=active", nil)
	lw := httptest.NewRecorder()
	server.router.ServeHTTP(lw, req)
	if lw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", lw.Code)
	}
	var resp PaginatedResponse
	if err := json.Unmarshal(lw.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("expected exactly the reactivated download in the active view, got total %d", resp.Total)
	}
}

func TestForceDownloadItem_ReForceDownloadedStaysRefusedAndUnchanged(t *testing.T) {
	db := setupTestDB(t)
	radarr := newRadarrDuneServer(t)
	newForceDownloadTestConfig(t, radarr.URL, "")
	server := NewServer()

	movie := models.Movie{TMDBID: 42, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)
	path := "/existing/Dune (2021).mkv"
	dl := models.DownloadInfo{
		URL:          "http://example.com/done",
		Status:       string(models.DownloadStatusCompleted),
		DownloadPath: &path,
		RetryCount:   2,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	db.Create(&dl)
	line := models.ProcessedLine{
		LineContent:    "done",
		LineHash:       "hash-done",
		LineURL:        strPtr2("http://example.com/done"),
		TvgName:        "Dune",
		ContentType:    models.ContentTypeMovies,
		MovieID:        &movie.ID,
		DownloadInfoID: &dl.ID,
		State:          models.StateDownloaded,
		ProcessedAt:    time.Now(),
	}
	db.Create(&line)

	w := postForce(server, line.ID)
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
	var body ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &body)
	if body.Error != "already_downloaded" {
		t.Errorf("expected already_downloaded, got %q", body.Error)
	}

	var after models.DownloadInfo
	db.First(&after, dl.ID)
	if after.Status != string(models.DownloadStatusCompleted) || after.IsForced || after.RetryCount != 2 || *after.DownloadPath != path {
		t.Errorf("expected download untouched, got %+v", after)
	}
	var lineAfter models.ProcessedLine
	db.First(&lineAfter, line.ID)
	if lineAfter.State != models.StateDownloaded {
		t.Errorf("expected line still downloaded, got %q", lineAfter.State)
	}
}
