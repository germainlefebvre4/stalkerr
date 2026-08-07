package processor

import (
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/external/tmdb"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/models"
)

// TestBackfillRichMetadata_LegacyMovieBackfilled verifies that a Movie with a
// tmdb_id but no poster_path gets backfilled with rich metadata.
func TestBackfillRichMetadata_LegacyMovieBackfilled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupEnrichTestDB(t)
	defer teardownTestDB(t)

	db := database.Get()

	movie := models.Movie{TMDBID: 42001, TMDBTitle: "Legacy Movie", TMDBYear: 2015}
	if err := db.Create(&movie).Error; err != nil {
		t.Fatalf("failed to create test movie: %v", err)
	}

	srv := newTMDBTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/movie/42001/external_ids" {
			w.Write([]byte(`{"imdb_id":"tt0001","tvdb_id":5001}`))
			return
		}
		w.Write([]byte(`{"id":42001,"title":"Legacy Movie","poster_path":"/legacy.jpg","overview":"An old classic."}`))
	})

	client := newTMDBClientForTest(t, srv.URL)
	stats, err := BackfillRichMetadata(db, client, logger.AppLogger())
	if err != nil {
		t.Fatalf("BackfillRichMetadata error: %v", err)
	}

	if stats.Processed != 1 || stats.Updated != 1 || stats.Errors != 0 {
		t.Errorf("unexpected stats: %+v", stats)
	}

	var updated models.Movie
	db.First(&updated, movie.ID)
	if updated.PosterPath == nil || *updated.PosterPath != "/legacy.jpg" {
		t.Errorf("expected PosterPath '/legacy.jpg', got %v", updated.PosterPath)
	}
	if updated.Overview == nil || *updated.Overview != "An old classic." {
		t.Errorf("expected Overview to be backfilled, got %v", updated.Overview)
	}
	if updated.IMDBID == nil || *updated.IMDBID != "tt0001" {
		t.Errorf("expected IMDBID 'tt0001', got %v", updated.IMDBID)
	}
	if updated.TVDBID == nil || *updated.TVDBID != 5001 {
		t.Errorf("expected TVDBID 5001, got %v", updated.TVDBID)
	}
}

// TestBackfillRichMetadata_NoOpWhenNothingNeeded verifies that no TMDB calls are
// made when every record already has a poster_path.
func TestBackfillRichMetadata_NoOpWhenNothingNeeded(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupEnrichTestDB(t)
	defer teardownTestDB(t)

	db := database.Get()

	poster := "/already-set.jpg"
	movie := models.Movie{TMDBID: 42002, TMDBTitle: "Fully Enriched Movie", TMDBYear: 2016, PosterPath: &poster}
	if err := db.Create(&movie).Error; err != nil {
		t.Fatalf("failed to create test movie: %v", err)
	}

	var callCount int32
	srv := newTMDBTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	})

	client := newTMDBClientForTest(t, srv.URL)
	stats, err := BackfillRichMetadata(db, client, logger.AppLogger())
	if err != nil {
		t.Fatalf("BackfillRichMetadata error: %v", err)
	}

	if stats.Processed != 0 {
		t.Errorf("expected Processed=0, got %d", stats.Processed)
	}
	if atomic.LoadInt32(&callCount) != 0 {
		t.Errorf("expected no TMDB API calls, got %d", callCount)
	}
}

// TestBackfillRichMetadata_TVShowDeduplication verifies that TV show rows sharing
// the same tmdb_id are backfilled from a single TMDB API call.
func TestBackfillRichMetadata_TVShowDeduplication(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupEnrichTestDB(t)
	defer teardownTestDB(t)

	db := database.Get()

	s1, s2 := 1, 2
	e1, e2 := 1, 1
	shows := []models.TVShow{
		{TMDBID: 8001, TMDBTitle: "Dedup Show", TMDBYear: 2018, Season: &s1, Episode: &e1},
		{TMDBID: 8001, TMDBTitle: "Dedup Show", TMDBYear: 2018, Season: &s2, Episode: &e2},
	}
	for i := range shows {
		if err := db.Create(&shows[i]).Error; err != nil {
			t.Fatalf("failed to create tvshow: %v", err)
		}
	}

	var detailsCalls int32
	srv := newTMDBTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/tv/8001/external_ids" {
			w.Write([]byte(`{"imdb_id":"tt9999","tvdb_id":6001}`))
			return
		}
		atomic.AddInt32(&detailsCalls, 1)
		w.Write([]byte(`{"id":8001,"name":"Dedup Show","poster_path":"/dedup.jpg","overview":"Shared show overview."}`))
	})

	client := newTMDBClientForTest(t, srv.URL)
	stats, err := BackfillRichMetadata(db, client, logger.AppLogger())
	if err != nil {
		t.Fatalf("BackfillRichMetadata error: %v", err)
	}

	if stats.Updated != 2 {
		t.Errorf("expected Updated=2, got %d", stats.Updated)
	}
	if atomic.LoadInt32(&detailsCalls) != 1 {
		t.Errorf("expected 1 TMDB details call (deduplication), got %d", detailsCalls)
	}

	for i := range shows {
		var updated models.TVShow
		db.First(&updated, shows[i].ID)
		if updated.PosterPath == nil || *updated.PosterPath != "/dedup.jpg" {
			t.Errorf("expected show %d to have PosterPath backfilled, got %v", shows[i].ID, updated.PosterPath)
		}
	}
}

// TestBackfillRichMetadata_PerRecordFailureIsolation verifies that a TMDB failure
// on one record does not prevent the other records from being backfilled.
func TestBackfillRichMetadata_PerRecordFailureIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupEnrichTestDB(t)
	defer teardownTestDB(t)

	db := database.Get()

	movieErr := models.Movie{TMDBID: 42003, TMDBTitle: "Errors Out", TMDBYear: 2019}
	movieOK := models.Movie{TMDBID: 42004, TMDBTitle: "Succeeds", TMDBYear: 2019}
	db.Create(&movieErr)
	db.Create(&movieOK)

	srv := newTMDBTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/movie/42003":
			http.Error(w, "internal error", http.StatusInternalServerError)
		case "/movie/42003/external_ids":
			http.Error(w, "internal error", http.StatusInternalServerError)
		case "/movie/42004":
			w.Write([]byte(`{"id":42004,"title":"Succeeds","poster_path":"/ok.jpg","overview":"All good."}`))
		case "/movie/42004/external_ids":
			w.Write([]byte(`{"imdb_id":"tt4242","tvdb_id":7001}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	client := newTMDBClientForTest(t, srv.URL)
	stats, err := BackfillRichMetadata(db, client, logger.AppLogger())
	if err != nil {
		t.Fatalf("BackfillRichMetadata error: %v", err)
	}

	if stats.Processed != 2 {
		t.Errorf("expected Processed=2, got %d", stats.Processed)
	}
	if stats.Errors == 0 {
		t.Error("expected at least 1 error for the failing record")
	}
	if stats.Updated != 1 {
		t.Errorf("expected Updated=1 for the surviving record, got %d", stats.Updated)
	}

	var updatedErr models.Movie
	db.First(&updatedErr, movieErr.ID)
	if updatedErr.PosterPath != nil {
		t.Errorf("expected failing movie to remain unbackfilled, got %v", updatedErr.PosterPath)
	}

	var updatedOK models.Movie
	db.First(&updatedOK, movieOK.ID)
	if updatedOK.PosterPath == nil || *updatedOK.PosterPath != "/ok.jpg" {
		t.Errorf("expected succeeding movie to be backfilled, got %v", updatedOK.PosterPath)
	}
}

// TestProcess_BackfillSkippedWhenTMDBDisabled verifies that Process() does not
// invoke the rich metadata backfill (and makes no TMDB calls) when SkipTMDB is set.
func TestProcess_BackfillSkippedWhenTMDBDisabled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	db := database.Get()

	movie := models.Movie{TMDBID: 42005, TMDBTitle: "Untouched Legacy Movie", TMDBYear: 2020}
	if err := db.Create(&movie).Error; err != nil {
		t.Fatalf("failed to create test movie: %v", err)
	}

	var callCount int32
	srv := newTMDBTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	})

	config.SetConfig(&config.Config{
		TMDB: config.TMDBConfig{
			Enabled:           true,
			APIKey:            "mock-key",
			Language:          "en-US",
			RequestsPerSecond: 0,
		},
	})

	tmdb.SetBaseURL(srv.URL)
	defer tmdb.SetBaseURL("https://api.themoviedb.org/3")

	tmpFile := createTestM3U(t, "#EXTM3U\n#EXTINF:-1,Test Channel\nhttp://example.com/live.m3u8")

	p, err := NewProcessor(tmpFile)
	if err != nil {
		t.Fatalf("NewProcessor failed: %v", err)
	}

	opts := ProcessOptions{
		Force:            true,
		BatchSize:        10,
		ProgressInterval: 100,
		SkipTMDB:         true,
	}

	if _, err := p.Process(opts); err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if atomic.LoadInt32(&callCount) != 0 {
		t.Errorf("expected no TMDB calls when SkipTMDB is set, got %d", callCount)
	}

	var unchanged models.Movie
	db.First(&unchanged, movie.ID)
	if unchanged.PosterPath != nil {
		t.Errorf("expected legacy movie to remain unbackfilled when SkipTMDB is set, got %v", unchanged.PosterPath)
	}
}
