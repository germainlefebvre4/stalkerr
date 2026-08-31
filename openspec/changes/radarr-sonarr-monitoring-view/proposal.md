## Why

Radarr and Sonarr integration today only exists as one-shot CLI batch commands (`stalkeer radarr`, `stalkeer sonarr`) that fetch missing items and immediately attempt to download them. There is no way to see, from the UI, what Radarr/Sonarr currently monitor and whether the local M3U playlist already has a matching entry for it. Users have to run the CLI blind and read terminal output to find out. A dedicated view removes that guesswork and lets users spot unmatched or partially-matched monitored content before triggering a download.

## What Changes

- Add a new "Radarr/Sonarr" tab in the frontend, independent from the existing Playlist tab, with two sections: Films (Radarr) and Séries (Sonarr).
- Add two backend proxy endpoints that live-query Radarr/Sonarr (no new persisted tables, no scheduled sync job) and enrich the result with the existing playlist-matching logic (`internal/matcher`), reusing the same TVDB → TMDB → fuzzy matching path the CLI download commands already use so the view's match status is always consistent with what a real download attempt would do.
- Both endpoints are paginated, and pagination happens **before** the matching computation: the full lightweight Radarr/Sonarr listing is fetched once per request, then only the requested page's items go through matching (DB lookups for movies, plus per-series episode fetch from Sonarr for the aggregate). This bounds per-request cost to the page size regardless of total catalog size, in particular the Sonarr per-series episode call, which is the most expensive operation.
- Sonarr results are aggregated per series (e.g. "8/12 episodes matched"), not broken out per episode, in the list view.
- Each row opens a sidepanel showing the matched local `Movie`/`TVShow` metadata (when matched) and the playlist occurrences found for it (resolution, state), reusing `FindMovieDownloadCandidates`/`FindTVShowDownloadCandidates`.
- Data is fetched on demand: an explicit "Actualiser" (refresh) action per section triggers the live fetch; there is no background polling or auto-refresh.
- Each section (Films/Séries) has its own independent loading and error state, so a Radarr outage does not prevent the Sonarr section from displaying, and vice versa.

## Capabilities

### New Capabilities
- `radarr-sonarr-monitoring-api`: backend endpoints that live-fetch Radarr's monitored movies and Sonarr's monitored series, paginate before computing playlist-match status, and expose per-item/per-series match results plus per-movie/per-series matched playlist occurrence detail.
- `radarr-sonarr-monitoring-view`: frontend tab listing Radarr-monitored movies and Sonarr-monitored series with their playlist-match status, manual refresh per section, independent per-section error handling, and a sidepanel showing matched playlist occurrences for a selected item.

### Modified Capabilities
(none - existing Radarr/Sonarr client behavior and matcher logic are reused as-is; new client methods are additive)

## Impact

- `internal/external/radarr`, `internal/external/sonarr`: new client methods to list all monitored movies/series (not just "missing"), and to fetch episodes for a specific series on demand.
- `internal/matcher`: new aggregate helper to compute per-series playlist-match counts from a series' monitored episodes.
- `internal/api`: two new handler(s)/routes for paginated Radarr/Sonarr monitoring listings, following the existing proxy-handler pattern (`searchTMDBProxy`) and pagination pattern (`parsePagination`).
- `frontend/src`: new tab component, API client methods, i18n strings (fr/en), reusing the existing sidepanel/drawer visual pattern from `PlaylistTab.tsx`.
- No new database tables or migrations. No new scheduled jobs. No new configuration (existing `cfg.Radarr`/`cfg.Sonarr` URL/API key are reused).
