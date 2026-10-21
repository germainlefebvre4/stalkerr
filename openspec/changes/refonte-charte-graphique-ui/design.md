# Design

## Context

See `proposal.md` for motivation. Current frontend state (observed):

- `frontend/src/variables.css` (125 lines) holds light tokens in `:root` and dark overrides in `[data-theme="dark"]`; `index.css` (1612 lines) styles everything through classes such as `.glass-header`, `.segmented-tabs-list`, `.card`.
- `index.html` has a pre-paint script that sets `data-theme="dark"` only when `localStorage.stalkeer_theme === 'dark'`; absent that, the app is light. `useTheme` and `useReduceMotion` (`data-reduce-motion` on `<html>`) already exist.
- Inter is referenced in `--font-family` but never loaded.
- `App.tsx` renders `FloatingHeader` (title, subtitle, settings button) then a Radix `Tabs.Root` whose `Tabs.List` uses emoji labels. Tab state lives in `localStorage` (`stalkeer_active_tab`) and the URL. Below 768px a bottom bar is rendered via CSS/`useMediaQuery`.
- ~320 inline `style={{…}}` already use `var(--token)`; 7 hard-coded colors live in `ManualOverrideDialog`, `ErrorsSidepanel`, `MediaOccurrenceDrawer`, plus a dozen `rgba(99,102,241…)` literals.
- Available data for the sidebar: `GET /health` (`api.getHealth`), `GET /api/v1/system/status` → `version` (`useSystemStatus`), and the errors list total (`useErrorsTab`, which only fetches while the Errors tab is active).

## Goals / Non-Goals

**Goals:**
- A coherent mission-control look delivered through tokens first, so the app visibly changes after phase 1 and stays usable after every phase.
- Desktop sidebar + telemetry bar; mobile structure preserved.
- Fully self-hosted (fonts, icons), no CDN.
- Visual regression safety through existing Vitest tests plus targeted new ones for the shell.

**Non-Goals:**
- No backend or API contract changes, no new product features, no artwork/poster imagery.
- No migration to a CSS framework or CSS-in-JS; plain CSS with custom properties stays.
- No rewrite of component logic; only markup/classes/styles.

## Decisions

### D1. Tokens first, keep existing token names
Keep the current token names (`--bg-app`, `--bg-card`, `--text-main`, `--primary-accent`, `--status-*`, …) and change their values, adding only what is missing (`--bg-surface-2`, `--font-mono`, `--led-*`). Rationale: the ~320 inline `var(--token)` usages and all of `index.css` re-theme for free, giving ~70% of the effect in phase 1. Alternative: rename to a new semantic scheme (`--surface-1/2/3`) — cleaner but forces touching every usage up front for no visual gain; deferred, may be done opportunistically in phase 4.
Remove (set to `none`/`0`) `--shadow-*`, `--accent-gradient*`, `--glass-*` consumers rather than deleting the variables at once; delete the variables when no reference remains (checked by grep in the final task).

### D2. Palette
Dark (default): background `#0a0e14`, surface `#11161f`, surface-2 `#171d28`, rule `#222a37`, text `#d7dde8`, secondary `#8b95a7`, accent/in-progress `#22d3ee`, success `#3fb950`, pending `#f5a524`, failed `#f85149`, neutral `#8b95a7`.
The exploratory "muted" value `#5a6577` is only ~3.2:1 on `#0a0e14` and would violate the 4.5:1 requirement of `frontend-design-system`; use a lighter muted text value (≈ `#7a8599`, to be verified with a contrast check in task 1.2) and keep `#5a6577` for non-text decoration only.
Light: neutral cool-gray background (`#f3f5f8`), white surfaces, 1px rules (`#d5dbe4`), a darker cyan accent (`#0e7490`) so the accent text meets 4.5:1 on white, same status hues darkened for contrast.

### D3. Typography: `@fontsource` packages
Add `@fontsource/ibm-plex-sans` (400, 500, 600) and `@fontsource/ibm-plex-mono` (400, 500), imported in `main.tsx`. Vite bundles the woff2 files, satisfying "no CDN". Only latin/latin-ext subsets are imported to limit weight. Alternative: JetBrains Mono for data — rejected to keep a single family with matching metrics. Alternative: Google Fonts link — rejected (homelab, offline). Mono is applied via a `.mono` utility class and on table cells/data spans; `font-variant-numeric: tabular-nums` for numeric columns.

### D4. Sidebar as a new component, same tab state
Create `Sidebar.tsx` and render it at ≥768px; keep `Tabs.Root` as the state owner so `activeTab`, URL sync and `localStorage` persistence are untouched. The sidebar renders `Tabs.List orientation="vertical"` with `Tabs.Trigger` entries so roving tabindex/arrow-key behavior is preserved. The mobile bottom bar remains as today (re-skinned and with lucide icons). `FloatingHeader` is reduced to the mobile title row (title + settings button); on desktop the settings button moves to the sidebar footer. Layout: CSS grid `grid-template-columns: 220px 1fr` at ≥768px, sidebar `position: sticky; height: 100vh`.
Alternative: collapsible icon-only sidebar — deferred; adds state and specs for little value now.

### D5. Sidebar data sources
- Version: `useSystemStatus` (`version`), shown as `v<version>`; hidden if unavailable.
- API health LED: `useHealthAndStats`-style polling of `api.getHealth()` every 10s (`ok` → success LED, failure → failed LED).
- Error badge: add a small `useErrorCount` hook that calls the existing downloads list endpoint with the combined problem filter (`COMBINED_PROBLEM_FILTER`) and `limit=1`, reading only the total, polled every 30s. Needed because `useErrorsTab` only fetches while the tab is active. No backend change; if the endpoint cannot return a total with `limit=1`, fall back to the existing limit/total fields it already returns (verified in task 2.4).

### D6. Telemetry bar replaces desktop KPI cards
Reuse the data already computed for the KPI block (stats, success ratio) and render it as `label value` pairs in a single flex row, `.mono` values, separated by 1px rules. The mobile KPI block keeps its markup and is re-skinned only. Downloads-specific indicators (active count, rate) are shown only when already available from existing hooks; no new endpoints.

### D7. Status LEDs through a single primitive
Introduce a `StatusLed` component (dot + label, `data-status` driving the token) and a `.led` CSS class; migrate the existing pastel badge classes to it. Pulse uses a CSS animation disabled by `[data-reduce-motion="true"]` and by `@media (prefers-reduced-motion: reduce)`, which satisfies the spec's OS-level clause for free.

### D8. Theme default
Change the `index.html` bootstrap script to set `data-theme="light"` only when the stored value is `'light'`, otherwise `'dark'` (set explicitly on `<html>`), and make `useTheme`'s fallback `'dark'`. Existing users with `stalkeer_theme === 'dark'` are unaffected; users with `'light'` keep light.

### D9. Inline style migration strategy
Phase 1 leaves inline styles alone (they follow tokens). Phase 3 introduces primitives (`.btn`, `.input`, `.table-dense`, `.drawer`, `.dialog`, `.led`). Phase 4 migrates screen by screen, starting with the three files containing hard-coded colors, then the heaviest (`RadarrSonarrTab` 65, `ErrorsSidepanel` 64, `MediaOccurrenceDrawer` 63). Rule: only move to classes what is visual (color, radius, border, font, density); dynamic layout values computed from props stay inline. A lint-style check (grep in a test or CI script) asserts no `#hex`/`rgba(` literals remain in `src/components/**/*.tsx`.

### D10. Favicon and mark
Replace the emoji data-URI with an inline SVG data-URI (simple satellite/radar glyph in cyan on dark) in `index.html`; reuse the same glyph (lucide `Satellite`/`Radar`) as the sidebar mark.

## Risks / Trade-offs

- [Existing tests assert on emojis, header or tab markup] → Update them within the phase that changes the markup (phase 2); run `npm test` at the end of each phase.
- [Large diff in `index.css` is hard to review] → Phase 1 only changes token values and neutralizes shadows/gradients/glass; structural CSS lands in phase 2/3 commits separately.
- [Dark-first may break contrast on a few ad-hoc colors] → Hard-coded colors are removed in phase 4's first task; contrast of muted/secondary text verified in phase 1.
- [Sidebar steals 220px, squeezing wide tables (Arr, Playlist)] → Tables keep full width within the content area; sidebar width is a token; verify Arr and Playlist at 768px and 1280px.
- [Error badge adds polling load] → 30s interval, `limit=1`, paused when the document is hidden.
- [Overlap with `downloads-hide-cancelled-and-search` on `DownloadsTab.tsx`] → Finish and archive that change before phase 4 on Downloads.
- [Fonts add bundle weight (~100–150 KB woff2)] → Limited weights/subsets; acceptable for an internal dashboard.

## Migration Plan

Ship as one change with four ordered phases; each phase leaves the app working and tests green, so it can be committed (and, if needed, released) independently.
Rollback: revert the phase commits; theme preference keys in `localStorage` are unchanged, so no user data migration is involved.
The one visible default change (dark for users with no stored preference) is announced in the changelog/README.
