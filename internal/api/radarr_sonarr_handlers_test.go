package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/glefebvre/stalkeer/internal/external/radarr"
	"github.com/glefebvre/stalkeer/internal/external/sonarr"
	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/gorm"
)

// countMovieQueryVars registers a callback that tracks the largest number of bound
// SQL variables seen in any query against the "movies" table, so a test can assert
// a query only ever touched a page's worth of movies, not the full catalog.
func countMovieQueryVars(db *gorm.DB) *int {
	max := 0
	db.Callback().Query().After("gorm:query").Register("test:count_movie_query_vars", func(tx *gorm.DB) {
		if tx.Statement.Table == "movies" {
			if n := len(tx.Statement.Vars); n > max {
				max = n
			}
		}
	})
	return &max
}

func TestListRadarrMonitoredMovies_NotConfigured(t *testing.T) {
	setupTestDB(t)
	newForceDownloadTestConfig(t, "", "")
	server := NewServer()

	req, _ := http.NewRequest("GET", "/api/v1/radarr/movies", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", w.Code, w.Body.String())
	}
	var body ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body.Error != "radarr_not_configured" {
		t.Errorf("expected error code radarr_not_configured, got %q", body.Error)
	}
}

func TestRadarrFailureDoesNotAffectSonarrEndpoint(t *testing.T) {
	setupTestDB(t)

	sonarrSeries := []sonarr.Series{
		{ID: 1, Title: "Reachable Series", TvdbID: 555, Monitored: true},
	}
	sonarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v3/series":
			json.NewEncoder(w).Encode(sonarrSeries)
		case r.URL.Path == "/api/v3/episode":
			json.NewEncoder(w).Encode([]sonarr.Episode{})
		default:
			t.Errorf("unexpected sonarr path %s", r.URL.Path)
		}
	}))
	defer sonarrServer.Close()

	// Radarr points at an address nothing listens on, simulating an unreachable
	// upstream, while Sonarr is fully reachable and configured.
	newForceDownloadTestConfig(t, "http://127.0.0.1:1", sonarrServer.URL)
	server := NewServer()

	radarrReq, _ := http.NewRequest("GET", "/api/v1/radarr/movies", nil)
	radarrW := httptest.NewRecorder()
	server.router.ServeHTTP(radarrW, radarrReq)
	if radarrW.Code != http.StatusBadGateway {
		t.Fatalf("expected radarr endpoint to report 502 upstream failure, got %d: %s", radarrW.Code, radarrW.Body.String())
	}
	var radarrErr ErrorResponse
	if err := json.Unmarshal(radarrW.Body.Bytes(), &radarrErr); err != nil {
		t.Fatalf("failed to decode radarr error body: %v", err)
	}
	if radarrErr.Error != "radarr_unreachable" {
		t.Errorf("expected radarr_unreachable, got %q", radarrErr.Error)
	}

	sonarrReq, _ := http.NewRequest("GET", "/api/v1/sonarr/series", nil)
	sonarrW := httptest.NewRecorder()
	server.router.ServeHTTP(sonarrW, sonarrReq)
	if sonarrW.Code != http.StatusOK {
		t.Fatalf("expected sonarr endpoint to remain unaffected by radarr failure, got %d: %s", sonarrW.Code, sonarrW.Body.String())
	}

	var sonarrResp struct {
		Data []SonarrSeriesListItem `json:"data"`
	}
	if err := json.Unmarshal(sonarrW.Body.Bytes(), &sonarrResp); err != nil {
		t.Fatalf("failed to decode sonarr response: %v", err)
	}
	if len(sonarrResp.Data) != 1 || sonarrResp.Data[0].Title != "Reachable Series" {
		t.Errorf("expected sonarr series list unaffected, got %+v", sonarrResp.Data)
	}
}

func TestListRadarrMonitoredMovies_PaginatesBeforeMatching(t *testing.T) {
	db := setupTestDB(t)

	// Seed a local movie matching every catalog entry's TVDB id, so that if the
	// handler matched against the *entire* catalog instead of just the requested
	// page, the batched query would bind far more than a page's worth of variables.
	const catalogSize = 200
	for i := 1; i <= catalogSize; i++ {
		tvdbID := 10000 + i
		movie := models.Movie{
			TMDBID:    90000 + i,
			TVDBID:    &tvdbID,
			TMDBTitle: fmt.Sprintf("Movie %d", i),
			TMDBYear:  2000 + (i % 20),
		}
		if err := db.Create(&movie).Error; err != nil {
			t.Fatalf("failed to seed local movie: %v", err)
		}
	}

	radarrMovies := make([]radarr.Movie, catalogSize)
	for i := 0; i < catalogSize; i++ {
		radarrMovies[i] = radarr.Movie{
			ID:        i + 1,
			Title:     fmt.Sprintf("Movie %d", i+1),
			Year:      2000 + (i % 20),
			TvdbID:    10000 + i + 1,
			Monitored: true,
		}
	}

	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(radarrMovies)
	}))
	defer radarrServer.Close()

	newForceDownloadTestConfig(t, radarrServer.URL, "")
	server := NewServer()

	maxVars := countMovieQueryVars(db)

	const pageSize = 5
	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/radarr/movies?limit=%d&offset=0", pageSize), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data  []RadarrMovieListItem `json:"data"`
		Total int64                 `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if int(resp.Total) != catalogSize {
		t.Errorf("expected total %d, got %d", catalogSize, resp.Total)
	}
	if len(resp.Data) != pageSize {
		t.Fatalf("expected %d items on the page, got %d", pageSize, len(resp.Data))
	}
	for _, item := range resp.Data {
		if !item.Matched {
			t.Errorf("expected movie %d to be matched (seeded local movie exists), got unmatched", item.RadarrID)
		}
	}

	// A batched query for the page should bind roughly pageSize variables, never
	// anywhere close to the full 200-item catalog.
	if *maxVars > pageSize*4 {
		t.Errorf("expected DB match query bound variables to scale with page size (~%d), got %d - matching does not appear to be limited to the requested page", pageSize, *maxVars)
	}
}

func TestListRadarrMonitoredMovies_Search(t *testing.T) {
	setupTestDB(t)

	radarrMovies := []radarr.Movie{
		{ID: 1, Title: "The Matrix", Year: 1999, Monitored: true},
		{ID: 2, Title: "Matrix Reloaded", Year: 2003, Monitored: true},
		{ID: 3, Title: "Inception", Year: 2010, Monitored: true},
	}
	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(radarrMovies)
	}))
	defer radarrServer.Close()

	newForceDownloadTestConfig(t, radarrServer.URL, "")
	server := NewServer()

	cases := []struct {
		name          string
		search        string
		expectedTotal int
		expectTitles  []string
	}{
		{"matching subset", "matrix", 2, []string{"The Matrix", "Matrix Reloaded"}},
		{"matching nothing", "no-such-title", 0, nil},
		{"omitted", "", 3, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/api/v1/radarr/movies"
			if tc.search != "" {
				url += "?search=" + tc.search
			}
			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
			}

			var resp struct {
				Data  []RadarrMovieListItem `json:"data"`
				Total int64                 `json:"total"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if int(resp.Total) != tc.expectedTotal {
				t.Errorf("expected total %d, got %d", tc.expectedTotal, resp.Total)
			}
			if len(resp.Data) != tc.expectedTotal {
				t.Fatalf("expected %d items, got %d", tc.expectedTotal, len(resp.Data))
			}
			for _, expected := range tc.expectTitles {
				found := false
				for _, item := range resp.Data {
					if item.Title == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected title %q in results, got %+v", expected, resp.Data)
				}
			}
		})
	}
}

func TestListSonarrMonitoredSeries_Search(t *testing.T) {
	setupTestDB(t)

	allSeries := []sonarr.Series{
		{ID: 1, Title: "Breaking Bad", TvdbID: 1001, Monitored: true},
		{ID: 2, Title: "Better Call Saul", TvdbID: 1002, Monitored: true},
		{ID: 3, Title: "The Wire", TvdbID: 1003, Monitored: true},
	}
	sonarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v3/series":
			json.NewEncoder(w).Encode(allSeries)
		case strings.HasPrefix(r.URL.Path, "/api/v3/episode"):
			json.NewEncoder(w).Encode([]sonarr.Episode{})
		default:
			t.Errorf("unexpected sonarr path %s", r.URL.Path)
		}
	}))
	defer sonarrServer.Close()

	newForceDownloadTestConfig(t, "", sonarrServer.URL)
	server := NewServer()

	cases := []struct {
		name          string
		search        string
		expectedTotal int
		expectTitles  []string
	}{
		{"matching subset", "bad", 1, []string{"Breaking Bad"}},
		{"matching nothing", "no-such-title", 0, nil},
		{"omitted", "", 3, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/api/v1/sonarr/series"
			if tc.search != "" {
				url += "?search=" + tc.search
			}
			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
			}

			var resp struct {
				Data  []SonarrSeriesListItem `json:"data"`
				Total int64                  `json:"total"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if int(resp.Total) != tc.expectedTotal {
				t.Errorf("expected total %d, got %d", tc.expectedTotal, resp.Total)
			}
			if len(resp.Data) != tc.expectedTotal {
				t.Fatalf("expected %d items, got %d", tc.expectedTotal, len(resp.Data))
			}
			for _, expected := range tc.expectTitles {
				found := false
				for _, item := range resp.Data {
					if item.Title == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected title %q in results, got %+v", expected, resp.Data)
				}
			}
		})
	}
}

func TestListRadarrSonarrStats_FullCatalogCounts(t *testing.T) {
	db := setupTestDB(t)

	const catalogSize = 45 // larger than one page (radarrSonarrDefaultPageSize=20)
	const matchedCount = 30

	// Matched movies use a distinct year (2000) with matching TVDB ids seeded
	// locally; unmatched movies use a different year (1975) with no local
	// counterpart at all, so fuzzy title+year matching can't accidentally cross
	// the two groups.
	for i := 1; i <= matchedCount; i++ {
		tvdbID := 10000 + i
		movie := models.Movie{
			TMDBID:    90000 + i,
			TVDBID:    &tvdbID,
			TMDBTitle: fmt.Sprintf("Alpha Movie %d", i),
			TMDBYear:  2000,
		}
		if err := db.Create(&movie).Error; err != nil {
			t.Fatalf("failed to seed local movie: %v", err)
		}
	}

	radarrMovies := make([]radarr.Movie, catalogSize)
	for i := 0; i < matchedCount; i++ {
		radarrMovies[i] = radarr.Movie{
			ID:        i + 1,
			Title:     fmt.Sprintf("Alpha Movie %d", i+1),
			Year:      2000,
			TvdbID:    10000 + i + 1,
			Monitored: true,
		}
	}
	for i := matchedCount; i < catalogSize; i++ {
		radarrMovies[i] = radarr.Movie{
			ID:        i + 1,
			Title:     fmt.Sprintf("Beta Movie %d", i+1),
			Year:      1975,
			TvdbID:    20000 + i + 1,
			Monitored: true,
		}
	}

	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(radarrMovies)
	}))
	defer radarrServer.Close()

	var episodeCalls int64
	sonarrSeries := make([]sonarr.Series, 7)
	for i := range sonarrSeries {
		sonarrSeries[i] = sonarr.Series{ID: i + 1, Title: fmt.Sprintf("Series %d", i+1), TvdbID: 5000 + i, Monitored: true}
	}
	sonarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v3/series":
			json.NewEncoder(w).Encode(sonarrSeries)
		case strings.HasPrefix(r.URL.Path, "/api/v3/episode"):
			atomic.AddInt64(&episodeCalls, 1)
			json.NewEncoder(w).Encode([]sonarr.Episode{})
		default:
			t.Errorf("unexpected sonarr path %s", r.URL.Path)
		}
	}))
	defer sonarrServer.Close()

	newForceDownloadTestConfig(t, radarrServer.URL, sonarrServer.URL)
	server := NewServer()

	req, _ := http.NewRequest("GET", "/api/v1/radarr-sonarr/stats", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp RadarrSonarrStatsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.RadarrMonitored == nil || *resp.RadarrMonitored != catalogSize {
		t.Errorf("expected radarr_monitored %d, got %v", catalogSize, resp.RadarrMonitored)
	}
	if resp.RadarrMatched == nil {
		t.Fatalf("expected radarr_matched to be set, got nil")
	}
	if *resp.RadarrMatched != matchedCount {
		t.Errorf("expected radarr_matched %d, got %d", matchedCount, *resp.RadarrMatched)
	}
	if resp.RadarrError != "" {
		t.Errorf("expected no radarr error, got %q", resp.RadarrError)
	}
	if resp.SonarrMonitored == nil || *resp.SonarrMonitored != len(sonarrSeries) {
		t.Errorf("expected sonarr_monitored %d, got %v", len(sonarrSeries), resp.SonarrMonitored)
	}
	if resp.SonarrError != "" {
		t.Errorf("expected no sonarr error, got %q", resp.SonarrError)
	}

	if calls := atomic.LoadInt64(&episodeCalls); calls != 0 {
		t.Errorf("expected no per-series episode fetches when computing sonarr_monitored, got %d", calls)
	}
}

func TestListRadarrSonarrStats_RadarrUnreachableSonarrStillReturned(t *testing.T) {
	setupTestDB(t)

	sonarrSeries := []sonarr.Series{
		{ID: 1, Title: "Reachable Series", TvdbID: 555, Monitored: true},
	}
	sonarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v3/series":
			json.NewEncoder(w).Encode(sonarrSeries)
		case strings.HasPrefix(r.URL.Path, "/api/v3/episode"):
			json.NewEncoder(w).Encode([]sonarr.Episode{})
		default:
			t.Errorf("unexpected sonarr path %s", r.URL.Path)
		}
	}))
	defer sonarrServer.Close()

	newForceDownloadTestConfig(t, "http://127.0.0.1:1", sonarrServer.URL)
	server := NewServer()

	req, _ := http.NewRequest("GET", "/api/v1/radarr-sonarr/stats", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp RadarrSonarrStatsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.RadarrError == "" {
		t.Errorf("expected a radarr error, got none")
	}
	if resp.RadarrMonitored != nil || resp.RadarrMatched != nil {
		t.Errorf("expected radarr counts to be nil on failure, got monitored=%v matched=%v", resp.RadarrMonitored, resp.RadarrMatched)
	}
	if resp.SonarrMonitored == nil || *resp.SonarrMonitored != 1 {
		t.Errorf("expected sonarr_monitored 1 despite radarr failure, got %v", resp.SonarrMonitored)
	}
	if resp.SonarrError != "" {
		t.Errorf("expected no sonarr error, got %q", resp.SonarrError)
	}
}

func TestListRadarrSonarrStats_SonarrUnreachableRadarrStillReturned(t *testing.T) {
	db := setupTestDB(t)

	tvdbID := 20000
	movie := models.Movie{TMDBID: 99999, TVDBID: &tvdbID, TMDBTitle: "Solo Movie", TMDBYear: 2020}
	if err := db.Create(&movie).Error; err != nil {
		t.Fatalf("failed to seed local movie: %v", err)
	}

	radarrMovies := []radarr.Movie{
		{ID: 1, Title: "Solo Movie", Year: 2020, TvdbID: 20000, Monitored: true},
	}
	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(radarrMovies)
	}))
	defer radarrServer.Close()

	newForceDownloadTestConfig(t, radarrServer.URL, "http://127.0.0.1:1")
	server := NewServer()

	req, _ := http.NewRequest("GET", "/api/v1/radarr-sonarr/stats", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp RadarrSonarrStatsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.RadarrMonitored == nil || *resp.RadarrMonitored != 1 {
		t.Errorf("expected radarr_monitored 1 despite sonarr failure, got %v", resp.RadarrMonitored)
	}
	if resp.RadarrMatched == nil || *resp.RadarrMatched != 1 {
		t.Errorf("expected radarr_matched 1, got %v", resp.RadarrMatched)
	}
	if resp.RadarrError != "" {
		t.Errorf("expected no radarr error, got %q", resp.RadarrError)
	}
	if resp.SonarrError == "" {
		t.Errorf("expected a sonarr error, got none")
	}
	if resp.SonarrMonitored != nil {
		t.Errorf("expected sonarr_monitored to be nil on failure, got %v", resp.SonarrMonitored)
	}
}

func TestListSonarrMonitoredSeries_PageSizeBoundsEpisodeFetches(t *testing.T) {
	setupTestDB(t)

	const catalogSize = 20
	const pageSize = 3

	allSeries := make([]sonarr.Series, catalogSize)
	for i := 0; i < catalogSize; i++ {
		allSeries[i] = sonarr.Series{
			ID:        i + 1,
			Title:     fmt.Sprintf("Series %d", i+1),
			TvdbID:    5000 + i + 1,
			Monitored: true,
		}
	}

	var episodeCalls int64
	sonarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v3/series":
			json.NewEncoder(w).Encode(allSeries)
		case strings.HasPrefix(r.URL.Path, "/api/v3/episode"):
			atomic.AddInt64(&episodeCalls, 1)
			json.NewEncoder(w).Encode([]sonarr.Episode{})
		default:
			t.Errorf("unexpected sonarr path %s", r.URL.Path)
		}
	}))
	defer sonarrServer.Close()

	newForceDownloadTestConfig(t, "", sonarrServer.URL)
	server := NewServer()

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/sonarr/series?limit=%d&offset=0", pageSize), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data  []SonarrSeriesListItem `json:"data"`
		Total int64                  `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if int(resp.Total) != catalogSize {
		t.Errorf("expected total %d, got %d", catalogSize, resp.Total)
	}
	if len(resp.Data) != pageSize {
		t.Fatalf("expected %d items on the page, got %d", pageSize, len(resp.Data))
	}

	calls := atomic.LoadInt64(&episodeCalls)
	if calls != pageSize {
		t.Errorf("expected exactly %d episode fetches (bounded by page size), got %d (catalog has %d series)", pageSize, calls, catalogSize)
	}
}
