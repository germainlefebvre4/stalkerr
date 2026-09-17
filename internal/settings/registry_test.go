package settings

import "testing"

// expectedRegistryKeys is the full set of Tier-1 overridable fields per the
// app-settings spec's "Overridable Applicative Settings Fields" requirement.
var expectedRegistryKeys = []string{
	"radarr.url", "radarr.api_key", "radarr.enabled", "radarr.sync_interval", "radarr.quality_profile_id",
	"sonarr.url", "sonarr.api_key", "sonarr.enabled", "sonarr.sync_interval", "sonarr.quality_profile_id",
	"tmdb.api_key", "tmdb.language", "tmdb.enabled", "tmdb.requests_per_second",
	"jellyfin.url", "jellyfin.api_key", "jellyfin.enabled",
	"jellyfin.playback_check_enabled", "jellyfin.playback_action", "jellyfin.playback_poll_interval_seconds",
	"notifications.enabled", "notifications.ntfy.enabled", "notifications.ntfy.server_url",
	"notifications.ntfy.topic", "notifications.ntfy.auth_token",
	"downloads.movies_path", "downloads.tvshows_path", "downloads.temp_dir", "downloads.max_parallel",
	"downloads.timeout", "downloads.retry_attempts", "downloads.resume_enabled",
	"downloads.progress_interval_mb", "downloads.progress_interval_seconds", "downloads.lock_timeout_minutes",
	"downloads.max_retry_attempts", "downloads.min_file_size_mb", "downloads.force_tier_probability",
	"downloads.throttle_rate_kbps",
	"logging.app.level", "logging.database.level", "logging.format",
	"m3u.update_interval",
}

// expectedRestartRequiredKeys is the exact set the app-settings spec's
// "Restart-Required Field Exposure" requirement names.
var expectedRestartRequiredKeys = map[string]bool{
	"tmdb.api_key":               true,
	"tmdb.language":              true,
	"tmdb.requests_per_second":   true,
	"downloads.timeout":          true,
	"downloads.retry_attempts":   true,
	"downloads.min_file_size_mb": true,
}

func TestRegistry_MatchesExpectedKeyList(t *testing.T) {
	got := make(map[string]bool, len(registry))
	for _, f := range registry {
		if got[f.Key] {
			t.Errorf("duplicate registry key %q", f.Key)
		}
		got[f.Key] = true
	}

	want := make(map[string]bool, len(expectedRegistryKeys))
	for _, k := range expectedRegistryKeys {
		want[k] = true
	}

	for k := range want {
		if !got[k] {
			t.Errorf("expected registry key %q missing from registry", k)
		}
	}
	for k := range got {
		if !want[k] {
			t.Errorf("unexpected registry key %q not in expected list", k)
		}
	}
}

func TestRegistry_RestartRequiredFlagsMatchSpec(t *testing.T) {
	for _, f := range registry {
		want := expectedRestartRequiredKeys[f.Key]
		if f.RestartRequired != want {
			t.Errorf("field %q: RestartRequired = %v, want %v", f.Key, f.RestartRequired, want)
		}
	}
}

func TestRegistry_SensitiveFieldsAreFlagged(t *testing.T) {
	sensitive := map[string]bool{
		"radarr.api_key":               true,
		"sonarr.api_key":               true,
		"tmdb.api_key":                 true,
		"jellyfin.api_key":             true,
		"notifications.ntfy.auth_token": true,
	}
	for _, f := range registry {
		want := sensitive[f.Key]
		if f.Sensitive != want {
			t.Errorf("field %q: Sensitive = %v, want %v", f.Key, f.Sensitive, want)
		}
	}
}

func TestFieldByKey_UnknownKeyNotFound(t *testing.T) {
	if _, ok := FieldByKey("does.not.exist"); ok {
		t.Error("expected unknown key to be absent from registry")
	}
}

func TestIsBootstrapKey(t *testing.T) {
	for _, key := range []string{
		"database.host", "database.port", "database.user", "database.password",
		"database.dbname", "database.sslmode", "api.port", "metrics.port",
		"metrics.enabled", "metrics.path",
	} {
		if !IsBootstrapKey(key) {
			t.Errorf("expected %q to be a bootstrap key", key)
		}
	}
	if IsBootstrapKey("radarr.url") {
		t.Error("radarr.url must not be a bootstrap key")
	}
}
