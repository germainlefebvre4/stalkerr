# playlist-grouped-media-view Specification

## Purpose

Defines the grouped, movie/TV-show-level alternative to the Playlist Items view: a sub-tab that collapses episode-level rows into one row per movie or per show, so users can see what was fetched recently without paging through every episode.

## Requirements

### Requirement: Grouped Media Sub-Tab
The Playlist page SHALL offer a "Films & Séries" sub-tab alongside the existing "Items" sub-tab, independent of the existing content-type filter buttons (All/Movies/TVShows), which continue to apply within either sub-tab.

#### Scenario: Switching to the grouped view
- **WHEN** the user selects the "Films & Séries" sub-tab
- **THEN** the system SHALL replace the per-item table with the grouped movie/TV-show table, keeping the currently active content-type, state, TMDB-enrichment, and search filters applied.

#### Scenario: Switching back to the items view
- **WHEN** the user selects the "Items" sub-tab while the grouped view is active
- **THEN** the system SHALL restore the existing per-item table with the same filters still applied.

### Requirement: Server-Side Grouped Listing Endpoint
The system SHALL expose `GET /api/v1/items/grouped`, which returns one entry per distinct movie (grouped by TMDB movie identity) and one entry per distinct TV show (grouped by TMDB show identity, independent of season/episode), paginated via `limit`/`offset` query parameters with the same semantics as `GET /api/v1/items`. The reported pagination total SHALL count distinct groups, not underlying items.

#### Scenario: A show with many episodes collapses into one group
- **WHEN** a TV show has 100 processed episodes across 5 seasons ingested overnight
- **THEN** `GET /api/v1/items/grouped` SHALL return exactly one entry for that show, and that entry SHALL count as one unit toward `limit`/`offset` pagination, not 100.

#### Scenario: Pagination reflects distinct groups
- **WHEN** the result set contains 3 distinct movies and 2 distinct TV shows matching the current filters
- **THEN** the endpoint's reported total SHALL be 5, regardless of how many underlying `processed_lines` rows belong to those movies/shows.

### Requirement: Grouped Row Content
Each movie or TV-show group entry SHALL include the title and year. Each TV-show group entry SHALL additionally include the range of seasons covered by the episodes contributing to that group (e.g. `S01`-`S05`), computed from the minimum and maximum non-null season values among those episodes; movie group entries SHALL NOT include a season range.

#### Scenario: Movie group has no season range
- **WHEN** a movie group entry is returned
- **THEN** it SHALL include its title and year and SHALL NOT include season range information.

#### Scenario: TV show group spans multiple seasons
- **WHEN** a TV show's contributing episodes span seasons 1 through 5
- **THEN** the group entry's season range SHALL be reported as spanning `S01` to `S05`.

#### Scenario: TV show group is a single season
- **WHEN** all of a TV show's contributing episodes belong to season 2
- **THEN** the group entry's season range SHALL be reported as a single season, `S02`, rather than a range.

### Requirement: Default Ordering by Most Recent Activity
`GET /api/v1/items/grouped` SHALL order results by the most recent `created_at` among each group's contributing items, descending, and SHALL NOT accept a `sort` parameter to change this ordering.

#### Scenario: A show with a new episode outranks an older, larger show
- **WHEN** Show A last received a new episode yesterday and Show B (with more total episodes) received a new episode last night
- **THEN** Show B SHALL be ordered before Show A in the grouped listing.

### Requirement: Existing Item Filters Apply Before Grouping
`GET /api/v1/items/grouped` SHALL accept the same `content_type`, `state`, `group_title`, `tvg_name`, and `tmdb_enriched` query parameters as `GET /api/v1/items`, with identical matching semantics, applied to the underlying items before they are grouped. A group SHALL be included in the response if and only if at least one of its underlying items matches the active filters, and any aggregated field (season range, most-recent-activity ordering) SHALL be computed only from the matching items.

#### Scenario: A group is hidden when no underlying item matches
- **WHEN** the `state=downloaded` filter is applied and a TV show has zero episodes in the `downloaded` state
- **THEN** that show SHALL NOT appear in the grouped listing.

#### Scenario: A group's season range reflects only the matching items
- **WHEN** the `state=downloaded` filter is applied and a TV show has episodes in seasons 1 through 5 but only its season-2 episodes are `downloaded`
- **THEN** that show's group entry SHALL appear with a season range of `S02`, not `S01`-`S05`.

### Requirement: Unmatched Items Grouped Into Fixed Pseudo-Groups
Items with no TMDB association (`content_type=movies` with no linked movie, or `content_type=tvshows` with no linked TV show) SHALL NOT be distributed into per-title groups. Instead, `GET /api/v1/items/grouped` SHALL report at most two fixed pseudo-group entries — one for unmatched movies and one for unmatched TV shows — each present only when at least one matching unmatched item exists.

#### Scenario: Unmatched movies produce a single pseudo-group
- **WHEN** 6 movie items have no linked movie record and match the active filters
- **THEN** the response SHALL include exactly one "unmatched movies" pseudo-group entry, not 6 separate entries.

#### Scenario: Pseudo-group absent when there is nothing unmatched
- **WHEN** every TV-show item matching the active filters has a linked TV show record
- **THEN** the response SHALL NOT include an "unmatched TV shows" pseudo-group entry.

### Requirement: Grouping Scope Limited to Movies and TV Shows
`GET /api/v1/items/grouped` SHALL only ever produce entries for `content_type` values `movies` and `tvshows`. Items with `content_type` of `channels` or `uncategorized` SHALL be excluded from the grouped listing regardless of the `content_type` filter value.

#### Scenario: Channels are never grouped
- **WHEN** the grouped endpoint is queried with no `content_type` filter
- **THEN** the response SHALL contain no entries derived from `channels` or `uncategorized` items.

### Requirement: Expanding a Group Lists Its Underlying Items
Clicking a group row (movie, TV show, or unmatched pseudo-group) in the "Films & Séries" view SHALL expand an inline section listing that group's underlying items, rendered with the same columns, state badges, and per-item actions (association/correction, pipeline reset) as the "Items" view's table, independently paginated from the group's own pagination.

#### Scenario: Expanding a TV show with many episodes
- **WHEN** the user clicks a TV-show group row whose 100 episodes span 5 seasons
- **THEN** the system SHALL expand an inline item list scoped to that show's episodes, showing a first page of items with its own pagination controls rather than all 100 at once.

#### Scenario: Expanding a movie group
- **WHEN** the user clicks a movie group row
- **THEN** the system SHALL expand an inline item list scoped to items linked to that movie, with the same columns and actions as the Items view.

#### Scenario: Expanding an unmatched pseudo-group
- **WHEN** the user clicks the "unmatched TV shows" pseudo-group row
- **THEN** the system SHALL expand an inline item list scoped to TV-show items with no linked TV show record.

### Requirement: Item Listing Supports Filtering by Movie or Show Identity
`GET /api/v1/items` SHALL accept two additional optional query parameters: `movie_id` (integer), restricting results to items linked to that movie, and `tmdb_id` (integer), restricting results to `tvshows`-content-type items whose linked TV show's TMDB id matches the given value regardless of season or episode.

#### Scenario: Filtering items by movie_id
- **WHEN** a client requests `GET /api/v1/items?movie_id=123`
- **THEN** the system SHALL return only items whose linked movie has id `123`.

#### Scenario: Filtering items by tmdb_id returns all of a show's episodes
- **WHEN** a client requests `GET /api/v1/items?content_type=tvshows&tmdb_id=456`
- **THEN** the system SHALL return every item linked to a TV show record whose TMDB id is `456`, across all of that show's seasons and episodes.

### Requirement: Date-Group Headers Reused for Grouped View
When the "Films & Séries" view is displayed, the system SHALL render the same day-based group headers ("Aujourd'hui"/"Hier"/full date, localized) used by the Items view, positioned above the first group entry of each distinct calendar day, using each group entry's most-recent-activity timestamp as its date.

#### Scenario: Groups from the same day share one header
- **WHEN** three group entries all have their most-recent activity on the same calendar day
- **THEN** the system SHALL render one date-group header above the first of the three, with no separator between the other two.
