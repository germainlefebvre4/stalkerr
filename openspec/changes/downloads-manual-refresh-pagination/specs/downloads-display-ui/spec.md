## MODIFIED Requirements

### Requirement: Download List Summary Row
The Downloads tab SHALL fetch enriched downloads from `GET /api/v1/downloads` on initial tab activation and when the user clicks the manual refresh control, and SHALL NOT poll or auto-refresh the list in the background. The tab SHALL show a loading state during a fetch, same as before. Instead of a detailed card, each download SHALL render as a compact summary row: on viewports at or above the mobile breakpoint, a table row; below it, a list card matching the Playlist tab's `mobile-list-card` pattern. Each summary row SHALL display, at minimum: the content type icon (🎬/📺/🔗) and title (falling back to the file name or URL when no `content.title` is available) with the year in parentheses when known, the status badge (same statuses/labels as before), and a compact progress indicator (percentage and/or size) for downloads whose status is `downloading` or `retrying`. The summary row SHALL NOT render file paths, technical specification chips, validation badges, genres, or error messages inline — that detail is available in the sidepanel (capability `downloads-details-sidepanel`). Each summary row SHALL be clickable/tappable to open that sidepanel for the corresponding download.

The status, type, and problem filter dropdowns SHALL keep updating query parameters, re-fetching downloads with the new filters, and resetting to page 1 on change, unchanged from before.

#### Scenario: Desktop table row shows only summary content
- **WHEN** the Downloads tab is rendered on a viewport at or above the mobile breakpoint
- **THEN** each download SHALL appear as one table row showing its type icon, title, year, status badge, and (for `downloading`/`retrying` items) a compact progress indicator, with no inline file path, technical specs, validation badges, genres, or error message

#### Scenario: Mobile card shows only summary content
- **WHEN** the Downloads tab is rendered on a viewport narrower than the mobile breakpoint
- **THEN** each download SHALL appear as one list card showing its type icon, title, year, and status badge, matching the visual pattern used by the Playlist tab's mobile list cards

#### Scenario: Missing content title falls back gracefully
- **WHEN** a download has no `content.title`
- **THEN** the summary row SHALL display the download's file name (derived from `download_path`) or, failing that, its `url`, instead of leaving the title blank

#### Scenario: Clicking a row opens the sidepanel
- **WHEN** the user clicks (or taps) a download's summary row
- **THEN** the frontend SHALL open the details sidepanel for that download (capability `downloads-details-sidepanel`)

#### Scenario: Manual refresh re-fetches the current page
- **WHEN** the user clicks the refresh control
- **THEN** the frontend SHALL re-fetch the current page of downloads with the active filters and `limit`/`offset`, without relying on any background timer

#### Scenario: No background auto-refresh
- **WHEN** the Downloads tab is active and the user takes no action
- **THEN** the frontend SHALL NOT issue any additional fetch to `/api/v1/downloads` after the initial load, until the user clicks refresh, changes a filter, changes the page, or changes the items-per-page value

## ADDED Requirements

### Requirement: Pagination Controls
The Downloads tab SHALL paginate the download list using the `limit`, `offset`, `total`, and `total_pages` fields returned by `GET /api/v1/downloads`, and SHALL render page navigation controls plus an items-per-page selector below the list.

#### Scenario: Navigating to another page
- **WHEN** the user navigates to a different page via the pagination controls
- **THEN** the frontend SHALL re-fetch downloads with the `offset` corresponding to that page for the current `limit`, using the active filters, and update the displayed list

#### Scenario: Changing items per page
- **WHEN** the user selects a different items-per-page value
- **THEN** the frontend SHALL re-fetch downloads with the new `limit`, resetting to page 1, using the active filters
