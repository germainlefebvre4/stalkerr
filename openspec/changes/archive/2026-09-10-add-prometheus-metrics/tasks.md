## 1. Configuration

- [x] 1.1 Add `MetricsConfig{Enabled bool, Port int, Path string}` to `internal/config/config.go` (defaults: `false`, `8081`, `/metrics`), wired the same way as `TMDBConfig`/`RadarrConfig`; verify with a config test asserting the defaults apply when the `metrics` block is absent
- [x] 1.2 Document the new keys in `config.yml.example` under a `metrics:` block, `enabled: false` by default
- [x] 1.3 Add the `metrics.*` keys to `docker-compose.yml`'s shared environment defaults and `charts/stalkerr/values.yaml`'s `config` block, both defaulting to disabled; verify `helm template` (or the chart's existing test/lint command) still succeeds

## 2. `job_runs` persistence (resume-downloads, enrich-tvdb)

- [x] 2.1 Add `models.JobRun` (table `job_runs`: `Action`, `Status`, `StartedAt`, `CompletedAt *time.Time`, `SucceededCount`, `FailedCount`, `SkippedCount`, `ErrorMessage *string`, `CreatedAt`, `UpdatedAt`) and register it in `internal/database/database.go`'s `AutoMigrate` list; verify the table is created on a fresh test database
- [x] 2.2 Wire `cmd/resume_downloads.go` to create a `job_runs` row (`status=in_progress`) at start and finalize it (`success`/`failed`, counts from `ResumeStats`, error message on crash) at completion, independent of `cfg.Metrics.Enabled`; verify with a test asserting a `job_runs` row exists after a run with metrics disabled
- [x] 2.3 Wire `cmd/enrich_tvdb.go` the same way, using `EnrichTVDBStats`; verify a partial/failed run still persists the counts accumulated up to the failure (per the `job-run-history` spec's "crash partway" scenario)
- [x] 2.4 Add a test covering a `job_runs` entry left `in_progress` (process still running) to confirm it's queryable with no completion time before finalization

## 3. `processing_logs` metadata-backfill columns (process)

- [x] 3.1 Add `MetadataBackfilledCount *int` and `MetadataBackfillErrorsCount *int` to `models.ProcessingLog`; verify `AutoMigrate` adds the columns to the existing table without touching existing rows (they read back `null`)
- [x] 3.2 Update `internal/processor/processor.go` (around the `BackfillRichMetadata` call at `processor.go:230`) to persist `BackfillStats.Updated`/`BackfillStats.Errors` onto the run's `processing_logs` entry when finalized; verify with a test asserting a `process` run's `processing_logs` row carries non-null backfill counts after a run that performs a backfill

## 4. TMDB circuit breaker accessor

- [x] 4.1 Add a method on `internal/external/tmdb.Client` exposing its circuit breaker's `State()` and `Failures()`; verify with a unit test that the accessor reflects the breaker's state after forcing a transition (e.g. simulated failures tripping it open)

## 5. Metrics collector

- [x] 5.1 Add `github.com/prometheus/client_golang` to `go.mod`
- [x] 5.2 Implement a custom `prometheus.Collector` (new package, e.g. `internal/metrics`) whose `Collect()` runs, per scrape: latest-row and `SUM`-across-history queries against `processing_logs` and `job_runs` keyed by action, a `COUNT ... GROUP BY status` and `SUM(bytes_downloaded)` query against `download_info`, a bucketed `COUNT(*) WHERE retry_count <= le` query for the retry distribution, and (when the TMDB client is non-nil) the circuit-breaker accessor from task 4.1; verify with a test that feeds a seeded test database and asserts the emitted metric families/labels/values match the `prometheus-metrics` spec's scenarios (last-run status/duration, cumulative vs. snapshot item counts including the backfill item types, downloads by status, retry buckets, breaker state)
- [x] 5.3 Verify the collector omits the TMDB circuit-breaker metrics entirely when the TMDB client is nil (spec scenario: "Circuit breaker metrics absent when TMDB is disabled")

## 6. Admin-port listener

- [x] 6.1 In `internal/api`, add a second `*http.Server` bound to `cfg.Metrics.Port`, started from `cmd/server.go` only when `cfg.Metrics.Enabled`, serving `cfg.Metrics.Path` via `promhttp.Handler()` wrapping the collector from task 5.2; registered with the existing `shutdown.Handler.Register` pattern; verify with a test/manual run that `GET :8081/metrics` returns `200` with a Prometheus text body when enabled
- [x] 6.2 Verify the metrics port is never opened when `cfg.Metrics.Enabled` is `false` (e.g. a connection attempt to the port fails) — covers the spec's "Metrics endpoint absent when disabled" scenario

## 7. End-to-end verification

- [x] 7.1 Run `openspec validate add-prometheus-metrics --strict` and confirm it still passes after implementation
- [x] 7.2 Manually exercise the full path with `metrics.enabled: true` in a local Compose stack: trigger `process`, `resume-downloads`, `enrich-tvdb`, and a download, then curl `:8081/metrics` and confirm every metric family described in `specs/prometheus-metrics/spec.md` appears with plausible values
- [x] 7.3 Confirm `docker-compose.yml` and the Helm chart still start cleanly with the new config keys left at their defaults (`metrics.enabled: false`), i.e. no behavior change for existing deployments
