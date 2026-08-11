# frontend-ihm-dashboard Specification

## Purpose
TBD - created by archiving change frontend-ihm-react-radix. Update Purpose after archive.
## Requirements
### Requirement: Playlist View with Filtering
The frontend SHALL render a comprehensive M3U playlist item table allowing:
- Searching by media name (`tvg_name`).
- Filtering/searching by group title (`group_title`).
- Filtering by content type (`all`, `movies`, `tvshows`).
- Filtering by processing state (`all`, `processed`, `pending`, `downloading`, `organizing`, `downloaded`, `failed`).
- Filtering by TMDB enrichment state (`all`, `yes`, `no`).
- Paginated navigation.

The frontend SHALL allow configuring the page size (10, 50, 100 entries per page), default to 50 entries, and persist this choice in the browser's `localStorage` under the key `stalkeer_playlist_limit`.

The frontend SHALL render an interactive pagination bar with page numbers, smart ellipsis (`...`) for navigation in large page counts, and quick jump buttons to the first (`<<`) and last (`>>`) pages, plus a direct "go to page" input allowing the user to type a page number and jump straight to it.

The frontend SHALL reflect the current page, page size, and all active filters (media name search, group search, content type, TMDB enrichment, pipeline state) in the browser URL's query string, and SHALL restore this exact state from the URL on page load or refresh.

#### Scenario: Filter items by Content Type, State, TMDB Enrichment and Search Title
- **WHEN** the user selects the "Movies" content tab, "Failed" status filter, "Non (Non Enrichi)" TMDB filter, types "Matrix" in the media name search, and types "ACTION" in the group search
- **THEN** the frontend SHALL fetch and display only playlist items of content type `movies`, in a `failed` state, not enriched by TMDB, whose media name contains "Matrix", and whose group contains "ACTION".

#### Scenario: Change page size configuration and persist preference
- **WHEN** the user selects "50" as the size from the page size dropdown
- **THEN** the frontend SHALL write "50" to `localStorage` under `stalkeer_playlist_limit`, reset the current page to 1, and fetch 50 playlist items from the backend API.

#### Scenario: Navigate with dynamic page numbers and range display
- **WHEN** the user clicks page "3" on the pagination bar
- **THEN** the frontend SHALL fetch and display the playlist items with the calculated offset, and render the record range indicator (e.g. "Affichage de 101 à 150 sur 1000 entrées").

#### Scenario: Jump directly to a chosen page number
- **WHEN** the user types "7" into the "go to page" input and confirms
- **THEN** the frontend SHALL navigate to page 7 (clamped to the valid page range) and fetch/display the corresponding playlist items.

#### Scenario: Restore pagination and filters from the URL after a refresh
- **WHEN** the user is viewing page 3 with an active content-type and state filter, and refreshes the browser
- **THEN** the frontend SHALL read the page, page size, and filters from the URL query string and re-render the exact same filtered page, instead of resetting to page 1 with no filters.

### Requirement: Mobile Collapsible Advanced Filters
On mobile viewports, the frontend SHALL render the Playlist advanced filter grid (media name search, group search, TMDB enrichment, pipeline state) inside a disclosure section that is collapsed by default, so the table/card list below it is visible without additional scrolling. The content-type filter (`all`, `movies`, `tvshows`) SHALL remain outside this disclosure and always visible on mobile. The disclosure's header SHALL display a count of advanced filters currently set to a non-default value (i.e. not `all` and not empty). The disclosure's collapsed/expanded state SHALL NOT be persisted across page loads or reflected in the URL or `localStorage`; it SHALL always start collapsed when the Playlist tab is mounted on a mobile viewport, regardless of whether an advanced filter is already active from a restored URL. On non-mobile viewports, the advanced filter grid SHALL continue to render inline and always visible, with no disclosure control.

#### Scenario: Advanced filters are collapsed on mobile by default
- **WHEN** the user opens the Playlist tab on a mobile viewport with no active advanced filters
- **THEN** the frontend SHALL render the content-type filter always visible, render the advanced filter grid collapsed inside a disclosure showing no active-filter count, and render the table/card list immediately below without requiring the user to scroll past the advanced filter grid

#### Scenario: Expanding the disclosure reveals the advanced filter grid
- **WHEN** the user taps the collapsed advanced-filters disclosure header on a mobile viewport
- **THEN** the frontend SHALL expand the disclosure to show the media name search, group search, TMDB enrichment, and pipeline state controls, without changing the content-type filter's visibility or position

#### Scenario: Active filter count is visible while collapsed
- **WHEN** the pipeline state filter is set to `failed` and the TMDB enrichment filter is set to `yes`, and the advanced-filters disclosure is collapsed
- **THEN** the frontend SHALL display a count of `2` active advanced filters on the disclosure header

#### Scenario: Disclosure state resets on remount regardless of restored filters
- **WHEN** the user navigates to the Playlist tab on a mobile viewport with a pipeline state filter of `failed` restored from the URL query string
- **THEN** the frontend SHALL still render the advanced-filters disclosure collapsed by default, while the restored `failed` filter SHALL remain applied to the fetched playlist items and reflected in the disclosure's active-filter count

#### Scenario: Desktop rendering is unaffected
- **WHEN** the user opens the Playlist tab on a non-mobile viewport
- **THEN** the frontend SHALL render the advanced filter grid inline and fully visible, with no disclosure control and no collapse/expand behavior

### Requirement: Real-time Monitoring Dashboard
The frontend SHALL present dedicated tabs to display active downloads (incorporating file-size progress bars) and execution logs, using polling to automatically keep content fresh. The interface MUST employ visual shimmers on loading and progress indicators, and active pulsing indicators for the API health state. The frontend SHALL also persist the active tab state in the browser's `localStorage` and restore it upon page refresh.

#### Scenario: Display active download progress bar and logs status
- **WHEN** the user navigates to the "Downloads" view
- **THEN** the frontend SHALL render active progress bars indicating the percentage of completion with animated gradient shimmer effects, and automatically re-fetch data every 5 seconds.

#### Scenario: Persist and restore active tab on refresh
- **WHEN** the user selects a tab and refreshes the browser page
- **THEN** the frontend SHALL write the selected tab name to `localStorage` upon switch, and read/restore it upon mount to active the correct view.

### Requirement: Modern Light-Themed Radix UI Dialog for File Re-organization
The frontend SHALL offer a modal dialog powered by Radix UI `Dialog` primitives that enables moving entire movie or TV show parent directories to a target directory.

#### Scenario: Confirm folder move from completed items
- **WHEN** the user clicks the "Déplacer ⇄" button on a completed item, selects a destination path, and confirms
- **THEN** the frontend SHALL issue a `POST` request to the media management move endpoint, display a success toast upon success, and refresh the UI state.

### Requirement: Statistics KPI Cards
The frontend SHALL display a horizontal grid of 4 visual statistics cards beneath the main header, fetching data from `/api/v1/stats`. These cards must display:
- Total items in the M3U playlist.
- Number of movies identified.
- Number of TV shows identified.
- Download success percentage (downloaded count vs failed count).

#### Scenario: Display global stats on dashboard load
- **WHEN** the dashboard loads or is refreshed
- **THEN** the frontend SHALL fetch statistics from `/api/v1/stats` and render them in the KPI grid with modern, translucent pastel backgrounds.

### Requirement: Pipeline Reset Action on Playlist
The frontend SHALL offer a "Réinitialiser ↻" button on each playlist table row for items that are processed, downloading, or failed. Clicking this button SHALL trigger a `POST` request to the appropriate backend endpoint (`/api/v1/movies/:id/reset` or `/api/v1/tvshows/:id/reset`).

#### Scenario: Trigger reset pipeline of a failed movie
- **WHEN** the user clicks the "Réinitialiser ↻" button on a failed movie row in the playlist table
- **THEN** the frontend SHALL send a `POST` request to `/api/v1/movies/:id/reset`, display a success toast on success, and refresh the playlist items list.

### Requirement: Playlist Creation Date Format
The frontend SHALL render the "Créé le" column of the playlist table using a fixed `DD/MM/YYYY` format with zero-padded day and month, independent of the browser's locale settings.

#### Scenario: Display a creation date with zero-padded day and month
- **WHEN** a playlist item's `created_at` value corresponds to March 5th, 2026
- **THEN** the "Créé le" column SHALL display "05/03/2026".

### Requirement: Pipeline State Visual Indicator
The frontend SHALL render the pipeline state badge (wherever the playlist item's `state` is displayed, including the playlist table and its details sidepanel) using a color that reflects whether the state represents success, an in-progress/waiting condition, or a failure:
- `processed` and `downloaded` SHALL use the success (green) style.
- `pending` SHALL use the pending/waiting (amber) style.
- `downloading` and `organizing` SHALL use the in-progress (blue) style.
- `failed` SHALL use the failure (red) style.

#### Scenario: Display a processed item with a success badge
- **WHEN** a playlist item has `state` equal to `processed`
- **THEN** the frontend SHALL render its state badge using the success (green) style, not the pending (amber) style.

#### Scenario: Display a pending item with a pending badge
- **WHEN** a playlist item has `state` equal to `pending`
- **THEN** the frontend SHALL render its state badge using the pending (amber) style.


