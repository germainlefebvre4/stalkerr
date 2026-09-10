## Why

Stalkeer's batch jobs (`process`, `download`, `m3u-download`, `resume-downloads`, `enrich-tvdb`) already run unattended on a schedule (Compose one-shots or Kubernetes CronJobs), but their outcomes are only visible by reading logs or querying the database by hand. The `server` Deployment also already reserves a dedicated `admin` port (8081, wired through `docker-compose.yml` and the Helm chart's Service/Deployment) that nothing has ever listened on. Exposing a Prometheus-format `/metrics` endpoint on that port — for a user who already runs their own Prometheus — turns "did last night's sync actually succeed" into a scrapeable, alertable signal instead of a manual log check, using data the application already computes or persists.

## What Changes

- Add a new `job_runs` table (GORM-managed, alongside the existing `processing_logs` table) that persists start/end time, status, and succeeded/failed/skipped counts for every `resume-downloads`, `enrich-tvdb`, and `backfill-metadata` invocation. This history is written unconditionally, independent of whether metrics exposition is enabled — it has standalone value as a durable run log for commands whose outcomes are today only logged and discarded.
- Add a config-gated `GET /metrics` endpoint, served on the existing admin port, in Prometheus text-exposition format, covering:
  - Run status, duration, and item counts (total processed / matched / skipped, both as a cumulative counter and as a last-run gauge) for every scheduled/batch action — sourced from `processing_logs` (`process`, `download`, `m3u-download`) and the new `job_runs` table (`resume-downloads`, `enrich-tvdb`, `backfill-metadata`) under one unified metric namespace and `action` label.
  - Downloads by status, a retry-count distribution, and total bytes downloaded, sourced from the existing `download_info` table.
  - The TMDB client's circuit breaker state and failure count (in-process; the only circuit breaker that lives inside the long-running `server` process).
- Add a `metrics` configuration block (`metrics.enabled`, `metrics.port`, `metrics.path`) following the existing `Enabled bool` pattern used by `tmdb`/`radarr`/`sonarr`/`m3u.download`. Default `enabled: false`. When disabled, the admin port's `/metrics` route is not registered at all (the endpoint is absent, not merely empty or unauthenticated).
- No Prometheus or Grafana provisioning is added anywhere (Helm chart, Compose, docs) — scraping this endpoint remains entirely the operator's own responsibility.

## Capabilities

### New Capabilities
- `job-run-history`: Persists a `job_runs` row per invocation of `resume-downloads`, `enrich-tvdb`, and `backfill-metadata`, recording status, duration, and succeeded/failed/skipped counts — independent of and useful without the metrics endpoint.
- `prometheus-metrics`: A config-gated `GET /metrics` endpoint on the admin port exposing job-run, download, and TMDB circuit-breaker signals in Prometheus text format.

### Modified Capabilities
(none — the admin port's existence and the ConfigMap delivery mechanism are unchanged; only new behavior is added on top of both, matching how prior integrations such as TMDB/Radarr/Sonarr added their own config blocks without amending `configuration-management`)

## Impact

- **New code**: `internal/models` (new `JobRun` model), a small metrics collector package (SQL-aggregate queries against `processing_logs`/`job_runs`/`download_info`, plus a TMDB circuit-breaker accessor), a new route/listener on the admin port in `internal/api`.
- **Modified code**: `internal/config` (new `MetricsConfig` struct), `cmd/resume_downloads.go`, `cmd/enrich_tvdb.go`, the backfill-metadata command (writes to `job_runs` on start/finish), `internal/database/database.go` (register `JobRun` in `AutoMigrate`), `internal/external/tmdb` (expose circuit-breaker state), `config.yml.example`.
- **Deployment surfaces**: `docker-compose.yml` and `charts/stalkerr/values.yaml` gain the new `metrics.*` keys (documented, default off); no new ports, services, or infra resources — the admin port is already provisioned everywhere it needs to be.
- **Out of scope**: Radarr/Sonarr live monitoring stats (expensive live upstream calls — not scrape-safe without a caching layer, not requested), per-request HTTP latency/error metrics for the API server, the M3U downloader's circuit breaker (lives in the short-lived `m3u-download` process, unreachable from `server`), and any Prometheus/Grafana provisioning.
