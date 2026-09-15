package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/settings"
)

func newIntegrationTestServer(t *testing.T) *Server {
	t.Helper()
	setupTestDB(t)
	tempDir := t.TempDir()
	config.SetConfig(&config.Config{
		Downloads: config.DownloadsConfig{
			MoviesPath:  filepath.Join(tempDir, "movies"),
			TVShowsPath: filepath.Join(tempDir, "tvshows"),
		},
	})
	return NewServer()
}

func postIntegrationTest(t *testing.T, server *Server, body integrationTestRequest) (*httptest.ResponseRecorder, ServiceStatus) {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/settings/integrations/test", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	var result ServiceStatus
	if w.Code == http.StatusOK {
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
	}
	return w, result
}

func TestTestIntegrationConnectivity_MissingURLRejected(t *testing.T) {
	server := newIntegrationTestServer(t)

	called := false
	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer radarrServer.Close()

	w, _ := postIntegrationTest(t, server, integrationTestRequest{Service: "radarr", URL: "", APIKey: "key"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing url, got %d: %s", w.Code, w.Body.String())
	}
	if called {
		t.Error("expected no outbound call for a request missing a url")
	}
}

func TestTestIntegrationConnectivity_UnsupportedServiceRejected(t *testing.T) {
	server := newIntegrationTestServer(t)

	called := false
	someServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer someServer.Close()

	w, _ := postIntegrationTest(t, server, integrationTestRequest{Service: "tmdb", URL: someServer.URL, APIKey: "key"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unsupported service, got %d: %s", w.Code, w.Body.String())
	}
	if called {
		t.Error("expected no outbound call for an unsupported service")
	}
}

func TestTestIntegrationConnectivity_RadarrOK(t *testing.T) {
	server := newIntegrationTestServer(t)

	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/system/status" {
			t.Errorf("expected radarr path /api/v3/system/status, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer radarrServer.Close()

	w, result := postIntegrationTest(t, server, integrationTestRequest{Service: "radarr", URL: radarrServer.URL, APIKey: "key"})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if result.Status != statusOK {
		t.Errorf("expected ok, got %+v", result)
	}
}

func TestTestIntegrationConnectivity_RadarrUnauthorized(t *testing.T) {
	server := newIntegrationTestServer(t)

	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid api key"))
	}))
	defer radarrServer.Close()

	_, result := postIntegrationTest(t, server, integrationTestRequest{Service: "radarr", URL: radarrServer.URL, APIKey: "bad-key"})

	if result.Status != statusKO || result.Reason != reasonUnauthorized {
		t.Errorf("expected ko/unauthorized, got %+v", result)
	}
}

func TestTestIntegrationConnectivity_RadarrUnreachable(t *testing.T) {
	server := newIntegrationTestServer(t)

	_, result := postIntegrationTest(t, server, integrationTestRequest{Service: "radarr", URL: "http://127.0.0.1:1", APIKey: "key"})

	if result.Status != statusKO || result.Reason != reasonUnreachable {
		t.Errorf("expected ko/unreachable, got %+v", result)
	}
}

func TestTestIntegrationConnectivity_RadarrTimeout(t *testing.T) {
	server := newIntegrationTestServer(t)

	original := systemStatusCheckTimeout
	systemStatusCheckTimeout = 30 * time.Millisecond
	defer func() { systemStatusCheckTimeout = original }()

	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	_, result := postIntegrationTest(t, server, integrationTestRequest{Service: "radarr", URL: slowServer.URL, APIKey: "key"})

	if result.Status != statusKO || result.Reason != reasonTimeout {
		t.Errorf("expected ko/timeout, got %+v", result)
	}
}

func TestTestIntegrationConnectivity_SonarrOK(t *testing.T) {
	server := newIntegrationTestServer(t)

	sonarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/system/status" {
			t.Errorf("expected sonarr path /api/v3/system/status, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer sonarrServer.Close()

	_, result := postIntegrationTest(t, server, integrationTestRequest{Service: "sonarr", URL: sonarrServer.URL, APIKey: "key"})

	if result.Status != statusOK {
		t.Errorf("expected ok, got %+v", result)
	}
}

func TestTestIntegrationConnectivity_JellyfinOK(t *testing.T) {
	server := newIntegrationTestServer(t)

	var gotToken string
	jellyfinServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/System/Info" {
			t.Errorf("expected jellyfin path /System/Info, got %s", r.URL.Path)
		}
		gotToken = r.Header.Get("X-Emby-Token")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer jellyfinServer.Close()

	_, result := postIntegrationTest(t, server, integrationTestRequest{Service: "jellyfin", URL: jellyfinServer.URL, APIKey: "jf-key"})

	if result.Status != statusOK {
		t.Errorf("expected ok, got %+v", result)
	}
	if gotToken != "jf-key" {
		t.Errorf("expected X-Emby-Token 'jf-key', got %q", gotToken)
	}
}

func TestTestIntegrationConnectivity_JellyfinUnauthorized(t *testing.T) {
	server := newIntegrationTestServer(t)

	jellyfinServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid api key"))
	}))
	defer jellyfinServer.Close()

	_, result := postIntegrationTest(t, server, integrationTestRequest{Service: "jellyfin", URL: jellyfinServer.URL, APIKey: "bad-key"})

	if result.Status != statusKO || result.Reason != reasonUnauthorized {
		t.Errorf("expected ko/unauthorized, got %+v", result)
	}
}

// TestTestIntegrationConnectivity_DoesNotPersistSettings verifies the
// integration-connectivity-test spec's "No Persistence" requirement: the
// endpoint never reads settings.Effective() (so the test's own supplied
// values are used, not the saved config) nor writes an override, even when
// called with values that differ from the currently saved configuration.
func TestTestIntegrationConnectivity_DoesNotPersistSettings(t *testing.T) {
	setupTestDB(t)
	tempDir := t.TempDir()
	config.SetConfig(&config.Config{
		Radarr: config.RadarrConfig{URL: "http://saved-radarr.example", APIKey: "saved-key"},
		Downloads: config.DownloadsConfig{
			MoviesPath:  filepath.Join(tempDir, "movies"),
			TVShowsPath: filepath.Join(tempDir, "tvshows"),
		},
	})
	server := NewServer()

	before := settings.Effective()

	radarrServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer radarrServer.Close()

	postIntegrationTest(t, server, integrationTestRequest{Service: "radarr", URL: radarrServer.URL, APIKey: "different-key"})

	after := settings.Effective()

	if before.Radarr.URL != after.Radarr.URL || before.Radarr.APIKey != after.Radarr.APIKey {
		t.Errorf("expected saved radarr settings unchanged, before=%+v after=%+v", before.Radarr, after.Radarr)
	}
	if after.Radarr.URL != "http://saved-radarr.example" || after.Radarr.APIKey != "saved-key" {
		t.Errorf("expected saved radarr settings to remain the originally configured values, got %+v", after.Radarr)
	}
}
