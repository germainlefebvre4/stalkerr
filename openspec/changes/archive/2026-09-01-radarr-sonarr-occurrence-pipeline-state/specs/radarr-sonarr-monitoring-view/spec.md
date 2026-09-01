## MODIFIED Requirements

### Requirement: Sidepanel shows matched playlist occurrences for a selected item
Selecting a movie or series row SHALL open a sidepanel showing the matched local metadata (when matched) and the list of matching playlist occurrences (at least resolution, processing status, and download status per occurrence, with processing status and download status shown as two separate statuses rather than a single combined pipeline state), including already-downloaded occurrences. Each occurrence row SHALL be selectable to open the full media detail drawer for that specific occurrence (see "Selecting an occurrence opens the full media detail drawer").

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

### Requirement: Séries episode rows expand to reveal their occurrences
Each monitored-episode row in the Séries sidepanel SHALL be expandable to reveal that episode's own playlist occurrences (at least resolution, processing status, and download status per occurrence, with processing status and download status shown as two separate statuses rather than a single combined pipeline state), mirroring the Films occurrence list, before any occurrence can be selected to open the full media detail drawer.

#### Scenario: Expanding a matched episode
- **WHEN** the user selects an episode row that has one or more matching playlist occurrences
- **THEN** the row SHALL expand to list each of its occurrences, and each listed occurrence SHALL be selectable to open the full media detail drawer

#### Scenario: Expanding an unmatched episode
- **WHEN** the user selects an episode row that has no matching playlist occurrences
- **THEN** the row SHALL indicate it has no occurrences rather than expanding to an empty or misleading list

#### Scenario: Expanded episode's occurrence row shows separate processing and download status
- **WHEN** an episode row is expanded to reveal its occurrences
- **THEN** each listed occurrence SHALL show its processing status and its download status as two distinct, independently visible statuses rather than a single combined state
