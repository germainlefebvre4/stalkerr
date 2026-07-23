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

The frontend SHALL render an interactive pagination bar with page numbers, smart ellipsis (`...`) for navigation in large page counts, and quick jump buttons to the first (`<<`) and last (`>>`) pages.

#### Scenario: Filter items by Content Type, State, TMDB Enrichment and Search Title
- **WHEN** the user selects the "Movies" content tab, "Failed" status filter, "Non (Non Enrichi)" TMDB filter, types "Matrix" in the media name search, and types "ACTION" in the group search
- **THEN** the frontend SHALL fetch and display only playlist items of content type `movies`, in a `failed` state, not enriched by TMDB, whose media name contains "Matrix", and whose group contains "ACTION".

#### Scenario: Change page size configuration and persist preference
- **WHEN** the user selects "50" as the size from the page size dropdown
- **THEN** the frontend SHALL write "50" to `localStorage` under `stalkeer_playlist_limit`, reset the current page to 1, and fetch 50 playlist items from the backend API.

#### Scenario: Navigate with dynamic page numbers and range display
- **WHEN** the user clicks page "3" on the pagination bar
- **THEN** the frontend SHALL fetch and display the playlist items with the calculated offset, and render the record range indicator (e.g. "Affichage de 101 à 150 sur 1000 entrées").

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


