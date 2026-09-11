## Why

The header currently carries two separate, growing icons (system status, language) while the "Filtres de Tri" tab consumes a full desktop-only tab slot for what is, in practice, an occasional configuration task. As more cross-cutting preferences accumulate (default tab, default page size, default playlist view — already scattered across `localStorage` keys with no discoverable UI), the header and tab bar are not a scalable home for them. Consolidating these into a single settings entry point reduces header clutter, frees a tab slot, and gives future preferences (and the requested light/dark theme) one discoverable home.

## What Changes

- Replace the header's system-status icon and language `<select>` with a single settings icon, identical on desktop and mobile, sharing the title's row on both.
- **BREAKING**: Remove the "Filtres de Tri" tab from the tab bar (desktop segmented tabs and the `VALID_TABS`/URL `tab` state); `?tab=filters` is no longer a valid tab value.
- Add a right-anchored settings drawer, opened from the new header icon, with collapsible sections:
  - **Apparence**: light/dark theme toggle, reduce-motion toggle.
  - **Langue**: the existing language switcher, relocated.
  - **Filtres**: the existing filter grid and creation flow, relocated, collapsed by default.
  - **Préférences**: default startup tab, default playlist page size, default playlist view (list/grouped) — surfacing the preferences already silently persisted in `localStorage`.
  - **Système**: the existing system-status content (service rows, disk usage, refresh action), relocated, collapsed by default.
  - **À propos**: app name and repository link.
- Introduce a dark theme token set and a `light`/`dark` mode switch (no theme tokens exist today — `variables.css` defines a single light palette).
- Explicitly out of scope: unifying the stats/logs polling intervals (they stay hardcoded at 10s/5s respectively) and a compact/comfortable density mode (density stays as-is, no setting).

## Capabilities

### New Capabilities
- `frontend-settings-drawer`: the settings drawer itself — header entry point, open/close behavior, section layout, the Apparence (theme + reduce-motion) and Préférences (default tab/page size/playlist view) sections, and the À propos section.

### Modified Capabilities
- `system-status-view`: the status icon and dialog are replaced by the drawer's collapsible "Système" subsection — fetch-on-first-expand instead of fetch-on-dialog-open, refresh action unchanged.
- `frontend-i18n`: the language switcher moves from a standalone header control into the drawer's "Langue" subsection.
- `frontend-filters-management`: the "Filtres de Tri" dedicated tab becomes a collapsible subsection inside the drawer; the filter grid, creation dialog, and delete flow are unchanged in behavior.
- `frontend-responsive-layout`: "Filtres" is removed from the desktop-only tab list and its mobile-narrowing fallback rule (it is no longer a tab at all, on any viewport); the tab-content styling requirement enumerating "all four tabs" drops to three (Playlist, Logs, Downloads).

## Impact

- Frontend only: `App.tsx` (`VALID_TABS`, tab list, `Tabs.Trigger` for filters, mobile fallback effect), `FloatingHeader.tsx` (replace status/language controls with one trigger), new `SettingsDrawer` component composing the relocated `SystemStatusDialog` content, `FiltersTab` content, and the language switcher.
- `variables.css`: new dark-mode token values alongside the existing light tokens.
- New `localStorage` keys for theme and reduce-motion; existing `stalkeer_active_tab` / `stalkeer_playlist_limit` / playlist-view keys become editable through the new Préférences section rather than only implicitly set.
- New i18n keys under a `settings` namespace (or extending `dialogs`/`common`) for the drawer's section labels.
- No backend/API changes.
