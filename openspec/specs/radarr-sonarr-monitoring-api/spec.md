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

### Requirement: Pagination is applied before match computation
Both listing endpoints SHALL fetch the full lightweight Radarr/Sonarr listing first, then compute playlist-match status only for the entries in the requested page, so that per-request cost (local DB lookups, and for Sonarr the per-series episode fetch from Sonarr) scales with the page size and not with the total catalog size.

#### Scenario: Large catalog, single page requested
- **WHEN** Radarr or Sonarr has a monitored catalog far larger than the requested page size
- **THEN** the endpoint SHALL perform match computation only for the entries belonging to the requested page, not the entire catalog

#### Scenario: Page size bounds Sonarr per-series episode fetches
- **WHEN** a page of `limit` series is requested from the Sonarr listing endpoint
- **THEN** the endpoint SHALL issue at most `limit` additional Sonarr episode-list requests to compute the aggregates for that page

#### Scenario: Default and maximum page size
- **WHEN** no `limit` is supplied
- **THEN** the endpoint SHALL apply a bounded default page size
- **WHEN** a `limit` larger than the endpoint's maximum allowed page size is supplied
- **THEN** the endpoint SHALL clamp it to that maximum rather than computing matches for an unbounded number of entries

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

### Requirement: No caching or persistence of monitoring results
Results SHALL be computed fresh on every request to these endpoints; the system SHALL NOT persist Radarr/Sonarr monitored items or their computed match status between requests.

#### Scenario: Repeated request after upstream state changes
- **WHEN** the same listing endpoint is called twice, and Radarr/Sonarr's monitored catalog changed between the two calls
- **THEN** the second response SHALL reflect the updated catalog without requiring any cache invalidation step
