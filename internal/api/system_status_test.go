package api

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/external/radarr"
	"github.com/glefebvre/stalkeer/internal/external/sonarr"
	"github.com/glefebvre/stalkeer/internal/external/tmdb"
)

func TestClassifyReachabilityError(t *testing.T) {
	t.Run("ctx deadline exceeded maps to timeout", func(t *testing.T) {
		err := &url.Error{Op: "Get", URL: "http://example.com", Err: context.DeadlineExceeded}
		if got := classifyReachabilityError(err); got != reasonTimeout {
			t.Errorf("expected %s, got %s", reasonTimeout, got)
		}
	})

	t.Run("radarr 401 maps to unauthorized", func(t *testing.T) {
		err := &radarr.StatusError{Code: http.StatusUnauthorized, Body: "unauthorized"}
		if got := classifyReachabilityError(err); got != reasonUnauthorized {
			t.Errorf("expected %s, got %s", reasonUnauthorized, got)
		}
	})

	t.Run("sonarr 401 maps to unauthorized", func(t *testing.T) {
		err := &sonarr.StatusError{Code: http.StatusUnauthorized}
		if got := classifyReachabilityError(err); got != reasonUnauthorized {
			t.Errorf("expected %s, got %s", reasonUnauthorized, got)
		}
	})

	t.Run("tmdb 401 maps to unauthorized", func(t *testing.T) {
		err := &tmdb.StatusError{Code: http.StatusUnauthorized}
		if got := classifyReachabilityError(err); got != reasonUnauthorized {
			t.Errorf("expected %s, got %s", reasonUnauthorized, got)
		}
	})

	t.Run("non-401 status error maps to unavailable", func(t *testing.T) {
		err := &radarr.StatusError{Code: http.StatusInternalServerError, Body: "boom"}
		if got := classifyReachabilityError(err); got != reasonUnavailable {
			t.Errorf("expected %s, got %s", reasonUnavailable, got)
		}
	})

	t.Run("connection refused maps to unreachable", func(t *testing.T) {
		err := &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connection refused")}
		if got := classifyReachabilityError(err); got != reasonUnreachable {
			t.Errorf("expected %s, got %s", reasonUnreachable, got)
		}
	})

	t.Run("dns error maps to unreachable", func(t *testing.T) {
		err := &net.DNSError{IsNotFound: true}
		if got := classifyReachabilityError(err); got != reasonUnreachable {
			t.Errorf("expected %s, got %s", reasonUnreachable, got)
		}
	})

	t.Run("generic error maps to unavailable", func(t *testing.T) {
		err := errors.New("boom")
		if got := classifyReachabilityError(err); got != reasonUnavailable {
			t.Errorf("expected %s, got %s", reasonUnavailable, got)
		}
	})
}

func newSystemStatusTestConfig(t *testing.T, radarrURL, radarrKey, sonarrURL, sonarrKey string) {
	t.Helper()
	tempDir := t.TempDir()
	config.SetConfig(&config.Config{
		Radarr: config.RadarrConfig{URL: radarrURL, APIKey: radarrKey},
		Sonarr: config.SonarrConfig{URL: sonarrURL, APIKey: sonarrKey},
		Downloads: config.DownloadsConfig{
			MoviesPath:  filepath.Join(tempDir, "movies"),
			TVShowsPath: filepath.Join(tempDir, "tvshows"),
		},
	})
}

func getSystemStatusResponse(t *testing.T, server *Server) SystemStatusResponse {
	t.Helper()
	req, _ := http.NewRequest("GET", "/api/v1/system/status", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp SystemStatusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return resp
}

func TestSystemStatus_OneDependencyFailureDoesNotAffectOthers(t *testing.T) {
	setupTestDB(t)

	sonarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/system/status" {
			t.Errorf("expected sonarr path /api/v3/system/status, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	}))
	defer sonarrServer.Close()

	// Radarr points at an address nothing listens on, simulating an
	// unreachable upstream, while Sonarr is fully reachable and configured.
	newSystemStatusTestConfig(t, "http://127.0.0.1:1", "radarr-key", sonarrServer.URL, "sonarr-key")
	server := NewServer()

	resp := getSystemStatusResponse(t, server)

	if resp.Database.Status != statusOK {
		t.Errorf("expected database ok, got %+v", resp.Database)
	}
	if resp.Radarr.Status != statusKO {
		t.Errorf("expected radarr ko, got %+v", resp.Radarr)
	}
	if resp.Sonarr.Status != statusOK {
		t.Errorf("expected sonarr unaffected by radarr's failure, got %+v", resp.Sonarr)
	}
	if resp.TMDB.Status != statusNotConfigured {
		t.Errorf("expected tmdb not_configured (disabled by default), got %+v", resp.TMDB)
	}
}

func TestSystemStatus_ReportsBuildMetadata(t *testing.T) {
	setupTestDB(t)

	originalVersion, originalCommit, originalDate := buildVersion, buildCommit, buildDate
	SetBuildInfo("1.2.3", "abc1234", "2026-09-10_12:00:00")
	defer func() { SetBuildInfo(originalVersion, originalCommit, originalDate) }()

	newSystemStatusTestConfig(t, "", "", "", "")
	server := NewServer()

	resp := getSystemStatusResponse(t, server)

	if resp.Version != "1.2.3" || resp.Commit != "abc1234" || resp.Date != "2026-09-10_12:00:00" {
		t.Errorf("expected build metadata in response, got %+v", resp)
	}
}

func TestSystemStatus_NotConfiguredSkipsOutboundCall(t *testing.T) {
	setupTestDB(t)

	calledRadarr := false
	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledRadarr = true
		w.WriteHeader(http.StatusOK)
	}))
	defer radarrServer.Close()

	// The URL is set, but the API key is empty: still "not configured" per
	// the URL/APIKey convention shared with the other Radarr/Sonarr
	// endpoints, and must never be dialed.
	newSystemStatusTestConfig(t, radarrServer.URL, "", "", "")
	server := NewServer()

	resp := getSystemStatusResponse(t, server)

	if resp.Radarr.Status != statusNotConfigured {
		t.Errorf("expected radarr not_configured, got %+v", resp.Radarr)
	}
	if resp.Sonarr.Status != statusNotConfigured {
		t.Errorf("expected sonarr not_configured, got %+v", resp.Sonarr)
	}
	if calledRadarr {
		t.Error("expected no outbound call to a not-configured Radarr")
	}
}

func TestSystemStatus_BoundedByPerCheckTimeout(t *testing.T) {
	setupTestDB(t)

	original := systemStatusCheckTimeout
	systemStatusCheckTimeout = 30 * time.Millisecond
	defer func() { systemStatusCheckTimeout = original }()

	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	newSystemStatusTestConfig(t, slowServer.URL, "radarr-key", "", "")
	server := NewServer()

	start := time.Now()
	resp := getSystemStatusResponse(t, server)
	elapsed := time.Since(start)

	if elapsed > 300*time.Millisecond {
		t.Errorf("expected response bounded by the per-check timeout (~30ms), took %v", elapsed)
	}
	if resp.Radarr.Status != statusKO || resp.Radarr.Reason != reasonTimeout {
		t.Errorf("expected radarr ko/timeout, got %+v", resp.Radarr)
	}
}
