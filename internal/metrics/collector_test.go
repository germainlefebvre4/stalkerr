package metrics

import (
	"testing"
	"time"

	"github.com/glefebvre/stalkeer/internal/external/tmdb"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCollectorTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.ProcessingLog{}, &models.JobRun{}, &models.DownloadInfo{}))
	return db
}

func gatherFamilies(t *testing.T, c *Collector) map[string]*dto.MetricFamily {
	t.Helper()

	reg := prometheus.NewPedanticRegistry()
	require.NoError(t, reg.Register(c))
	families, err := reg.Gather()
	require.NoError(t, err)

	out := make(map[string]*dto.MetricFamily, len(families))
	for _, f := range families {
		out[f.GetName()] = f
	}
	return out
}

func findMetric(family *dto.MetricFamily, labels map[string]string) *dto.Metric {
	if family == nil {
		return nil
	}
	for _, m := range family.Metric {
		got := make(map[string]string, len(m.Label))
		for _, lp := range m.Label {
			got[lp.GetName()] = lp.GetValue()
		}
		if len(got) != len(labels) {
			continue
		}
		match := true
		for k, v := range labels {
			if got[k] != v {
				match = false
				break
			}
		}
		if match {
			return m
		}
	}
	return nil
}

func metricValue(m *dto.Metric) float64 {
	if m == nil {
		return 0
	}
	if m.Gauge != nil {
		return m.Gauge.GetValue()
	}
	if m.Counter != nil {
		return m.Counter.GetValue()
	}
	return 0
}

// Last run status/duration exposed, sourced from job_runs for a
// job_runs-backed action.
func TestCollector_LastRunStatusAndDuration_JobRunsBacked(t *testing.T) {
	db := setupCollectorTestDB(t)
	started := time.Now().Add(-42 * time.Second)
	completed := started.Add(42 * time.Second)
	require.NoError(t, db.Create(&models.JobRun{
		Action: "enrich-tvdb", Status: "success",
		StartedAt: started, CompletedAt: &completed,
		SucceededCount: 3,
	}).Error)

	families := gatherFamilies(t, New(db, nil))

	success := findMetric(families["stalkeer_job_last_run_success"], map[string]string{"action": "enrich-tvdb"})
	require.NotNil(t, success)
	require.Equal(t, 1.0, metricValue(success))

	duration := findMetric(families["stalkeer_job_last_run_duration_seconds"], map[string]string{"action": "enrich-tvdb"})
	require.NotNil(t, duration)
	require.InDelta(t, 42.0, metricValue(duration), 0.5)
}

// Actions from both processing_logs and job_runs appear as different label
// values under the same metric family, not separate source-specific names.
func TestCollector_ActionsFromBothSourcesShareOneFamily(t *testing.T) {
	db := setupCollectorTestDB(t)
	now := time.Now()
	require.NoError(t, db.Create(&models.ProcessingLog{Action: "process_m3u", Status: "success", StartedAt: now, CompletedAt: &now}).Error)
	require.NoError(t, db.Create(&models.JobRun{Action: "resume-downloads", Status: "success", StartedAt: now, CompletedAt: &now}).Error)

	families := gatherFamilies(t, New(db, nil))
	family := families["stalkeer_job_last_run_success"]

	require.NotNil(t, findMetric(family, map[string]string{"action": "process"}))
	require.NotNil(t, findMetric(family, map[string]string{"action": "resume-downloads"}))
}

// Item counts: cumulative sums across history vs a snapshot of only the
// latest run, including the process action's metadata-backfill item types.
func TestCollector_ItemCountsCumulativeAndSnapshot(t *testing.T) {
	db := setupCollectorTestDB(t)
	now := time.Now()
	older := now.Add(-time.Hour)

	movies1, backfilled1, errors1 := 40, 3, 1
	require.NoError(t, db.Create(&models.ProcessingLog{
		Action: "process_m3u", Status: "success", StartedAt: older, CompletedAt: &older,
		MoviesCount: &movies1, MetadataBackfilledCount: &backfilled1, MetadataBackfillErrorsCount: &errors1,
	}).Error)

	movies2, backfilled2, errors2 := 12, 5, 1
	require.NoError(t, db.Create(&models.ProcessingLog{
		Action: "process_m3u", Status: "success", StartedAt: now, CompletedAt: &now,
		MoviesCount: &movies2, MetadataBackfilledCount: &backfilled2, MetadataBackfillErrorsCount: &errors2,
	}).Error)

	families := gatherFamilies(t, New(db, nil))

	cumulative := findMetric(families["stalkeer_items_total"], map[string]string{"action": "process", "item_type": "movies"})
	require.NotNil(t, cumulative)
	require.Equal(t, 52.0, metricValue(cumulative))

	snapshot := findMetric(families["stalkeer_items_last_run"], map[string]string{"action": "process", "item_type": "movies"})
	require.NotNil(t, snapshot)
	require.Equal(t, 12.0, metricValue(snapshot))

	backfillSnapshot := findMetric(families["stalkeer_items_last_run"], map[string]string{"action": "process", "item_type": "metadata_backfilled"})
	require.NotNil(t, backfillSnapshot)
	require.Equal(t, 5.0, metricValue(backfillSnapshot))

	backfillErrSnapshot := findMetric(families["stalkeer_items_last_run"], map[string]string{"action": "process", "item_type": "metadata_backfill_errors"})
	require.NotNil(t, backfillErrSnapshot)
	require.Equal(t, 1.0, metricValue(backfillErrSnapshot))
}

// Downloads exposed by status, plus total bytes downloaded as a snapshot.
func TestCollector_DownloadsByStatusAndBytes(t *testing.T) {
	db := setupCollectorTestDB(t)
	seed := func(status string, bytes int64) {
		require.NoError(t, db.Create(&models.DownloadInfo{Status: status, BytesDownloaded: &bytes}).Error)
	}
	for i := 0; i < 5; i++ {
		seed("completed", 100)
	}
	seed("failed", 0)
	seed("failed", 0)
	seed("in_progress", 50)

	families := gatherFamilies(t, New(db, nil))

	completed := findMetric(families["stalkeer_downloads"], map[string]string{"status": "completed"})
	require.NotNil(t, completed)
	require.Equal(t, 5.0, metricValue(completed))

	failed := findMetric(families["stalkeer_downloads"], map[string]string{"status": "failed"})
	require.NotNil(t, failed)
	require.Equal(t, 2.0, metricValue(failed))

	bytesFamily := families["stalkeer_downloads_bytes_downloaded"]
	require.NotNil(t, bytesFamily)
	require.Len(t, bytesFamily.Metric, 1)
	require.Equal(t, 550.0, metricValue(bytesFamily.Metric[0]))
}

// Retry-count distribution: bucketed counts consistent with the actual
// distribution of tracked downloads.
func TestCollector_RetryBucketsReflectDistribution(t *testing.T) {
	db := setupCollectorTestDB(t)
	seed := func(retry int) {
		require.NoError(t, db.Create(&models.DownloadInfo{Status: "pending", RetryCount: retry}).Error)
	}
	for i := 0; i < 10; i++ {
		seed(0)
	}
	for i := 0; i < 3; i++ {
		seed(1)
	}
	seed(5)

	families := gatherFamilies(t, New(db, nil))
	family := families["stalkeer_download_retry_count_bucket"]

	require.Equal(t, 10.0, metricValue(findMetric(family, map[string]string{"le": "0"})))
	require.Equal(t, 13.0, metricValue(findMetric(family, map[string]string{"le": "1"})))
	require.Equal(t, 13.0, metricValue(findMetric(family, map[string]string{"le": "3"})))
	require.Equal(t, 14.0, metricValue(findMetric(family, map[string]string{"le": "+Inf"})))
}

// TMDB circuit breaker state/failures exposed when a TMDB client is present.
func TestCollector_CircuitBreakerExposedWhenTMDBEnabled(t *testing.T) {
	db := setupCollectorTestDB(t)
	client := tmdb.NewClient(tmdb.Config{APIKey: "test-key"})

	families := gatherFamilies(t, New(db, client))

	closed := findMetric(families["stalkeer_tmdb_circuit_breaker_state"], map[string]string{"name": "tmdb", "state": "closed"})
	require.NotNil(t, closed)
	require.Equal(t, 1.0, metricValue(closed))

	open := findMetric(families["stalkeer_tmdb_circuit_breaker_state"], map[string]string{"name": "tmdb", "state": "open"})
	require.NotNil(t, open)
	require.Equal(t, 0.0, metricValue(open))

	failures := findMetric(families["stalkeer_tmdb_circuit_breaker_failures"], map[string]string{"name": "tmdb"})
	require.NotNil(t, failures)
	require.Equal(t, 0.0, metricValue(failures))
}

// 5.3: the collector omits the TMDB circuit-breaker metrics entirely when no
// TMDB client is available (TMDB integration disabled).
func TestCollector_CircuitBreakerAbsentWhenTMDBDisabled(t *testing.T) {
	db := setupCollectorTestDB(t)

	families := gatherFamilies(t, New(db, nil))

	require.Nil(t, families["stalkeer_tmdb_circuit_breaker_state"])
	require.Nil(t, families["stalkeer_tmdb_circuit_breaker_failures"])
}
