## Context

See proposal.md - Why. Relevant current state:
- `database.HealthCheck()` already exists and backs `/health`.
- `internal/external/{radarr,sonarr,tmdb}` clients expose only business-data methods today; none has a reachability/ping call.
- Radarr/Sonarr clients are constructed **per-request** directly in `internal/api/radarr_sonarr_handlers.go` from `config.Get()` (checked via `cfg.Radarr.URL == "" || cfg.Radarr.APIKey == ""` for "not configured"); the TMDB client is instead built once in `NewServer()` and held on `Server.tmdbClient` (nil when disabled). The new endpoint follows both existing conventions rather than introducing a new client-lifecycle pattern.
- `internal/downloader/diskspace.go` already computes available/free/total/used% for a path via `unix.Statfs`, used today only as a pre-download guard.
- The frontend already has an error-code-driven translation pattern (`ApiError.code` + `useApiErrorMessage`) used elsewhere - reusing it keeps the KO reason handling consistent with the rest of the app instead of introducing a second, free-text error convention.

## Goals / Non-Goals

**Goals:**
- Aggregate DB/Radarr/Sonarr/TMDB/disk into one bounded-latency, on-demand response.
- Keep every check independent: one dependency's failure never affects another's reported status.
- Reuse existing conventions (per-request client construction, error-code translation) instead of introducing new ones.

**Non-Goals:**
- Surfacing Radarr/Sonarr's own internal `/api/v3/health` diagnostics (indexers, download clients, root folders). Deferred; the chosen check only validates that stalkeer can reach and authenticate against Radarr/Sonarr's API.
- Latency measurements, version numbers, or any metric beyond OK/KO/not-configured plus a short reason.
- Background polling, caching between requests, or any alerting/badge tied to health state.
- Historical/trend tracking of past status.

## Decisions

### Aggregation lives in `internal/api`, not a new package
The aggregator is a thin composition of already-owned building blocks (one DB ping, up to three short HTTP calls, a handful of `statfs` calls) with no independent business logic of its own. A new handler file (e.g. `internal/api/system_status.go`) matches how `radarr_sonarr_handlers.go` already composes per-request clients. A separate `internal/systemstatus` package was considered and rejected as unnecessary layering for logic this thin and this tied to the HTTP layer.

### Reachability calls
- Radarr/Sonarr: call `GET /api/v3/system/status` (new client method) - the same lightweight, authenticated call Radarr/Sonarr's own UI uses to confirm connectivity. Rejected `/api/v3/health`: richer, but returns the services' *own* internal diagnostics, which is explicitly out of scope (Non-Goals).
- TMDB: call `GET /3/configuration` (new client method) - validates the API key without consuming search quota or requiring extra scopes.
- Each reachability call runs under its own short `context.WithTimeout` (starting at 5s), independent of the client's default business-call timeout (30s) - a diagnostic check must fail fast, not wait as long as a real data fetch would.

### Concurrent, independent checks
The DB check, the three integration checks, and the disk usage checks all run concurrently (goroutines fanning in to the response), so total handler latency is bounded by the slowest single check (~5s), not their sum. Each check's failure is captured independently and never short-circuits the others, satisfying "Each upstream service's availability is reported independently" from the spec.

### KO reasons are stable codes, not raw error text
Reachability failures are classified into a small fixed set of reason codes (e.g. `unreachable`, `unauthorized`, `timeout`, `unavailable`) rather than passed through as raw error strings. The frontend translates these the same way it already translates `ApiError.code` via `useApiErrorMessage`, keeping one error-presentation convention across the app instead of adding a second free-text one for this feature alone.

### Disk deduplication by device id, not by path string
For each configured path, resolve it to the nearest existing ancestor directory (same walk-up already done in `GetDiskSpace`), then read that directory's device id via `os.Stat(...).Sys().(*syscall.Stat_t).Dev`. Paths sharing a device id are grouped into a single reported entry. `st_dev` from `stat(2)` was chosen over `Statfs`'s `f_fsid` because `f_fsid` is unreliable/zero on some filesystem types (notably some overlay/tmpfs configurations common in Docker), while `st_dev` is reliably populated for any path.

## Risks / Trade-offs

- [Risk] A dependency that accepts a TCP connection but never completes its HTTP response could still consume the full per-check timeout → [Mitigation] bounded context timeout per check (5s), checks run in parallel so the endpoint's total wait is capped at that same bound, not the sum of all checks.
- [Risk] Device-id-based dedup could, in unusual bind-mount/overlay setups, either merge two paths that are logically distinct or fail to merge two paths on the same physical disk → [Mitigation] worst case is a harmless extra or duplicate entry, not an incorrect health reading; acceptable given the "minimal diagnostic" scope.
- [Risk] A fixed set of KO reason codes may be too coarse to explain unusual failures → [Mitigation] acceptable per the agreed minimal scope; logs remain the fallback for exotic cases.

## Migration Plan

Purely additive: one new endpoint, one new frontend entry point and dialog. No schema or data migration. Rollback is a plain revert of the deploy.

## Open Questions

- Exact per-check timeout value (proposed: 5s) is tunable during implementation without affecting specs or approach.
