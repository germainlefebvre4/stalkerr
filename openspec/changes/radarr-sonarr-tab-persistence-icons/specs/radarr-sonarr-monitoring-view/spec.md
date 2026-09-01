## MODIFIED Requirements

### Requirement: Radarr/Sonarr tab organizes content into three sub-tabs
The Radarr/Sonarr monitoring tab SHALL organize its content into three sub-tabs: Résumé, Radarr, and Sonarr. Only one sub-tab's content SHALL be visible at a time. The selected sub-tab SHALL persist across a page refresh. The Radarr and Sonarr sub-tab triggers SHALL each display their respective solution's icon alongside their text label; the Résumé sub-tab trigger SHALL display a text label only.

#### Scenario: Sub-tabs are reachable within the monitoring tab
- **WHEN** the user opens the Radarr/Sonarr monitoring tab
- **THEN** three sub-tabs SHALL be available: Résumé, Radarr, Sonarr

#### Scenario: Radarr sub-tab shows the Films section
- **WHEN** the user selects the Radarr sub-tab
- **THEN** the Films section (movies table, pagination, search field) SHALL be displayed, and the Séries section and Résumé content SHALL NOT be displayed

#### Scenario: Sonarr sub-tab shows the Séries section
- **WHEN** the user selects the Sonarr sub-tab
- **THEN** the Séries section (series table, pagination, search field) SHALL be displayed, and the Films section and Résumé content SHALL NOT be displayed

#### Scenario: Selected sub-tab survives a page refresh
- **WHEN** the user selects the Radarr or Sonarr sub-tab and then refreshes the page
- **THEN** the same sub-tab SHALL remain selected after the page reloads

#### Scenario: Radarr and Sonarr sub-tabs display their solution's icon
- **WHEN** the Radarr/Sonarr monitoring tab is displayed
- **THEN** the Radarr sub-tab trigger SHALL display the Radarr icon next to its label, and the Sonarr sub-tab trigger SHALL display the Sonarr icon next to its label
