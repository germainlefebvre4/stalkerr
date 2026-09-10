## Why

The Home dashboard and the arr-suite "Résumé" sub-tab both render a matched-vs-monitored progress bar for Radarr, but Sonarr's card only shows a raw monitored count with no matched breakdown or progress bar. This asymmetry exists because computing "matched" for Sonarr requires a per-series episode fetch from Sonarr, unlike Radarr's single local-DB batch query - but that per-series computation is already built and cached (`sonarrMatchStatusCache`) to power the existing Séries match-status filter, so the summary cards can reuse it instead of adding a new unbounded cost.

## What Changes

- `GET /api/v1/radarr-sonarr/stats` gains a `sonarr_matched` field: the count of monitored series with at least one matched monitored episode, computed by consulting (and populating on miss) the existing `sonarrMatchStatusCache` for every currently monitored series.
- The Home tab's Sonarr subsection adds a matched count and a progress bar, mirroring the existing Radarr subsection's presentation.
- The arr-suite "Résumé" sub-tab's Sonarr card adds a matched/unmatched breakdown and a progress bar, mirroring the existing Radarr card's presentation.
- The `radarr-sonarr-monitoring-api` spec's fan-out-avoidance requirement is narrowed: the *monitored total* still must never trigger per-series Sonarr calls, but the *matched count* is now explicitly allowed to consult and populate the match-status cache, accepting a one-time full fan-out on a cold cache (e.g. right after backend startup or a manual Séries refresh).

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `radarr-sonarr-monitoring-api`: the stats endpoint response gains `sonarr_matched`, computed via the Sonarr match-status cache; the existing no-fan-out requirement is scoped to the monitored-total computation only.
- `radarr-sonarr-monitoring-view`: the Résumé sub-tab's Sonarr card gains a matched/unmatched breakdown and progress bar, matching the Radarr card's presentation.
- `frontend-home-dashboard`: the Home tab's Radarr/Sonarr summary card gains a Sonarr matched count and progress bar, matching the existing Radarr subsection.

## Impact

- Backend: `internal/api/radarr_sonarr_handlers.go` (`RadarrSonarrStatsResponse`, `listRadarrSonarrStats`), reusing `internal/api/sonarr_match_cache.go` (`sonarrMatchCache.matchedStatus`).
- Frontend: `frontend/src/types.ts` (`RadarrSonarrStats`), `frontend/src/components/HomeTab.tsx`, `frontend/src/components/RadarrSonarrTab.tsx`.
- No database schema changes, no new endpoints, no breaking changes to the existing `sonarr_monitored` field or the Séries match-status filter behavior.
