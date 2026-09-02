## MODIFIED Requirements

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

### Requirement: No caching or persistence of monitoring results, except the Sonarr match-status filter cache
Results SHALL be computed fresh on every request to these endpoints; the system SHALL NOT persist Radarr/Sonarr monitored items, occurrence counts, or Radarr match status between requests. The sole exception is an in-memory, per-process cache of each Sonarr-monitored series' matched/unmatched status, used only to serve the match-status filter without a full per-series episode fan-out on every filtered request.

#### Scenario: Repeated request after upstream state changes
- **WHEN** the same listing endpoint is called twice, and Radarr/Sonarr's monitored catalog changed between the two calls
- **THEN** the second response SHALL reflect the updated catalog without requiring any cache invalidation step, except for Sonarr matched/unmatched status as covered by the match-status cache requirement

## ADDED Requirements

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
