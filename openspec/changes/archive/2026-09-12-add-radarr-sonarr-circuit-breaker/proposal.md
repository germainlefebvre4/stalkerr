## Why

TMDB and the M3U downloader already protect themselves from a downed dependency with `internal/circuitbreaker`: after a few consecutive failures they fail fast instead of letting every caller wait out a full timeout. Radarr and Sonarr have no such protection. Worse, their built-in retry (`retry.Do(..., apperrors.IsRetryable)`) almost never actually retries: the raw errors returned by `internal/external/httpclient` are never classified into `apperrors.AppError`, so `apperrors.IsRetryable` returns false on the first failure and the retry loop exits immediately. During a Radarr/Sonarr outage today, every one of the ~14 call sites across the API handlers and the `download` command independently blocks for its full timeout on every request, and the system status page's Radarr/Sonarr checks are just as blind between calls. Since the shared `httpclient` package (added by `extract-shared-http-helpers`) now gives every Radarr/Sonarr call a single execution chokepoint, both gaps can be closed by reusing the existing, already-tested circuit breaker rather than building something new.

## What Changes

- Add a shared circuit breaker to `internal/external/httpclient`: `Get`, `GetPage`, and `Put` execute the HTTP call through it, so every Radarr and Sonarr request is protected without per-call-site wiring.
- Fix the `apperrors.IsRetryable` classification gap: errors from `httpclient` (network/connection failures, timeouts, non-2xx responses) are wrapped into `apperrors.AppError` with the appropriate code (e.g. service-unavailable, timeout, rate-limited) before reaching `retry.Do`'s classifier, so transient failures are actually retried before counting against the breaker.
- Give each service (Radarr, Sonarr) one shared, long-lived breaker instance instead of a fresh one per call: owned by `api.Server` (built once in `NewServer()`, following the existing `tmdbClient` pattern) for the API server, and built once per run in `cmd/download.go`. Every `radarr.New`/`sonarr.New` call site is threaded a breaker via a new `Config.Breaker` field.
- Route the Radarr/Sonarr reachability check (`Client.SystemStatus`, used by the system status endpoint) through the same shared breaker instead of bypassing it as it does today.
- **BREAKING** (behavioral, not API-shape): when a service's breaker is open, the system status endpoint reports that service as KO immediately, without attempting a live reachability call - an explicit, documented exception to the endpoint's existing "compute every section fresh on each call" guarantee.

## Capabilities

### New Capabilities
- `radarr-sonarr-resilience`: circuit-breaker and retry-classification behavior shared by the Radarr and Sonarr clients - failure accounting, open/half-open/closed transitions, fail-fast behavior, and which errors are treated as retryable.

### Modified Capabilities
- `system-status-api`: the Radarr/Sonarr reachability check short-circuits to KO when that service's circuit breaker is open, instead of always performing a live check.

## Impact

- **Code**: `internal/external/httpclient/httpclient.go` (breaker execution, error classification), `internal/external/radarr/radarr.go`, `internal/external/sonarr/sonarr.go` (`Config.Breaker`, `SystemStatus` routed through the breaker), `internal/apperrors/errors.go` (or a new classification helper alongside it), `internal/api/api.go` (`Server` gains `radarrBreaker`/`sonarrBreaker`), `internal/api/system_status.go`, `internal/api/radarr_sonarr_handlers.go`, `internal/api/force_download.go`, `internal/api/resync_path.go` (thread the shared breaker into each `radarr.Config`/`sonarr.Config` literal), `cmd/download.go` (build one breaker per service per run).
- **Tests**: new unit tests for the breaker wiring and error classification in `httpclient`/`radarr`/`sonarr`; `internal/api/system_status_test.go` gains coverage for the open-circuit short-circuit path.
- **No** DB migrations, no config changes, no changes to the Radarr/Sonarr external API contracts, no frontend changes required (the status endpoint's response shape is unchanged, only the KO reason and latency in the open-circuit case).
