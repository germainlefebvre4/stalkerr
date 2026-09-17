## MODIFIED Requirements

### Requirement: Films section lists Radarr monitored movies with match status
The tab SHALL include a Films section listing Radarr movies matching the active État filter (Monitored, by default - see "État status filter control on the Films section"), each showing at minimum its title, year, whether it has a matching entry in the local playlist, and its État (Monitored, Unmonitored, or Missing), with pagination controls.

#### Scenario: Matched vs unmatched movies are visually distinguishable
- **WHEN** the Films section is displayed
- **THEN** movies with a playlist match SHALL be visually distinguished from movies without one (e.g. a status badge)

#### Scenario: Monitored, Unmonitored, and Missing movies are visually distinguishable
- **WHEN** the Films section is displayed
- **THEN** a movie's État (Monitored, Unmonitored, or Missing) SHALL be shown via a badge distinct from the existing matched/unmatched badge, and the three États SHALL be visually distinguishable from one another

#### Scenario: Paginating the Films section
- **WHEN** the user navigates to another page in the Films section
- **THEN** the section SHALL request and display that page's movies with their match status

### Requirement: Séries section lists Sonarr monitored series with aggregate match status
The tab SHALL include a Séries section listing Sonarr series matching the active État filter (Monitored, by default - see "État status filter control on the Séries section"), each showing at minimum its title, an aggregate ratio of matched vs. monitored episodes (e.g. "8/12"), and its own État (Monitored, Unmonitored, or Missing), with pagination controls. Per-episode detail is not shown in this list view.

#### Scenario: Series aggregate ratio is displayed
- **WHEN** the Séries section is displayed
- **THEN** each series SHALL show the count of monitored episodes matched in the playlist over the total count of monitored episodes

#### Scenario: Fully matched series is distinguishable from partially or unmatched series
- **WHEN** a series has all monitored episodes matched, some matched, or none matched
- **THEN** the three cases SHALL be visually distinguishable from one another

#### Scenario: Monitored, Unmonitored, and Missing series are visually distinguishable
- **WHEN** the Séries section is displayed
- **THEN** a series' État (Monitored, Unmonitored, or Missing) SHALL be shown via a badge distinct from the existing matched-ratio badge, and the three États SHALL be visually distinguishable from one another

## ADDED Requirements

### Requirement: État status filter control on the Films section
The Radarr sub-tab SHALL provide an État filter control, separate from the existing match-status filter, offering three cumulative (multi-select) values - Monitored, Unmonitored, Missing - that filters the Films table by Radarr's own monitored/file-presence state across the entire catalog, applied before pagination. Multiple selected values SHALL combine as a logical AND, matching the backend filter's semantics (see the `radarr-sonarr-monitoring-api` capability). The control SHALL default to Monitored selected on first load. The selected combination SHALL persist across a page refresh, following the same URL-persistence pattern already used for the existing match-status filter.

#### Scenario: Default selection preserves today's view
- **WHEN** the user opens the Films section without having chosen an État filter before
- **THEN** only Monitored SHALL be selected, and the table SHALL show the same set of movies it would have shown before this filter existed

#### Scenario: Selecting Missing narrows the table
- **WHEN** the user selects only "Missing" in the État filter
- **THEN** the table SHALL update to show only movies that are monitored and missing a file, re-paginated from the first page

#### Scenario: Selecting Unmonitored reveals movies not shown before this feature
- **WHEN** the user selects only "Unmonitored" in the État filter
- **THEN** the table SHALL update to show only unmonitored movies, which were not previously retrievable through this table

#### Scenario: Selecting a contradictory combination shows an empty table
- **WHEN** the user selects both "Unmonitored" and "Missing" (or both "Monitored" and "Unmonitored")
- **THEN** the table SHALL display its empty-results state, since no movie can satisfy both conditions at once - this is expected, not an error

#### Scenario: État filter combines with the existing match-status filter
- **WHEN** the user has both an État selection and a match-status selection active
- **THEN** the table SHALL show only movies satisfying both filters together

#### Scenario: Selection persists across a page refresh
- **WHEN** the user changes the État filter selection and then refreshes the page
- **THEN** the same combination SHALL remain selected and applied after the page reloads

### Requirement: État status filter control on the Séries section
The Sonarr sub-tab SHALL provide an État filter control, separate from the existing match-status filter, offering three cumulative (multi-select) values - Monitored, Unmonitored, Missing - that filters the Séries table by Sonarr's own monitored/file-presence state across the entire catalog, applied before pagination. Multiple selected values SHALL combine as a logical AND, matching the backend filter's semantics (see the `radarr-sonarr-monitoring-api` capability). The control SHALL default to Monitored selected on first load. The selected combination SHALL persist across a page refresh, following the same URL-persistence pattern already used for the existing match-status filter.

#### Scenario: Default selection preserves today's view
- **WHEN** the user opens the Séries section without having chosen an État filter before
- **THEN** only Monitored SHALL be selected, and the table SHALL show the same set of series it would have shown before this filter existed

#### Scenario: Selecting Missing narrows the table
- **WHEN** the user selects only "Missing" in the État filter
- **THEN** the table SHALL update to show only series that are monitored with at least one monitored episode missing a file, re-paginated from the first page

#### Scenario: Selecting Unmonitored reveals series not shown before this feature
- **WHEN** the user selects only "Unmonitored" in the État filter
- **THEN** the table SHALL update to show only unmonitored series, which were not previously retrievable through this table

#### Scenario: Selecting a contradictory combination shows an empty table
- **WHEN** the user selects both "Unmonitored" and "Missing" (or both "Monitored" and "Unmonitored")
- **THEN** the table SHALL display its empty-results state, since no series can satisfy both conditions at once - this is expected, not an error

#### Scenario: État filter combines with the existing match-status filter
- **WHEN** the user has both an État selection and a match-status selection active
- **THEN** the table SHALL show only series satisfying both filters together

#### Scenario: Selection persists across a page refresh
- **WHEN** the user changes the État filter selection and then refreshes the page
- **THEN** the same combination SHALL remain selected and applied after the page reloads

### Requirement: Sidepanel header shows the selected movie's or series' État
The sidepanel opened for a selected movie or series SHALL display that item's État (Monitored, Unmonitored, or Missing) next to its title/year header, using the same badge convention as the table.

#### Scenario: Sidepanel header shows État for a movie
- **WHEN** the user selects a movie row and its sidepanel opens
- **THEN** the header SHALL display that movie's État alongside its title and year

#### Scenario: Sidepanel header shows État for a series
- **WHEN** the user selects a series row and its sidepanel opens
- **THEN** the header SHALL display that series' État alongside its title and year

### Requirement: Episode rows display their own État alongside their matched/unmatched status
Each monitored-episode row in the Séries sidepanel's per-episode breakdown SHALL display its own État (Monitored or Missing, derived from Sonarr's own per-episode monitored/file-presence state), in addition to its existing matched/unmatched playlist badge.

#### Scenario: Missing episode is visually distinguishable
- **WHEN** an episode row is displayed for an episode that is monitored in Sonarr but has no file there
- **THEN** the row SHALL display a Missing État badge distinct from its matched/unmatched playlist badge

#### Scenario: Monitored, complete episode shows no Missing indicator
- **WHEN** an episode row is displayed for an episode that is monitored and has a file in Sonarr
- **THEN** the row's État badge SHALL indicate Monitored without a Missing indication

### Requirement: État badge renders as a compact status dot on mobile viewports
On mobile-width viewports, every État badge introduced by this capability (table row, mobile card, sidepanel header, episode row) SHALL render as a small colored status dot with a short label, following the same compact-indicator convention already used elsewhere in this tab's mobile summary cards, instead of the full-size text badge used on desktop.

#### Scenario: Mobile card shows a compact État indicator
- **WHEN** the Films or Séries section is displayed on a mobile-width viewport
- **THEN** each mobile list card SHALL show its État as a compact colored dot with a short label, alongside its existing matched/unmatched badge, without the card's layout overflowing or wrapping awkwardly

#### Scenario: Mobile episode row shows a compact État indicator
- **WHEN** the Séries sidepanel's per-episode breakdown is displayed on a mobile-width viewport
- **THEN** each episode row SHALL show its État as a compact colored dot with a short label, alongside its existing matched/unmatched badge
