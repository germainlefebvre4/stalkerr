## MODIFIED Requirements

### Requirement: Responsive Card Density

Below the mobile breakpoint, the KPI cards, filter cards, and download cards SHALL reduce their padding and font sizes and ensure all interactive elements (buttons, badges) meet a minimum touch target size of `44px` in height. Filter cards and download cards SHALL preserve all information currently shown on desktop. The KPI cards grid SHALL instead render collapsed by default behind a single-row tap-to-expand toggle showing a label and a chevron affordance, with no numeric values visible until expanded; tapping the toggle row SHALL reveal the full KPI grid in place, and tapping it again SHALL collapse it back. This collapsed/expanded state SHALL be independent per page load (not persisted across reloads or navigation) and SHALL NOT affect any other UI state.

#### Scenario: Download card remains fully readable and tappable on a narrow viewport
- **WHEN** the viewport is narrower than the mobile breakpoint and the Downloads view is active
- **THEN** the frontend SHALL render each download card with all of its current desktop information (title, technical specs, validation badges, progress) visible without horizontal overflow, and any action button (e.g. "Déplacer") SHALL have a tappable height of at least `44px`.

#### Scenario: KPI cards grid starts collapsed on a narrow viewport
- **WHEN** the viewport is narrower than the mobile breakpoint, on any tab, and the page has just loaded or been navigated to
- **THEN** the frontend SHALL render the KPI cards grid collapsed to a single toggle row with a label and chevron, with none of the 4 KPI cards' numeric values visible, and the toggle row SHALL have a tappable height of at least `44px`.

#### Scenario: Tapping the collapsed KPI toggle expands the grid in place
- **WHEN** the viewport is narrower than the mobile breakpoint and the user taps the collapsed KPI toggle row
- **THEN** the frontend SHALL expand the KPI cards grid in place to show all 4 cards with their current mobile styling, without navigating away from the active tab or losing any other page state.

#### Scenario: Tapping the expanded KPI toggle collapses the grid back
- **WHEN** the viewport is narrower than the mobile breakpoint and the KPI cards grid is currently expanded
- **THEN** tapping the toggle row SHALL collapse the grid back to the single toggle row.

#### Scenario: Desktop KPI grid is unchanged
- **WHEN** the viewport is at or above the mobile breakpoint
- **THEN** the frontend SHALL continue to render the full KPI cards grid always visible, with no collapse toggle.
