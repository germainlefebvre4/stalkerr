package radarr

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

// Client represents a Radarr API client
type Client struct {
	http *httpclient.Client
}

// Config holds Radarr client configuration
type Config struct {
	BaseURL     string
	APIKey      string
	Timeout     time.Duration
	RetryConfig retry.Config
	Logger      *logger.Logger
}

// Movie represents a Radarr movie
type Movie struct {
	ID               int       `json:"id"`
	Title            string    `json:"title"`
	Year             int       `json:"year"`
	TvdbID           int       `json:"tvdbId"`
	TMDBID           int       `json:"tmdbId"`
	Path             string    `json:"path"`
	Monitored        bool      `json:"monitored"`
	HasFile          bool      `json:"hasFile"`
	SizeOnDisk       int64     `json:"sizeOnDisk"`
	Added            time.Time `json:"added"`
	QualityProfileID int       `json:"qualityProfileId"`
}

// FetchOptions controls how many records are fetched. Limit 0 means unlimited.
type FetchOptions struct {
	Limit int
}

// New creates a new Radarr client
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

// GetMissingMovies retrieves all monitored movies that are not downloaded by paginating
// the wanted/missing endpoint with server-side filtering.
// GetMissingMovies retrieves missing movies by paginating the wanted/missing
// endpoint. Pagination stops when all records are fetched or opts.Limit is reached
// (0 = unlimited).
func (c *Client) GetMissingMovies(ctx context.Context, opts FetchOptions) ([]Movie, error) {
	const ps = 1000
	var all []Movie
	for page := 1; ; page++ {
		endpoint := fmt.Sprintf("/api/v3/wanted/missing?page=%d&pageSize=%d&sortKey=title&sortDirection=ascending", page, ps)

		var records []Movie
		var total int
		err := retry.Do(ctx, c.http.Retry, func() error {
			r, t, err := c.getPagedMovies(ctx, endpoint)
			if err != nil {
				return err
			}
			records = r
			total = t
			return nil
		}, apperrors.IsRetryable)

		if err != nil {
			return nil, apperrors.ExternalServiceError("radarr", "failed to get missing movies", err)
		}

		all = append(all, records...)

		if c.http.Logger != nil {
			c.http.Logger.Info(fmt.Sprintf("radarr: fetched page %d (%d/%d movies)", page, len(all), total))
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

// GetAllMovies retrieves every movie known to Radarr in a single call, regardless
// of monitored/hasFile status. Unlike GetMissingMovies (which paginates the
// wanted/missing endpoint and only returns eligible-for-download movies), this
// returns the full lightweight catalog so callers can decide which subset
// (e.g. monitored-only) and which page to work with.
func (c *Client) GetAllMovies(ctx context.Context) ([]Movie, error) {
	endpoint := "/api/v3/movie"

	var movies []Movie
	err := retry.Do(ctx, c.http.Retry, func() error {
		m, err := c.getMovies(ctx, endpoint)
		if err != nil {
			return err
		}
		movies = m
		return nil
	}, apperrors.IsRetryable)

	if err != nil {
		return nil, apperrors.ExternalServiceError("radarr", "failed to get all movies", err)
	}

	return movies, nil
}

// GetMovieDetails retrieves detailed information for a specific movie
func (c *Client) GetMovieDetails(ctx context.Context, id int) (*Movie, error) {
	endpoint := fmt.Sprintf("/api/v3/movie/%d", id)

	var movie Movie
	err := retry.Do(ctx, c.http.Retry, func() error {
		m, err := c.getMovie(ctx, endpoint)
		if err != nil {
			return err
		}
		movie = *m
		return nil
	}, apperrors.IsRetryable)

	if err != nil {
		return nil, apperrors.ExternalServiceError("radarr", "failed to get movie details", err)
	}

	return &movie, nil
}

// GetMovieByTMDBID looks up a movie directly by TMDB ID, independent of the
// missing/wanted list, so it also finds movies that already have a file.
// Returns (nil, nil) when Radarr has no matching movie (a distinguishable
// "not found" outcome), and (nil, err) when the lookup itself fails.
func (c *Client) GetMovieByTMDBID(ctx context.Context, tmdbID int) (*Movie, error) {
	endpoint := fmt.Sprintf("/api/v3/movie?tmdbId=%d", tmdbID)

	var movies []Movie
	err := retry.Do(ctx, c.http.Retry, func() error {
		m, err := c.getMovies(ctx, endpoint)
		if err != nil {
			return err
		}
		movies = m
		return nil
	}, apperrors.IsRetryable)

	if err != nil {
		return nil, apperrors.ExternalServiceError("radarr", "failed to get movie by tmdb id", err)
	}

	if len(movies) == 0 {
		return nil, nil
	}

	return &movies[0], nil
}

// UpdateMovie updates a movie in Radarr
func (c *Client) UpdateMovie(ctx context.Context, movie *Movie) error {
	endpoint := fmt.Sprintf("/api/v3/movie/%d", movie.ID)

	err := retry.Do(ctx, c.http.Retry, func() error {
		return c.putMovie(ctx, endpoint, movie)
	}, apperrors.IsRetryable)

	if err != nil {
		return apperrors.ExternalServiceError("radarr", "failed to update movie", err)
	}

	return nil
}

// StatusError wraps a non-2xx HTTP response so callers can distinguish an
// authorization failure (401) from other reachability failures without
// parsing message strings. StatusCode() satisfies the structural
// `interface{ StatusCode() int }` the aggregation endpoint's classifier
// checks for, without that package needing to import radarr.
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
// Radarr's own system/status endpoint - the same call Radarr's UI uses to
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

func (c *Client) getPagedMovies(ctx context.Context, endpoint string) ([]Movie, int, error) {
	return httpclient.GetPage[Movie](ctx, c.http, endpoint)
}

func (c *Client) getMovies(ctx context.Context, endpoint string) ([]Movie, error) {
	return httpclient.Get[[]Movie](ctx, c.http, endpoint)
}

func (c *Client) getMovie(ctx context.Context, endpoint string) (*Movie, error) {
	movie, err := httpclient.Get[Movie](ctx, c.http, endpoint)
	if err != nil {
		return nil, err
	}
	return &movie, nil
}

func (c *Client) putMovie(ctx context.Context, endpoint string, movie *Movie) error {
	return httpclient.Put(ctx, c.http, endpoint, movie)
}

func (c *Client) newRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Request, error) {
	return c.http.NewRequest(ctx, method, endpoint, body)
}
