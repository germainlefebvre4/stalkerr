## Context

See `proposal.md` - Why for motivation. Relevant constraints established during exploration:

- The `server` binary (`internal/api`, Gin) is the only long-running process; `process`, `download`, `m3u-download`, `resume-downloads`, `enrich-tvdb` are one-shot CLI invocations (Compose one-shots or Kubernetes CronJobs) that start, run, and exit — they cannot themselves be scraped.
- `docker-compose.yml` and `charts/stalkerr/{values.yaml,templates/{service,deployment}.yaml}` already provision a second "admin" port (8081) on the `server` container/Service, distinct from the API port (8080). Nothing currently listens on it — `internal/api/api.go` registers no route for it.
- `internal/models.ProcessingLog` (table `processing_logs`) already persists per-run stats for `process`/`download`/`m3u-download` (see the `processing-run-statistics` capability). `resume-downloads`/`enrich-tvdb`/`backfill-metadata` compute equivalent stats in-memory (`ResumeStats`, `EnrichTVDBStats`, `BackfillStats`) but only log them — nothing persists past process exit.
- `internal/circuitbreaker.CircuitBreaker` has no registry; only the instance owned by `internal/external/tmdb.Client` lives inside the long-running `server` process (the M3U downloader's breaker lives inside the short-lived `m3u-download` process and is unreachable).
- The project uses GORM with `db.AutoMigrate(...)` (`internal/database/database.go`) — no separate migration tooling, so a new table is a new model added to that list.

## Goals / Non-Goals

**Goals:**
- Serve `/metrics` from the `server` process only, reading current state from Postgres and in-process structures at scrape time (pull model), rather than pushing/streaming events.
- Keep the feature fully inert (no open port, no behavior change) when `metrics.enabled` is `false`.
- Reuse existing persisted data (`processing_logs`, `download_info`) wherever it already exists; add the minimum new persistence (`job_runs`) needed for the three commands that currently discard their stats.

**Non-Goals:**
- Building or documenting a Prometheus/Grafana stack — scraping this endpoint is the operator's responsibility (see proposal.md).
- True Prometheus histograms built from in-process `Observe()` calls — retry-count and duration are reconstructed from durable state at scrape time, not accumulated as the events happen (see Decisions).
- Live, in-flight progress of a currently running batch job — a pull endpoint on the always-up `server` cannot observe a separate process's live progress; only its outcome once persisted.
- Radarr/Sonarr live stats and per-request HTTP metrics — out of scope per proposal.md.

## Decisions

### A second, independently-toggleable HTTP listener on the admin port
`api.Server` gains a second `*http.Server` bound to `cfg.Metrics.Port`, started only when `cfg.Metrics.Enabled`, registered with the same `shutdown.Handler` pattern already used for the main API server (`cmd/server.go`). It is a separate listener rather than a route on the existing 8080 router so that "disabled" means the port itself never opens — matching the `prometheus-metrics` spec's requirement that the endpoint be absent, not just unauthenticated, when off. It reuses the admin port already reserved end-to-end in Compose and the Helm chart, so no new port, Service entry, or container port needs provisioning.

**Alternative considered**: add `/metrics` as a route on the existing 8080 router, gated by a middleware check. Rejected — it would leave the port always open and mixes an internal/operational surface with the public API surface fronted by the same CORS policy.

### `client_golang` with a custom `Collector`, not instrumented counters
Use Prometheus's official Go client (`prometheus/client_golang`), but implement the job-run/download/circuit-breaker metrics as a custom `prometheus.Collector` whose `Collect()` runs the SQL aggregate queries and reads the TMDB breaker's state at scrape time — not as package-level counters updated via `Inc()`/`Observe()` calls scattered through the codebase. This matches the data's actual shape: it already lives durably in Postgres (or, for the circuit breaker, in a live in-process struct), so re-deriving it at scrape time is simpler and correct-by-construction (no risk of in-process counters drifting from the DB, no loss of state across `server` restarts for the DB-backed metrics).

**Alternative considered**: instrument each code path with `promauto` counters/histograms as events happen. Rejected for the DB-backed metrics — it would duplicate state that already exists durably in Postgres, and in-process counters reset on every `server` restart, which is a real risk for a personal deployment restarted often. Kept in mind as the right tool if per-request HTTP metrics are added later (explicitly out of scope here).

### Cumulative counters and last-run gauges both computed from history, not accumulated live
The cumulative counters (`stalkeer_items_total{action,item_type}` per the spec) are computed as `SUM(...)` over all historical `processing_logs`/`job_runs` rows for that action, re-summed on every scrape. This is a genuine Prometheus counter (monotonically non-decreasing as long as rows aren't deleted) despite not being accumulated via `Inc()`. Snapshot gauges read only the latest row per action (`ORDER BY started_at DESC LIMIT 1`).

**Trade-off accepted**: if operational cleanup ever deletes old `processing_logs`/`job_runs` rows (no such cleanup exists today), the cumulative counter would appear to reset, which Prometheus's `rate()`/`increase()` functions handle gracefully (they treat a decrease as a reset), so this is safe even though it's not the textbook pattern.

### `job_runs`: one new generic table, not an extension of `processing_logs`
New GORM model `models.JobRun` (table `job_runs`): `Action`, `Status`, `StartedAt`, `CompletedAt *time.Time`, `SucceededCount`, `FailedCount`, `SkippedCount`, `ErrorMessage *string`, plus `CreatedAt`/`UpdatedAt`. `resume-downloads`, `enrich-tvdb`, and the backfill-metadata command each create a row at start (`status=in_progress`) and finalize it at completion, mirroring `processing_logs`'s existing start/finalize lifecycle. Written unconditionally (see proposal.md), independent of `cfg.Metrics.Enabled`.

**Alternative considered**: add nullable `succeeded_count`/`failed_count`/`skipped_count` columns to `processing_logs` and reuse it for all six actions. Rejected — `processing_logs`'s existing columns (`MoviesCount`, `TMDBMatchedCount`, ...) are specific to the M3U processing pipeline's semantics; forcing `resume-downloads`'s "resumed/failed/skipped" into that shape (or vice versa) would misrepresent one or the other. A second table keeps each capability's persistence honest about what it actually measures, matching this repo's existing split between `processing-run-statistics` (persistence) and the capabilities that read it.

### TMDB circuit breaker: read directly, no new registry
`tmdb.Client` gains a method exposing its existing (already-private) `*circuitbreaker.CircuitBreaker`'s `State()`/`Failures()`. The metrics collector calls this only when `api.Server.tmdbClient != nil` (it is nil when TMDB is disabled — see `api.go`'s existing construction logic), satisfying the spec's "absent when TMDB disabled" scenario without a generic breaker registry, since exactly one breaker instance is ever reachable from `server`.

### Retry-count distribution as SQL-bucketed counts, not a real histogram
`stalkeer_download_retry_count_bucket{le=...}` is computed as `COUNT(*) WHERE retry_count <= le` for a small fixed set of bucket boundaries (e.g. `0, 1, 2, 3, 5, +Inf`) chosen to match the small integer range `DownloadInfo.RetryCount` actually takes (bounded by `Downloads.RetryAttempts`/`MaxRetryAttempts` config). This produces a metric shape compatible with `histogram_quantile()` in PromQL without needing a true client-side histogram (which would require instrumenting the download path directly and would reset on restart, losing the point of reading durable state — see the `client_golang` decision above).

### Labels are drawn from small, fixed, known sets
`action` (6 known CLI commands), `item_type`/`result` (a handful of known outcome kinds), `status` (`download_info`'s existing status enum), `name` (currently only `"tmdb"`). No label is ever derived from user-controlled or unbounded data (e.g. no per-item or per-filename labels), avoiding Prometheus cardinality blowup.

## Risks / Trade-offs

- **[Risk]** Every scrape runs several `COUNT`/`SUM`/`GROUP BY` queries against Postgres → at typical Prometheus scrape intervals (15-30s) against this project's data volumes (personal-scale, single user), the load is negligible. **Mitigation**: none built in; revisit with a short-TTL cache only if someone configures a sub-5s scrape interval, which is not a realistic personal-monitoring setup.
- **[Risk]** `resume-downloads` is a manual/recovery command today (the `download` command already resumes interrupted streams automatically per-stream), not part of the scheduled pipeline — its `job_runs` history will be sparse and its "last run" metrics may go stale for long periods under normal operation. **Mitigation**: none needed functionally; worth noting in any alerting the operator sets up so a stale `resume-downloads` timestamp isn't mistaken for a broken schedule.
- **[Risk]** Adding a second `http.Server`/listener inside `api.Server` slightly increases startup/shutdown complexity. **Mitigation**: reuse the existing `shutdown.Handler.Register` pattern already used for the primary server and the database, so shutdown ordering stays consistent and testable the same way.

## Migration Plan

- `models.JobRun` added to `internal/database/database.go`'s `AutoMigrate` list — purely additive (new table), applied automatically on next `server`/CLI startup like every other GORM-managed table today; no backfill needed since there is no historical data to migrate.
- `metrics` config block defaults `enabled: false` everywhere (`config.yml.example`, Compose `environment` defaults, Helm `values.yaml`) — deploying this change changes no runtime behavior until an operator opts in.
- Rollback: setting `metrics.enabled: false` (or leaving it unset) fully disables the new listener with no other effect; the `job_runs` table can be left in place harmlessly or dropped manually if the change is reverted entirely.

## Open Questions

- Exact Prometheus metric names/help text and the precise retry-count bucket boundaries are implementation-level naming choices that don't change any spec behavior — left to `tasks.md`/implementation.
