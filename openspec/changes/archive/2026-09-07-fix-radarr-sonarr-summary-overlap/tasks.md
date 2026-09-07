## 1. Fix Résumé summary cards layout

- [x] 1.1 In `frontend/src/components/RadarrSonarrTab.tsx`, replace the outer `<div style={{ display: 'grid', ... }}>` wrapper (Résumé sub-tab, ~line 250) with `<section className="home-grid">` — reusing the existing card-grid classes already defined in `index.css` and used by `HomeTab.tsx` (the `kpi-grid`/`kpi-card`/`StatsKPICards.tsx` names originally specified do not exist anywhere in the codebase; see note below).
- [x] 1.2 Replace the Radarr summary `<div className="table-flush" style={{...}}>` block (~line 251) with `<div className="home-card">`, moving the heading into a `home-card-title` element and the monitored/matched/unmatched stats into a `home-fields` wrapper, preserving the existing `statsError` / `stats?.radarr_error` / loading branches and translation keys unchanged.
- [x] 1.3 Replace the Sonarr summary `<div className="table-flush" style={{...}}>` block (~line 273) the same way, preserving its existing single-stat (monitored only) content and error/loading branches unchanged.
- [x] 1.4 Verify `frontend/src/index.css`'s `.table-flush` class (margin-left/right: -2rem on desktop) is no longer applied anywhere in the Résumé sub-tab, and remains applied only to the Films/Séries/occurrence tables (~lines 213, 350, 447), by grepping `table-flush` in `RadarrSonarrTab.tsx` and confirming exactly 3 remaining matches.

> **Note on deviation from original tasks:** the proposal/tasks described restyling with `kpi-grid`/`kpi-card`/`kpi-card-title`/`kpi-card-value`/`kpi-card-subtitle` classes "already defined in `index.css` and used by `StatsKPICards.tsx`". Neither that component nor those classes exist anywhere in the repository. The actual existing analogous pattern — already rendering this same Radarr/Sonarr summary data as side-by-side cards without overlap — is `HomeTab.tsx`'s `.home-grid`/`.home-card`/`.home-card-title`/`.home-card-subheading`/`.home-fields` (defined in `index.css` around lines 81-124, with a mobile override at ~955-972 that stacks to a single column below 768px). Confirmed with the user and implemented using `.home-card` instead of the nonexistent `kpi-card`.

## 2. Verify

- [x] 2.1 Run the frontend test suite (`RadarrSonarrTab.test.tsx` and any snapshot/lint checks) and verify all existing tests pass unmodified.
- [x] 2.2 Start the frontend dev server, open the Radarr/Sonarr tab's Résumé sub-tab at a desktop width (≥768px), and verify the two cards sit side by side without overlapping, matching the visual style of the Home tab's stat cards.
- [x] 2.3 Resize to a mobile width (<768px) and verify the two cards stack in a single column (per `.home-grid`'s mobile override) without overlap or horizontal overflow.
