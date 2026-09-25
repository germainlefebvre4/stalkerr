package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/models"
)

func TestListDownloadsEnriched_DuplicateDownloadProblemFilter(t *testing.T) {
	db := setupTestDB(t)

	movieA := models.Movie{TMDBID: 1, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movieA)
	movieB := models.Movie{TMDBID: 2, TMDBTitle: "Arrival", TMDBYear: 2016}
	db.Create(&movieB)

	pathA1 := "/media/movies/Dune (2021)/Dune (2021).mkv"
	pathA2 := "/media/movies/Dune (2021) (1)/Dune (2021).mkv"
	dlA1 := models.DownloadInfo{URL: "http://example.com/dune-1", Status: "completed", DownloadPath: &pathA1, CreatedAt: time.Now(), UpdatedAt: time.Now(), CompletedAt: timePtr(time.Now())}
	db.Create(&dlA1)
	dlA2 := models.DownloadInfo{URL: "http://example.com/dune-2", Status: "completed", DownloadPath: &pathA2, CreatedAt: time.Now(), UpdatedAt: time.Now(), CompletedAt: timePtr(time.Now())}
	db.Create(&dlA2)

	plA1 := models.ProcessedLine{LineContent: "dune 1", LineHash: "hash-dune-1", TvgName: "Dune", ContentType: models.ContentTypeMovies, MovieID: &movieA.ID, DownloadInfoID: &dlA1.ID, State: models.StateDownloaded, ProcessedAt: time.Now()}
	db.Create(&plA1)
	plA2 := models.ProcessedLine{LineContent: "dune 2", LineHash: "hash-dune-2", TvgName: "Dune", ContentType: models.ContentTypeMovies, MovieID: &movieA.ID, DownloadInfoID: &dlA2.ID, State: models.StateDownloaded, ProcessedAt: time.Now()}
	db.Create(&plA2)

	pathB := "/media/movies/Arrival (2016)/Arrival (2016).mkv"
	dlB := models.DownloadInfo{URL: "http://example.com/arrival", Status: "completed", DownloadPath: &pathB, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	db.Create(&dlB)
	plB := models.ProcessedLine{LineContent: "arrival", LineHash: "hash-arrival", TvgName: "Arrival", ContentType: models.ContentTypeMovies, MovieID: &movieB.ID, DownloadInfoID: &dlB.ID, State: models.StateDownloaded, ProcessedAt: time.Now()}
	db.Create(&plB)

	server := NewServer()

	// 5.1: problem=duplicate_download returns only the duplicate group's occurrences.
	req, _ := http.NewRequest("GET", "/api/v1/downloads?problem=duplicate_download", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data  []DownloadEnrichedResponse `json:"data"`
		Total int64                      `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Total != 2 {
		t.Fatalf("expected 2 duplicate occurrences, got %d", resp.Total)
	}
	ids := map[uint]bool{}
	for _, item := range resp.Data {
		ids[item.ID] = true
	}
	if !ids[dlA1.ID] || !ids[dlA2.ID] {
		t.Errorf("expected both Dune occurrences in duplicate_download results, got %+v", ids)
	}
	if ids[dlB.ID] {
		t.Error("did not expect the single Arrival occurrence in duplicate_download results")
	}

	// 5.2: duplicate_siblings is populated for a duplicate row, omitted for a non-duplicate row.
	reqAll, _ := http.NewRequest("GET", "/api/v1/downloads", nil)
	wAll := httptest.NewRecorder()
	server.router.ServeHTTP(wAll, reqAll)

	var respAll struct {
		Data []DownloadEnrichedResponse `json:"data"`
	}
	if err := json.Unmarshal(wAll.Body.Bytes(), &respAll); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	var foundA1, foundB bool
	for _, item := range respAll.Data {
		if item.ID == dlA1.ID {
			foundA1 = true
			if len(item.DuplicateSiblings) != 1 {
				t.Fatalf("expected 1 sibling for dlA1, got %d", len(item.DuplicateSiblings))
			}
			if item.DuplicateSiblings[0].DownloadInfoID != dlA2.ID {
				t.Errorf("expected sibling id %d, got %d", dlA2.ID, item.DuplicateSiblings[0].DownloadInfoID)
			}
			if item.DuplicateSiblings[0].DownloadPath != pathA2 {
				t.Errorf("expected sibling path %q, got %q", pathA2, item.DuplicateSiblings[0].DownloadPath)
			}
		}
		if item.ID == dlB.ID {
			foundB = true
			if item.DuplicateSiblings != nil {
				t.Errorf("expected duplicate_siblings omitted for a non-duplicate row, got %+v", item.DuplicateSiblings)
			}
		}
	}
	if !foundA1 || !foundB {
		t.Fatal("expected both dlA1 and dlB in the default (unfiltered) response")
	}

	// 5.3: the raw JSON omits duplicate_siblings entirely for the non-duplicate row.
	raw := string(wAll.Body.Bytes())
	if !json.Valid([]byte(raw)) {
		t.Fatal("expected valid JSON response")
	}
}

// 5.3: the combined default filter (missing_year,year_mismatch,unknown_format,duplicate_download)
// includes a duplicate_download-only match alongside the other three reasons.
func TestListDownloadsEnriched_CombinedFilterIncludesDuplicateDownload(t *testing.T) {
	db := setupTestDB(t)

	movie := models.Movie{TMDBID: 1, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)

	// Both occurrences have a well-formed, year-correct path, so the only
	// applicable "problem" is duplicate_download.
	pathA := "/media/movies/Dune (2021)/Dune (2021).mkv"
	pathB := "/media/movies/Dune (2021) (1)/Dune (2021).mkv"
	dlA := models.DownloadInfo{URL: "http://example.com/dune-a", Status: "completed", DownloadPath: &pathA, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	db.Create(&dlA)
	dlB := models.DownloadInfo{URL: "http://example.com/dune-b", Status: "completed", DownloadPath: &pathB, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	db.Create(&dlB)

	plA := models.ProcessedLine{LineContent: "dune a", LineHash: "hash-a", TvgName: "Dune", ContentType: models.ContentTypeMovies, MovieID: &movie.ID, DownloadInfoID: &dlA.ID, State: models.StateDownloaded, ProcessedAt: time.Now()}
	db.Create(&plA)
	plB := models.ProcessedLine{LineContent: "dune b", LineHash: "hash-b", TvgName: "Dune", ContentType: models.ContentTypeMovies, MovieID: &movie.ID, DownloadInfoID: &dlB.ID, State: models.StateDownloaded, ProcessedAt: time.Now()}
	db.Create(&plB)

	server := NewServer()

	req, _ := http.NewRequest("GET", "/api/v1/downloads?problem=missing_year,year_mismatch,unknown_format,duplicate_download", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data  []DownloadEnrichedResponse `json:"data"`
		Total int64                      `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Total != 2 {
		t.Fatalf("expected both duplicate occurrences to match the combined filter, got %d", resp.Total)
	}
}

func timePtr(t time.Time) *time.Time { return &t }
