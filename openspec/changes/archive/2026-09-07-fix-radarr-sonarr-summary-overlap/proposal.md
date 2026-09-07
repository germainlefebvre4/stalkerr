## Why

On the Radarr/Sonarr monitoring tab, the Résumé sub-tab's two summary cards (Films/Radarr and Séries/Sonarr) visually overlap on desktop. Both cards reuse the `table-flush` CSS class, which applies `margin-left/right: -2rem` on viewports ≥768px so that a full-width `<table>` can bleed flush to its parent card's edges. That negative margin was copy-pasted onto these two summary cards even though they are side-by-side grid cells, not full-width tables, so each card's edges overlap into its neighbor's space instead of bleeding into the surrounding `.tab-panel` padding.

## What Changes

- Remove the `table-flush` class from the two Résumé summary card containers in `RadarrSonarrTab.tsx` (it is only appropriate for the full-width occurrence/films/series tables, where it stays unchanged).
- Restyle those two cards using the existing `kpi-grid` / `kpi-card` / `kpi-card-title` / `kpi-card-value` / `kpi-card-subtitle` classes (already defined in `index.css` and used by `StatsKPICards.tsx`) instead of ad-hoc inline styles, so the Résumé tab matches the app's established stat-card visual language (hover state, accent bar, responsive mobile grid) and stops duplicating layout rules inline.
- Preserve existing loading/error/data content and copy exactly: Radarr shows monitored/matched/unmatched, Sonarr shows monitored only — no behavior or requirement change, purely visual.

## Capabilities

This is a pure visual/styling fix: the Résumé sub-tab's documented behavior (`radarr-sonarr-monitoring-view`, "Résumé sub-tab shows a monitoring summary") is unchanged — same data, same fields, same sub-tab structure. No capability specs are added or modified (`skip_specs: true` is set for this change).

### New Capabilities
(none)

### Modified Capabilities
(none)

## Impact

- `frontend/src/components/RadarrSonarrTab.tsx` (Résumé sub-tab JSX, lines ~249-291)
- No CSS changes needed: reuses existing `.kpi-grid` / `.kpi-card` rules in `frontend/src/index.css`
- No API, schema, or backend impact
