package main

import (
	"time"

	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/gorm"
)

// startJobRun creates a job_runs row with status "in_progress" for the given
// action, recording the moment the invocation starts. Written unconditionally,
// independent of whether Prometheus metrics exposition is enabled.
func startJobRun(db *gorm.DB, action string) (*models.JobRun, error) {
	run := &models.JobRun{
		Action:    action,
		Status:    "in_progress",
		StartedAt: time.Now(),
	}
	if err := db.Create(run).Error; err != nil {
		return nil, err
	}
	return run, nil
}

// finishJobRun finalizes a job_runs row with its outcome status, item counts,
// and optional error message.
func finishJobRun(db *gorm.DB, run *models.JobRun, status string, succeeded, failed, skipped int, errMsg string) {
	now := time.Now()
	run.Status = status
	run.CompletedAt = &now
	run.SucceededCount = succeeded
	run.FailedCount = failed
	run.SkippedCount = skipped
	if errMsg != "" {
		run.ErrorMessage = &errMsg
	}
	db.Save(run)
}
