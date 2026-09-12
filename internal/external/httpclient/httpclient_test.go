package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	apperrors "github.com/glefebvre/stalkeer/internal/apperrors"
	"github.com/glefebvre/stalkeer/internal/circuitbreaker"
	"github.com/glefebvre/stalkeer/internal/retry"
)

type testItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func testClient(baseURL string) *Client {
	return New(baseURL, "test-key", 5*time.Second, retry.Config{MaxAttempts: 1}, nil, nil)
}

func testClientWithBreaker(baseURL string, breaker *circuitbreaker.CircuitBreaker) *Client {
	return New(baseURL, "test-key", 5*time.Second, retry.Config{MaxAttempts: 1}, nil, breaker)
}

// fakeStatusError mimics the shape of radarr.StatusError/sonarr.StatusError
// (an error exposing StatusCode() int) without importing either package.
type fakeStatusError struct{ code int }

func (e *fakeStatusError) Error() string   { return fmt.Sprintf("status %d", e.code) }
func (e *fakeStatusError) StatusCode() int { return e.code }

func TestGetDecodesSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "test-key" {
			t.Errorf("expected X-Api-Key header to be set")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testItem{ID: 1, Name: "widget"})
	}))
	defer server.Close()

	item, err := Get[testItem](context.Background(), testClient(server.URL), "/items/1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.ID != 1 || item.Name != "widget" {
		t.Errorf("unexpected item: %+v", item)
	}
}

func TestGetPageDecodesSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			TotalRecords int        `json:"totalRecords"`
			Records      []testItem `json:"records"`
		}{
			TotalRecords: 2,
			Records:      []testItem{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}},
		})
	}))
	defer server.Close()

	records, total, err := GetPage[testItem](context.Background(), testClient(server.URL), "/items")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected total 2, got %d", total)
	}
	if len(records) != 2 || records[0].ID != 1 || records[1].ID != 2 {
		t.Errorf("unexpected records: %+v", records)
	}
}

func TestGetNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	_, err := Get[testItem](context.Background(), testClient(server.URL), "/items/1")
	if err == nil {
		t.Fatal("expected an error")
	}
	wantMsg := "unexpected status code 500: boom"
	if !strings.Contains(err.Error(), wantMsg) {
		t.Errorf("expected error to contain %q, got %q", wantMsg, err.Error())
	}
	if !apperrors.IsRetryable(err) {
		t.Errorf("expected a 500 response to be classified as retryable, got %v", err)
	}
}

func TestPutNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("nope"))
	}))
	defer server.Close()

	err := Put(context.Background(), testClient(server.URL), "/items/1", testItem{ID: 1})
	if err == nil {
		t.Fatal("expected an error")
	}
	want := "unexpected status code 400: nope"
	if err.Error() != want {
		t.Errorf("expected error %q, got %q", want, err.Error())
	}
}

func TestPutSuccessAcceptsOKAndAccepted(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusAccepted} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
		}))

		err := Put(context.Background(), testClient(server.URL), "/items/1", testItem{ID: 1})
		server.Close()

		if err != nil {
			t.Errorf("status %d: unexpected error: %v", status, err)
		}
	}
}

func TestGetDecodeFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	_, err := Get[testItem](context.Background(), testClient(server.URL), "/items/1")
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.HasPrefix(err.Error(), "failed to decode response: ") {
		t.Errorf("expected decode error prefix, got %q", err.Error())
	}
}

func TestNewRequestMarshalFailure(t *testing.T) {
	c := testClient("http://example.invalid")

	_, err := c.NewRequest(context.Background(), "PUT", "/items/1", func() {})
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.HasPrefix(err.Error(), "failed to marshal request body: ") {
		t.Errorf("expected marshal error prefix, got %q", err.Error())
	}
}

func TestBreakerOpenBlocksSubsequentCalls(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(testItem{ID: 1, Name: "widget"})
	}))
	defer server.Close()

	breaker := circuitbreaker.New(circuitbreaker.Config{
		MaxFailures:         2,
		Timeout:             time.Minute,
		MaxHalfOpenRequests: 1,
		IsSuccessful:        IsSuccessful,
	})

	// Drive the breaker open with genuine transport failures against an
	// address nothing listens on, sharing the same breaker instance the
	// client below uses - exactly how one breaker is shared across every
	// call to a service in production.
	deadClient := testClientWithBreaker("http://127.0.0.1:1", breaker)
	for i := 0; i < 2; i++ {
		if _, err := Get[testItem](context.Background(), deadClient, "/items/1"); err == nil {
			t.Fatalf("expected connection failure on attempt %d", i)
		}
	}
	if breaker.State() != circuitbreaker.StateOpen {
		t.Fatalf("expected breaker to be open after repeated failures, got %s", breaker.State())
	}

	client := testClientWithBreaker(server.URL, breaker)

	if _, err := Get[testItem](context.Background(), client, "/items/1"); !errors.Is(err, circuitbreaker.ErrOpenState) {
		t.Errorf("expected Get to return ErrOpenState, got %v", err)
	}
	if _, _, err := GetPage[testItem](context.Background(), client, "/items"); !errors.Is(err, circuitbreaker.ErrOpenState) {
		t.Errorf("expected GetPage to return ErrOpenState, got %v", err)
	}
	if err := Put(context.Background(), client, "/items/1", testItem{ID: 1}); !errors.Is(err, circuitbreaker.ErrOpenState) {
		t.Errorf("expected Put to return ErrOpenState, got %v", err)
	}
	if calls != 0 {
		t.Errorf("expected no request to reach the test server while the breaker is open, got %d calls", calls)
	}
}

func TestClassifyDoError(t *testing.T) {
	t.Run("timeout is retryable and CodeServiceTimeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(50 * time.Millisecond)
		}))
		defer server.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		_, err := Get[testItem](ctx, testClient(server.URL), "/items/1")
		if err == nil {
			t.Fatal("expected a timeout error")
		}
		if !apperrors.IsRetryable(err) {
			t.Errorf("expected timeout error to be retryable, got %v", err)
		}
		if apperrors.GetErrorCode(err) != apperrors.CodeServiceTimeout {
			t.Errorf("expected CodeServiceTimeout, got %s", apperrors.GetErrorCode(err))
		}
	})

	t.Run("connection failure is retryable and CodeServiceUnavailable", func(t *testing.T) {
		_, err := Get[testItem](context.Background(), testClient("http://127.0.0.1:1"), "/items/1")
		if err == nil {
			t.Fatal("expected a connection error")
		}
		if !apperrors.IsRetryable(err) {
			t.Errorf("expected connection error to be retryable, got %v", err)
		}
		if apperrors.GetErrorCode(err) != apperrors.CodeServiceUnavailable {
			t.Errorf("expected CodeServiceUnavailable, got %s", apperrors.GetErrorCode(err))
		}
	})

	t.Run("429 response is retryable and CodeRateLimited", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("slow down"))
		}))
		defer server.Close()

		_, err := Get[testItem](context.Background(), testClient(server.URL), "/items/1")
		if !apperrors.IsRetryable(err) {
			t.Errorf("expected 429 to be retryable, got %v", err)
		}
		if apperrors.GetErrorCode(err) != apperrors.CodeRateLimited {
			t.Errorf("expected CodeRateLimited, got %s", apperrors.GetErrorCode(err))
		}
	})

	t.Run("5xx response is retryable and CodeServiceUnavailable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("down"))
		}))
		defer server.Close()

		_, err := Get[testItem](context.Background(), testClient(server.URL), "/items/1")
		if !apperrors.IsRetryable(err) {
			t.Errorf("expected 5xx to be retryable, got %v", err)
		}
		if apperrors.GetErrorCode(err) != apperrors.CodeServiceUnavailable {
			t.Errorf("expected CodeServiceUnavailable, got %s", apperrors.GetErrorCode(err))
		}
	})

	t.Run("other 4xx response is not retryable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("nope"))
		}))
		defer server.Close()

		_, err := Get[testItem](context.Background(), testClient(server.URL), "/items/1")
		if err == nil {
			t.Fatal("expected an error")
		}
		if apperrors.IsRetryable(err) {
			t.Errorf("expected 404 to not be retryable, got %v", err)
		}
	})
}

func TestIsSuccessful(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, true},
		{"retryable service-unavailable error", apperrors.Wrap(fmt.Errorf("boom"), apperrors.CodeServiceUnavailable, "unavailable"), false},
		{"retryable timeout error", apperrors.Wrap(fmt.Errorf("boom"), apperrors.CodeServiceTimeout, "timeout"), false},
		{"retryable rate-limited error", apperrors.Wrap(fmt.Errorf("boom"), apperrors.CodeRateLimited, "rate limited"), false},
		{"non-retryable validation error", apperrors.New(apperrors.CodeValidation, "bad input"), true},
		{"status-shaped 401 counts as success", &fakeStatusError{code: http.StatusUnauthorized}, true},
		{"status-shaped 400 counts as success", &fakeStatusError{code: http.StatusBadRequest}, true},
		{"status-shaped 404 counts as success", &fakeStatusError{code: http.StatusNotFound}, true},
		{"status-shaped 429 counts as failure", &fakeStatusError{code: http.StatusTooManyRequests}, false},
		{"status-shaped 503 counts as failure", &fakeStatusError{code: http.StatusServiceUnavailable}, false},
		{"plain error with no status counts as success", fmt.Errorf("boom"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSuccessful(tt.err); got != tt.want {
				t.Errorf("IsSuccessful(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
