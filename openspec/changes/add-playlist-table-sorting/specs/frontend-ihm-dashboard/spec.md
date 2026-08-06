## MODIFIED Requirements

### Requirement: Playlist View with Filtering
The frontend SHALL render a comprehensive M3U playlist item table allowing:
- Searching by media name (`tvg_name`).
- Filtering/searching by group title (`group_title`).
- Filtering by content type (`all`, `movies`, `tvshows`).
- Filtering by processing state (`all`, `processed`, `pending`, `downloading`, `organizing`, `downloaded`, `failed`).
- Filtering by TMDB enrichment state (`all`, `yes`, `no`).
- Paginated navigation.
- Sorting by clicking any of the following column headers: media name, group/category, TMDB enrichment, pipeline state, created date, downloaded date. Clicking a header not currently sorted SHALL sort by that column ascending; clicking the currently-sorted header SHALL toggle between ascending and descending; the sorted header SHALL display a visual indicator of the active sort direction.

The frontend SHALL allow configuring the page size (10, 50, 100 entries per page), default to 50 entries, and persist this choice in the browser's `localStorage` under the key `stalkeer_playlist_limit`.

The frontend SHALL render an interactive pagination bar with page numbers, smart ellipsis (`...`) for navigation in large page counts, and quick jump buttons to the first (`<<`) and last (`>>`) pages, plus a direct "go to page" input allowing the user to type a page number and jump straight to it.

The frontend SHALL reflect the current page, page size, all active filters (media name search, group search, content type, TMDB enrichment, pipeline state), and the active sort column/direction in the browser URL's query string, and SHALL restore this exact state from the URL on page load or refresh.

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
- **WHEN** the user is viewing page 3 with an active content-type and state filter and a column sorted descending, and refreshes the browser
- **THEN** the frontend SHALL read the page, page size, filters, and sort column/direction from the URL query string and re-render the exact same filtered, sorted page, instead of resetting to page 1 with no filters and the default sort.

#### Scenario: Sort playlist table by a column header
- **WHEN** the user clicks the "Créé le" column header while no sort is active
- **THEN** the frontend SHALL request items sorted by `created_at` ascending, reset the current page to 1, and display an ascending sort indicator on that header.

#### Scenario: Toggle sort direction on repeated header click
- **WHEN** the user clicks a column header that is already the active ascending sort
- **THEN** the frontend SHALL request items sorted by the same column descending, reset the current page to 1, and display a descending sort indicator on that header.

#### Scenario: List the most recently downloaded items
- **WHEN** the user filters by pipeline state "downloaded" and clicks the "Téléchargé le" column header twice (to sort descending)
- **THEN** the frontend SHALL request items with `state=downloaded` sorted by `downloaded_at` descending, so the most recently downloaded items appear first.

### Requirement: Playlist Creation Date Format
The frontend SHALL render the "Créé le" column of the playlist table using a fixed `DD/MM/YYYY` format with zero-padded day and month, independent of the browser's locale settings. The frontend SHALL render a new "Téléchargé le" column using the same fixed `DD/MM/YYYY` format when a playlist item has a non-null `downloaded_at`, and SHALL render a placeholder (e.g. "—") when `downloaded_at` is `null`.

#### Scenario: Display a creation date with zero-padded day and month
- **WHEN** a playlist item's `created_at` value corresponds to March 5th, 2026
- **THEN** the "Créé le" column SHALL display "05/03/2026".

#### Scenario: Display a downloaded date with zero-padded day and month
- **WHEN** a playlist item's `downloaded_at` value corresponds to March 5th, 2026
- **THEN** the "Téléchargé le" column SHALL display "05/03/2026".

#### Scenario: Display placeholder for items not yet downloaded
- **WHEN** a playlist item's `downloaded_at` is `null`
- **THEN** the "Téléchargé le" column SHALL display "—".
