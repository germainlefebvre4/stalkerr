// Package settings resolves the application's "effective" configuration:
// file/env/default values from internal/config, with Tier-1 (applicative)
// fields overridden by rows stored in the database, and read-only
// visibility into which Tier-0 (bootstrap) fields cannot be overridden.
//
// internal/config cannot depend on internal/database (the reverse import
// already exists, for the DSN), so this package sits above both and is the
// one place the API server and every CLI command should read Tier-1
// configuration from, in place of config.Get(). See design.md's "A new
// internal/settings package hosts the merge".
package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/logger"
	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/gorm/clause"
)

// ErrUnknownField is returned when a key does not match any registered
// overridable field.
var ErrUnknownField = errors.New("unknown settings field")

// ErrBootstrapField is returned when a caller attempts to store or clear an
// override for a Tier-0 (bootstrap) field, which is read-only by design.
var ErrBootstrapField = errors.New("field is bootstrap configuration and cannot be overridden")

// Origin values reported for a field's effective source. See the
// app-settings spec's "Field Origin Exposure" requirement: file and
// environment-variable origins are not distinguished from each other.
const (
	OriginInterface = "interface"
	OriginConfig    = "config"
)

// Effective returns the application configuration with every stored Tier-1
// override applied on top of config.Get()'s file/env/default-resolved
// values, and its M3U source list replaced by the resolved effective list
// (see EffectiveSources). It never fails: a stored override that can no
// longer be decoded (e.g. its field was removed from the registry) is
// logged and skipped, leaving that field's file/env/default value in place.
func Effective() *config.Config {
	cfgCopy := *config.Get()

	for key, raw := range loadOverridesMap() {
		field, ok := FieldByKey(key)
		if !ok {
			continue
		}
		value, err := decodeValue(field.Kind, raw)
		if err != nil {
			logger.AppLogger().WithFields(map[string]interface{}{
				"key":   key,
				"error": err.Error(),
			}).Warn("failed to decode stored settings override, ignoring")
			continue
		}
		field.Set(&cfgCopy, value)
	}

	cfgCopy.M3U.Sources = EffectiveSources()
	return &cfgCopy
}

// Origin reports whether key's effective value currently comes from a
// stored override ("interface") or from file/env/default configuration
// ("config"). key may be any overridable or bootstrap field.
func Origin(key string) string {
	if hasOverride(key) {
		return OriginInterface
	}
	return OriginConfig
}

// SetOverride stores raw (a JSON-encoded scalar) as key's override,
// replacing any previous override for the same key (see the app-settings
// spec's "Re-Overriding a Field Replaces Its Previous Value"). It rejects a
// bootstrap key with ErrBootstrapField, a key outside the registry with
// ErrUnknownField, and a value that does not decode to the field's
// registered type.
func SetOverride(key string, raw json.RawMessage) error {
	if IsBootstrapKey(key) {
		return ErrBootstrapField
	}
	field, ok := FieldByKey(key)
	if !ok {
		return ErrUnknownField
	}
	if _, err := decodeValue(field.Kind, string(raw)); err != nil {
		return fmt.Errorf("invalid value for %s: %w", key, err)
	}

	row := models.SettingsOverride{Key: key, Value: string(raw), UpdatedAt: time.Now()}
	return database.Get().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&row).Error
}

// ClearOverride removes key's stored override, if any, reverting its
// effective value to file/env/default configuration. It is idempotent:
// clearing a key with no stored override is not an error. It rejects a
// bootstrap key with ErrBootstrapField and a key outside the registry with
// ErrUnknownField.
func ClearOverride(key string) error {
	if IsBootstrapKey(key) {
		return ErrBootstrapField
	}
	if _, ok := FieldByKey(key); !ok {
		return ErrUnknownField
	}
	return database.Get().Where("key = ?", key).Delete(&models.SettingsOverride{}).Error
}

func hasOverride(key string) bool {
	db := database.Get()
	if db == nil {
		return false
	}
	var row models.SettingsOverride
	return db.First(&row, "key = ?", key).Error == nil
}

func loadOverridesMap() map[string]string {
	m := make(map[string]string)
	db := database.Get()
	if db == nil {
		return m
	}
	var rows []models.SettingsOverride
	if err := db.Find(&rows).Error; err != nil {
		logger.AppLogger().WithFields(map[string]interface{}{"error": err.Error()}).Warn("failed to load settings overrides")
		return m
	}
	for _, r := range rows {
		m[r.Key] = r.Value
	}
	return m
}

func decodeValue(kind FieldKind, raw string) (interface{}, error) {
	switch kind {
	case KindString:
		var v string
		err := json.Unmarshal([]byte(raw), &v)
		return v, err
	case KindInt:
		var v int
		err := json.Unmarshal([]byte(raw), &v)
		return v, err
	case KindInt64:
		var v int64
		err := json.Unmarshal([]byte(raw), &v)
		return v, err
	case KindFloat64:
		var v float64
		err := json.Unmarshal([]byte(raw), &v)
		return v, err
	case KindBool:
		var v bool
		err := json.Unmarshal([]byte(raw), &v)
		return v, err
	default:
		return nil, fmt.Errorf("unsupported field kind %d", kind)
	}
}
