## MODIFIED Requirements

### Requirement: Sidepanel shows matched playlist occurrences for a selected item
Selecting a movie or series row SHALL open a sidepanel showing the matched local metadata (when matched) and the list of matching playlist occurrences (at least resolution and pipeline state per occurrence), including already-downloaded occurrences. Each occurrence row SHALL be selectable to open the full media detail drawer for that specific occurrence (see "Selecting an occurrence opens the full media detail drawer").

#### Scenario: Selecting a matched movie
- **WHEN** the user selects a movie that has a playlist match
- **THEN** the sidepanel SHALL display the matched local metadata and every playlist occurrence found for it, regardless of pipeline state

#### Scenario: Selecting an unmatched movie
- **WHEN** the user selects a movie with no playlist match
- **THEN** the sidepanel SHALL clearly indicate no playlist match was found, without displaying an error

#### Scenario: Selecting a series shows per-episode breakdown
- **WHEN** the user selects a series row from the aggregated Séries list
- **THEN** the sidepanel SHALL display the per-episode match detail underlying that series' aggregate ratio, with each monitored episode's season and episode number formatted as zero-padded two-digit numbers separated by a space (e.g. "S01 E01")

## ADDED Requirements

### Requirement: Selecting an occurrence opens the full media detail drawer
Selecting a playlist occurrence row (a movie's occurrence, or one of a series episode's occurrences after expanding it) SHALL open the same media detail drawer used by the Playlist tab for that occurrence - TMDB metadata, pipeline state, M3U provenance, raw line and stream URL, and the force-download action - rather than only the resolution/state summary shown in the occurrences list.

#### Scenario: Opening a movie occurrence's detail
- **WHEN** the user selects an occurrence row in a matched movie's sidepanel
- **THEN** the full media detail drawer SHALL open for that occurrence, including its force-download action if eligible

#### Scenario: Detail drawer stacks beside the occurrences sidepanel on desktop
- **WHEN** the media detail drawer is opened from the occurrences sidepanel on a desktop-width viewport
- **THEN** the detail drawer SHALL appear as an additional panel positioned immediately to the left of the occurrences sidepanel, with both panels visible and the occurrences sidepanel unchanged

#### Scenario: Detail view replaces the sidepanel on mobile
- **WHEN** the media detail drawer is opened from the occurrences sidepanel on a mobile-width viewport
- **THEN** the detail view SHALL replace the occurrences sidepanel's content in place, and the user SHALL be able to return to the occurrences list from it

### Requirement: Séries episode rows expand to reveal their occurrences
Each monitored-episode row in the Séries sidepanel SHALL be expandable to reveal that episode's own playlist occurrences (at least resolution and pipeline state per occurrence), mirroring the Films occurrence list, before any occurrence can be selected to open the full media detail drawer.

#### Scenario: Expanding a matched episode
- **WHEN** the user selects an episode row that has one or more matching playlist occurrences
- **THEN** the row SHALL expand to list each of its occurrences, and each listed occurrence SHALL be selectable to open the full media detail drawer

#### Scenario: Expanding an unmatched episode
- **WHEN** the user selects an episode row that has no matching playlist occurrences
- **THEN** the row SHALL indicate it has no occurrences rather than expanding to an empty or misleading list
