## Context

`PlaylistTab.tsx` currently renders the "Items"/"Films & Séries" pair and the "All"/"Movies"/"TV Shows" pair as two separate rows of pill buttons (`btn-primary`/`btn-secondary`), styled entirely with inline `style={{...}}` plus a few global CSS classes in `index.css`. There is no Tailwind, no icon library, and no existing `Toggle`/`Switch` component anywhere in the frontend — the only "toggle" pattern in the codebase is a pair of buttons with a conditional class. Icons elsewhere in the app are plain unicode/emoji characters inline in JSX (`🔍`, `⚙️`, `▾`, `▸`), and tooltips are always the native `title` attribute (Radix Tooltip is not installed). Mobile/desktop branching in this exact file already uses a `useIsMobile()` boolean to pick between two JSX blocks (see the advanced-filters section), rather than CSS media queries. See proposal.md for the motivation.

## Goals / Non-Goals

**Goals:**
- Introduce a minimal, dependency-free toggle switch component styled to match the app's existing visual language (CSS variables already used by `.btn-primary`/`.btn-secondary`).
- Merge the two selector rows into one without changing `playlistView`/`playlistFilter` state shape or data-fetching behavior.
- Keep the merged row on one line at common mobile widths by shrinking the filter buttons, not by shrinking or hiding the switch.

**Non-Goals:**
- No generalized/reusable `<Switch>` component library for the rest of the app — this is a single-purpose element scoped to this row. If a second use case appears later, it can be extracted then.
- No change to `playlistFilter` (content-type) semantics, ordering, or the underlying `/api/v1/items` vs `/api/v1/items/grouped` data flow.
- No new icon library or design system dependency.

## Decisions

**Switch markup: native checkbox + label, not a from-scratch `div`/ARIA widget.**
Use `<label>` wrapping a visually-hidden `<input type="checkbox" checked={playlistView === 'grouped'} onChange={...} />` plus styled `<span>` track/knob elements. This gets keyboard focus, spacebar toggling, and screen-reader checked-state semantics for free, matching how `ManualOverrideDialog.tsx` already uses a native checkbox elsewhere. Alternative considered: a `role="switch"` `<button>` — rejected as more ARIA plumbing to hand-roll for no behavioral gain over a native checkbox.

**Icons: plain unicode/emoji flanking the track, no icon library.**
`☰` (list) on the "Items" side, `🎬` (clapperboard) on the "Films & Séries" side, both `aria-hidden` since the accessible name comes from the checkbox's `title`/label, not the decorative icons. This follows the project's existing emoji-as-icon convention rather than introducing `lucide-react` or similar.

**Tooltip: native `title` attribute on the switch's `<label>`, dynamic per current state.**
Text describes the *destination* state (e.g. "Basculer vers Films & Séries" when currently on Items), consistent with the project's exclusive use of native `title` for tooltips. New i18n keys (e.g. `view.switchToGrouped` / `view.switchToItems`) replace the now-unused `view.items` / `view.grouped` button-label strings in `frontend/src/locales/{en,fr}/playlist.json`.

**Row layout: single flex row, filter-button group as one flex child, switch as a sibling with `marginLeft: 'auto'`.**
The three content-type buttons stay wrapped in their own inner `flex`/`wrap` div (unchanged internally), which becomes one flex item in the outer row; the switch is the second flex item with `marginLeft: 'auto'` so it's pushed right regardless of how much space the button group takes, and stays right-aligned even if the row wraps onto two lines on an unexpectedly narrow viewport. Alternative considered: `justify-content: space-between` on the outer row — rejected because a wrapped single-item row resolves to flex-start under `space-between` in most browsers, which would left-align the switch instead of keeping it right-aligned as specced.

**Mobile compacting: `isMobile`-conditional inline styles, not a CSS media query.**
Reduce the filter buttons' `padding` (e.g. `0.45rem 1rem` → `0.35rem 0.6rem`) and the row's `gap` (e.g. `0.5rem` → `0.35rem`) when `useIsMobile()` is true, mirroring the existing conditional-JSX pattern already used a few lines below in the same component for the advanced-filters block, rather than adding a new `@media` block in `index.css` disconnected from that pattern. The switch itself keeps a fixed, already-compact size on both breakpoints.

## Risks / Trade-offs

- [Icon-only switch loses the explicit text labels "Items"/"Films & Séries"] → Mitigated by the dynamic `title` tooltip and the distinct flanking icons; acceptable per explicit user direction (icons/tooltip only, no visible label).
- [`360px`-wide phones are unusually narrow; padding/gap reduction alone might not be enough if translated button labels are long in some locale] → If it doesn't fit for a given locale's longer strings, the outer row's existing `flexWrap: 'wrap'` still provides a wrap fallback with the switch staying right-aligned on its own line (per the row-layout decision above), so there is no hard breakage, only a graceful two-line fallback.
- [First custom toggle switch in the codebase, no established visual pattern to copy exactly] → Styled directly from the same CSS custom properties (`--accent-gradient`, `--border-color`, `--radius-sm`, etc.) already driving `.btn-primary`/`.btn-secondary`, so it reads as part of the same design system rather than a one-off.
