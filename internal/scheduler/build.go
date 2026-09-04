package scheduler

import (
	"context"
	"fmt"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/downloader"
	"github.com/glefebvre/stalkeer/internal/external/radarr"
	"github.com/glefebvre/stalkeer/internal/external/sonarr"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/matcher"
	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/gorm"
)

// BuildDeps bundles the collaborators BuildStreams needs to construct a run's
// full stream set. Radarr/Sonarr clients are accepted as interfaces so tests can
// substitute fakes without spinning up HTTP servers when that isn't needed.
type BuildDeps struct {
	Config       *config.Config
	Radarr       RadarrClient
	Sonarr       SonarrClient
	DB           *gorm.DB
	StateManager *downloader.StateManager

	// MonitoredMovieTMDBIDs and MonitoredSeriesTVDBIDs are an optional
	// snapshot of Radarr's/Sonarr's current monitored status, keyed by
	// Movie.TMDBID / TVDBID, true when monitored. Both are nil-safe: a nil
	// (or missing-key) lookup means "unconfirmed", not "unmonitored", so
	// omitting them preserves BuildStreams' previous behavior entirely.
	// Only mergeIncompleteDownloads consults them (see its doc comment).
	MonitoredMovieTMDBIDs  map[int]bool
	MonitoredSeriesTVDBIDs map[int]bool
}

// RadarrClient is the subset of *radarr.Client BuildStreams depends on.
type RadarrClient interface {
	GetMissingMovies(ctx context.Context, opts radarr.FetchOptions) ([]radarr.Movie, error)
	GetMovieByTMDBID(ctx context.Context, tmdbID int) (*radarr.Movie, error)
}

// SonarrClient is the subset of *sonarr.Client BuildStreams depends on.
type SonarrClient interface {
	GetMissingEpisodes(ctx context.Context, opts sonarr.FetchOptions) ([]sonarr.Episode, error)
	GetSeriesDetails(ctx context.Context, id int) (*sonarr.Series, error)
	GetSeriesByTVDBID(ctx context.Context, tvdbID int) (*sonarr.Series, error)
}

type tvdbSeasonKey struct {
	tvdbID int
	season int
}

// BuildStreams fetches missing movies (Radarr) and missing episodes (Sonarr),
// matches them against the local database, classifies tier-1 (missing) and
// tier-2 (already-downloaded, upgrade-eligible) candidates, and merges any
// incomplete/interrupted downloads into their owning stream. Stale download
// locks are cleaned up before any stream is constructed.
//
// It returns the full, flat stream set for the run: movie streams, every
// series' season streams (not just the earliest — NewScheduler is responsible
// for exposing only the earliest per series as claimable), and tier-2 streams.
func BuildStreams(ctx context.Context, deps BuildDeps) ([]*Stream, error) {
	if err := deps.StateManager.CleanupStaleLocks(ctx); err != nil {
		return nil, err
	}

	var streams []*Stream

	movieStreams, movieByID, err := buildTier1MovieStreams(ctx, deps)
	if err != nil {
		return nil, err
	}
	streams = append(streams, movieStreams...)

	seriesStreams, tvdbSeasonStreams, err := buildTier1SeriesStreams(ctx, deps)
	if err != nil {
		return nil, err
	}
	streams = append(streams, seriesStreams...)

	tier2MovieStreams, err := buildTier2MovieStreams(ctx, deps)
	if err != nil {
		return nil, err
	}

	tier2SeriesStreams, err := buildTier2SeriesStreams(ctx, deps)
	if err != nil {
		return nil, err
	}

	extraStreams, err := mergeIncompleteDownloads(ctx, deps, movieByID, tvdbSeasonStreams)
	if err != nil {
		return nil, err
	}
	streams = append(streams, extraStreams...)

	// movieByID and tvdbSeasonStreams may have gained entries from
	// mergeIncompleteDownloads synthesizing a new tier-1 stream, so this must
	// run after step 5 to see the final tier-1 set.
	streams = append(streams, dedupTier2MovieStreams(tier2MovieStreams, movieByID)...)
	streams = append(streams, dedupTier2SeriesStreams(tier2SeriesStreams, tvdbSeasonStreams)...)

	return streams, nil
}

// dedupTier2MovieStreams drops any tier-2 movie stream whose movie already has
// a tier-1 stream this run - tier-1 always wins for a given work unit.
func dedupTier2MovieStreams(tier2 []*Stream, movieByID map[uint]*Stream) []*Stream {
	tier1Keys := make(map[string]bool, len(movieByID))
	for id := range movieByID {
		tier1Keys[fmt.Sprintf("movie:%d", id)] = true
	}

	var kept []*Stream
	for _, s := range tier2 {
		if !tier1Keys[s.SourceKey] {
			kept = append(kept, s)
		}
	}
	return kept
}

// dedupTier2SeriesStreams drops any tier-2 series-season stream whose
// (tvdbID, season) already has a tier-1 stream this run. Tier-1 and tier-2
// series streams use different SourceKey formats (Sonarr's internal series ID,
// shared across a series' seasons, vs TVDB ID+season, one per season - see
// design.md), so matching happens via the shared tvdbSeasonKey identity,
// formatted the way tier-2 encodes it in SourceKey.
func dedupTier2SeriesStreams(tier2 []*Stream, tvdbSeasonStreams map[tvdbSeasonKey]*Stream) []*Stream {
	tier1Keys := make(map[string]bool, len(tvdbSeasonStreams))
	for key := range tvdbSeasonStreams {
		tier1Keys[fmt.Sprintf("series:tvdb:%d:season:%d", key.tvdbID, key.season)] = true
	}

	var kept []*Stream
	for _, s := range tier2 {
		if !tier1Keys[s.SourceKey] {
			kept = append(kept, s)
		}
	}
	return kept
}

func buildTier1MovieStreams(ctx context.Context, deps BuildDeps) ([]*Stream, map[uint]*Stream, error) {
	if deps.Radarr == nil {
		return nil, make(map[uint]*Stream), nil
	}

	missing, err := deps.Radarr.GetMissingMovies(ctx, radarr.FetchOptions{})
	if err != nil {
		return nil, nil, err
	}

	var streams []*Stream
	byID := make(map[uint]*Stream)

	for _, movie := range missing {
		dbMovie, _, _, err := matcher.MatchMovieByTVDB(deps.DB, movie.TvdbID, movie.TMDBID, movie.Title, movie.Year)
		if err != nil {
			continue
		}

		candidates, err := matcher.FindMovieDownloadCandidates(deps.DB, dbMovie.ID)
		if err != nil {
			return nil, nil, err
		}
		if len(candidates) == 0 {
			continue
		}

		baseDestPath, _ := downloader.BuildRadarrDestPath(movie.Path, deps.Config.Downloads.MoviesPath, movie.Title, movie.Year)
		stream := &Stream{
			Tier:      Tier1,
			SourceKey: fmt.Sprintf("movie:%d", dbMovie.ID),
			Items: []Item{{
				DisplayName: fmt.Sprintf("%s (%d)", movie.Title, movie.Year),
				BaseDestDir: baseDestPath,
				Candidates:  candidates,
			}},
		}
		streams = append(streams, stream)
		byID[dbMovie.ID] = stream
	}

	return streams, byID, nil
}

type seasonBuild struct {
	seriesID int
	season   int
	tvdbID   int
	items    []Item
}

func buildTier1SeriesStreams(ctx context.Context, deps BuildDeps) ([]*Stream, map[tvdbSeasonKey]*Stream, error) {
	if deps.Sonarr == nil {
		return nil, make(map[tvdbSeasonKey]*Stream), nil
	}

	missing, err := deps.Sonarr.GetMissingEpisodes(ctx, sonarr.FetchOptions{})
	if err != nil {
		return nil, nil, err
	}

	seriesCache := make(map[int]*sonarr.Series)
	seasonMap := make(map[tvdbSeasonKey]*seasonBuild)
	var seasonOrder []tvdbSeasonKey

	for _, ep := range missing {
		series, ok := seriesCache[ep.SeriesID]
		if !ok {
			s, err := deps.Sonarr.GetSeriesDetails(ctx, ep.SeriesID)
			if err != nil {
				continue
			}
			series = s
			seriesCache[ep.SeriesID] = series
		}

		dbShow, _, _, err := matcher.MatchTVShowByTVDB(deps.DB, series.TvdbID, 0, series.Title, ep.SeasonNumber, ep.EpisodeNumber)
		if err != nil {
			continue
		}

		candidates, err := matcher.FindTVShowDownloadCandidates(deps.DB, dbShow.ID)
		if err != nil {
			return nil, nil, err
		}
		if len(candidates) == 0 {
			continue
		}

		baseDestPath, _ := downloader.BuildSonarrDestPath(
			series.Path, deps.Config.Downloads.TVShowsPath, series.Title, series.Year, ep.SeasonNumber, ep.EpisodeNumber,
		)

		key := tvdbSeasonKey{tvdbID: series.TvdbID, season: ep.SeasonNumber}
		sb, ok := seasonMap[key]
		if !ok {
			sb = &seasonBuild{seriesID: ep.SeriesID, season: ep.SeasonNumber, tvdbID: series.TvdbID}
			seasonMap[key] = sb
			seasonOrder = append(seasonOrder, key)
		}
		sb.items = append(sb.items, Item{
			DisplayName: fmt.Sprintf("%s S%02dE%02d", series.Title, ep.SeasonNumber, ep.EpisodeNumber),
			BaseDestDir: baseDestPath,
			Candidates:  candidates,
			Episode:     ep.EpisodeNumber,
		})
	}

	var streams []*Stream
	byTVDBSeason := make(map[tvdbSeasonKey]*Stream)
	for _, key := range seasonOrder {
		sb := seasonMap[key]
		SortSeasonItems(sb.items)
		stream := &Stream{
			Tier:      Tier1,
			SourceKey: fmt.Sprintf("series:%d", sb.seriesID),
			SeriesID:  sb.seriesID,
			Season:    sb.season,
			Items:     sb.items,
		}
		streams = append(streams, stream)
		byTVDBSeason[key] = stream
	}

	return streams, byTVDBSeason, nil
}

// buildTier2MovieStreams classifies movies that already have a successful
// download but whose best untried candidate is strictly better (language,
// then resolution) than the one already downloaded, as tier-2 upgrade
// streams. A leftover candidate that is merely eligible but equal-or-worse no
// longer qualifies. The destination path is sourced from a live Radarr
// lookup, matching tier-1, so an accepted upgrade replaces the original file
// instead of landing in a different folder.
func buildTier2MovieStreams(ctx context.Context, deps BuildDeps) ([]*Stream, error) {
	var movies []models.Movie
	if err := deps.DB.Find(&movies).Error; err != nil {
		return nil, err
	}

	var streams []*Stream
	for _, movie := range movies {
		downloadedLine, err := findDownloadedLine(deps.DB, "movie_id", movie.ID)
		if err != nil {
			return nil, err
		}
		if downloadedLine == nil {
			continue
		}

		candidates, err := matcher.FindMovieDownloadCandidates(deps.DB, movie.ID)
		if err != nil {
			return nil, err
		}
		if len(candidates) == 0 {
			continue
		}
		if candidateRank(&candidates[0]) >= candidateRank(downloadedLine) {
			continue
		}

		moviePath := ""
		if deps.Radarr != nil {
			radarrMovie, err := deps.Radarr.GetMovieByTMDBID(ctx, movie.TMDBID)
			if err != nil {
				logger.AppLogger().WithFields(map[string]interface{}{
					"movie_id": movie.ID,
					"tmdb_id":  movie.TMDBID,
					"error":    err,
				}).Warn("tier-2: failed to look up movie in Radarr, skipping upgrade for this run")
				continue
			}
			if radarrMovie != nil {
				moviePath = radarrMovie.Path
			}
		}

		baseDestPath, _ := downloader.BuildRadarrDestPath(moviePath, deps.Config.Downloads.MoviesPath, movie.TMDBTitle, movie.TMDBYear)
		streams = append(streams, &Stream{
			Tier:      Tier2,
			SourceKey: fmt.Sprintf("movie:%d", movie.ID),
			Items: []Item{{
				DisplayName: fmt.Sprintf("%s (%d)", movie.TMDBTitle, movie.TMDBYear),
				BaseDestDir: baseDestPath,
				Candidates:  candidates,
			}},
		})
	}

	return streams, nil
}

// buildTier2SeriesStreams classifies TV episodes that already have a
// successful download but whose best untried candidate is strictly better
// (language, then resolution) than the one already downloaded, as tier-2
// upgrade streams, grouped one stream per (series, season). Tier-2 streams do
// not participate in the ascending-season-order queue: every one of their
// episodes is already in a terminal (downloaded) state, so there is no
// in-progress season to protect. The destination path is sourced from a live
// Sonarr lookup, matching tier-1, so an accepted upgrade replaces the
// original file instead of landing in a different folder.
func buildTier2SeriesStreams(ctx context.Context, deps BuildDeps) ([]*Stream, error) {
	var episodes []models.TVShow
	if err := deps.DB.Find(&episodes).Error; err != nil {
		return nil, err
	}

	seasonMap := make(map[tvdbSeasonKey]*seasonBuild)
	var seasonOrder []tvdbSeasonKey
	seriesPathCache := make(map[int]string)

	for _, ep := range episodes {
		if ep.TVDBID == nil || ep.Season == nil || ep.Episode == nil {
			continue
		}

		downloadedLine, err := findDownloadedLine(deps.DB, "tv_show_id", ep.ID)
		if err != nil {
			return nil, err
		}
		if downloadedLine == nil {
			continue
		}

		candidates, err := matcher.FindTVShowDownloadCandidates(deps.DB, ep.ID)
		if err != nil {
			return nil, err
		}
		if len(candidates) == 0 {
			continue
		}
		if candidateRank(&candidates[0]) >= candidateRank(downloadedLine) {
			continue
		}

		seriesPath, ok := seriesPathCache[*ep.TVDBID]
		if !ok {
			if deps.Sonarr != nil {
				series, err := deps.Sonarr.GetSeriesByTVDBID(ctx, *ep.TVDBID)
				if err != nil {
					logger.AppLogger().WithFields(map[string]interface{}{
						"tvdb_id": *ep.TVDBID,
						"error":   err,
					}).Warn("tier-2: failed to look up series in Sonarr, skipping upgrade for this run")
					continue
				}
				if series != nil {
					seriesPath = series.Path
				}
			}
			seriesPathCache[*ep.TVDBID] = seriesPath
		}

		baseDestPath, _ := downloader.BuildSonarrDestPath(seriesPath, deps.Config.Downloads.TVShowsPath, ep.TMDBTitle, ep.TMDBYear, *ep.Season, *ep.Episode)

		key := tvdbSeasonKey{tvdbID: *ep.TVDBID, season: *ep.Season}
		sb, ok := seasonMap[key]
		if !ok {
			sb = &seasonBuild{season: *ep.Season, tvdbID: *ep.TVDBID}
			seasonMap[key] = sb
			seasonOrder = append(seasonOrder, key)
		}
		sb.items = append(sb.items, Item{
			DisplayName: fmt.Sprintf("%s (%d) - S%02dE%02d", ep.TMDBTitle, ep.TMDBYear, *ep.Season, *ep.Episode),
			BaseDestDir: baseDestPath,
			Candidates:  candidates,
			Episode:     *ep.Episode,
		})
	}

	var streams []*Stream
	for _, key := range seasonOrder {
		sb := seasonMap[key]
		SortSeasonItems(sb.items)
		streams = append(streams, &Stream{
			Tier:      Tier2,
			SourceKey: fmt.Sprintf("series:tvdb:%d:season:%d", key.tvdbID, key.season),
			Season:    sb.season,
			Items:     sb.items,
		})
	}

	return streams, nil
}

// findDownloadedLine returns the best-ranked already-downloaded ProcessedLine
// for the given movie/TV-show occurrence (by candidateRank), or nil if none
// has been downloaded yet. When more than one occurrence has been downloaded,
// the best-ranked one represents what the user already effectively has, so
// tier-2 gating compares untried candidates against it rather than an
// arbitrary one.
func findDownloadedLine(db *gorm.DB, column string, id uint) (*models.ProcessedLine, error) {
	var lines []models.ProcessedLine
	if err := db.Where(fmt.Sprintf("%s = ? AND state = ?", column), id, models.StateDownloaded).Find(&lines).Error; err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, nil
	}

	best := &lines[0]
	for i := 1; i < len(lines); i++ {
		if candidateRank(&lines[i]) < candidateRank(best) {
			best = &lines[i]
		}
	}
	return best, nil
}

// mergeIncompleteDownloads attaches any incomplete/interrupted DownloadInfo
// record to the item it belongs to (matched by movie, or by series+season+
// episode), so it is resumed as the next item attempted in its own stream
// rather than requiring a separate resume command.
//
// A crashed download leaves its ProcessedLine stuck in the "downloading"
// state, which falls outside the "processed"/"failed" states the normal
// missing-content matcher requires — so the owning stream may not have been
// constructed by buildTier1*Streams at all. When that happens here, a stream
// is synthesized directly from the DownloadInfo's preloaded Movie/TVShow
// association instead of being dropped. A synthesized series stream has no
// known Sonarr series ID, so (uniquely among tier-1 streams) it does not
// participate in the ascending-season-order queue; this is an accepted,
// narrow trade-off for an edge case outside the normal fetch path.
//
// An incomplete download whose movie/series is confirmed unmonitored in
// Radarr/Sonarr (deps.MonitoredMovieTMDBIDs/MonitoredSeriesTVDBIDs) is left
// out of this run's stream set entirely instead of being resumed. A miss in
// either map (nil map, movie/series not currently in the fetched library, or
// the library fetch failed this run) is always treated as "unconfirmed",
// never as "unmonitored", so it is still resumed as before.
func mergeIncompleteDownloads(ctx context.Context, deps BuildDeps, movieByID map[uint]*Stream, seriesByTVDBSeason map[tvdbSeasonKey]*Stream) ([]*Stream, error) {
	incomplete, err := deps.StateManager.GetIncompleteDownloads(ctx, deps.Config.Downloads.MaxRetryAttempts, 0)
	if err != nil {
		return nil, err
	}

	var extra []*Stream
	touchedSeasons := make(map[tvdbSeasonKey]bool)

	for i := range incomplete {
		download := &incomplete[i]
		if len(download.ProcessedLines) == 0 {
			continue
		}
		line := &download.ProcessedLines[0]

		switch line.ContentType {
		case models.ContentTypeMovies:
			if line.Movie == nil {
				continue
			}
			if monitored, known := deps.MonitoredMovieTMDBIDs[line.Movie.TMDBID]; known && !monitored {
				// Confirmed no longer monitored in Radarr: don't resume it,
				// whether or not a tier-1 stream already exists for it.
				continue
			}
			stream, ok := movieByID[line.Movie.ID]
			if !ok {
				candidates, err := matcher.FindMovieDownloadCandidates(deps.DB, line.Movie.ID)
				if err != nil {
					return nil, err
				}
				baseDestPath, _ := downloader.BuildRadarrDestPath("", deps.Config.Downloads.MoviesPath, line.Movie.TMDBTitle, line.Movie.TMDBYear)
				stream = &Stream{
					Tier:      Tier1,
					SourceKey: fmt.Sprintf("movie:%d", line.Movie.ID),
					Items: []Item{{
						DisplayName: fmt.Sprintf("%s (%d)", line.Movie.TMDBTitle, line.Movie.TMDBYear),
						BaseDestDir: baseDestPath,
						Candidates:  candidates,
					}},
				}
				movieByID[line.Movie.ID] = stream
				extra = append(extra, stream)
			}
			attachResume(&stream.Items[0], line, download)

		case models.ContentTypeTVShows:
			show := line.TVShow
			if show == nil || show.TVDBID == nil || show.Season == nil || show.Episode == nil {
				continue
			}
			if monitored, known := deps.MonitoredSeriesTVDBIDs[*show.TVDBID]; known && !monitored {
				// Confirmed no longer monitored in Sonarr: don't resume any
				// of this series' incomplete episodes.
				continue
			}
			key := tvdbSeasonKey{tvdbID: *show.TVDBID, season: *show.Season}
			stream, ok := seriesByTVDBSeason[key]
			if !ok {
				candidates, err := matcher.FindTVShowDownloadCandidates(deps.DB, show.ID)
				if err != nil {
					return nil, err
				}
				baseDestPath, _ := downloader.BuildSonarrDestPath("", deps.Config.Downloads.TVShowsPath, show.TMDBTitle, show.TMDBYear, *show.Season, *show.Episode)
				stream = &Stream{
					Tier:      Tier1,
					SourceKey: fmt.Sprintf("series:tvdb:%d", *show.TVDBID),
					Season:    *show.Season,
					Items: []Item{{
						DisplayName: fmt.Sprintf("%s (%d) - S%02dE%02d", show.TMDBTitle, show.TMDBYear, *show.Season, *show.Episode),
						BaseDestDir: baseDestPath,
						Candidates:  candidates,
						Episode:     *show.Episode,
					}},
				}
				seriesByTVDBSeason[key] = stream
				extra = append(extra, stream)
				attachResume(&stream.Items[0], line, download)
				continue
			}

			found := false
			for idx := range stream.Items {
				if stream.Items[idx].Episode == *show.Episode {
					attachResume(&stream.Items[idx], line, download)
					found = true
					break
				}
			}
			if !found {
				candidates, err := matcher.FindTVShowDownloadCandidates(deps.DB, show.ID)
				if err != nil {
					return nil, err
				}
				baseDestPath, _ := downloader.BuildSonarrDestPath("", deps.Config.Downloads.TVShowsPath, show.TMDBTitle, show.TMDBYear, *show.Season, *show.Episode)
				item := Item{
					DisplayName: fmt.Sprintf("%s (%d) - S%02dE%02d", show.TMDBTitle, show.TMDBYear, *show.Season, *show.Episode),
					BaseDestDir: baseDestPath,
					Candidates:  candidates,
					Episode:     *show.Episode,
				}
				attachResume(&item, line, download)
				stream.Items = append(stream.Items, item)
			}
			touchedSeasons[key] = true
		}
	}

	for key := range touchedSeasons {
		SortSeasonItems(seriesByTVDBSeason[key].Items)
	}

	return extra, nil
}

// attachResume marks item as resumable and ensures the specific ProcessedLine
// tied to the incomplete download is among its candidates: that line's own
// state (e.g. stuck "downloading" from a crashed run) may fall outside
// FindMovieDownloadCandidates/FindTVShowDownloadCandidates' normal state
// filter, so without this it could otherwise be resumed-marked but have
// nothing to actually attempt.
func attachResume(item *Item, line *models.ProcessedLine, download *models.DownloadInfo) {
	item.ResumeInfo = download
	for _, c := range item.Candidates {
		if c.ID == line.ID {
			return
		}
	}
	item.Candidates = append([]models.ProcessedLine{*line}, item.Candidates...)
}

// ApplyLimit caps the number of distinct work units (movies or TV series)
// considered for this run, applied before the random draw so it still
// respects season-order/tier rules on the remaining subset: an entire series'
// season chain is kept or dropped together, never split.
func ApplyLimit(streams []*Stream, limit int) []*Stream {
	if limit <= 0 {
		return streams
	}

	order := make([]string, 0)
	seen := make(map[string]bool)
	for _, s := range streams {
		if !seen[s.SourceKey] {
			seen[s.SourceKey] = true
			order = append(order, s.SourceKey)
		}
	}
	if len(order) > limit {
		order = order[:limit]
	}

	allowed := make(map[string]bool, len(order))
	for _, key := range order {
		allowed[key] = true
	}

	var result []*Stream
	for _, s := range streams {
		if allowed[s.SourceKey] {
			result = append(result, s)
		}
	}
	return result
}
