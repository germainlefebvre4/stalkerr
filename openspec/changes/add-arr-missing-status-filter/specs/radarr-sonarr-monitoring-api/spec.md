## MODIFIED Requirements

### Requirement: List Radarr monitored movies with playlist match status
The system SHALL expose a paginated endpoint that lists Radarr movies matching the requested status filter (see "Optional status filter for both listing endpoints"), each annotated with whether a matching entry exists in the local playlist database, and with Radarr's own `monitored` state and whether Radarr already has a file for it (`hasFile`). When no status filter is supplied, the endpoint SHALL list only movies currently monitored in Radarr, exactly as it did before the status filter existed.

#### Scenario: Monitored movie with a playlist match
- **WHEN** a Radarr-monitored movie's TVDB ID, TMDB ID, or fuzzy title+year matches a local `Movie` record
- **THEN** the response entry for that movie SHALL report a matched status referencing the local record

#### Scenario: Monitored movie with no playlist match
- **WHEN** a Radarr-monitored movie matches no local `Movie` record by TVDB ID, TMDB ID, or fuzzy title+year
- **THEN** the response entry for that movie SHALL report an unmatched status, and the request SHALL still succeed (HTTP 200) for the remaining entries

#### Scenario: Match status is independent of playlist pipeline state
- **WHEN** a matched movie's only playlist occurrences are already in the `downloaded` state
- **THEN** the movie SHALL still report a matched status (matching is not restricted to occurrences in `processed`/`failed` state)

#### Scenario: No status filter supplied preserves today's behavior
- **WHEN** a Radarr movies listing request omits the status filter
- **THEN** the endpoint SHALL return only movies currently monitored in Radarr, identical in scope to its behavior before this capability's status filter existed

### Requirement: List Sonarr monitored series with per-series playlist match aggregate
The system SHALL expose a paginated endpoint that lists Sonarr series matching the requested status filter (see "Optional status filter for both listing endpoints"), each annotated with an aggregate count of how many of its monitored episodes have a matching entry in the local playlist database (e.g. "8 of 12 monitored episodes matched"), and with Sonarr's own `monitored` state and whether the series is Missing (monitored with at least one monitored episode lacking a file in Sonarr). When no status filter is supplied, the endpoint SHALL list only series currently monitored in Sonarr, exactly as it did before the status filter existed.

#### Scenario: Series with partial playlist matches
- **WHEN** a Sonarr-monitored series has N monitored episodes and M of them (M <= N) have a matching local `TVShow` record
- **THEN** the response entry for that series SHALL report matched count M and monitored count N

#### Scenario: Series with no monitored episodes
- **WHEN** a Sonarr-monitored series has zero monitored episodes
- **THEN** the response entry SHALL report a matched count of 0 and a monitored count of 0, without error

#### Scenario: No status filter supplied preserves today's behavior
- **WHEN** a Sonarr series listing request omits the status filter
- **THEN** the endpoint SHALL return only series currently monitored in Sonarr, identical in scope to its behavior before this capability's status filter existed

## ADDED Requirements

### Requirement: Optional status filter for both listing endpoints
The Radarr movies listing endpoint and the Sonarr series listing endpoint SHALL each accept an optional, repeatable status filter parameter whose values are `monitored`, `unmonitored`, and `missing`, applied to the full catalog before pagination. An entry is `monitored` when Radarr/Sonarr reports it as monitored; `unmonitored` when it does not; `missing` when it is monitored and lacks a file (for a Radarr movie: monitored and `hasFile` is false; for a Sonarr series: monitored and at least one monitored episode lacks a file in Sonarr). When multiple values are supplied, an entry SHALL be included only if it satisfies every supplied value's condition simultaneously (logical AND), not if it satisfies any one of them. When the parameter is omitted entirely, the endpoint SHALL behave as if only `monitored` were supplied, preserving today's default behavior.

#### Scenario: Single value narrows to that status
- **WHEN** a listing request supplies `status=missing`
- **THEN** the response SHALL include only entries that are both monitored and missing a file

#### Scenario: Multiple values combine as a logical AND
- **WHEN** a listing request supplies `status=monitored&status=missing`
- **THEN** the response SHALL include only entries satisfying both conditions at once, which for these two values is equivalent to `status=missing` alone

#### Scenario: Contradictory combination yields an empty result
- **WHEN** a listing request supplies `status=unmonitored&status=missing`, or `status=monitored&status=unmonitored`
- **THEN** the response SHALL return HTTP 200 with an empty `data` array and `total` of 0, since no entry can simultaneously satisfy both conditions - this is the correct, expected outcome of the logical AND, not an error

#### Scenario: No status filter supplied
- **WHEN** a listing request omits the status filter entirely
- **THEN** the endpoint SHALL return only monitored entries, exactly as it did before this filter existed

#### Scenario: Status filter combined with search and match-status filter
- **WHEN** a listing request supplies a status filter together with a `search` term and/or a match-status filter
- **THEN** the endpoint SHALL apply all supplied constraints together before pagination, returning only entries satisfying every one of them

#### Scenario: Status filter requires no additional upstream calls
- **WHEN** a listing request supplies any combination of status filter values
- **THEN** the endpoint SHALL determine each entry's monitored/unmonitored/missing status entirely from the same Radarr/Sonarr catalog fetch already used to build the unfiltered listing, without issuing additional per-item upstream requests

### Requirement: Sonarr series episode list includes each episode's monitored and file-presence state
The endpoint that returns a Sonarr series' full per-episode list (used to render that series' per-episode breakdown) SHALL include, for each episode, whether it is monitored in Sonarr and whether Sonarr has a file for it, in addition to the existing local-playlist match detail.

#### Scenario: Episode fields reflect Sonarr's own state
- **WHEN** the per-episode list for a Sonarr series is requested
- **THEN** each episode entry SHALL report Sonarr's own `monitored` and `hasFile` values, independent of and in addition to that episode's local-playlist match status

#### Scenario: Missing episode is derivable from the response
- **WHEN** an episode is monitored in Sonarr but Sonarr has no file for it
- **THEN** the episode entry's fields SHALL be sufficient for a client to determine that the episode is Missing, without an additional request
