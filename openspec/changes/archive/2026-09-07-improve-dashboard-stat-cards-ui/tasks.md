## 1. Dependency and shared CSS foundation

- [x] 1.1 Add `lucide-react` to `frontend/package.json` and verify `npm install` (or the project's package manager) completes without errors
- [x] 1.2 Add a hero-metric class (e.g. `.home-card-hero`) to `frontend/src/index.css`, distinct from `.home-card-title`/`.home-fields`, with its own desktop and mobile (`@media (max-width: 767.98px)`) font sizes, and verify it renders at a visibly larger size than `.home-fields` text in both breakpoints
- [x] 1.3 Add a per-card icon/status-badge header layout to `.home-card` (icon + title on the left, status badge on the right, wrapping gracefully on narrow widths) and verify no overflow/clipping down to a 320px viewport width
- [x] 1.4 Verify `.status-dot` renders correctly as a mobile replacement for `.badge-success`/`.badge-failed` by temporarily rendering both side-by-side in a scratch page or Storybook-less manual check, then remove the scratch check
- [x] 1.5 Add a disclosure-control style (button/text pattern, `44px` minimum tappable height) for the group-titles "+N autres" control, and verify it meets `44px` height at the mobile breakpoint

## 2. Home dashboard tab (`HomeTab.tsx`)

- [x] 2.1 Restructure the "Last Run" card: status badge (success/failure/in-progress) at top, execution date/time + duration, then movies/TV/new-items/TMDB matched/TMDB unmatched in a compact secondary grid, and verify all four existing loading/empty/unavailable/error states still render correctly
- [x] 2.2 Implement the group-titles disclosure (local `useState`, collapsed by default, "+N autres" label) in the Last Run card, and verify: (a) a short list renders fully with no disclosure, (b) a long list renders truncated with a correct hidden-count label, (c) expanding reveals the full list
- [x] 2.3 Restructure the "Catalog" card: total items as hero metric, movies/TV shows as secondary detail, download success percentage as a progress bar (reusing `Progress.Root`/`Progress.Indicator` + `.progress-root`/`.progress-indicator`) alongside its numeric value, and verify the bar's fill percentage matches `getDownloadSuccessRatio()`
- [x] 2.4 Restructure the "Radarr/Sonarr" card: brand icon + status badge per subsection (success when no error, failure when `radarr_error`/`sonarr_error`/`radarrSonarrStatsError` is set), Radarr matched/monitored ratio as a progress bar, and verify both the all-success and the one-service-errors scenarios render correctly
- [x] 2.5 Restructure the "Downloads & Errors" card: downloads total as hero metric, errors count with a status badge (neutral when zero, failure-styled when greater than zero), and verify both the zero-errors and non-zero-errors cases render the correct badge style
- [x] 2.6 Verify all four cards still pass their existing loading-state checks (`latestLogLoading`, `radarrSonarrStatsLoading`, stats `null` states) without a layout shift once data arrives

## 3. Radarr/Sonarr Résumé sub-tab (`RadarrSonarrTab.tsx`)

- [x] 3.1 Restructure the "Films (Radarr)" summary card to use the same hero-metric/status-badge/progress-bar layout as Home's Radarr/Sonarr card (monitored as hero, matched-ratio progress bar, unmatched as secondary detail, Radarr icon + status badge), and verify the existing error/retry button still works
- [x] 3.2 Restructure the "Séries (Sonarr)" summary card to use the same layout without a progress bar (monitored count as hero metric only, Sonarr icon + status badge), and verify the existing error/retry button still works
- [x] 3.3 Verify the "one section fails, the other still loads" behavior (existing `statsError`/`stats.radarr_error`/`stats.sonarr_error` branches) still renders each card's status badge independently

## 4. Localization

- [x] 4.1 Add any new translation keys used by the group-titles disclosure label (e.g. a "+{{count}} autres" pattern) to `frontend/src/locales/en/home.json` and `frontend/src/locales/fr/home.json`, and verify both locales render without falling back to raw keys
- [x] 4.2 Add any new translation keys used by new status-badge text (if the badge renders text beyond existing `resume.*`/`radarrSonarr.*`/`lastRun.*` strings) to the relevant `home.json` and `radarrSonarr.json` locale files in both `en` and `fr`, and verify no missing-key warnings appear in the console when switching languages

## 5. Cross-page verification

- [x] 5.1 Run the existing frontend test suite (`HomeTab.test.tsx` and any `RadarrSonarrTab` tests) and update assertions that depended on the old flat-text markup, verifying all tests pass
- [x] 5.2 Manually load the Home tab and the Radarr/Sonarr Résumé sub-tab side by side at a desktop width and confirm both use visually identical card chrome (icon placement, badge style, hero metric size, progress-bar style)
- [x] 5.3 Manually resize to a mobile viewport (< 768px) and confirm on both pages: single-column stacking, `.status-dot` replacing full badges, hero metric legible at its own size, and the group-titles disclosure (Home only) meeting the `44px` tap target
- [x] 5.4 Confirm no other component in the codebase regressed by re-running a repo-wide search for `.home-card`/`.home-grid` usages and checking each result still renders as expected
