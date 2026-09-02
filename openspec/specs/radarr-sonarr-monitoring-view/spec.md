# radarr-sonarr-monitoring-view Specification

## Purpose

Give users a dedicated view of what Radarr and Sonarr currently monitor, and whether the local M3U playlist already has a matching entry for each item, so they no longer have to infer this from CLI output.

## Requirements

### Requirement: Dedicated Radarr/Sonarr monitoring tab
The frontend SHALL provide a tab, separate from the existing Playlist tab, dedicated to displaying Radarr-monitored movies and Sonarr-monitored series with their playlist-match status.

#### Scenario: Tab is reachable from the main navigation
- **WHEN** the user opens the application
- **THEN** a tab for Radarr/Sonarr monitoring SHALL be available alongside the existing tabs (Playlist, etc.)

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

### Requirement: Films section lists Radarr monitored movies with match status
The tab SHALL include a Films section listing Radarr-monitored movies, each showing at minimum its title, year, and whether it has a matching entry in the local playlist, with pagination controls.

#### Scenario: Matched vs unmatched movies are visually distinguishable
- **WHEN** the Films section is displayed
- **THEN** movies with a playlist match SHALL be visually distinguished from movies without one (e.g. a status badge)

#### Scenario: Paginating the Films section
- **WHEN** the user navigates to another page in the Films section
- **THEN** the section SHALL request and display that page's movies with their match status

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

### Requirement: Occurrence count column in the Films table
The Films table SHALL include a column showing each movie's total playlist occurrence count, counting every matching playlist entry including duplicates at different resolutions or qualities.

#### Scenario: Movie with multiple quality occurrences
- **WHEN** a matched movie has several playlist occurrences in different resolutions
- **THEN** the Occurrences column SHALL display the total count of those occurrences

#### Scenario: Unmatched movie shows zero occurrences
- **WHEN** a movie has no playlist match
- **THEN** the Occurrences column SHALL display 0

### Requirement: Séries section lists Sonarr monitored series with aggregate match status
The tab SHALL include a Séries section listing Sonarr-monitored series, each showing at minimum its title and an aggregate ratio of matched vs. monitored episodes (e.g. "8/12"), with pagination controls. Per-episode detail is not shown in this list view.

#### Scenario: Series aggregate ratio is displayed
- **WHEN** the Séries section is displayed
- **THEN** each series SHALL show the count of monitored episodes matched in the playlist over the total count of monitored episodes

#### Scenario: Fully matched series is distinguishable from partially or unmatched series
- **WHEN** a series has all monitored episodes matched, some matched, or none matched
- **THEN** the three cases SHALL be visually distinguishable from one another

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

### Requirement: Occurrence count column in the Séries table
The Séries table SHALL include a column showing each series' total playlist occurrence count, aggregated across all of its monitored episodes and counting every matching playlist entry including duplicates at different resolutions or qualities.

#### Scenario: Series with matched episodes having multiple occurrences
- **WHEN** a series has several monitored episodes, each with one or more playlist occurrences
- **THEN** the Occurrences column SHALL display the sum of occurrences across all of that series' monitored episodes

#### Scenario: Series with no matched episodes shows zero occurrences
- **WHEN** a series has zero monitored episodes with a playlist match
- **THEN** the Occurrences column SHALL display 0

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

### Requirement: Selecting an occurrence opens the full media detail drawer
Selecting a playlist occurrence row (a movie's occurrence, or one of a series episode's occurrences after expanding it) SHALL open the same media detail drawer used by the Playlist tab for that occurrence - TMDB metadata, pipeline state, M3U provenance, raw line and stream URL, the "Associate" action, and the force-download action - rather than only the resolution/state summary shown in the occurrences list.

#### Scenario: Opening a movie occurrence's detail
- **WHEN** the user selects an occurrence row in a matched movie's sidepanel
- **THEN** the full media detail drawer SHALL open for that occurrence, including its "Associate" action and its force-download action if eligible

#### Scenario: Detail drawer stacks beside the occurrences sidepanel on desktop
- **WHEN** the media detail drawer is opened from the occurrences sidepanel on a desktop-width viewport
- **THEN** the detail drawer SHALL appear as an additional panel positioned immediately to the left of the occurrences sidepanel, with both panels visible and the occurrences sidepanel unchanged

#### Scenario: Detail view replaces the sidepanel on mobile
- **WHEN** the media detail drawer is opened from the occurrences sidepanel on a mobile-width viewport
- **THEN** the detail view SHALL replace the occurrences sidepanel's content in place, and the user SHALL be able to return to the occurrences list from it

#### Scenario: Associate action available from the Radarr/Sonarr detail drawer
- **WHEN** the user opens the media detail drawer from the Radarr/Sonarr tab's occurrences sidepanel
- **THEN** the "Associate" action SHALL be present and functional, letting the user correct or confirm the occurrence's TMDB match without leaving the Radarr/Sonarr tab

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
