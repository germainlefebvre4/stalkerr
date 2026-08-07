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

// BackfillRichMetadata fetches all Movie and TVShow records that have a tmdb_id
// but a missing poster_path, queries TMDB for details and external IDs, and
// persists poster_path, overview, imdb_id, and tvdb_id. TVShow records are
// deduplicated by tmdb_id so TMDB is queried only once per unique show. A
// per-record TMDB failure is logged and skipped without failing the run.
func BackfillRichMetadata(db *gorm.DB, client *tmdb.Client, log *logger.Logger) (*BackfillStats, error) {
	stats := &BackfillStats{}

	if err := backfillMoviesRichMetadata(db, client, log, stats); err != nil {
		return stats, err
	}
	if err := backfillTVShowsRichMetadata(db, client, log, stats); err != nil {
		return stats, err
	}

	return stats, nil
}

func backfillMoviesRichMetadata(db *gorm.DB, client *tmdb.Client, log *logger.Logger, stats *BackfillStats) error {
	const batchSize = 100
	offset := 0

	for {
		var movies []models.Movie
		if err := db.Where("poster_path IS NULL AND tmdb_id != 0").
			Offset(offset).Limit(batchSize).Find(&movies).Error; err != nil {
			return fmt.Errorf("failed to query movies for backfill: %w", err)
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
				}).Warn("failed to fetch movie details during rich metadata backfill")
				continue
			}

			externalIDs, err := client.GetMovieExternalIDs(movie.TMDBID)
			if err != nil {
				log.WithFields(map[string]interface{}{
					"movie_id": movie.ID,
					"tmdb_id":  movie.TMDBID,
					"error":    err,
				}).Warn("failed to fetch movie external IDs during rich metadata backfill")
			}

			updates := map[string]interface{}{
				"poster_path": details.PosterPath,
				"overview":    details.Overview,
			}
			if externalIDs != nil {
				updates["imdb_id"] = externalIDs.IMDBID
				// Only fill tvdb_id if the record doesn't already have one.
				if externalIDs.TVDBID != nil && movie.TVDBID == nil {
					updates["tvdb_id"] = externalIDs.TVDBID
				}
			}

			if err := db.Model(movie).Updates(updates).Error; err != nil {
				stats.Errors++
				log.WithFields(map[string]interface{}{
					"movie_id": movie.ID,
					"error":    err,
				}).Warn("failed to persist rich metadata backfill for movie")
				continue
			}
			stats.Updated++
		}

		offset += batchSize
	}

	return nil
}

func backfillTVShowsRichMetadata(db *gorm.DB, client *tmdb.Client, log *logger.Logger, stats *BackfillStats) error {
	const batchSize = 100
	offset := 0

	type cacheEntry struct {
		posterPath *string
		overview   string
		imdbID     *string
		tvdbID     *int
		hadErr     bool
	}
	cache := make(map[int]cacheEntry)

	for {
		var shows []models.TVShow
		if err := db.Where("poster_path IS NULL AND tmdb_id != 0").
			Offset(offset).Limit(batchSize).Find(&shows).Error; err != nil {
			return fmt.Errorf("failed to query tvshows for backfill: %w", err)
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
					}).Warn("failed to fetch TV show details during rich metadata backfill")
					cache[show.TMDBID] = cacheEntry{hadErr: true}
					continue
				}

				externalIDs, err := client.GetTVShowExternalIDs(show.TMDBID)
				if err != nil {
					log.WithFields(map[string]interface{}{
						"tvshow_id": show.ID,
						"tmdb_id":   show.TMDBID,
						"error":     err,
					}).Warn("failed to fetch TV show external IDs during rich metadata backfill")
				}

				entry = cacheEntry{posterPath: details.PosterPath, overview: details.Overview}
				if externalIDs != nil {
					entry.imdbID = externalIDs.IMDBID
					entry.tvdbID = externalIDs.TVDBID
				}
				cache[show.TMDBID] = entry
			}

			if entry.hadErr {
				stats.Errors++
				continue
			}

			updates := map[string]interface{}{
				"poster_path": entry.posterPath,
				"overview":    entry.overview,
			}
			if entry.imdbID != nil {
				updates["imdb_id"] = entry.imdbID
			}
			// Only fill tvdb_id if the record doesn't already have one.
			if entry.tvdbID != nil && show.TVDBID == nil {
				updates["tvdb_id"] = entry.tvdbID
			}

			if err := db.Model(show).Updates(updates).Error; err != nil {
				stats.Errors++
				log.WithFields(map[string]interface{}{
					"tvshow_id": show.ID,
					"error":     err,
				}).Warn("failed to persist rich metadata backfill for TV show")
				continue
			}
			stats.Updated++
		}

		offset += batchSize
	}

	return nil
}
