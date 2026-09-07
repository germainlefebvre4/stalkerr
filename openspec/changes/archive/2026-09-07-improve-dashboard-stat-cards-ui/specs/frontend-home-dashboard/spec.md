## MODIFIED Requirements

### Requirement: Last Processing Run Summary
The Home tab SHALL display a summary of the most recent processing run, fetched from `GET /api/v1/processing-logs`: its execution date/time (`started_at`), elapsed duration (computed from `started_at` and `completed_at`, or presented as "in progress" when `completed_at` is absent), its status, its movies count, TV shows count, new-items count, TMDB matched count, TMDB unmatched count, and its list of `group_title` values, all as defined by capability `processing-run-statistics`. When no processing run has ever been recorded, the Home tab SHALL render an empty state instead of blank or zeroed fields. The card SHALL present the run's status as a status badge (success, failure, or in-progress style, reusing the app's existing badge styles) at the top of the card, followed by the execution date/time and duration, with the remaining counts (movies, TV shows, new items, TMDB matched, TMDB unmatched) arranged in a compact secondary grid below. When the `group_title` list contains more entries than fit on a single line, the Home tab SHALL collapse it behind a disclosure control showing a count of the remaining hidden entries (e.g. "+3 autres"), expandable to reveal the full list.

#### Scenario: Display a completed run's summary
- **WHEN** the most recent processing run completed successfully with a recorded `started_at`, `completed_at`, movies count, TV shows count, new-items count, TMDB matched/unmatched counts, and group title list
- **THEN** the Home tab SHALL render a success-styled status badge, the execution date/time, the elapsed duration between `started_at` and `completed_at`, and each of those counts and the group title list, arranged per the card's status-badge/secondary-grid layout

#### Scenario: Display an in-progress run's summary
- **WHEN** the most recent processing run's status is `in_progress` (no `completed_at` yet)
- **THEN** the Home tab SHALL render an in-progress-styled status badge and indicate the run is still in progress instead of showing a fixed elapsed duration

#### Scenario: Display a failed run's summary
- **WHEN** the most recent processing run's status is `failed`
- **THEN** the Home tab SHALL render a failure-styled status badge alongside the execution date/time, duration, and counts

#### Scenario: Display statistics absent for a pre-migration run
- **WHEN** the most recent processing run predates capability `processing-run-statistics` and its statistics fields are absent
- **THEN** the Home tab SHALL render a placeholder for the unavailable statistics rather than showing `0` or an empty list as if the run legitimately processed nothing

#### Scenario: No processing run has ever been recorded
- **WHEN** `GET /api/v1/processing-logs` returns no entries
- **THEN** the Home tab SHALL render an empty state for the last-run summary section instead of blank or zeroed values

#### Scenario: Long group-titles list is collapsed behind a disclosure
- **WHEN** the most recent run's `group_title` list contains more entries than fit on a single line
- **THEN** the Home tab SHALL render the visible subset followed by a disclosure control labeled with the count of hidden entries, and SHALL NOT display the full comma-joined list until the disclosure is expanded

#### Scenario: Expanding the group-titles disclosure reveals the full list
- **WHEN** the user activates the collapsed group-titles disclosure
- **THEN** the Home tab SHALL expand it to show every group title, and the disclosure control SHALL meet the `44px` minimum touch target on mobile viewports

### Requirement: Catalog Overview Summary
The Home tab SHALL display the total number of playlist items, the number of movies identified, the number of TV shows identified, and the download success percentage (downloaded count vs failed count), fetched from `GET /api/v1/stats`, matching the metrics previously shown by the removed KPI cards banner (capability `frontend-ihm-dashboard`). The card SHALL present the total items count as its hero metric, with movies count, TV shows count shown as secondary detail, and the download success percentage rendered as a progress bar (reusing the app's existing progress-bar styling) alongside its numeric value.

#### Scenario: Catalog overview renders on Home tab load
- **WHEN** the Home tab loads or is refreshed
- **THEN** the frontend SHALL fetch statistics from `/api/v1/stats` and render the total items as the card's hero metric, the movies and TV shows counts as secondary detail, and the download success percentage as a progress bar with its numeric value

### Requirement: Radarr/Sonarr Summary
The Home tab SHALL display a summary of Radarr and Sonarr monitoring, fetched from `GET /api/v1/radarr-sonarr/stats`: the Radarr monitored count, the Radarr matched count, and the Sonarr monitored count. When either service reports an error (unreachable or not configured), the Home tab SHALL display that service's error state instead of a missing or zeroed count, without preventing the other service's summary from rendering. Each service's subsection SHALL display that service's existing brand icon alongside a status badge reflecting whether it loaded successfully or reported an error, and the Radarr subsection SHALL render its matched-vs-monitored ratio as a progress bar alongside the numeric counts.

#### Scenario: Both services report counts
- **WHEN** `GET /api/v1/radarr-sonarr/stats` returns non-null `radarr_monitored`, `radarr_matched`, and `sonarr_monitored` values
- **THEN** the Home tab SHALL render the Radarr monitored/matched counts with a matched-ratio progress bar, the Sonarr monitored count, and a success-styled status badge alongside each service's brand icon

#### Scenario: One service is unreachable while the other succeeds
- **WHEN** `GET /api/v1/radarr-sonarr/stats` returns a `radarr_error` of `radarr_unreachable` alongside a valid `sonarr_monitored` value
- **THEN** the Home tab SHALL display the Radarr error state with a failure-styled status badge while still rendering the Sonarr monitored count with a success-styled status badge

### Requirement: Downloads and Errors Summary
The Home tab SHALL display the total number of downloads (from the existing Downloads listing) and the total number of downloads flagged with a naming problem (`missing_year`, `year_mismatch`, or `unknown_format`, matching the "Erreurs" tab's combined filter defined by capability `downloads-errors-view`). The card SHALL present the total downloads count as its hero metric, with the errors count shown alongside a status badge (neutral when zero, failure-styled when greater than zero).

#### Scenario: Downloads and errors counts render on Home tab load
- **WHEN** the Home tab loads or is refreshed
- **THEN** the frontend SHALL fetch the total downloads count and the combined-reason errors count, and render the downloads count as the card's hero metric alongside the errors count and its status badge

#### Scenario: Errors count is visually flagged when non-zero
- **WHEN** the combined-reason errors count is greater than zero
- **THEN** the frontend SHALL render the errors count with a failure-styled status badge instead of the neutral style used when there are no errors

### Requirement: Home Tab Responsive Layout
Below the mobile breakpoint defined by capability `frontend-responsive-layout`, the Home tab's summary sections SHALL render as vertically stacked cards, reducing padding and font sizes and ensuring interactive elements meet the `44px` minimum touch target, consistent with the "Responsive Card Density" rules applied elsewhere. At or above the mobile breakpoint, the Home tab's sections SHALL render using the desktop card styling, unchanged in layout density. Below the mobile breakpoint, each card's status badge SHALL be replaced by the app's existing compact status-dot indicator to conserve horizontal space, and each card's hero metric SHALL use its own responsive type scale distinct from the secondary-detail text size.

#### Scenario: Home tab sections stack vertically on mobile
- **WHEN** the viewport is narrower than the mobile breakpoint and the Home tab is active
- **THEN** the frontend SHALL render the last-run summary, catalog overview, Radarr/Sonarr summary, and downloads/errors summary as vertically stacked cards, with no horizontal scrolling required to read any card's content

#### Scenario: Home tab sections render unaffected on desktop
- **WHEN** the viewport is at or above the mobile breakpoint and the Home tab is active
- **THEN** the frontend SHALL render the Home tab's sections using the existing desktop card density

#### Scenario: Status badges become compact dots on mobile
- **WHEN** the viewport is narrower than the mobile breakpoint
- **THEN** each card SHALL render its status using the compact status-dot indicator instead of the full status badge used on desktop

#### Scenario: Hero metric remains legible on mobile
- **WHEN** the viewport is narrower than the mobile breakpoint
- **THEN** each card's hero metric SHALL render at a larger size than the card's secondary detail text, distinct from the uniform secondary-detail font size
