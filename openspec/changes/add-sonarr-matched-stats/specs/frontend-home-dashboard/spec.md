## MODIFIED Requirements

### Requirement: Radarr/Sonarr Summary
The Home tab SHALL display a summary of Radarr and Sonarr monitoring, fetched from `GET /api/v1/radarr-sonarr/stats`: the Radarr monitored count, the Radarr matched count, the Sonarr monitored count, and the Sonarr matched count. When either service reports an error (unreachable or not configured), the Home tab SHALL display that service's error state instead of a missing or zeroed count, without preventing the other service's summary from rendering. Each service's subsection SHALL display that service's existing brand icon alongside a status badge reflecting whether it loaded successfully or reported an error, and both the Radarr and Sonarr subsections SHALL render their matched-vs-monitored ratio as a progress bar alongside the numeric counts.

#### Scenario: Both services report counts
- **WHEN** `GET /api/v1/radarr-sonarr/stats` returns non-null `radarr_monitored`, `radarr_matched`, `sonarr_monitored`, and `sonarr_matched` values
- **THEN** the Home tab SHALL render the Radarr monitored/matched counts with a matched-ratio progress bar, the Sonarr monitored/matched counts with a matched-ratio progress bar, and a success-styled status badge alongside each service's brand icon

#### Scenario: One service is unreachable while the other succeeds
- **WHEN** `GET /api/v1/radarr-sonarr/stats` returns a `radarr_error` of `radarr_unreachable` alongside valid `sonarr_monitored` and `sonarr_matched` values
- **THEN** the Home tab SHALL display the Radarr error state with a failure-styled status badge while still rendering the Sonarr monitored/matched counts and progress bar with a success-styled status badge
