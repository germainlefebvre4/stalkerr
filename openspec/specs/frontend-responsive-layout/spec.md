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

Below the mobile breakpoint, the frontend SHALL render the mobile-eligible section tabs — Home, Playlist, Logs, Downloads, and any other tab not explicitly marked desktop-only (currently "Erreurs" is desktop-only, per capability `downloads-errors-view`) — as a fixed bottom tab bar spanning the full viewport width, with every mobile-eligible tab visible without horizontal scrolling. "Home" SHALL be the first (leftmost) entry. Selecting a tab SHALL switch the active section exactly as the desktop segmented tabs do today, including persisting the active tab in `localStorage`. "Filtres" is no longer a tab on any viewport (its configuration now lives in the Settings drawer, per capability `frontend-filters-management`), so it requires no entry in this bar and no mobile-narrowing fallback.

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

### Requirement: Filter and Download Card Density

Below the mobile breakpoint, filter cards and download cards SHALL reduce their padding and font sizes and ensure all interactive elements (buttons, badges) meet a minimum touch target size of `44px` in height. Filter cards SHALL preserve all information currently shown on desktop. Download cards SHALL show their essential fields always visible (per the Mobile Download Card Progressive Disclosure requirement), with secondary technical detail collapsed by default and reachable per-card via tap.

#### Scenario: Download card remains fully readable and tappable on a narrow viewport
- **WHEN** the viewport is narrower than the mobile breakpoint and the Downloads view is active
- **THEN** the frontend SHALL render each download card with its essential fields (title, year, status badge, `download_path`, size/progress line, error message when failed, and the "Déplacer" button when applicable) visible without horizontal overflow, any action button (e.g. "Déplacer") SHALL have a tappable height of at least `44px`, and the card's expand/collapse control SHALL also meet the `44px` minimum tappable height.

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

The Playlist details sidepanel SHALL continue to render as a right-anchored drawer on all viewport sizes. Below the mobile breakpoint, the drawer SHALL reduce its internal padding and typography sizing for the narrower width, and its close control SHALL have a tappable area of at least `44px` by `44px`. The drawer's own content SHALL NOT require horizontal scrolling to read at any viewport width: any unbreakable technical value (e.g. a hash or identifier) SHALL either wrap or be confined to its own explicitly scrollable sub-container, never forcing the drawer itself to overflow horizontally.

#### Scenario: Sidepanel close control is easily tappable on mobile
- **WHEN** the details sidepanel is open on a viewport narrower than the mobile breakpoint
- **THEN** the frontend SHALL render the close control with a tappable area of at least `44px` by `44px`.

#### Scenario: Sidepanel content never introduces its own horizontal scroll on mobile
- **WHEN** the details sidepanel is open on a viewport narrower than the mobile breakpoint
- **THEN** the frontend SHALL render the panel's content within the viewport width, with no horizontal scrollbar appearing inside the panel regardless of the length of any technical field it displays.

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

### Requirement: Compact Playlist Selector Row on Mobile
Below the mobile breakpoint, the Playlist page's merged selector row (the "All"/"Movies"/"TV Shows" content-type filter buttons together with the Items/Grouped view toggle switch) SHALL reduce the filter buttons' padding and inter-button spacing so the row fits within the viewport width without wrapping to a second line, at viewport widths of `360px` and above, while the toggle switch remains right-aligned within the row.

#### Scenario: Selector row fits on one line on a narrow viewport
- **WHEN** the viewport width is `360px` and the Playlist view is active
- **THEN** the frontend SHALL render the content-type filter buttons and the view toggle switch on a single row, with no wrapping, and the toggle switch positioned at the right edge of the row.

#### Scenario: Desktop row spacing is unchanged
- **WHEN** the viewport is at or above the mobile breakpoint
- **THEN** the frontend SHALL continue to render the content-type filter buttons with their existing (non-reduced) padding and gap.

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
- **THEN** the frontend SHALL continue to render the tab content area with its current card styling (background, border, border-radius, box-shadow, and side padding), unchanged from current desktop behavior.

#### Scenario: A tap-triggered hover state does not reintroduce a box-shadow on mobile
- **WHEN** the viewport is narrower than the mobile breakpoint and the user taps the tab content area, leaving it in a browser-applied `:hover` state
- **THEN** the frontend SHALL NOT display a box-shadow or a hover-lift transform on that tab content area, and this state SHALL remain visually identical to the non-hovered state until the viewport is at or above the mobile breakpoint
