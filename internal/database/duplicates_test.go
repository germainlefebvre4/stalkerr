package database_test

import (
	"testing"

	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/testutil"
)

// TestFindDuplicateDownloadGroups_DifferentPathsIsDuplicate covers the spec
// scenario "Two completed occurrences at different paths is a duplicate".
func TestFindDuplicateDownloadGroups_DifferentPathsIsDuplicate(t *testing.T) {
	db := testutil.TestDB(t)
	defer testutil.CleanupDB(t, db)

	movie := testutil.CreateMovie(db)

	dl1 := testutil.CreateDownloadInfo(db, testutil.WithDownloadPath("/movies/A/movie.mkv"))
	dl2 := testutil.CreateDownloadInfo(db, testutil.WithDownloadPath("/movies/B/movie.mkv"))

	testutil.CreateProcessedLine(db, testutil.WithMovieID(movie.ID), testutil.WithState(models.StateDownloaded), testutil.WithDownloadInfoID(dl1.ID))
	testutil.CreateProcessedLine(db, testutil.WithMovieID(movie.ID), testutil.WithState(models.StateDownloaded), testutil.WithDownloadInfoID(dl2.ID))

	groups, err := database.FindDuplicateDownloadGroups(db)
	testutil.AssertNoError(t, err, "unexpected error")

	if len(groups) != 1 {
		t.Fatalf("expected 1 duplicate group, got %d", len(groups))
	}
	if groups[0].MovieID == nil || *groups[0].MovieID != movie.ID {
		t.Errorf("expected group for movie %d, got %+v", movie.ID, groups[0])
	}
	if len(groups[0].Occurrences) != 2 {
		t.Errorf("expected 2 occurrences, got %d", len(groups[0].Occurrences))
	}
}

// TestFindDuplicateDownloadGroups_SamePathIsNotDuplicate covers "A same-path
// tier-2 upgrade is not a duplicate".
func TestFindDuplicateDownloadGroups_SamePathIsNotDuplicate(t *testing.T) {
	db := testutil.TestDB(t)
	defer testutil.CleanupDB(t, db)

	movie := testutil.CreateMovie(db)

	dl1 := testutil.CreateDownloadInfo(db, testutil.WithDownloadPath("/movies/A/movie.mkv"))
	dl2 := testutil.CreateDownloadInfo(db, testutil.WithDownloadPath("/movies/A/movie.mkv"))

	testutil.CreateProcessedLine(db, testutil.WithMovieID(movie.ID), testutil.WithState(models.StateDownloaded), testutil.WithDownloadInfoID(dl1.ID))
	testutil.CreateProcessedLine(db, testutil.WithMovieID(movie.ID), testutil.WithState(models.StateDownloaded), testutil.WithDownloadInfoID(dl2.ID))

	groups, err := database.FindDuplicateDownloadGroups(db)
	testutil.AssertNoError(t, err, "unexpected error")

	if len(groups) != 0 {
		t.Fatalf("expected no duplicate groups for a same-path pair, got %d", len(groups))
	}
}

// TestFindDuplicateDownloadGroups_ForcedDownloadExcluded covers "A forced
// download is excluded from the comparison".
func TestFindDuplicateDownloadGroups_ForcedDownloadExcluded(t *testing.T) {
	db := testutil.TestDB(t)
	defer testutil.CleanupDB(t, db)

	movie := testutil.CreateMovie(db)

	automatic := testutil.CreateDownloadInfo(db, testutil.WithDownloadPath("/movies/A/movie.mkv"))
	forced := testutil.CreateDownloadInfo(db, testutil.WithDownloadPath("/movies/B/movie.mkv"), testutil.WithForced(true))

	testutil.CreateProcessedLine(db, testutil.WithMovieID(movie.ID), testutil.WithState(models.StateDownloaded), testutil.WithDownloadInfoID(automatic.ID))
	testutil.CreateProcessedLine(db, testutil.WithMovieID(movie.ID), testutil.WithState(models.StateDownloaded), testutil.WithDownloadInfoID(forced.ID))

	groups, err := database.FindDuplicateDownloadGroups(db)
	testutil.AssertNoError(t, err, "unexpected error")

	if len(groups) != 0 {
		t.Fatalf("expected no duplicate group when one occurrence is forced, got %d", len(groups))
	}
}

// TestFindDuplicateDownloadGroups_SingleCompletedIsNeverDuplicate covers "A
// single completed occurrence is never a duplicate", including when other
// non-completed occurrences exist for the same movie.
func TestFindDuplicateDownloadGroups_SingleCompletedIsNeverDuplicate(t *testing.T) {
	db := testutil.TestDB(t)
	defer testutil.CleanupDB(t, db)

	movie := testutil.CreateMovie(db)

	completed := testutil.CreateDownloadInfo(db, testutil.WithDownloadPath("/movies/A/movie.mkv"))
	failed := testutil.CreateDownloadInfo(db, testutil.WithDownloadPath("/movies/B/movie.mkv"), testutil.WithDownloadStatus(models.DownloadStatusFailed))

	testutil.CreateProcessedLine(db, testutil.WithMovieID(movie.ID), testutil.WithState(models.StateDownloaded), testutil.WithDownloadInfoID(completed.ID))
	testutil.CreateProcessedLine(db, testutil.WithMovieID(movie.ID), testutil.WithState(models.StateFailed), testutil.WithDownloadInfoID(failed.ID))

	groups, err := database.FindDuplicateDownloadGroups(db)
	testutil.AssertNoError(t, err, "unexpected error")

	if len(groups) != 0 {
		t.Fatalf("expected no duplicate group with only one completed occurrence, got %d", len(groups))
	}
}

// TestFindDuplicateDownloadGroups_TVEpisodes covers the same grouping rule
// applied to tv_show_id (each TVShow row is one specific episode).
func TestFindDuplicateDownloadGroups_TVEpisodes(t *testing.T) {
	db := testutil.TestDB(t)
	defer testutil.CleanupDB(t, db)

	episode := testutil.CreateTVShow(db)

	dl1 := testutil.CreateDownloadInfo(db, testutil.WithDownloadPath("/tv/Show/S01E01-a.mkv"))
	dl2 := testutil.CreateDownloadInfo(db, testutil.WithDownloadPath("/tv/Show/S01E01-b.mkv"))

	testutil.CreateProcessedLine(db, testutil.WithTVShowID(episode.ID), testutil.WithState(models.StateDownloaded), testutil.WithDownloadInfoID(dl1.ID))
	testutil.CreateProcessedLine(db, testutil.WithTVShowID(episode.ID), testutil.WithState(models.StateDownloaded), testutil.WithDownloadInfoID(dl2.ID))

	groups, err := database.FindDuplicateDownloadGroups(db)
	testutil.AssertNoError(t, err, "unexpected error")

	if len(groups) != 1 {
		t.Fatalf("expected 1 duplicate group, got %d", len(groups))
	}
	if groups[0].TVShowID == nil || *groups[0].TVShowID != episode.ID {
		t.Errorf("expected group for tv_show %d, got %+v", episode.ID, groups[0])
	}
}
