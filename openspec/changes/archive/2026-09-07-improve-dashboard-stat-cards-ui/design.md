## Context

See `proposal.md` — Why / What Changes. The relevant current-state facts that shape this design:

- `HomeTab.tsx` and `RadarrSonarrTab.tsx` (its `resume` sub-tab) both render `<section className="home-card">` blocks inside a `<div className="home-grid">`, using the exact same global CSS classes from `frontend/src/index.css` (`.home-grid`, `.home-card`, `.home-card-title`, `.home-card-subheading`, `.home-fields`, `.home-card-loading`, `.home-empty-state`, `.home-unavailable`, `.home-error`). No other component in the codebase references these classes.
- The app already has: brand icons (`assets/icons/radarr.svg`, `sonarr.svg`), a badge system (`.badge`, `.badge-success/progress/failed/pending/neutral`), a Radix-based progress bar (`.progress-root`/`.progress-indicator`, used today in `DownloadsSummaryList.tsx`), and a compact `.status-dot` variant explicitly built for mobile status indicators (currently unused by any component).
- Mobile breakpoint is `(max-width: 767.98px)`, defined once (`useIsMobile`/`MOBILE_BREAKPOINT_QUERY`) and mirrored in a CSS `@media` query — both must stay in sync if touched.
- No icon library is currently a dependency; only hand-authored brand SVGs exist.

## Goals / Non-Goals

**Goals:**
- Extend `.home-card`/`.home-grid` (and closely related classes) once, so `HomeTab.tsx` and `RadarrSonarrTab.tsx`'s Résumé sub-tab both pick up the same visual language automatically.
- Reuse existing visual primitives (badges, progress bar, brand icons) rather than inventing new ones, except for the one deliberate addition: a generic icon set for cards that have no brand icon.
- Keep all data-fetching, hooks, and API contracts unchanged — this is a presentation-layer change plus one new local UI interaction (group-titles disclosure).

**Non-Goals:**
- No backend/API changes.
- No change to which data is fetched or how errors/loading states are computed (`useHomeDashboard`, `useRadarrSonarr`, etc. are untouched).
- No redesign of other `.home-grid`-adjacent-but-unrelated patterns (`.filter-grid`, `.filter-card`) — those are visually similar but structurally separate and out of scope.
- No dark-mode introduction (the app has no dark theme today; not being added here).

## Decisions

**1. Icon library: `lucide-react`.**
Chosen over hand-drawn SVGs because the app needs several distinct generic icons (Last Run/clock, Catalog/library, Downloads & Errors) with consistent stroke weight and sizing, and over a heavier icon-font/kit because `lucide-react` is tree-shakable (each icon is its own ES export, so bundle impact is limited to the ~4-5 icons actually imported) and has no runtime CSS dependency, matching the app's plain-CSS styling approach.

**2. Status indicator: existing `.badge-success`/`.badge-failed` on desktop, existing `.status-dot` on mobile — no new component.**
Both classes already exist and already carry the right semantic colors via `variables.css` status tokens. The card only needs to pick which one to render based on `useIsMobile()`, the same pattern `DownloadsSummaryList.tsx` already uses to branch between a full badge (desktop table) and a more compact mobile treatment.

**3. Hero metric gets a new class (`.home-card-hero` or similar), not a reuse of `.home-card-title`/`.home-fields`.**
`.home-fields` is deliberately uniform (one font-size for all rows, including its mobile override). Overloading it for the hero number would force every row in every card to share the hero's larger size. A dedicated class keeps the secondary rows exactly as they render today and only changes the one promoted value per card.

**4. Progress bar reuse: same `Progress.Root`/`Progress.Indicator` + `.progress-root`/`.progress-indicator` used by `DownloadsSummaryList.tsx`, not a new bar component.**
The existing bar is already responsive (`width: 100%`) and already carries the app's gradient/shimmer styling. Reusing it verbatim keeps the two progress-bar instances in the app visually identical.

**5. Group-titles disclosure: local component state (`useState<boolean>`), not persisted.**
The proposal's new interaction only needs to survive as long as the card is mounted — there is no cross-session or cross-tab reason to remember "expanded" (the underlying run itself is transient — a new run replaces the old summary on the next completed processing run). This matches the project's existing convention of *not* persisting purely cosmetic UI toggles (e.g. the mobile playlist advanced-filters disclosure explicitly resets to collapsed on every mount, per `frontend-ihm-dashboard`'s "Mobile Collapsible Advanced Filters" precedent).

**6. Résumé sub-tab's simpler cards (Sonarr: monitored count only, no ratio) still consume the same base classes.**
The hero-metric/status-badge/icon treatment applies to every card; the progress bar only renders when a card has a meaningful ratio to show (Radarr matched/monitored, Catalog download success rate). A card without a ratio (Sonarr, both in Home and in Résumé) simply omits the progress-bar element — no variant class needed, the component-level JSX just doesn't render one.

## Risks / Trade-offs

- **[Risk] Restyling shared classes touches two pages at once — a mistake affects both simultaneously.** → Mitigation: because the blast radius is fully known (only `HomeTab.tsx` and `RadarrSonarrTab.tsx` reference these classes — confirmed by search), manual verification of both pages after the CSS change is sufficient; no other page can silently regress.
- **[Risk] New `lucide-react` dependency adds a small amount of bundle size.** → Mitigation: import only the specific icons needed (named ES imports), not the full icon set; this keeps tree-shaken bundle impact to a few KB.
- **[Risk] New i18n strings needed (e.g. "+N autres" disclosure label, any new status-badge text) must be added to both `en` and `fr` locale files (`home.json`, `radarrSonarr.json`) or the UI falls back to raw keys.** → Mitigation: tasks.md will explicitly enumerate the locale keys to add per locale file.
- **[Trade-off] Mobile uses `.status-dot` (no text) instead of the full badge, trading explicit status text for horizontal space.** → Accepted: the dot's color alone (green/red) is legible at a glance, consistent with how `.status-dot`'s doc comment already describes its intended mobile use; the desktop view retains the full text badge for users who want it.
