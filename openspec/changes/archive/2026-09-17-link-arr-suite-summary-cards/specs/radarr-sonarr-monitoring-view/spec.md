## ADDED Requirements

### Requirement: Résumé cards navigate to their detailed sub-tab
The Résumé sub-tab's Radarr card and Sonarr card SHALL each act as a navigation control to their respective detailed sub-tab. Clicking anywhere on the Radarr card SHALL switch the active sub-tab to Radarr. Clicking anywhere on the Sonarr card SHALL switch the active sub-tab to Sonarr. Both cards SHALL present a visual affordance (e.g. hover and cursor feedback) indicating they are interactive.

#### Scenario: Clicking the Films (Radarr) card opens the Radarr sub-tab
- **WHEN** the user is on the Résumé sub-tab and clicks the "Films (Radarr)" card
- **THEN** the active sub-tab switches to Radarr, showing the detailed Films list

#### Scenario: Clicking the Séries (Sonarr) card opens the Sonarr sub-tab
- **WHEN** the user is on the Résumé sub-tab and clicks the "Séries (Sonarr)" card
- **THEN** the active sub-tab switches to Sonarr, showing the detailed Séries list

#### Scenario: Card hover indicates interactivity
- **WHEN** the user hovers over the Radarr or Sonarr card on the Résumé sub-tab
- **THEN** the card shows a visual affordance (e.g. cursor pointer and/or hover styling) indicating it can be clicked
