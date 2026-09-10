package main

import (
	"testing"

	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 2.4: a job_runs entry is queryable with no completion time from the moment
// the invocation starts, before it is finalized (job-run-history capability's
// "a running invocation is visible before completion" scenario).
func TestStartJobRun_VisibleBeforeCompletion(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.JobRun{}))
	database.SetDB(db)

	run, err := startJobRun(db, "resume-downloads")
	require.NoError(t, err)

	var reloaded models.JobRun
	require.NoError(t, db.First(&reloaded, run.ID).Error)
	require.Equal(t, "in_progress", reloaded.Status)
	require.Nil(t, reloaded.CompletedAt)
}
