# Proposal

## Why

The arr-suite page (Films, Séries, Résumé) now fails almost every time with `502 sonarr_unreachable` / `radarr_unreachable`, each after exactly 10 s. The Radarr/Sonarr instances answer slowly since their storage moved from iSCSI to NFS, and the monitoring endpoints reuse `existenceCheckTimeout` (10 s), a budget designed for a single fail-fast lookup during a forced download. Even a single `GET /api/v3/series` (71 series) or `GET /api/v3/movie` (89 movies) no longer fits. The infrastructure cannot be fixed right now, and the handlers also swallow the underlying error, so the real cause (timeout, refused connection, open breaker) is invisible in the logs.

## What Changes

- Introduce a dedicated, hard-coded upstream time budget for the Radarr/Sonarr monitoring endpoints: 30 s per upstream call and per handler, 60 s for the stats endpoint (which shares one context across its Radarr and Sonarr phases).
- Raise the HTTP server `WriteTimeout` from 15 s to 65 s so a slow-but-successful answer is not cut off by the server before the handler returns.
- Log the underlying error when fetching the full Radarr movie catalog or Sonarr series catalog fails in the listing handlers, instead of only returning the generic 502.
- `existenceCheckTimeout` (10 s) is unchanged and keeps serving force-download, resync, duplicate handling and system-status checks.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `radarr-sonarr-monitoring-api`: adds a requirement that the monitoring endpoints tolerate slow upstream responses (a bounded budget well above the fail-fast 10 s) and that an upstream fetch failure is logged with its cause.

## Impact

- Code: `internal/api/radarr_sonarr_handlers.go` (timeout constants, five handlers, failure logging), `internal/api/api.go` (`WriteTimeout`).
- No API contract, schema, frontend or configuration change.
- Behavior: the page can take up to 30 s to answer when an arr instance is degraded instead of failing at 10 s. This masks, but does not fix, the slow NFS-backed instances; a degraded service still trips the shared circuit breaker as before.
