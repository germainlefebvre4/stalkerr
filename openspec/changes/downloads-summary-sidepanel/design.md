## Context

`DownloadsTab.tsx` currently renders each `DownloadEnriched` as a self-contained `download-card` with all detail inline (see proposal.md - Why). The Playlist tab already implements the target pattern: `PlaylistItemsTable.tsx` renders a compact table (desktop) / `mobile-list-card` list (mobile) and `PlaylistTab.tsx` owns a `selectedItem` state plus a Radix `Dialog`-based drawer (`drawer-overlay`/`drawer-content` CSS, already themed and responsive) for full detail. `useDownloads` polls `GET /api/v1/downloads` every 5 seconds while the tab is active and replaces the whole `downloads` array on each tick.

## Goals / Non-Goals

**Goals:**
- Reuse the existing table/mobile-card and `Dialog` drawer patterns/CSS as-is; no new visual language.
- Keep the sidepanel showing live data for an actively downloading item without the user closing/reopening it.
- Preserve all existing Move/Rename behavior and API calls unchanged — only the trigger's location moves.

**Non-Goals:**
- No backend or API changes; no new fields.
- No pagination changes to the Downloads list (still capped at 20 via `api.getDownloads(20, ...)`, unchanged in this change).
- No change to the filter bar (status/type/problem selects) beyond leaving it exactly as-is.

## Decisions

**Selection state: id, not object.** Unlike `PlaylistTab`'s `selectedItem: PlaylistItem | null` (captured once, never re-derived, safe because Playlist never overwrites its array behind an open drawer), Downloads polls every 5s and replaces `downloads` wholesale. Storing the selected item by reference would freeze the drawer's data at click time. Instead `DownloadsTab` SHALL hold `selectedId: number | null` and derive the displayed item each render as `downloads.find(d => d.id === selectedId) ?? null`. When that lookup returns `null` while `selectedId` is still set (item left the list, e.g. a filter change), the `Dialog`'s `open` prop becomes `false` and `selectedId` SHALL be reset to `null` in the same effect — this is what gives "auto-close" for free, no separate close animation state needed.

**New component: `DownloadsSummaryList` (mirrors `PlaylistItemsTable`).** A new file, `frontend/src/components/DownloadsSummaryList.tsx`, takes `downloads`, `loading`, and `onRowClick`, and renders the desktop `custom-table`/`table-flush` markup plus the `isMobile` branch using `mobile-list-card`, following `PlaylistItemsTable.tsx`'s existing branching (`useIsMobile()`) rather than CSS-only responsive hiding — same approach the codebase already uses. `DownloadsTab.tsx` keeps owning the `Dialog` drawer and the filter bar, same division of responsibility as `PlaylistTab.tsx` / `PlaylistItemsTable.tsx`.

**Drawer content ported directly from the current card's conditional blocks.** The existing `download-card` JSX already computes everything the drawer needs (`isLowQuality`, `hasYearIssue`, `hasYearMismatch`, `hasFormatIssue`, `progress`, `filepathBase`); that logic moves into the drawer's render function largely unchanged, just re-hosted. No new derivation logic beyond the `selectedId` lookup above.

**Move/Rename triggers relocate, dialogs do not change.** `onOpenMoveDialog`/`onOpenRenameDialog` props on `DownloadsTab` are unchanged; only the buttons that call them move from the row into the drawer's actions section, gated on `status === 'completed'` as before.

**Error message section reuses the existing gating rule.** `status === 'failed'` gates rendering, verbatim same condition as today's card, just relocated into the drawer.

## Risks / Trade-offs

- [Risk] A poll tick could replace `downloads` with a new array reference while the drawer is open and the user is mid-interaction (e.g. hovering a copy button) → not a real risk here: the drawer has no scroll-reset or focus-loss-sensitive controls, and Radix `Dialog` content re-rendering with fresh props does not remount the dialog itself.
- [Risk] `selectedId` outliving its item without the effect running (e.g. batched update) → mitigate by deriving `open={!!selectedItem}` directly from the lookup result each render (as `PlaylistTab` already does with `open={!!selectedItem}`), not from a separately-tracked boolean, so there is no window where `open` and the lookup disagree.
- [Trade-off] Splitting `DownloadsTab.tsx` into a tab shell + new `DownloadsSummaryList.tsx` is a bit more surface area than editing in place, but matches the existing Playlist split and keeps the row-rendering testable in isolation.

## Migration Plan

Pure frontend UI change behind no feature flag; ships as a normal PR. Existing `DownloadsTab.test.tsx` assertions about the error banner move to opening the drawer first (click the row) before asserting on the message, since the banner is no longer inline in the list.
