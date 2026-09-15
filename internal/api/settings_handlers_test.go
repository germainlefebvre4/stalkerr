package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/settings"
)

func setupSettingsTestConfig() {
	config.SetConfig(&config.Config{
		Radarr: config.RadarrConfig{
			URL:     "http://file-radarr.example.com",
			Enabled: true,
		},
		Database: config.DatabaseConfig{
			Host:     "db-host",
			Port:     5432,
			User:     "stalkeer",
			Password: "s3cret-db-pass",
			DBName:   "stalkeer",
			SSLMode:  "disable",
		},
		API:     config.APIConfig{Port: 8080},
		Metrics: config.MetricsConfig{Enabled: false, Port: 8081, Path: "/metrics"},
	})
}

func TestListSettings_ReturnsEffectiveValuesWithOrigin(t *testing.T) {
	setupTestDB(t)
	setupSettingsTestConfig()
	server := NewServer()

	req, _ := http.NewRequest("GET", "/api/v1/settings", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var body struct {
		Settings []SettingsFieldResponse `json:"settings"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	var radarrURL *SettingsFieldResponse
	for i := range body.Settings {
		if body.Settings[i].Key == "radarr.url" {
			radarrURL = &body.Settings[i]
		}
	}
	if radarrURL == nil {
		t.Fatal("expected radarr.url in settings list")
	}
	if radarrURL.Value != "http://file-radarr.example.com" {
		t.Errorf("radarr.url value = %v, want file value", radarrURL.Value)
	}
	if radarrURL.Origin != "config" {
		t.Errorf("radarr.url origin = %q, want config", radarrURL.Origin)
	}
}

func TestSetSettingsField_SuccessUpdatesOriginToInterface(t *testing.T) {
	setupTestDB(t)
	setupSettingsTestConfig()
	server := NewServer()

	body, _ := json.Marshal(SetSettingsFieldRequest{Value: json.RawMessage(`"http://override.example.com"`)})
	req, _ := http.NewRequest("PUT", "/api/v1/settings/radarr.url", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var field SettingsFieldResponse
	if err := json.Unmarshal(w.Body.Bytes(), &field); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if field.Origin != "interface" {
		t.Errorf("origin = %q, want interface", field.Origin)
	}
	if field.Value != "http://override.example.com" {
		t.Errorf("value = %v, want override", field.Value)
	}
}

func TestSetSettingsField_UnknownKeyReturns404(t *testing.T) {
	setupTestDB(t)
	setupSettingsTestConfig()
	server := NewServer()

	body, _ := json.Marshal(SetSettingsFieldRequest{Value: json.RawMessage(`"value"`)})
	req, _ := http.NewRequest("PUT", "/api/v1/settings/does.not.exist", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	var errResp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &errResp)
	if errResp.Error != "unknown_settings_field" {
		t.Errorf("error code = %q, want unknown_settings_field", errResp.Error)
	}
}

func TestSetSettingsField_BootstrapKeyReturns400(t *testing.T) {
	setupTestDB(t)
	setupSettingsTestConfig()
	server := NewServer()

	body, _ := json.Marshal(SetSettingsFieldRequest{Value: json.RawMessage(`"evil-host"`)})
	req, _ := http.NewRequest("PUT", "/api/v1/settings/database.host", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	var errResp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &errResp)
	if errResp.Error != "bootstrap_field_readonly" {
		t.Errorf("error code = %q, want bootstrap_field_readonly", errResp.Error)
	}
}

func TestClearSettingsField_RevertsToConfigValue(t *testing.T) {
	setupTestDB(t)
	setupSettingsTestConfig()
	server := NewServer()

	putBody, _ := json.Marshal(SetSettingsFieldRequest{Value: json.RawMessage(`"http://override.example.com"`)})
	putReq, _ := http.NewRequest("PUT", "/api/v1/settings/radarr.url", bytes.NewBuffer(putBody))
	putReq.Header.Set("Content-Type", "application/json")
	server.router.ServeHTTP(httptest.NewRecorder(), putReq)

	delReq, _ := http.NewRequest("DELETE", "/api/v1/settings/radarr.url", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, delReq)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var field SettingsFieldResponse
	json.Unmarshal(w.Body.Bytes(), &field)
	if field.Origin != "config" {
		t.Errorf("origin after clear = %q, want config", field.Origin)
	}
	if field.Value != "http://file-radarr.example.com" {
		t.Errorf("value after clear = %v, want file value", field.Value)
	}
}

func TestGetBootstrapSettings_MasksDatabasePassword(t *testing.T) {
	setupTestDB(t)
	setupSettingsTestConfig()
	server := NewServer()

	req, _ := http.NewRequest("GET", "/api/v1/settings/origin", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		Bootstrap []BootstrapFieldResponse `json:"bootstrap"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)

	var password *BootstrapFieldResponse
	for i := range body.Bootstrap {
		if body.Bootstrap[i].Key == "database.password" {
			password = &body.Bootstrap[i]
		}
	}
	if password == nil {
		t.Fatal("expected database.password in bootstrap response")
	}
	if password.Value != nil {
		t.Errorf("expected database.password value to be omitted, got %v", password.Value)
	}
	if password.IsSet == nil || !*password.IsSet {
		t.Error("expected database.password IsSet=true")
	}
	if password.Origin != "config" {
		t.Errorf("database.password origin = %q, want config", password.Origin)
	}
}

func TestSetSettingsField_LoggingOverrideAppliesLiveWithoutRestart(t *testing.T) {
	setupTestDB(t)
	setupSettingsTestConfig()

	// Baseline: app logging at "error", so a Debug call produces no output.
	logger.InitializeLoggersWithFormat("error", "error", "json")

	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = origStdout }()

	server := NewServer()
	body, _ := json.Marshal(SetSettingsFieldRequest{Value: json.RawMessage(`"debug"`)})
	req, _ := http.NewRequest("PUT", "/api/v1/settings/logging.app.level", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	respW := httptest.NewRecorder()
	server.router.ServeHTTP(respW, req)
	if respW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", respW.Code, respW.Body.String())
	}

	logger.AppLogger().Debug("visible-after-live-override")

	w.Close()
	os.Stdout = origStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)

	if !strings.Contains(buf.String(), "visible-after-live-override") {
		t.Errorf("expected debug log to be emitted immediately after the override, without a restart; got: %q", buf.String())
	}
}

func TestListSettings_RestartRequiredFlagMatchesExactFieldSet(t *testing.T) {
	setupTestDB(t)
	setupSettingsTestConfig()
	server := NewServer()

	req, _ := http.NewRequest("GET", "/api/v1/settings", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	var body struct {
		Settings []SettingsFieldResponse `json:"settings"`
	}
	json.Unmarshal(w.Body.Bytes(), &body)

	expectedRestartRequired := map[string]bool{
		"tmdb.api_key":               true,
		"tmdb.language":              true,
		"tmdb.requests_per_second":   true,
		"downloads.timeout":          true,
		"downloads.retry_attempts":   true,
		"downloads.min_file_size_mb": true,
	}
	for _, f := range body.Settings {
		want := expectedRestartRequired[f.Key]
		if f.RestartRequired != want {
			t.Errorf("field %q: restart_required = %v, want %v", f.Key, f.RestartRequired, want)
		}
	}
}

// 10.2: an override set through the settings API must be visible to
// whatever calls settings.Effective() next, with no restart of any
// long-lived process in between - this is exactly what a freshly-invoked
// CLI command (process/download/m3u-download) does on every run, since
// every Tier-1 call site now reads settings.Effective() instead of
// config.Get() (see design.md's Migration Plan and tasks 6.1/6.2).
func TestSettingsOverride_VisibleToEffectiveConfigWithoutRestart(t *testing.T) {
	setupTestDB(t)
	setupSettingsTestConfig()
	server := NewServer()

	body, _ := json.Marshal(SetSettingsFieldRequest{Value: json.RawMessage(`"http://cli-visible-override.example.com"`)})
	req, _ := http.NewRequest("PUT", "/api/v1/settings/radarr.url", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	eff := settings.Effective()
	if eff.Radarr.URL != "http://cli-visible-override.example.com" {
		t.Errorf("expected settings.Effective() to reflect the override set via the API, got %q", eff.Radarr.URL)
	}
}

// PUT to a bootstrap field via the settings endpoint must not succeed for
// any of the enumerated bootstrap keys, not just database.host.
func TestSetSettingsField_RejectsEveryBootstrapKey(t *testing.T) {
	setupTestDB(t)
	setupSettingsTestConfig()
	server := NewServer()

	for _, key := range []string{
		"database.host", "database.port", "database.user", "database.password",
		"database.dbname", "database.sslmode", "api.port", "metrics.port",
		"metrics.enabled", "metrics.path",
	} {
		body, _ := json.Marshal(SetSettingsFieldRequest{Value: json.RawMessage(`"x"`)})
		req, _ := http.NewRequest("PUT", "/api/v1/settings/"+key, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("key %q: expected 400, got %d", key, w.Code)
		}
	}
}
