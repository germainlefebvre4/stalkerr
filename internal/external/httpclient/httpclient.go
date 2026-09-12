// Package httpclient provides a small shared HTTP client used by peer
// external-service integrations (radarr, sonarr) to build requests, check
// response status codes, and decode JSON responses without each duplicating
// the same plumbing.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	apperrors "github.com/glefebvre/stalkeer/internal/apperrors"
	"github.com/glefebvre/stalkeer/internal/circuitbreaker"
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

	// Breaker protects every call this client makes when non-nil. Get,
	// GetPage, and Put call c.HTTP.Do directly when it's nil, so a client
	// without a breaker behaves exactly as before.
	Breaker *circuitbreaker.CircuitBreaker
}

// New creates a new Client.
func New(baseURL, apiKey string, timeout time.Duration, retryCfg retry.Config, log *logger.Logger, breaker *circuitbreaker.CircuitBreaker) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTP: &http.Client{
			Timeout: timeout,
		},
		Retry:   retryCfg,
		Logger:  log,
		Breaker: breaker,
	}
}

// Execute runs fn through the client's circuit breaker when one is
// configured, or calls it directly when Breaker is nil. Shared by
// Get/GetPage/Put (wrapping only the HTTP round trip) and by callers like
// radarr/sonarr's SystemStatus (wrapping a whole single, unretried check).
func (c *Client) Execute(fn func() error) error {
	if c.Breaker == nil {
		return fn()
	}
	return c.Breaker.Execute(fn)
}

// doRequest executes req through Execute, classifying a transport-level
// failure into a retryable apperrors.AppError so retry.Do's classifier and
// the circuit breaker's IsSuccessful predicate can recognize it as the
// service's fault without inspecting error strings.
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	err := c.Execute(func() error {
		var doErr error
		resp, doErr = c.HTTP.Do(req)
		if doErr != nil {
			return classifyDoError(doErr)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// classifyDoError wraps a transport-level failure from c.HTTP.Do into a
// retryable apperrors.AppError: a timeout (including a ctx deadline, which
// surfaces as *url.Error wrapping context.DeadlineExceeded) becomes
// CodeServiceTimeout, any other network/connection failure becomes
// CodeServiceUnavailable.
func classifyDoError(err error) error {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return apperrors.Wrap(err, apperrors.CodeServiceTimeout, "request timed out")
	}
	return apperrors.Wrap(err, apperrors.CodeServiceUnavailable, "service unavailable")
}

// statusCoder is implemented by an error exposing the HTTP status code that
// caused it (e.g. radarr/sonarr's StatusError, used by SystemStatus), so
// IsSuccessful can single out a server-side failure worth counting against
// the breaker from a client error that isn't.
type statusCoder interface {
	StatusCode() int
}

// IsSuccessful is the circuit breaker success predicate shared by every
// Radarr/Sonarr breaker (see circuitbreaker.Config.IsSuccessful): an error
// only counts as a breaker failure when it's genuinely the service's
// fault - a retryable apperrors.AppError (network/timeout/429/5xx, from
// Get/GetPage/Put) or a StatusError-shaped error with a 429/5xx code (from
// SystemStatus). Anything else - nil, or a non-retryable client error like
// a bad request or bad credentials (400/401/404) - counts as success, so it
// never trips the circuit and its real reason keeps reaching the caller.
func IsSuccessful(err error) bool {
	if err == nil {
		return true
	}
	if apperrors.IsRetryable(err) {
		return false
	}
	var sc statusCoder
	if errors.As(err, &sc) {
		code := sc.StatusCode()
		if code == http.StatusTooManyRequests || code >= http.StatusInternalServerError {
			return false
		}
	}
	return true
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
// status code is not one of want. A 429 or 5xx response is classified into a
// retryable apperrors.AppError (CodeRateLimited / CodeServiceUnavailable);
// any other unwanted status is left unclassified (not retryable), matching
// apperrors.IsRetryable's default of false.
func CheckStatus(resp *http.Response, want ...int) error {
	for _, w := range want {
		if resp.StatusCode == w {
			return nil
		}
	}

	body, _ := io.ReadAll(resp.Body)
	err := fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))

	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		return apperrors.Wrap(err, apperrors.CodeRateLimited, "rate limited")
	case resp.StatusCode >= http.StatusInternalServerError:
		return apperrors.Wrap(err, apperrors.CodeServiceUnavailable, "service unavailable")
	default:
		return err
	}
}

// Get performs a GET request against endpoint and decodes a 200 JSON
// response body into T.
func Get[T any](ctx context.Context, c *Client, endpoint string) (T, error) {
	var zero T

	req, err := c.NewRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return zero, err
	}

	resp, err := c.doRequest(req)
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

	resp, err := c.doRequest(req)
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

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return CheckStatus(resp, http.StatusOK, http.StatusAccepted)
}
