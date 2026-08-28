## ADDED Requirements

### Requirement: Compact Playlist Selector Row on Mobile
Below the mobile breakpoint, the Playlist page's merged selector row (the "All"/"Movies"/"TV Shows" content-type filter buttons together with the Items/Grouped view toggle switch) SHALL reduce the filter buttons' padding and inter-button spacing so the row fits within the viewport width without wrapping to a second line, at viewport widths of `360px` and above, while the toggle switch remains right-aligned within the row.

#### Scenario: Selector row fits on one line on a narrow viewport
- **WHEN** the viewport width is `360px` and the Playlist view is active
- **THEN** the frontend SHALL render the content-type filter buttons and the view toggle switch on a single row, with no wrapping, and the toggle switch positioned at the right edge of the row.

#### Scenario: Desktop row spacing is unchanged
- **WHEN** the viewport is at or above the mobile breakpoint
- **THEN** the frontend SHALL continue to render the content-type filter buttons with their existing (non-reduced) padding and gap.
