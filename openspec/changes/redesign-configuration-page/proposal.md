## Why

The Configuration page (which replaced the former Settings drawer in `app-settings-overrides`) inherited that drawer's single-scrolling-page presentation and simply grew every new domain (Radarr, Sonarr, TMDB, Jellyfin, Notifications, Downloads tuning, logging, M3U sources) on top of what it already held (Apparence, Langue, Filtres, Préférences, Système, À propos). The result mixes three unrelated concerns — client-side preferences, backend-editable configuration, and read-only diagnostics — in one flat list of sections with no grouping logic, buries read-only bootstrap fields inside the same "Avancé" disclosure as editable ones, requires one Save click per field across ~50 fields, renders even a single boolean on the full page width, and gives no at-a-glance view of what is currently overridden or which changes still need a restart to take effect.

## What Changes

- The Configuration page is restructured from an ordered list of always-expanded/disclosure sections into **tabs**: Général (Apparence, Langue, Préférences), Intégrations (Radarr, Sonarr, TMDB, Jellyfin), Contenu (Filtres, Sources M3U), Notifications, Avancé (Downloads tuning, Logging, M3U update interval — editable fields only), Système (health, disk, build version, GitHub link, plus the bootstrap read-only fields moved out of Avancé).
- Tabs become **URL-addressable** (deep-linkable, restores browser back/forward) and the **last active tab is remembered** across sessions; today the page has zero URL state (a boolean toggle swaps out the whole tab bar).
- A **summary banner** at the top of the page shows the total count of active overrides and how many require a restart; each tab label carries a badge with its own override count and a marker when at least one of its fields requires a restart.
- Overridable fields render as **cards** grouped by logical unit (one card per integration/domain), laid out in a **responsive grid capped to a readable content width** — compact groups sit multiple-per-row, larger groups (e.g. Downloads tuning) span the full row — instead of a single full-width column where even a boolean stretches edge to edge.
- The **save model** changes from one Save button per field to one save action per card/group, appearing only when that group has an unsaved change (dirty-state detection); the existing per-field "reset to config" becomes a small icon instead of a persistent text button.
- **Boolean fields render as toggle switches** (reusing the existing Apparence toggle component), replacing the current true/false text `<select>`.
- The **Interface/Config origin indicator** becomes a compact visual marker with full detail on hover/focus, replacing today's persistent text badge on every field.
- The **"create a new item" interaction is unified** across Filtres and Sources M3U — today's two separate dialog implementations for the same list+CRUD shape converge on one pattern.
- A **quick client-side search** filters visible settings fields by label.
- **BREAKING**: none.
- **Scope note (added mid-implementation)**: this was planned as a frontend-only presentation and interaction change, but implementing the unified Filtres/Sources M3U create dialog surfaced that `frontend-filters-management`'s existing "Filters List View" requirement (origin vs. override display) and `filter-override-policy`'s existing "Origin Filter Configuration Exposure" and "Single Active Runtime Override Per Attribute" requirements were never actually implemented server-side. A new read-only `GET /api/v1/filters/origin` endpoint and an attribute-replace fix to `createFilter`/`updateFilter` are added to make those already-approved requirements true; see `design.md`'s "Filters backend gap". No override semantics beyond that fix, storage shape, or other endpoint contracts change.

## Capabilities

### New Capabilities
None — this change reworks how existing capabilities are presented and edited; no new backend or frontend capability is introduced.

### Modified Capabilities
- `frontend-configuration-page`: page structure changes from an ordered list of always-expanded/disclosure sections to tab-based navigation (Général/Intégrations/Contenu/Notifications/Avancé/Système), with URL-addressable tabs, last-active-tab memory, a summary banner, per-tab override/restart badge counts, and a quick search over visible settings fields.
- `frontend-app-settings-management`: Radarr/Sonarr/TMDB/Jellyfin/Notifications/Downloads/logging fields render as cards in a responsive grid instead of full-width rows; save action moves from per-field to per-card with dirty-state detection; boolean fields use a toggle switch instead of a true/false select; the origin badge becomes a compact indicator with hover/focus detail; bootstrap read-only fields move from the "Avancé" grouping to the "Système" tab.
- `frontend-filters-management`: the "add a filter" interaction is unified with the Sources M3U creation pattern (same trigger/dialog shape/replace-warning convention), and the Filtres section moves under the "Contenu" tab instead of a page-level disclosure.
- `system-status-view`: reachable via the Configuration page's "Système" tab — supersedes stale "Settings drawer" wording left over from the prior drawer-to-page migration.
- `frontend-i18n`: language switcher reachable via the Configuration page's "Général" tab — supersedes stale "Settings drawer" wording.

## Impact

- Frontend only: `ConfigurationPage.tsx` restructured into a tab container plus tab panels; `IntegrationsSection.tsx`, `NotificationsSection.tsx`, `AdvancedSection.tsx`, `FiltersSection.tsx`, `M3uSourcesSection.tsx`, `SystemStatusSection.tsx` adapted to the card/grid layout and relocated under their respective tabs; `SettingsFieldRow.tsx` reworked for toggle-switch booleans, the compact origin indicator, and group-level dirty-state save instead of a per-field save button; `BootstrapFieldRow` relocated into the Système tab.
- Likely new shared components (exact shape deferred to `design.md`): a tab container with URL sync, a card/grid layout primitive, a dirty-state save bar, a settings search input.
- No backend/API changes: override resolution, storage, and endpoint contracts (`/api/v1/settings`, `/api/v1/m3u/sources`, `/api/v1/filters`, `/api/v1/system/status`) are unchanged; this only changes how the frontend presents and edits them.
- Locale files (`src/locales/{en,fr}/settings.json`, `dialogs.json`) updated for new tab labels, search placeholder, and summary banner strings.
- Out of scope: any change to `internal/settings`, `internal/config`, the override storage model, the Helm chart's config delivery, or secret encryption at rest — all unaffected by this change.
