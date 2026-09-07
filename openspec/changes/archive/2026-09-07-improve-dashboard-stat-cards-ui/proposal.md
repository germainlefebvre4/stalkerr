## Why

The Home dashboard tab and the Radarr/Sonarr "Résumé" sub-tab both use the same `.home-card`/`.home-grid` stat-tile pattern, but it is purely flat text: a title followed by a vertical list of `<strong>value</strong> label` rows. There is no visual hierarchy (every number has the same weight), no color-coded status (beyond red error text), and the "Last Run" card is noticeably more crowded (7+ fields plus a group-title list) than the other cards it sits next to. This makes both pages harder to scan at a glance than they need to be, even though the app already has the visual building blocks (badges, a progress bar, brand icons) to fix this elsewhere.

## What Changes

- Extend the shared stat-card component/classes (`.home-card`, `.home-grid`) used by both the Home dashboard tab and the Radarr/Sonarr Résumé sub-tab with:
  - A dominant "hero" metric per card (largest, most important number), with the remaining fields listed beneath it at a smaller, secondary weight.
  - A per-card icon — the existing Radarr/Sonarr brand icons for those subsections, and a new generic icon (via a new `lucide-react` dependency) for the Catalog, Downloads & Errors, and Last Run cards — plus a status indicator (reusing the existing `.badge-success`/`.badge-failed` classes) reflecting whether that card's data loaded cleanly or hit an error.
  - A progress bar (reusing the existing `.progress-root`/`.progress-indicator` pattern already used in the Downloads view) to visualize matched/monitored and download-success ratios, instead of only stating them as numbers.
- Restructure the "Last Run" card specifically: a status badge and run date/duration at the top, remaining counts (movies, TV shows, new items, TMDB matched/unmatched) arranged in a compact secondary grid, and the `group_titles` list collapsed behind a "+N autres" disclosure when it is long, instead of always printing the full comma-joined list.
- On mobile viewports, replace the full status badge with the existing compact `.status-dot` indicator to save horizontal space, and give the hero metric its own responsive type scale distinct from the uniform `.home-fields` text size; any new interactive element (the group-titles disclosure) meets the existing `44px` minimum touch target.
- Add `lucide-react` as a new frontend dependency for the generic (non-brand) card icons.

This is a shared-component change: because both pages already consume the exact same CSS classes, updating them once means the Home dashboard and the Radarr/Sonarr Résumé sub-tab evolve together and stay visually consistent.

## Capabilities

### New Capabilities
_None — this extends the presentation of two existing capabilities without introducing a new one._

### Modified Capabilities
- `frontend-home-dashboard`: the Last Run, Catalog, Radarr/Sonarr, and Downloads & Errors summary requirements gain hero-metric/status-badge/progress-bar presentation rules; the Last Run summary gains a group-titles disclosure behavior; the Home Tab Responsive Layout requirement gains mobile-specific rules for the compact status indicator and the hero-metric type scale.
- `radarr-sonarr-monitoring-view`: the "Résumé sub-tab shows a monitoring summary" requirement gains the same hero-metric/icon/status-badge/progress-bar presentation rules, so the sub-tab stays visually consistent with the Home dashboard's stat cards.

## Impact

- Frontend only, no backend/API changes — all underlying data is already fetched by existing endpoints; this changes presentation and adds one new client-only interaction (the group-titles disclosure).
- `frontend/src/index.css` and `frontend/src/variables.css`: new/updated stat-card classes (hero metric, per-card icon slot, status indicator, disclosure).
- `frontend/src/components/HomeTab.tsx` and `frontend/src/components/RadarrSonarrTab.tsx`: updated markup to use the restructured stat-card pattern.
- `frontend/package.json`: new dependency, `lucide-react`.
