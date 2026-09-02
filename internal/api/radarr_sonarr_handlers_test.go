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

func TestListRadarrMonitoredMovies_MatchStatusFilter(t *testing.T) {
	db := setupTestDB(t)

	// 3 matched (seeded local movie via TVDB id), 2 unmatched.
	matchedIDs := []int{1, 2, 3}
	for _, i := range matchedIDs {
		tvdbID := 10000 + i
		movie := models.Movie{TMDBID: 90000 + i, TVDBID: &tvdbID, TMDBTitle: fmt.Sprintf("Matched %d", i), TMDBYear: 2000}
		if err := db.Create(&movie).Error; err != nil {
			t.Fatalf("failed to seed local movie: %v", err)
		}
	}

	radarrMovies := []radarr.Movie{
		{ID: 1, Title: "Matched 1", Year: 2000, TvdbID: 10001, Monitored: true},
		{ID: 2, Title: "Matched 2", Year: 2000, TvdbID: 10002, Monitored: true},
		{ID: 3, Title: "Matched 3", Year: 2000, TvdbID: 10003, Monitored: true},
		{ID: 4, Title: "Unmatched 1", Year: 1975, TvdbID: 20004, Monitored: true},
		{ID: 5, Title: "Unmatched 2", Year: 1975, TvdbID: 20005, Monitored: true},
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
		filter        string
		expectedTotal int
	}{
		{"matched", "matched", 3},
		{"no_match", "no_match", 2},
		{"no filter", "", 5},
		{"invalid filter treated as none", "bogus", 5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/api/v1/radarr/movies"
			if tc.filter != "" {
				url += "?filter=" + tc.filter
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
			for _, item := range resp.Data {
				expectMatched := tc.filter == "matched"
				if tc.filter == "matched" || tc.filter == "no_match" {
					if item.Matched != expectMatched {
						t.Errorf("expected item %q matched=%v for filter %q, got %v", item.Title, expectMatched, tc.filter, item.Matched)
					}
				}
			}
		})
	}
}

func TestListRadarrMonitoredMovies_FilteredTotalReflectsFilteredCount(t *testing.T) {
	db := setupTestDB(t)

	// 10 movies total, spanning multiple pages pre-filter (pageSize=4), but only
	// 3 matched, so a filtered request should report far fewer pages than the
	// unfiltered catalog would.
	const catalogSize = 10
	const matchedCount = 3
	const pageSize = 4

	for i := 1; i <= matchedCount; i++ {
		tvdbID := 10000 + i
		movie := models.Movie{TMDBID: 90000 + i, TVDBID: &tvdbID, TMDBTitle: fmt.Sprintf("Matched %d", i), TMDBYear: 2000}
		if err := db.Create(&movie).Error; err != nil {
			t.Fatalf("failed to seed local movie: %v", err)
		}
	}

	radarrMovies := make([]radarr.Movie, catalogSize)
	for i := 0; i < matchedCount; i++ {
		radarrMovies[i] = radarr.Movie{ID: i + 1, Title: fmt.Sprintf("Matched %d", i+1), Year: 2000, TvdbID: 10000 + i + 1, Monitored: true}
	}
	for i := matchedCount; i < catalogSize; i++ {
		radarrMovies[i] = radarr.Movie{ID: i + 1, Title: fmt.Sprintf("Unmatched %d", i+1), Year: 1975, TvdbID: 20000 + i + 1, Monitored: true}
	}

	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(radarrMovies)
	}))
	defer radarrServer.Close()

	newForceDownloadTestConfig(t, radarrServer.URL, "")
	server := NewServer()

	req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/radarr/movies?filter=matched&limit=%d&offset=0", pageSize), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if int(resp.Total) != matchedCount {
		t.Errorf("expected filtered total %d, got %d", matchedCount, resp.Total)
	}
	if resp.TotalPages != 1 {
		t.Errorf("expected 1 page for %d filtered items at page size %d, got %d", matchedCount, pageSize, resp.TotalPages)
	}
}

func TestListRadarrMonitoredMovies_FilterCombinedWithSearch(t *testing.T) {
	db := setupTestDB(t)

	tvdbID := 10001
	movie := models.Movie{TMDBID: 90001, TVDBID: &tvdbID, TMDBTitle: "Alpha Matched", TMDBYear: 2000}
	if err := db.Create(&movie).Error; err != nil {
		t.Fatalf("failed to seed local movie: %v", err)
	}

	radarrMovies := []radarr.Movie{
		{ID: 1, Title: "Alpha Matched", Year: 2000, TvdbID: 10001, Monitored: true},
		{ID: 2, Title: "Alpha Unmatched", Year: 1975, TvdbID: 20002, Monitored: true},
		{ID: 3, Title: "Beta Unmatched", Year: 1975, TvdbID: 20003, Monitored: true},
	}
	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(radarrMovies)
	}))
	defer radarrServer.Close()

	newForceDownloadTestConfig(t, radarrServer.URL, "")
	server := NewServer()

	req, _ := http.NewRequest("GET", "/api/v1/radarr/movies?search=alpha&filter=no_match", nil)
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

	if resp.Total != 1 || len(resp.Data) != 1 || resp.Data[0].Title != "Alpha Unmatched" {
		t.Fatalf("expected exactly 'Alpha Unmatched' for search=alpha&filter=no_match, got %+v (total %d)", resp.Data, resp.Total)
	}
}

func TestListRadarrMonitoredMovies_OccurrenceCount(t *testing.T) {
	db := setupTestDB(t)

	tvdbID := 10001
	movie := models.Movie{TMDBID: 90001, TVDBID: &tvdbID, TMDBTitle: "Matched Movie", TMDBYear: 2000}
	if err := db.Create(&movie).Error; err != nil {
		t.Fatalf("failed to seed local movie: %v", err)
	}

	lineURL := "http://example.com/stream.mkv"
	res720p, res1080p := "720p", "1080p"
	occurrences := []models.ProcessedLine{
		{MovieID: &movie.ID, TvgName: "Matched Movie 720p", LineURL: &lineURL, LineContent: "#EXTINF", LineHash: "hash-1", GroupTitle: "Movies", ContentType: models.ContentTypeMovies, State: models.StateProcessed, Resolution: &res720p},
		{MovieID: &movie.ID, TvgName: "Matched Movie 1080p", LineURL: &lineURL, LineContent: "#EXTINF", LineHash: "hash-2", GroupTitle: "Movies", ContentType: models.ContentTypeMovies, State: models.StateDownloaded, Resolution: &res1080p},
	}
	for i := range occurrences {
		if err := db.Create(&occurrences[i]).Error; err != nil {
			t.Fatalf("failed to seed processed line: %v", err)
		}
	}

	radarrMovies := []radarr.Movie{
		{ID: 1, Title: "Matched Movie", Year: 2000, TvdbID: 10001, Monitored: true},
		{ID: 2, Title: "Unmatched Movie", Year: 1975, TvdbID: 20002, Monitored: true},
	}
	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(radarrMovies)
	}))
	defer radarrServer.Close()

	newForceDownloadTestConfig(t, radarrServer.URL, "")
	server := NewServer()

	req, _ := http.NewRequest("GET", "/api/v1/radarr/movies", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data []RadarrMovieListItem `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	for _, item := range resp.Data {
		switch item.Title {
		case "Matched Movie":
			if item.OccurrenceCount != 2 {
				t.Errorf("expected 2 occurrences for matched movie, got %d", item.OccurrenceCount)
			}
		case "Unmatched Movie":
			if item.OccurrenceCount != 0 {
				t.Errorf("expected 0 occurrences for unmatched movie, got %d", item.OccurrenceCount)
			}
		default:
			t.Errorf("unexpected item %q", item.Title)
		}
	}
}

// sonarrEpisodeFixtures maps a Sonarr series ID to the episodes GetEpisodesBySeriesID
// should return for it, letting a test server behave differently per series.
func sonarrTestServer(t *testing.T, allSeries []sonarr.Series, episodesBySeriesID map[int][]sonarr.Episode) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v3/series":
			json.NewEncoder(w).Encode(allSeries)
		case strings.HasPrefix(r.URL.Path, "/api/v3/episode"):
			seriesID := 0
			fmt.Sscanf(r.URL.Query().Get("seriesId"), "%d", &seriesID)
			json.NewEncoder(w).Encode(episodesBySeriesID[seriesID])
		default:
			t.Errorf("unexpected sonarr path %s", r.URL.Path)
		}
	}))
}

func TestListSonarrMonitoredSeries_MatchStatusFilter(t *testing.T) {
	db := setupTestDB(t)
	sonarrMatchCache.clear()

	// Series 1: has a matched monitored episode. Series 2 and 3: no local match.
	s1e1, e1 := 1, 1
	tvdbMatched := 5001
	if err := db.Create(&models.TVShow{TMDBID: 9999, TVDBID: &tvdbMatched, TMDBTitle: "Matched Show", Season: &s1e1, Episode: &e1}).Error; err != nil {
		t.Fatalf("failed to seed local episode: %v", err)
	}

	allSeries := []sonarr.Series{
		{ID: 1, Title: "Matched Show", TvdbID: tvdbMatched, Monitored: true},
		{ID: 2, Title: "Unmatched Show A", TvdbID: 5002, Monitored: true},
		{ID: 3, Title: "Unmatched Show B", TvdbID: 5003, Monitored: true},
	}
	episodesBySeriesID := map[int][]sonarr.Episode{
		1: {{SeasonNumber: 1, EpisodeNumber: 1, Monitored: true}},
		2: {{SeasonNumber: 1, EpisodeNumber: 1, Monitored: true}},
		3: {{SeasonNumber: 1, EpisodeNumber: 1, Monitored: true}},
	}
	sonarrServer := sonarrTestServer(t, allSeries, episodesBySeriesID)
	defer sonarrServer.Close()

	newForceDownloadTestConfig(t, "", sonarrServer.URL)
	server := NewServer()

	cases := []struct {
		name          string
		filter        string
		expectedTotal int
	}{
		{"matched", "matched", 1},
		{"no_match", "no_match", 2},
		{"no filter", "", 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			url := "/api/v1/sonarr/series"
			if tc.filter != "" {
				url += "?filter=" + tc.filter
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
		})
	}
}

func TestListSonarrMonitoredSeries_FilteredTotalAndSearch(t *testing.T) {
	db := setupTestDB(t)
	sonarrMatchCache.clear()

	s1e1, e1 := 1, 1
	tvdbMatched := 5001
	if err := db.Create(&models.TVShow{TMDBID: 9999, TVDBID: &tvdbMatched, TMDBTitle: "Alpha Matched", Season: &s1e1, Episode: &e1}).Error; err != nil {
		t.Fatalf("failed to seed local episode: %v", err)
	}

	allSeries := []sonarr.Series{
		{ID: 1, Title: "Alpha Matched", TvdbID: tvdbMatched, Monitored: true},
		{ID: 2, Title: "Alpha Unmatched", TvdbID: 5002, Monitored: true},
		{ID: 3, Title: "Beta Unmatched", TvdbID: 5003, Monitored: true},
	}
	episodesBySeriesID := map[int][]sonarr.Episode{
		1: {{SeasonNumber: 1, EpisodeNumber: 1, Monitored: true}},
		2: {{SeasonNumber: 1, EpisodeNumber: 1, Monitored: true}},
		3: {{SeasonNumber: 1, EpisodeNumber: 1, Monitored: true}},
	}
	sonarrServer := sonarrTestServer(t, allSeries, episodesBySeriesID)
	defer sonarrServer.Close()

	newForceDownloadTestConfig(t, "", sonarrServer.URL)
	server := NewServer()

	req, _ := http.NewRequest("GET", "/api/v1/sonarr/series?search=alpha&filter=no_match", nil)
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

	if resp.Total != 1 || len(resp.Data) != 1 || resp.Data[0].Title != "Alpha Unmatched" {
		t.Fatalf("expected exactly 'Alpha Unmatched' for search=alpha&filter=no_match, got %+v (total %d)", resp.Data, resp.Total)
	}
}

func TestListSonarrMonitoredSeries_RefreshInvalidatesCache(t *testing.T) {
	setupTestDB(t)
	sonarrMatchCache.clear()

	tvdbID := 5001
	allSeries := []sonarr.Series{
		{ID: 1, Title: "Series One", TvdbID: tvdbID, Monitored: true},
	}
	var episodeCalls int64
	sonarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/api/v3/series":
			json.NewEncoder(w).Encode(allSeries)
		case strings.HasPrefix(r.URL.Path, "/api/v3/episode"):
			atomic.AddInt64(&episodeCalls, 1)
			json.NewEncoder(w).Encode([]sonarr.Episode{{SeasonNumber: 1, EpisodeNumber: 1, Monitored: true}})
		default:
			t.Errorf("unexpected sonarr path %s", r.URL.Path)
		}
	}))
	defer sonarrServer.Close()

	newForceDownloadTestConfig(t, "", sonarrServer.URL)
	server := NewServer()

	// First filtered request populates the cache: one episode fetch for the
	// filter-membership check via the cache, plus one for the page-scoped detail.
	req, _ := http.NewRequest("GET", "/api/v1/sonarr/series?filter=matched", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	callsAfterFirst := atomic.LoadInt64(&episodeCalls)

	// A second filtered request should reuse the cache for the filter-membership
	// check, only paying the page-scoped detail fetch again.
	req2, _ := http.NewRequest("GET", "/api/v1/sonarr/series?filter=matched", nil)
	w2 := httptest.NewRecorder()
	server.router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w2.Code, w2.Body.String())
	}
	callsAfterSecond := atomic.LoadInt64(&episodeCalls)
	if callsAfterSecond-callsAfterFirst >= callsAfterFirst {
		t.Errorf("expected fewer episode fetches on cache hit; first=%d, second increment=%d", callsAfterFirst, callsAfterSecond-callsAfterFirst)
	}

	// A refresh request must invalidate the cache, causing the filter-membership
	// check to refetch for the series again.
	req3, _ := http.NewRequest("GET", "/api/v1/sonarr/series?filter=matched&refresh=true", nil)
	w3 := httptest.NewRecorder()
	server.router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w3.Code, w3.Body.String())
	}
	callsAfterRefresh := atomic.LoadInt64(&episodeCalls)
	if callsAfterRefresh-callsAfterSecond < callsAfterFirst {
		t.Errorf("expected refresh to force a recompute (episode fetch count to jump back up), got increment %d (first request cost %d)", callsAfterRefresh-callsAfterSecond, callsAfterFirst)
	}
}

func TestListSonarrMonitoredSeries_OccurrenceCount(t *testing.T) {
	db := setupTestDB(t)

	// Two monitored episodes for the series, each with its own occurrence count.
	s1e1, e1 := 1, 1
	s1e2, e2 := 1, 2
	tvdbID := 5001
	ep1 := models.TVShow{TMDBID: 1001, TVDBID: &tvdbID, TMDBTitle: "Series One", Season: &s1e1, Episode: &e1}
	ep2 := models.TVShow{TMDBID: 1002, TVDBID: &tvdbID, TMDBTitle: "Series One", Season: &s1e2, Episode: &e2}
	if err := db.Create(&ep1).Error; err != nil {
		t.Fatalf("failed to seed episode 1: %v", err)
	}
	if err := db.Create(&ep2).Error; err != nil {
		t.Fatalf("failed to seed episode 2: %v", err)
	}

	lineURL := "http://example.com/stream.mkv"
	res720p, res1080p := "720p", "1080p"
	lines := []models.ProcessedLine{
		{TVShowID: &ep1.ID, TvgName: "Series One S01E01 720p", LineURL: &lineURL, LineContent: "#EXTINF", LineHash: "hash-1", GroupTitle: "Series", ContentType: models.ContentTypeTVShows, State: models.StateProcessed, Resolution: &res720p},
		{TVShowID: &ep1.ID, TvgName: "Series One S01E01 1080p", LineURL: &lineURL, LineContent: "#EXTINF", LineHash: "hash-2", GroupTitle: "Series", ContentType: models.ContentTypeTVShows, State: models.StateDownloaded, Resolution: &res1080p},
		{TVShowID: &ep2.ID, TvgName: "Series One S01E02 720p", LineURL: &lineURL, LineContent: "#EXTINF", LineHash: "hash-3", GroupTitle: "Series", ContentType: models.ContentTypeTVShows, State: models.StateProcessed, Resolution: &res720p},
	}
	for i := range lines {
		if err := db.Create(&lines[i]).Error; err != nil {
			t.Fatalf("failed to seed processed line: %v", err)
		}
	}

	allSeries := []sonarr.Series{
		{ID: 1, Title: "Series One", TvdbID: tvdbID, Monitored: true},
	}
	episodesBySeriesID := map[int][]sonarr.Episode{
		1: {
			{SeasonNumber: 1, EpisodeNumber: 1, Monitored: true},
			{SeasonNumber: 1, EpisodeNumber: 2, Monitored: true},
		},
	}
	sonarrServer := sonarrTestServer(t, allSeries, episodesBySeriesID)
	defer sonarrServer.Close()

	newForceDownloadTestConfig(t, "", sonarrServer.URL)
	server := NewServer()

	req, _ := http.NewRequest("GET", "/api/v1/sonarr/series", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data []SonarrSeriesListItem `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 series, got %d", len(resp.Data))
	}
	if resp.Data[0].OccurrenceCount != 3 {
		t.Errorf("expected summed occurrence count 3 across both episodes, got %d", resp.Data[0].OccurrenceCount)
	}
	if resp.Data[0].MatchedCount != 2 || resp.Data[0].MonitoredCount != 2 {
		t.Errorf("expected matched_count/monitored_count 2/2, got %d/%d", resp.Data[0].MatchedCount, resp.Data[0].MonitoredCount)
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
