## Purpose

Defines how the dashboard's navigation, tables, cards, and sidepanel adapt below a mobile viewport width, so that browsing and switching sections never requires horizontal scrolling on a phone.

## ADDED Requirements

### Requirement: Mobile Breakpoint

The frontend SHALL treat any viewport with a width strictly below `768px` as "mobile" and apply the mobile layout rules defined by this capability. Viewports at or above `768px` SHALL continue to use the existing desktop layout unchanged.

#### Scenario: Viewport narrower than the breakpoint uses the mobile layout
- **WHEN** the browser viewport width is `480px`
- **THEN** the frontend SHALL render the mobile bottom tab bar, card-based Playlist/Logs views, and mobile card/sidepanel styling defined by this capability.

#### Scenario: Viewport at or above the breakpoint uses the desktop layout
- **WHEN** the browser viewport width is `1024px`
- **THEN** the frontend SHALL render the existing segmented tab pills and the full `<table>` layout for Playlist and Logs, unchanged from current desktop behavior.

### Requirement: Mobile Bottom Tab Navigation

Below the mobile breakpoint, the frontend SHALL render the four section tabs (Playlist, Filters, Logs, Downloads) as a fixed bottom tab bar spanning the full viewport width, with all four tabs visible without scrolling. Selecting a tab SHALL switch the active section exactly as the desktop segmented tabs do today, including persisting the active tab in `localStorage`.

#### Scenario: All four tabs are reachable without horizontal scrolling
- **WHEN** the viewport is narrower than the mobile breakpoint
- **THEN** the frontend SHALL display all four tab entries (Playlist, Filters, Logs, Downloads) simultaneously in the bottom tab bar, each individually tappable, with no horizontal overflow or scroll required to reach any of them.

#### Scenario: Selecting a bottom tab switches section and persists the choice
- **WHEN** the user taps the "Downloads" entry in the mobile bottom tab bar
- **THEN** the frontend SHALL activate the Downloads section and write `"downloads"` to `localStorage` under the same key used by the desktop tabs.

### Requirement: Playlist and Logs Render as Cards on Mobile

Below the mobile breakpoint, the Playlist and Logs views SHALL render their list of items as a vertically stacked list of cards (one card per item) instead of a `<table>`, and SHALL NOT require horizontal scrolling to read any item's primary information. Each card SHALL show the item's primary identifying text and its status/state badge. Tapping a card SHALL open the same details sidepanel that clicking a table row opens on desktop, unchanged in content.

#### Scenario: Playlist items render as cards without horizontal scroll on mobile
- **WHEN** the viewport is narrower than the mobile breakpoint and the Playlist view is active
- **THEN** the frontend SHALL render each playlist item as a stacked card showing its media name and pipeline state badge, with no horizontal scroll container required to view a card's content.

#### Scenario: Tapping a mobile Playlist card opens the existing details sidepanel
- **WHEN** the user taps a Playlist card on a mobile viewport
- **THEN** the frontend SHALL open the same details sidepanel (drawer) that opens when clicking a row on desktop, with identical content.

#### Scenario: Logs render as cards without horizontal scroll on mobile
- **WHEN** the viewport is narrower than the mobile breakpoint and the Logs view is active
- **THEN** the frontend SHALL render each log entry as a stacked card showing its action and status badge, with no horizontal scroll container required to view a card's content.

### Requirement: Full-Width Table Rendering on Desktop

At or above the mobile breakpoint, the Playlist and Logs tables SHALL render using the full available content width of the viewport, rather than being constrained by the page container's and tab panel's standard padding, so that more column content is visible without introducing horizontal scrolling under normal column content lengths.

#### Scenario: Table spans the available width on desktop
- **WHEN** the viewport is at or above the mobile breakpoint and the Playlist view is active
- **THEN** the frontend SHALL render the Playlist table extending to the full width available within the viewport, rather than being inset by the tab panel's default padding.

### Requirement: Responsive Card Density

Below the mobile breakpoint, the KPI cards, filter cards, and download cards SHALL reduce their padding and font sizes and ensure all interactive elements (buttons, badges) meet a minimum touch target size of `44px` in height, while preserving all information currently shown on desktop.

#### Scenario: Download card remains fully readable and tappable on a narrow viewport
- **WHEN** the viewport is narrower than the mobile breakpoint and the Downloads view is active
- **THEN** the frontend SHALL render each download card with all of its current desktop information (title, technical specs, validation badges, progress) visible without horizontal overflow, and any action button (e.g. "Déplacer") SHALL have a tappable height of at least `44px`.

### Requirement: Sidepanel Mobile Ergonomics

The Playlist details sidepanel SHALL continue to render as a right-anchored drawer on all viewport sizes. Below the mobile breakpoint, the drawer SHALL reduce its internal padding and typography sizing for the narrower width, and its close control SHALL have a tappable area of at least `44px` by `44px`.

#### Scenario: Sidepanel close control is easily tappable on mobile
- **WHEN** the details sidepanel is open on a viewport narrower than the mobile breakpoint
- **THEN** the frontend SHALL render the close control with a tappable area of at least `44px` by `44px`.
