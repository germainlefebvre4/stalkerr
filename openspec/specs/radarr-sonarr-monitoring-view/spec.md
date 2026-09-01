# radarr-sonarr-monitoring-view Specification

## Purpose

Give users a dedicated view of what Radarr and Sonarr currently monitor, and whether the local M3U playlist already has a matching entry for each item, so they no longer have to infer this from CLI output.

## Requirements

### Requirement: Dedicated Radarr/Sonarr monitoring tab
The frontend SHALL provide a tab, separate from the existing Playlist tab, dedicated to displaying Radarr-monitored movies and Sonarr-monitored series with their playlist-match status.

#### Scenario: Tab is reachable from the main navigation
- **WHEN** the user opens the application
- **THEN** a tab for Radarr/Sonarr monitoring SHALL be available alongside the existing tabs (Playlist, etc.)

### Requirement: Films section lists Radarr monitored movies with match status
The tab SHALL include a Films section listing Radarr-monitored movies, each showing at minimum its title, year, and whether it has a matching entry in the local playlist, with pagination controls.

#### Scenario: Matched vs unmatched movies are visually distinguishable
- **WHEN** the Films section is displayed
- **THEN** movies with a playlist match SHALL be visually distinguished from movies without one (e.g. a status badge)

#### Scenario: Paginating the Films section
- **WHEN** the user navigates to another page in the Films section
- **THEN** the section SHALL request and display that page's movies with their match status

### Requirement: Séries section lists Sonarr monitored series with aggregate match status
The tab SHALL include a Séries section listing Sonarr-monitored series, each showing at minimum its title and an aggregate ratio of matched vs. monitored episodes (e.g. "8/12"), with pagination controls. Per-episode detail is not shown in this list view.

#### Scenario: Series aggregate ratio is displayed
- **WHEN** the Séries section is displayed
- **THEN** each series SHALL show the count of monitored episodes matched in the playlist over the total count of monitored episodes

#### Scenario: Fully matched series is distinguishable from partially or unmatched series
- **WHEN** a series has all monitored episodes matched, some matched, or none matched
- **THEN** the three cases SHALL be visually distinguishable from one another

### Requirement: Manual refresh per section
Each section (Films, Séries) SHALL only fetch data from its backend endpoint when the user explicitly triggers a refresh action for that section (initial load counts as one such trigger); the view SHALL NOT poll or auto-refresh in the background.

#### Scenario: Refresh action re-fetches only its own section
- **WHEN** the user triggers the refresh action in the Films section
- **THEN** only the Films section SHALL re-fetch and update; the Séries section SHALL remain unchanged until its own refresh is triggered

#### Scenario: No background polling
- **WHEN** the tab remains open without user interaction
- **THEN** neither section SHALL issue additional fetch requests on its own

### Requirement: Independent error handling per section
A failure to load one section SHALL NOT prevent the other section from loading or displaying its own data.

#### Scenario: Radarr unavailable, Sonarr available
- **WHEN** the Films section's request fails (upstream unavailable or misconfigured)
- **THEN** the Films section SHALL display an error state local to itself while the Séries section loads and displays normally, independent of the Films section's state

#### Scenario: Retry after failure
- **WHEN** a section is in an error state
- **THEN** the user SHALL be able to trigger that section's refresh action again to retry, without reloading the page or affecting the other section

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
