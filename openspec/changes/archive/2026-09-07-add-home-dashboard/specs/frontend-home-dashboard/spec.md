## Purpose

A single Home/Dashboard landing tab, on both desktop and mobile, that aggregates the outcome of the last M3U processing run together with catalog, Radarr/Sonarr, downloads, and error counts, replacing the previous always-visible KPI banner.

## ADDED Requirements

### Requirement: Home Tab as Default Landing Page
The frontend SHALL add a "Home" tab, positioned first among the section tabs on both the desktop segmented tab list and the mobile bottom tab bar. The Home tab SHALL follow the same active-tab persistence rules as the other tabs (capability `frontend-ihm-dashboard`'s "Real-time Monitoring Dashboard" requirement): reflected in the URL's `tab` query parameter, persisted to `localStorage`, and restored on refresh. When no `tab` URL parameter and no stored `localStorage` value are present, the frontend SHALL activate the Home tab by default.

#### Scenario: Home is the default tab on first load
- **WHEN** the user opens the application for the first time, with no `tab` query parameter and no `stalkeer_active_tab` value in `localStorage`
- **THEN** the frontend SHALL activate the Home tab

#### Scenario: Home appears first in both desktop and mobile navigation
- **WHEN** the user views the tab navigation, on desktop or on a mobile viewport
- **THEN** the frontend SHALL render "Home" as the first entry, before Playlist, Logs, Downloads, and the other tabs

#### Scenario: A restored tab choice still takes priority over the Home default
- **WHEN** the user previously selected the "Downloads" tab and refreshes the browser
- **THEN** the frontend SHALL restore the "Downloads" tab as active, not fall back to Home

### Requirement: Last Processing Run Summary
The Home tab SHALL display a summary of the most recent processing run, fetched from `GET /api/v1/processing-logs`: its execution date/time (`started_at`), elapsed duration (computed from `started_at` and `completed_at`, or presented as "in progress" when `completed_at` is absent), its status, its movies count, TV shows count, new-items count, TMDB matched count, TMDB unmatched count, and its list of `group_title` values, all as defined by capability `processing-run-statistics`. When no processing run has ever been recorded, the Home tab SHALL render an empty state instead of blank or zeroed fields.

#### Scenario: Display a completed run's summary
- **WHEN** the most recent processing run completed successfully with a recorded `started_at`, `completed_at`, movies count, TV shows count, new-items count, TMDB matched/unmatched counts, and group title list
- **THEN** the Home tab SHALL render the execution date/time, the elapsed duration between `started_at` and `completed_at`, and each of those counts and the group title list

#### Scenario: Display an in-progress run's summary
- **WHEN** the most recent processing run's status is `in_progress` (no `completed_at` yet)
- **THEN** the Home tab SHALL indicate the run is still in progress instead of showing a fixed elapsed duration

#### Scenario: Display statistics absent for a pre-migration run
- **WHEN** the most recent processing run predates capability `processing-run-statistics` and its statistics fields are absent
- **THEN** the Home tab SHALL render a placeholder for the unavailable statistics rather than showing `0` or an empty list as if the run legitimately processed nothing

#### Scenario: No processing run has ever been recorded
- **WHEN** `GET /api/v1/processing-logs` returns no entries
- **THEN** the Home tab SHALL render an empty state for the last-run summary section instead of blank or zeroed values

### Requirement: Catalog Overview Summary
The Home tab SHALL display the total number of playlist items, the number of movies identified, the number of TV shows identified, and the download success percentage (downloaded count vs failed count), fetched from `GET /api/v1/stats`, matching the metrics previously shown by the removed KPI cards banner (capability `frontend-ihm-dashboard`).

#### Scenario: Catalog overview renders on Home tab load
- **WHEN** the Home tab loads or is refreshed
- **THEN** the frontend SHALL fetch statistics from `/api/v1/stats` and render the total items, movies, TV shows, and download success percentage within the Home tab

### Requirement: Radarr/Sonarr Summary
The Home tab SHALL display a summary of Radarr and Sonarr monitoring, fetched from `GET /api/v1/radarr-sonarr/stats`: the Radarr monitored count, the Radarr matched count, and the Sonarr monitored count. When either service reports an error (unreachable or not configured), the Home tab SHALL display that service's error state instead of a missing or zeroed count, without preventing the other service's summary from rendering.

#### Scenario: Both services report counts
- **WHEN** `GET /api/v1/radarr-sonarr/stats` returns non-null `radarr_monitored`, `radarr_matched`, and `sonarr_monitored` values
- **THEN** the Home tab SHALL render the Radarr monitored/matched counts and the Sonarr monitored count

#### Scenario: One service is unreachable while the other succeeds
- **WHEN** `GET /api/v1/radarr-sonarr/stats` returns a `radarr_error` of `radarr_unreachable` alongside a valid `sonarr_monitored` value
- **THEN** the Home tab SHALL display the Radarr error state while still rendering the Sonarr monitored count

### Requirement: Downloads and Errors Summary
The Home tab SHALL display the total number of downloads (from the existing Downloads listing) and the total number of downloads flagged with a naming problem (`missing_year`, `year_mismatch`, or `unknown_format`, matching the "Erreurs" tab's combined filter defined by capability `downloads-errors-view`).

#### Scenario: Downloads and errors counts render on Home tab load
- **WHEN** the Home tab loads or is refreshed
- **THEN** the frontend SHALL fetch the total downloads count and the combined-reason errors count, and render both within the Home tab

### Requirement: Home Tab Responsive Layout
Below the mobile breakpoint defined by capability `frontend-responsive-layout`, the Home tab's summary sections SHALL render as vertically stacked cards, reducing padding and font sizes and ensuring interactive elements meet the `44px` minimum touch target, consistent with the "Responsive Card Density" rules applied elsewhere. At or above the mobile breakpoint, the Home tab's sections SHALL render using the desktop card styling, unchanged in layout density.

#### Scenario: Home tab sections stack vertically on mobile
- **WHEN** the viewport is narrower than the mobile breakpoint and the Home tab is active
- **THEN** the frontend SHALL render the last-run summary, catalog overview, Radarr/Sonarr summary, and downloads/errors summary as vertically stacked cards, with no horizontal scrolling required to read any card's content

#### Scenario: Home tab sections render unaffected on desktop
- **WHEN** the viewport is at or above the mobile breakpoint and the Home tab is active
- **THEN** the frontend SHALL render the Home tab's sections using the existing desktop card density
