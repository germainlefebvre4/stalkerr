package models

import "time"

// SettingsOverride represents a single runtime-overridden application
// settings field, keyed by its dotted config path (e.g. "radarr.api_key").
// See the app-settings spec.
type SettingsOverride struct {
	Key       string    `gorm:"primaryKey;type:varchar(255)" json:"key"`
	Value     string    `gorm:"type:text;not null" json:"value"` // JSON-encoded scalar
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

// TableName specifies the table name for SettingsOverride
func (SettingsOverride) TableName() string {
	return "settings_overrides"
}
