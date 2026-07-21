package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glefebvre/stalkeer/internal/api"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/testutil"
)

func TestResetHandlers(t *testing.T) {
	// Initialize test database
	db := testutil.TestDB(t)
	defer testutil.CleanupDB(t, db)
	database.SetDB(db)

	// Create test movie and tvshow
	movie := testutil.CreateMovie(db)
	testutil.CreateProcessedLine(db, func(pl *models.ProcessedLine) {
		pl.MovieID = &movie.ID
		pl.LineHash = "movie-line-hash"
	})

	tvshow := testutil.CreateTVShow(db)
	testutil.CreateProcessedLine(db, func(pl *models.ProcessedLine) {
		pl.TVShowID = &tvshow.ID
		pl.LineHash = "tvshow-line-hash"
	})

	// Create the API server
	server := api.NewServer()

	// 1. Test POST /api/v1/movies/:id/reset - Success
	req1, _ := http.NewRequest("POST", "/api/v1/movies/1/reset", nil)
	w1 := httptest.NewRecorder()
	server.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", w1.Code)
	}

	var resp1 map[string]interface{}
	if err := json.Unmarshal(w1.Body.Bytes(), &resp1); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp1["status"] != "success" {
		t.Errorf("expected success status, got %s", resp1["status"])
	}

	// Verify that the processed line is gone
	var movieLineCount int64
	db.Model(&models.ProcessedLine{}).Where("movie_id = ?", movie.ID).Count(&movieLineCount)
	if movieLineCount != 0 {
		t.Errorf("expected 0 processed lines for movie after reset, got %d", movieLineCount)
	}

	// 2. Test POST /api/v1/movies/:id/reset - Not Found
	req2, _ := http.NewRequest("POST", "/api/v1/movies/999/reset", nil)
	w2 := httptest.NewRecorder()
	server.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNotFound {
		t.Errorf("expected status 404 Not Found, got %d", w2.Code)
	}

	// 3. Test POST /api/v1/tvshows/:id/reset - Success
	req3, _ := http.NewRequest("POST", "/api/v1/tvshows/1/reset", nil)
	w3 := httptest.NewRecorder()
	server.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", w3.Code)
	}

	var resp3 map[string]interface{}
	if err := json.Unmarshal(w3.Body.Bytes(), &resp3); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp3["status"] != "success" {
		t.Errorf("expected success status, got %s", resp3["status"])
	}

	// Verify that the processed line is gone
	var tvshowLineCount int64
	db.Model(&models.ProcessedLine{}).Where("tv_show_id = ?", tvshow.ID).Count(&tvshowLineCount)
	if tvshowLineCount != 0 {
		t.Errorf("expected 0 processed lines for tvshow after reset, got %d", tvshowLineCount)
	}

	// 4. Test POST /api/v1/tvshows/:id/reset - Not Found
	req4, _ := http.NewRequest("POST", "/api/v1/tvshows/999/reset", nil)
	w4 := httptest.NewRecorder()
	server.ServeHTTP(w4, req4)

	if w4.Code != http.StatusNotFound {
		t.Errorf("expected status 404 Not Found, got %d", w4.Code)
	}
}
