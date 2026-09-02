## Why

The Radarr and Sonarr sub-tabs let users see whether their monitored movies/series have a matching playlist entry, but they can only browse the full paginated list to find matched or unmatched items - there is no way to isolate just the unmatched ones (the ones that actually need action) or to see, at a glance, how many playlist occurrences back a given match. Both gaps force users to page through the full catalog or open each row's sidepanel one at a time.

## What Changes

- Add a match-status filter (`matched` / `no match`) to both the Films (Radarr) and Séries (Sonarr) tables, applied server-side before pagination so page counts and totals reflect the filtered set.
  - For Sonarr, "matched" means at least one of the series' monitored episodes has a playlist match; "no match" means zero do. This is independent of the existing matched/monitored ratio badge, which continues to show the detailed partial ratio.
- Add an "Occurrences" column to both tables showing the total number of playlist occurrences (`ProcessedLine` rows) found for that row's media - counting every matching playlist entry, including duplicates at different resolutions/qualities, not just distinct matched episodes/movies.
- **Radarr**: compute match status across the entire monitored catalog before pagination (extends the existing DB-only batch match, already done this way for the Résumé stats endpoint) so the filter and pagination stay consistent.
- **Sonarr**: introduce an in-memory, per-process cache of each monitored series' matched/unmatched status, since determining it requires fetching episodes from Sonarr per series (a network fan-out). The cache is populated lazily (a filtered request with no cached entry computes and stores it) and invalidated only by the existing manual refresh action - no background polling or scheduled refresh is introduced.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `radarr-sonarr-monitoring-api`: the Radarr and Sonarr listing endpoints gain an optional match-status filter parameter and an occurrence count per item; the existing "pagination before match computation" and "no caching or persistence" requirements are amended to describe the filtered-request path and the new Sonarr match-status cache.
- `radarr-sonarr-monitoring-view`: the Films and Séries sections gain a match-status filter control and an Occurrences column.

## Impact

- Backend: `internal/api/radarr_sonarr_handlers.go` (`listRadarrMonitoredMovies`, `listSonarrMonitoredSeries`), `internal/matcher/matcher.go` (new/adjusted batch occurrence-count queries), a new in-memory Sonarr match-status cache component.
- Frontend: `frontend/src/components/RadarrSonarrTab.tsx`, `frontend/src/hooks/useRadarrSonarr.ts` (filter state, occurrence count rendering), likely `frontend/src/hooks/useRadarrSonarrView.ts` or a new URL-persisted filter hook following the `useDownloads.ts` pattern.
- No database schema changes; no new persistence beyond the in-memory Sonarr cache.
