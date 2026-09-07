## MODIFIED Requirements

### Requirement: Résumé sub-tab shows a monitoring summary
The Résumé sub-tab SHALL display, for Radarr, the total number of monitored movies and how many of them are matched vs. unmatched in the local playlist, computed across the full monitored catalog; and, for Sonarr, the total number of monitored series only, without a matched/unmatched breakdown. Each service's card SHALL display that service's existing brand icon alongside a status badge reflecting whether it loaded successfully or reported an error, matching the same status-badge, hero-metric, and progress-bar presentation used by the Home dashboard's Radarr/Sonarr summary card. The Radarr card SHALL render its matched-vs-monitored ratio as a progress bar alongside the numeric counts.

#### Scenario: Radarr summary shows full-catalog matched/unmatched counts
- **WHEN** the user opens the Résumé sub-tab
- **THEN** it SHALL display the total count of Radarr-monitored movies as the card's hero metric, the count of those matched vs. unmatched, a matched-ratio progress bar, and a success-styled status badge alongside the Radarr icon, reflecting the entire monitored catalog

#### Scenario: Sonarr summary shows only a total count
- **WHEN** the user opens the Résumé sub-tab
- **THEN** it SHALL display the total count of Sonarr-monitored series as the card's hero metric, without a matched/unmatched breakdown, alongside a success-styled status badge and the Sonarr icon

#### Scenario: One summary source failing does not block the other
- **WHEN** the Radarr summary data fails to load
- **THEN** the Sonarr summary count SHALL still display normally with its own status badge, and vice versa

#### Scenario: A failing section shows a failure-styled status badge
- **WHEN** either the Radarr or the Sonarr summary section fails to load
- **THEN** that section's card SHALL display a failure-styled status badge alongside its existing error message, without affecting the other section's status badge
