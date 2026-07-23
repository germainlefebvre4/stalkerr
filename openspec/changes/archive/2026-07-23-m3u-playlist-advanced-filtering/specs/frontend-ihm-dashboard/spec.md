## MODIFIED Requirements

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
