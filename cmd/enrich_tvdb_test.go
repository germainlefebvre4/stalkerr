package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glefebvre/stalkeer/internal/database"
	"github.com/glefebvre/stalkeer/internal/external/tmdb"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/glefebvre/stalkeer/internal/processor"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupEnrichTVDBTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Movie{}, &models.TVShow{}, &models.JobRun{}))

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	database.SetDB(db)
	return db
}

// 2.3: a crash partway through an enrich-tvdb run still persists the
// succeeded/failed/skipped counts accumulated before the failure, plus a
// non-empty error message (job-run-history capability's "crash partway"
// scenario).
func TestRunEnrichTVDB_CrashPartwayPersistsPartialCounts(t *testing.T) {
	db := setupEnrichTVDBTestDB(t)

	// One movie that will be updated successfully before the crash.
	movie := models.Movie{TMDBID: 111, TMDBTitle: "Movie"}
	require.NoError(t, db.Create(&movie).Error)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tvdb_id": 555}`))
	}))
	defer srv.Close()

	tmdb.SetBaseURL(srv.URL)
	client := tmdb.NewClient(tmdb.Config{APIKey: "test-key", Language: "en-US"})

	// Simulate a fatal error partway through the run: the tvshows table
	// disappears after movies (which succeed) are processed but before the
	// tvshows step's query runs.
	require.NoError(t, db.Migrator().DropTable(&models.TVShow{}))

	_, err := runEnrichTVDB(db, client, processor.EnrichTVDBOptions{})
	require.Error(t, err)

	var run models.JobRun
	require.NoError(t, db.Where("action = ?", "enrich-tvdb").First(&run).Error)
	require.Equal(t, "failed", run.Status)
	require.Equal(t, 1, run.SucceededCount)
	require.NotNil(t, run.ErrorMessage)
	require.NotEmpty(t, *run.ErrorMessage)
}
