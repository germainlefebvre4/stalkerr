## MODIFIED Requirements

### Requirement: Mobile Bottom Tab Navigation

Below the mobile breakpoint, the frontend SHALL render the mobile-eligible section tabs — Home, Playlist, Logs, Downloads, and any other tab not explicitly marked desktop-only (currently "Filtres" and "Erreurs" are desktop-only; "Erreurs" per capability `downloads-errors-view`) — as a fixed bottom tab bar spanning the full viewport width, with every mobile-eligible tab visible without horizontal scrolling. "Home" SHALL be the first (leftmost) entry. Selecting a tab SHALL switch the active section exactly as the desktop segmented tabs do today, including persisting the active tab in `localStorage`. If "Filtres" is the active tab and the viewport narrows below the mobile breakpoint, the frontend SHALL fall back to another mobile-eligible tab instead of continuing to render "Filtres" or a related empty/broken layout, mirroring the existing "Erreurs" fallback behavior.

#### Scenario: All four tabs are reachable without horizontal scrolling
- **WHEN** the viewport is narrower than the mobile breakpoint
- **THEN** the frontend SHALL display all mobile-eligible tab entries, with "Home" first, simultaneously in the bottom tab bar, each individually tappable, with no horizontal overflow or scroll required to reach any of them

#### Scenario: Filtres is absent from the mobile bottom tab bar
- **WHEN** the viewport is narrower than the mobile breakpoint
- **THEN** the frontend SHALL NOT show a "Filtres" entry in the mobile bottom tab bar

#### Scenario: Narrowing the viewport while Filtres is active falls back
- **WHEN** the "Filtres" tab is currently active and the viewport is resized to below the mobile breakpoint
- **THEN** the frontend SHALL switch the active tab to another mobile-eligible tab instead of continuing to render the Filtres tab

#### Scenario: Selecting a bottom tab switches section and persists the choice
- **WHEN** the user taps the "Downloads" entry in the mobile bottom tab bar
- **THEN** the frontend SHALL activate the Downloads section and write `"downloads"` to `localStorage` under the same key used by the desktop tabs.

## REMOVED Requirements

### Requirement: Responsive Card Density
**Reason**: The KPI cards grid this requirement described no longer exists — it was removed together with the KPI banner (capability `frontend-ihm-dashboard`'s REMOVED "Statistics KPI Cards" requirement, superseded by capability `frontend-home-dashboard`). The filter-card and download-card density rules this requirement also covered are unaffected by that removal and are re-added below as "Filter and Download Card Density", scoped to the cards that remain.
**Migration**: No user-facing change to filter cards or download cards. Any code or design reference to the KPI cards grid's collapse/expand toggle should be removed; filter-card and download-card density behavior is unchanged, now specified in the "Filter and Download Card Density" requirement added below.

Below the mobile breakpoint, the KPI cards, filter cards, and download cards SHALL reduce their padding and font sizes and ensure all interactive elements (buttons, badges) meet a minimum touch target size of `44px` in height. Filter cards SHALL preserve all information currently shown on desktop. Download cards SHALL show their essential fields always visible (per the Mobile Download Card Progressive Disclosure requirement), with secondary technical detail collapsed by default and reachable per-card via tap. The KPI cards grid SHALL instead render collapsed by default behind a single-row tap-to-expand toggle showing a label and a chevron affordance, with no numeric values visible until expanded; tapping the toggle row SHALL reveal the full KPI grid in place, and tapping it again SHALL collapse it back. This collapsed/expanded state SHALL be independent per page load (not persisted across reloads or navigation) and SHALL NOT affect any other UI state.

#### Scenario: Download card remains fully readable and tappable on a narrow viewport
- **WHEN** the viewport is narrower than the mobile breakpoint and the Downloads view is active
- **THEN** the frontend SHALL render each download card with its essential fields (title, year, status badge, `download_path`, size/progress line, error message when failed, and the "Déplacer" button when applicable) visible without horizontal overflow, any action button (e.g. "Déplacer") SHALL have a tappable height of at least `44px`, and the card's expand/collapse control SHALL also meet the `44px` minimum tappable height.

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

## ADDED Requirements

### Requirement: Filter and Download Card Density
Below the mobile breakpoint, filter cards and download cards SHALL reduce their padding and font sizes and ensure all interactive elements (buttons, badges) meet a minimum touch target size of `44px` in height. Filter cards SHALL preserve all information currently shown on desktop. Download cards SHALL show their essential fields always visible (per the Mobile Download Card Progressive Disclosure requirement), with secondary technical detail collapsed by default and reachable per-card via tap.

#### Scenario: Download card remains fully readable and tappable on a narrow viewport
- **WHEN** the viewport is narrower than the mobile breakpoint and the Downloads view is active
- **THEN** the frontend SHALL render each download card with its essential fields (title, year, status badge, `download_path`, size/progress line, error message when failed, and the "Déplacer" button when applicable) visible without horizontal overflow, any action button (e.g. "Déplacer") SHALL have a tappable height of at least `44px`, and the card's expand/collapse control SHALL also meet the `44px` minimum tappable height.
