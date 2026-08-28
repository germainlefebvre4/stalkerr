## 1. Component state and markup

- [ ] 1.1 Add a local `useState<boolean>` (default `false`, meaning collapsed) to `StatsKPICards` and verify it does not affect the existing `stats`/`getDownloadSuccessRatio` props contract.
- [ ] 1.2 Add a toggle row (label + chevron, e.g. "📊 Statistiques ▾/▴") above the `.kpi-grid` markup, rendered unconditionally, that flips the boolean on click/tap and verify it renders in the DOM on both mobile and desktop widths via browser dev tools.
- [ ] 1.3 Apply a collapsed/expanded class (e.g. `kpi-grid--collapsed`) to `.kpi-grid` driven by the local state and verify the class toggles correctly on tap in React DevTools.

## 2. Mobile-only CSS

- [ ] 2.1 In `index.css`, inside the existing `@media (max-width: 767.98px)` block, hide `.kpi-grid` when `.kpi-grid--collapsed` is present and show the toggle row, verified by resizing the browser below 768px and confirming only the toggle row is visible by default.
- [ ] 2.2 Style the toggle row to meet the `44px` minimum tappable height and match existing mobile card spacing/typography, verified with a browser dev tools element inspection of computed height.
- [ ] 2.3 Ensure the toggle row is hidden (or inert) at/above the `768px` breakpoint and `.kpi-grid` always renders expanded on desktop regardless of the local state, verified by resizing the browser above 768px and confirming all 4 KPI cards remain visible with no toggle row shown.

## 3. Verification

- [ ] 3.1 Manually verify on a mobile-width viewport: page loads with the KPI grid collapsed on every tab (Playlist, Filters, Logs, Downloads), tapping the toggle row expands the grid in place showing correct stats values, tapping again collapses it back.
- [ ] 3.2 Manually verify that `/api/v1/stats` is still fetched and the values are correct immediately upon first expand (no loading flicker), confirming the fetch is unaffected by the collapsed state.
- [ ] 3.3 Manually verify desktop (>= 768px) is unchanged: KPI grid always visible, no toggle row present.
