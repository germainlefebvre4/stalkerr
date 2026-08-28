## MODIFIED Requirements

### Requirement: Grouped Media Sub-Tab
The Playlist page SHALL offer a way to switch between the "Items" view and the "Films & Séries" (grouped) view via a single icon-only toggle switch, positioned on the same row as the existing content-type filter buttons (All/Movies/TVShows) and right-aligned within that row, independent of those filter buttons, which continue to apply within either view.

#### Scenario: Switching to the grouped view
- **WHEN** the user activates the toggle switch to its "grouped" position
- **THEN** the system SHALL replace the per-item table with the grouped movie/TV-show table, keeping the currently active content-type, state, TMDB-enrichment, and search filters applied.

#### Scenario: Switching back to the items view
- **WHEN** the user activates the toggle switch to its "items" position while the grouped view is active
- **THEN** the system SHALL restore the existing per-item table with the same filters still applied.

#### Scenario: Toggle switch shares a row with the content-type filters
- **WHEN** the Playlist page is displayed at any viewport width
- **THEN** the toggle switch SHALL render on the same row as the "All"/"Movies"/"TV Shows" content-type filter buttons, right-aligned within that row.
