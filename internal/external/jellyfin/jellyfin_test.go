package jellyfin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/retry"
)

func TestNew(t *testing.T) {
	cfg := Config{
		BaseURL: "http://localhost:8096",
		APIKey:  "test-key",
	}

	client := New(cfg)

	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.baseURL != cfg.BaseURL {
		t.Errorf("expected baseURL %s, got %s", cfg.BaseURL, client.baseURL)
	}
	if client.apiKey != cfg.APIKey {
		t.Errorf("expected apiKey %s, got %s", cfg.APIKey, client.apiKey)
	}
}

func TestNotifyPathsUpdated(t *testing.T) {
	var gotMethod, gotPath, gotToken string
	var gotBody mediaUpdatedRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotToken = r.Header.Get("X-Emby-Token")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 5 * time.Second,
		RetryConfig: retry.Config{
			MaxAttempts: 1,
		},
	})

	err := client.NotifyPathsUpdated(context.Background(), []string{"/media/movies/Foo (2020)", "/media/tvshows/Bar (2021)/Season 01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected POST method, got %s", gotMethod)
	}
	if gotPath != "/Library/Media/Updated" {
		t.Errorf("expected path /Library/Media/Updated, got %s", gotPath)
	}
	if gotToken != "test-key" {
		t.Errorf("expected X-Emby-Token header 'test-key', got %q", gotToken)
	}
	if len(gotBody.Updates) != 2 {
		t.Fatalf("expected 2 updates, got %d", len(gotBody.Updates))
	}
	if gotBody.Updates[0].Path != "/media/movies/Foo (2020)" || gotBody.Updates[0].UpdateType != "Created" {
		t.Errorf("unexpected first update: %+v", gotBody.Updates[0])
	}
	if gotBody.Updates[1].Path != "/media/tvshows/Bar (2021)/Season 01" || gotBody.Updates[1].UpdateType != "Created" {
		t.Errorf("unexpected second update: %+v", gotBody.Updates[1])
	}
}

func TestNotifyPathsUpdatedEmptyPathsIsNoop(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 5 * time.Second,
		RetryConfig: retry.Config{
			MaxAttempts: 1,
		},
	})

	err := client.NotifyPathsUpdated(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected no request to be sent for an empty paths slice")
	}
}

func TestNotifyPathsUpdatedNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid api key"))
	}))
	defer server.Close()

	client := New(Config{
		BaseURL: server.URL,
		APIKey:  "bad-key",
		Timeout: 5 * time.Second,
		RetryConfig: retry.Config{
			MaxAttempts: 1,
		},
	})

	err := client.NotifyPathsUpdated(context.Background(), []string{"/media/movies/Foo (2020)"})
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestNotifyPathsUpdatedNetworkError(t *testing.T) {
	client := New(Config{
		BaseURL: "http://127.0.0.1:1",
		APIKey:  "test-key",
		Timeout: 500 * time.Millisecond,
		RetryConfig: retry.Config{
			MaxAttempts: 1,
		},
	})

	err := client.NotifyPathsUpdated(context.Background(), []string{"/media/movies/Foo (2020)"})
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

func TestSystemStatusOK(t *testing.T) {
	var gotPath, gotToken string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotToken = r.Header.Get("X-Emby-Token")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"Version":"10.8.0"}`))
	}))
	defer server.Close()

	client := New(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 5 * time.Second,
		RetryConfig: retry.Config{
			MaxAttempts: 1,
		},
	})

	if err := client.SystemStatus(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/System/Info" {
		t.Errorf("expected path /System/Info, got %s", gotPath)
	}
	if gotToken != "test-key" {
		t.Errorf("expected X-Emby-Token header 'test-key', got %q", gotToken)
	}
}

func TestSystemStatusUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid api key"))
	}))
	defer server.Close()

	client := New(Config{
		BaseURL: server.URL,
		APIKey:  "bad-key",
		Timeout: 5 * time.Second,
		RetryConfig: retry.Config{
			MaxAttempts: 1,
		},
	})

	err := client.SystemStatus(context.Background())
	if err == nil {
		t.Fatal("expected error for 401 response")
	}

	var statusErr *StatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("expected *StatusError, got %T: %v", err, err)
	}
	if statusErr.StatusCode() != http.StatusUnauthorized {
		t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, statusErr.StatusCode())
	}
}

func TestSystemStatusConnectionFailure(t *testing.T) {
	client := New(Config{
		BaseURL: "http://127.0.0.1:1",
		APIKey:  "test-key",
		Timeout: 500 * time.Millisecond,
		RetryConfig: retry.Config{
			MaxAttempts: 1,
		},
	})

	err := client.SystemStatus(context.Background())
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}

	var statusErr *StatusError
	if errors.As(err, &statusErr) {
		t.Fatalf("expected a connection-level error, not *StatusError: %v", err)
	}
}

func TestSessionsDecodesIdlePausedAndPlayingSessions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Sessions" {
			t.Errorf("expected path /Sessions, got %s", r.URL.Path)
		}
		if r.Header.Get("X-Emby-Token") != "test-key" {
			t.Errorf("expected X-Emby-Token header 'test-key', got %q", r.Header.Get("X-Emby-Token"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{"PlayState": {"IsPaused": false}},
			{"NowPlayingItem": {"Name": "Paused Movie"}, "PlayState": {"IsPaused": true}},
			{"NowPlayingItem": {"Name": "Playing Movie"}, "PlayState": {"IsPaused": false}}
		]`))
	}))
	defer server.Close()

	client := New(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 5 * time.Second,
	})

	sessions, err := client.Sessions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(sessions))
	}

	if sessions[0].IsActivelyPlaying() {
		t.Error("idle session (no NowPlayingItem) must not be actively playing")
	}
	if sessions[1].IsActivelyPlaying() {
		t.Error("paused session must not be actively playing")
	}
	if !sessions[2].IsActivelyPlaying() {
		t.Error("playing, non-paused session must be actively playing")
	}
}

func TestSessionsNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid api key"))
	}))
	defer server.Close()

	client := New(Config{BaseURL: server.URL, APIKey: "bad-key", Timeout: 5 * time.Second})

	_, err := client.Sessions(context.Background())
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}

func TestActivePlaybackTrueWhenSessionPlaying(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"NowPlayingItem": {"Name": "X"}, "PlayState": {"IsPaused": false}}]`))
	}))
	defer server.Close()

	client := New(Config{BaseURL: server.URL, APIKey: "test-key", Timeout: 5 * time.Second})

	if !client.ActivePlayback(context.Background()) {
		t.Error("expected ActivePlayback to report true")
	}
}

func TestActivePlaybackFalseWhenNoSessionsPlaying(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"PlayState": {"IsPaused": false}}]`))
	}))
	defer server.Close()

	client := New(Config{BaseURL: server.URL, APIKey: "test-key", Timeout: 5 * time.Second})

	if client.ActivePlayback(context.Background()) {
		t.Error("expected ActivePlayback to report false for an idle session")
	}
}

// TestActivePlaybackFailsOpenOnError covers the "Fail-Open on Jellyfin
// Unreachable" requirement: a connection error, timeout, or auth failure
// must never be reported as active playback.
func TestActivePlaybackFailsOpenOnError(t *testing.T) {
	client := New(Config{
		BaseURL: "http://127.0.0.1:1",
		APIKey:  "test-key",
		Timeout: 500 * time.Millisecond,
	})

	if client.ActivePlayback(context.Background()) {
		t.Error("expected fail-open (false) when Jellyfin is unreachable")
	}
}

func TestNotifyPathsUpdatedTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Timeout: 20 * time.Millisecond,
		RetryConfig: retry.Config{
			MaxAttempts: 1,
		},
	})

	err := client.NotifyPathsUpdated(context.Background(), []string{"/media/movies/Foo (2020)"})
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
