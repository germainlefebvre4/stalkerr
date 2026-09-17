package models

import "time"

// BandwidthScheduleWindow is one weekly bandwidth schedule window: one or
// more days of the week, a start and end time-of-day (server local time,
// "HH:MM" 24h format), and the action to apply while the window is active.
// An end time earlier than the start time denotes a window spanning past
// midnight. See the download-bandwidth-schedule spec.
type BandwidthScheduleWindow struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	DaysOfWeek StringList `gorm:"type:text;not null" json:"days_of_week"`
	StartTime  string     `gorm:"type:varchar(5);not null" json:"start_time"`
	EndTime    string     `gorm:"type:varchar(5);not null" json:"end_time"`
	Action     string     `gorm:"type:varchar(20);not null" json:"action"`
	CreatedAt  time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"not null" json:"updated_at"`
}

// TableName specifies the table name for BandwidthScheduleWindow.
func (BandwidthScheduleWindow) TableName() string {
	return "bandwidth_schedule_windows"
}
