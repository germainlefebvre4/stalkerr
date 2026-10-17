# Design

## Context

See proposal.md - Why. Three limits stack on every monitoring request, and the tightest one wins:

```
 request budget (context.WithTimeout)   10s   existenceCheckTimeout
 http.Client.Timeout (per upstream call) 10s   existenceCheckTimeout
 http.Server.WriteTimeout                15s   cuts the response write
```

`existenceCheckTimeout` lives in `force_download.go` and is shared by force-download, resync, duplicate handling and the monitoring handlers. Retries are off (`MaxAttempts: 1`) for the monitoring handlers, and the Radarr/Sonarr clients go through one process-wide circuit breaker per service (5 failures, 60 s open).

Production logs show both `GET /radarr/movies` and `GET /sonarr/series` ending at exactly 10 s with `502`. The Sonarr body message (`failed to reach Sonarr`) is the one returned when the first catalog call fails, so the per-series episode fan-out is not involved.

## Goals / Non-Goals

**Goals:**
- Let the monitoring view work against slow, but responsive, Radarr/Sonarr instances.
- Make the cause of an upstream fetch failure visible in the logs.

**Non-Goals:**
- Fixing the slow instances (their SQLite config volume moved from iSCSI to NFS; an infrastructure change).
- Reducing upstream load: the Sonarr fan-out bounds, caching or `singleflight` of the catalog reads, and breaker tuning are left for a separate change.
- Making the budgets configurable.

## Decisions

**Dedicated constants instead of raising `existenceCheckTimeout`.** The monitoring handlers get `arrListTimeout` (30 s) and `arrStatsTimeout` (60 s), defined next to the handlers. Raising the shared constant would also slow down the fail-fast paths (force-download validation, resync, duplicates) that are meant to answer quickly. Alternative rejected: one constant for everything, simpler but changes unrelated endpoints.

**Hard-coded, not a setting.** The change is a stopgap until the infrastructure is fixed; a registry entry, UI field and tests would outlive its purpose. It can become a setting later if the need persists.

**Stats gets a wider overall budget.** The stats handler calls Radarr then Sonarr under one context, with the Sonarr phase possibly fanning out. Each upstream call keeps the 30 s client timeout, while the shared context is 60 s so a slow Radarr does not leave Sonarr with no time.

**`WriteTimeout` raised to 65 s.** It must exceed the longest handler budget (`arrStatsTimeout`) with margin; otherwise the server aborts the write while the handler still runs and the client sees a dropped connection instead of an answer. It is server-wide, but it only loosens a ceiling: handlers that finish fast are unaffected. Alternative rejected: a per-route write deadline via `http.ResponseController`, more precise but more machinery for a temporary change.

**Log the underlying error in the two list handlers.** Both handlers currently discard `err` when the catalog fetch fails, so a timeout, a refused connection and an open breaker all look identical. A `Warn` entry with the error through the application logger (same pattern as the other API handlers) gives that distinction at no behavior cost. Stats, per-item detail and per-series fan-out failures are left as they are to keep the change small.

## Risks / Trade-offs

- [The page can take up to 30 s to load while an arr instance is degraded] → Accepted for now; the root cause is the infrastructure. The frontend already shows a loading state per section.
- [A slow upstream still trips the circuit breaker] → Timeouts count as failures as before. A longer budget makes this slower to happen, not impossible; breaker tuning is a non-goal here.
- [A longer `WriteTimeout` keeps slow client connections open longer] → Bounded at 65 s and limited to a small, self-hosted deployment.
- [Slow responses make the three parallel requests on tab activation compete for the same instances] → Unchanged from today; revisit with the load-reduction follow-up if it remains a problem.
