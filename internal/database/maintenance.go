package database

import (
	"fmt"

	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/gorm"
)

// PruneProcessedLines deletes processed_lines whose line_hash is not in the activeHashes list.
// If hard is false, lines in StateDownloaded, StateDownloading, or StateOrganizing are preserved.
// If hard is true, all inactive lines are deleted.
func PruneProcessedLines(gdb *gorm.DB, activeHashes []string, hard bool) (int64, error) {
	if gdb == nil {
		gdb = db
	}
	if len(activeHashes) == 0 {
		return 0, fmt.Errorf("active hashes list cannot be empty")
	}

	query := gdb.Model(&models.ProcessedLine{}).Where("line_hash NOT IN ?", activeHashes)
	if !hard {
		query = query.Where("state NOT IN ?", []string{
			string(models.StateDownloaded),
			string(models.StateDownloading),
			string(models.StateOrganizing),
		})
	}

	result := query.Delete(&models.ProcessedLine{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to prune processed lines: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// CleanupOrphanedMetadata deletes movies and tvshows that are no longer referenced by any processed_lines.
// Returns the number of pruned movies, pruned tvshows, and any error.
func CleanupOrphanedMetadata(gdb *gorm.DB) (int64, int64, error) {
	if gdb == nil {
		gdb = db
	}

	var prunedMovies int64
	var prunedTVShows int64

	err := gdb.Transaction(func(tx *gorm.DB) error {
		// Clean up orphaned movies using GORM query builder
		movieResult := tx.Where("id NOT IN (SELECT DISTINCT movie_id FROM processed_lines WHERE movie_id IS NOT NULL)").Delete(&models.Movie{})
		if movieResult.Error != nil {
			return fmt.Errorf("failed to clean up orphaned movies: %w", movieResult.Error)
		}
		prunedMovies = movieResult.RowsAffected

		// Clean up orphaned tvshows using GORM query builder
		tvshowResult := tx.Where("id NOT IN (SELECT DISTINCT tv_show_id FROM processed_lines WHERE tv_show_id IS NOT NULL)").Delete(&models.TVShow{})
		if tvshowResult.Error != nil {
			return fmt.Errorf("failed to clean up orphaned tvshows: %w", tvshowResult.Error)
		}
		prunedTVShows = tvshowResult.RowsAffected

		return nil
	})

	if err != nil {
		return 0, 0, err
	}

	return prunedMovies, prunedTVShows, nil
}

// ResetMovie surgically deletes all processed_lines associated with a specific movie ID.
// Returns the number of deleted processed_lines and any error.
func ResetMovie(gdb *gorm.DB, movieID uint) (int64, error) {
	if gdb == nil {
		gdb = db
	}

	// First verify that the movie exists
	var movie models.Movie
	if err := gdb.First(&movie, movieID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, gorm.ErrRecordNotFound
		}
		return 0, fmt.Errorf("failed to check movie existence: %w", err)
	}

	result := gdb.Where("movie_id = ?", movieID).Delete(&models.ProcessedLine{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete processed lines for movie %d: %w", movieID, result.Error)
	}

	return result.RowsAffected, nil
}

// ResetTVShow surgically deletes all processed_lines associated with a specific TV show ID.
// Returns the number of deleted processed_lines and any error.
func ResetTVShow(gdb *gorm.DB, tvshowID uint) (int64, error) {
	if gdb == nil {
		gdb = db
	}

	// First verify that the TV show exists
	var tvshow models.TVShow
	if err := gdb.First(&tvshow, tvshowID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, gorm.ErrRecordNotFound
		}
		return 0, fmt.Errorf("failed to check TV show existence: %w", err)
	}

	result := gdb.Where("tv_show_id = ?", tvshowID).Delete(&models.ProcessedLine{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete processed lines for TV show %d: %w", tvshowID, result.Error)
	}

	return result.RowsAffected, nil
}
