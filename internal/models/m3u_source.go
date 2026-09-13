package models

import "time"

// M3USourceConfig represents a runtime-stored M3U source, identified by
// name. A row here replaces the config.yml-defined source of the same name
// entirely (never merged field-by-field), or adds a new source that exists
// only at runtime if no config.yml source shares its name. See the
// m3u-source-overrides spec.
type M3USourceConfig struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	FilePath string `gorm:"type:text;not null" json:"file_path"`

	// Download settings, mirroring config.M3UDownloadConfig.
	Enabled        bool    `gorm:"not null;default:false" json:"enabled"`
	URL            string  `gorm:"type:text" json:"url"`
	ArchiveDir     string  `gorm:"type:text" json:"archive_dir"`
	RetentionCount int     `gorm:"not null;default:0" json:"retention_count"`
	MaxFileSizeMB  int64   `gorm:"not null;default:0" json:"max_file_size_mb"`
	TimeoutSeconds int     `gorm:"not null;default:0" json:"timeout_seconds"`
	RetryAttempts  int     `gorm:"not null;default:0" json:"retry_attempts"`
	AuthUsername   string  `gorm:"type:varchar(255)" json:"auth_username,omitempty"`
	AuthPassword   *string `gorm:"type:text" json:"-"` // sensitive; never serialized raw, see Sensitive Source Field Masking

	IsRuntime bool      `gorm:"not null;default:true;index" json:"is_runtime"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
}

// TableName specifies the table name for M3USourceConfig
func (M3USourceConfig) TableName() string {
	return "m3u_source_configs"
}
