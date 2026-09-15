## 1. Page layout alignment

- [x] 1.1 In `ConfigurationPage.tsx`, wrap the header row, `.configuration-summary-banner`, `.settings-search-input-wrap`, `Tabs.List`, and `Tabs.Content` panels in a single container element and give it `max-width: 960px; margin: 0 auto` in `index.css`; remove the now-redundant individual `max-width`/`margin: 0 auto` from `.configuration-summary-banner`, `.settings-search-input-wrap`, and `.settings-tab-panel`. Verify by inspecting the rendered page at a viewport wider than 960px: the header title/close button, tab bar, search input, and tab content all share the same left/right edges.
- [x] 1.2 Verify `ConfigurationPage.test.tsx` still passes unchanged (layout-only change, no behavioral assertions expected to break); run `npm test` in `frontend/`.

## 2. Quick search matches group titles

- [x] 2.1 In `SettingsGroupCard.tsx`, compute a `titleMatches` boolean (`title` present and matching `query`) and use the unfiltered `fields` array (skip the label filter) whenever `titleMatches` is true, per the "Search matches a group's title" scenario in `frontend-configuration-page`. Verify with a new test in `SettingsGroupCard.test.tsx`: rendering with `title="Radarr"`, unrelated field labels, and `searchQuery="rad"` shows all of the group's fields.
- [x] 2.2 Confirm the existing "Search filters fields within the active tab" behavior (label-only match, title not matching) is unchanged by running `IntegrationsSection.test.tsx` and `AdvancedSection.test.tsx`.

## 3. Enumerated select field control for log levels

- [x] 3.1 Add `'select'` to `SettingsGroupFieldSpec['type']` and an optional `options: { value: string; labelKey: string }[]` in `SettingsGroupCard.tsx`.
- [x] 3.2 In `SettingsFieldRow.tsx`, render a `<select>` populated from `options` (each option's `labelKey` passed through `t()`) when `type === 'select'`, alongside the existing `boolean`/text/number branches. Verify with a new test in `SettingsFieldRow.test.tsx` covering rendering the options and staging a selected value as a pending change via `onChange`.
- [x] 3.3 In `settingsFieldGroups.ts`, define the four-level option list (`debug`, `info`, `warn`, `error` → `Debug`/`Info`/`Warning`/`Erreur` label keys) once and set `type: 'select'` plus that `options` list on `logging.app.level` and `logging.database.level` in `LOGGING_FIELDS`.
- [x] 3.4 Add the four label keys to `frontend/src/locales/en/settings.json` and `fr/settings.json` (e.g. under `logLevels.debug/info/warn/error`). Verify `AdvancedSection.test.tsx` renders a `<select>` with the four options for both log level fields instead of a text input.

## 4. M3U source card: hide absent auth password

- [x] 4.1 In `M3uSourcesSection.tsx`, render the auth-password line only `{source.has_auth_password && (...)}` instead of always rendering it with a "Not set" fallback, per the "Auth password status shown only when set" scenario. Verify with `M3uSourcesSection.test.tsx`: a source with `has_auth_password: false` renders no auth-password text; one with `has_auth_password: true` renders the "set" indicator.

## 5. M3U source dialog: enabled control first

- [x] 5.1 In `M3uSourceDialog.tsx`, move the "Download enabled" checkbox block to immediately after the name field's `<div>` and before the file path field, per the "Enabled control appears first in the dialog" scenario. Verify by reviewing the rendered form's DOM order (add/adjust a test if `M3uSourceDialog` has or gains a test file; otherwise confirm manually via the running app since no test file exists today).

## 6. Système tab: remove bootstrap, reorganize into cards

- [x] 6.1 Delete `BootstrapConfigCard.tsx` and `BootstrapConfigCard.test.tsx`; remove its import and `<BootstrapConfigCard bootstrap={bootstrap} .../>` usage from `ConfigurationPage.tsx`'s "system" tab panel.
- [x] 6.2 Remove the now-fully-unused bootstrap plumbing: the `bootstrap` state/fetch in `useAppSettings.ts` (and its destructuring in `ConfigurationPage.tsx`), the `getBootstrapSettings` call in `services/api.ts`, the `BootstrapField` type in `types.ts`, and the `readOnly`/`BootstrapField`-union support in `SettingsFieldRow.tsx` (confirmed unused by any other caller). Verify `useAppSettings`-related tests still pass and `tsc`/build has no unused-export or type errors.
- [x] 6.3 Remove the unused `fields.bootstrap.*`, `groups.bootstrap`, and `bootstrapHint` keys from `frontend/src/locales/en/settings.json` and `fr/settings.json`.
- [x] 6.4 In `SystemStatusSection.tsx`, restructure the service-health rows, disk usage block, and build-info block into `settings-group-card` elements inside a `settings-cards-grid` (e.g. one card for health, one for disk, one for build/about), reusing the existing CSS classes rather than the current ad-hoc inline-styled `<div>`s. Verify `SystemStatusSection.test.tsx` still passes (update selectors if they targeted the removed inline structure) and the tab visually matches the rest of the page's card layout.

## 7. Full verification

- [x] 7.1 Run the full frontend test suite (`npm test` in `frontend/`) and confirm all tests pass, including the new/updated ones from tasks 2-6. (3 pre-existing failures in `DownloadsTab.test.tsx`/`ErrorsTab.test.tsx` pagination, unrelated to this change and present on the base branch, remain.)
- [x] 7.2 Run the app locally (per the project's `run` workflow) and manually walk the Configuration page: verify tab bar/header alignment at a wide viewport, search "rad" surfaces the Radarr card, both log level fields show the four-option dropdown, an M3U source without a password shows no password line, the M3U dialog's enabled checkbox is first, and the "Système" tab shows no database/API/metrics fields with its remaining content in cards.
