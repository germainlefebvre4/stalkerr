## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: Playlist View with Filtering
The frontend SHALL render a comprehensive M3U playlist item table allowing searching by group title and filtering by content type (`movie` or `tvshow`) AND processing state (`all`, `processed`, `pending`, `downloading`, `downloaded`, `failed`).

#### Scenario: Filter items by Content Type, State and Search Title
- **WHEN** the user selects the "Movies" content tab, "Failed" status filter, and types "Matrix" in the search input
- **THEN** the frontend SHALL fetch and display only playlist items of content type `movies`, in a `failed` state, and whose group titles contain "Matrix".

### Requirement: Real-time Monitoring Dashboard
The frontend SHALL present dedicated tabs to display active downloads (incorporating file-size progress bars) and execution logs, using polling to automatically keep content fresh. The interface MUST employ visual shimmers on loading and progress indicators, and active pulsing indicators for the API health state.

#### Scenario: Display active download progress bar and logs status
- **WHEN** the user navigates to the "Downloads" view
- **THEN** the frontend SHALL render active progress bars indicating the percentage of completion with animated gradient shimmer effects, and automatically re-fetch data every 5 seconds.
