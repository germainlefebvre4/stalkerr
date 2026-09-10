package models

import "time"

// JobRun represents a durable run history entry for a standalone CLI command
// invocation (e.g. resume-downloads, enrich-tvdb) whose stats would otherwise
// only be logged and lost once the process exits.
type JobRun struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Action         string     `gorm:"type:varchar(100);not null" json:"action"`
	Status         string     `gorm:"type:varchar(50);not null" json:"status"` // "success", "failed", "in_progress"
	StartedAt      time.Time  `gorm:"not null" json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	SucceededCount int        `gorm:"not null;default:0" json:"succeeded_count"`
	FailedCount    int        `gorm:"not null;default:0" json:"failed_count"`
	SkippedCount   int        `gorm:"not null;default:0" json:"skipped_count"`
	ErrorMessage   *string    `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt      time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"not null" json:"updated_at"`
}

// TableName specifies the table name for JobRun
func (JobRun) TableName() string {
	return "job_runs"
}
