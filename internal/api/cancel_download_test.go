package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/models"
)

func TestCancelDownload_InvalidID(t *testing.T) {
	_ = setupTestDB(t)
	server := NewServer()

	req, _ := http.NewRequest("POST", "/api/v1/downloads/not-a-number/cancel", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCancelDownload_NotFound(t *testing.T) {
	_ = setupTestDB(t)
	server := NewServer()

	req, _ := http.NewRequest("POST", "/api/v1/downloads/999/cancel", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", w.Code, w.Body.String())
	}
}

// 4.2: refused with 409 for each ineligible status.
func TestCancelDownload_RefusedForIneligibleStatuses(t *testing.T) {
	tests := []struct {
		name   string
		status models.DownloadStatus
	}{
		{"completed", models.DownloadStatusCompleted},
		{"downloading", models.DownloadStatusDownloading},
		{"already cancelled", models.DownloadStatusCancelled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			server := NewServer()

			dl := models.DownloadInfo{Status: string(tt.status)}
			db.Create(&dl)

			req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/downloads/%d/cancel", dl.ID), nil)
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			if w.Code != http.StatusConflict {
				t.Fatalf("expected status 409, got %d: %s", w.Code, w.Body.String())
			}

			var after models.DownloadInfo
			db.First(&after, dl.ID)
			if after.Status != string(tt.status) {
				t.Errorf("expected status unchanged (%q), got %q", tt.status, after.Status)
			}
		})
	}
}

// 4.3: an eligible cancel sets DownloadInfo.Status/error_message and every
// linked ProcessedLine.State, and leaves a sibling occurrence of the same
// movie untouched.
func TestCancelDownload_SuccessCancelsOccurrenceAndLeavesSiblingUntouched(t *testing.T) {
	tests := []struct {
		name   string
		status models.DownloadStatus
	}{
		{"pending", models.DownloadStatusPending},
		{"failed", models.DownloadStatusFailed},
		{"retrying", models.DownloadStatusRetrying},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			server := NewServer()

			movie := models.Movie{TMDBID: 42, TMDBTitle: "Dune", TMDBYear: 2021}
			db.Create(&movie)

			dl := models.DownloadInfo{Status: string(tt.status)}
			db.Create(&dl)

			line := models.ProcessedLine{
				LineContent: "dune content",
				LineHash:    "hash-cancel-target-" + tt.name,
				TvgName:     "Dune",
				GroupTitle:  "Movies",
				ContentType: models.ContentTypeMovies,
				MovieID:     &movie.ID,
				State:       models.StateProcessed,
				ProcessedAt: time.Now(),
			}
			db.Create(&line)
			db.Model(&line).Update("download_info_id", dl.ID)

			// Sibling occurrence of the same movie, untouched by the cancel.
			siblingDL := models.DownloadInfo{Status: string(models.DownloadStatusCompleted)}
			db.Create(&siblingDL)
			sibling := models.ProcessedLine{
				LineContent: "dune sibling content",
				LineHash:    "hash-cancel-sibling-" + tt.name,
				TvgName:     "Dune",
				GroupTitle:  "Movies",
				ContentType: models.ContentTypeMovies,
				MovieID:     &movie.ID,
				State:       models.StateDownloaded,
				ProcessedAt: time.Now(),
			}
			db.Create(&sibling)
			db.Model(&sibling).Update("download_info_id", siblingDL.ID)

			req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/downloads/%d/cancel", dl.ID), nil)
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
			}

			var resp CancelDownloadResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}
			if resp.Status != string(models.DownloadStatusCancelled) {
				t.Errorf("expected response status %q, got %q", models.DownloadStatusCancelled, resp.Status)
			}

			var updatedDL models.DownloadInfo
			db.First(&updatedDL, dl.ID)
			if updatedDL.Status != string(models.DownloadStatusCancelled) {
				t.Errorf("expected DownloadInfo status cancelled, got %q", updatedDL.Status)
			}
			if updatedDL.ErrorMessage == nil || *updatedDL.ErrorMessage != cancelDownloadErrorMessage {
				t.Errorf("expected error_message %q, got %v", cancelDownloadErrorMessage, updatedDL.ErrorMessage)
			}

			var updatedLine models.ProcessedLine
			db.First(&updatedLine, line.ID)
			if updatedLine.State != models.StateCancelled {
				t.Errorf("expected ProcessedLine state cancelled, got %q", updatedLine.State)
			}

			// Sibling untouched.
			var siblingAfter models.DownloadInfo
			db.First(&siblingAfter, siblingDL.ID)
			if siblingAfter.Status != string(models.DownloadStatusCompleted) {
				t.Errorf("expected sibling DownloadInfo status unchanged, got %q", siblingAfter.Status)
			}
			var siblingLineAfter models.ProcessedLine
			db.First(&siblingLineAfter, sibling.ID)
			if siblingLineAfter.State != models.StateDownloaded {
				t.Errorf("expected sibling ProcessedLine state unchanged, got %q", siblingLineAfter.State)
			}
		})
	}
}

// 4.4: the route resolves (router-level check, distinct from any individual
// handler-behavior test above).
func TestCancelDownload_RouteResolves(t *testing.T) {
	db := setupTestDB(t)
	server := NewServer()

	dl := models.DownloadInfo{Status: string(models.DownloadStatusPending)}
	db.Create(&dl)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/downloads/%d/cancel", dl.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Fatalf("expected the cancel route to resolve, got 404 (route not registered?): %s", w.Body.String())
	}
}
