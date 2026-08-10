package processor

import (
	"fmt"

	"github.com/glefebvre/stalkeer/internal/external/tmdb"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/gorm"
)

// BackfillStats holds the results of a rich-metadata backfill run.
type BackfillStats struct {
	Processed int
	Updated   int
	Errors    int
}

// BackfillRichMetadata queries all Movie and TVShow records that have a tmdb_id
// but are still missing poster_path, fetches their details and external IDs from
// TMDB, and persists poster_path, overview, imdb_id, and tvdb_id. TVShow records
// are deduplicated by tmdb_id so TMDB is called only once per unique show.
func BackfillRichMetadata(db *gorm.DB, client *tmdb.Client, log *logger.Logger) (*BackfillStats, error) {
	stats := &BackfillStats{}

	if err := backfillMoviesMetadata(db, client, log, stats); err != nil {
		return stats, err
	}
	if err := backfillTVShowsMetadata(db, client, log, stats); err != nil {
		return stats, err
	}

	return stats, nil
}

func backfillMoviesMetadata(db *gorm.DB, client *tmdb.Client, log *logger.Logger, stats *BackfillStats) error {
	const batchSize = 100
	offset := 0

	for {
		var movies []models.Movie
		if err := db.Where("tmdb_id != 0 AND poster_path IS NULL").
			Offset(offset).Limit(batchSize).Find(&movies).Error; err != nil {
			return fmt.Errorf("failed to query movies for metadata backfill: %w", err)
		}
		if len(movies) == 0 {
			break
		}

		for i := range movies {
			stats.Processed++
			movie := &movies[i]

			details, err := client.GetMovieDetails(movie.TMDBID)
			if err != nil {
				stats.Errors++
				log.WithFields(map[string]interface{}{
					"movie_id": movie.ID,
					"tmdb_id":  movie.TMDBID,
					"error":    err,
				}).Warn("failed to backfill movie metadata: details lookup failed")
				continue
			}

			externalIDs, err := client.GetMovieExternalIDs(movie.TMDBID)
			if err != nil {
				log.WithFields(map[string]interface{}{
					"movie_id": movie.ID,
					"tmdb_id":  movie.TMDBID,
					"error":    err,
				}).Warn("failed to backfill movie external IDs")
			}

			updates := map[string]interface{}{
				"poster_path": details.PosterPath,
				"overview":    details.Overview,
			}
			if externalIDs != nil {
				updates["imdb_id"] = externalIDs.IMDBID
				if externalIDs.TVDBID != nil {
					updates["tvdb_id"] = externalIDs.TVDBID
				}
			}

			if err := db.Model(movie).Updates(updates).Error; err != nil {
				stats.Errors++
				log.WithFields(map[string]interface{}{
					"movie_id": movie.ID,
					"error":    err,
				}).Warn("failed to persist backfilled movie metadata")
				continue
			}
			stats.Updated++
		}

		offset += batchSize
	}

	return nil
}

func backfillTVShowsMetadata(db *gorm.DB, client *tmdb.Client, log *logger.Logger, stats *BackfillStats) error {
	const batchSize = 100
	offset := 0

	// Cache TMDB ID -> fetched details/external IDs to avoid redundant API calls
	// across rows sharing the same show (different seasons/episodes).
	type cacheEntry struct {
		details     *tmdb.TVShowDetails
		externalIDs *tmdb.ExternalIDs
		hadErr      bool
	}
	cache := make(map[int]cacheEntry)

	for {
		var shows []models.TVShow
		if err := db.Where("tmdb_id != 0 AND poster_path IS NULL").
			Offset(offset).Limit(batchSize).Find(&shows).Error; err != nil {
			return fmt.Errorf("failed to query tvshows for metadata backfill: %w", err)
		}
		if len(shows) == 0 {
			break
		}

		for i := range shows {
			stats.Processed++
			show := &shows[i]

			entry, cached := cache[show.TMDBID]
			if !cached {
				details, err := client.GetTVShowDetails(show.TMDBID)
				if err != nil {
					stats.Errors++
					log.WithFields(map[string]interface{}{
						"tvshow_id": show.ID,
						"tmdb_id":   show.TMDBID,
						"error":     err,
					}).Warn("failed to backfill tvshow metadata: details lookup failed")
					cache[show.TMDBID] = cacheEntry{hadErr: true}
					continue
				}

				externalIDs, err := client.GetTVShowExternalIDs(show.TMDBID)
				if err != nil {
					log.WithFields(map[string]interface{}{
						"tvshow_id": show.ID,
						"tmdb_id":   show.TMDBID,
						"error":     err,
					}).Warn("failed to backfill tvshow external IDs")
				}

				entry = cacheEntry{details: details, externalIDs: externalIDs}
				cache[show.TMDBID] = entry
			}

			if entry.hadErr {
				stats.Errors++
				continue
			}

			updates := map[string]interface{}{
				"poster_path": entry.details.PosterPath,
				"overview":    entry.details.Overview,
			}
			if entry.externalIDs != nil {
				updates["imdb_id"] = entry.externalIDs.IMDBID
				if entry.externalIDs.TVDBID != nil {
					updates["tvdb_id"] = entry.externalIDs.TVDBID
				}
			}

			if err := db.Model(show).Updates(updates).Error; err != nil {
				stats.Errors++
				log.WithFields(map[string]interface{}{
					"tvshow_id": show.ID,
					"error":     err,
				}).Warn("failed to persist backfilled tvshow metadata")
				continue
			}
			stats.Updated++
		}

		offset += batchSize
	}

	return nil
}
