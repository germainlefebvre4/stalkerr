package models

import "time"

// ManualMapping represents a persistent manual override/mapping for a malformed raw title
type ManualMapping struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	TvgName     string      `gorm:"type:varchar(255);not null;uniqueIndex:idx_manual_mappings_unique,composite:tvg_group" json:"tvg_name"`
	GroupTitle  string      `gorm:"type:varchar(255);not null;uniqueIndex:idx_manual_mappings_unique,composite:tvg_group" json:"group_title"`
	ContentType ContentType `gorm:"type:varchar(20);not null" json:"content_type"`
	TMDBID      int         `gorm:"not null" json:"tmdb_id"`
	Season      *int        `json:"season,omitempty"`
	Episode     *int        `json:"episode,omitempty"`
	CreatedAt   time.Time   `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time   `gorm:"not null" json:"updated_at"`
}

// TableName specifies the table name for ManualMapping
func (ManualMapping) TableName() string {
	return "manual_mappings"
}
