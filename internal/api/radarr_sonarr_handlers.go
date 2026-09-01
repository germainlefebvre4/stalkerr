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
	RadarrID int    `json:"radarr_id"`
	Title    string `json:"title"`
	Year     int    `json:"year"`
	HasFile  bool   `json:"has_file"`
	Matched  bool   `json:"matched"`
	MovieID  *uint  `json:"movie_id,omitempty"`
}

// SonarrSeriesListItem is one row of the Sonarr monitoring list, aggregated per series.
type SonarrSeriesListItem struct {
	SonarrID       int    `json:"sonarr_id"`
	Title          string `json:"title"`
	Year           int    `json:"year"`
	MatchedCount   int    `json:"matched_count"`
	MonitoredCount int    `json:"monitored_count"`
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

// listRadarrMonitoredMovies handles GET /api/v1/radarr/movies?limit&offset.
// Pagination happens before matching: the full Radarr movie list is fetched once,
// filtered to monitored movies and sorted, then only the requested page's movies
// are matched against the local playlist database. See
// openspec/changes/radarr-sonarr-monitoring-view.
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

	monitored := make([]radarr.Movie, 0, len(allMovies))
	for _, m := range allMovies {
		if m.Monitored {
			monitored = append(monitored, m)
		}
	}
	sort.Slice(monitored, func(i, j int) bool {
		return strings.ToLower(monitored[i].Title) < strings.ToLower(monitored[j].Title)
	})

	total := len(monitored)
	pageMovies := sliceRadarrMoviesPage(monitored, offset, limit)

	db := database.Get()
	matches, err := matcher.MatchMoviesBatch(db, pageMovies)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "failed to compute playlist match status",
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

// listSonarrMonitoredSeries handles GET /api/v1/sonarr/series?limit&offset.
// Pagination happens before the per-series episode fetch/matching: the full
// monitored series list is fetched once, sorted, then only the requested page's
// series get their monitored episodes fetched from Sonarr (concurrently, bounded
// by the page size) and matched against the local playlist database.
func (s *Server) listSonarrMonitoredSeries(c *gin.Context) {
	cfg := config.Get()
	if cfg.Sonarr.URL == "" || cfg.Sonarr.APIKey == "" {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   "sonarr_not_configured",
			Message: "Sonarr is not configured",
		})
		return
	}

	limit, offset := parsePaginationBounded(c, radarrSonarrDefaultPageSize, radarrSonarrMaxPageSize)

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

	sort.Slice(allSeries, func(i, j int) bool {
		return strings.ToLower(allSeries[i].Title) < strings.ToLower(allSeries[j].Title)
	})

	total := len(allSeries)
	pageSeries := sliceSonarrSeriesPage(allSeries, offset, limit)

	type aggregateResult struct {
		matched, monitored int
		err                error
	}
	aggregates := make([]aggregateResult, len(pageSeries))

	db := database.Get()
	var wg sync.WaitGroup
	for i, series := range pageSeries {
		wg.Add(1)
		go func(i int, series sonarr.Series) {
			defer wg.Done()

			episodes, err := client.GetEpisodesBySeriesID(ctx, series.ID)
			if err != nil {
				aggregates[i].err = err
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

			matched, monitoredTotal, err := matcher.MatchSeriesEpisodesAggregate(db, series.TvdbID, monitoredEpisodes)
			if err != nil {
				aggregates[i].err = err
				return
			}
			aggregates[i].matched = matched
			aggregates[i].monitored = monitoredTotal
		}(i, series)
	}
	wg.Wait()

	items := make([]SonarrSeriesListItem, len(pageSeries))
	for i, series := range pageSeries {
		if aggregates[i].err != nil {
			c.JSON(http.StatusBadGateway, ErrorResponse{
				Error:   "sonarr_unreachable",
				Message: "failed to fetch episodes from Sonarr for one or more series",
			})
			return
		}
		items[i] = SonarrSeriesListItem{
			SonarrID:       series.ID,
			Title:          series.Title,
			Year:           series.Year,
			MatchedCount:   aggregates[i].matched,
			MonitoredCount: aggregates[i].monitored,
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
