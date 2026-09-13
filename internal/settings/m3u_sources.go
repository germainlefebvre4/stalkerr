package settings

import (
	"errors"
	"time"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/gorm/clause"
)

// ErrM3USourceNotFound is returned when deleting a name for which no
// runtime source is stored (whether or not an origin source of that name
// exists in config.yml). See the m3u-source-overrides spec's "Removing a
// Runtime Source" requirement.
var ErrM3USourceNotFound = errors.New("no runtime source stored for this name")

// M3USourceEntry pairs a resolved M3U source with whether it currently
// comes from a stored runtime row (true) or the config.yml-defined origin
// (false). See the m3u-source-overrides spec.
type M3USourceEntry struct {
	Source    config.M3USourceConfig
	IsRuntime bool
}

// HasAuthPassword reports whether this entry currently has a
// download.auth_password set, without exposing its value. See "Sensitive
// Source Field Masking".
func (e M3USourceEntry) HasAuthPassword() bool {
	return e.Source.Download.AuthPassword != ""
}

// M3USourceInput is the create/update payload for a runtime M3U source. A
// nil AuthPassword means "omitted": keep the name's current effective
// password unchanged. A non-nil pointer to "" clears it; any other value
// replaces it. See "Sensitive Source Field Masking".
type M3USourceInput struct {
	FilePath       string
	Enabled        bool
	URL            string
	ArchiveDir     string
	RetentionCount int
	MaxFileSizeMB  int64
	TimeoutSeconds int
	RetryAttempts  int
	AuthUsername   string
	AuthPassword   *string
}

// OriginSources returns the config.yml-defined M3U sources, independently
// of any stored runtime source. See "Origin M3U Source Exposure".
func OriginSources() []M3USourceEntry {
	src := config.Get().M3U.Sources
	entries := make([]M3USourceEntry, len(src))
	for i, s := range src {
		entries[i] = M3USourceEntry{Source: s, IsRuntime: false}
	}
	return entries
}

// EffectiveEntries resolves the effective M3U source list: for each name
// known from either origin configuration or stored runtime sources, the
// runtime source if one exists for that name, otherwise the origin source.
// Fields are never merged between an origin and a runtime source sharing a
// name. See "Effective Source List Resolution".
func EffectiveEntries() []M3USourceEntry {
	origin := config.Get().M3U.Sources
	runtimeByName := runtimeSourcesByName()

	var entries []M3USourceEntry
	seen := make(map[string]bool, len(origin))

	for _, s := range origin {
		if r, ok := runtimeByName[s.Name]; ok {
			entries = append(entries, M3USourceEntry{Source: rowToSource(r), IsRuntime: true})
		} else {
			entries = append(entries, M3USourceEntry{Source: s, IsRuntime: false})
		}
		seen[s.Name] = true
	}

	for name, r := range runtimeByName {
		if !seen[name] {
			entries = append(entries, M3USourceEntry{Source: rowToSource(r), IsRuntime: true})
		}
	}

	return entries
}

// EffectiveSources is EffectiveEntries stripped down to the plain
// config.M3USourceConfig shape internal consumers (the M3U downloader,
// Effective()) need, with real (unmasked) auth_password values - masking
// only applies at the API response layer.
func EffectiveSources() []config.M3USourceConfig {
	entries := EffectiveEntries()
	sources := make([]config.M3USourceConfig, len(entries))
	for i, e := range entries {
		sources[i] = e.Source
	}
	return sources
}

// SetSource stores a runtime M3U source for name, replacing any existing
// runtime source for the same name in place (upsert by name; never a
// duplicate, never merged field-by-field with a prior value other than
// auth_password's omission handling). See "Runtime Source, Keyed By Name".
func SetSource(name string, input M3USourceInput) (config.M3USourceConfig, error) {
	var password string
	if input.AuthPassword == nil {
		password = currentEffectivePassword(name)
	} else {
		password = *input.AuthPassword
	}

	row := models.M3USourceConfig{
		Name:           name,
		FilePath:       input.FilePath,
		Enabled:        input.Enabled,
		URL:            input.URL,
		ArchiveDir:     input.ArchiveDir,
		RetentionCount: input.RetentionCount,
		MaxFileSizeMB:  input.MaxFileSizeMB,
		TimeoutSeconds: input.TimeoutSeconds,
		RetryAttempts:  input.RetryAttempts,
		AuthUsername:   input.AuthUsername,
		IsRuntime:      true,
		UpdatedAt:      time.Now(),
	}
	if password != "" {
		row.AuthPassword = &password
	}

	if err := database.Get().Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"file_path", "enabled", "url", "archive_dir", "retention_count",
			"max_file_size_mb", "timeout_seconds", "retry_attempts",
			"auth_username", "auth_password", "is_runtime", "updated_at",
		}),
	}).Create(&row).Error; err != nil {
		return config.M3USourceConfig{}, err
	}

	return rowToSource(row), nil
}

// DeleteSource removes the stored runtime source for name. If an origin
// source of the same name exists, the effective source for that name
// reverts to it; otherwise the name disappears from the effective list
// entirely. Deleting a name with no stored runtime source (whether or not
// an origin source exists for it) returns ErrM3USourceNotFound and never
// affects the origin source. See "Removing a Runtime Source".
func DeleteSource(name string) error {
	result := database.Get().Where("name = ?", name).Delete(&models.M3USourceConfig{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrM3USourceNotFound
	}
	return nil
}

func currentEffectivePassword(name string) string {
	for _, e := range EffectiveEntries() {
		if e.Source.Name == name {
			return e.Source.Download.AuthPassword
		}
	}
	return ""
}

func runtimeSourcesByName() map[string]models.M3USourceConfig {
	m := make(map[string]models.M3USourceConfig)
	db := database.Get()
	if db == nil {
		return m
	}
	var rows []models.M3USourceConfig
	if err := db.Find(&rows).Error; err != nil {
		logger.AppLogger().WithFields(map[string]interface{}{"error": err.Error()}).Warn("failed to load M3U runtime sources")
		return m
	}
	for _, r := range rows {
		m[r.Name] = r
	}
	return m
}

func rowToSource(row models.M3USourceConfig) config.M3USourceConfig {
	var password string
	if row.AuthPassword != nil {
		password = *row.AuthPassword
	}
	return config.M3USourceConfig{
		Name:     row.Name,
		FilePath: row.FilePath,
		Download: config.M3UDownloadConfig{
			Enabled:        row.Enabled,
			URL:            row.URL,
			ArchiveDir:     row.ArchiveDir,
			RetentionCount: row.RetentionCount,
			MaxFileSizeMB:  row.MaxFileSizeMB,
			TimeoutSeconds: row.TimeoutSeconds,
			RetryAttempts:  row.RetryAttempts,
			AuthUsername:   row.AuthUsername,
			AuthPassword:   password,
		},
	}
}
