## MODIFIED Requirements

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
