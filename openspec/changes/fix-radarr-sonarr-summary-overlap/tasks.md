## 1. Fix Résumé summary cards layout

- [ ] 1.1 In `frontend/src/components/RadarrSonarrTab.tsx`, replace the outer `<div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: '1.25rem' }}>` wrapper (Résumé sub-tab, ~line 250) with `<section className="kpi-grid">`, and verify it renders as a `<section>` with class `kpi-grid` in the compiled output.
- [ ] 1.2 Replace the Radarr summary `<div className="table-flush" style={{...}}>` block (~line 251) with `<div className="kpi-card">`, moving the heading into a `kpi-card-title` span and the monitored/matched/unmatched stats into `kpi-card-value` / `kpi-card-subtitle` spans, preserving the existing `statsError` / `stats?.radarr_error` / loading branches and translation keys unchanged, and verify by inspecting the rendered DOM (no more inline `padding`/`border` styles, correct `kpi-card-*` classes present).
- [ ] 1.3 Replace the Sonarr summary `<div className="table-flush" style={{...}}>` block (~line 273) the same way, preserving its existing single-stat (monitored only) content and error/loading branches unchanged.
- [ ] 1.4 Verify `frontend/src/index.css`'s `.table-flush` class (margin-left/right: -2rem on desktop) is no longer applied anywhere in the Résumé sub-tab, and remains applied only to the Films/Séries/occurrence tables (~lines 213, 350, 447), by grepping `table-flush` in `RadarrSonarrTab.tsx` and confirming exactly 3 remaining matches.

## 2. Verify

- [ ] 2.1 Run the frontend test suite (`RadarrSonarrTab.test.tsx` and any snapshot/lint checks) and verify all existing tests pass unmodified.
- [ ] 2.2 Start the frontend dev server, open the Radarr/Sonarr tab's Résumé sub-tab at a desktop width (≥768px), and verify the two cards sit side by side without overlapping, matching the visual style (hover, accent bar) of the KPI cards elsewhere in the app (e.g. Playlist tab).
- [ ] 2.3 Resize to a mobile width (<768px) and verify the two cards stack in the existing 2-column `kpi-grid` mobile layout without overlap or horizontal overflow.
