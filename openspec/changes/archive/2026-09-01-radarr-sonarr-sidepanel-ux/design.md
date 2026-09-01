## Context

See proposal.md for motivation. Two things in the current code shape this design:

- `PlaylistGroupedView.tsx` already implements the exact interaction this change needs elsewhere in the app: a single-open accordion row with a rotating chevron (`▸`), and an expanded section rendered with a visually distinct background/border rather than as another row of the same table. That's the closest prior art in this codebase.
- The mobile sub-tab bug is a CSS class collision: `.segmented-tabs-list { display: none; }` (mobile breakpoint) is shared between `App.tsx`'s top-level nav (which has a `.mobile-tab-bar` replacement) and `RadarrSonarrTab.tsx`'s own Résumé/Radarr/Sonarr sub-tabs (which have no replacement). The fix must not touch the app-level nav's mobile behavior.

## Goals / Non-Goals

**Goals:**
- Season grouping, selection clarity, and the misclick fix for the Sonarr episode sidepanel.
- Restore sub-tab switching on mobile, scoped to `RadarrSonarrTab.tsx`.
- Mobile-appropriate rendering for sidepanel tables (movie occurrences, episode list, episode occurrences).

**Non-Goals:**
- No backend/API changes.
- No changes to the app's top-level navigation or its mobile bottom bar.
- No changes to the top-level Films/Séries list mobile rendering (already has a working `mobile-list-card` treatment).
- Does not implement URL-based sub-tab persistence or solution icons — that's the separate, already in-progress `radarr-sonarr-tab-persistence-icons` change.

## Decisions

**Accordion state shape**: replace the single `expandedEpisode: {season, episode} | null` state with two independent single-value states, `expandedSeason: number | null` and `expandedEpisode: number | null` (episode number alone, since only one season can be open at a time). Selecting a season resets `expandedEpisode` to `null`. This keeps both accordions (season-level, episode-level) simple and independent rather than nesting them into one compound key.

**Visual pattern for season sections**: reuse `PlaylistGroupedView`'s established pattern (rotating chevron, header row with matched/total stat, expanded content in a visually distinct background/border block) instead of inventing a new visual language for grouping in this codebase.

**Selected-episode visibility**: give the expanded episode row an explicit active state (conditional class, e.g. `clickable-row clickable-row--active`, plus the same chevron treatment used for seasons) so it reads as "open" independent of the content that appears below it.

**Misclick fix**: stop rendering the nested occurrences as another `<tr>` in the same table with the same `clickable-row` styling as episode rows. Render them in a framed block (indentation + distinct background + padding + border, mirroring `PlaylistGroupedView`'s `renderExpandedSection`) so they read as content *inside* the open episode rather than a continuation of the episode list, and are no longer visually adjacent to the next episode row without a clear boundary.
- Alternative considered: just add spacing (margin) before the next episode row without changing the nested table's styling. Rejected — it would reduce accidental clicks somewhat but wouldn't fix the underlying confusion of two visually identical row types stacked in the same list.

**Mobile sub-tab switcher**: give `RadarrSonarrTab.tsx`'s `Tabs.List` its own class (e.g. `radarr-sonarr-subtabs`) instead of relying solely on the shared `.segmented-tabs-list`, and add a mobile-breakpoint override for that class so it stays visible/usable on mobile (horizontally scrollable pill row), without touching `.segmented-tabs-list`'s existing mobile-hide rule used by the app-level nav.
- Alternative considered: give it a `.mobile-tab-bar`-style fixed bottom bar of its own. Rejected — a second fixed bottom bar would visually compete with the app's existing one.

**Mobile sidepanel tables**: reuse the existing `mobile-list-card` pattern (already used for the top-level Films/Séries lists in this same file) for movie occurrences, season/episode rows, and episode occurrences, gated by the `useIsMobile()` hook already in use here — consistent with the rest of the component instead of a new pattern.

**Coordination with `radarr-sonarr-tab-persistence-icons`**: that change (in progress, 0/7 tasks) also modifies the "three sub-tabs" requirement and the same `Tabs.List`/`Tabs.Trigger` region for URL persistence and solution icons. This change does not implement that scope. Whichever change is implemented first should land before the other starts on this file, to avoid conflicting edits to the same lines.

## Risks / Trade-offs

- [Risk] Overlapping edits with `radarr-sonarr-tab-persistence-icons` on `Tabs.List`/`Tabs.Trigger` in `RadarrSonarrTab.tsx` → Mitigation: sequence implementation (land one change before starting the other's edits to that region), per the Decisions note above.
- [Risk] Reframing the nested occurrences table changes its visual weight relative to the Films occurrences table → Mitigation: keep the same badges/typography/columns; only change the surrounding container (background, indentation, spacing), not the data presentation itself.
- [Risk] Client-side season grouping assumes `SonarrSeriesEpisodesResponse.episodes` can be grouped and sorted without a guaranteed API order → Mitigation: group and sort defensively client-side (by season, then episode) rather than assuming API order.
