package models_test

import (
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// legacyProcessingLog mirrors processing_logs before the metadata-backfill
// columns were added, to simulate migrating a pre-existing table.
type legacyProcessingLog struct {
	ID        uint      `gorm:"primaryKey"`
	Action    string    `gorm:"type:varchar(100);not null"`
	ItemCount int       `gorm:"not null;default:0"`
	Status    string    `gorm:"type:varchar(50);not null"`
	StartedAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (legacyProcessingLog) TableName() string {
	return "processing_logs"
}

// 3.1: AutoMigrate adds the new nullable metadata-backfill columns to an
// existing processing_logs table without touching pre-existing rows, which
// read the new fields back as null rather than 0.
func TestProcessingLog_AutoMigrate_AddsBackfillColumnsWithoutTouchingExistingRows(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.AutoMigrate(&legacyProcessingLog{}))
	preExisting := legacyProcessingLog{Action: "process_m3u", Status: "success", StartedAt: time.Now()}
	require.NoError(t, db.Create(&preExisting).Error)

	require.NoError(t, db.AutoMigrate(&models.ProcessingLog{}))

	var reloaded models.ProcessingLog
	require.NoError(t, db.First(&reloaded, preExisting.ID).Error)
	require.Equal(t, "process_m3u", reloaded.Action)
	require.Nil(t, reloaded.MetadataBackfilledCount)
	require.Nil(t, reloaded.MetadataBackfillErrorsCount)
}
