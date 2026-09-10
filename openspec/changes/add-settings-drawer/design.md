## Context

See `proposal.md` - Why/What Changes for motivation and scope. Relevant current state:

- `FloatingHeader.tsx` renders the system-status icon (opens `SystemStatusDialog`, a Radix `Dialog`) and the language `<select>` as two independent controls.
- `App.tsx` owns `VALID_TABS`, the desktop `Tabs.List`, and the mobile bottom tab bar (`tabs` array); "filters" is one of the seven tab values, already excluded from the mobile bar and from `MOBILE_FALLBACK_TAB` handling.
- `FiltersTab.tsx` is a `Tabs.Content` panel with its own data fetching (`useFilters`) triggered by `activeTab === 'filters'`.
- `variables.css` defines a single, light-only token palette (`--bg-*`, `--text-*`, `--status-*`, `--border-*`, `--shadow-*`); no dark tokens, no `data-theme`/`prefers-color-scheme` handling exist anywhere in the codebase.
- Preferences already live in `localStorage` under ad hoc keys (`stalkeer_active_tab`, `stalkeer_playlist_limit`) with no discoverable UI; language persistence is handled by `i18next-browser-languagedetector`.
- All dialogs in the app (`SystemStatusDialog`, `CreateFilterDialog`, `ManualOverrideDialog`, `MoveFolderDialog`, `RenameFolderDialog`, `RunItemsDialog`) use Radix UI `Dialog` primitives; there is no existing drawer/sheet or collapsible/accordion primitive in use.

## Goals / Non-Goals

**Goals:**
- One header entry point, one drawer, hosting the six sections defined in the specs.
- A reusable dark token set that every existing component picks up for free because they already consume `var(--...)` tokens.
- No behavior change to the relocated Filtres/Système/Langue controls beyond their entry point.

**Non-Goals:**
- No `prefers-color-scheme` auto-detection — the theme toggle is a plain two-state, user-set switch (per proposal, explicitly a "toggle").
- No unification of the stats (10s) and logs (5s) polling intervals.
- No density/compact mode.
- No backend or API changes — this is frontend-only.

## Decisions

**Drawer = Radix `Dialog.Content` styled as a right-anchored sheet, not a new dependency.**
The codebase has no drawer/sheet primitive, but it already depends on `@radix-ui/react-dialog` and uses it for six other modals. Radix's own drawer/sheet pattern (documented by Radix) is exactly `Dialog.Root` + `Dialog.Content` with drawer-specific CSS (fixed to the right edge, full height, slide-in transform) instead of the centered-modal CSS `dialog-content` class already in `index.css`. Alternative considered: adding a dedicated drawer library (e.g. Vaul) — rejected, it would be the only non-Radix UI primitive in the app for one screen.

**Collapsible sections (Filtres, Système) = local `useState<boolean>`, not `@radix-ui/react-collapsible`.**
Only two sections need expand/collapse, with no animation requirement beyond what the rest of the app already has. A boolean per section plus a conditional render meets every scenario in the specs (fetch-on-first-expand, content revealed in place). Alternative considered: pulling in `@radix-ui/react-collapsible` for its accessibility affordances — deferred; can be swapped in later without a spec change if keyboard/ARIA gaps surface, since the specs describe behavior (expand reveals content, first expand fetches) not the mechanism.

**Theme tokens: a parallel `[data-theme="dark"]` block in `variables.css`, toggled via a `data-theme` attribute on `<html>`.**
Every component already styles through `var(--token)`, so redefining the same custom property names under a `[data-theme="dark"]` selector (mirroring how `:root` defines the light values) re-themes the whole app without touching component code. The toggle sets `document.documentElement.dataset.theme` and persists the choice under a new `stalkeer_theme` `localStorage` key, following the existing `stalkeer_*` naming convention. Absent a stored value, no `data-theme` attribute is set and the app renders with the light `:root` values (matches the spec's "absent a stored preference, use light theme"). Alternative considered: CSS-in-JS theme objects — rejected, would mean rewriting every component's inline `style={{ color: 'var(--text-secondary)' }}` usage instead of reusing it as-is.

**Reduce-motion: gate the two known animations, not a global `prefers-reduced-motion` override.**
Two effects exist today: the toast's inline `animation: 'contentShow 150ms ease-out'` style in `App.tsx`, and the `.card:hover` transform/shadow transition in `index.css` (itself already scoped to `@media (hover: hover) and (pointer: fine)`, so it never fires on touch). The toggle persists under `stalkeer_reduce_motion` and sets a `data-reduce-motion="true"` attribute on `<html>`; the toast's animation becomes conditional on that attribute (inline style branch) and the card-hover CSS rule is qualified with `:root:not([data-reduce-motion="true"])`.

**Default startup tab / default playlist page size / default playlist view: additive writers of the existing `localStorage` keys, not a new preference layer.**
`stalkeer_active_tab` and `stalkeer_playlist_limit` are already read on mount by `readInitialActiveTab()` and `usePlaylist()` respectively; the Préférences section becomes a second place that writes the same keys (without navigating/re-rendering the currently active tab or view), rather than introducing parallel "default" keys that could drift from what normal navigation persists. This keeps `App.tsx`'s existing URL-param-then-`localStorage` precedence and `frontend-ihm-dashboard`'s existing persistence requirement unchanged — no delta needed there.

**Removing "filters" from `VALID_TABS` relies on the existing `isValid`/default fallback — no bespoke migration code.**
`TAB_URL_SCHEMA.isValid` already rejects any value not in `VALID_TABS` and `readInitialActiveTab()` already falls back to `'home'` when the stored/URL tab is invalid. Dropping `'filters'` from `VALID_TABS` makes any old `?tab=filters` link or stale `stalkeer_active_tab: "filters"` value degrade to Home automatically, through the same path new invalid values already take.

**Settings entry point icon must be visually distinct from the Logs tab's `⚙️`.**
Concretely: do not reuse `⚙️` for the settings trigger. Exact glyph is an implementation choice (see Open Questions).

## Risks / Trade-offs

- **Dark-theme completeness** → the retrofit assumes every visible color already flows through a `var(--token)`. A handful of components use inline hex/rgba values directly (e.g. the toast's `rgba(...)` shadow tints, KPI gradient stops). Mitigation: grep for hardcoded `#`/`rgb(`/`rgba(` in `.tsx`/`.css` during implementation and either route them through tokens or explicitly accept them as intentionally theme-invariant (shadows, gradients).
- **Drawer-as-styled-Dialog on mobile** → a full-height right-anchored sheet needs different sizing than the existing centered `dialog-content` class; getting the mobile breakpoint treatment wrong could reintroduce horizontal scroll. Mitigation: reuse the same `768px` breakpoint and touch-target rules already established by `frontend-responsive-layout`.
- **Two independent writers of `stalkeer_active_tab`/`stalkeer_playlist_limit`** (normal navigation and the new Préférences fields) → could surprise a user who expects "Préférences" to be the sole source of truth. Mitigation: the specs above make the "does not navigate the current session" behavior explicit; document it in the drawer's UI copy during implementation.

## Migration Plan

Frontend-only, no data migration, no feature flag (single self-hosted deployment). Ship as one change: remove the two header controls and the Filtres tab, add the drawer, add dark tokens. Rollback is a plain revert — no persisted server-side state is introduced.

## Open Questions

- Exact glyph for the settings entry point icon (must differ from the Logs tab's `⚙️`). Deferred to implementation/visual review; does not affect any requirement.
