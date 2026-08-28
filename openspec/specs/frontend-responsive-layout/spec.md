# frontend-responsive-layout Specification

## Purpose
Defines how the dashboard's navigation, tables, cards, and sidepanel adapt below a mobile viewport width, so that browsing and switching sections never requires horizontal scrolling on a phone.

## Requirements

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

### Requirement: Mobile Download Card Progressive Disclosure

Below the mobile breakpoint, each download card SHALL always show its essential fields — content title and year, status badge, the full `download_path`, a size/progress line, the error message when the download has failed, and the "Déplacer" button when the download is completed — and SHALL collapse all other desktop fields (format, resolution, duration, the year/format/low-quality validation badges, genres) into a per-card expandable section that is hidden by default. This applies uniformly across all download statuses (pending, downloading, retrying, completed, failed). Desktop layout (viewport at or above the mobile breakpoint) is unaffected and continues to show all fields always visible, unchanged.

#### Scenario: Essential fields are always visible on a collapsed mobile download card
- **WHEN** the viewport is narrower than the mobile breakpoint and a download card is in its default (collapsed) state
- **THEN** the frontend SHALL display the content title and year, the status badge, the full `download_path`, the size/progress line, the error message (if the download has failed), and the "Déplacer" button (if the download is completed), without requiring the user to tap the card.

#### Scenario: download_path replaces the raw URL and folder/file box on mobile
- **WHEN** the viewport is narrower than the mobile breakpoint
- **THEN** the frontend SHALL render the download's full `download_path` as a single line in place of the desktop's separate raw URL line and boxed folder-name/file-name display.

#### Scenario: Downloading or retrying items show progress instead of a duplicate size
- **WHEN** the viewport is narrower than the mobile breakpoint and a download's status is "downloading" or "retrying"
- **THEN** the frontend SHALL show only the `downloaded / total (percent)` progress line and progress bar as the card's size/progress line, and SHALL NOT additionally render a separate fixed file-size value alongside it.

#### Scenario: Completed or failed items show a fixed size
- **WHEN** the viewport is narrower than the mobile breakpoint and a download's status is "completed" or "failed"
- **THEN** the frontend SHALL show the download's file size as the card's size/progress line.

#### Scenario: Tapping a card reveals its secondary details
- **WHEN** the viewport is narrower than the mobile breakpoint and the user taps a collapsed download card (or its expand control)
- **THEN** the frontend SHALL expand that card in place to reveal its format, resolution, duration, validation badges (year, format, low quality), and genres, without navigating away from the Downloads list.

#### Scenario: Each card's expanded state is independent
- **WHEN** the viewport is narrower than the mobile breakpoint and the user expands one download card among several displayed
- **THEN** the frontend SHALL leave all other download cards in their current (collapsed or expanded) state, unaffected by that action.

#### Scenario: Desktop download cards are unaffected
- **WHEN** the viewport is at or above the mobile breakpoint
- **THEN** the frontend SHALL continue to render each download card with all fields always visible and no expand/collapse control, unchanged from current desktop behavior.

### Requirement: Sidepanel Mobile Ergonomics

The Playlist details sidepanel SHALL continue to render as a right-anchored drawer on all viewport sizes. Below the mobile breakpoint, the drawer SHALL reduce its internal padding and typography sizing for the narrower width, and its close control SHALL have a tappable area of at least `44px` by `44px`.

#### Scenario: Sidepanel close control is easily tappable on mobile
- **WHEN** the details sidepanel is open on a viewport narrower than the mobile breakpoint
- **THEN** the frontend SHALL render the close control with a tappable area of at least `44px` by `44px`.

### Requirement: Playlist Pagination Bar Is Compact on Mobile

Below the mobile breakpoint, the Playlist pagination bar SHALL render the first/previous/next/last controls (`<<`, `<`, `>`, `>>`) together with a `Page X / Y` text indicator in place of the numbered page-button list, and SHALL NOT render the "go to page" input/button group. The bar SHALL fit within the viewport width without requiring horizontal scrolling, regardless of the total number of pages.

#### Scenario: Compact pagination bar renders on mobile
- **WHEN** the viewport is narrower than the mobile breakpoint, the Playlist view is active, and there is more than one page of results
- **THEN** the frontend SHALL render the first/previous/next/last controls and a `Page X / Y` text indicator, SHALL NOT render individual numbered page buttons or the "go to page" input/button group, and the pagination bar SHALL require no horizontal scrolling to view in full.

#### Scenario: Desktop pagination bar is unchanged
- **WHEN** the viewport is at or above the mobile breakpoint
- **THEN** the frontend SHALL continue to render the full pagination bar as defined by `frontend-ihm-dashboard` (numbered page buttons with ellipsis, first/last jump buttons, and the "go to page" input), unchanged from current desktop behavior.

#### Scenario: Navigation controls remain functional in the compact mobile bar
- **WHEN** the viewport is narrower than the mobile breakpoint and the user taps the next-page (`>`) control while not on the last page
- **THEN** the frontend SHALL navigate to the next page, fetch/display its items, and update the `Page X / Y` indicator to reflect the new current page.
