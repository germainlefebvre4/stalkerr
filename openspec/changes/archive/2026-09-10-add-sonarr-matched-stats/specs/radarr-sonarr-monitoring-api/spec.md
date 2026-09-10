## MODIFIED Requirements

### Requirement: Sonarr monitored series total count without per-series fan-out
The system SHALL expose the total number of Sonarr-monitored series without fetching each series' episode list from Sonarr. This no-fan-out constraint applies only to the monitored total; the matched count reported alongside it is computed separately per the "Sonarr full-catalog matched series stats" requirement, which is explicitly permitted to consult and populate the Sonarr match-status cache.

#### Scenario: Total count avoids per-series episode fetches
- **WHEN** the Sonarr monitored-series total is requested
- **THEN** the system SHALL derive it from the monitored series listing alone, without issuing an additional Sonarr episode-list request per series

### Requirement: No caching or persistence of monitoring results, except the Sonarr match-status filter cache
Results SHALL be computed fresh on every request to these endpoints; the system SHALL NOT persist Radarr/Sonarr monitored items, occurrence counts, or Radarr match status between requests. The sole exception is an in-memory, per-process cache of each Sonarr-monitored series' matched/unmatched status, used to serve the match-status filter and the Sonarr full-catalog matched-count stat without a full per-series episode fan-out on every request once warm.

#### Scenario: Repeated request after upstream state changes
- **WHEN** the same listing endpoint is called twice, and Radarr/Sonarr's monitored catalog changed between the two calls
- **THEN** the second response SHALL reflect the updated catalog without requiring any cache invalidation step, except for Sonarr matched/unmatched status as covered by the match-status cache requirement

### Requirement: Sonarr match-status cache backs the match-status filter
The system SHALL maintain an in-memory, per-process cache mapping each Sonarr-monitored series to its matched/unmatched status (as defined by the match-status filter requirement), shared by the match-status filter and the Sonarr full-catalog matched-count stats endpoint, populated lazily and invalidated only by the existing manual refresh action for the Séries section - never by a background timer or scheduled job.

#### Scenario: Cache miss on first filtered request
- **WHEN** a match-status-filtered Sonarr listing request is made and a monitored series has no cached status (or its cached status was invalidated)
- **THEN** the system SHALL compute that series' matched/unmatched status by fetching its episodes from Sonarr, store the result in the cache, and use it to satisfy the filter

#### Scenario: Cache hit on subsequent filtered requests
- **WHEN** a match-status-filtered Sonarr listing request is made for a series whose status is already cached
- **THEN** the system SHALL use the cached status without an additional Sonarr episode-list fetch for that series

#### Scenario: Stats endpoint populates the cache on a cold start
- **WHEN** the Sonarr full-catalog matched-count stat is requested and one or more monitored series have no cached status
- **THEN** the system SHALL compute those series' matched/unmatched status by fetching their episodes from Sonarr, store each result in the cache, and use it to compute the matched count

#### Scenario: Stats endpoint reuses a warm cache
- **WHEN** the Sonarr full-catalog matched-count stat is requested and every monitored series already has a cached status
- **THEN** the system SHALL compute the matched count entirely from the cache, without issuing any additional Sonarr episode-list requests

#### Scenario: Manual refresh invalidates the cache
- **WHEN** the user triggers the Séries section's manual refresh action
- **THEN** the cached matched/unmatched status for affected series SHALL be invalidated so the next filtered request recomputes it

#### Scenario: No background refresh of the cache
- **WHEN** time passes with the application running and no manual refresh is triggered
- **THEN** the system SHALL NOT recompute or expire cached entries on its own

#### Scenario: Cache is process-local and not persisted
- **WHEN** the backend process restarts
- **THEN** the cache SHALL start empty; cached status SHALL NOT be persisted to the database or any external store

## ADDED Requirements

### Requirement: Sonarr full-catalog matched series stats via match-status cache
The system SHALL expose, alongside the Sonarr monitored-series total, the count of those monitored series with at least one matched monitored episode (as defined by the match-status filter's per-series matched/unmatched semantics), computed over the entire monitored catalog by consulting the Sonarr match-status cache.

#### Scenario: Stats reflect the full monitored catalog
- **WHEN** the stats endpoint is called
- **THEN** it SHALL return both the total monitored series count and the matched-series count, both computed over every currently monitored series, not just one page

#### Scenario: A series counts as matched with at least one matched episode
- **WHEN** the stats endpoint computes the matched-series count
- **THEN** a series SHALL count as matched when at least one of its monitored episodes has a playlist match, matching the semantics already used by the Séries match-status filter

#### Scenario: Episode fetch failure during stats computation
- **WHEN** fetching a monitored series' episodes from Sonarr fails while computing the matched-series count
- **THEN** the stats endpoint SHALL report the Sonarr section as unreachable (the same `sonarr_error` used elsewhere for Sonarr connectivity failures) rather than returning a partial or incorrect matched count, without affecting the Radarr section of the same response
