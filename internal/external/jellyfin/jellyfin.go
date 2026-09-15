package jellyfin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	apperrors "github.com/glefebvre/stalkeer/internal/apperrors"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/retry"
)

// Client represents a Jellyfin API client
type Client struct {
	baseURL     string
	apiKey      string
	httpClient  *http.Client
	retryConfig retry.Config
	logger      *logger.Logger
}

// Config holds Jellyfin client configuration
type Config struct {
	BaseURL     string
	APIKey      string
	Timeout     time.Duration
	RetryConfig retry.Config
	Logger      *logger.Logger
}

// mediaUpdate represents a single changed path in a Library/Media/Updated payload
type mediaUpdate struct {
	Path       string `json:"Path"`
	UpdateType string `json:"UpdateType"`
}

// mediaUpdatedRequest is the request body for Jellyfin's Library/Media/Updated endpoint
type mediaUpdatedRequest struct {
	Updates []mediaUpdate `json:"Updates"`
}

// New creates a new Jellyfin client. The default timeout is short (10s) since
// this call must never meaningfully delay a CLI run.
func New(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	if cfg.RetryConfig.MaxAttempts == 0 {
		cfg.RetryConfig = retry.DefaultConfig()
	}

	return &Client{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		retryConfig: cfg.RetryConfig,
		logger:      cfg.Logger,
	}
}

// NotifyPathsUpdated reports changed library paths to Jellyfin so it can
// rescan just those folders instead of waiting for its own periodic full
// library scan. An empty paths slice is a no-op: no request is sent.
func (c *Client) NotifyPathsUpdated(ctx context.Context, paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	updates := make([]mediaUpdate, 0, len(paths))
	for _, p := range paths {
		updates = append(updates, mediaUpdate{Path: p, UpdateType: "Created"})
	}
	body := mediaUpdatedRequest{Updates: updates}

	err := retry.Do(ctx, c.retryConfig, func() error {
		return c.postMediaUpdated(ctx, body)
	}, apperrors.IsRetryable)

	if err != nil {
		return apperrors.ExternalServiceError("jellyfin", "failed to notify library paths updated", err)
	}

	return nil
}

func (c *Client) postMediaUpdated(ctx context.Context, body mediaUpdatedRequest) error {
	req, err := c.newRequest(ctx, http.MethodPost, "/Library/Media/Updated", body)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// StatusError wraps a non-2xx HTTP response so callers can distinguish an
// authorization failure (401) from other reachability failures without
// parsing message strings. StatusCode() satisfies the structural
// `interface{ StatusCode() int }` the aggregation endpoint's classifier
// checks for, without that package needing to import jellyfin.
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

// SystemStatus performs a lightweight, unretried reachability-and-auth check
// against Jellyfin's /System/Info endpoint, which (unlike the unauthenticated
// /System/Info/Public) requires the X-Emby-Token header and returns 401 on an
// invalid key. Unlike NotifyPathsUpdated, it does not go through retry.Do: a
// diagnostic check must fail fast under the caller's ctx deadline rather than
// retry like a real data fetch.
func (c *Client) SystemStatus(ctx context.Context) error {
	req, err := c.newRequest(ctx, http.MethodGet, "/System/Info", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
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

func (c *Client) newRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Request, error) {
	url := c.baseURL + endpoint

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Emby-Token", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}
