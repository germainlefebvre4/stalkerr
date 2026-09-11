package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// withM3USourcesConfig writes a temp config.yaml containing a minimal
// m3u.sources list, chdirs into that directory for the duration of the test,
// and restores the original working directory on cleanup. This satisfies the
// m3u.sources non-empty requirement for tests that don't otherwise care about
// M3U configuration.
func withM3USourcesConfig(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	configContent := `
m3u:
  sources:
    - name: default
      file_path: /tmp/test.m3u
`
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir to temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(origWd)
	})
}

func TestLoad_WithDefaults(t *testing.T) {
	withM3USourcesConfig(t)

	// Set required environment variables
	os.Setenv("STALKEER_DATABASE_USER", "testuser")
	os.Setenv("STALKEER_DATABASE_DBNAME", "testdb")
	defer func() {
		os.Unsetenv("STALKEER_DATABASE_USER")
		os.Unsetenv("STALKEER_DATABASE_DBNAME")
	}()

	// Reset cfg to nil to force reload
	cfg = nil

	err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	config := Get()
	if config.Database.Host != "localhost" {
		t.Errorf("expected default host 'localhost', got %s", config.Database.Host)
	}
	if config.Database.Port != 5432 {
		t.Errorf("expected default port 5432, got %d", config.Database.Port)
	}
	if config.Logging.Level != "info" {
		t.Errorf("expected default log level 'info', got %s", config.Logging.Level)
	}
	if config.API.Port != 8080 {
		t.Errorf("expected default API port 8080, got %d", config.API.Port)
	}
	if config.Downloads.ForceTierProbability != 0.1 {
		t.Errorf("expected default force_tier_probability 0.1, got %v", config.Downloads.ForceTierProbability)
	}
	if config.Downloads.MinFileSizeMB != 1 {
		t.Errorf("expected default min_file_size_mb 1, got %v", config.Downloads.MinFileSizeMB)
	}
	if config.Jellyfin.Enabled != false {
		t.Errorf("expected jellyfin.enabled to default to false, got %v", config.Jellyfin.Enabled)
	}
	if config.Jellyfin.URL != "" {
		t.Errorf("expected jellyfin.url to default to empty, got %q", config.Jellyfin.URL)
	}
	if config.Jellyfin.APIKey != "" {
		t.Errorf("expected jellyfin.api_key to default to empty, got %q", config.Jellyfin.APIKey)
	}
	if config.Metrics.Enabled != false {
		t.Errorf("expected metrics.enabled to default to false, got %v", config.Metrics.Enabled)
	}
	if config.Metrics.Port != 8081 {
		t.Errorf("expected metrics.port to default to 8081, got %d", config.Metrics.Port)
	}
	if config.Metrics.Path != "/metrics" {
		t.Errorf("expected metrics.path to default to '/metrics', got %q", config.Metrics.Path)
	}
}

// TestLoad_NotificationsDefaults covers the notifications section's defaults
// separately from TestLoad_WithDefaults to keep that function's cyclomatic
// complexity within the project's configured limit.
func TestLoad_NotificationsDefaults(t *testing.T) {
	withM3USourcesConfig(t)

	os.Setenv("STALKEER_DATABASE_USER", "testuser")
	os.Setenv("STALKEER_DATABASE_DBNAME", "testdb")
	defer func() {
		os.Unsetenv("STALKEER_DATABASE_USER")
		os.Unsetenv("STALKEER_DATABASE_DBNAME")
	}()

	cfg = nil
	if err := Load(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	config := Get()
	if config.Notifications.Enabled != false {
		t.Errorf("expected notifications.enabled to default to false, got %v", config.Notifications.Enabled)
	}
	if config.Notifications.Ntfy.Enabled != false {
		t.Errorf("expected notifications.ntfy.enabled to default to false, got %v", config.Notifications.Ntfy.Enabled)
	}
	if config.Notifications.Ntfy.ServerURL != "" {
		t.Errorf("expected notifications.ntfy.server_url to default to empty, got %q", config.Notifications.Ntfy.ServerURL)
	}
	if config.Notifications.Ntfy.Topic != "" {
		t.Errorf("expected notifications.ntfy.topic to default to empty, got %q", config.Notifications.Ntfy.Topic)
	}
	if config.Notifications.Ntfy.AuthToken != "" {
		t.Errorf("expected notifications.ntfy.auth_token to default to empty, got %q", config.Notifications.Ntfy.AuthToken)
	}
}

func TestLoad_MinFileSizeMBOverride(t *testing.T) {
	withM3USourcesConfig(t)

	os.Setenv("STALKEER_DATABASE_USER", "testuser")
	os.Setenv("STALKEER_DATABASE_DBNAME", "testdb")
	os.Setenv("STALKEER_DOWNLOADS_MIN_FILE_SIZE_MB", "5")
	defer func() {
		os.Unsetenv("STALKEER_DATABASE_USER")
		os.Unsetenv("STALKEER_DATABASE_DBNAME")
		os.Unsetenv("STALKEER_DOWNLOADS_MIN_FILE_SIZE_MB")
	}()

	cfg = nil
	err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	config := Get()
	if config.Downloads.MinFileSizeMB != 5 {
		t.Errorf("expected overridden min_file_size_mb 5, got %v", config.Downloads.MinFileSizeMB)
	}
}

func TestLoad_NotificationsEnvOverride(t *testing.T) {
	withM3USourcesConfig(t)

	os.Setenv("STALKEER_DATABASE_USER", "testuser")
	os.Setenv("STALKEER_DATABASE_DBNAME", "testdb")
	os.Setenv("STALKEER_NOTIFICATIONS_ENABLED", "true")
	os.Setenv("STALKEER_NOTIFICATIONS_NTFY_ENABLED", "true")
	os.Setenv("STALKEER_NOTIFICATIONS_NTFY_SERVER_URL", "https://ntfy.sh")
	os.Setenv("STALKEER_NOTIFICATIONS_NTFY_TOPIC", "stalkeer-alerts")
	os.Setenv("STALKEER_NOTIFICATIONS_NTFY_AUTH_TOKEN", "tk_test_token")
	defer func() {
		os.Unsetenv("STALKEER_DATABASE_USER")
		os.Unsetenv("STALKEER_DATABASE_DBNAME")
		os.Unsetenv("STALKEER_NOTIFICATIONS_ENABLED")
		os.Unsetenv("STALKEER_NOTIFICATIONS_NTFY_ENABLED")
		os.Unsetenv("STALKEER_NOTIFICATIONS_NTFY_SERVER_URL")
		os.Unsetenv("STALKEER_NOTIFICATIONS_NTFY_TOPIC")
		os.Unsetenv("STALKEER_NOTIFICATIONS_NTFY_AUTH_TOKEN")
	}()

	cfg = nil
	if err := Load(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	config := Get()
	if !config.Notifications.Enabled {
		t.Error("expected notifications.enabled to be overridden to true")
	}
	if !config.Notifications.Ntfy.Enabled {
		t.Error("expected notifications.ntfy.enabled to be overridden to true")
	}
	if config.Notifications.Ntfy.ServerURL != "https://ntfy.sh" {
		t.Errorf("expected notifications.ntfy.server_url 'https://ntfy.sh', got %q", config.Notifications.Ntfy.ServerURL)
	}
	if config.Notifications.Ntfy.Topic != "stalkeer-alerts" {
		t.Errorf("expected notifications.ntfy.topic 'stalkeer-alerts', got %q", config.Notifications.Ntfy.Topic)
	}
	if config.Notifications.Ntfy.AuthToken != "tk_test_token" {
		t.Errorf("expected notifications.ntfy.auth_token 'tk_test_token', got %q", config.Notifications.Ntfy.AuthToken)
	}
}

func TestLoad_M3USourcesFromYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configContent := `
database:
  user: testuser
  dbname: testdb
m3u:
  sources:
    - name: provider-a
      file_path: /tmp/provider-a.m3u
      download:
        enabled: true
        url: "http://provider-a.example.com/playlist.m3u"
    - name: provider-b
      file_path: /tmp/provider-b.m3u
      download:
        enabled: true
        url: "http://provider-b.example.com/playlist.m3u"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir to temp dir: %v", err)
	}
	defer os.Chdir(origWd)

	cfg = nil
	if err := Load(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	config := Get()
	if len(config.M3U.Sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(config.M3U.Sources))
	}
	if config.M3U.Sources[0].Name != "provider-a" {
		t.Errorf("expected first source name 'provider-a', got %s", config.M3U.Sources[0].Name)
	}
	if config.M3U.Sources[0].FilePath != "/tmp/provider-a.m3u" {
		t.Errorf("expected first source file path '/tmp/provider-a.m3u', got %s", config.M3U.Sources[0].FilePath)
	}
	if config.M3U.Sources[1].Name != "provider-b" {
		t.Errorf("expected second source name 'provider-b', got %s", config.M3U.Sources[1].Name)
	}
	if config.M3U.Sources[1].Download.URL != "http://provider-b.example.com/playlist.m3u" {
		t.Errorf("expected second source download URL, got %s", config.M3U.Sources[1].Download.URL)
	}
}

func TestLoad_EmptySourcesRejected(t *testing.T) {
	// Reset viper's global state so no m3u.sources config leaks in from a
	// config file read by an earlier test in this package.
	viper.Reset()

	os.Setenv("STALKEER_DATABASE_USER", "testuser")
	os.Setenv("STALKEER_DATABASE_DBNAME", "testdb")
	defer func() {
		os.Unsetenv("STALKEER_DATABASE_USER")
		os.Unsetenv("STALKEER_DATABASE_DBNAME")
	}()

	cfg = nil
	err := Load()
	if err == nil {
		t.Fatalf("expected error when m3u.sources is absent, got nil")
	}
	if !strings.Contains(err.Error(), "m3u.sources must be a non-empty list") {
		t.Errorf("expected error about m3u.sources, got: %s", err.Error())
	}
}

func TestValidate_InvalidLogLevel(t *testing.T) {
	withM3USourcesConfig(t)

	os.Setenv("STALKEER_DATABASE_USER", "testuser")
	os.Setenv("STALKEER_DATABASE_DBNAME", "testdb")
	os.Setenv("STALKEER_LOGGING_LEVEL", "invalid")
	defer func() {
		os.Unsetenv("STALKEER_DATABASE_USER")
		os.Unsetenv("STALKEER_DATABASE_DBNAME")
		os.Unsetenv("STALKEER_LOGGING_LEVEL")
	}()

	cfg = nil
	err := Load()
	if err == nil {
		t.Fatalf("expected error for invalid log level, got nil")
	}
	if !strings.Contains(err.Error(), "logging.level must be one of") {
		t.Errorf("expected error about log level, got: %s", err.Error())
	}
}

func TestGetAppLogLevel_ModularConfig(t *testing.T) {
	cfg := &Config{
		Logging: LoggingConfig{
			App: LogLevelConfig{Level: "debug"},
		},
	}

	level := cfg.GetAppLogLevel()
	if level != "debug" {
		t.Errorf("expected app log level 'debug', got %s", level)
	}
}

func TestGetAppLogLevel_LegacyFallback(t *testing.T) {
	cfg := &Config{
		Logging: LoggingConfig{
			Level: "warn",
		},
	}

	level := cfg.GetAppLogLevel()
	if level != "warn" {
		t.Errorf("expected app log level 'warn' from legacy config, got %s", level)
	}
}

func TestGetAppLogLevel_DefaultFallback(t *testing.T) {
	cfg := &Config{
		Logging: LoggingConfig{},
	}

	level := cfg.GetAppLogLevel()
	if level != "info" {
		t.Errorf("expected default app log level 'info', got %s", level)
	}
}

func TestGetAppLogLevel_Priority(t *testing.T) {
	cfg := &Config{
		Logging: LoggingConfig{
			Level: "warn",
			App:   LogLevelConfig{Level: "debug"},
		},
	}

	level := cfg.GetAppLogLevel()
	if level != "debug" {
		t.Errorf("expected app.level to take priority over legacy level, got %s", level)
	}
}

func TestGetDatabaseLogLevel_ModularConfig(t *testing.T) {
	cfg := &Config{
		Logging: LoggingConfig{
			Database: LogLevelConfig{Level: "error"},
		},
	}

	level := cfg.GetDatabaseLogLevel()
	if level != "error" {
		t.Errorf("expected database log level 'error', got %s", level)
	}
}

func TestGetDatabaseLogLevel_LegacyFallback(t *testing.T) {
	cfg := &Config{
		Logging: LoggingConfig{
			Level: "debug",
		},
	}

	level := cfg.GetDatabaseLogLevel()
	if level != "debug" {
		t.Errorf("expected database log level 'debug' from legacy config, got %s", level)
	}
}

func TestIsUsingLegacyLogging_True(t *testing.T) {
	cfg := &Config{
		Logging: LoggingConfig{
			Level: "info",
		},
	}

	if !cfg.IsUsingLegacyLogging() {
		t.Error("expected IsUsingLegacyLogging to return true")
	}
}

func TestIsUsingLegacyLogging_False_WithModularConfig(t *testing.T) {
	cfg := &Config{
		Logging: LoggingConfig{
			Level: "info",
			App:   LogLevelConfig{Level: "debug"},
		},
	}

	if cfg.IsUsingLegacyLogging() {
		t.Error("expected IsUsingLegacyLogging to return false when modular config is set")
	}
}

func TestIsUsingLegacyLogging_False_NoLegacy(t *testing.T) {
	cfg := &Config{
		Logging: LoggingConfig{
			App:      LogLevelConfig{Level: "debug"},
			Database: LogLevelConfig{Level: "warn"},
		},
	}

	if cfg.IsUsingLegacyLogging() {
		t.Error("expected IsUsingLegacyLogging to return false when only modular config is set")
	}
}

func TestValidate_ModularLogLevels(t *testing.T) {
	tests := []struct {
		name        string
		appLevel    string
		dbLevel     string
		expectError bool
	}{
		{"valid debug levels", "debug", "debug", false},
		{"valid info levels", "info", "info", false},
		{"valid warn levels", "warn", "warn", false},
		{"valid error levels", "error", "error", false},
		{"invalid app level", "invalid", "info", true},
		{"invalid db level", "info", "invalid", true},
		{"both invalid", "bad", "wrong", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg = &Config{
				Database: DatabaseConfig{
					User:   "testuser",
					DBName: "testdb",
				},
				M3U: M3UConfig{
					Sources: []M3USourceConfig{{Name: "default", FilePath: "/tmp/test.m3u"}},
				},
				Logging: LoggingConfig{
					App:      LogLevelConfig{Level: tt.appLevel},
					Database: LogLevelConfig{Level: tt.dbLevel},
				},
			}

			err := validate()
			if tt.expectError && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}
