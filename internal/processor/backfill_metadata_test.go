package processor

import (
	"net/http"
	"strings"
	"testing"

	"github.com/glefebvre/stalkeer/internal/classifier"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/filter"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/parser"
)

// TestBackfillRichMetadata_MovieBackfilled verifies that a legacy Movie record
// (tmdb_id set, poster_path null) is backfilled with poster/overview/external IDs.
func TestBackfillRichMetadata_MovieBackfilled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	db := database.Get()
	movie := models.Movie{TMDBID: 321, TMDBTitle: "Legacy Movie", TMDBYear: 2018}
	if err := db.Create(&movie).Error; err != nil {
		t.Fatalf("failed to create legacy movie: %v", err)
	}

	srv := newTMDBTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "external_ids") {
			w.Write([]byte(`{"imdb_id": "tt9999999", "tvdb_id": 4242}`))
			return
		}
		w.Write([]byte(`{"id": 321, "title": "Legacy Movie", "release_date": "2018-01-01", "poster_path": "/legacy.jpg", "overview": "Legacy overview.", "genres": []}`))
	})

	client := newTMDBClientForTest(t, srv.URL)
	stats, err := BackfillRichMetadata(db, client, logger.AppLogger())
	if err != nil {
		t.Fatalf("BackfillRichMetadata error: %v", err)
	}

	if stats.Updated != 1 {
		t.Errorf("expected Updated=1, got %d", stats.Updated)
	}

	var updated models.Movie
	if err := db.First(&updated, movie.ID).Error; err != nil {
		t.Fatalf("failed to load movie: %v", err)
	}
	if updated.PosterPath == nil || *updated.PosterPath != "/legacy.jpg" {
		t.Errorf("expected poster_path=/legacy.jpg, got %v", updated.PosterPath)
	}
	if updated.Overview == nil || *updated.Overview != "Legacy overview." {
		t.Errorf("expected overview set, got %v", updated.Overview)
	}
	if updated.IMDBID == nil || *updated.IMDBID != "tt9999999" {
		t.Errorf("expected imdb_id=tt9999999, got %v", updated.IMDBID)
	}
	if updated.TVDBID == nil || *updated.TVDBID != 4242 {
		t.Errorf("expected tvdb_id=4242, got %v", updated.TVDBID)
	}
}

// TestBackfillRichMetadata_NoOpWhenNothingNeedsBackfill verifies that no TMDB
// calls are made when every record already has poster_path populated.
func TestBackfillRichMetadata_NoOpWhenNothingNeedsBackfill(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	db := database.Get()
	poster := "/already-there.jpg"
	movie := models.Movie{TMDBID: 654, TMDBTitle: "Already Enriched", TMDBYear: 2022, PosterPath: &poster}
	if err := db.Create(&movie).Error; err != nil {
		t.Fatalf("failed to create movie: %v", err)
	}

	callCount := 0
	srv := newTMDBTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
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
	if callCount != 0 {
		t.Errorf("expected no TMDB API calls, got %d", callCount)
	}
}

// TestBackfillRichMetadata_TVShowDeduplication verifies that multiple TVShow
// rows sharing the same tmdb_id result in only one details + one external-IDs
// TMDB call.
func TestBackfillRichMetadata_TVShowDeduplication(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	db := database.Get()
	s1, s2 := 1, 2
	e1, e2, e3 := 1, 2, 1
	shows := []models.TVShow{
		{TMDBID: 8765, TMDBTitle: "Dedup Show", TMDBYear: 2020, Season: &s1, Episode: &e1},
		{TMDBID: 8765, TMDBTitle: "Dedup Show", TMDBYear: 2020, Season: &s1, Episode: &e2},
		{TMDBID: 8765, TMDBTitle: "Dedup Show", TMDBYear: 2020, Season: &s2, Episode: &e3},
	}
	for i := range shows {
		if err := db.Create(&shows[i]).Error; err != nil {
			t.Fatalf("failed to create tvshow: %v", err)
		}
	}

	detailsCalls, externalCalls := 0, 0
	srv := newTMDBTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "external_ids") {
			externalCalls++
			w.Write([]byte(`{"imdb_id": "tt1112223", "tvdb_id": 111}`))
			return
		}
		detailsCalls++
		w.Write([]byte(`{"id": 8765, "name": "Dedup Show", "first_air_date": "2020-01-01", "poster_path": "/dedup.jpg", "overview": "Dedup overview.", "genres": []}`))
	})

	client := newTMDBClientForTest(t, srv.URL)
	stats, err := BackfillRichMetadata(db, client, logger.AppLogger())
	if err != nil {
		t.Fatalf("BackfillRichMetadata error: %v", err)
	}

	if stats.Updated != 3 {
		t.Errorf("expected Updated=3, got %d", stats.Updated)
	}
	if detailsCalls != 1 {
		t.Errorf("expected 1 details call (deduplication), got %d", detailsCalls)
	}
	if externalCalls != 1 {
		t.Errorf("expected 1 external-ids call (deduplication), got %d", externalCalls)
	}
}

// TestBackfillRichMetadata_PerRecordFailureIsolation verifies that a TMDB
// lookup failure for one record does not abort the overall backfill run.
func TestBackfillRichMetadata_PerRecordFailureIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	db := database.Get()
	movie1 := models.Movie{TMDBID: 111, TMDBTitle: "Error Movie", TMDBYear: 2019}
	movie2 := models.Movie{TMDBID: 222, TMDBTitle: "OK Movie", TMDBYear: 2019}
	db.Create(&movie1)
	db.Create(&movie2)

	srv := newTMDBTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/111") {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "external_ids") {
			w.Write([]byte(`{"imdb_id": "tt2223334", "tvdb_id": 222}`))
			return
		}
		w.Write([]byte(`{"id": 222, "title": "OK Movie", "release_date": "2019-01-01", "poster_path": "/ok.jpg", "overview": "OK overview.", "genres": []}`))
	})

	client := newTMDBClientForTest(t, srv.URL)
	stats, err := BackfillRichMetadata(db, client, logger.AppLogger())
	if err != nil {
		t.Fatalf("BackfillRichMetadata error: %v", err)
	}

	if stats.Errors == 0 {
		t.Error("expected at least 1 error")
	}
	if stats.Updated != 1 {
		t.Errorf("expected Updated=1, got %d", stats.Updated)
	}

	var okMovie models.Movie
	db.First(&okMovie, movie2.ID)
	if okMovie.PosterPath == nil || *okMovie.PosterPath != "/ok.jpg" {
		t.Errorf("expected poster_path=/ok.jpg on surviving record, got %v", okMovie.PosterPath)
	}

	var errMovie models.Movie
	db.First(&errMovie, movie1.ID)
	if errMovie.PosterPath != nil {
		t.Errorf("expected poster_path to remain nil on failed record, got %v", *errMovie.PosterPath)
	}
}

// TestProcess_BackfillSkippedWhenSkipTMDB verifies that the automatic backfill
// makes no TMDB calls when the process run uses --skip-tmdb, even though a
// legacy record needing backfill exists.
func TestProcess_BackfillSkippedWhenSkipTMDB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	db := database.Get()
	movie := models.Movie{TMDBID: 999, TMDBTitle: "Legacy Movie", TMDBYear: 2017}
	if err := db.Create(&movie).Error; err != nil {
		t.Fatalf("failed to create legacy movie: %v", err)
	}

	callCount := 0
	srv := newTMDBTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id": 999, "title": "Legacy Movie", "release_date": "2017-01-01", "poster_path": "/x.jpg", "overview": "x", "genres": []}`))
	})
	client := newTMDBClientForTest(t, srv.URL)

	tmpFile := createTestM3U(t, "#EXTM3U\n#EXTINF:-1,Some Channel\nhttp://example.com/stream.ts")

	f := filter.NewManager()
	if err := f.LoadAll(); err != nil {
		t.Fatalf("failed to load filters: %v", err)
	}

	p := &Processor{
		filePath:   tmpFile,
		parser:     parser.NewParserWithLogger(tmpFile, logger.AppLogger()),
		classifier: classifier.New(),
		filter:     f,
		tmdbClient: client,
		logger:     logger.AppLogger(),
		db:         db,
	}

	stats, err := p.Process(ProcessOptions{SkipTMDB: true})
	if err != nil {
		t.Fatalf("Process error: %v", err)
	}

	if callCount != 0 {
		t.Errorf("expected no TMDB calls when SkipTMDB is set, got %d", callCount)
	}
	if stats.MetadataBackfilled != 0 {
		t.Errorf("expected MetadataBackfilled=0 when SkipTMDB is set, got %d", stats.MetadataBackfilled)
	}

	var unchanged models.Movie
	db.First(&unchanged, movie.ID)
	if unchanged.PosterPath != nil {
		t.Errorf("expected legacy movie to remain unbackfilled, got poster_path=%v", *unchanged.PosterPath)
	}
}
