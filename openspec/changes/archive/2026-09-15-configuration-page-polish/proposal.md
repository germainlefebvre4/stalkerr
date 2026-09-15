## Why

`redesign-configuration-page` shipped the tab-based Configuration page, but several rough edges surfaced once it was in daily use: the tab bar and page header visually drift from the rest of the page's centered content, the quick search only matches field labels so searching an integration's name (e.g. "rad" for Radarr) hides its whole card instead of surfacing it, the two logging-level fields are free-text even though the backend only accepts four fixed values, the M3U source card announces an absent password instead of just omitting it, the M3U source dialog buries its own enable/disable toggle at the bottom of the form, and the "Système" tab still shows raw database/API/metrics connection variables that carry little value for an end user.

## What Changes

- The tab bar (`.settings-tabs-list`) and the page header (title + close button) are re-aligned with the rest of the Configuration page's centered, 960px-capped content instead of spanning the full page width.
- The quick settings search additionally matches each settings group/card's title (e.g. "Radarr", "Sonarr"), not only individual field labels; a title match keeps the whole card visible.
- `logging.app.level` and `logging.database.level` render as a `<select>` with the four backend-accepted technical values (`debug`, `info`, `warn`, `error`) mapped to localized display labels (Debug, Info, Warning, Erreur), instead of a free-text input.
- The M3U source card only displays the "Auth password" line when a password is actually set; it is omitted entirely otherwise.
- The M3U source create/edit dialog moves the "Download enabled" checkbox to the top of the form, right after the source name field, ahead of the connection/tuning fields.
- The "Système" tab drops the read-only bootstrap fields for database connection, API port, and metrics (port/enabled/path) — the `BootstrapConfigCard` component and its backing requirement are removed — and the tab's remaining content (service health, disk usage, build info) is rearranged into the same card layout used by the rest of the Configuration page instead of ad-hoc stacked sections.
- **BREAKING**: none.

## Capabilities

### New Capabilities
None — every change below refines how `redesign-configuration-page`'s already-shipped capabilities are presented; no new backend or frontend capability is introduced.

### Modified Capabilities
- `frontend-configuration-page`: the "Quick Settings Search" requirement is extended to also match a settings group's title, not only its fields' labels.
- `frontend-app-settings-management`: the "Bootstrap Configuration Display" requirement is removed (database/API/metrics fields are no longer shown anywhere in the Configuration page); the "Settings Field Layout" requirement's example group list drops the now-removed bootstrap group; a new "Enumerated Field Control" requirement (sibling to "Boolean Field Control") covers select-based fields with a fixed value set, applied to the two logging-level fields; the "M3U Sources Management" requirement is amended so the auth password status is shown only when a password is set, and so the source dialog's enabled/disabled control appears before the source's other configuration fields.

## Impact

- Frontend only: `ConfigurationPage.tsx` (header/tabs alignment), `SettingsGroupCard.tsx` and `SettingsFieldRow.tsx` (title-aware search, new `'select'` field type), `settingsFieldGroups.ts` (log level fields gain `type: 'select'` + options), `M3uSourcesSection.tsx` (conditional password line) and `M3uSourceDialog.tsx` (field order), `SystemStatusSection.tsx` and `BootstrapConfigCard.tsx` (bootstrap card removed, remaining Système content moved to `settings-group-card`/`settings-cards-grid`), `index.css` (tab bar/header alignment).
- No backend/API changes: `internal/settings`, `internal/config`, override resolution and storage, and all endpoint contracts are unchanged — the four log level values and the M3U `has_auth_password` boolean already exist server-side and are only newly consumed correctly by the frontend.
- Locale files (`src/locales/{en,fr}/settings.json`) updated for the log level display labels; the bootstrap-related labels (`fields.bootstrap.*`, `groups.bootstrap`, `bootstrapHint`) become unused and are removed.
- Out of scope: `logging.format` (json/text) is a similarly fixed-value field but was not part of the reported issue and is left as free text; the M3U dialog's other field ordering (URL, file path, retention/retry) is unchanged beyond moving the enabled checkbox.
