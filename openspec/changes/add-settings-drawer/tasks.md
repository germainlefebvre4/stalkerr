## 1. Theme tokens and preference plumbing

- [ ] 1.1 Add a `[data-theme="dark"]` block to `frontend/src/variables.css` with dark equivalents for every existing token (`--bg-*`, `--text-*`, `--status-*`, `--border-*`, `--shadow-*`), and grep `.tsx`/`.css` for hardcoded `#`/`rgb(`/`rgba(` colors outside tokens; route through tokens or note as intentionally theme-invariant. Verify: toggling `data-theme` on `<html>` in devtools re-themes every screen with no unstyled/invisible element.
- [ ] 1.2 Add a `useTheme` hook (or equivalent) that reads/writes `stalkeer_theme` in `localStorage` and sets `document.documentElement.dataset.theme`, defaulting to no attribute (light) when unset. Verify: reloading after selecting dark restores dark without a flash of light content.
- [ ] 1.3 Add a `useReduceMotion` hook that reads/writes `stalkeer_reduce_motion` in `localStorage` and sets a `data-reduce-motion="true"` attribute on `<html>`. Gate the toast's `contentShow` animation in `App.tsx` behind it, and qualify the `.card:hover` transition rule in `index.css` with `:root:not([data-reduce-motion="true"])`. Verify: with the toggle on, triggering a toast shows no entrance animation and hovering a card shows no lift.

## 2. Settings drawer shell

- [ ] 2.1 Create `frontend/src/components/SettingsDrawer.tsx`: a Radix `Dialog.Root`/`Dialog.Content` styled as a right-anchored, full-height sheet (new CSS class in `index.css`, distinct from the existing centered `.dialog-content`), with a close control. Verify: opening/closing via the trigger and the close control works, and closing leaves the currently active tab unchanged.
- [ ] 2.2 Add the six section headers in order (Apparence, Langue, Filtres, Préférences, Système, À propos) inside the drawer, with Filtres and Système as collapsed-by-default disclosures (local `useState<boolean>` per section) and the other four always-expanded. Verify: opening the drawer shows all six in order with Filtres/Système collapsed.
- [ ] 2.3 Replace `FloatingHeader.tsx`'s system-status icon and language `<select>` with a single settings trigger button (icon distinct from the Logs tab's `⚙️`, per `design.md`), on the same header row as the title on both desktop and mobile (verify at `375px` and `1280px` viewport widths). Wire it to open `SettingsDrawer`.

## 3. Apparence and Préférences sections

- [ ] 3.1 Build the Apparence section: theme toggle (light/dark) wired to `useTheme`, and a reduce-animations toggle wired to `useReduceMotion`. Verify: switching themes re-renders immediately; both preferences survive a page reload.
- [ ] 3.2 Build the Préférences section: a startup-tab selector that writes `stalkeer_active_tab` without changing the current session's active tab; a default playlist page-size selector and default playlist-view selector that write `stalkeer_playlist_limit` and the existing playlist-view `localStorage` key, applying live if the Playlist tab is mounted. Verify: setting "Downloads" as startup tab while on Home leaves Home active now, and a subsequent reload with no `tab` URL param opens Downloads.
- [ ] 3.3 Build the À propos section: app name and a link to the source repository. Verify: link opens the repo in a new tab.

## 4. Relocate Système and Langue

- [ ] 4.1 Move `SystemStatusDialog`'s content (service rows, disk usage, refresh action) into the drawer's Système disclosure, fetching on first expand only (reuse `useSystemStatus`), with no fetch before first expand and a fresh fetch on every re-expand. Remove the standalone `SystemStatusDialog` trigger and its `isSystemStatusOpen` state from `App.tsx`. Verify: `system-status-view` spec scenarios (fetch on first expand only, no polling, refresh works, states/disk rows render).
- [ ] 4.2 Move the language `<select>` from `FloatingHeader.tsx` into the drawer's Langue section, preserving `i18next` wiring (`i18n.changeLanguage`, `stalkeer_language` persistence). Verify: switching language inside the drawer re-renders the app and `<html lang>` in the language chosen, and persists across a reload without reopening the drawer.

## 5. Relocate Filtres and remove the tab

- [ ] 5.1 Move `FiltersTab`'s content (filter grid, create/delete flows) into the drawer's Filtres disclosure, fetching via `useFilters` on first expand instead of on `activeTab === 'filters'`. Keep `CreateFilterDialog` wiring unchanged. Verify: `frontend-filters-management` spec scenarios (origin vs. override display, create, delete) all still pass with the section as the entry point.
- [ ] 5.2 Remove `'filters'` from `VALID_TABS` in `App.tsx`, remove the `Tabs.Trigger` for "filters" from the desktop segmented tabs, and drop `'filters'` from the mobile-narrowing fallback effect (it's no longer reachable on any viewport, so the fallback condition simplifies to just `'errors'`). Verify: `?tab=filters` in the URL falls back to Home (existing `isValid`/default behavior), and no "Filtres" entry appears anywhere in the tab bar on desktop or mobile.
- [ ] 5.3 Update the `frontend-responsive-layout`-owned "four tabs" full-bleed styling scope (`index.css`/component) to the three remaining tabs (Playlist, Logs, Downloads); confirm the Settings drawer is unaffected since it is not a `Tabs.Content` panel. Verify: mobile full-bleed styling still applies correctly to Playlist/Logs/Downloads.

## 6. Translations

- [ ] 6.1 Add a `settings` i18n namespace (or extend `common`/`dialogs`) with keys for the drawer's section labels, theme/reduce-motion toggles, and Préférences fields, in both `frontend/src/locales/en/` and `frontend/src/locales/fr/`. Verify: no missing-key warnings in either language with the drawer open and every section expanded.

## 7. Verification pass

- [ ] 7.1 Run the existing frontend test suite (`SystemStatusDialog.test.tsx` and any tests referencing the "filters" tab or `FloatingHeader` will need updating to reflect the new drawer-based structure) and confirm it passes.
- [ ] 7.2 Manually verify in a browser at both a desktop and a `375px`-wide viewport: single settings icon shares the title row on mobile, drawer opens/closes, all six sections behave as specified, dark theme and reduce-motion persist across reload, and no regression in Playlist/Logs/Downloads/Radarr-Sonarr/Errors/Home tabs.
