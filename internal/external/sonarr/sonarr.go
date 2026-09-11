package sonarr

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	apperrors "github.com/glefebvre/stalkeer/internal/apperrors"
	"github.com/glefebvre/stalkeer/internal/external/httpclient"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/retry"
)

// Client represents a Sonarr API client
type Client struct {
	http *httpclient.Client
}

// Config holds Sonarr client configuration
type Config struct {
	BaseURL     string
	APIKey      string
	Timeout     time.Duration
	RetryConfig retry.Config
	Logger      *logger.Logger
}

// Series represents a Sonarr series
type Series struct {
	ID                int       `json:"id"`
	Title             string    `json:"title"`
	Year              int       `json:"year"`
	TvdbID            int       `json:"tvdbId"`
	Path              string    `json:"path"`
	Monitored         bool      `json:"monitored"`
	SeasonCount       int       `json:"seasonCount"`
	EpisodeFileCount  int       `json:"episodeFileCount"`
	TotalEpisodeCount int       `json:"totalEpisodeCount"`
	Added             time.Time `json:"added"`
	QualityProfileID  int       `json:"qualityProfileId"`
}

// Episode represents a Sonarr episode
type Episode struct {
	ID            int       `json:"id"`
	SeriesID      int       `json:"seriesId"`
	Title         string    `json:"title"`
	SeasonNumber  int       `json:"seasonNumber"`
	EpisodeNumber int       `json:"episodeNumber"`
	HasFile       bool      `json:"hasFile"`
	Monitored     bool      `json:"monitored"`
	AirDate       string    `json:"airDate"`
	AirDateUtc    time.Time `json:"airDateUtc"`
}

// FetchOptions controls how many records are fetched. Limit 0 means unlimited.
type FetchOptions struct {
	Limit int
}

// New creates a new Sonarr client
func New(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	if cfg.RetryConfig.MaxAttempts == 0 {
		cfg.RetryConfig = retry.DefaultConfig()
	}

	return &Client{
		http: httpclient.New(cfg.BaseURL, cfg.APIKey, cfg.Timeout, cfg.RetryConfig, cfg.Logger),
	}
}

// GetMissingSeries retrieves all monitored series with missing episodes
func (c *Client) GetMissingSeries(ctx context.Context) ([]Series, error) {
	endpoint := "/api/v3/series"

	var allSeries []Series
	err := retry.Do(ctx, c.http.Retry, func() error {
		series, err := c.getSeries(ctx, endpoint)
		if err != nil {
			return err
		}
		allSeries = series
		return nil
	}, apperrors.IsRetryable)

	if err != nil {
		return nil, apperrors.ExternalServiceError("sonarr", "failed to get series", err)
	}

	// Filter for monitored series with missing episodes
	var missing []Series
	for _, s := range allSeries {
		if s.Monitored && s.EpisodeFileCount < s.TotalEpisodeCount {
			missing = append(missing, s)
		}
	}

	return missing, nil
}

// GetAllSeries retrieves every series known to Sonarr in a single call,
// regardless of monitored/episode-file status. Unlike GetMissingSeries/
// GetAllMonitoredSeries (which only ever return monitored series), this
// returns the full catalog so callers can distinguish a series that is
// genuinely absent from Sonarr from one that is present but unmonitored -
// GetAllMonitoredSeries's own client-side filter makes that distinction
// impossible from its result alone.
func (c *Client) GetAllSeries(ctx context.Context) ([]Series, error) {
	endpoint := "/api/v3/series"

	var allSeries []Series
	err := retry.Do(ctx, c.http.Retry, func() error {
		series, err := c.getSeries(ctx, endpoint)
		if err != nil {
			return err
		}
		allSeries = series
		return nil
	}, apperrors.IsRetryable)

	if err != nil {
		return nil, apperrors.ExternalServiceError("sonarr", "failed to get all series", err)
	}

	return allSeries, nil
}

// GetAllMonitoredSeries retrieves every monitored series in Sonarr, regardless of
// missing-episode status. Unlike GetMissingSeries (which also requires
// EpisodeFileCount < TotalEpisodeCount), a fully-downloaded monitored series is
// still included here.
func (c *Client) GetAllMonitoredSeries(ctx context.Context) ([]Series, error) {
	allSeries, err := c.GetAllSeries(ctx)
	if err != nil {
		return nil, err
	}

	var monitored []Series
	for _, s := range allSeries {
		if s.Monitored {
			monitored = append(monitored, s)
		}
	}

	return monitored, nil
}

// GetSeriesDetails retrieves detailed information for a specific series
func (c *Client) GetSeriesDetails(ctx context.Context, id int) (*Series, error) {
	endpoint := fmt.Sprintf("/api/v3/series/%d", id)

	var series Series
	err := retry.Do(ctx, c.http.Retry, func() error {
		s, err := c.getSingleSeries(ctx, endpoint)
		if err != nil {
			return err
		}
		series = *s
		return nil
	}, apperrors.IsRetryable)

	if err != nil {
		return nil, apperrors.ExternalServiceError("sonarr", "failed to get series details", err)
	}

	return &series, nil
}

// GetMissingEpisodes retrieves missing episodes by paginating the wanted/missing
// endpoint. Pagination stops when all records are fetched or opts.Limit is reached
// (0 = unlimited). Episodes are sorted by series title, season, and episode number.
func (c *Client) GetMissingEpisodes(ctx context.Context, opts FetchOptions) ([]Episode, error) {
	const ps = 1000
	var all []Episode
	for page := 1; ; page++ {
		endpoint := fmt.Sprintf("/api/v3/wanted/missing?page=%d&pageSize=%d&sortKey=series.sortTitle&sortDirection=ascending", page, ps)

		var records []Episode
		var total int
		err := retry.Do(ctx, c.http.Retry, func() error {
			r, t, err := c.getEpisodes(ctx, endpoint)
			if err != nil {
				return err
			}
			records = r
			total = t
			return nil
		}, apperrors.IsRetryable)

		if err != nil {
			return nil, apperrors.ExternalServiceError("sonarr", "failed to get missing episodes", err)
		}

		all = append(all, records...)

		if c.http.Logger != nil {
			c.http.Logger.Info(fmt.Sprintf("sonarr: fetched page %d (%d/%d episodes)", page, len(all), total))
		}

		if opts.Limit > 0 && len(all) >= opts.Limit {
			all = all[:opts.Limit]
			break
		}
		if len(all) >= total || len(records) == 0 {
			break
		}
	}
	return all, nil
}

// GetEpisodeDetails retrieves detailed information for a specific episode
func (c *Client) GetEpisodeDetails(ctx context.Context, id int) (*Episode, error) {
	endpoint := fmt.Sprintf("/api/v3/episode/%d", id)

	var episode Episode
	err := retry.Do(ctx, c.http.Retry, func() error {
		ep, err := c.getEpisode(ctx, endpoint)
		if err != nil {
			return err
		}
		episode = *ep
		return nil
	}, apperrors.IsRetryable)

	if err != nil {
		return nil, apperrors.ExternalServiceError("sonarr", "failed to get episode details", err)
	}

	return &episode, nil
}

// GetSeriesByTVDBID looks up a series directly by TVDB ID, independent of the
// missing/wanted list, so it also finds series that already have files.
// Returns (nil, nil) when Sonarr has no matching series (a distinguishable
// "not found" outcome), and (nil, err) when the lookup itself fails.
func (c *Client) GetSeriesByTVDBID(ctx context.Context, tvdbID int) (*Series, error) {
	endpoint := fmt.Sprintf("/api/v3/series?tvdbId=%d", tvdbID)

	var series []Series
	err := retry.Do(ctx, c.http.Retry, func() error {
		s, err := c.getSeries(ctx, endpoint)
		if err != nil {
			return err
		}
		series = s
		return nil
	}, apperrors.IsRetryable)

	if err != nil {
		return nil, apperrors.ExternalServiceError("sonarr", "failed to get series by tvdb id", err)
	}

	if len(series) == 0 {
		return nil, nil
	}

	return &series[0], nil
}

// GetEpisodesBySeriesID retrieves every episode belonging to a Sonarr series,
// independent of "missing" status.
func (c *Client) GetEpisodesBySeriesID(ctx context.Context, seriesID int) ([]Episode, error) {
	endpoint := fmt.Sprintf("/api/v3/episode?seriesId=%d", seriesID)

	var episodes []Episode
	err := retry.Do(ctx, c.http.Retry, func() error {
		eps, err := c.getEpisodeList(ctx, endpoint)
		if err != nil {
			return err
		}
		episodes = eps
		return nil
	}, apperrors.IsRetryable)

	if err != nil {
		return nil, apperrors.ExternalServiceError("sonarr", "failed to get episodes by series id", err)
	}

	return episodes, nil
}

// FindEpisodeByTVDBID looks up a series by TVDB ID and, among its episodes,
// the one matching the given season/episode number, filtering client-side
// since Sonarr has no direct season+episode-number query. Returns
// (nil, nil, nil) when the series or the specific episode isn't found, and
// (nil, nil, err) when a live call fails.
func (c *Client) FindEpisodeByTVDBID(ctx context.Context, tvdbID, season, episode int) (*Series, *Episode, error) {
	series, err := c.GetSeriesByTVDBID(ctx, tvdbID)
	if err != nil {
		return nil, nil, err
	}
	if series == nil {
		return nil, nil, nil
	}

	episodes, err := c.GetEpisodesBySeriesID(ctx, series.ID)
	if err != nil {
		return nil, nil, err
	}

	for i := range episodes {
		if episodes[i].SeasonNumber == season && episodes[i].EpisodeNumber == episode {
			return series, &episodes[i], nil
		}
	}

	return nil, nil, nil
}

// UpdateEpisode updates an episode in Sonarr
func (c *Client) UpdateEpisode(ctx context.Context, episode *Episode) error {
	endpoint := fmt.Sprintf("/api/v3/episode/%d", episode.ID)

	err := retry.Do(ctx, c.http.Retry, func() error {
		return c.putEpisode(ctx, endpoint, episode)
	}, apperrors.IsRetryable)

	if err != nil {
		return apperrors.ExternalServiceError("sonarr", "failed to update episode", err)
	}

	return nil
}

// StatusError wraps a non-2xx HTTP response so callers can distinguish an
// authorization failure (401) from other reachability failures without
// parsing message strings. StatusCode() satisfies the structural
// `interface{ StatusCode() int }` the aggregation endpoint's classifier
// checks for, without that package needing to import sonarr.
type StatusError struct {
	Code int
	Body string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("unexpected status code %d: %s", e.Code, e.Body)
}

// StatusCode returns the HTTP status code that produced this error.
func (e *StatusError) StatusCode() int {
	return e.Code
}

// SystemStatus performs a lightweight, unretried reachability check against
// Sonarr's own system/status endpoint - the same call Sonarr's UI uses to
// confirm connectivity. Unlike the other client methods, it does not go
// through retry.Do: a diagnostic check must fail fast under the caller's ctx
// deadline rather than retry like a real data fetch.
func (c *Client) SystemStatus(ctx context.Context) error {
	req, err := c.newRequest(ctx, "GET", "/api/v3/system/status", nil)
	if err != nil {
		return err
	}

	resp, err := c.http.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &StatusError{Code: resp.StatusCode, Body: string(body)}
	}

	return nil
}

func (c *Client) getSeries(ctx context.Context, endpoint string) ([]Series, error) {
	return httpclient.Get[[]Series](ctx, c.http, endpoint)
}

func (c *Client) getSingleSeries(ctx context.Context, endpoint string) (*Series, error) {
	series, err := httpclient.Get[Series](ctx, c.http, endpoint)
	if err != nil {
		return nil, err
	}
	return &series, nil
}

func (c *Client) getEpisodes(ctx context.Context, endpoint string) ([]Episode, int, error) {
	return httpclient.GetPage[Episode](ctx, c.http, endpoint)
}

func (c *Client) getEpisodeList(ctx context.Context, endpoint string) ([]Episode, error) {
	return httpclient.Get[[]Episode](ctx, c.http, endpoint)
}

func (c *Client) getEpisode(ctx context.Context, endpoint string) (*Episode, error) {
	episode, err := httpclient.Get[Episode](ctx, c.http, endpoint)
	if err != nil {
		return nil, err
	}
	return &episode, nil
}

func (c *Client) putEpisode(ctx context.Context, endpoint string, episode *Episode) error {
	return httpclient.Put(ctx, c.http, endpoint, episode)
}

func (c *Client) newRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Request, error) {
	return c.http.NewRequest(ctx, method, endpoint, body)
}
