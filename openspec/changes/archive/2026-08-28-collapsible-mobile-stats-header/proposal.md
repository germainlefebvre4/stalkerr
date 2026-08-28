## Why

On mobile, the Statistics KPI Cards grid (`StatsKPICards`, rendered above the tabs in `App.tsx`) always shows all 4 cards, pushing the active tab's content further down the screen on every page load and on every tab. The user wants this header collapsed by default on mobile, with the numbers available on tap rather than always occupying two rows of screen height.

## What Changes

- On mobile viewports only (< 768px), the Statistics KPI Cards grid SHALL start collapsed, showing a single compact row with a label (e.g. "📊 Statistiques") and a chevron affordance, no numeric values visible.
- Tapping the collapsed row SHALL expand it in place to show the existing 2x2 KPI grid (Playlist total, Movies, TV shows, Success rate) with its current mobile styling (reduced padding/font). Tapping again (or tapping the row while expanded) SHALL collapse it back.
- This collapsed/expanded state is local to the KPI header component, not persisted (resets to collapsed on reload/navigation), and independent of any other UI state.
- Desktop layout (viewport >= 768px) is unchanged: the KPI grid remains always visible, with no toggle control.
- The collapse/expand applies uniformly across all tabs, since the KPI header renders above the tab content on every tab.
- **BREAKING**: none (mobile-only display change; no API or data contract changes). This modifies the existing `frontend-responsive-layout` requirement that mandates KPI cards preserve all desktop information always-visible on mobile — that requirement is superseded by the collapsed-by-default, tap-to-expand behavior described above, consistent with the same progressive-disclosure pattern already adopted for mobile Downloads cards.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `frontend-ihm-dashboard`: The "Statistics KPI Cards" requirement gains mobile-specific display behavior — collapsed by default with a tap-to-expand toggle, in addition to the existing always-visible desktop behavior.
- `frontend-responsive-layout`: The requirement that KPI cards preserve all desktop information always-visible on mobile is superseded for the KPI grid specifically — it becomes collapsed-by-default with an accessible tap target to reveal it, matching the precedent set for mobile Downloads cards.

## Impact

- `frontend/src/components/StatsKPICards.tsx`: add local collapsed/expanded state and a toggle row, rendered mobile-only.
- `frontend/src/index.css`: new mobile-only styles for the collapsed toggle row and expand/collapse behavior of `.kpi-grid` (chevron affordance, transition).
- No backend/API changes.
