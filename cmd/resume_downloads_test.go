package main

import (
	"context"
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/config"
	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/downloader"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupResumeDownloadsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&models.DownloadInfo{},
		&models.ProcessedLine{},
		&models.JobRun{},
	))

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	database.SetDB(db)
	return db
}

// 2.2: a resume-downloads run persists a job_runs row, independent of whether
// metrics exposition is enabled (job-run-history capability).
func TestRunResumeDownloads_PersistsJobRunWithMetricsDisabled(t *testing.T) {
	db := setupResumeDownloadsTestDB(t)
	config.SetConfig(&config.Config{Metrics: config.MetricsConfig{Enabled: false}})

	dl := downloader.New(5*time.Second, 3, 1)
	helper := downloader.NewResumeHelper(dl.GetStateManager(), dl)
	opts := downloader.ResumeOptions{DryRun: true}

	_, err := runResumeDownloads(context.Background(), db, helper, opts)
	require.NoError(t, err)

	var run models.JobRun
	require.NoError(t, db.Where("action = ?", "resume-downloads").First(&run).Error)
	require.Equal(t, "success", run.Status)
	require.NotNil(t, run.CompletedAt)
}
