## Context

`PlaylistTab.tsx` renders items returned by `usePlaylist()` (backed by `GET /api/v1/items`, already sorted/paginated server-side per `playlist-item-sorting`). Desktop renders a `<table>` with one `<tr>` per item; mobile renders one `.mobile-list-card` per item. Both currently render items as a flat list with no grouping. See `proposal.md` for motivation.

## Goals / Non-Goals

**Goals:**
- Compute day-based grouping client-side from the items array already present in state, for both the desktop table and mobile card list.
- Keep the change additive and localized to rendering — no new API params, no new state beyond what's derivable from existing `playlist` / `playlistSort` / `playlistOrder`.

**Non-Goals:**
- Grouping/labeling behavior for `downloaded_at` or any column other than `created_at`.
- Sticky/floating group headers while scrolling.
- Persisting or configuring the grouping behavior (e.g. no user toggle to disable it).

## Decisions

**Compute grouping in the render layer of `PlaylistTab.tsx`, not in `usePlaylist.ts`.** The grouping is a pure function of `playlist` + `playlistSort` + `playlistOrder`, all already returned by the hook. Keeping it in the component avoids widening the hook's public contract or adding derived state that must be kept in sync with fetches.

**Represent a "day boundary" as a lookup built once per render pass, not per item.** Walk the already-sorted `playlist` array once, and for each index record whether it starts a new calendar day (comparing local calendar day of `created_at` against the previous item, or "no previous item" for index 0 — which always starts a group when `playlistSort === 'created_at'`). This is O(n) over the current page (max 100 items per `usePlaylist`'s `VALID_LIMITS`), computed inline in the render — no memoization needed at this scale.

**Label resolution is a small pure helper, not a component.** A function `getDateGroupLabel(date: Date, t: TFunction): string` returns the translated "Today"/"Yesterday" string or falls back to `formatDate(date, locale)` for older days — reusing the exact formatting already used for the date cells, so the header and the cell below it never disagree on date format.

**"Today"/"Yesterday" comparison uses the viewer's local calendar day**, via `new Date()` and comparing year/month/day components — consistent with how `created_at` is already displayed to the user (each `formatDate` call implicitly renders in the browser's local time zone). No server/client timezone reconciliation is needed since this is purely a display-layer grouping on top of an already-fetched value.

**Alternative considered — separate `<thead>`-style section per group:** rejected. Restructuring the `<table>` into multiple `<tbody>` groups (one per day) would work but complicates the existing single-`<tbody>` map and the `colSpan` loading/empty states for no real benefit over inserting a header `<tr>` inline before the relevant row.

**Desktop header row:** a full-width `<tr>` with a single `<td colSpan={7}>` inserted immediately before the first `<tr>` of a new group, styled as a thin, muted, non-interactive divider (small text + a hairline border), not a colored badge — consistent with the "léger" requirement and distinct from the existing `.badge-*` and `.pulse-dot` treatments used for pipeline state.

**Mobile header:** a small `<div>` with the same label, inserted before the first `.mobile-list-card` of a new group, styled consistently with the desktop header (shared CSS class) but without a `<tr>`/`colSpan` wrapper.

## Risks / Trade-offs

**[Risk] Grouping silently disappears when the user sorts by another column, which could read as "the feature broke" rather than "intentionally not applicable".** → Mitigation: this is the explicitly agreed behavior (see proposal); no indicator is added for the disabled state to keep the change minimal, matching the "léger" requirement.

**[Risk] A day boundary computed only from the current page's items could show two separate "Yesterday" headers if, hypothetically, items on the same day were non-contiguous within a page.** → Mitigation: cannot happen under this design, since grouping is only active when `sort === 'created_at'`, which guarantees same-day items are always contiguous within a server-sorted page.

## Migration Plan

Purely additive frontend change; no data migration, no feature flag. Ships in the normal frontend build/deploy. Rollback is a plain revert of the touched files.
