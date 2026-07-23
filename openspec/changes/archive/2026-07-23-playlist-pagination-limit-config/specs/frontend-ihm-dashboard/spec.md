## MODIFIED Requirements

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
