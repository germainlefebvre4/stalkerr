## MODIFIED Requirements

### Requirement: Radarr/Sonarr tab organizes content into three sub-tabs
The Radarr/Sonarr monitoring tab SHALL organize its content into three sub-tabs: Résumé, Radarr, and Sonarr. Only one sub-tab's content SHALL be visible at a time. The selected sub-tab SHALL persist across a page refresh. The Radarr and Sonarr sub-tab triggers SHALL each display their respective solution's icon alongside their text label; the Résumé sub-tab trigger SHALL display a text label only. The sub-tab switcher SHALL remain reachable on mobile-width viewports, independent of the app's top-level navigation.

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

#### Scenario: Sub-tab switcher is usable on mobile
- **WHEN** the user opens the Radarr/Sonarr monitoring tab on a mobile-width viewport
- **THEN** the user SHALL be able to switch between the Résumé, Radarr, and Sonarr sub-tabs, using a control distinct from the app's bottom tab-bar navigation

### Requirement: Selecting a series shows per-episode breakdown
Selecting a series row from the aggregated Séries list SHALL open a sidepanel displaying the per-episode match detail underlying that series' aggregate ratio, with each monitored episode's season and episode number formatted as zero-padded two-digit numbers separated by a space (e.g. "S01 E01"). Episodes SHALL be grouped by season into collapsible sections, collapsed by default, each showing the matched/total ratio of monitored episodes within that season. At most one season section SHALL be expanded at a time; expanding a season SHALL collapse any other expanded season.

#### Scenario: Selecting a series shows per-episode breakdown grouped by season
- **WHEN** the user selects a series row from the aggregated Séries list
- **THEN** the sidepanel SHALL display the per-episode match detail underlying that series' aggregate ratio, grouped into per-season sections, with each monitored episode's season and episode number formatted as zero-padded two-digit numbers separated by a space (e.g. "S01 E01")

#### Scenario: Season sections are collapsed by default and show their stats
- **WHEN** the sidepanel opens for a selected series
- **THEN** every season section SHALL be collapsed, and each SHALL display the count of monitored episodes matched in the playlist over the total count of monitored episodes for that season

#### Scenario: Expanding a season collapses the previously expanded season
- **WHEN** the user expands a season section while another season section is already expanded
- **THEN** the previously expanded season SHALL collapse and only the newly selected season's episodes SHALL be visible

### Requirement: Séries episode rows expand to reveal their occurrences
Each monitored-episode row in the Séries sidepanel SHALL be expandable to reveal that episode's own playlist occurrences (at least resolution, processing status, and download status per occurrence, with processing status and download status shown as two separate statuses rather than a single combined pipeline state), mirroring the Films occurrence list, before any occurrence can be selected to open the full media detail drawer. At most one episode row SHALL be expanded at a time; expanding an episode SHALL collapse any other expanded episode. An expanded episode row SHALL be visually distinguishable from collapsed episode rows. The occurrences revealed by an expanded episode SHALL be visually distinguishable from the episode rows above and below them, so that an occurrence cannot be mistaken for the next episode row.

#### Scenario: Expanding a matched episode
- **WHEN** the user selects an episode row that has one or more matching playlist occurrences
- **THEN** the row SHALL expand to list each of its occurrences, and each listed occurrence SHALL be selectable to open the full media detail drawer

#### Scenario: Expanding an unmatched episode
- **WHEN** the user selects an episode row that has no matching playlist occurrences
- **THEN** the row SHALL indicate it has no occurrences rather than expanding to an empty or misleading list

#### Scenario: Expanded episode's occurrence row shows separate processing and download status
- **WHEN** an episode row is expanded to reveal its occurrences
- **THEN** each listed occurrence SHALL show its processing status and its download status as two distinct, independently visible statuses rather than a single combined state

#### Scenario: Expanded episode row is visually distinguishable from other rows
- **WHEN** an episode row is expanded
- **THEN** that row SHALL be styled differently from collapsed episode rows, so the user can tell which episode is currently open

#### Scenario: Expanding a different episode collapses the previous one
- **WHEN** the user selects an episode row while another episode row is already expanded
- **THEN** the previously expanded episode SHALL collapse and only the newly selected episode's occurrences SHALL be visible

#### Scenario: Occurrences of an expanded episode are not mistaken for the next episode row
- **WHEN** an episode row is expanded to reveal its occurrences
- **THEN** the occurrences SHALL be visually separated (e.g. indentation, spacing, distinct styling) from both the expanded episode row and the next episode row, rather than sharing the same row styling and sitting directly adjacent to it

### Requirement: Sidepanel shows matched playlist occurrences for a selected item
Selecting a movie or series row SHALL open a sidepanel showing the matched local metadata (when matched) and the list of matching playlist occurrences (at least resolution, processing status, and download status per occurrence, with processing status and download status shown as two separate statuses rather than a single combined pipeline state), including already-downloaded occurrences. Each occurrence row SHALL be selectable to open the full media detail drawer for that specific occurrence (see "Selecting an occurrence opens the full media detail drawer"). On mobile-width viewports, occurrence and episode tables within the sidepanel SHALL render as a mobile-appropriate list layout instead of the fixed-width desktop table.

#### Scenario: Selecting a matched movie
- **WHEN** the user selects a movie that has a playlist match
- **THEN** the sidepanel SHALL display the matched local metadata and every playlist occurrence found for it, regardless of pipeline state

#### Scenario: Selecting an unmatched movie
- **WHEN** the user selects a movie with no playlist match
- **THEN** the sidepanel SHALL clearly indicate no playlist match was found, without displaying an error

#### Scenario: Selecting a series shows per-episode breakdown
- **WHEN** the user selects a series row from the aggregated Séries list
- **THEN** the sidepanel SHALL display the per-episode match detail underlying that series' aggregate ratio, with each monitored episode's season and episode number formatted as zero-padded two-digit numbers separated by a space (e.g. "S01 E01")

#### Scenario: Occurrence row shows separate processing and download status
- **WHEN** a matched movie's occurrence row is displayed in the sidepanel
- **THEN** the row SHALL show its processing status and its download status as two distinct, independently visible statuses rather than a single combined state

#### Scenario: Sidepanel tables adapt to mobile viewports
- **WHEN** the sidepanel is displayed on a mobile-width viewport
- **THEN** its occurrence and episode tables SHALL render as a mobile-appropriate list layout instead of the fixed-width desktop table, without requiring horizontal scrolling to read a row
