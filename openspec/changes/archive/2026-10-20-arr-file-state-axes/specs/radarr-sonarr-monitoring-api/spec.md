# Spec Delta

## ADDED Requirements

### Requirement: File state is exposed as an axis independent of monitored state
Every Radarr movie entry, Sonarr series entry, and Sonarr episode entry returned by the monitoring endpoints SHALL carry a `file_state` of `downloaded`, `partial`, or `none`, independent of its `monitored` flag. For movies and episodes it is `downloaded` when Radarr/Sonarr has a file and `none` otherwise (`partial` never applies). For series it is derived from downloaded vs. expected episodes (see "Series file state counts only aired episodes").

#### Scenario: Movie with a file
- **WHEN** Radarr reports a movie with `hasFile` true
- **THEN** its entry SHALL report `file_state` `downloaded`, whether the movie is monitored or not

#### Scenario: Movie without a file
- **WHEN** Radarr reports a movie with `hasFile` false
- **THEN** its entry SHALL report `file_state` `none`, and SHALL NOT report `partial`

#### Scenario: Episode file state
- **WHEN** the per-episode list of a Sonarr series is requested
- **THEN** each episode entry SHALL report `file_state` `downloaded` when Sonarr has a file for it and `none` otherwise, independent of its `monitored` flag

#### Scenario: Unmonitored item that has its file
- **WHEN** an unmonitored movie or series has all its files on disk
- **THEN** its entry SHALL report `monitored` false and `file_state` `downloaded`

### Requirement: Series file state counts only aired episodes
A Sonarr series entry's `file_state` SHALL be `downloaded` when every episode that has already aired has a file, `none` when no episode has a file, and `partial` otherwise. Episodes that have not aired yet SHALL NOT count against the series. The entry SHALL also expose the number of downloaded episodes and the number of aired episodes so a client can display a ratio.

#### Scenario: Continuing series with all aired episodes downloaded
- **WHEN** a series has aired episodes all present on disk and further episodes not yet aired
- **THEN** its entry SHALL report `file_state` `downloaded`

#### Scenario: Some aired episodes missing
- **WHEN** a series has a file for some but not all of its aired episodes
- **THEN** its entry SHALL report `file_state` `partial` and expose the downloaded and aired episode counts (for example 18 and 28)

#### Scenario: No episode downloaded
- **WHEN** a series has aired episodes but none has a file
- **THEN** its entry SHALL report `file_state` `none`

#### Scenario: Series with no aired episodes
- **WHEN** a series has no aired episode yet
- **THEN** its entry SHALL report `file_state` `none` without error

### Requirement: Optional file filter for both listing endpoints
The Radarr movies and Sonarr series listing endpoints SHALL each accept an optional, repeatable `file` parameter whose values are `downloaded`, `partial`, and `none`, applied to the full catalog before pagination. An entry is included when its `file_state` equals any supplied value (logical OR). When the parameter is omitted, no file-state restriction SHALL apply.

#### Scenario: Single file value
- **WHEN** a listing request supplies `file=downloaded`
- **THEN** the response SHALL include only entries whose `file_state` is `downloaded`

#### Scenario: Multiple file values combine as OR
- **WHEN** a listing request supplies `file=none&file=partial`
- **THEN** the response SHALL include entries whose `file_state` is `none` or `partial`

#### Scenario: File filter omitted
- **WHEN** a listing request omits the `file` parameter
- **THEN** entries SHALL NOT be restricted by file state

#### Scenario: File filter combined with the other filters
- **WHEN** a listing request supplies `file` together with `status`, `search`, and/or a match-status filter
- **THEN** the endpoint SHALL return only entries satisfying every supplied filter, applied before pagination

#### Scenario: File filter requires no additional upstream calls
- **WHEN** a listing request supplies any `file` value
- **THEN** the endpoint SHALL determine each entry's file state from the same catalog fetch already used for the unfiltered listing, without per-item upstream requests

### Requirement: Optional monitored-status filter for both listing endpoints
The Radarr movies and Sonarr series listing endpoints SHALL each accept an optional, repeatable `status` parameter with values `monitored` and `unmonitored`, applied to the full catalog before pagination. Supplied values combine as logical OR. When omitted, the endpoint SHALL behave as if only `monitored` were supplied, preserving today's default. The former `missing` value is no longer recognized.

#### Scenario: Single value narrows to that status
- **WHEN** a listing request supplies `status=unmonitored`
- **THEN** the response SHALL include only unmonitored entries

#### Scenario: Both values return everything
- **WHEN** a listing request supplies `status=monitored&status=unmonitored`
- **THEN** the response SHALL include monitored and unmonitored entries alike

#### Scenario: No status filter supplied
- **WHEN** a listing request omits the status filter entirely
- **THEN** the endpoint SHALL return only monitored entries, exactly as it did before this filter existed

#### Scenario: Retired missing value is ignored
- **WHEN** a listing request supplies `status=missing`
- **THEN** the value SHALL be ignored like any other unrecognized value, and the endpoint SHALL behave as if the parameter had been omitted, returning HTTP 200

#### Scenario: Missing view is expressed through both axes
- **WHEN** a listing request supplies `status=monitored&file=none&file=partial`
- **THEN** the response SHALL include only monitored entries whose file state is `none` or `partial`

#### Scenario: Status filter combined with search and match-status filter
- **WHEN** a listing request supplies a status filter together with a `search` term and/or a match-status filter
- **THEN** the endpoint SHALL apply all supplied constraints together before pagination, returning only entries satisfying every one of them

#### Scenario: Status filter requires no additional upstream calls
- **WHEN** a listing request supplies any combination of status filter values
- **THEN** the endpoint SHALL determine each entry's monitored status entirely from the same Radarr/Sonarr catalog fetch already used to build the unfiltered listing, without issuing additional per-item upstream requests

## MODIFIED Requirements

### Requirement: List Radarr monitored movies with playlist match status
The system SHALL expose a paginated endpoint that lists Radarr movies matching the requested status filter (see "Optional monitored-status filter for both listing endpoints") and file filter (see "Optional file filter for both listing endpoints"), each annotated with whether a matching entry exists in the local playlist database, with Radarr's own `monitored` state, and with its `file_state`. When no status filter is supplied, the endpoint SHALL list only movies currently monitored in Radarr, exactly as it did before the status filter existed.

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

#### Scenario: Entry reports no missing flag
- **WHEN** a Radarr movie entry is returned
- **THEN** it SHALL report `monitored` and `file_state` and SHALL NOT report a separate `missing` flag

### Requirement: List Sonarr monitored series with per-series playlist match aggregate
The system SHALL expose a paginated endpoint that lists Sonarr series matching the requested status filter (see "Optional monitored-status filter for both listing endpoints") and file filter (see "Optional file filter for both listing endpoints"), each annotated with an aggregate count of how many of its monitored episodes have a matching entry in the local playlist database (e.g. "8 of 12 monitored episodes matched"), with Sonarr's own `monitored` state, and with its `file_state` and episode counts (see "Series file state counts only aired episodes"). When no status filter is supplied, the endpoint SHALL list only series currently monitored in Sonarr, exactly as it did before the status filter existed.

#### Scenario: Series with partial playlist matches
- **WHEN** a Sonarr-monitored series has N monitored episodes and M of them (M <= N) have a matching local `TVShow` record
- **THEN** the response entry for that series SHALL report matched count M and monitored count N

#### Scenario: Series with no monitored episodes
- **WHEN** a Sonarr-monitored series has zero monitored episodes
- **THEN** the response entry SHALL report a matched count of 0 and a monitored count of 0, without error

#### Scenario: No status filter supplied preserves today's behavior
- **WHEN** a Sonarr series listing request omits the status filter
- **THEN** the endpoint SHALL return only series currently monitored in Sonarr, identical in scope to its behavior before this capability's status filter existed

#### Scenario: Entry reports no missing flag
- **WHEN** a Sonarr series entry is returned
- **THEN** it SHALL report `monitored`, `file_state` and the downloaded/aired episode counts, and SHALL NOT report a separate `missing` flag

### Requirement: Sonarr series episode list includes each episode's monitored and file-presence state
The endpoint that returns a Sonarr series' full per-episode list (used to render that series' per-episode breakdown) SHALL include, for each episode, whether it is monitored in Sonarr and its `file_state`, in addition to the existing local-playlist match detail.

#### Scenario: Episode fields reflect Sonarr's own state
- **WHEN** the per-episode list for a Sonarr series is requested
- **THEN** each episode entry SHALL report Sonarr's own `monitored` value and a `file_state` derived from whether Sonarr has a file for it, independent of and in addition to that episode's local-playlist match status

#### Scenario: Missing episode is derivable from the response
- **WHEN** an episode is monitored in Sonarr but Sonarr has no file for it
- **THEN** the episode entry SHALL report `monitored` true and `file_state` `none`, which is sufficient for a client to treat it as missing without an additional request

## REMOVED Requirements

### Requirement: Optional status filter for both listing endpoints
**Reason**: The status filter mixed monitoring and file presence (`missing`) and combined values with AND; monitoring and file state are now independent axes.
**Migration**: Use "Optional monitored-status filter for both listing endpoints" for monitored/unmonitored and "Optional file filter for both listing endpoints" for file presence; the former `status=missing` is `status=monitored&file=none&file=partial`.
