## Why

The dashboard's current "bandeau statistiques" (4 KPI cards fixed above the tabs) only shows catalog totals and gives no visibility into what the last M3U processing run actually did (when it ran, how long it took, how many movies/tvshows it picked up, how many titles matched TMDB vs didn't, which `group_title`s were pulled in). On mobile, the "Filtres" tab also takes a permanent slot in the bottom tab bar even though sorting-filter configuration is a low-frequency, desktop-oriented admin task. Users need a single landing page that surfaces the last run's outcome plus the catalog, Radarr/Sonarr, downloads, and error counts already computed elsewhere in the app, and the mobile nav should stop dedicating scarce bottom-bar space to the Filtres tab.

## What Changes

- Add a new "Home" / "Dashboard" tab, shown first in the tab order on both desktop and mobile, that becomes the app's default landing tab.
- **BREAKING**: Remove the always-visible KPI cards banner (`StatsKPICards`, the `kpi-grid`/`kpi-toggle-row` block rendered above the tabs) from every tab; its 4 metrics (and more) now live only inside the Home tab.
- The Home tab aggregates:
  - Last processing run: execution date/time, elapsed duration, status, movies/tvshows picked up, new titles discovered, TMDB matched/unmatched counts, and the list of `group_title`s retrieved by that run.
  - Catalog overview: total playlist items, movies identified, TV shows identified, download success percentage (existing `/api/v1/stats` data).
  - Radarr/Sonarr summary: monitored/matched counts (existing `/api/v1/radarr-sonarr/stats` data).
  - Downloads summary: total downloads count (and by status).
  - Errors summary: count of downloads with naming problems (existing Erreurs-tab query).
- Persist per-run statistics on `processing_logs` so the Home tab's "last run" section doesn't need to be recomputed by scanning items: movies count, TV shows count, new-items count, TMDB matched count, TMDB unmatched count, and the distinct `group_title` list for that run. The processor already computes movies/tvshows/TMDB-matched/TMDB-not-found counts in memory (`internal/processor/processor.go` `Statistics`) and already knows create-vs-update per item (`saveBatch`) — these are newly persisted, not newly computed.
- **BREAKING**: Remove "Filtres" from the mobile bottom tab bar, making it desktop-only (same pattern already used for the "Erreurs" tab): if "Filtres" is active and the viewport narrows below the mobile breakpoint, fall back to another tab.

## Capabilities

### New Capabilities
- `frontend-home-dashboard`: the new Home/Dashboard tab — its layout, the data it aggregates from existing and new endpoints, and its default-landing-tab behavior on desktop and mobile.
- `processing-run-statistics`: computing and persisting per-run statistics (movies/tvshows counts, new-items count, TMDB matched/unmatched counts, distinct `group_title` list) on the `processing_logs` entry created by each M3U processing run.

### Modified Capabilities
- `api-processing-logs`: `GET /api/v1/processing-logs` response entries gain the new per-run statistics fields introduced by `processing-run-statistics`.
- `frontend-ihm-dashboard`: remove the "Statistics KPI Cards" requirement (the banner is superseded by the Home tab); the default active tab on load changes from `playlist` to `home`.
- `frontend-responsive-layout`: "Mobile Bottom Tab Navigation" changes from the fixed (Playlist, Filters, Logs, Downloads[, Radarr-Sonarr]) set to include "Home" and exclude "Filtres" (which becomes desktop-only, mirroring the existing Erreurs-tab pattern); the "Responsive Card Density" KPI-grid collapse/expand requirement is removed since the KPI grid no longer exists at the page level.

## Impact

- **Backend**: `internal/models/log.go` (new `ProcessingLog` columns), `internal/processor/processor.go` (`Statistics` struct gains `NewItems`, `TMDBMatched`/`TMDBNotFound` already present, distinct group-title tracking; `saveBatch` and `updateProcessingLog` persist them), `internal/api/dto.go` / `internal/api/handlers.go` (processing-logs list response includes the new fields), `internal/database/database.go` (`AutoMigrate` picks up the new `ProcessingLog` columns automatically, no manual migration file needed).
- **Frontend**: `frontend/src/App.tsx` (tab list, `VALID_TABS`, default tab, mobile fallback logic, removal of `StatsKPICards` rendering), new `HomeTab` component and supporting hook(s) reusing/extending `useHealthAndStats`, `useLogs`, `useRadarrSonarr`, `useDownloads`/`useErrorsTab` data, `frontend/src/components/StatsKPICards.tsx` (removed or repurposed into the Home tab), i18n locale entries (`frontend/src/locales/*/common.json` and a new namespace for the Home tab), `frontend/src/index.css`/`variables.css` (new dashboard layout styles, removal of the now-unused `kpi-toggle-row`/page-level `kpi-grid` styles).
- **Specs**: `openspec/specs/frontend-ihm-dashboard/spec.md`, `openspec/specs/frontend-responsive-layout/spec.md`, `openspec/specs/api-processing-logs/spec.md` (deltas), plus two new spec directories.
