package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/models"
)

// 4.1: an id that doesn't belong to any duplicate group is refused (404,
// not_a_duplicate) without changing any record.
func TestResolveDuplicateDownload_NotADuplicate(t *testing.T) {
	db := setupTestDB(t)
	server := NewServer()

	moviePath := t.TempDir()
	filePath := filepath.Join(moviePath, "movie.mkv")
	if err := os.WriteFile(filePath, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	movie := models.Movie{TMDBID: 42, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)

	dl := models.DownloadInfo{URL: "http://example.com/dune", Status: string(models.DownloadStatusCompleted), DownloadPath: &filePath}
	db.Create(&dl)
	line := models.ProcessedLine{
		LineContent: "dune", LineHash: "hash-dune-single", TvgName: "Dune",
		ContentType: models.ContentTypeMovies, MovieID: &movie.ID, DownloadInfoID: &dl.ID,
		State: models.StateDownloaded, ProcessedAt: time.Now(),
	}
	db.Create(&line)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/downloads/%d/resolve-duplicate", dl.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", w.Code, w.Body.String())
	}

	var body ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body.Error != "not_a_duplicate" {
		t.Errorf("expected error code not_a_duplicate, got %q", body.Error)
	}

	var unchanged models.DownloadInfo
	db.First(&unchanged, dl.ID)
	if unchanged.Status != string(models.DownloadStatusCompleted) {
		t.Errorf("expected status to remain completed, got %q", unchanged.Status)
	}
}

func TestResolveDuplicateDownload_InvalidID(t *testing.T) {
	_ = setupTestDB(t)
	server := NewServer()

	req, _ := http.NewRequest("POST", "/api/v1/downloads/not-a-number/resolve-duplicate", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

// 4.2/4.3: one occurrence's file is missing, the other's is present - only the
// missing one is cancelled, and the response reports both outcomes.
func TestResolveDuplicateDownload_OneMissingOnePresent(t *testing.T) {
	db := setupTestDB(t)
	server := NewServer()

	dir := t.TempDir()
	presentPath := filepath.Join(dir, "present.mkv")
	if err := os.WriteFile(presentPath, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	missingPath := filepath.Join(dir, "missing.mkv") // never created

	movie := models.Movie{TMDBID: 42, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)

	dlPresent := models.DownloadInfo{URL: "http://example.com/present", Status: string(models.DownloadStatusCompleted), DownloadPath: &presentPath}
	db.Create(&dlPresent)
	dlMissing := models.DownloadInfo{URL: "http://example.com/missing", Status: string(models.DownloadStatusCompleted), DownloadPath: &missingPath}
	db.Create(&dlMissing)

	linePresent := models.ProcessedLine{
		LineContent: "present", LineHash: "hash-present", TvgName: "Dune",
		ContentType: models.ContentTypeMovies, MovieID: &movie.ID, DownloadInfoID: &dlPresent.ID,
		State: models.StateDownloaded, ProcessedAt: time.Now(),
	}
	db.Create(&linePresent)
	lineMissing := models.ProcessedLine{
		LineContent: "missing", LineHash: "hash-missing", TvgName: "Dune",
		ContentType: models.ContentTypeMovies, MovieID: &movie.ID, DownloadInfoID: &dlMissing.ID,
		State: models.StateDownloaded, ProcessedAt: time.Now(),
	}
	db.Create(&lineMissing)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/downloads/%d/resolve-duplicate", dlMissing.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ResolveDuplicateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if len(resp.Occurrences) != 2 {
		t.Fatalf("expected 2 occurrences in response, got %d", len(resp.Occurrences))
	}

	outcomes := map[uint]string{}
	for _, o := range resp.Occurrences {
		outcomes[o.DownloadInfoID] = o.Outcome
	}
	if outcomes[dlPresent.ID] != duplicateOutcomeUnchanged {
		t.Errorf("expected present occurrence to be unchanged, got %q", outcomes[dlPresent.ID])
	}
	if outcomes[dlMissing.ID] != duplicateOutcomeCancelled {
		t.Errorf("expected missing occurrence to be cancelled, got %q", outcomes[dlMissing.ID])
	}

	var present models.DownloadInfo
	db.First(&present, dlPresent.ID)
	if present.Status != string(models.DownloadStatusCompleted) {
		t.Errorf("expected present occurrence's status to remain completed, got %q", present.Status)
	}

	var missing models.DownloadInfo
	db.First(&missing, dlMissing.ID)
	if missing.Status != string(models.DownloadStatusCancelled) {
		t.Errorf("expected missing occurrence's status to become cancelled, got %q", missing.Status)
	}

	var missingLine models.ProcessedLine
	db.First(&missingLine, lineMissing.ID)
	if missingLine.State != models.StateCancelled {
		t.Errorf("expected missing occurrence's processed_line state to become cancelled, got %q", missingLine.State)
	}

	var presentLine models.ProcessedLine
	db.First(&presentLine, linePresent.ID)
	if presentLine.State != models.StateDownloaded {
		t.Errorf("expected present occurrence's processed_line state to remain downloaded, got %q", presentLine.State)
	}
}

// 4.2: both files present - nothing changes, group still reported as
// duplicate afterwards.
func TestResolveDuplicateDownload_BothPresent(t *testing.T) {
	db := setupTestDB(t)
	server := NewServer()

	dir := t.TempDir()
	pathA := filepath.Join(dir, "a.mkv")
	pathB := filepath.Join(dir, "b.mkv")
	if err := os.WriteFile(pathA, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	if err := os.WriteFile(pathB, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	movie := models.Movie{TMDBID: 42, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)

	dlA := models.DownloadInfo{URL: "http://example.com/a", Status: string(models.DownloadStatusCompleted), DownloadPath: &pathA}
	db.Create(&dlA)
	dlB := models.DownloadInfo{URL: "http://example.com/b", Status: string(models.DownloadStatusCompleted), DownloadPath: &pathB}
	db.Create(&dlB)

	lineA := models.ProcessedLine{
		LineContent: "a", LineHash: "hash-a", TvgName: "Dune",
		ContentType: models.ContentTypeMovies, MovieID: &movie.ID, DownloadInfoID: &dlA.ID,
		State: models.StateDownloaded, ProcessedAt: time.Now(),
	}
	db.Create(&lineA)
	lineB := models.ProcessedLine{
		LineContent: "b", LineHash: "hash-b", TvgName: "Dune",
		ContentType: models.ContentTypeMovies, MovieID: &movie.ID, DownloadInfoID: &dlB.ID,
		State: models.StateDownloaded, ProcessedAt: time.Now(),
	}
	db.Create(&lineB)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/downloads/%d/resolve-duplicate", dlA.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ResolveDuplicateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	for _, o := range resp.Occurrences {
		if o.Outcome != duplicateOutcomeUnchanged {
			t.Errorf("expected all occurrences unchanged, got %q for %d", o.Outcome, o.DownloadInfoID)
		}
	}

	var a, b models.DownloadInfo
	db.First(&a, dlA.ID)
	db.First(&b, dlB.ID)
	if a.Status != string(models.DownloadStatusCompleted) || b.Status != string(models.DownloadStatusCompleted) {
		t.Error("expected both occurrences to remain completed")
	}
}

// 4.2: both files missing - both are cancelled.
func TestResolveDuplicateDownload_BothMissing(t *testing.T) {
	db := setupTestDB(t)
	server := NewServer()

	dir := t.TempDir()
	pathA := filepath.Join(dir, "a-gone.mkv")
	pathB := filepath.Join(dir, "b-gone.mkv")

	movie := models.Movie{TMDBID: 42, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)

	dlA := models.DownloadInfo{URL: "http://example.com/a", Status: string(models.DownloadStatusCompleted), DownloadPath: &pathA}
	db.Create(&dlA)
	dlB := models.DownloadInfo{URL: "http://example.com/b", Status: string(models.DownloadStatusCompleted), DownloadPath: &pathB}
	db.Create(&dlB)

	lineA := models.ProcessedLine{
		LineContent: "a", LineHash: "hash-a-gone", TvgName: "Dune",
		ContentType: models.ContentTypeMovies, MovieID: &movie.ID, DownloadInfoID: &dlA.ID,
		State: models.StateDownloaded, ProcessedAt: time.Now(),
	}
	db.Create(&lineA)
	lineB := models.ProcessedLine{
		LineContent: "b", LineHash: "hash-b-gone", TvgName: "Dune",
		ContentType: models.ContentTypeMovies, MovieID: &movie.ID, DownloadInfoID: &dlB.ID,
		State: models.StateDownloaded, ProcessedAt: time.Now(),
	}
	db.Create(&lineB)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/downloads/%d/resolve-duplicate", dlA.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ResolveDuplicateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	for _, o := range resp.Occurrences {
		if o.Outcome != duplicateOutcomeCancelled {
			t.Errorf("expected all occurrences cancelled, got %q for %d", o.Outcome, o.DownloadInfoID)
		}
	}

	var a, b models.DownloadInfo
	db.First(&a, dlA.ID)
	db.First(&b, dlB.ID)
	if a.Status != string(models.DownloadStatusCancelled) || b.Status != string(models.DownloadStatusCancelled) {
		t.Error("expected both occurrences to become cancelled")
	}
}
