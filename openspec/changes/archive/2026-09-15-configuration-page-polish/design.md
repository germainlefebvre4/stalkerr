## Context

See `proposal.md` for motivation. All six fixes live inside the already-shipped Configuration page (`frontend/src/components/ConfigurationPage.tsx` and its child sections/cards) introduced by `redesign-configuration-page`. No backend change is involved anywhere in this change.

## Goals / Non-Goals

**Goals:**
- Fix the page-width layout drift between the tab bar/header and the rest of the centered content, without introducing a new drift point for future sections.
- Extend search and add an enumerated-`<select>` field type using the existing `SettingsGroupFieldSpec`/`SettingsFieldRow` extension points, not a parallel mechanism.
- Remove the bootstrap card cleanly (component, keys, tests) rather than leaving it disabled or empty.

**Non-Goals:**
- Redesigning the visual style (colors, spacing scale, card look) of the Configuration page — this only fixes alignment and content, reusing existing classes/tokens.
- Touching `logging.format` or any other fixed-value field beyond the two log level fields named in the proposal.
- Any change to how M3U/settings overrides are stored or resolved server-side.

## Decisions

### 1. Page-width alignment: one wrapping container, not five individual caps
Today `.configuration-summary-banner`, `.settings-search-input-wrap`, and `.settings-tab-panel` each independently declare `max-width: 960px; margin: 0 auto`, while `.settings-tabs-list` and the header row declare neither — hence the drift. Rather than adding the same two declarations to the tab bar and header (a fourth and fifth copy of the same rule), wrap the page's header, summary banner, search input, tab list, and tab panels in a single container element that carries `max-width: 960px; margin: 0 auto` once, and drop the now-redundant per-element declarations.
- **Alternative considered**: add `max-width`/`margin: auto` directly to `.settings-tabs-list` and the header row (matches the existing per-element pattern, smaller diff). Rejected because it repeats the exact class of bug being fixed here — the next new top-level element (e.g. a future banner) would need the same reminder and could just as easily be forgotten again.

### 2. Enumerated field type shape
Add `'select'` to `SettingsGroupFieldSpec['type']` (alongside `'text' | 'number' | 'boolean'`), with a new optional `options: { value: string; labelKey: string }[]` used only when `type === 'select'`. `SettingsFieldRow` renders a `<select>` from `options` when `type === 'select'`, mapping each `labelKey` through `t()` the same way field labels already are. `settingsFieldGroups.ts` defines the four-level option list once and reuses it for both `logging.app.level` and `logging.database.level`.
- **Alternative considered**: a generic `enum` type name instead of `select`, to mirror the backend's terminology. Kept `select` because it names the rendered control, consistent with the existing `'boolean'` type naming the toggle control it renders (per "Boolean Field Control"), not the data shape.

### 3. Search title-matching keeps the whole group, does not also highlight/filter within a title-matched group
Per the "Search matches a group's title" scenario, matching "rad" against "Radarr" shows every Radarr field, not just ones whose own label also happens to match. `SettingsGroupCard` computes `visibleFields` as today (label match) but skips that filter entirely — falls back to the full `fields` array — whenever the group's own title already matches the query.
- **Alternative considered**: matching the title AND still filtering to only label-matching fields, unioned with "show all if title matches". Equivalent outcome for this proposal's scope; implementing it as an early-return (title matches → use unfiltered `fields`) is simpler and was chosen.

### 4. Bootstrap card removal is a deletion, not a hide
`BootstrapConfigCard.tsx` and `BootstrapConfigCard.test.tsx` are deleted outright (not conditionally rendered as empty), its import and usage removed from `ConfigurationPage.tsx`'s "Système" tab content, and the now-unused `fields.bootstrap.*`, `groups.bootstrap`, and `bootstrapHint` locale keys are removed from both `en/settings.json` and `fr/settings.json`. The `BootstrapField` type and the `useAppSettings` hook's `bootstrap` fetch may still be used elsewhere (verify before removing them from `useAppSettings`/`types.ts`); if `bootstrap` becomes unused everywhere, remove it too rather than leaving a dead fetch.

### 5. Système tab content reorganized into existing card primitives
The service-health rows, disk usage, and build-info blocks in `SystemStatusSection.tsx` move from ad-hoc inline-styled `<div>`s into `settings-group-card` elements inside a `settings-cards-grid`, matching every other tab. This is a pure presentational refactor of existing JSX/CSS classes already used elsewhere on the page; no new component is introduced.

## Risks / Trade-offs

- [Consolidating the 960px cap into one wrapper touches every child of the page] → Low risk: purely a layout-container change, each child's own internal styling is untouched; verify visually across the six tabs after the change.
- [Removing `bootstrap` fetch/type if it turns out to be used elsewhere, e.g. a different diagnostics view] → Mitigation: grep for `useAppSettings().bootstrap` and `BootstrapField` usage outside `BootstrapConfigCard` before deleting anything beyond the card itself; keep the hook/type if any other consumer exists.
