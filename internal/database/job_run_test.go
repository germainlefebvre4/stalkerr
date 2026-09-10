package database_test

import (
	"testing"

	"github.com/glefebvre/stalkeer/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestJobRun_AutoMigrate_CreatesTable(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(&models.JobRun{}); err != nil {
		t.Fatalf("failed to auto-migrate JobRun: %v", err)
	}

	if !db.Migrator().HasTable(&models.JobRun{}) {
		t.Fatal("expected job_runs table to exist after migration")
	}

	if err := db.Create(&models.JobRun{Action: "resume-downloads", Status: "in_progress"}).Error; err != nil {
		t.Fatalf("expected to insert into job_runs, got error: %v", err)
	}
}
