## Why

The Radarr/Sonarr monitoring tab currently stacks the Films (Radarr) and Séries (Sonarr) sections vertically in a single view, with no way to jump directly to one, no way to filter either table by title, and no at-a-glance summary of how much of the monitored catalog is matched. As the monitored catalog grows, scrolling past one section to reach the other and hunting for a specific title in a paginated table both become friction.

## What Changes

- Restructure the Radarr/Sonarr monitoring tab into three nested sub-tabs: **Résumé** (summary/stats), **Radarr** (the existing Films section), **Sonarr** (the existing Séries section).
- Add a search field to the Radarr sub-tab that filters the movies table by title, server-side, replacing the current page with matching results (not limited to the already-loaded page).
- Add a search field to the Sonarr sub-tab that filters the series table by title, server-side, same behavior.
- Add a Résumé sub-tab showing:
  - Radarr: total monitored movies, and a matched/unmatched breakdown computed across the full monitored catalog (not just the loaded page).
  - Sonarr: total monitored series only (no matched/unmatched breakdown - computing it would require fetching every series' episode list from Sonarr individually, which the existing per-page endpoint deliberately avoids doing for the full catalog).
- **BREAKING**: `GET /api/v1/radarr/movies` and `GET /api/v1/sonarr/series` gain an optional `search` query parameter; existing callers without it are unaffected.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `radarr-sonarr-monitoring-view`: the single-page Films/Séries layout becomes three sub-tabs (Résumé, Radarr, Sonarr); the Radarr and Sonarr sub-tabs each gain a title search field; a new Résumé sub-tab is added.
- `radarr-sonarr-monitoring-api`: the movie and series listing endpoints gain an optional `search` parameter filtering by title before pagination; a new lightweight stats endpoint (or endpoints) reports full-catalog matched/unmatched movie counts and total monitored series count.

## Impact

- `frontend/src/components/RadarrSonarrTab.tsx`: restructured around a nested `Tabs.Root` (Résumé / Radarr / Sonarr); each of the Radarr and Sonarr panels gains a search input.
- `frontend/src/hooks/useRadarrSonarr.ts`: search term state per section, reset to page 1 on search change, and a fetch for the new stats data.
- `frontend/src/services/api.ts`: `listRadarrMovies`/`listSonarrSeries` accept a `search` parameter; a new client method for the stats endpoint(s).
- `internal/api/radarr_sonarr_handlers.go`: `search` query parameter filtering (title substring, case-insensitive) applied to the in-memory monitored list before the pagination slice, for both movies and series; a new handler computing full-catalog Radarr matched/unmatched counts (via `matcher.MatchMoviesBatch` over the entire monitored list) and the Sonarr monitored-series total.
- `frontend/src/locales/{fr,en}/radarrSonarr.json`: new strings for the sub-tab labels, search placeholders/empty-results text, and the Résumé section.
- No changes to the existing per-item detail endpoints (`.../matches`, `.../episodes`) or to the sidepanel/drawer behavior.
