## Context

See `proposal.md` - Why. Both listing endpoints (`internal/api/radarr_sonarr_handlers.go`) currently paginate before computing match status, so per-request cost scales with page size, not catalog size (see `radarr-sonarr-monitoring-api` spec). This works because nothing today needs match status for entries outside the current page. A match-status filter breaks that assumption: filtering correctly requires knowing every entry's match status before deciding which page it falls on.

Radarr matching (`matcher.MatchMoviesBatch`) is local-DB-only and already runs over the full monitored catalog elsewhere (the `/stats` endpoint), so extending it to the filtered-list path is cheap. Sonarr matching requires one Sonarr episode-list API call per series (`client.GetEpisodesBySeriesID`), which the code deliberately avoids running for the whole catalog (see the existing "Sonarr monitored series total count without per-series fan-out" requirement, and the `/stats` endpoint that skips it). A naive full-catalog fan-out on every filtered request would reintroduce that cost.

## Goals / Non-Goals

**Goals:**
- Filter both tables by match status before pagination, so filtered totals/pages are correct.
- Avoid a full Sonarr episode fan-out on every filtered request.
- Add an occurrence count column without changing the existing page-scoped cost model for match computation.
- Keep the existing "no polling" characteristic of the Radarr/Sonarr tab intact.

**Non-Goals:**
- Persisting Sonarr match status to the database, or introducing any scheduled/background job.
- Changing the existing partial matched/monitored ratio badge already shown for series - the new filter is a separate, binary view of the same underlying data.
- Real-time cache invalidation when the playlist or Sonarr's monitored set changes outside of a user-triggered refresh.

## Decisions

### Radarr: compute full-catalog match status inline, only when filtering
When `filter` is present, `listRadarrMonitoredMovies` runs `matcher.MatchMoviesBatch` over the entire monitored (post-search) list instead of just the requested page, filters by the requested status, then paginates. When `filter` is absent, the existing page-scoped behavior is unchanged. `MatchMoviesBatch` already batches DB lookups regardless of input size (at most two queries), so this doesn't introduce N+1 queries - the only change is that the batch spans the full catalog instead of one page, matching what `/stats` already does today.

Alternative considered: always compute full-catalog match status, drop the page-scoped path entirely. Rejected - it would silently make every unfiltered request pay the "full catalog" cost, which is unnecessary for the common case and changes an existing, explicitly-specified performance characteristic for no benefit.

### Sonarr: in-memory lazy cache of matched/unmatched status, invalidated by manual refresh
A new package-level (or server-scoped) in-memory cache stores `series_id -> matched (bool)`, keyed off the same Sonarr instance. When a filtered Sonarr list request needs a series' status and it isn't cached, the handler fetches that series' episodes and computes `matched = (episodes matched > 0)`, then stores it. Subsequent filtered requests reuse the cached value until the user triggers the Séries section's manual refresh, which clears the cache (or at least the entries touched by that refresh).

This keeps the "no background polling" property (see `radarr-sonarr-monitoring-view` - "Manual refresh per section"): the cache only ever changes as a side effect of a request the user already caused (a filtered list load, or an explicit refresh), never a timer.

Alternatives considered:
- **Cache in the database** - persists across restarts, but adds schema/migration surface for a value that's cheap to recompute and already explicitly out of scope (proposal's Impact: "No database schema changes"). Rejected for this change's scope.
- **Scheduled background refresh (cron)** - keeps the cache warm without a user action, but is a new architectural pattern (this codebase has no polling/cron for this feature area) and was explicitly rejected during exploration in favor of staying consistent with the existing explicit-trigger model.
- **No cache, accept full fan-out per filtered request** - simplest, but reintroduces the exact per-catalog Sonarr API fan-out the current design deliberately avoids; risks timeouts/rate-limiting on large libraries.

### Occurrence count stays page-scoped, computed alongside existing match data
For Radarr, add one grouped COUNT query (`processed_lines` grouped by `movie_id`) for the page's matched movie IDs, run alongside the existing `MatchMoviesBatch` call. For Sonarr, switch the per-page per-series computation from `matcher.MatchSeriesEpisodesAggregate` (which only counts matched episodes) to `matcher.MatchSeriesEpisodesDetail` (which already loads full occurrence lists per episode), and sum `len(Occurrences)` across a series' episodes. This adds one occurrence query per page for Radarr and switches Sonarr's existing per-series query to a slightly heavier variant that was already built for the sidepanel detail view - no new fan-out shape, same page-scoped cost profile as today.

### Filter and occurrence values are independent of the Sonarr cache for Radarr
The Sonarr cache exists solely to avoid the full-catalog episode fan-out for the *filter*. The occurrence count column is unrelated to the cache (it's page-scoped for both services) and does not read from or populate it.

## Risks / Trade-offs

- **[Risk]** First filtered Sonarr request after a cold start (or after refresh) still pays the full fan-out cost, potentially slow for a large monitored catalog. → **Mitigation**: this only happens once per cache lifetime (or once per refresh), not per filtered page turn; matches the cost the "Matched" Résumé stat would already incur if it did full-catalog Sonarr matching (which it explicitly avoids) - document the latency trade-off in the tasks/testing so it's a known, accepted cost rather than a regression.
- **[Risk]** Cache can go stale if Sonarr's monitored episodes or the playlist change without the user hitting refresh, so a filtered view could miss a series that became matched. → **Mitigation**: this mirrors the existing "manual refresh only, no polling" model already specified for the rest of the tab (see "Manual refresh per section"); the cache does not make staleness worse than the existing unfiltered view already tolerates between refreshes, since even an unfiltered page load only reflects data as of its own request.
- **[Risk]** In-memory cache means matched/unmatched status can differ between backend replicas if the deployment ever runs more than one instance. → **Mitigation**: out of scope for this change (the app's Helm chart and current deployment model targets a single instance for this stateful matching concern); note as a constraint rather than solve now.

## Open Questions

- Exact cache eviction granularity on manual refresh (clear the whole Sonarr cache vs. only entries touched by the refreshed page) is an implementation detail that doesn't affect the spec or task breakdown - can be decided during implementation.
