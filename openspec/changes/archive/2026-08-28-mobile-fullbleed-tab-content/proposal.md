## Why

On mobile, the active tab's content sits inside `.card.tab-panel`: a white box with its own border, rounded corners, and shadow, nested inside `.app-container`'s side padding. Stacking these two layers of inset consumes roughly a fifth of a phone's width in padding and borders before any real content (list rows, table cells) is reached, which is the opposite of what a narrow screen needs. The header and KPI cards can keep their card treatment, but the tab content — the part the user is actually scanning — should reclaim that width by dropping its own extra inset layer, while still lining up with the header and KPI block's edges rather than bleeding past them.

## What Changes

- Below the mobile breakpoint (`< 768px`), the active tab's content area (`.tab-panel`) SHALL no longer render as a card: no background fill distinct from the page, no border, no border-radius, no box-shadow, and no side padding of its own.
- The tab content area's left/right edges SHALL align with the header and KPI block's edges — both are direct children of `.app-container` and share its side padding, so the tab content SHALL NOT add extra side inset (the old card padding) and SHALL NOT bleed past that padding to the viewport edge either.
- Below the mobile breakpoint, a single 1px top border SHALL mark the seam between the KPI block (or its collapsed toggle row) and the tab content area below it, using a border tone visibly distinct from the page background (the existing `--border-color` token is too close to `--bg-app` to read as a divider at this seam).
- This applies uniformly to all four tabs (Playlist, Filters, Logs, Downloads).
- The header and KPI cards keep their current card styling and side gutter, unchanged.
- Item-level cards inside each tab (`mobile-list-card`, `download-card`, `filter-card`) keep their own background, border, and radius, unchanged — only the outer tab-panel wrapper loses its card treatment.
- Desktop layout (`>= 768px`) is unaffected.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `frontend-responsive-layout`: adds a requirement that the mobile tab content area renders without card chrome and without its own side padding, aligned edge-to-edge with the header/KPI block (via the shared `.app-container` padding), with a hairline top divider separating it from the KPI block, instead of nesting inside a bordered/shadowed card.

## Impact

- `frontend/src/index.css`: mobile (`max-width: 767.98px`) rules for `.tab-panel` change from card padding to no side padding + top-divider styling, aligned with the header/KPI block's edges; `--border-color` is not reused for this seam, a more visible tone is introduced instead.
- `frontend/src/variables.css`: likely needs a new token (e.g. `--border-color-strong`) for the divider tone, so it isn't a one-off hardcoded value.
- No component/markup changes expected — `PlaylistTab.tsx`, `FiltersTab.tsx`, `LogsTab.tsx`, `DownloadsTab.tsx` keep their `className="card tab-panel"` Tabs.Content wrapper; the visual change is scoped to the mobile media query.
