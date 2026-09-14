## 1. Shared field primitives

- [x] 1.1 Promote `ToggleSwitch` out of `ConfigurationPage.tsx` into its own exported component and repoint the existing Apparence theme/reduce-motion toggles to it; verify `ConfigurationPage.test.tsx`'s existing toggle assertions still pass.
- [x] 1.2 Add a `readOnly` mode to `SettingsFieldRow` (value + compact origin indicator, no save/pending controls) and delete `BootstrapFieldRow`, switching its call site to the new mode; verify `SettingsFieldRow.test.tsx` covers read-only rendering and no test file references `BootstrapFieldRow` anymore.
- [x] 1.3 Replace `SettingsFieldRow`'s boolean `<select>` with the shared `ToggleSwitch`, remove its per-field Save button, and replace the text "reset to config" button with an icon-only control; verify `SettingsFieldRow.test.tsx` asserts a toggle renders for boolean fields and no per-field Save button remains.
- [x] 1.4 Replace the field's persistent "Interface"/"Config" text badge with a compact indicator that reveals the full label via a Radix `Tooltip` on hover/focus; verify a test asserts the label text is absent until hover/focus.

## 2. Dirty-state group save

- [x] 2.1 Build a group-card component that renders a set of `SettingsFieldRow`s, tracks a pending-changes map keyed by field, and shows a save/discard action bar only while that map is non-empty; verify a new test covers: typing stages a pending change with no API call, confirming submits every pending field, cancelling discards them all.
- [x] 2.2 Wire `IntegrationsSection`, `NotificationsSection`, and Avancé's editable groups (Downloads tuning, Logging, M3U update interval) to render through the group-card component instead of individually Save-buttoned rows; verify each section's existing test file is updated and passes.
- [x] 2.3 Verify the reset-to-config icon still submits immediately and independently of any other field's pending state in the same card, with a test exercising a card that has both a pending edit on one field and an immediate reset on another.

## 3. Tab container and URL/localStorage state

- [x] 3.1 Build the six-tab container for `ConfigurationPage.tsx` using the app's existing Radix `Tabs` primitive, with tabs Général, Intégrations, Contenu, Notifications, Avancé, Système in that order; verify `ConfigurationPage.test.tsx` asserts the six tabs render in order.
- [x] 3.2 Add `"settings"` as a recognized value of the existing `tab` URL query param/`readInitialActiveTab`/`patchTabURLState` mechanism, and add a `settingsTab` param plus a `stalkeer_configuration_tab` localStorage key with precedence URL > localStorage > `"general"`; verify a test covers each precedence level.
- [x] 3.3 Verify selecting a tab updates `settingsTab` in the URL and localStorage, and that loading a URL with a given `settingsTab` opens the Configuration page directly on that tab, with a test that mounts the app at such a URL.
- [x] 3.4 Verify the page's existing "Close"/back affordance still restores the main tab that was active before Configuration was opened, regardless of which Configuration-page tab is active, via updated `App.test.tsx` coverage.

## 4. Reorganize sections into tabs

- [x] 4.1 Move Apparence, Langue, and Préférences into the "Général" tab as three cards using the shared card/grid layout (960px max width, ≤6-field cards packing multi-column); verify `ConfigurationPage.test.tsx` locates them under "Général".
- [x] 4.2 Move `IntegrationsSection` (Radarr/Sonarr/TMDB/Jellyfin) under the "Intégrations" tab, unchanged in field content; verify its existing test still passes under the new container.
- [x] 4.3 Move `NotificationsSection` under the "Notifications" tab; verify its existing test still passes.
- [x] 4.4 Move Downloads tuning, Logging, and M3U update interval (from `AdvancedSection`) under the "Avancé" tab as editable-only groups, applying the >6-fields-spans-full-width rule to Downloads tuning; verify the updated test reflects the new grouping and no longer includes bootstrap fields.
- [x] 4.5 Move the bootstrap read-only fields into the "Système" tab as their own read-only card; verify a test asserts they render under "Système" and not under "Avancé".
- [x] 4.6 Move `SystemStatusSection` (health/disk/build) and the "À propos" name/GitHub link into the "Système" tab alongside the bootstrap card; verify `SystemStatusSection.test.tsx` passes and an "À propos" assertion is present under "Système".
- [x] 4.7 Move `FiltersSection` and `M3uSourcesSection` under the "Contenu" tab; verify their existing test files pass under the new container.
- [x] 4.8 Change `useSystemStatus`'s fetch trigger from "section expanded" to "tab activated" (same on-demand-only semantics, new trigger), and fix its stale "header icon's dialog" code comment; verify "no fetch before first activation" and "reactivating fetches again" with a test.

## 5. Overview summary and search

- [x] 5.1 On Configuration page mount, fetch settings, M3U (effective + origin), and filters (override + origin) data needed to compute override counts, independently of `useSystemStatus` (which stays on-demand-only); verify with a test that these requests fire once on mount regardless of the initially active tab.
- [x] 5.2 Compute and render the summary banner (total override count, distinct restart-required count, and the "no overrides" empty state) from that prefetched data; verify with a test covering all three banner states.
- [x] 5.3 Add the per-tab override-count badge and restart-required marker to the Intégrations, Contenu, Notifications, and Avancé tab labels; verify with a test that a tab with zero overrides shows no badge and a tab with a restart-required override shows the marker.
- [x] 5.4 Add the quick search input, filtering the active tab's visible field rows and Filtres/Sources M3U entries by label (case-insensitive substring), persisting the query across tab switches; verify with a test covering filter-while-typing, clearing, and persistence across a tab switch.

## 6. Backend: filter origin exposure and override replacement

_Added mid-implementation: `frontend-filters-management`'s Filters List View and `filter-override-policy`'s Origin Filter Configuration Exposure / Single Active Runtime Override Per Attribute requirements were never actually implemented server-side. See `design.md`'s "Filters backend gap" for why this is in scope here._

- [x] 6.1 Add a `GET /api/v1/filters/origin` endpoint returning, for each attribute (`group_title`, `tvg_name`), the `config.yml`-defined include/exclude patterns (empty lists when unconfigured); verify with a backend test covering both attributes and the no-patterns-configured case.
- [x] 6.2 Make `createFilter` and `updateFilter` enforce at most one active runtime override per attribute by deleting any existing runtime override for the target attribute before saving the new/updated one; verify with backend tests for "create replaces an existing same-attribute override" and "update onto an attribute with a different existing override replaces it".

## 7. Unify create/edit dialogs

- [x] 7.1 Add `getFilterOrigin()` to `api.ts`; wire `useFilters`/`FiltersSection` to fetch it alongside `GET /api/v1/filters` and render one block per attribute (Group Title, TVG Name) showing the origin patterns (labeled origin/system) and, if present, the active runtime override (labeled as an active override); verify with a new `FiltersSection` test covering both the no-override and has-override cases.
- [x] 7.2 Extract the shared dialog interaction shape (header trigger, Radix `Dialog`, inline replace-warning banner) used by both filter and M3U source creation, refactoring `CreateFilterDialog` and `M3uSourceDialog` onto it — `M3uSourceDialog`'s native `confirm()` replace-check becomes the same inline banner; verify both existing test files pass and a test asserts both use the same warning-banner markup pattern.
- [x] 7.3 Add "Reprendre la config actuelle" to `CreateFilterDialog`, populating Include/Exclude from the active override for the selected attribute if one exists (from the data fetched in 7.1) or from its origin patterns otherwise; add the inline replace-warning banner when the selected attribute already has an active runtime override (detected from the already-fetched `GET /api/v1/filters` data — no extra request needed); verify with a new `CreateFilterDialog` test covering both the load-current-config and replace-warning behaviors.
- [x] 7.4 Verify the "load current configuration into the form" (Filtres) and pre-fill-on-edit (Sources M3U) behaviors both work correctly after the refactor by running their test suites.

## 8. Locale and cleanup

- [x] 8.1 Add/update locale keys in `src/locales/{en,fr}/settings.json` for the six tab labels, the summary banner, per-tab badges, and the search placeholder; verify keys are identical between `en` and `fr`.
- [x] 8.2 Update `src/locales/{en,fr}/dialogs.json`'s `systemStatus.*` keys if their surrounding copy still implies a dialog/drawer rather than a tab; verify no orphaned keys remain.
- [x] 8.3 Remove every remaining "Settings drawer" reference in code comments and user-facing strings (e.g. `useSystemStatus.ts`); verify with a repo-wide search of `frontend/src` that none remain.

## 9. Verification

- [x] 9.1 Run the full frontend test suite and verify it passes.
- [x] 9.2 Run the full backend test suite and verify it passes.
- [x] 9.3 Start the app locally and manually walk through: opening Configuration from the header icon, switching all six tabs, checking the summary banner/badges against actual overrides, staging and saving a group of fields, discarding a pending change, resetting a single overridden field to config, searching for a field, creating/replacing an M3U source and a filter through the unified dialog pattern, and deep-linking directly to a `settingsTab` URL; verify each behaves as specified.
