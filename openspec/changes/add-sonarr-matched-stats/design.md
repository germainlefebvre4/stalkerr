## Context

See proposal.md - Why. Two relevant pieces of existing code this design builds on:
- `sonarrMatchCache` (`internal/api/sonarr_match_cache.go`): an in-memory, per-process `map[int]bool` of Sonarr series ID -> "has at least one matched monitored episode", populated lazily via `matchedStatus(ctx, client, db, series)`, cleared only by the Séries section's manual refresh (`refresh=true` on `GET /api/v1/sonarr/series`).
- `listSonarrMonitoredSeries` (`internal/api/radarr_sonarr_handlers.go`): when a match-status `filter` is requested, it already resolves every monitored series' cached status concurrently (one goroutine per series, `sync.WaitGroup`), and fails the whole request with `sonarr_unreachable` if any single series' episode fetch errors.

`listRadarrSonarrStats` is the handler this change extends; it currently computes `radarr_matched` from a single local-DB batch query and `sonarr_monitored` from the monitored-series listing alone, with no Sonarr fan-out.

## Goals / Non-Goals

**Goals:**
- Add a `sonarr_matched` count to `GET /api/v1/radarr-sonarr/stats`, reusing `sonarrMatchCache` so a warm cache costs zero extra Sonarr calls.
- Render the same matched-vs-monitored progress bar for Sonarr that Radarr already has, on both the Home tab and the arr-suite Résumé sub-tab.

**Non-Goals:**
- Bounding or throttling the concurrency of a cold-cache fan-out - out of scope for this change; the existing filtered-listing endpoint has the same unbounded per-request fan-out today, so this does not introduce a new class of risk.
- Changing the Séries section's per-page listing, its match-status filter, or the cache's invalidation rules - all unchanged.
- Making `sonarr_matched` independently nullable while `sonarr_monitored` succeeds - the two are computed together and share the same error path (see Decisions).

## Decisions

**Compute `sonarr_matched` via concurrent cache lookups, mirroring the filtered-listing pattern.** After fetching `GetAllMonitoredSeries`, spawn one goroutine per series calling `sonarrMatchCache.matchedStatus(ctx, client, db, series)` (same call already used by the filter), gated by a `sync.WaitGroup`, then count `matched == true`. Alternative considered: a sequential loop - rejected because on a cold cache it would serialize N Sonarr round-trips instead of bounding latency to roughly the slowest single call, and it would diverge from the concurrency pattern already established for the same underlying operation.

**Treat the Sonarr section as all-or-nothing on any per-series fetch failure.** If any goroutine's `matchedStatus` call errors, the whole Sonarr section of the response reports `sonarr_error: "sonarr_unreachable"` (both `sonarr_monitored` and `sonarr_matched` become `nil`), instead of returning a partial matched count. Alternative considered: report a matched count computed only from the series that succeeded - rejected because a silently-partial ratio would misrepresent the real match rate and there is no way for the frontend to distinguish "partial" from "complete" without a new field; this also matches the existing all-or-nothing behavior `listSonarrMonitoredSeries` already applies to its filtered path.

**Frontend mirrors the existing Radarr card exactly.** `RadarrSonarrStats.sonarr_matched: number | null` is added next to `sonarr_monitored`; both `HomeTab.tsx` and `RadarrSonarrTab.tsx` compute a `sonarrMatchedRatio` with the same `Math.round((matched / monitored) * 100)` guarded-division logic already used for `radarrMatchedRatio`, and render a second `<Progress.Root>`/`<Progress.Indicator>` pair. The arr-suite Résumé Sonarr card additionally gains the matched/unmatched secondary-grid row Radarr's card already has (reusing the existing `resume.matched`/`resume.unmatched` i18n keys), since today it only renders a bare hero metric.

## Risks / Trade-offs

[Risk] A cold cache (first stats request after a backend restart, or right after the Séries section's manual refresh clears it) fans out one concurrent Sonarr episode-list request per currently-monitored series, which could be a large burst against a single Sonarr instance → Mitigation: this is the same cost the match-status filter already accepts today; the user explicitly chose this trade-off (reusing the cache) over a "best-effort, cache-only" alternative during exploration. If Sonarr-side rate limiting becomes an observed problem, bounding concurrency is a follow-up, not a blocker here.

[Risk] The Home tab's Radarr/Sonarr summary card calls the same `/stats` endpoint, so a cold cache now makes the Home tab's first load pay the Sonarr fan-out cost too, even for users who never open the Sonarr or Résumé sub-tabs → Mitigation: accepted as part of the chosen approach; the existing `existenceCheckTimeout` already bounds worst-case request latency the same way it does for Radarr today.

[Risk] If the backend is rolled back after a frontend deploy (or vice versa), the frontend may receive a response without `sonarr_matched` → Mitigation: treat a missing/`undefined` `sonarr_matched` the same as `null` (ratio computed as 0, progress bar not rendered), matching how the existing Radarr fields already degrade when absent.

## Migration Plan

No database changes. Deploy backend and frontend together; the new field is additive so either can briefly lag the other without breaking the existing Radarr display. Rollback is a plain revert of both commits - `sonarr_matched` disappearing from the response degrades to today's "monitored count only" display, not an error.

## Open Questions

- Should the cold-cache fan-out be concurrency-bounded (e.g. a semaphore) to protect a Sonarr instance with a very large monitored catalog? Deferred: the existing filtered-listing endpoint carries the same unbounded pattern today, so this change doesn't regress anything; revisit only if it proves to be a real-world problem.
