package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/external/sonarr"
	"github.com/glefebvre/stalkeer/internal/models"
)

func newForceDownloadTestConfig(t *testing.T, radarrURL, sonarrURL string) {
	t.Helper()
	tempDir := t.TempDir()
	config.SetConfig(&config.Config{
		Radarr: config.RadarrConfig{URL: radarrURL, APIKey: "radarr-key"},
		Sonarr: config.SonarrConfig{URL: sonarrURL, APIKey: "sonarr-key"},
		Downloads: config.DownloadsConfig{
			MoviesPath:    filepath.Join(tempDir, "movies"),
			TVShowsPath:   filepath.Join(tempDir, "tvshows"),
			TempDir:       filepath.Join(tempDir, "tmp"),
			Timeout:       10,
			RetryAttempts: 1,
		},
	})
}

func TestForceDownloadItem_InvalidID(t *testing.T) {
	_ = setupTestDB(t)
	newForceDownloadTestConfig(t, "", "")
	server := NewServer()

	req, _ := http.NewRequest("POST", "/api/v1/items/not-a-number/force-download", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestForceDownloadItem_NotFound(t *testing.T) {
	_ = setupTestDB(t)
	newForceDownloadTestConfig(t, "", "")
	server := NewServer()

	req, _ := http.NewRequest("POST", "/api/v1/items/999/force-download", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestForceDownloadItem_NotMatched(t *testing.T) {
	db := setupTestDB(t)
	// Point Radarr/Sonarr at unreachable addresses: if the handler contacted them
	// despite the item being unmatched, this test would fail with a network error
	// instead of the expected clean refusal.
	newForceDownloadTestConfig(t, "http://127.0.0.1:1", "http://127.0.0.1:1")
	server := NewServer()

	line := models.ProcessedLine{
		LineContent: "uncategorized content",
		LineHash:    "hash-not-matched",
		LineURL:     strPtr2("http://example.com/stream"),
		TvgName:     "Some Channel",
		ContentType: models.ContentTypeUncategorized,
		State:       models.StateProcessed,
		ProcessedAt: time.Now(),
	}
	db.Create(&line)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/items/%d/force-download", line.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d: %s", w.Code, w.Body.String())
	}

	var count int64
	db.Model(&models.DownloadInfo{}).Count(&count)
	if count != 0 {
		t.Errorf("expected no DownloadInfo rows created, got %d", count)
	}
}

func TestForceDownloadItem_AlreadyDownloaded(t *testing.T) {
	db := setupTestDB(t)
	newForceDownloadTestConfig(t, "http://127.0.0.1:1", "")
	server := NewServer()

	movie := models.Movie{TMDBID: 1, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)

	line := models.ProcessedLine{
		LineContent: "dune content",
		LineHash:    "hash-already-downloaded",
		LineURL:     strPtr2("http://example.com/stream"),
		TvgName:     "Dune",
		ContentType: models.ContentTypeMovies,
		MovieID:     &movie.ID,
		State:       models.StateDownloaded,
		ProcessedAt: time.Now(),
	}
	db.Create(&line)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/items/%d/force-download", line.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestForceDownloadItem_AlreadyDownloading(t *testing.T) {
	db := setupTestDB(t)
	newForceDownloadTestConfig(t, "http://127.0.0.1:1", "")
	server := NewServer()

	movie := models.Movie{TMDBID: 1, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)

	line := models.ProcessedLine{
		LineContent: "dune content",
		LineHash:    "hash-already-downloading",
		LineURL:     strPtr2("http://example.com/stream"),
		TvgName:     "Dune",
		ContentType: models.ContentTypeMovies,
		MovieID:     &movie.ID,
		State:       models.StateDownloading,
		ProcessedAt: time.Now(),
	}
	db.Create(&line)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/items/%d/force-download", line.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestForceDownloadItem_MovieNotFoundInRadarr(t *testing.T) {
	db := setupTestDB(t)

	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{})
	}))
	defer radarrServer.Close()

	newForceDownloadTestConfig(t, radarrServer.URL, "")
	server := NewServer()

	movie := models.Movie{TMDBID: 42, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)

	line := models.ProcessedLine{
		LineContent: "dune content",
		LineHash:    "hash-not-found-radarr",
		LineURL:     strPtr2("http://example.com/stream"),
		TvgName:     "Dune",
		ContentType: models.ContentTypeMovies,
		MovieID:     &movie.ID,
		State:       models.StateProcessed,
		ProcessedAt: time.Now(),
	}
	db.Create(&line)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/items/%d/force-download", line.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d: %s", w.Code, w.Body.String())
	}

	var count int64
	db.Model(&models.DownloadInfo{}).Count(&count)
	if count != 0 {
		t.Errorf("expected no DownloadInfo rows created on refusal, got %d", count)
	}
}

func TestForceDownloadItem_RadarrCheckFails(t *testing.T) {
	db := setupTestDB(t)

	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer radarrServer.Close()

	newForceDownloadTestConfig(t, radarrServer.URL, "")
	server := NewServer()

	movie := models.Movie{TMDBID: 42, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)

	line := models.ProcessedLine{
		LineContent: "dune content",
		LineHash:    "hash-radarr-check-failed",
		LineURL:     strPtr2("http://example.com/stream"),
		TvgName:     "Dune",
		ContentType: models.ContentTypeMovies,
		MovieID:     &movie.ID,
		State:       models.StateProcessed,
		ProcessedAt: time.Now(),
	}
	db.Create(&line)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/items/%d/force-download", line.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d: %s", w.Code, w.Body.String())
	}

	var count int64
	db.Model(&models.DownloadInfo{}).Count(&count)
	if count != 0 {
		t.Errorf("expected no DownloadInfo rows created on refusal, got %d", count)
	}
}

func TestForceDownloadItem_TVEpisodeMissingTVDBID(t *testing.T) {
	db := setupTestDB(t)
	newForceDownloadTestConfig(t, "", "http://127.0.0.1:1")
	server := NewServer()

	season, episode := 1, 2
	tvshow := models.TVShow{TMDBID: 7, TMDBTitle: "Breaking Bad", TMDBYear: 2008, Season: &season, Episode: &episode}
	db.Create(&tvshow)

	line := models.ProcessedLine{
		LineContent: "bb content",
		LineHash:    "hash-tvdb-nil",
		LineURL:     strPtr2("http://example.com/stream"),
		TvgName:     "Breaking Bad",
		ContentType: models.ContentTypeTVShows,
		TVShowID:    &tvshow.ID,
		State:       models.StateProcessed,
		ProcessedAt: time.Now(),
	}
	db.Create(&line)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/items/%d/force-download", line.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 (fail-closed on nil TVDBID), got %d: %s", w.Code, w.Body.String())
	}

	var count int64
	db.Model(&models.DownloadInfo{}).Count(&count)
	if count != 0 {
		t.Errorf("expected no DownloadInfo rows created on refusal, got %d", count)
	}
}

// blockingDownloadSource serves content only after release() is called, and reports
// via hitCount how many times it has been requested.
type blockingDownloadSource struct {
	server   *httptest.Server
	release  chan struct{}
	hitCount chan int
}

func newBlockingDownloadSource(t *testing.T, content []byte) *blockingDownloadSource {
	t.Helper()
	b := &blockingDownloadSource{
		release:  make(chan struct{}),
		hitCount: make(chan int, 100),
	}
	hits := 0
	b.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		select {
		case b.hitCount <- hits:
		default:
		}
		<-b.release
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}))
	return b
}

func (b *blockingDownloadSource) Close() { b.server.Close() }

func TestForceDownloadItem_MovieSuccess_AcceptedBeforeTransferCompletes(t *testing.T) {
	db := setupTestDB(t)

	moviePath := t.TempDir()
	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 5, "title": "Dune", "year": 2021, "tmdbId": 42, "path": moviePath, "monitored": true},
		})
	}))
	defer radarrServer.Close()

	newForceDownloadTestConfig(t, radarrServer.URL, "")
	server := NewServer()

	// The transfer must never be attempted in the request path, so use a source
	// that would block forever if it were ever hit.
	source := newBlockingDownloadSource(t, []byte("movie bytes"))
	defer source.Close()

	movie := models.Movie{TMDBID: 42, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)

	resolution := "1080p"
	line := models.ProcessedLine{
		LineContent: "dune content",
		LineHash:    "hash-movie-success",
		LineURL:     &source.server.URL,
		TvgName:     "Dune",
		ContentType: models.ContentTypeMovies,
		MovieID:     &movie.ID,
		Resolution:  &resolution,
		State:       models.StateProcessed,
		ProcessedAt: time.Now(),
	}
	db.Create(&line)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/items/%d/force-download", line.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d: %s", w.Code, w.Body.String())
	}

	var resp ForceDownloadResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Status != "queued" {
		t.Errorf("expected status 'queued', got %q", resp.Status)
	}

	// The transfer must not have been attempted: no hit should reach the download
	// source within a short timeout.
	select {
	case <-source.hitCount:
		t.Fatal("expected the file transfer not to start in the request path")
	case <-time.After(200 * time.Millisecond):
	}

	// Verify the resolution-suffixed path was persisted, and the record is pending.
	var dl models.DownloadInfo
	if err := db.First(&dl).Error; err != nil {
		t.Fatalf("expected a DownloadInfo row to exist: %v", err)
	}
	if dl.Status != string(models.DownloadStatusPending) {
		t.Errorf("expected DownloadInfo status %q, got %q", models.DownloadStatusPending, dl.Status)
	}
	if dl.DownloadPath == nil {
		t.Fatal("expected DownloadPath to be persisted")
	}
	if !containsSubstring(*dl.DownloadPath, "[1080p]") {
		t.Errorf("expected resolution-suffixed path, got %q", *dl.DownloadPath)
	}

	var lineAfter models.ProcessedLine
	db.First(&lineAfter, line.ID)
	if lineAfter.State != models.StateProcessed {
		t.Errorf("expected ProcessedLine state unchanged (processed), got %q", lineAfter.State)
	}
}

// 5.2 A forced download whose occurrence has a known language and VFQ variant
// carries all three tags (resolution, language, variant) in its download path.
func TestForceDownloadItem_MovieSuccess_LanguageAndVariantTagged(t *testing.T) {
	db := setupTestDB(t)

	moviePath := t.TempDir()
	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 5, "title": "Dune", "year": 2021, "tmdbId": 43, "path": moviePath, "monitored": true},
		})
	}))
	defer radarrServer.Close()

	newForceDownloadTestConfig(t, radarrServer.URL, "")
	server := NewServer()

	source := newBlockingDownloadSource(t, []byte("movie bytes"))
	defer source.Close()

	movie := models.Movie{TMDBID: 43, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)

	resolution := "1080p"
	language := "MULTI"
	frenchVariant := "VFQ"
	line := models.ProcessedLine{
		LineContent:   "dune content",
		LineHash:      "hash-movie-lang-variant",
		LineURL:       &source.server.URL,
		TvgName:       "Dune",
		ContentType:   models.ContentTypeMovies,
		MovieID:       &movie.ID,
		Resolution:    &resolution,
		Language:      &language,
		FrenchVariant: &frenchVariant,
		State:         models.StateProcessed,
		ProcessedAt:   time.Now(),
	}
	db.Create(&line)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/items/%d/force-download", line.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d: %s", w.Code, w.Body.String())
	}

	select {
	case <-source.hitCount:
		t.Fatal("expected the file transfer not to start in the request path")
	case <-time.After(200 * time.Millisecond):
	}

	var dl models.DownloadInfo
	if err := db.First(&dl).Error; err != nil {
		t.Fatalf("expected a DownloadInfo row to exist: %v", err)
	}
	if dl.DownloadPath == nil {
		t.Fatal("expected DownloadPath to be persisted")
	}
	if !containsSubstring(*dl.DownloadPath, "[1080p][MULTI][VFQ]") {
		t.Errorf("expected resolution/language/variant-tagged path, got %q", *dl.DownloadPath)
	}
}

// TestForceDownloadItem_RepeatedRequest_ReusesDownloadInfo replaces the former
// TestForceDownloadItem_ConcurrentRequest_SecondRejectedWithoutDuplicateTransfer:
// now that the transfer is deferred to the next download cron run instead of
// starting synchronously, a still-processed occurrence never reaches
// StateDownloading between two back-to-back requests, so both are accepted and
// persistForceDownloadPath's existing item.DownloadInfoID != nil branch reuses
// the same record instead of creating a second one.
func TestForceDownloadItem_RepeatedRequest_ReusesDownloadInfo(t *testing.T) {
	db := setupTestDB(t)

	moviePath := t.TempDir()
	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 5, "title": "Dune", "year": 2021, "tmdbId": 42, "path": moviePath, "monitored": true},
		})
	}))
	defer radarrServer.Close()

	newForceDownloadTestConfig(t, radarrServer.URL, "")
	server := NewServer()

	line := models.ProcessedLine{
		LineContent: "dune content",
		LineHash:    "hash-movie-repeated",
		LineURL:     strPtr2("http://example.com/stream"),
		TvgName:     "Dune",
		ContentType: models.ContentTypeMovies,
		State:       models.StateProcessed,
		ProcessedAt: time.Now(),
	}
	movie := models.Movie{TMDBID: 42, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)
	line.MovieID = &movie.ID
	db.Create(&line)

	req1, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/items/%d/force-download", line.ID), nil)
	w1 := httptest.NewRecorder()
	server.router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusAccepted {
		t.Fatalf("expected first request to be accepted (202), got %d: %s", w1.Code, w1.Body.String())
	}

	req2, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/items/%d/force-download", line.ID), nil)
	w2 := httptest.NewRecorder()
	server.router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusAccepted {
		t.Fatalf("expected second request to also be accepted (202), got %d: %s", w2.Code, w2.Body.String())
	}

	var downloadInfoCount int64
	db.Model(&models.DownloadInfo{}).Count(&downloadInfoCount)
	if downloadInfoCount != 1 {
		t.Errorf("expected exactly one DownloadInfo row (second request reused the first), got %d", downloadInfoCount)
	}
}

func TestForceDownloadItem_SiblingOccurrenceUnaffected(t *testing.T) {
	db := setupTestDB(t)

	moviePath := t.TempDir()
	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 5, "title": "Dune", "year": 2021, "tmdbId": 42, "path": moviePath, "monitored": true},
		})
	}))
	defer radarrServer.Close()

	newForceDownloadTestConfig(t, radarrServer.URL, "")
	server := NewServer()

	movie := models.Movie{TMDBID: 42, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)

	siblingPath := "/existing/sibling/Dune (2021).mkv"
	siblingDownload := models.DownloadInfo{
		URL:          "http://example.com/sibling",
		Status:       string(models.DownloadStatusCompleted),
		DownloadPath: &siblingPath,
	}
	db.Create(&siblingDownload)

	sibling := models.ProcessedLine{
		LineContent:    "dune sd content",
		LineHash:       "hash-sibling-downloaded",
		LineURL:        strPtr2("http://example.com/sibling-stream"),
		TvgName:        "Dune",
		ContentType:    models.ContentTypeMovies,
		MovieID:        &movie.ID,
		DownloadInfoID: &siblingDownload.ID,
		State:          models.StateDownloaded,
		ProcessedAt:    time.Now(),
	}
	db.Create(&sibling)

	target := models.ProcessedLine{
		LineContent: "dune hd content",
		LineHash:    "hash-target-forced",
		LineURL:     strPtr2("http://example.com/target-stream"),
		TvgName:     "Dune",
		ContentType: models.ContentTypeMovies,
		MovieID:     &movie.ID,
		State:       models.StateProcessed,
		ProcessedAt: time.Now(),
	}
	db.Create(&target)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/items/%d/force-download", target.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d: %s", w.Code, w.Body.String())
	}

	var targetAfter models.ProcessedLine
	db.First(&targetAfter, target.ID)
	if targetAfter.DownloadInfoID == nil {
		t.Fatal("expected target to have a DownloadInfo linked")
	}
	if *targetAfter.DownloadInfoID == siblingDownload.ID {
		t.Error("expected the forced download to use a distinct DownloadInfo row from the sibling")
	}

	var targetDL models.DownloadInfo
	db.First(&targetDL, *targetAfter.DownloadInfoID)
	if targetDL.Status != string(models.DownloadStatusPending) {
		t.Errorf("expected target DownloadInfo status %q, got %q", models.DownloadStatusPending, targetDL.Status)
	}

	var siblingAfter models.DownloadInfo
	db.First(&siblingAfter, siblingDownload.ID)
	if siblingAfter.DownloadPath == nil || *siblingAfter.DownloadPath != siblingPath {
		t.Errorf("expected sibling DownloadPath unchanged (%q), got %v", siblingPath, siblingAfter.DownloadPath)
	}
	if siblingAfter.Status != string(models.DownloadStatusCompleted) {
		t.Errorf("expected sibling status unchanged (completed), got %q", siblingAfter.Status)
	}

	var siblingLine models.ProcessedLine
	db.First(&siblingLine, sibling.ID)
	if siblingLine.State != models.StateDownloaded {
		t.Errorf("expected sibling ProcessedLine state unchanged (downloaded), got %q", siblingLine.State)
	}
}

func TestForceDownloadItem_MovieNotMonitored(t *testing.T) {
	db := setupTestDB(t)

	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 5, "title": "Dune", "year": 2021, "tmdbId": 42, "path": t.TempDir(), "monitored": false},
		})
	}))
	defer radarrServer.Close()

	newForceDownloadTestConfig(t, radarrServer.URL, "")
	server := NewServer()

	movie := models.Movie{TMDBID: 42, TMDBTitle: "Dune", TMDBYear: 2021}
	db.Create(&movie)

	line := models.ProcessedLine{
		LineContent: "dune content",
		LineHash:    "hash-movie-not-monitored",
		LineURL:     strPtr2("http://example.com/stream"),
		TvgName:     "Dune",
		ContentType: models.ContentTypeMovies,
		MovieID:     &movie.ID,
		State:       models.StateProcessed,
		ProcessedAt: time.Now(),
	}
	db.Create(&line)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/items/%d/force-download", line.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d: %s", w.Code, w.Body.String())
	}

	var body ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body.Error != "not_monitored" {
		t.Errorf("expected error code not_monitored, got %q", body.Error)
	}

	var count int64
	db.Model(&models.DownloadInfo{}).Count(&count)
	if count != 0 {
		t.Errorf("expected no DownloadInfo rows created, got %d", count)
	}
}

func TestForceDownloadItem_EpisodeSeriesNotMonitored(t *testing.T) {
	db := setupTestDB(t)

	tvdbID := 555
	season, episode := 1, 2

	sonarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v3/series":
			json.NewEncoder(w).Encode([]sonarr.Series{
				{ID: 1, Title: "Breaking Bad", TvdbID: tvdbID, Monitored: false, Path: t.TempDir()},
			})
		case "/api/v3/episode":
			json.NewEncoder(w).Encode([]sonarr.Episode{
				{ID: 10, SeriesID: 1, SeasonNumber: season, EpisodeNumber: episode},
			})
		default:
			t.Errorf("unexpected sonarr path %s", r.URL.Path)
		}
	}))
	defer sonarrServer.Close()

	newForceDownloadTestConfig(t, "", sonarrServer.URL)
	server := NewServer()

	tvshow := models.TVShow{TMDBID: 7, TVDBID: &tvdbID, TMDBTitle: "Breaking Bad", TMDBYear: 2008, Season: &season, Episode: &episode}
	db.Create(&tvshow)

	line := models.ProcessedLine{
		LineContent: "bb content",
		LineHash:    "hash-series-not-monitored",
		LineURL:     strPtr2("http://example.com/stream"),
		TvgName:     "Breaking Bad",
		ContentType: models.ContentTypeTVShows,
		TVShowID:    &tvshow.ID,
		State:       models.StateProcessed,
		ProcessedAt: time.Now(),
	}
	db.Create(&line)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/items/%d/force-download", line.ID), nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d: %s", w.Code, w.Body.String())
	}

	var body ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body.Error != "not_monitored" {
		t.Errorf("expected error code not_monitored, got %q", body.Error)
	}

	var count int64
	db.Model(&models.DownloadInfo{}).Count(&count)
	if count != 0 {
		t.Errorf("expected no DownloadInfo rows created, got %d", count)
	}
}

func strPtr2(s string) *string { return &s }

func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}
