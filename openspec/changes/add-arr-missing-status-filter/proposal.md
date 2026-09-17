## Why

The arr-suite Radarr/Sonarr tabs only ever show monitored items and only ever badge them by local-playlist match status. There is no way to see, at a glance or via a filter, that a monitored movie/series is missing its file(s) in Radarr/Sonarr itself, nor to see unmonitored items at all — the backend currently drops them before they reach the frontend.

## What Changes

- Expose Radarr's own `monitored`/`hasFile` fields and Sonarr's own `monitored`/episode file-completeness fields through the Radarr movies listing, the Sonarr series listing, and the Sonarr per-series episode detail endpoint, instead of discarding them server-side as today.
- Derive a "Missing" state from that data: a Radarr movie is Missing when `monitored && !hasFile`; a Sonarr series is Missing when `monitored && episodeFileCount < totalEpisodeCount`; a Sonarr episode is Missing when `monitored && !hasFile`. This is independent of local-playlist match status.
- **BREAKING** (internal, single consumer — the arr-suite tab): the Radarr movies and Sonarr series listing endpoints stop unconditionally excluding unmonitored items. A new `status` query parameter (repeatable, values `monitored` / `unmonitored` / `missing`) controls which items are included. Omitting it preserves today's exact behavior (monitored-only, unfiltered by missing/unmonitored).
- Add a new "État" status filter (Monitored / Unmonitored / Missing, multi-select and cumulative) to both the Films and Séries sections of the arr-suite tab, independent of and combinable with the existing match-status (Matched / No match) filter. Multiple selected values combine as a logical AND (an entry must satisfy every selected value), matching the multi-select's own internal logic — contradictory combinations (e.g. Unmonitored + Missing) correctly yield an empty result. Defaults to "Monitored" selected, matching today's behavior. Persisted to the URL like the existing filter.
- Add an "État" badge to both tables (new column) and to the movie/series header in the sidepanel; for Sonarr, also add a per-episode "État" badge in the season/episode breakdown, alongside the existing matched/unmatched badge.
- On mobile-width viewports, the new "État" badge renders as a colored status dot with a short label (reusing the existing summary-card `renderStatusIndicator` convention) instead of a full text badge, so the existing mobile card layout is not crowded by a second full badge.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `radarr-sonarr-monitoring-api`: listing/detail endpoints stop hard-filtering to monitored-only, gain a `status` filter parameter, and expose the raw monitored/file-presence fields needed to derive "Missing".
- `radarr-sonarr-monitoring-view`: Films and Séries sections gain the "État" status filter and badge (table, sidepanel header, per-episode breakdown), with a mobile-specific compact rendering.

## Impact

- Backend: `internal/external/radarr/radarr.go`, `internal/external/sonarr/sonarr.go` (no client changes expected — `GetAllMovies`/`GetAllSeries` already return the needed fields), `internal/api/radarr_sonarr_handlers.go` (remove the hard monitored-only filter, add `status` query parsing, extend response structs).
- Frontend: `frontend/src/types.ts` (extend `RadarrMovieListItem`, `SonarrSeriesListItem`, `SonarrSeriesEpisodeItem`), `frontend/src/services/api.ts` (pass the new `status` param), `frontend/src/hooks/useRadarrSonarr.ts` (new URL-persisted filter state), `frontend/src/components/RadarrSonarrTab.tsx` (new filter control, badge rendering in both tables, sidepanel header, and season/episode breakdown, plus mobile-specific rendering).
- No database schema changes; no new upstream API calls (the fields already come back from the existing `GetAllMovies`/`GetAllSeries`/`GetEpisodesBySeriesID` calls, just previously discarded).
