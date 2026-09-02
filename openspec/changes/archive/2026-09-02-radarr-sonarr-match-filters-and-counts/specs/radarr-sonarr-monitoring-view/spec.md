## ADDED Requirements

### Requirement: Match-status filter control on the Films section
The Radarr sub-tab SHALL provide a match-status filter control (All / Matched / No match) that filters the Films table by playlist match status across the entire monitored catalog, applied before pagination. The selected filter SHALL persist across a page refresh, following the same URL-persistence pattern already used for the sub-tab selection.

#### Scenario: Filtering to unmatched movies
- **WHEN** the user selects "No match" in the Films filter control
- **THEN** the table SHALL update to show only movies with no playlist match, re-paginated from the first page

#### Scenario: Filtering to matched movies
- **WHEN** the user selects "Matched" in the Films filter control
- **THEN** the table SHALL update to show only movies with a playlist match, re-paginated from the first page

#### Scenario: Filter persists across a page refresh
- **WHEN** the user selects a match-status filter and then refreshes the page
- **THEN** the same filter SHALL remain selected and applied after the page reloads

#### Scenario: Filter combined with search
- **WHEN** the user has both a search term and a match-status filter active
- **THEN** the table SHALL show only movies satisfying both constraints

#### Scenario: Selecting "All" restores the unfiltered list
- **WHEN** the user selects "All" in the Films filter control
- **THEN** the table SHALL return to showing every monitored movie regardless of match status

### Requirement: Match-status filter control on the Séries section
The Sonarr sub-tab SHALL provide a match-status filter control (All / Matched / No match) that filters the Séries table by whether a series has at least one matched monitored episode, applied before pagination. The selected filter SHALL persist across a page refresh, following the same URL-persistence pattern already used for the sub-tab selection.

#### Scenario: Filtering to series with no matched episodes
- **WHEN** the user selects "No match" in the Séries filter control
- **THEN** the table SHALL update to show only series with zero matched monitored episodes, re-paginated from the first page

#### Scenario: Filtering to series with at least one matched episode
- **WHEN** the user selects "Matched" in the Séries filter control
- **THEN** the table SHALL update to show only series with at least one matched monitored episode, re-paginated from the first page

#### Scenario: Filter persists across a page refresh
- **WHEN** the user selects a match-status filter and then refreshes the page
- **THEN** the same filter SHALL remain selected and applied after the page reloads

#### Scenario: Filter combined with search
- **WHEN** the user has both a search term and a match-status filter active
- **THEN** the table SHALL show only series satisfying both constraints

#### Scenario: Selecting "All" restores the unfiltered list
- **WHEN** the user selects "All" in the Séries filter control
- **THEN** the table SHALL return to showing every monitored series regardless of match status

### Requirement: Occurrence count column in the Films table
The Films table SHALL include a column showing each movie's total playlist occurrence count, counting every matching playlist entry including duplicates at different resolutions or qualities.

#### Scenario: Movie with multiple quality occurrences
- **WHEN** a matched movie has several playlist occurrences in different resolutions
- **THEN** the Occurrences column SHALL display the total count of those occurrences

#### Scenario: Unmatched movie shows zero occurrences
- **WHEN** a movie has no playlist match
- **THEN** the Occurrences column SHALL display 0

### Requirement: Occurrence count column in the Séries table
The Séries table SHALL include a column showing each series' total playlist occurrence count, aggregated across all of its monitored episodes and counting every matching playlist entry including duplicates at different resolutions or qualities.

#### Scenario: Series with matched episodes having multiple occurrences
- **WHEN** a series has several monitored episodes, each with one or more playlist occurrences
- **THEN** the Occurrences column SHALL display the sum of occurrences across all of that series' monitored episodes

#### Scenario: Series with no matched episodes shows zero occurrences
- **WHEN** a series has zero monitored episodes with a playlist match
- **THEN** the Occurrences column SHALL display 0
