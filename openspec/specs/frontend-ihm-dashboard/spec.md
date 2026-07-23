# frontend-ihm-dashboard Specification

## Purpose
TBD - created by archiving change frontend-ihm-react-radix. Update Purpose after archive.
## Requirements
### Requirement: Playlist View with Filtering
The frontend SHALL render a comprehensive M3U playlist item table allowing searching by name, filtering by content type (`movie` or `tvshow`), and paginated navigation.
The frontend SHALL allow configuring the page size (10, 50, 100 entries per page), default to 50 entries, and persist this choice in the browser's `localStorage` under the key `stalkeer_playlist_limit`.
The frontend SHALL render an interactive pagination bar with page numbers, smart ellipsis (`...`) for navigation in large page counts, and quick jump buttons to the first (`<<`) and last (`>>`) pages.

#### Scenario: Filter items by Content Type and Search Title
- **WHEN** the user selects the "Movies" filter tab and types "Matrix" in the search input
- **THEN** the frontend SHALL fetch and display only playlist items of type `movies` whose names contain "Matrix".

#### Scenario: Change page size configuration and persist preference
- **WHEN** the user selects "50" as the page size from the page size dropdown
- **THEN** the frontend SHALL write "50" to `localStorage` under `stalkeer_playlist_limit`, reset the current page to 1, and fetch 50 playlist items from the backend API.

#### Scenario: Navigate with dynamic page numbers and range display
- **WHEN** the user clicks page "3" on the pagination bar
- **THEN** the frontend SHALL fetch and display the playlist items with the calculated offset, and render the record range indicator (e.g. "Affichage de 101 à 150 sur 1000 entrées").

### Requirement: Real-time Monitoring Dashboard
The frontend SHALL present dedicated tabs or tables to display active downloads (incorporating file-size progress bars) and execution logs, using polling to automatically keep content fresh. The frontend SHALL also persist the active tab state in the browser's `localStorage` and restore it upon page refresh.

#### Scenario: Display active download progress bar and logs status
- **WHEN** the user navigates to the "Downloads" view
- **THEN** the frontend SHALL render active progress bars indicating the percentage of completion for elements in the `downloading` state, and automatically re-fetch the data every 10 seconds.

#### Scenario: Persist and restore active tab on refresh
- **WHEN** the user selects a tab and refreshes the browser page
- **THEN** the frontend SHALL write the selected tab name to `localStorage` upon switch, and read/restore it upon mount to active the correct view.

### Requirement: Modern Light-Themed Radix UI Dialog for File Re-organization
The frontend SHALL offer a modal dialog powered by Radix UI `Dialog` primitives that enables moving entire movie or TV show parent directories to a target directory.

#### Scenario: Confirm folder move from completed items
- **WHEN** the user clicks the "Déplacer ⇄" button on a completed item, selects a destination path, and confirms
- **THEN** the frontend SHALL issue a `POST` request to the media management move endpoint, display a success toast upon success, and refresh the UI state.

