// Package metrics implements a custom prometheus.Collector exposing job-run,
// download, and TMDB circuit-breaker signals for stalkeer's admin-port
// /metrics endpoint. See openspec/changes/add-prometheus-metrics/design.md
// for the rationale behind deriving every metric from durable state at
// scrape time instead of accumulating counters as events happen.
package metrics

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/glefebvre/stalkeer/internal/circuitbreaker"
	"github.com/glefebvre/stalkeer/internal/external/tmdb"
	"github.com/glefebvre/stalkeer/internal/models"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

const namespace = "stalkeer"

// retryBuckets are the retry-count boundaries used for the download retry
// distribution, chosen to match the small integer range DownloadInfo.RetryCount
// actually takes (bounded by Downloads.RetryAttempts/MaxRetryAttempts config).
var retryBuckets = []int{0, 1, 2, 3, 5}

// processingLogActions maps the raw processing_logs.action column values this
// collector recognizes to the action label value exposed on metrics, keeping
// labels drawn from a small, fixed, known set rather than whatever strings
// happen to be in the table.
var processingLogActions = map[string]string{
	"process_m3u":  "process",
	"download":     "download",
	"m3u-download": "m3u-download",
}

// jobRunActions are the job_runs.action column values this collector
// recognizes; the raw value is also the exposed action label.
var jobRunActions = []string{"resume-downloads", "enrich-tvdb"}

// breakerStates are the circuit breaker states exposed as a labeled gauge.
var breakerStates = []circuitbreaker.State{
	circuitbreaker.StateClosed,
	circuitbreaker.StateOpen,
	circuitbreaker.StateHalfOpen,
}

// Collector is a custom prometheus.Collector that derives all its metrics at
// scrape time from durable state (processing_logs, job_runs, download_info)
// and, when available, the TMDB client's in-process circuit breaker.
type Collector struct {
	db         *gorm.DB
	tmdbClient *tmdb.Client

	lastRunSuccess    *prometheus.Desc
	lastRunDuration   *prometheus.Desc
	itemsTotal        *prometheus.Desc
	itemsLastRun      *prometheus.Desc
	downloadsByStatus *prometheus.Desc
	downloadsBytes    *prometheus.Desc
	retryBucket       *prometheus.Desc
	breakerState      *prometheus.Desc
	breakerFailures   *prometheus.Desc
}

// New creates a metrics Collector. tmdbClient may be nil (e.g. TMDB
// integration disabled), in which case Collect omits the TMDB circuit
// breaker metrics entirely.
func New(db *gorm.DB, tmdbClient *tmdb.Client) *Collector {
	return &Collector{
		db:         db,
		tmdbClient: tmdbClient,

		lastRunSuccess: prometheus.NewDesc(
			namespace+"_job_last_run_success",
			"Whether the action's most recent run completed successfully (1) or not (0).",
			[]string{"action"}, nil,
		),
		lastRunDuration: prometheus.NewDesc(
			namespace+"_job_last_run_duration_seconds",
			"Duration of the action's most recent completed run, in seconds.",
			[]string{"action"}, nil,
		),
		itemsTotal: prometheus.NewDesc(
			namespace+"_items_total",
			"Cumulative count of items by type across all historical runs of an action.",
			[]string{"action", "item_type"}, nil,
		),
		itemsLastRun: prometheus.NewDesc(
			namespace+"_items_last_run",
			"Count of items by type from the action's most recent run only.",
			[]string{"action", "item_type"}, nil,
		),
		downloadsByStatus: prometheus.NewDesc(
			namespace+"_downloads",
			"Current count of tracked downloads by status.",
			[]string{"status"}, nil,
		),
		downloadsBytes: prometheus.NewDesc(
			namespace+"_downloads_bytes_downloaded",
			"Current total bytes downloaded across all tracked downloads (a snapshot, not a monotonic total).",
			nil, nil,
		),
		retryBucket: prometheus.NewDesc(
			namespace+"_download_retry_count_bucket",
			"Cumulative count of tracked downloads with a retry count less than or equal to le.",
			[]string{"le"}, nil,
		),
		breakerState: prometheus.NewDesc(
			namespace+"_tmdb_circuit_breaker_state",
			"TMDB client circuit breaker state (1 for the current state, 0 for the others). Absent entirely when TMDB is disabled.",
			[]string{"name", "state"}, nil,
		),
		breakerFailures: prometheus.NewDesc(
			namespace+"_tmdb_circuit_breaker_failures",
			"TMDB client circuit breaker current failure count. Absent entirely when TMDB is disabled.",
			[]string{"name"}, nil,
		),
	}
}

// Describe implements prometheus.Collector.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.lastRunSuccess
	ch <- c.lastRunDuration
	ch <- c.itemsTotal
	ch <- c.itemsLastRun
	ch <- c.downloadsByStatus
	ch <- c.downloadsBytes
	ch <- c.retryBucket
	ch <- c.breakerState
	ch <- c.breakerFailures
}

// Collect implements prometheus.Collector, running its queries against the
// current database state on every scrape (a pull model - see design.md).
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	for rawAction, label := range processingLogActions {
		c.collectProcessingLogAction(ch, rawAction, label)
	}
	for _, action := range jobRunActions {
		c.collectJobRunAction(ch, action)
	}

	c.collectDownloads(ch)
	c.collectRetryBuckets(ch)

	if c.tmdbClient != nil {
		c.collectCircuitBreaker(ch)
	}
}

func (c *Collector) collectProcessingLogAction(ch chan<- prometheus.Metric, rawAction, label string) {
	var latest models.ProcessingLog
	if err := c.db.Where("action = ?", rawAction).Order("started_at DESC").Limit(1).Find(&latest).Error; err != nil || latest.ID == 0 {
		return
	}

	c.emitLastRun(ch, label, latest.Status, latest.StartedAt, latest.CompletedAt)

	snapshot := map[string]int{
		"movies":                   intOrZero(latest.MoviesCount),
		"tv_shows":                 intOrZero(latest.TVShowsCount),
		"new_items":                intOrZero(latest.NewItemsCount),
		"tmdb_matched":             intOrZero(latest.TMDBMatchedCount),
		"tmdb_unmatched":           intOrZero(latest.TMDBUnmatchedCount),
		"metadata_backfilled":      intOrZero(latest.MetadataBackfilledCount),
		"metadata_backfill_errors": intOrZero(latest.MetadataBackfillErrorsCount),
	}
	for itemType, value := range snapshot {
		ch <- prometheus.MustNewConstMetric(c.itemsLastRun, prometheus.GaugeValue, float64(value), label, itemType)
	}

	var sums struct {
		Movies                 sql.NullInt64 `gorm:"column:movies"`
		TVShows                sql.NullInt64 `gorm:"column:tv_shows"`
		NewItems               sql.NullInt64 `gorm:"column:new_items"`
		TMDBMatched            sql.NullInt64 `gorm:"column:tmdb_matched"`
		TMDBUnmatched          sql.NullInt64 `gorm:"column:tmdb_unmatched"`
		MetadataBackfilled     sql.NullInt64 `gorm:"column:metadata_backfilled"`
		MetadataBackfillErrors sql.NullInt64 `gorm:"column:metadata_backfill_errors"`
	}
	c.db.Model(&models.ProcessingLog{}).Where("action = ?", rawAction).
		Select(`
			SUM(movies_count) AS movies,
			SUM(tv_shows_count) AS tv_shows,
			SUM(new_items_count) AS new_items,
			SUM(tmdb_matched_count) AS tmdb_matched,
			SUM(tmdb_unmatched_count) AS tmdb_unmatched,
			SUM(metadata_backfilled_count) AS metadata_backfilled,
			SUM(metadata_backfill_errors_count) AS metadata_backfill_errors
		`).Scan(&sums)

	cumulative := map[string]int64{
		"movies":                   sums.Movies.Int64,
		"tv_shows":                 sums.TVShows.Int64,
		"new_items":                sums.NewItems.Int64,
		"tmdb_matched":             sums.TMDBMatched.Int64,
		"tmdb_unmatched":           sums.TMDBUnmatched.Int64,
		"metadata_backfilled":      sums.MetadataBackfilled.Int64,
		"metadata_backfill_errors": sums.MetadataBackfillErrors.Int64,
	}
	for itemType, value := range cumulative {
		ch <- prometheus.MustNewConstMetric(c.itemsTotal, prometheus.CounterValue, float64(value), label, itemType)
	}
}

func (c *Collector) collectJobRunAction(ch chan<- prometheus.Metric, action string) {
	var latest models.JobRun
	if err := c.db.Where("action = ?", action).Order("started_at DESC").Limit(1).Find(&latest).Error; err != nil || latest.ID == 0 {
		return
	}

	c.emitLastRun(ch, action, latest.Status, latest.StartedAt, latest.CompletedAt)

	snapshot := map[string]int{
		"succeeded": latest.SucceededCount,
		"failed":    latest.FailedCount,
		"skipped":   latest.SkippedCount,
	}
	for itemType, value := range snapshot {
		ch <- prometheus.MustNewConstMetric(c.itemsLastRun, prometheus.GaugeValue, float64(value), action, itemType)
	}

	var sums struct {
		Succeeded sql.NullInt64 `gorm:"column:succeeded"`
		Failed    sql.NullInt64 `gorm:"column:failed"`
		Skipped   sql.NullInt64 `gorm:"column:skipped"`
	}
	c.db.Model(&models.JobRun{}).Where("action = ?", action).
		Select("SUM(succeeded_count) AS succeeded, SUM(failed_count) AS failed, SUM(skipped_count) AS skipped").
		Scan(&sums)

	cumulative := map[string]int64{
		"succeeded": sums.Succeeded.Int64,
		"failed":    sums.Failed.Int64,
		"skipped":   sums.Skipped.Int64,
	}
	for itemType, value := range cumulative {
		ch <- prometheus.MustNewConstMetric(c.itemsTotal, prometheus.CounterValue, float64(value), action, itemType)
	}
}

func (c *Collector) emitLastRun(ch chan<- prometheus.Metric, action, status string, startedAt time.Time, completedAt *time.Time) {
	success := 0.0
	if status == "success" {
		success = 1.0
	}
	ch <- prometheus.MustNewConstMetric(c.lastRunSuccess, prometheus.GaugeValue, success, action)

	// A still-running latest run has no duration to report yet - a pull
	// endpoint on the always-up server cannot observe a separate process's
	// live progress, only its outcome once persisted (see design.md).
	if completedAt != nil {
		duration := completedAt.Sub(startedAt).Seconds()
		ch <- prometheus.MustNewConstMetric(c.lastRunDuration, prometheus.GaugeValue, duration, action)
	}
}

func (c *Collector) collectDownloads(ch chan<- prometheus.Metric) {
	var statusCounts []struct {
		Status string `gorm:"column:status"`
		Count  int64  `gorm:"column:count"`
	}
	c.db.Model(&models.DownloadInfo{}).Select("status, COUNT(*) AS count").Group("status").Scan(&statusCounts)
	for _, sc := range statusCounts {
		ch <- prometheus.MustNewConstMetric(c.downloadsByStatus, prometheus.GaugeValue, float64(sc.Count), sc.Status)
	}

	var totalBytes struct {
		Total sql.NullInt64 `gorm:"column:total"`
	}
	c.db.Model(&models.DownloadInfo{}).Select("SUM(bytes_downloaded) AS total").Scan(&totalBytes)
	ch <- prometheus.MustNewConstMetric(c.downloadsBytes, prometheus.GaugeValue, float64(totalBytes.Total.Int64))
}

func (c *Collector) collectRetryBuckets(ch chan<- prometheus.Metric) {
	for _, boundary := range retryBuckets {
		var count int64
		c.db.Model(&models.DownloadInfo{}).Where("retry_count <= ?", boundary).Count(&count)
		ch <- prometheus.MustNewConstMetric(c.retryBucket, prometheus.GaugeValue, float64(count), strconv.Itoa(boundary))
	}
	var total int64
	c.db.Model(&models.DownloadInfo{}).Count(&total)
	ch <- prometheus.MustNewConstMetric(c.retryBucket, prometheus.GaugeValue, float64(total), "+Inf")
}

func (c *Collector) collectCircuitBreaker(ch chan<- prometheus.Metric) {
	status := c.tmdbClient.CircuitBreakerStatus()
	for _, state := range breakerStates {
		value := 0.0
		if status.State == state {
			value = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.breakerState, prometheus.GaugeValue, value, "tmdb", state.String())
	}
	ch <- prometheus.MustNewConstMetric(c.breakerFailures, prometheus.GaugeValue, float64(status.Failures), "tmdb")
}

func intOrZero(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
