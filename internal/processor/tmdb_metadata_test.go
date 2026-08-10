package processor

import (
	"net/http"
	"strings"
	"testing"

	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/external/tmdb"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/models"
)

// newTestProcessor builds a minimal Processor sufficient for exercising the
// TMDB enrichment methods directly, bypassing M3U parsing/classification.
func newTestProcessor(client *tmdb.Client) *Processor {
	return &Processor{
		tmdbClient: client,
		logger:     logger.AppLogger(),
		db:         database.Get(),
	}
}

// TestEnrichMovieWithTMDBID_PersistsRichMetadata verifies that poster_path,
// overview, imdb_id, and tvdb_id are persisted from the already-fetched TMDB
// detail and external-ID responses, with no additional TMDB call.
func TestEnrichMovieWithTMDBID_PersistsRichMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	srv := newTMDBTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "external_ids") {
			w.Write([]byte(`{"imdb_id": "tt1234567", "tvdb_id": 555}`))
			return
		}
		w.Write([]byte(`{"id": 42, "title": "Test Movie", "release_date": "2020-01-01", "poster_path": "/poster.jpg", "overview": "A great movie.", "genres": []}`))
	})

	client := newTMDBClientForTest(t, srv.URL)
	p := newTestProcessor(client)

	line := &models.ProcessedLine{}
	stats := &Statistics{}
	if err := p.enrichMovieWithTMDBID(line, 42, "en-US", stats); err != nil {
		t.Fatalf("enrichMovieWithTMDBID error: %v", err)
	}

	if line.MovieID == nil {
		t.Fatal("expected MovieID to be set")
	}

	var movie models.Movie
	if err := database.Get().First(&movie, *line.MovieID).Error; err != nil {
		t.Fatalf("failed to load persisted movie: %v", err)
	}

	if movie.PosterPath == nil || *movie.PosterPath != "/poster.jpg" {
		t.Errorf("expected poster_path=/poster.jpg, got %v", movie.PosterPath)
	}
	if movie.Overview == nil || *movie.Overview != "A great movie." {
		t.Errorf("expected overview set, got %v", movie.Overview)
	}
	if movie.IMDBID == nil || *movie.IMDBID != "tt1234567" {
		t.Errorf("expected imdb_id=tt1234567, got %v", movie.IMDBID)
	}
	if movie.TVDBID == nil || *movie.TVDBID != 555 {
		t.Errorf("expected tvdb_id=555, got %v", movie.TVDBID)
	}
}

// TestEnrichMovieWithTMDBID_ExternalIDsUnavailable verifies that poster/overview
// are still persisted from details when the external-IDs lookup fails, while
// imdb_id/tvdb_id are left null.
func TestEnrichMovieWithTMDBID_ExternalIDsUnavailable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	srv := newTMDBTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "external_ids") {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id": 43, "title": "No External IDs Movie", "release_date": "2021-01-01", "poster_path": "/poster2.jpg", "overview": "Another movie.", "genres": []}`))
	})

	client := newTMDBClientForTest(t, srv.URL)
	p := newTestProcessor(client)

	line := &models.ProcessedLine{}
	stats := &Statistics{}
	if err := p.enrichMovieWithTMDBID(line, 43, "en-US", stats); err != nil {
		t.Fatalf("enrichMovieWithTMDBID error: %v", err)
	}

	var movie models.Movie
	if err := database.Get().First(&movie, *line.MovieID).Error; err != nil {
		t.Fatalf("failed to load persisted movie: %v", err)
	}

	if movie.PosterPath == nil || *movie.PosterPath != "/poster2.jpg" {
		t.Errorf("expected poster_path=/poster2.jpg, got %v", movie.PosterPath)
	}
	if movie.IMDBID != nil {
		t.Errorf("expected imdb_id to remain nil, got %v", *movie.IMDBID)
	}
	if movie.TVDBID != nil {
		t.Errorf("expected tvdb_id to remain nil, got %v", *movie.TVDBID)
	}
}

// TestEnrichTVShowWithTMDBID_PersistsRichMetadata verifies that poster_path,
// overview, imdb_id, and tvdb_id are persisted for TV show enrichment.
func TestEnrichTVShowWithTMDBID_PersistsRichMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	setupTestDB(t)
	defer teardownTestDB(t)

	srv := newTMDBTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "external_ids") {
			w.Write([]byte(`{"imdb_id": "tt7654321", "tvdb_id": 777}`))
			return
		}
		w.Write([]byte(`{"id": 99, "name": "Test Show", "first_air_date": "2019-05-01", "poster_path": "/showposter.jpg", "overview": "A great show.", "genres": []}`))
	})

	client := newTMDBClientForTest(t, srv.URL)
	p := newTestProcessor(client)

	season, episode := 1, 1
	line := &models.ProcessedLine{}
	stats := &Statistics{}
	if err := p.enrichTVShowWithTMDBID(line, 99, &season, &episode, "en-US", stats); err != nil {
		t.Fatalf("enrichTVShowWithTMDBID error: %v", err)
	}

	if line.TVShowID == nil {
		t.Fatal("expected TVShowID to be set")
	}

	var show models.TVShow
	if err := database.Get().First(&show, *line.TVShowID).Error; err != nil {
		t.Fatalf("failed to load persisted tvshow: %v", err)
	}

	if show.PosterPath == nil || *show.PosterPath != "/showposter.jpg" {
		t.Errorf("expected poster_path=/showposter.jpg, got %v", show.PosterPath)
	}
	if show.Overview == nil || *show.Overview != "A great show." {
		t.Errorf("expected overview set, got %v", show.Overview)
	}
	if show.IMDBID == nil || *show.IMDBID != "tt7654321" {
		t.Errorf("expected imdb_id=tt7654321, got %v", show.IMDBID)
	}
	if show.TVDBID == nil || *show.TVDBID != 777 {
		t.Errorf("expected tvdb_id=777, got %v", show.TVDBID)
	}
}
