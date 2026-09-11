// Package httpclient provides a small shared HTTP client used by peer
// external-service integrations (radarr, sonarr) to build requests, check
// response status codes, and decode JSON responses without each duplicating
// the same plumbing.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/retry"
)

// Client holds the shared configuration needed to call a JSON HTTP API that
// authenticates via an API key header.
type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
	Retry   retry.Config
	Logger  *logger.Logger
}

// New creates a new Client.
func New(baseURL, apiKey string, timeout time.Duration, retryCfg retry.Config, log *logger.Logger) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTP: &http.Client{
			Timeout: timeout,
		},
		Retry:  retryCfg,
		Logger: log,
	}
}

// NewRequest builds an HTTP request against the client's BaseURL, marshaling
// body as JSON when non-nil, and setting the X-Api-Key/Content-Type/Accept
// headers common to these APIs.
func (c *Client) NewRequest(ctx context.Context, method, endpoint string, body any) (*http.Request, error) {
	url := c.BaseURL + endpoint

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

	req.Header.Set("X-Api-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// CheckStatus returns an error describing the response body when resp's
// status code is not one of want.
func CheckStatus(resp *http.Response, want ...int) error {
	for _, w := range want {
		if resp.StatusCode == w {
			return nil
		}
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
}

// Get performs a GET request against endpoint and decodes a 200 JSON
// response body into T.
func Get[T any](ctx context.Context, c *Client, endpoint string) (T, error) {
	var zero T

	req, err := c.NewRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return zero, err
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()

	if err := CheckStatus(resp, http.StatusOK); err != nil {
		return zero, err
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return zero, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// GetPage performs a GET request against endpoint and decodes a 200 JSON
// response body shaped as {"totalRecords": int, "records": [T]}.
func GetPage[T any](ctx context.Context, c *Client, endpoint string) ([]T, int, error) {
	req, err := c.NewRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if err := CheckStatus(resp, http.StatusOK); err != nil {
		return nil, 0, err
	}

	var page struct {
		TotalRecords int `json:"totalRecords"`
		Records      []T `json:"records"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return page.Records, page.TotalRecords, nil
}

// Put performs a PUT request against endpoint with body marshaled as JSON,
// accepting a 200 or 202 response and discarding its body.
func Put(ctx context.Context, c *Client, endpoint string, body any) error {
	req, err := c.NewRequest(ctx, "PUT", endpoint, body)
	if err != nil {
		return err
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return CheckStatus(resp, http.StatusOK, http.StatusAccepted)
}
