## ADDED Requirements

### Requirement: A work unit yields at most one claimable stream
For a given `SourceKey` (a single movie, or a single series-season), `BuildStreams` SHALL produce at most one stream per run. When a work unit qualifies for both tier-1 (still missing) and tier-2 (already-downloaded, upgrade-eligible) classification in the same run, the tier-1 stream SHALL take precedence and no tier-2 stream SHALL be built for that same `SourceKey`.

#### Scenario: Movie missing per Radarr but already marked downloaded locally
- **WHEN** a movie is still reported as missing by Radarr, and the local database already has a `downloaded` `ProcessedLine` for it plus another eligible candidate
- **THEN** `BuildStreams` SHALL produce only the tier-1 stream for that movie, and SHALL NOT also produce a tier-2 stream for it

#### Scenario: Series-season missing per Sonarr but already marked downloaded locally
- **WHEN** a series-season has at least one episode still reported as missing by Sonarr, and the local database already has a `downloaded` `ProcessedLine` for another episode of that same season plus an eligible candidate
- **THEN** `BuildStreams` SHALL produce only the tier-1 stream for that series-season, and SHALL NOT also produce a tier-2 stream for it

#### Scenario: No destination path is targeted by two independently-scheduled downloads
- **WHEN** streams are claimed and drained in the same run
- **THEN** no two claimable streams SHALL resolve to the same destination path, so a completed download is never silently overwritten by another download of the same work unit later in the same run
