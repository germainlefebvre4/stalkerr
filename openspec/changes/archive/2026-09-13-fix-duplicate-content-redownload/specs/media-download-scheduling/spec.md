## MODIFIED Requirements

### Requirement: A work unit yields at most one claimable stream
For a given `SourceKey` (a single movie, or a single series-season), `BuildStreams` SHALL produce at most one stream per run. A movie, or an individual episode within a series-season, qualifies for tier-1 classification only while the local database has no `downloaded` `ProcessedLine` for it; Radarr's or Sonarr's still-reporting-missing status alone SHALL NOT requalify an already-downloaded movie or episode for tier-1. When a work unit qualifies for both tier-1 (still missing, not yet downloaded locally) and tier-2 (already-downloaded, upgrade-eligible) classification in the same run, the tier-1 stream SHALL take precedence and no tier-2 stream SHALL be built for that same `SourceKey`.

#### Scenario: Movie missing per Radarr but already marked downloaded locally
- **WHEN** a movie is still reported as missing by Radarr, and the local database already has a `downloaded` `ProcessedLine` for it plus another eligible candidate
- **THEN** `BuildStreams` SHALL NOT produce a tier-1 stream for that movie based on Radarr's missing status alone; it SHALL only produce a stream for that movie if the separate tier-2 strictly-better-candidate condition is met

#### Scenario: Series-season missing per Sonarr but already marked downloaded locally
- **WHEN** a series-season stream is built, and one of its episodes already has a `downloaded` `ProcessedLine` locally plus another eligible candidate, while other episodes of the same season are still reported missing by Sonarr
- **THEN** `BuildStreams` SHALL exclude the already-downloaded episode from that season stream's items based on Sonarr's missing status alone, while still including the season's other genuinely-missing episodes

#### Scenario: No destination path is targeted by two independently-scheduled downloads
- **WHEN** streams are claimed and drained in the same run
- **THEN** no two claimable streams SHALL resolve to the same destination path, so a completed download is never silently overwritten by another download of the same work unit later in the same run
