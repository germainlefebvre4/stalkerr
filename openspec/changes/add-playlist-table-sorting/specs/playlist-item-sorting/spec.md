## Purpose

Defines the server-side sorting contract for the playlist item listing endpoint, including which fields can be sorted, how sort direction is validated, and how items without TMDB enrichment are ordered relative to enriched ones.

## ADDED Requirements

### Requirement: Sortable Fields for Playlist Item Listing
The system SHALL allow `GET /api/v1/items` requests to specify a `sort` query parameter selecting the field used to order results, and an `order` query parameter (`asc` or `desc`) selecting the sort direction. The system SHALL accept the following values for `sort`: `tvg_name`, `group_title`, `state`, `created_at`, `downloaded_at`, `tmdb_title`. When `sort` is omitted, the system SHALL default to `created_at`. When `order` is omitted, the system SHALL default to `desc`.

If `sort` is provided but is not one of the accepted values, the system SHALL respond with `400 Bad Request` and `ErrorResponse.error` set to `"invalid_sort_field"`. If `order` is provided but is not `asc` or `desc` (case-insensitive), the system SHALL respond with `400 Bad Request` and `ErrorResponse.error` set to `"invalid_sort_order"`, and SHALL NOT execute the underlying query with the invalid value.

#### Scenario: Sort playlist items by group title ascending
- **WHEN** a client makes a `GET` request to `/api/v1/items?sort=group_title&order=asc`
- **THEN** the system SHALL return items ordered by `group_title` in ascending alphabetical order.

#### Scenario: Sort playlist items by pipeline state
- **WHEN** a client makes a `GET` request to `/api/v1/items?sort=state&order=asc`
- **THEN** the system SHALL return items ordered by their `state` value in ascending alphabetical order.

#### Scenario: Reject an unsupported sort field
- **WHEN** a client makes a `GET` request to `/api/v1/items?sort=line_hash`
- **THEN** the system SHALL respond with `400 Bad Request` and `ErrorResponse.error` set to `"invalid_sort_field"`.

#### Scenario: Reject a malformed sort order value
- **WHEN** a client makes a `GET` request to `/api/v1/items?sort=created_at&order=created_at;DROP TABLE processed_lines`
- **THEN** the system SHALL respond with `400 Bad Request` and `ErrorResponse.error` set to `"invalid_sort_order"`, and SHALL NOT execute the query.

#### Scenario: Default ordering when no sort is specified
- **WHEN** a client makes a `GET` request to `/api/v1/items` with no `sort` or `order` parameter
- **THEN** the system SHALL return items ordered by `created_at` descending, matching prior behavior.

### Requirement: Downloaded-At Timestamp
The system SHALL record a `downloaded_at` timestamp on each playlist item at the moment its `state` transitions to `downloaded`, and SHALL expose this value via `ItemResponse.downloaded_at`. Items that have never reached the `downloaded` state SHALL have a `null` `downloaded_at`.

#### Scenario: Downloaded-at is set when an item completes download
- **WHEN** a playlist item's pipeline transitions its `state` to `downloaded`
- **THEN** the system SHALL set that item's `downloaded_at` to the time of the transition, and subsequent `GET /api/v1/items`/`GET /api/v1/items/:id` responses SHALL include that value in `downloaded_at`.

#### Scenario: Downloaded-at is absent for items not yet downloaded
- **WHEN** a client requests a playlist item whose `state` is `pending`, `downloading`, `organizing`, `processed`, or `failed`
- **THEN** the response's `downloaded_at` field SHALL be `null`.

#### Scenario: List items sorted by most recently downloaded
- **WHEN** a client makes a `GET` request to `/api/v1/items?state=downloaded&sort=downloaded_at&order=desc`
- **THEN** the system SHALL return only items in the `downloaded` state, ordered from the most recently downloaded to the least recently downloaded.

### Requirement: TMDB Title Sort Places Non-Enriched Items Last
When `sort=tmdb_title`, the system SHALL order items by the enriched TMDB title (the linked movie's title if `content_type` is `movies`, or the linked TV show's title if `content_type` is `tvshows`), and SHALL place items with no TMDB enrichment (no linked movie or TV show) at the end of the result set regardless of the requested `order` direction.

#### Scenario: Non-enriched items sort last in ascending order
- **WHEN** a client makes a `GET` request to `/api/v1/items?sort=tmdb_title&order=asc`
- **THEN** the system SHALL return enriched items ordered by TMDB title from A to Z, followed by all non-enriched items, in that order.

#### Scenario: Non-enriched items sort last in descending order
- **WHEN** a client makes a `GET` request to `/api/v1/items?sort=tmdb_title&order=desc`
- **THEN** the system SHALL return enriched items ordered by TMDB title from Z to A, followed by all non-enriched items, in that order.
