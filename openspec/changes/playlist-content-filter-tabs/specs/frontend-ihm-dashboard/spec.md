## MODIFIED Requirements

### Requirement: Playlist View with Filtering
The frontend SHALL render a comprehensive M3U playlist item table allowing:
- Searching by media name (`tvg_name`).
- Filtering/searching by group title (`group_title`).
- Filtering by content type (`all`, `movies`, `tvshows`) via a segmented tab control, using the same tab mechanism (roving tabindex, arrow-key navigation, `data-state="active"` styling) as the Radarr/Sonarr suite's sub-tabs.
- Filtering by processing state (`all`, `processed`, `pending`, `downloading`, `organizing`, `downloaded`, `failed`).
- Filtering by TMDB enrichment state (`all`, `yes`, `no`).
- Paginated navigation.
- Sorting by clicking any of the following column headers: media name, group/category, TMDB enrichment, pipeline state, created date, downloaded date. Clicking a header not currently sorted SHALL sort by that column ascending; clicking the currently-sorted header SHALL toggle between ascending and descending; the sorted header SHALL display a visual indicator of the active sort direction.

The frontend SHALL allow configuring the page size (10, 50, 100 entries per page), default to 10 entries, and persist this choice in the browser's `localStorage` under the key `stalkeer_playlist_limit`.

The frontend SHALL render an interactive pagination bar with page numbers, smart ellipsis (`...`) for navigation in large page counts, and quick jump buttons to the first (`<<`) and last (`>>`) pages, plus a direct "go to page" input allowing the user to type a page number and jump straight to it.

The frontend SHALL reflect the current page, page size, all active filters (media name search, group search, content type, TMDB enrichment, pipeline state), and the active sort column/direction in the browser URL's query string, and SHALL restore this exact state from the URL on page load or refresh.

#### Scenario: Filter items by Content Type, State, TMDB Enrichment and Search Title
- **WHEN** the user selects the "Movies" content tab, "Failed" status filter, "Non (Non Enrichi)" TMDB filter, types "Matrix" in the media name search, and types "ACTION" in the group search
- **THEN** the frontend SHALL fetch and display only playlist items of content type `movies`, in a `failed` state, not enriched by TMDB, whose media name contains "Matrix", and whose group contains "ACTION".

#### Scenario: Default to 10 entries per page with no stored preference
- **WHEN** the user opens the Playlist view for the first time, with no `stalkeer_playlist_limit` value in `localStorage` and no `limit` value in the URL query string
- **THEN** the frontend SHALL fetch and display 10 playlist items per page.

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

#### Scenario: Navigate the content-type filter with arrow keys
- **WHEN** the "Movies" content-type tab is focused and the user presses the right arrow key
- **THEN** the frontend SHALL move focus to the "TV Shows" content-type tab, following the same roving-tabindex keyboard navigation as the Radarr/Sonarr sub-tabs

#### Scenario: Active content-type tab is visually indicated
- **WHEN** the "TV Shows" content-type tab is selected
- **THEN** the frontend SHALL mark it as the active tab via the same `data-state="active"` styling mechanism used by the Radarr/Sonarr sub-tabs, and the previously selected tab SHALL no longer show the active styling
