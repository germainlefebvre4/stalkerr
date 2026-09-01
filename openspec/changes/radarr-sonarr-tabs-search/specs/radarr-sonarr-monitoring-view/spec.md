## ADDED Requirements

### Requirement: Radarr/Sonarr tab organizes content into three sub-tabs
The Radarr/Sonarr monitoring tab SHALL organize its content into three sub-tabs: Résumé, Radarr, and Sonarr. Only one sub-tab's content SHALL be visible at a time.

#### Scenario: Sub-tabs are reachable within the monitoring tab
- **WHEN** the user opens the Radarr/Sonarr monitoring tab
- **THEN** three sub-tabs SHALL be available: Résumé, Radarr, Sonarr

#### Scenario: Radarr sub-tab shows the Films section
- **WHEN** the user selects the Radarr sub-tab
- **THEN** the Films section (movies table, pagination, search field) SHALL be displayed, and the Séries section and Résumé content SHALL NOT be displayed

#### Scenario: Sonarr sub-tab shows the Séries section
- **WHEN** the user selects the Sonarr sub-tab
- **THEN** the Séries section (series table, pagination, search field) SHALL be displayed, and the Films section and Résumé content SHALL NOT be displayed

### Requirement: Search field filters the Radarr movies table by title
The Radarr sub-tab SHALL provide a search field that filters the movies table by title across the entire monitored catalog, not only the currently loaded page.

#### Scenario: Entering a search term filters results
- **WHEN** the user types a search term into the Radarr search field
- **THEN** the table SHALL update to show only movies whose title matches the term, re-paginated from the first page

#### Scenario: Search term matches nothing
- **WHEN** the search term matches no monitored movie
- **THEN** the table SHALL display an empty-results state rather than an error

#### Scenario: Clearing the search term restores the full list
- **WHEN** the user clears the search field
- **THEN** the table SHALL return to showing the full paginated, unfiltered list

### Requirement: Search field filters the Sonarr series table by title
The Sonarr sub-tab SHALL provide a search field that filters the series table by title across the entire monitored catalog, not only the currently loaded page.

#### Scenario: Entering a search term filters results
- **WHEN** the user types a search term into the Sonarr search field
- **THEN** the table SHALL update to show only series whose title matches the term, re-paginated from the first page

#### Scenario: Search term matches nothing
- **WHEN** the search term matches no monitored series
- **THEN** the table SHALL display an empty-results state rather than an error

#### Scenario: Clearing the search term restores the full list
- **WHEN** the user clears the search field
- **THEN** the table SHALL return to showing the full paginated, unfiltered list

### Requirement: Résumé sub-tab shows a monitoring summary
The Résumé sub-tab SHALL display, for Radarr, the total number of monitored movies and how many of them are matched vs. unmatched in the local playlist, computed across the full monitored catalog; and, for Sonarr, the total number of monitored series only, without a matched/unmatched breakdown.

#### Scenario: Radarr summary shows full-catalog matched/unmatched counts
- **WHEN** the user opens the Résumé sub-tab
- **THEN** it SHALL display the total count of Radarr-monitored movies and the count of those matched vs. unmatched, reflecting the entire monitored catalog

#### Scenario: Sonarr summary shows only a total count
- **WHEN** the user opens the Résumé sub-tab
- **THEN** it SHALL display the total count of Sonarr-monitored series, without a matched/unmatched breakdown

#### Scenario: One summary source failing does not block the other
- **WHEN** the Radarr summary data fails to load
- **THEN** the Sonarr summary count SHALL still display normally, and vice versa
