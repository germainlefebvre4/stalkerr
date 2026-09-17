package settings

import "github.com/glefebvre/stalkeer/internal/config"

// FieldKind identifies the Go scalar type a registered field holds, so a
// stored JSON-encoded override value can be decoded into the right type.
type FieldKind int

const (
	KindString FieldKind = iota
	KindInt
	KindInt64
	KindFloat64
	KindBool
)

// FieldSpec describes one overridable (Tier-1) applicative configuration
// field: its dotted key, its type, whether its value is sensitive (never
// returned raw), whether changing it requires an application restart to
// take effect, and how to read/write it on a *config.Config.
//
// This is a hand-written registry rather than a reflection-based mapping:
// an invalid or unregistered key is rejected at the API boundary instead of
// silently no-oping, and adding a field is a one-line, compile-checked
// entry (see design.md's "An explicit, hand-written field registry").
type FieldSpec struct {
	Key             string
	Kind            FieldKind
	Sensitive       bool
	RestartRequired bool
	get             func(c *config.Config) interface{}
	set             func(c *config.Config, v interface{})
}

// Get reads this field's current value from c.
func (f FieldSpec) Get(c *config.Config) interface{} {
	return f.get(c)
}

// Set writes v (already decoded to this field's Kind) onto c.
func (f FieldSpec) Set(c *config.Config, v interface{}) {
	f.set(c, v)
}

// registry lists every Tier-1 (applicative) overridable field. See the
// app-settings spec's "Overridable Applicative Settings Fields" requirement.
var registry = []FieldSpec{
	// Radarr
	{Key: "radarr.url", Kind: KindString,
		get: func(c *config.Config) interface{} { return c.Radarr.URL },
		set: func(c *config.Config, v interface{}) { c.Radarr.URL = v.(string) }},
	{Key: "radarr.api_key", Kind: KindString, Sensitive: true,
		get: func(c *config.Config) interface{} { return c.Radarr.APIKey },
		set: func(c *config.Config, v interface{}) { c.Radarr.APIKey = v.(string) }},
	{Key: "radarr.enabled", Kind: KindBool,
		get: func(c *config.Config) interface{} { return c.Radarr.Enabled },
		set: func(c *config.Config, v interface{}) { c.Radarr.Enabled = v.(bool) }},
	{Key: "radarr.sync_interval", Kind: KindInt,
		get: func(c *config.Config) interface{} { return c.Radarr.SyncInterval },
		set: func(c *config.Config, v interface{}) { c.Radarr.SyncInterval = v.(int) }},
	{Key: "radarr.quality_profile_id", Kind: KindInt,
		get: func(c *config.Config) interface{} { return c.Radarr.QualityProfileID },
		set: func(c *config.Config, v interface{}) { c.Radarr.QualityProfileID = v.(int) }},

	// Sonarr
	{Key: "sonarr.url", Kind: KindString,
		get: func(c *config.Config) interface{} { return c.Sonarr.URL },
		set: func(c *config.Config, v interface{}) { c.Sonarr.URL = v.(string) }},
	{Key: "sonarr.api_key", Kind: KindString, Sensitive: true,
		get: func(c *config.Config) interface{} { return c.Sonarr.APIKey },
		set: func(c *config.Config, v interface{}) { c.Sonarr.APIKey = v.(string) }},
	{Key: "sonarr.enabled", Kind: KindBool,
		get: func(c *config.Config) interface{} { return c.Sonarr.Enabled },
		set: func(c *config.Config, v interface{}) { c.Sonarr.Enabled = v.(bool) }},
	{Key: "sonarr.sync_interval", Kind: KindInt,
		get: func(c *config.Config) interface{} { return c.Sonarr.SyncInterval },
		set: func(c *config.Config, v interface{}) { c.Sonarr.SyncInterval = v.(int) }},
	{Key: "sonarr.quality_profile_id", Kind: KindInt,
		get: func(c *config.Config) interface{} { return c.Sonarr.QualityProfileID },
		set: func(c *config.Config, v interface{}) { c.Sonarr.QualityProfileID = v.(int) }},

	// TMDB (restart-required: baked into the TMDB client at server boot)
	{Key: "tmdb.api_key", Kind: KindString, Sensitive: true, RestartRequired: true,
		get: func(c *config.Config) interface{} { return c.TMDB.APIKey },
		set: func(c *config.Config, v interface{}) { c.TMDB.APIKey = v.(string) }},
	{Key: "tmdb.language", Kind: KindString, RestartRequired: true,
		get: func(c *config.Config) interface{} { return c.TMDB.Language },
		set: func(c *config.Config, v interface{}) { c.TMDB.Language = v.(string) }},
	{Key: "tmdb.enabled", Kind: KindBool,
		get: func(c *config.Config) interface{} { return c.TMDB.Enabled },
		set: func(c *config.Config, v interface{}) { c.TMDB.Enabled = v.(bool) }},
	{Key: "tmdb.requests_per_second", Kind: KindFloat64, RestartRequired: true,
		get: func(c *config.Config) interface{} { return c.TMDB.RequestsPerSecond },
		set: func(c *config.Config, v interface{}) { c.TMDB.RequestsPerSecond = v.(float64) }},

	// Jellyfin
	{Key: "jellyfin.url", Kind: KindString,
		get: func(c *config.Config) interface{} { return c.Jellyfin.URL },
		set: func(c *config.Config, v interface{}) { c.Jellyfin.URL = v.(string) }},
	{Key: "jellyfin.api_key", Kind: KindString, Sensitive: true,
		get: func(c *config.Config) interface{} { return c.Jellyfin.APIKey },
		set: func(c *config.Config, v interface{}) { c.Jellyfin.APIKey = v.(string) }},
	{Key: "jellyfin.enabled", Kind: KindBool,
		get: func(c *config.Config) interface{} { return c.Jellyfin.Enabled },
		set: func(c *config.Config, v interface{}) { c.Jellyfin.Enabled = v.(bool) }},
	{Key: "jellyfin.playback_check_enabled", Kind: KindBool,
		get: func(c *config.Config) interface{} { return c.Jellyfin.PlaybackCheckEnabled },
		set: func(c *config.Config, v interface{}) { c.Jellyfin.PlaybackCheckEnabled = v.(bool) }},
	{Key: "jellyfin.playback_action", Kind: KindString,
		get: func(c *config.Config) interface{} { return c.Jellyfin.PlaybackAction },
		set: func(c *config.Config, v interface{}) { c.Jellyfin.PlaybackAction = v.(string) }},
	{Key: "jellyfin.playback_poll_interval_seconds", Kind: KindInt,
		get: func(c *config.Config) interface{} { return c.Jellyfin.PlaybackPollIntervalSeconds },
		set: func(c *config.Config, v interface{}) { c.Jellyfin.PlaybackPollIntervalSeconds = v.(int) }},

	// Notifications
	{Key: "notifications.enabled", Kind: KindBool,
		get: func(c *config.Config) interface{} { return c.Notifications.Enabled },
		set: func(c *config.Config, v interface{}) { c.Notifications.Enabled = v.(bool) }},
	{Key: "notifications.ntfy.enabled", Kind: KindBool,
		get: func(c *config.Config) interface{} { return c.Notifications.Ntfy.Enabled },
		set: func(c *config.Config, v interface{}) { c.Notifications.Ntfy.Enabled = v.(bool) }},
	{Key: "notifications.ntfy.server_url", Kind: KindString,
		get: func(c *config.Config) interface{} { return c.Notifications.Ntfy.ServerURL },
		set: func(c *config.Config, v interface{}) { c.Notifications.Ntfy.ServerURL = v.(string) }},
	{Key: "notifications.ntfy.topic", Kind: KindString,
		get: func(c *config.Config) interface{} { return c.Notifications.Ntfy.Topic },
		set: func(c *config.Config, v interface{}) { c.Notifications.Ntfy.Topic = v.(string) }},
	{Key: "notifications.ntfy.auth_token", Kind: KindString, Sensitive: true,
		get: func(c *config.Config) interface{} { return c.Notifications.Ntfy.AuthToken },
		set: func(c *config.Config, v interface{}) { c.Notifications.Ntfy.AuthToken = v.(string) }},

	// Downloads (three restart-required: baked into the Downloader instance at server boot)
	{Key: "downloads.movies_path", Kind: KindString,
		get: func(c *config.Config) interface{} { return c.Downloads.MoviesPath },
		set: func(c *config.Config, v interface{}) { c.Downloads.MoviesPath = v.(string) }},
	{Key: "downloads.tvshows_path", Kind: KindString,
		get: func(c *config.Config) interface{} { return c.Downloads.TVShowsPath },
		set: func(c *config.Config, v interface{}) { c.Downloads.TVShowsPath = v.(string) }},
	{Key: "downloads.temp_dir", Kind: KindString,
		get: func(c *config.Config) interface{} { return c.Downloads.TempDir },
		set: func(c *config.Config, v interface{}) { c.Downloads.TempDir = v.(string) }},
	{Key: "downloads.max_parallel", Kind: KindInt,
		get: func(c *config.Config) interface{} { return c.Downloads.MaxParallel },
		set: func(c *config.Config, v interface{}) { c.Downloads.MaxParallel = v.(int) }},
	{Key: "downloads.timeout", Kind: KindInt, RestartRequired: true,
		get: func(c *config.Config) interface{} { return c.Downloads.Timeout },
		set: func(c *config.Config, v interface{}) { c.Downloads.Timeout = v.(int) }},
	{Key: "downloads.retry_attempts", Kind: KindInt, RestartRequired: true,
		get: func(c *config.Config) interface{} { return c.Downloads.RetryAttempts },
		set: func(c *config.Config, v interface{}) { c.Downloads.RetryAttempts = v.(int) }},
	{Key: "downloads.resume_enabled", Kind: KindBool,
		get: func(c *config.Config) interface{} { return c.Downloads.ResumeEnabled },
		set: func(c *config.Config, v interface{}) { c.Downloads.ResumeEnabled = v.(bool) }},
	{Key: "downloads.progress_interval_mb", Kind: KindInt64,
		get: func(c *config.Config) interface{} { return c.Downloads.ProgressIntervalMB },
		set: func(c *config.Config, v interface{}) { c.Downloads.ProgressIntervalMB = v.(int64) }},
	{Key: "downloads.progress_interval_seconds", Kind: KindInt,
		get: func(c *config.Config) interface{} { return c.Downloads.ProgressIntervalSeconds },
		set: func(c *config.Config, v interface{}) { c.Downloads.ProgressIntervalSeconds = v.(int) }},
	{Key: "downloads.lock_timeout_minutes", Kind: KindInt,
		get: func(c *config.Config) interface{} { return c.Downloads.LockTimeoutMinutes },
		set: func(c *config.Config, v interface{}) { c.Downloads.LockTimeoutMinutes = v.(int) }},
	{Key: "downloads.max_retry_attempts", Kind: KindInt,
		get: func(c *config.Config) interface{} { return c.Downloads.MaxRetryAttempts },
		set: func(c *config.Config, v interface{}) { c.Downloads.MaxRetryAttempts = v.(int) }},
	{Key: "downloads.min_file_size_mb", Kind: KindInt64, RestartRequired: true,
		get: func(c *config.Config) interface{} { return c.Downloads.MinFileSizeMB },
		set: func(c *config.Config, v interface{}) { c.Downloads.MinFileSizeMB = v.(int64) }},
	{Key: "downloads.force_tier_probability", Kind: KindFloat64,
		get: func(c *config.Config) interface{} { return c.Downloads.ForceTierProbability },
		set: func(c *config.Config, v interface{}) { c.Downloads.ForceTierProbability = v.(float64) }},
	{Key: "downloads.throttle_rate_kbps", Kind: KindInt,
		get: func(c *config.Config) interface{} { return c.Downloads.ThrottleRateKbps },
		set: func(c *config.Config, v interface{}) { c.Downloads.ThrottleRateKbps = v.(int) }},

	// Logging (applied live: an override re-invokes logger.InitializeLoggersWithFormat)
	{Key: "logging.app.level", Kind: KindString,
		get: func(c *config.Config) interface{} { return c.Logging.App.Level },
		set: func(c *config.Config, v interface{}) { c.Logging.App.Level = v.(string) }},
	{Key: "logging.database.level", Kind: KindString,
		get: func(c *config.Config) interface{} { return c.Logging.Database.Level },
		set: func(c *config.Config, v interface{}) { c.Logging.Database.Level = v.(string) }},
	{Key: "logging.format", Kind: KindString,
		get: func(c *config.Config) interface{} { return c.Logging.Format },
		set: func(c *config.Config, v interface{}) { c.Logging.Format = v.(string) }},

	// M3U: the refresh interval is a scalar field, covered here; the
	// m3u.sources list itself is covered separately by the
	// m3u-source-overrides mechanism (see internal/settings/m3u_sources.go).
	{Key: "m3u.update_interval", Kind: KindInt,
		get: func(c *config.Config) interface{} { return c.M3U.UpdateInterval },
		set: func(c *config.Config, v interface{}) { c.M3U.UpdateInterval = v.(int) }},
}

var registryByKey = func() map[string]FieldSpec {
	m := make(map[string]FieldSpec, len(registry))
	for _, f := range registry {
		m[f.Key] = f
	}
	return m
}()

// FieldByKey returns the registered field spec for key, if any.
func FieldByKey(key string) (FieldSpec, bool) {
	f, ok := registryByKey[key]
	return f, ok
}

// Registry returns every overridable field spec.
func Registry() []FieldSpec {
	return registry
}

// bootstrapKeys lists every Tier-0 field: read-only, never overridable,
// because the system needs its file/env-resolved value before it can reach
// its database or bind its ports. See the app-settings spec's "Bootstrap
// Configuration Remains Read-Only" requirement.
var bootstrapKeys = map[string]bool{
	"database.host":     true,
	"database.port":     true,
	"database.user":     true,
	"database.password": true,
	"database.dbname":   true,
	"database.sslmode":  true,
	"api.port":          true,
	"metrics.port":      true,
	"metrics.enabled":   true,
	"metrics.path":      true,
}

// IsBootstrapKey reports whether key is a Tier-0 (bootstrap) field.
func IsBootstrapKey(key string) bool {
	return bootstrapKeys[key]
}

// excludedKeys lists scalar config.Config fields that are deliberately
// neither overridable (Tier-1) nor bootstrap (Tier-0): they exist in
// config.yml/env for backward compatibility but are not surfaced through
// the settings mechanism at all.
var excludedKeys = map[string]string{
	// Deprecated, superseded by logging.app.level/logging.database.level
	// (see config.Config.GetAppLogLevel/GetDatabaseLogLevel). Not exposed
	// for override so the new UI steers users to the non-deprecated fields
	// instead of encouraging continued use of the legacy one.
	"logging.level": "deprecated, superseded by logging.app.level / logging.database.level",
}

// IsExcludedKey reports whether key is a known scalar field intentionally
// excluded from both the overridable registry and the bootstrap list.
func IsExcludedKey(key string) bool {
	_, ok := excludedKeys[key]
	return ok
}
