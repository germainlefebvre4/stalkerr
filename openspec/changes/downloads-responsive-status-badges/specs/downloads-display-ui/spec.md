## MODIFIED Requirements

### Requirement: Download List Summary Row
The Downloads tab SHALL fetch enriched downloads from `GET /api/v1/downloads` on initial tab activation and when the user clicks the manual refresh control, and SHALL NOT poll or auto-refresh the list in the background. The tab SHALL show a loading state during a fetch, same as before. Instead of a detailed card, each download SHALL render as a compact summary row: on viewports at or above the mobile breakpoint, a table row; below it, a list card matching the Playlist tab's `mobile-list-card` pattern. Each summary row SHALL display, at minimum: the content type icon (🎬/📺/🔗) and title (falling back to the file name or URL when no `content.title` is available) with the year in parentheses when known, the status badge, and a compact progress indicator (percentage and/or size) for downloads whose status is `downloading` or `retrying`. The summary row SHALL NOT render file paths, technical specification chips, validation badges, genres, or error messages inline — that detail is available in the sidepanel (capability `downloads-details-sidepanel`). Each summary row SHALL be clickable/tappable to open that sidepanel for the corresponding download.

On mobile (viewport narrower than the mobile breakpoint), the status badge SHALL show its emoji alone, with no accompanying word, for the `completed`, `pending`, `downloading`, and `failed` (no retries) statuses; each of these four statuses SHALL use a distinct emoji so that none of them can be confused for another without relying on text. A `failed` download that has been retried SHALL show its emoji followed by the retry count in parentheses (e.g. `❌ (3×)`), without the word "Échec"/"Failed". On desktop (viewport at or above the mobile breakpoint), the status badge SHALL show that same emoji followed by its status text label (e.g. `✅ Complété`, `❌ Échec`), and a retried `failed` download SHALL show the emoji, the "Échec"/"Failed" text, and the retry count together (e.g. `❌ Échec (3×)`). The `retrying` status badge is unchanged on both viewports (emoji plus its word label). Regardless of status or viewport, the status badge SHALL render on a single line and SHALL NOT wrap its content onto multiple lines at any viewport width, even when the download's title is long enough to compress the space available to the badge.

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

#### Scenario: Completed, pending, downloading, and no-retry failed badges show emoji only
- **WHEN** the viewport is narrower than the mobile breakpoint and a download's status is `completed`, `pending`, `downloading`, or `failed` with a retry count of zero
- **THEN** its status badge SHALL display only that status's emoji, with no text label, and that emoji SHALL be different from the emoji used for each of the other three statuses

#### Scenario: Completed, pending, downloading, and no-retry failed badges show emoji and text on desktop
- **WHEN** the viewport is at or above the mobile breakpoint and a download's status is `completed`, `pending`, `downloading`, or `failed` with a retry count of zero
- **THEN** its status badge SHALL display that status's emoji followed by its text label (e.g. `✅ Complété`, `❌ Échec`)

#### Scenario: Failed badge with retries shows the retry count instead of the status word
- **WHEN** the viewport is narrower than the mobile breakpoint and a download's status is `failed` and its retry count is greater than zero (e.g. 3)
- **THEN** its status badge SHALL display the failed emoji followed by the retry count in parentheses (e.g. `❌ (3×)`), and SHALL NOT display the word "Échec"/"Failed"

#### Scenario: Failed badge with retries shows the status word and retry count on desktop
- **WHEN** the viewport is at or above the mobile breakpoint and a download's status is `failed` and its retry count is greater than zero (e.g. 3)
- **THEN** its status badge SHALL display the failed emoji, the "Échec"/"Failed" text, and the retry count together (e.g. `❌ Échec (3×)`)

#### Scenario: Status badge never wraps even with a long title
- **WHEN** a download's title is long enough to compress the horizontal space left for the status badge, on any viewport width
- **THEN** the status badge SHALL remain on a single line, with its emoji (and text, when present) never breaking onto a second line
