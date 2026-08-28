## ADDED Requirements

### Requirement: Full-Bleed Tab Content on Mobile

Below the mobile breakpoint, the active tab's content area SHALL render without card styling: no background fill distinct from the page background, no border, no border-radius, no box-shadow, and no side padding of its own. Its left and right edges SHALL align with the header and KPI block's edges (both share `.app-container`'s side padding) — the tab content SHALL NOT add an extra layer of side inset beyond that padding, and SHALL NOT bleed past it to the viewport edge either. A single 1px top border SHALL mark the seam between the KPI block (in either its collapsed toggle-row or expanded grid state) and the tab content area below it, using a border tone that is visibly distinct from the page background. This requirement applies uniformly to all four tabs (Playlist, Filters, Logs, Downloads). It does not affect the header, the KPI cards, or any item-level card (`mobile-list-card`, `download-card`, `filter-card`), which keep their existing card styling and side gutter unchanged.

#### Scenario: Tab content aligns with the header and KPI block's edges on mobile
- **WHEN** the viewport is narrower than the mobile breakpoint and any tab is active
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
