## ADDED Requirements

### Requirement: Desktop Sidebar Navigation
At or above the mobile breakpoint, the frontend SHALL render the section navigation as a fixed left sidebar listing every section (Home, Playlist, Logs, Downloads, Arr, Errors), each with an icon and a text label, and SHALL NOT render a top header with segmented tab pills. Selecting a section SHALL switch it and persist the active tab exactly as the mobile bottom tab bar does.

#### Scenario: Sidebar replaces header and tab pills on desktop
- **WHEN** the viewport is at or above the mobile breakpoint
- **THEN** the frontend SHALL render a left sidebar containing all six sections and SHALL NOT render segmented tab pills or a glass header

#### Scenario: Selecting a sidebar entry switches section and persists the choice
- **WHEN** the user clicks the "Downloads" entry in the sidebar
- **THEN** the frontend SHALL activate the Downloads section and write `"downloads"` to `localStorage` under the same key used by the mobile bottom tab bar

#### Scenario: Active section is marked
- **WHEN** a section is active
- **THEN** its sidebar entry SHALL be visually distinguished from the others and expose `aria-current="page"`

### Requirement: Sidebar Identity and Health Footer
The desktop sidebar SHALL show the application name and version at its top, and, at its bottom, an LED indicating whether the backend API is reachable (with a text label) above the settings entry point.

#### Scenario: Sidebar shows name and version
- **WHEN** the sidebar is displayed
- **THEN** its top SHALL show the application name and the application version

#### Scenario: Footer shows API health
- **WHEN** the backend health endpoint answers successfully
- **THEN** the sidebar footer SHALL show a green LED labelled as healthy, and a failed LED with a label when the endpoint cannot be reached

### Requirement: Sidebar Error Counter
The sidebar entry for the Errors section SHALL display a numeric badge with the current number of errors when that number is greater than zero, and SHALL NOT display a badge when it is zero.

#### Scenario: Badge shown when errors exist
- **WHEN** the backend reports 2 errors
- **THEN** the Errors entry in the sidebar SHALL display a badge reading "2"

#### Scenario: No badge when there are no errors
- **WHEN** the backend reports 0 errors
- **THEN** the Errors entry in the sidebar SHALL display no badge

### Requirement: Desktop Telemetry Bar
At or above the mobile breakpoint, the key indicators previously shown as KPI cards SHALL be rendered as a single thin telemetry bar at the top of the content area, as inline label/value pairs in monospace, without card styling. The same indicators SHALL remain available on mobile as the existing collapsible KPI block.

#### Scenario: Indicators render as a single bar on desktop
- **WHEN** the viewport is at or above the mobile breakpoint
- **THEN** the frontend SHALL render the indicators as label/value pairs in one horizontal bar, with no per-indicator card background, border or shadow

#### Scenario: Mobile keeps the collapsible KPI block
- **WHEN** the viewport is narrower than the mobile breakpoint
- **THEN** the frontend SHALL render the collapsible KPI block as before, using the mission control tokens

## MODIFIED Requirements

### Requirement: Mobile Breakpoint

The frontend SHALL treat any viewport with a width strictly below `768px` as "mobile" and apply the mobile layout rules defined by this capability. Viewports at or above `768px` SHALL use the desktop layout (sidebar navigation, telemetry bar and full `<table>` views).

#### Scenario: Viewport narrower than the breakpoint uses the mobile layout
- **WHEN** the browser viewport width is `480px`
- **THEN** the frontend SHALL render the mobile bottom tab bar, card-based Playlist/Logs views, and mobile card/sidepanel styling defined by this capability.

#### Scenario: Viewport at or above the breakpoint uses the desktop layout
- **WHEN** the browser viewport width is `1024px`
- **THEN** the frontend SHALL render the desktop sidebar navigation and the full `<table>` layout for Playlist and Logs, with the same columns and behavior as before.

### Requirement: Mobile Bottom Tab Navigation

Below the mobile breakpoint, the frontend SHALL render the mobile-eligible section tabs — Home, Playlist, Logs, Downloads, and any other tab not explicitly marked desktop-only (currently "Erreurs" is desktop-only, per capability `downloads-errors-view`) — as a fixed bottom tab bar spanning the full viewport width, with every mobile-eligible tab visible without horizontal scrolling. "Home" SHALL be the first (leftmost) entry. Selecting a tab SHALL switch the active section exactly as the desktop sidebar does, including persisting the active tab in `localStorage`. "Filtres" is no longer a tab on any viewport (its configuration now lives in the Settings drawer, per capability `frontend-filters-management`), so it requires no entry in this bar and no mobile-narrowing fallback.

#### Scenario: All four tabs are reachable without horizontal scrolling
- **WHEN** the viewport is narrower than the mobile breakpoint
- **THEN** the frontend SHALL display all mobile-eligible tab entries (Home, Playlist, Logs, Downloads), with "Home" first, simultaneously in the bottom tab bar, each individually tappable, with no horizontal overflow or scroll required to reach any of them

#### Scenario: Filtres is absent from the mobile bottom tab bar
- **WHEN** the viewport is narrower than the mobile breakpoint
- **THEN** the frontend SHALL NOT show a "Filtres" entry in the mobile bottom tab bar

#### Scenario: Narrowing the viewport while Filtres is active falls back
- **WHEN** the viewport narrows below the mobile breakpoint
- **THEN** the frontend SHALL NOT need to fall back away from a "Filtres" tab, since "Filtres" is never an active tab on any viewport — its configuration now lives in the Settings drawer instead of the tab bar

#### Scenario: Selecting a bottom tab switches section and persists the choice
- **WHEN** the user taps the "Downloads" entry in the mobile bottom tab bar
- **THEN** the frontend SHALL activate the Downloads section and write `"downloads"` to `localStorage` under the same key used by the desktop sidebar.

### Requirement: Full-Bleed Tab Content on Mobile

Below the mobile breakpoint, the active tab's content area SHALL render without card styling: no background fill distinct from the page background, no border, no border-radius, no box-shadow, and no side padding of its own. Its left and right edges SHALL align with the header and KPI block's edges (both share `.app-container`'s side padding) — the tab content SHALL NOT add an extra layer of side inset beyond that padding, and SHALL NOT bleed past it to the viewport edge either. A single 1px top border SHALL mark the seam between the KPI block (in either its collapsed toggle-row or expanded grid state) and the tab content area below it, using a border tone that is visibly distinct from the page background. This requirement applies uniformly to the three remaining tabs (Playlist, Logs, Downloads); it does not apply to the Settings drawer, which is a separate overlay surface rather than a `Tabs.Content` panel. It does not affect the header, the KPI cards, or any item-level card (`mobile-list-card`, `download-card`, `filter-card`), which keep their existing card styling and side gutter unchanged. The "no box-shadow" guarantee SHALL hold at all times, including while the tab content area is in a browser-applied `:hover` state left stuck by a touch tap — no hover-triggered styling SHALL reintroduce a box-shadow or a lift effect on a touch device.

#### Scenario: Tab content aligns with the header and KPI block's edges on mobile
- **WHEN** the viewport is narrower than the mobile breakpoint and any of the three remaining tabs is active
- **THEN** the frontend SHALL render that tab's content area with no visible background distinct from the page, no border, no border-radius, no box-shadow, and no side padding of its own, so its left/right edges line up with the header and KPI block's edges instead of being inset further or bleeding past them.

#### Scenario: A hairline divider separates the KPI block from the content below it
- **WHEN** the viewport is narrower than the mobile breakpoint
- **THEN** the frontend SHALL render a 1px top border at the seam between the KPI block (collapsed or expanded) and the tab content area, in a tone visibly distinct from the page background, rather than reusing a border tone indistinguishable from it.

#### Scenario: Item-level cards are unaffected
- **WHEN** the viewport is narrower than the mobile breakpoint
- **THEN** the frontend SHALL continue to render `mobile-list-card`, `download-card`, and `filter-card` elements with their own background, border, and border-radius, unchanged.

#### Scenario: Desktop tab content is unchanged
- **WHEN** the viewport is at or above the mobile breakpoint
- **THEN** the frontend SHALL keep rendering the tab content area as a bordered panel with side padding, restyled per capability `frontend-design-system` (1px border, small radius, no box-shadow).

#### Scenario: A tap-triggered hover state does not reintroduce a box-shadow on mobile
- **WHEN** the viewport is narrower than the mobile breakpoint and the user taps the tab content area, leaving it in a browser-applied `:hover` state
- **THEN** the frontend SHALL NOT display a box-shadow or a hover-lift transform on that tab content area, and this state SHALL remain visually identical to the non-hovered state until the viewport is at or above the mobile breakpoint
