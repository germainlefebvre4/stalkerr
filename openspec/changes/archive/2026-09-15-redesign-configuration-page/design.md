## Context

See `proposal.md` for motivation. Relevant current-state constraints (from the codebase, as of the `app-settings-overrides` change):

- `ConfigurationPage.tsx` is not a route: `App.tsx` holds a plain `isSettingsOpen` boolean that swaps the entire tab bar for the page. There is no URL representation of "Configuration is open" or of which of its ten sections is visible — the app's existing `tab` query-param/`localStorage` mechanism (`readInitialActiveTab`, `patchTabURLState`) only tracks the Home/Downloads/Playlist/Logs/Errors tabs.
- Ten sections render in a fixed order, four always-expanded (Apparence, Langue, Préférences, À propos) and six as `DisclosureSection` accordions (Filtres, Sources M3U, Intégrations, Notifications, Avancé, Système), each fetching its own data lazily on first expand via a `useEffect` gated on its own `isExpanded` boolean. There is no rule distinguishing which pattern a section uses.
- `SettingsFieldRow.tsx` renders one field per row: label, an "Interface"/"Config" text badge, a "Restart required" text badge, a masked input for sensitive fields, a boolean rendered as a `<select>` of `"true"`/`"false"`, and its own Save + "Reset to config" buttons — one field, one Save action.
- `AdvancedSection.tsx` locally duplicates a non-exported `BootstrapFieldRow` (same visual shape as `SettingsFieldRow`, no Save/Reset) to render the ten read-only bootstrap fields (`database.*`, `api.port`, `metrics.*`) inside the same "Avancé" disclosure as the thirteen editable Downloads-tuning fields and three Logging fields.
- `FiltersSection.tsx` delegates creation to `CreateFilterDialog.tsx` (a Radix `Dialog`, per `frontend-filters-management`); `M3uSourcesSection.tsx` delegates to its own `M3uSourceDialog.tsx` — two separate implementations of the same "list of named items, create/edit via dialog, warn before replacing an existing name" shape.
- `useAppSettings`, `useM3uSources`, and `useFilters` each fetch their own data independently; nothing today aggregates "how many overrides are currently active" across them. `useSystemStatus` fetches only on demand (no auto-fetch on mount), per `system-status-view`'s explicit "no fetch before expand" requirement — unaffected by this change.
- `system-status-view` and `frontend-i18n` still reference the retired "Settings drawer" in their requirement text — a wording gap left over from the drawer-to-page migration, unrelated to this change's own scope but touched here since both capabilities' entry points are being renamed again anyway.

## Goals / Non-Goals

**Goals:**
- Replace the ten-section single scroll with six tabs (Général, Intégrations, Contenu, Notifications, Avancé, Système), each independently addressable by URL and remembered across sessions.
- Give the page a global, at-a-glance view of active overrides and pending restarts without visiting every tab.
- Replace the one-Save-button-per-field model with one dirty-state save action per logical group of fields, without changing what an override actually stores.
- Lay out field groups in a responsive card grid capped to a readable width instead of a single full-width column.
- Make the four general UX fixes agreed with the user: real toggle switches for booleans, a compact origin indicator, one shared "create a new item" interaction for Filtres and Sources M3U, and a quick client-side settings search.

**Non-Goals:**
- No backend/API change to app-settings, M3U sources, or system status: override storage, resolution (`settings.Effective()`), endpoint contracts, and the field registry are untouched for those domains. The one exception, discovered during implementation, is described under "Filters backend gap" below.
- No dedicated mobile redesign of the Configuration page. The page keeps behaving as it does today with respect to `frontend-responsive-layout` (that capability already exempts it as "a separate overlay surface"); the new grid degrades to fewer columns at narrow widths by ordinary CSS wrapping, but a tailored mobile tab/card treatment is left for a later change if needed.
- No change to what counts as an "override" or to restart-required semantics — only to how those existing facts are surfaced (per-field, per-tab, and in aggregate).
- No encryption, secret-handling, or M3U/filter override-mechanism changes — out of scope, unaffected by presentation.

## Decisions

### Tabs, not disclosures; grouped by nature, not by arrival order
Six tabs replace the ten sections: **Général** (Apparence, Langue, Préférences), **Intégrations** (Radarr, Sonarr, TMDB, Jellyfin), **Contenu** (Filtres, Sources M3U), **Notifications**, **Avancé** (Downloads tuning, Logging, M3U update interval — editable fields only), **Système** (health, disk, build version, À propos' name/repo link, and the bootstrap read-only fields moved out of Avancé). Within a tab, the former section identities (e.g. "Apparence", "Langue" inside Général) are kept as named sub-groups/cards, not flattened — so requirements naming those sections (Theme Toggle, Reduce Motion Toggle, the two Préférences requirements) need no wording change, only their container does.

Implemented with the same Radix `Tabs` primitive already used for the app's main Home/Downloads/Playlist/Logs tab bar, for visual and accessibility consistency and to avoid a new dependency.

**Alternative considered**: mirror the ten sections 1:1 as ten tabs — rejected as too granular (several would be single-card tabs) and it doesn't fix the "editable/read-only mixed together" problem in "Avancé" that motivated this change.

### URL addressability reuses the existing tab-state mechanism
The Configuration page becomes a recognized value (`"settings"`) of the app's existing `tab` URL query param/`localStorage` pair, instead of the disconnected `isSettingsOpen` boolean — so `?tab=settings` alone opens it, consistent with how `?tab=downloads` already opens the Downloads tab. A second param, `settingsTab`, carries which of the six internal tabs is active (`general | integrations | content | notifications | advanced | system`), read with the same precedence order the app already uses for its main tab (`tab` param): URL param first, then a new `localStorage` key (`stalkeer_configuration_tab`) for the last one used, then `"general"` as the default. Selecting a tab updates `settingsTab` in the URL (via the existing `patchTabURLState` helper) and in `localStorage`. The main app's previously-active tab (what "Close" returns to) keeps being tracked exactly as today — it is orthogonal to which Configuration-page tab is active.

**Alternative considered**: a fully separate route/history entry for Configuration (e.g. via a router) — rejected, the app has no router today and the existing query-param mechanism already solves "deep-linkable, back-button-friendly, remembers last value" for the main tabs; reusing it is the smaller change.

### Prefetch enough to make the summary banner true on open
The summary banner and per-tab override badges need to know, as soon as the Configuration page mounts, how many app-settings fields, M3U sources, and filters currently carry an override — before the user has visited those tabs. This page fetches `getSettings()`, `getM3uSources()`/`getM3uSourcesOrigin()`, and `getFilters()`/the filters origin endpoint once when the page mounts (a handful of lightweight GETs already used elsewhere), rather than only when each tab is first activated. `useSystemStatus` (health/disk/build) is unaffected and keeps its existing on-demand-only fetch — it is diagnostic, not an override, and isn't part of the counted totals.

**Alternative considered**: keep every dataset fully lazy and show a banner that fills in progressively as tabs are visited, or reads "0" until then — rejected as actively misleading (a "0 overrides" banner the user hasn't earned yet is worse than a short extra request burst on open).

### One dirty-state save per group; "reset to config" stays immediate and per-field
Typing a new value into a field stages a pending change (visually marked, not yet submitted). The group/card it belongs to then shows a save action ("Enregistrer" / "Annuler") only while it has at least one pending field; confirming submits every pending field in that group as stored overrides (existing per-field `PUT /api/v1/settings/:key` and `PUT /api/v1/m3u/sources/:name` calls, fired together, not a new bulk endpoint); cancelling discards the drafts. Clearing an existing override (the "reset to config" affordance, now a small icon rather than a text button) stays an immediate, single-field action independent of any other pending changes in the same group — it already behaves this way today and changing it to a staged action would make "revert" surprising (a user reverting a mistake would expect it to take effect right away, not queue behind unrelated edits).

**Alternative considered**: autosave on blur — rejected per user preference; also would fight the "sensitive field starts empty" convention (a blur-triggered save could easily fire on an untouched masked field).

### Settings field groups render as cards in a capped, responsive grid
The Configuration page's content column is capped at **960px**, centered within the tab panel. Within that width, a field group (e.g. Radarr, Sonarr, Logging, Apparence) with **6 or fewer fields** renders as a compact card with a minimum width of **280px**, so multiple compact cards pack into a row (up to three at 960px); a group with **more than 6 fields** (Downloads tuning: 13; the bootstrap group: 10) always spans the tab panel's full width. Within a card, fields lay out on an internal 2-column grid for compact/short-value fields (boolean, number, short select) and full-card width for long-value fields (URL, text, secret) — this is what actually fixes "a boolean on 1600px of width."

**Alternative considered**: a fixed N-column grid for every group regardless of field count — rejected, it's what produced the "boolean alone on a huge row" complaint in the first place for small groups, or would cramp Downloads' 13 fields into an unreadably narrow card.

### Shared components replace the per-field Save button and the duplicated read-only row
- `SettingsFieldRow` drops its own Save button and its always-visible text origin badge; it keeps rendering one field (label, control, compact origin indicator, restart marker, reset-to-config icon) and gains a `readOnly` mode, replacing the separately-maintained `BootstrapFieldRow`.
- A new wrapping component (name TBD in tasks, e.g. `SettingsGroupCard`) renders a set of `SettingsFieldRow`s as one card, owns the pending-changes set, and renders the group's save/discard bar when dirty.
- Booleans render through the existing `ToggleSwitch` component (already used for Apparence today), promoted to a shared/exported component instead of being local to `ConfigurationPage.tsx`.
- The origin indicator becomes a small marker (e.g. a colored dot/left-border) with the full "Interface"/"Config" label available via a Radix `Tooltip` on hover/focus, rather than a permanent text pill on every row.
- Filtres' and Sources M3U's create/edit dialogs converge on one shared interaction shape: a "+" trigger in the section's own header, a Radix `Dialog`, and an inline warning banner inside the dialog (not a native `confirm()`) when submission would replace an existing item. The exact shared component boundary (a common dialog shell vs. two call sites of shared building blocks) is an implementation choice left to `tasks.md`.
- A single search input, scoped to the currently active tab and persisted across tab switches (in React state, not the URL), filters visible field rows and Filtres/Sources M3U list entries by matching the query against their translated label.

### Filters backend gap (discovered during implementation)

Task 6 (dialog unification) assumed `CreateFilterDialog` already had a "load current configuration" affordance and a replace-warning, and that `FiltersSection` already grouped by attribute and showed origin vs. override, per the pre-existing (unchanged by this proposal) `frontend-filters-management` spec. None of that is actually implemented: `FiltersSection.tsx` renders a flat list with no origin data, `CreateFilterDialog.tsx` has no "Reprendre la config actuelle" button, and — the root cause — **no backend endpoint exposes the `config.yml` origin filter patterns** at all (`GET /api/v1/filters` only returns database-backed runtime rows), and `createFilter`/`updateFilter` don't enforce `filter-override-policy`'s "at most one active runtime override per attribute" (they only enforce a unique `name`, so nothing stops two runtime rows sharing an attribute, and `internal/filter`'s `Matches()` ANDs together every runtime filter it finds for an attribute instead of picking one).

This is a pre-existing gap against already-approved specs (`frontend-filters-management`, `filter-override-policy`), not something this proposal's UI rework introduces. The user chose to close it as part of this change rather than defer it. Scope added, kept to exactly what's needed to make the unified dialog and the "Filters List View" requirement true:
- A new read-only `GET /api/v1/filters/origin` endpoint returning, per attribute, the `config.yml` include/exclude patterns (empty when unconfigured) — satisfies `filter-override-policy`'s Origin Filter Configuration Exposure requirement.
- `createFilter`/`updateFilter` now delete any existing runtime override for the target attribute before saving the new one, so at most one remains — satisfies `filter-override-policy`'s Single Active Runtime Override Per Attribute requirement.
- Not touched: `internal/filter`'s `Matches()` multi-filter-AND behavior for any pre-existing rows that already violate the one-per-attribute invariant (the API-level fix above prevents new violations; a data migration for existing bad rows is out of scope), and the separate include/exclude pattern encoding question (`internal/filter` JSON-decodes stored patterns while the create dialog submits a raw string) — unrelated to origin exposure and left as-is.

## Risks / Trade-offs

- **More eager fetching on page open.** Prefetching settings/M3U/filters data as soon as Configuration mounts (instead of only per-tab-activation) means a handful of extra parallel requests every time the page opens, even if the user only ever looks at one tab. → Accepted: these are lightweight reads already used elsewhere in the app, and the alternative (a banner that lies or requires visiting every tab) is worse.
- **Grouping/save semantics diverge slightly from a strict "per field" model.** A user who edits two fields in the same card and only wants to keep one now must use the per-field reset (before saving) or edit only the field they intend to change — they can no longer "half-save" a card by clicking two different per-field buttons. → Accepted: this is the explicit trade of the new save model; the per-field "reset to config" icon and the "Annuler" discard action together still cover both "undo one field" and "undo everything I haven't saved yet."
- **Two more stale-wording capabilities touched.** `system-status-view` and `frontend-i18n` still say "Settings drawer"; fixing that here (since their entry points are being renamed to a tab anyway) is opportunistic scope, not something the user asked for directly. → Kept minimal: only the location wording changes, no behavioral change to either capability beyond the fetch-trigger rename in `system-status-view` (expand → activate).

## Migration Plan

Purely additive/reshuffling on the frontend; no data migration, no backend change, no rollback complexity beyond reverting the frontend build. Suggested implementation order (detailed further in `tasks.md`):
1. Promote `ToggleSwitch` to a shared component; extend `SettingsFieldRow` with a `readOnly` mode and drop `BootstrapFieldRow`.
2. Build the tab container (URL/localStorage-synced) and the six tab panels, moving existing section content in without behavior changes first.
3. Add the `SettingsGroupCard` wrapper (dirty-state, grid layout) and switch Intégrations/Notifications/Avancé/Général's field groups to it.
4. Add the summary banner and per-tab badges, wired to the prefetched data.
5. Converge Filtres' and Sources M3U's create dialogs on the shared interaction shape.
6. Add the quick search input.
7. Update locale files and the two stale-wording specs (`system-status-view`, `frontend-i18n`).

## Open Questions

- Exact shared-component boundary for the unified create/edit dialog (one generic component vs. two call sites sharing smaller pieces) — implementation detail, does not change any spec requirement, left to `tasks.md`.
