package api

import (
	"context"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/external/radarr"
	"github.com/glefebvre/stalkeer/internal/external/sonarr"
	"github.com/glefebvre/stalkeer/internal/matcher"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/retry"
)

// radarrSonarrDefaultPageSize/radarrSonarrMaxPageSize bound these endpoints' page
// size much tighter than the app's general defaultLimit/maxLimit: unlike a plain
// DB list, a Sonarr page fetches one extra live episode list per series, so an
// unbounded page size would fan out into an unbounded number of upstream calls.
const (
	radarrSonarrDefaultPageSize = 20
	radarrSonarrMaxPageSize     = 50
)

// RadarrMovieListItem is one row of the Radarr monitoring list.
type RadarrMovieListItem struct {
	RadarrID        int    `json:"radarr_id"`
	Title           string `json:"title"`
	Year            int    `json:"year"`
	HasFile         bool   `json:"has_file"`
	Matched         bool   `json:"matched"`
	MovieID         *uint  `json:"movie_id,omitempty"`
	OccurrenceCount int    `json:"occurrence_count"`
}

// SonarrSeriesListItem is one row of the Sonarr monitoring list, aggregated per series.
type SonarrSeriesListItem struct {
	SonarrID        int    `json:"sonarr_id"`
	Title           string `json:"title"`
	Year            int    `json:"year"`
	MatchedCount    int    `json:"matched_count"`
	MonitoredCount  int    `json:"monitored_count"`
	OccurrenceCount int    `json:"occurrence_count"`
}

// matchStatusFilter parses the "matched"/"no_match" match-status filter query
// parameter shared by the Radarr and Sonarr listing endpoints. Any other value
// (including absent/empty) is treated as no filter.
func matchStatusFilter(c *gin.Context) string {
	switch v := strings.TrimSpace(c.Query("filter")); v {
	case "matched", "no_match":
		return v
	default:
		return ""
	}
}

// OccurrenceResponse is one playlist occurrence, regardless of pipeline state.
type OccurrenceResponse struct {
	ID         uint    `json:"id"`
	Resolution *string `json:"resolution,omitempty"`
	State      string  `json:"state"`
}

// RadarrMovieMatchesResponse is the sidepanel detail for a single Radarr movie.
type RadarrMovieMatchesResponse struct {
	Matched     bool                 `json:"matched"`
	Movie       *MovieResponse       `json:"movie,omitempty"`
	Occurrences []OccurrenceResponse `json:"occurrences"`
}

// SonarrSeriesEpisodeItem is one monitored episode's match detail, underlying a
// series' aggregate ratio in the list view.
type SonarrSeriesEpisodeItem struct {
	Season      int                  `json:"season"`
	Episode     int                  `json:"episode"`
	Matched     bool                 `json:"matched"`
	Occurrences []OccurrenceResponse `json:"occurrences"`
}

// SonarrSeriesEpisodesResponse is the sidepanel detail for a single Sonarr series.
type SonarrSeriesEpisodesResponse struct {
	Episodes []SonarrSeriesEpisodeItem `json:"episodes"`
}

// listRadarrMonitoredMovies handles GET /api/v1/radarr/movies?limit&offset&filter&search.
// Pagination happens before matching: the full Radarr movie list is fetched once,
// filtered to monitored movies and sorted, then only the requested page's movies
// are matched against the local playlist database. When an optional match-status
// `filter` (matched/no_match) is supplied, match status is instead computed for
// the entire monitored (post-search) list before filtering and paginating, so the
// filtered total and page are accurate - this is local-DB-only and cheap, mirroring
// what the /stats endpoint already does. See
// openspec/changes/radarr-sonarr-monitoring-view and
// openspec/changes/radarr-sonarr-match-filters-and-counts.
func (s *Server) listRadarrMonitoredMovies(c *gin.Context) {
	cfg := config.Get()
	if cfg.Radarr.URL == "" || cfg.Radarr.APIKey == "" {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "radarr_not_configured",
			Message: "Radarr is not configured",
		})
		return
	}

	limit, offset := parsePaginationBounded(c, radarrSonarrDefaultPageSize, radarrSonarrMaxPageSize)

	ctx, cancel := context.WithTimeout(c.Request.Context(), existenceCheckTimeout)
	defer cancel()

	client := radarr.New(radarr.Config{
		BaseURL:     cfg.Radarr.URL,
		APIKey:      cfg.Radarr.APIKey,
		Timeout:     existenceCheckTimeout,
		RetryConfig: retry.Config{MaxAttempts: 1},
	})

	allMovies, err := client.GetAllMovies(ctx)
	if err != nil {
		c.JSON(http.StatusBadGateway, ErrorResponse{
			Error:   "radarr_unreachable",
			Message: "failed to reach Radarr",
		})
		return
	}

	search := strings.ToLower(strings.TrimSpace(c.Query("search")))

	monitored := make([]radarr.Movie, 0, len(allMovies))
	for _, m := range allMovies {
		if !m.Monitored {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(m.Title), search) {
			continue
		}
		monitored = append(monitored, m)
	}
	sort.Slice(monitored, func(i, j int) bool {
		return strings.ToLower(monitored[i].Title) < strings.ToLower(monitored[j].Title)
	})

	db := database.Get()
	filter := matchStatusFilter(c)

	var total int
	var pageMovies []radarr.Movie
	var matches map[int]matcher.MovieMatchResult

	if filter != "" {
		allMatches, err := matcher.MatchMoviesBatch(db, monitored)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "database_error",
				Message: "failed to compute playlist match status",
			})
			return
		}

		filtered := make([]radarr.Movie, 0, len(monitored))
		for _, m := range monitored {
			matched := allMatches[m.ID].Matched
			if (filter == "matched" && matched) || (filter == "no_match" && !matched) {
				filtered = append(filtered, m)
			}
		}

		total = len(filtered)
		pageMovies = sliceRadarrMoviesPage(filtered, offset, limit)
		matches = allMatches
	} else {
		total = len(monitored)
		pageMovies = sliceRadarrMoviesPage(monitored, offset, limit)

		pageMatches, err := matcher.MatchMoviesBatch(db, pageMovies)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "database_error",
				Message: "failed to compute playlist match status",
			})
			return
		}
		matches = pageMatches
	}

	matchedMovieIDs := make([]uint, 0, len(pageMovies))
	for _, m := range pageMovies {
		result := matches[m.ID]
		if result.Matched && result.Movie != nil {
			matchedMovieIDs = append(matchedMovieIDs, result.Movie.ID)
		}
	}
	occurrenceCounts, err := matcher.CountMovieOccurrencesBatch(db, matchedMovieIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "failed to compute playlist occurrence counts",
		})
		return
	}

	items := make([]RadarrMovieListItem, len(pageMovies))
	for i, m := range pageMovies {
		result := matches[m.ID]
		item := RadarrMovieListItem{
			RadarrID: m.ID,
			Title:    m.Title,
			Year:     m.Year,
			HasFile:  m.HasFile,
			Matched:  result.Matched,
		}
		if result.Matched && result.Movie != nil {
			movieID := result.Movie.ID
			item.MovieID = &movieID
			item.OccurrenceCount = occurrenceCounts[movieID]
		}
		items[i] = item
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       items,
		Total:      int64(total),
		Limit:      limit,
		Offset:     offset,
		TotalPages: totalPagesFor(total, limit),
	})
}

// listSonarrMonitoredSeries handles GET /api/v1/sonarr/series?limit&offset&filter&search&refresh.
// Pagination happens before the per-series episode fetch/matching: the full
// monitored series list is fetched once, sorted, then only the requested page's
// series get their monitored episodes fetched from Sonarr (concurrently, bounded
// by the page size) and matched against the local playlist database.
//
// When an optional match-status `filter` (matched/no_match) is supplied, every
// monitored (post-search) series' matched status is instead resolved via the
// Sonarr match-status cache (populating it on a miss) before filtering and
// paginating, avoiding a full per-series episode fan-out on every filtered
// request. A truthy `refresh` query parameter (set by the Séries section's
// manual refresh action) clears that cache first. See
// openspec/changes/radarr-sonarr-match-filters-and-counts.
func (s *Server) listSonarrMonitoredSeries(c *gin.Context) {
	cfg := config.Get()
	if cfg.Sonarr.URL == "" || cfg.Sonarr.APIKey == "" {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "sonarr_not_configured",
			Message: "Sonarr is not configured",
		})
		return
	}

	if refresh, _ := strconv.ParseBool(c.Query("refresh")); refresh {
		sonarrMatchCache.clear()
	}

	limit, offset := parsePaginationBounded(c, radarrSonarrDefaultPageSize, radarrSonarrMaxPageSize)
	filter := matchStatusFilter(c)

	ctx, cancel := context.WithTimeout(c.Request.Context(), existenceCheckTimeout)
	defer cancel()

	client := sonarr.New(sonarr.Config{
		BaseURL:     cfg.Sonarr.URL,
		APIKey:      cfg.Sonarr.APIKey,
		Timeout:     existenceCheckTimeout,
		RetryConfig: retry.Config{MaxAttempts: 1},
	})

	allSeries, err := client.GetAllMonitoredSeries(ctx)
	if err != nil {
		c.JSON(http.StatusBadGateway, ErrorResponse{
			Error:   "sonarr_unreachable",
			Message: "failed to reach Sonarr",
		})
		return
	}

	search := strings.ToLower(strings.TrimSpace(c.Query("search")))
	if search != "" {
		filtered := make([]sonarr.Series, 0, len(allSeries))
		for _, s := range allSeries {
			if strings.Contains(strings.ToLower(s.Title), search) {
				filtered = append(filtered, s)
			}
		}
		allSeries = filtered
	}

	sort.Slice(allSeries, func(i, j int) bool {
		return strings.ToLower(allSeries[i].Title) < strings.ToLower(allSeries[j].Title)
	})

	db := database.Get()

	var total int
	var pageSeries []sonarr.Series

	if filter != "" {
		type statusResult struct {
			matched bool
			err     error
		}
		statuses := make([]statusResult, len(allSeries))
		var wg sync.WaitGroup
		for i, series := range allSeries {
			wg.Add(1)
			go func(i int, series sonarr.Series) {
				defer wg.Done()
				matched, err := sonarrMatchCache.matchedStatus(ctx, client, db, series)
				statuses[i] = statusResult{matched: matched, err: err}
			}(i, series)
		}
		wg.Wait()

		filtered := make([]sonarr.Series, 0, len(allSeries))
		for i, series := range allSeries {
			if statuses[i].err != nil {
				c.JSON(http.StatusBadGateway, ErrorResponse{
					Error:   "sonarr_unreachable",
					Message: "failed to fetch episodes from Sonarr for one or more series",
				})
				return
			}
			matched := statuses[i].matched
			if (filter == "matched" && matched) || (filter == "no_match" && !matched) {
				filtered = append(filtered, series)
			}
		}

		total = len(filtered)
		pageSeries = sliceSonarrSeriesPage(filtered, offset, limit)
	} else {
		total = len(allSeries)
		pageSeries = sliceSonarrSeriesPage(allSeries, offset, limit)
	}

	type detailResult struct {
		details []matcher.SeriesEpisodeMatch
		err     error
	}
	details := make([]detailResult, len(pageSeries))

	var wg sync.WaitGroup
	for i, series := range pageSeries {
		wg.Add(1)
		go func(i int, series sonarr.Series) {
			defer wg.Done()

			episodes, err := client.GetEpisodesBySeriesID(ctx, series.ID)
			if err != nil {
				details[i].err = err
				return
			}

			var monitoredEpisodes []matcher.SeasonEpisode
			for _, ep := range episodes {
				if ep.Monitored {
					monitoredEpisodes = append(monitoredEpisodes, matcher.SeasonEpisode{
						Season:  ep.SeasonNumber,
						Episode: ep.EpisodeNumber,
					})
				}
			}

			result, err := matcher.MatchSeriesEpisodesDetail(db, series.TvdbID, monitoredEpisodes)
			if err != nil {
				details[i].err = err
				return
			}
			details[i].details = result
		}(i, series)
	}
	wg.Wait()

	items := make([]SonarrSeriesListItem, len(pageSeries))
	for i, series := range pageSeries {
		if details[i].err != nil {
			c.JSON(http.StatusBadGateway, ErrorResponse{
				Error:   "sonarr_unreachable",
				Message: "failed to fetch episodes from Sonarr for one or more series",
			})
			return
		}

		matchedCount := 0
		occurrenceCount := 0
		for _, ep := range details[i].details {
			if ep.Matched {
				matchedCount++
				occurrenceCount += len(ep.Occurrences)
			}
		}

		items[i] = SonarrSeriesListItem{
			SonarrID:        series.ID,
			Title:           series.Title,
			Year:            series.Year,
			MatchedCount:    matchedCount,
			MonitoredCount:  len(details[i].details),
			OccurrenceCount: occurrenceCount,
		}
	}

	c.JSON(http.StatusOK, PaginatedResponse{
		Data:       items,
		Total:      int64(total),
		Limit:      limit,
		Offset:     offset,
		TotalPages: totalPagesFor(total, limit),
	})
}

// RadarrSonarrStatsResponse is the full-catalog monitoring summary for the
// Résumé sub-tab. Each service is reported independently: if one upstream is
// unreachable or unconfigured, its counts are left nil and its *_error field
// explains why, while the other service's counts are still returned normally.
type RadarrSonarrStatsResponse struct {
	RadarrMonitored *int   `json:"radarr_monitored"`
	RadarrMatched   *int   `json:"radarr_matched"`
	RadarrError     string `json:"radarr_error,omitempty"`
	SonarrMonitored *int   `json:"sonarr_monitored"`
	SonarrError     string `json:"sonarr_error,omitempty"`
}

// listRadarrSonarrStats handles GET /api/v1/radarr-sonarr/stats, reporting
// full-catalog Radarr matched/unmatched counts and the Sonarr monitored-series
// total for the Résumé sub-tab. Unlike the paginated listing endpoints, the
// Radarr matched count is computed over the entire monitored list in a single
// batch: Radarr matching only touches the local DB, so this is one query, not
// an upstream fan-out. The Sonarr total intentionally skips per-series episode
// fetches (see radarr-sonarr-monitoring-api spec).
func (s *Server) listRadarrSonarrStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), existenceCheckTimeout)
	defer cancel()

	cfg := config.Get()
	resp := RadarrSonarrStatsResponse{}

	if cfg.Radarr.URL == "" || cfg.Radarr.APIKey == "" {
		resp.RadarrError = "radarr_not_configured"
	} else {
		radarrClient := radarr.New(radarr.Config{
			BaseURL:     cfg.Radarr.URL,
			APIKey:      cfg.Radarr.APIKey,
			Timeout:     existenceCheckTimeout,
			RetryConfig: retry.Config{MaxAttempts: 1},
		})

		if allMovies, err := radarrClient.GetAllMovies(ctx); err != nil {
			resp.RadarrError = "radarr_unreachable"
		} else {
			monitored := make([]radarr.Movie, 0, len(allMovies))
			for _, m := range allMovies {
				if m.Monitored {
					monitored = append(monitored, m)
				}
			}

			db := database.Get()
			matches, err := matcher.MatchMoviesBatch(db, monitored)
			if err != nil {
				resp.RadarrError = "database_error"
			} else {
				matchedCount := 0
				for _, m := range monitored {
					if matches[m.ID].Matched {
						matchedCount++
					}
				}
				monitoredCount := len(monitored)
				resp.RadarrMonitored = &monitoredCount
				resp.RadarrMatched = &matchedCount
			}
		}
	}

	if cfg.Sonarr.URL == "" || cfg.Sonarr.APIKey == "" {
		resp.SonarrError = "sonarr_not_configured"
	} else {
		sonarrClient := sonarr.New(sonarr.Config{
			BaseURL:     cfg.Sonarr.URL,
			APIKey:      cfg.Sonarr.APIKey,
			Timeout:     existenceCheckTimeout,
			RetryConfig: retry.Config{MaxAttempts: 1},
		})

		if allSeries, err := sonarrClient.GetAllMonitoredSeries(ctx); err != nil {
			resp.SonarrError = "sonarr_unreachable"
		} else {
			monitoredCount := len(allSeries)
			resp.SonarrMonitored = &monitoredCount
		}
	}

	c.JSON(http.StatusOK, resp)
}

// getRadarrMovieMatches handles GET /api/v1/radarr/movies/:id/matches, returning
// the matched local Movie metadata (if any) and its full playlist occurrence list,
// regardless of pipeline state.
func (s *Server) getRadarrMovieMatches(c *gin.Context) {
	cfg := config.Get()
	if cfg.Radarr.URL == "" || cfg.Radarr.APIKey == "" {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "radarr_not_configured",
			Message: "Radarr is not configured",
		})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "invalid radarr movie id",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), existenceCheckTimeout)
	defer cancel()

	client := radarr.New(radarr.Config{
		BaseURL:     cfg.Radarr.URL,
		APIKey:      cfg.Radarr.APIKey,
		Timeout:     existenceCheckTimeout,
		RetryConfig: retry.Config{MaxAttempts: 1},
	})

	movie, err := client.GetMovieDetails(ctx, id)
	if err != nil {
		c.JSON(http.StatusBadGateway, ErrorResponse{
			Error:   "radarr_unreachable",
			Message: "failed to reach Radarr",
		})
		return
	}

	db := database.Get()
	matches, err := matcher.MatchMoviesBatch(db, []radarr.Movie{*movie})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "failed to compute playlist match status",
		})
		return
	}

	result := matches[movie.ID]
	response := RadarrMovieMatchesResponse{Matched: result.Matched, Occurrences: []OccurrenceResponse{}}
	if result.Matched && result.Movie != nil {
		movieResp := toMovieResponse(*result.Movie)
		response.Movie = &movieResp

		occurrences, err := matcher.FindAllMovieOccurrences(db, result.Movie.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "database_error",
				Message: "failed to fetch playlist occurrences",
			})
			return
		}
		response.Occurrences = toOccurrenceResponses(occurrences)
	}

	c.JSON(http.StatusOK, response)
}

// getSonarrSeriesEpisodes handles GET /api/v1/sonarr/series/:id/episodes,
// returning per-monitored-episode match detail underlying a series' aggregate
// ratio in the list view.
func (s *Server) getSonarrSeriesEpisodes(c *gin.Context) {
	cfg := config.Get()
	if cfg.Sonarr.URL == "" || cfg.Sonarr.APIKey == "" {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "sonarr_not_configured",
			Message: "Sonarr is not configured",
		})
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "invalid sonarr series id",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), existenceCheckTimeout)
	defer cancel()

	client := sonarr.New(sonarr.Config{
		BaseURL:     cfg.Sonarr.URL,
		APIKey:      cfg.Sonarr.APIKey,
		Timeout:     existenceCheckTimeout,
		RetryConfig: retry.Config{MaxAttempts: 1},
	})

	series, err := client.GetSeriesDetails(ctx, id)
	if err != nil {
		c.JSON(http.StatusBadGateway, ErrorResponse{
			Error:   "sonarr_unreachable",
			Message: "failed to reach Sonarr",
		})
		return
	}

	episodes, err := client.GetEpisodesBySeriesID(ctx, series.ID)
	if err != nil {
		c.JSON(http.StatusBadGateway, ErrorResponse{
			Error:   "sonarr_unreachable",
			Message: "failed to fetch episodes from Sonarr",
		})
		return
	}

	var monitored []matcher.SeasonEpisode
	for _, ep := range episodes {
		if ep.Monitored {
			monitored = append(monitored, matcher.SeasonEpisode{Season: ep.SeasonNumber, Episode: ep.EpisodeNumber})
		}
	}

	db := database.Get()
	details, err := matcher.MatchSeriesEpisodesDetail(db, series.TvdbID, monitored)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "failed to compute episode match status",
		})
		return
	}

	items := make([]SonarrSeriesEpisodeItem, len(details))
	for i, d := range details {
		item := SonarrSeriesEpisodeItem{
			Season:      d.Season,
			Episode:     d.Episode,
			Matched:     d.Matched,
			Occurrences: []OccurrenceResponse{},
		}
		if d.Matched {
			item.Occurrences = toOccurrenceResponses(d.Occurrences)
		}
		items[i] = item
	}

	c.JSON(http.StatusOK, SonarrSeriesEpisodesResponse{Episodes: items})
}

// parsePaginationBounded is parsePagination with caller-supplied default/max
// bounds, for endpoints (like the Radarr/Sonarr monitoring lists) whose page cost
// is far higher per-item than a plain DB list.
func parsePaginationBounded(c *gin.Context, defaultLimit, maxLimit int) (limit, offset int) {
	limit = defaultLimit
	offset = 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			if limit > maxLimit {
				limit = maxLimit
			}
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	return limit, offset
}

func totalPagesFor(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(limit)))
}

func sliceRadarrMoviesPage(movies []radarr.Movie, offset, limit int) []radarr.Movie {
	if offset >= len(movies) {
		return []radarr.Movie{}
	}
	end := offset + limit
	if end > len(movies) {
		end = len(movies)
	}
	return movies[offset:end]
}

func sliceSonarrSeriesPage(series []sonarr.Series, offset, limit int) []sonarr.Series {
	if offset >= len(series) {
		return []sonarr.Series{}
	}
	end := offset + limit
	if end > len(series) {
		end = len(series)
	}
	return series[offset:end]
}

func toOccurrenceResponses(lines []models.ProcessedLine) []OccurrenceResponse {
	responses := make([]OccurrenceResponse, len(lines))
	for i, line := range lines {
		responses[i] = OccurrenceResponse{
			ID:         line.ID,
			Resolution: line.Resolution,
			State:      string(line.State),
		}
	}
	return responses
}
