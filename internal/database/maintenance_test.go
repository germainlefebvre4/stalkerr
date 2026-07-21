package database_test

import (
	"testing"

	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/testutil"
	"gorm.io/gorm"
)

func TestPruneProcessedLines(t *testing.T) {
	db := testutil.TestDB(t)
	defer testutil.CleanupDB(t, db)

	// Create test processed lines with unique hashes
	testutil.CreateProcessedLine(db, func(pl *models.ProcessedLine) {
		pl.LineHash = "hash1"
		pl.State = models.StateProcessed
	})
	testutil.CreateProcessedLine(db, func(pl *models.ProcessedLine) {
		pl.LineHash = "hash2"
		pl.State = models.StateDownloaded
	})
	testutil.CreateProcessedLine(db, func(pl *models.ProcessedLine) {
		pl.LineHash = "hash3"
		pl.State = models.StateFailed
	})

	// Active hashes is only hash1. So hash2 and hash3 are obsolete (expired).
	activeHashes := []string{"hash1"}

	// 1. Soft Prune (hard = false)
	// Only line3 should be pruned because it's failed and not in active list.
	// line2 is downloaded, so it is preserved.
	pruned, err := database.PruneProcessedLines(db, activeHashes, false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pruned != 1 {
		t.Errorf("expected 1 pruned line, got %d", pruned)
	}

	// Verify what remains in the DB
	var remaining []models.ProcessedLine
	db.Find(&remaining)
	if len(remaining) != 2 {
		t.Errorf("expected 2 remaining lines, got %d", len(remaining))
	}

	// 2. Hard Prune (hard = true)
	// Now let's prune with hard = true.
	// line2 should be pruned now as well, even though it's downloaded.
	pruned, err = database.PruneProcessedLines(db, activeHashes, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pruned != 1 {
		t.Errorf("expected 1 pruned line, got %d", pruned)
	}

	db.Find(&remaining)
	if len(remaining) != 1 {
		t.Errorf("expected 1 remaining line, got %d", len(remaining))
	}
	if remaining[0].LineHash != "hash1" {
		t.Errorf("expected remaining line to be hash1, got %s", remaining[0].LineHash)
	}
}

func TestCleanupOrphanedMetadata(t *testing.T) {
	db := testutil.TestDB(t)
	defer testutil.CleanupDB(t, db)

	// Create test movie and tvshow with unique titles to avoid unique index conflicts
	movie1 := testutil.CreateMovie(db, func(m *models.Movie) { 
		m.TMDBID = 1 
		m.TMDBTitle = "Test Movie 1"
	})
	testutil.CreateMovie(db, func(m *models.Movie) { 
		m.TMDBID = 2 
		m.TMDBTitle = "Test Movie 2" // Will be orphaned
	})

	tvshow1 := testutil.CreateTVShow(db, func(t *models.TVShow) { 
		t.TMDBID = 10 
		t.TMDBTitle = "Test Show 1"
	})
	testutil.CreateTVShow(db, func(t *models.TVShow) { 
		t.TMDBID = 20 
		t.TMDBTitle = "Test Show 2" // Will be orphaned
	})

	// Associate only movie1 and tvshow1 with a processed line
	testutil.CreateProcessedLine(db, func(pl *models.ProcessedLine) {
		pl.MovieID = &movie1.ID
		pl.LineHash = "hash-movie"
	})
	testutil.CreateProcessedLine(db, func(pl *models.ProcessedLine) {
		pl.TVShowID = &tvshow1.ID
		pl.LineHash = "hash-tv"
	})

	// Run cleanup
	prunedMovies, prunedTVShows, err := database.CleanupOrphanedMetadata(db)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if prunedMovies != 1 {
		t.Errorf("expected 1 pruned movie, got %d", prunedMovies)
	}
	if prunedTVShows != 1 {
		t.Errorf("expected 1 pruned tvshow, got %d", prunedTVShows)
	}

	// Verify only movie1 and tvshow1 remain
	var movies []models.Movie
	db.Find(&movies)
	if len(movies) != 1 || movies[0].ID != movie1.ID {
		t.Errorf("expected only movie1 to remain")
	}

	var tvshows []models.TVShow
	db.Find(&tvshows)
	if len(tvshows) != 1 || tvshows[0].ID != tvshow1.ID {
		t.Errorf("expected only tvshow1 to remain")
	}
}

func TestResetMovieAndTVShow(t *testing.T) {
	db := testutil.TestDB(t)
	defer testutil.CleanupDB(t, db)

	// Create test movie
	movie := testutil.CreateMovie(db)
	testutil.CreateProcessedLine(db, func(pl *models.ProcessedLine) {
		pl.MovieID = &movie.ID
		pl.LineHash = "movie-line"
	})

	// Create test tvshow
	tvshow := testutil.CreateTVShow(db)
	testutil.CreateProcessedLine(db, func(pl *models.ProcessedLine) {
		pl.TVShowID = &tvshow.ID
		pl.LineHash = "tvshow-line"
	})

	// Test ResetMovie
	rows, err := database.ResetMovie(db, movie.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row deleted, got %d", rows)
	}

	// Verify processed line is gone
	var count int64
	db.Model(&models.ProcessedLine{}).Where("movie_id = ?", movie.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 processed lines for movie, got %d", count)
	}

	// Test ResetMovie on non-existent
	_, err = database.ResetMovie(db, 9999)
	if err != gorm.ErrRecordNotFound {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}

	// Test ResetTVShow
	rows, err = database.ResetTVShow(db, tvshow.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row deleted, got %d", rows)
	}

	// Verify processed line is gone
	db.Model(&models.ProcessedLine{}).Where("tv_show_id = ?", tvshow.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 processed lines for tvshow, got %d", count)
	}

	// Test ResetTVShow on non-existent
	_, err = database.ResetTVShow(db, 9999)
	if err != gorm.ErrRecordNotFound {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}
}
