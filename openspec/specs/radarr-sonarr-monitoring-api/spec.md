# radarr-sonarr-monitoring-api Specification

## Purpose

Expose read-only, on-demand backend endpoints that report Radarr's and Sonarr's currently monitored movies/series together with whether the local M3U playlist already has a matching entry for each, without persisting any new data or running any background sync.

## Requirements

### Requirement: List Radarr monitored movies with playlist match status
The system SHALL expose a paginated endpoint that lists every movie currently monitored in Radarr (regardless of whether Radarr already has a file for it), each annotated with whether a matching entry exists in the local playlist database.

#### Scenario: Monitored movie with a playlist match
- **WHEN** a Radarr-monitored movie's TVDB ID, TMDB ID, or fuzzy title+year matches a local `Movie` record
- **THEN** the response entry for that movie SHALL report a matched status referencing the local record

#### Scenario: Monitored movie with no playlist match
- **WHEN** a Radarr-monitored movie matches no local `Movie` record by TVDB ID, TMDB ID, or fuzzy title+year
- **THEN** the response entry for that movie SHALL report an unmatched status, and the request SHALL still succeed (HTTP 200) for the remaining entries

#### Scenario: Match status is independent of playlist pipeline state
- **WHEN** a matched movie's only playlist occurrences are already in the `downloaded` state
- **THEN** the movie SHALL still report a matched status (matching is not restricted to occurrences in `processed`/`failed` state)

### Requirement: Optional search parameter filters both listing endpoints by title
The Radarr movies listing endpoint and the Sonarr series listing endpoint SHALL each accept an optional `search` query parameter that filters the monitored catalog to entries whose title case-insensitively contains the given term, applied before pagination.

#### Scenario: Search term filters the full catalog before pagination
- **WHEN** a listing request includes a `search` term
- **THEN** the endpoint SHALL filter the full monitored catalog by title (case-insensitive substring match) before slicing the requested page, and `total` SHALL reflect the filtered count

#### Scenario: No search term supplied
- **WHEN** a listing request omits `search` (or supplies an empty value)
- **THEN** the endpoint SHALL behave exactly as it does today, unfiltered

#### Scenario: Search term matches nothing
- **WHEN** a `search` term matches no entry in the monitored catalog
- **THEN** the endpoint SHALL return HTTP 200 with an empty `data` array and `total` of 0, not an error

### Requirement: Optional match-status filter for both listing endpoints
The Radarr movies listing endpoint and the Sonarr series listing endpoint SHALL each accept an optional match-status filter parameter with the value `matched`, `no_match`, or absent (no filtering), applied to the full monitored catalog before pagination. For Sonarr, a series SHALL be considered `matched` when at least one of its monitored episodes has a playlist match, and `no_match` when none do, independent of the partial matched/monitored ratio already reported.

#### Scenario: Radarr filter=matched
- **WHEN** a Radarr listing request sets the match-status filter to `matched`
- **THEN** the response SHALL include only monitored movies with a playlist match

#### Scenario: Radarr filter=no_match
- **WHEN** a Radarr listing request sets the match-status filter to `no_match`
- **THEN** the response SHALL include only monitored movies with no playlist match

#### Scenario: Sonarr filter=matched
- **WHEN** a Sonarr listing request sets the match-status filter to `matched`
- **THEN** the response SHALL include only monitored series with at least one matched monitored episode

#### Scenario: Sonarr filter=no_match
- **WHEN** a Sonarr listing request sets the match-status filter to `no_match`
- **THEN** the response SHALL include only monitored series with zero matched monitored episodes

#### Scenario: No filter supplied
- **WHEN** a listing request omits the match-status filter (or supplies no value)
- **THEN** the endpoint SHALL behave exactly as it does without this feature, returning both matched and unmatched entries

#### Scenario: Match-status filter combined with search
- **WHEN** a listing request supplies both a `search` term and a match-status filter
- **THEN** the endpoint SHALL apply both constraints together before pagination, returning only entries satisfying both

### Requirement: Occurrence count per Radarr movie and Sonarr series
Each entry in the Radarr movies listing and the Sonarr series listing SHALL include the total number of matching local playlist occurrences for that entry, counting every occurrence (including duplicates at different resolutions or qualities) rather than only distinct matched episodes or movies. This count SHALL be computed only for the entries in the requested page, following the same page-scoped cost model as match status for unfiltered requests.

#### Scenario: Matched movie with duplicate quality occurrences
- **WHEN** a matched Radarr movie has several playlist occurrences in different resolutions
- **THEN** its occurrence count SHALL equal the total number of those occurrences, not 1

#### Scenario: Unmatched movie has zero occurrences
- **WHEN** a Radarr movie has no playlist match
- **THEN** its occurrence count SHALL be 0

#### Scenario: Series occurrence count aggregates across monitored episodes
- **WHEN** a Sonarr series has multiple monitored episodes, each with its own playlist occurrences
- **THEN** the series' occurrence count SHALL be the sum of occurrences across all of its monitored episodes

#### Scenario: Series with no matched episodes has zero occurrences
- **WHEN** a Sonarr series has zero monitored episodes with a playlist match
- **THEN** its occurrence count SHALL be 0

### Requirement: Radarr full-catalog matched/unmatched movie stats
The system SHALL expose an endpoint reporting the total number of Radarr-monitored movies and how many of them have a matching local playlist entry, computed across the entire monitored catalog rather than a single page.

#### Scenario: Stats reflect the full monitored catalog
- **WHEN** the stats endpoint is called
- **THEN** it SHALL return the total monitored movie count and the matched count, both computed over every currently monitored movie, not just one page

#### Scenario: Matching stays local-only
- **WHEN** the stats endpoint computes the matched count
- **THEN** it SHALL determine match status using only the local playlist database (as the existing per-page match computation does), without additional live Radarr calls beyond fetching the monitored movie list

### Requirement: Sonarr monitored series total count without per-series fan-out
The system SHALL expose the total number of Sonarr-monitored series without fetching each series' episode list from Sonarr.

#### Scenario: Total count avoids per-series episode fetches
- **WHEN** the Sonarr monitored-series total is requested
- **THEN** the system SHALL derive it from the monitored series listing alone, without issuing an additional Sonarr episode-list request per series

### Requirement: List Sonarr monitored series with per-series playlist match aggregate
The system SHALL expose a paginated endpoint that lists every series currently monitored in Sonarr, each annotated with an aggregate count of how many of its monitored episodes have a matching entry in the local playlist database (e.g. "8 of 12 monitored episodes matched").

#### Scenario: Series with partial playlist matches
- **WHEN** a Sonarr-monitored series has N monitored episodes and M of them (M <= N) have a matching local `TVShow` record
- **THEN** the response entry for that series SHALL report matched count M and monitored count N

#### Scenario: Series with no monitored episodes
- **WHEN** a Sonarr-monitored series has zero monitored episodes
- **THEN** the response entry SHALL report a matched count of 0 and a monitored count of 0, without error

### Requirement: Pagination is applied before match computation, except when filtering by match status
Both listing endpoints SHALL fetch the full lightweight Radarr/Sonarr listing first. When no match-status filter is requested, the endpoint SHALL compute playlist-match status only for the entries in the requested page, so that per-request cost (local DB lookups, and for Sonarr the per-series episode fetch from Sonarr) scales with the page size and not with the total catalog size. When a match-status filter is requested, the endpoint SHALL compute match status for the entire monitored catalog before filtering and paginating, so that the filtered `total` and each returned page are accurate.

#### Scenario: Large catalog, single page requested, no filter
- **WHEN** Radarr or Sonarr has a monitored catalog far larger than the requested page size and no match-status filter is requested
- **THEN** the endpoint SHALL perform match computation only for the entries belonging to the requested page, not the entire catalog

#### Scenario: Page size bounds Sonarr per-series episode fetches when unfiltered
- **WHEN** a page of `limit` series is requested from the Sonarr listing endpoint with no match-status filter
- **THEN** the endpoint SHALL issue at most `limit` additional Sonarr episode-list requests to compute the aggregates for that page

#### Scenario: Default and maximum page size
- **WHEN** no `limit` is supplied
- **THEN** the endpoint SHALL apply a bounded default page size
- **WHEN** a `limit` larger than the endpoint's maximum allowed page size is supplied
- **THEN** the endpoint SHALL clamp it to that maximum rather than computing matches for an unbounded number of entries

#### Scenario: Filtered request computes match status for the full catalog before pagination
- **WHEN** a listing request includes a match-status filter
- **THEN** the endpoint SHALL compute match status for every entry in the monitored catalog before applying the filter and slicing the requested page, so `total` reflects the filtered count

#### Scenario: Filtered Radarr request stays local-only
- **WHEN** the Radarr movies listing endpoint computes match status for the full catalog to satisfy a match-status filter
- **THEN** it SHALL determine match status using only the local playlist database, without additional live Radarr calls beyond fetching the monitored movie list

#### Scenario: Filtered Sonarr request relies on the match-status cache
- **WHEN** the Sonarr series listing endpoint computes match status for the full catalog to satisfy a match-status filter
- **THEN** it SHALL use the Sonarr match-status cache (see "Sonarr match-status cache backs the match-status filter") rather than fetching every monitored series' episodes from Sonarr on every filtered request

### Requirement: Matched playlist occurrences detail for a single monitored item
The system SHALL expose an endpoint returning, for a single Radarr-matched movie or Sonarr-matched series+episode, the full list of matching local playlist occurrences (at least resolution and pipeline state per occurrence), regardless of each occurrence's pipeline state.

#### Scenario: Matched movie with multiple quality occurrences
- **WHEN** a matched movie has several playlist occurrences in different resolutions and states (including `downloaded`)
- **THEN** the detail response SHALL list every occurrence with its resolution and state

#### Scenario: Matched item with no playlist occurrences
- **WHEN** a local `Movie`/`TVShow` record matches the requested item but has no associated playlist occurrences
- **THEN** the detail response SHALL return an empty occurrence list rather than an error

#### Scenario: Requested item has no local match at all
- **WHEN** the requested Radarr movie or Sonarr series+episode has no corresponding local `Movie`/`TVShow` record
- **THEN** the detail response SHALL indicate no match found, distinct from "matched but no occurrences"

### Requirement: Each upstream service's availability is reported independently
A failure to reach Radarr SHALL NOT affect the Sonarr listing endpoint's ability to serve a response, and vice versa, since the two run as independent endpoints.

#### Scenario: Radarr unreachable, Sonarr reachable
- **WHEN** Radarr does not respond or errors, and Sonarr is configured and reachable
- **THEN** the Radarr listing endpoint SHALL return an error response distinguishing "upstream unavailable" from other error causes, while the Sonarr listing endpoint remains unaffected and continues to serve results

#### Scenario: Service not configured
- **WHEN** Radarr or Sonarr has no URL/API key configured
- **THEN** the corresponding listing endpoint SHALL return a response indicating the service is not configured, distinguishable from an upstream connectivity failure

### Requirement: No caching or persistence of monitoring results, except the Sonarr match-status filter cache
Results SHALL be computed fresh on every request to these endpoints; the system SHALL NOT persist Radarr/Sonarr monitored items, occurrence counts, or Radarr match status between requests. The sole exception is an in-memory, per-process cache of each Sonarr-monitored series' matched/unmatched status, used only to serve the match-status filter without a full per-series episode fan-out on every filtered request.

#### Scenario: Repeated request after upstream state changes
- **WHEN** the same listing endpoint is called twice, and Radarr/Sonarr's monitored catalog changed between the two calls
- **THEN** the second response SHALL reflect the updated catalog without requiring any cache invalidation step, except for Sonarr matched/unmatched status as covered by the match-status cache requirement

### Requirement: Sonarr match-status cache backs the match-status filter
The system SHALL maintain an in-memory, per-process cache mapping each Sonarr-monitored series to its matched/unmatched status (as defined by the match-status filter requirement), populated lazily and invalidated only by the existing manual refresh action for the Séries section - never by a background timer or scheduled job.

#### Scenario: Cache miss on first filtered request
- **WHEN** a match-status-filtered Sonarr listing request is made and a monitored series has no cached status (or its cached status was invalidated)
- **THEN** the system SHALL compute that series' matched/unmatched status by fetching its episodes from Sonarr, store the result in the cache, and use it to satisfy the filter

#### Scenario: Cache hit on subsequent filtered requests
- **WHEN** a match-status-filtered Sonarr listing request is made for a series whose status is already cached
- **THEN** the system SHALL use the cached status without an additional Sonarr episode-list fetch for that series

#### Scenario: Manual refresh invalidates the cache
- **WHEN** the user triggers the Séries section's manual refresh action
- **THEN** the cached matched/unmatched status for affected series SHALL be invalidated so the next filtered request recomputes it

#### Scenario: No background refresh of the cache
- **WHEN** time passes with the application running and no manual refresh is triggered
- **THEN** the system SHALL NOT recompute or expire cached entries on its own

#### Scenario: Cache is process-local and not persisted
- **WHEN** the backend process restarts
- **THEN** the cache SHALL start empty; cached status SHALL NOT be persisted to the database or any external store
