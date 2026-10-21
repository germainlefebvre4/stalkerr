# Tasks

## 1. Phase 1 — Foundations (tokens, fonts, theme default)

- [ ] 1.1 Add `@fontsource/ibm-plex-sans` (400/500/600) and `@fontsource/ibm-plex-mono` (400/500) to `frontend/package.json`, import latin subsets in `src/main.tsx`; verify `npm run build` emits woff2 files in `dist` and the built CSS contains no `fonts.googleapis.com` reference
- [ ] 1.2 Rewrite `src/variables.css`: new dark values in the default scope, light values in `[data-theme="light"]`, add `--bg-surface-2`, `--font-mono`, LED tokens, radii 4–6px, `--shadow-*` set to `none`, gradient tokens to a flat accent; verify contrast ≥ 4.5:1 for text/secondary/muted/status on their surfaces in both themes with a small script or documented manual check
- [ ] 1.3 Neutralize in `src/index.css` every `box-shadow`, `backdrop-filter`, gradient background and card hover-lift transform, and replace the `rgba(99,102,241…)` literals with tokens; verify `grep -nE "backdrop-filter|linear-gradient|translateY" src/index.css` returns nothing outside intended animations
- [ ] 1.4 Switch the theme default to dark: update the pre-paint script in `index.html` and the fallback in `src/hooks/useTheme.ts`; update/add tests in `useTheme` and `ConfigurationPage.test.tsx` for "no stored preference → dark" and "stored light → light", and verify they pass
- [ ] 1.5 Update the "Apparence" reduce-motion wiring so `[data-reduce-motion="true"]` and `prefers-reduced-motion` disable the toast entrance and LED pulse (no card lift remains); verify with a test that the attribute removes the animations
- [ ] 1.6 Run `npm test`, `npm run lint` and `npm run build`; open the app in both themes and confirm every screen is readable (no light-only literals left) — phase 1 exit check

## 2. Phase 2 — Shell (sidebar, telemetry bar, icons, favicon)

- [ ] 2.1 Create `src/components/Sidebar.tsx` (app mark + name + version, vertical `Tabs.List` with lucide icons for Home, Playlist, Logs, Downloads, Arr, Errors, footer with health LED and settings button), rendered at ≥768px inside the existing `Tabs.Root`; verify with a component test that clicking an entry changes the active tab, writes `stalkeer_active_tab`, and the active entry has `aria-current="page"`
- [ ] 2.2 Reduce `FloatingHeader.tsx` to the mobile title row (title + settings button, no subtitle glass, no emoji) and move the desktop settings entry into the sidebar footer; verify exactly one settings button exists at 1024px and at 480px (update `App.test.tsx`)
- [ ] 2.3 Replace emoji labels in `App.tsx` tab triggers and the mobile bottom bar with lucide icons + labels; verify no emoji remains in navigation markup (test) and that the Settings icon differs from the Logs icon
- [ ] 2.4 Add `src/hooks/useErrorCount.ts` (existing downloads list endpoint, combined problem filter, `limit=1`, total only, 30s poll, paused when document hidden) and a health poll for the sidebar LED; verify with hook tests that the badge shows for total > 0 and is hidden at 0, and confirm the endpoint returns a usable total with `limit=1` (adjust per design D5 if not)
- [ ] 2.5 Implement the grid shell CSS (`220px 1fr` at ≥768px, sticky full-height sidebar, mobile unchanged) and the desktop telemetry bar replacing KPI cards; keep the collapsible KPI block on mobile; verify at 768px, 1280px and 480px that nothing overflows horizontally (manual check with screenshots) and tests for both breakpoints pass
- [ ] 2.6 Replace the emoji favicon in `index.html` with an SVG data-URI mark and reuse the glyph in the sidebar; verify the tab icon renders in the browser
- [ ] 2.7 Run `npm test`, `npm run lint`, `npm run build` — phase 2 exit check

## 3. Phase 3 — Primitives

- [ ] 3.1 Add `StatusLed` component and `.led` styles (dot + label, `data-status`, pulse on active states, steady when motion is reduced); verify with unit tests for each status and for the reduced-motion case
- [ ] 3.2 Migrate existing pastel badge classes/usages (downloads, playlist, logs, errors, Arr) to `StatusLed`/LED styles; verify no `.status-badge`-style pastel class remains in use (grep) and existing tests are updated and passing
- [ ] 3.3 Define dense table styles (`.table-dense`: 1px rules, compact rows, mono numeric/path columns with tabular numerals) and apply them to Playlist, Logs, Errors and Downloads tables; verify visually at 1280px and via snapshot-free assertions that mono class is on size/path/date cells
- [ ] 3.4 Restyle buttons, inputs, selects, tabs inside pages (Radix sub-tabs), toasts, tooltips with flat tokens and a visible accent focus ring; verify keyboard focus ring visibility in both themes (manual check)
- [ ] 3.5 Restyle drawers and dialogs (`.drawer`, `.dialog`, Radix Dialog) with border instead of shadow and small radius; verify no `box-shadow`/`backdrop-filter` remains in overlay CSS (grep) and dialog/drawer tests still pass
- [ ] 3.6 Run `npm test`, `npm run lint`, `npm run build` — phase 3 exit check

## 4. Phase 4 — Screens (inline-style migration and consistency)

- [ ] 4.1 Coordinate with `downloads-hide-cancelled-and-search`: confirm it is finished and archived (or merged) before touching `DownloadsTab.tsx`; verify via `openspec list` that it is no longer in progress
- [ ] 4.2 Replace the hard-coded color literals in `ManualOverrideDialog.tsx`, `ErrorsSidepanel.tsx` and `MediaOccurrenceDrawer.tsx` with tokens; verify `grep -nE "#[0-9a-fA-F]{3,8}\b|rgba?\(" src/components/*.tsx` (excluding tests) returns nothing
- [ ] 4.3 Migrate visual inline styles to classes in `RadarrSonarrTab.tsx` (65 inline), keeping dynamic props-driven values inline; verify `RadarrSonarrTab.test.tsx` passes and the screen is checked in both themes
- [ ] 4.4 Migrate `ErrorsSidepanel.tsx` (64) and `MediaOccurrenceDrawer.tsx` (63) likewise; verify their tests pass and drawers render correctly in both themes and on mobile
- [ ] 4.5 Migrate `DownloadsTab.tsx` and downloads cards (mono for size/rate/path, LEDs, telemetry-consistent density); verify `DownloadsTab.test.tsx` passes and the mobile download card keeps its progressive disclosure
- [ ] 4.6 Migrate Home, Playlist, Logs, Pagination, Settings/Configuration components to the new primitives; verify their tests pass and each screen is checked in both themes at 480px and 1280px
- [ ] 4.7 Remove unused tokens/classes (`--accent-gradient*`, `--glass-*`, `--shadow-*`, `.glass-header`, `.segmented-tabs-*` if unused); verify by grep that nothing references the removed names and `npm run build` is clean
- [ ] 4.8 Add a guard test (or CI grep step in `frontend-ci`) that fails when color literals reappear in `src/components/**/*.tsx`; verify it fails on a deliberately inserted literal and passes otherwise
- [ ] 4.9 Update docs: note the dark default and the design system (tokens, fonts, LED, primitives) in the frontend README/`docs/`; verify the documented `npm run build`/`npm test` commands run as written
- [ ] 4.10 Final integration check: `npm test`, `npm run lint`, `npm run build`, then walk through every section in dark and light at 480px/768px/1280px and run `openspec validate refonte-charte-graphique-ui --strict`

## Workflow follow-up

- Archive the change after review (`/opsx:archive`), which syncs the delta specs into `openspec/specs/`.
