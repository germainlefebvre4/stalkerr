## MODIFIED Requirements

### Requirement: Quick Settings Search
The Configuration page SHALL provide a search input that filters the settings fields and Filtres/Sources M3U list entries visible on the active tab by matching the query text (case-insensitive substring match) against their translated label, hiding non-matching entries. The search SHALL also match the query against each settings group/card's translated title; when a group's title matches, that entire group SHALL remain visible even if none of its individual field labels match. The search query SHALL persist when the user switches tabs.

#### Scenario: Search filters fields within the active tab
- **WHEN** the user is on the "Intégrations" tab and types "api" into the search input
- **THEN** only fields whose label matches "api" SHALL remain visible on that tab, and other fields SHALL be hidden

#### Scenario: Search matches a group's title
- **WHEN** the user is on the "Intégrations" tab and types "rad" into the search input
- **THEN** the "Radarr" card SHALL remain visible with all of its fields, even though no individual field label contains "rad"

#### Scenario: Clearing the search shows every field again
- **WHEN** the user clears the search input
- **THEN** every field in the active tab SHALL be visible again

#### Scenario: Search query persists across tab switches
- **WHEN** the user has typed a query while on one tab and switches to another tab
- **THEN** the same query SHALL remain in the search input and SHALL filter the newly active tab's visible entries
