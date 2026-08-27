## 1. Date-Grouping Utility

- [x] 1.1 Add a pure helper (e.g. in `frontend/src/utils/date.ts`) that, given the sorted `playlist` array, returns for each index whether it starts a new local-calendar-day group, and verify with unit tests covering: first item always starts a group, consecutive same-day items don't, a day boundary mid-array does.
- [x] 1.2 Add a `getDateGroupLabel(date, t, locale)` helper returning the translated "Today"/"Yesterday" string or `formatDate(date, locale)` for older days, and verify with unit tests covering today, yesterday, and an older date across at least `fr` and `en`.

## 2. Localization

- [x] 2.1 Add `dateGroups.today` and `dateGroups.yesterday` keys to `frontend/src/locales/fr/playlist.json` and `frontend/src/locales/en/playlist.json`, and verify both locale files remain valid JSON and the keys resolve via `t('dateGroups.today')` / `t('dateGroups.yesterday')`.

## 3. Desktop Table Rendering

- [x] 3.1 In `PlaylistTab.tsx`, when `playlistSort === 'created_at'`, compute group boundaries from `playlist` using the 1.1 helper and render a full-width `<tr><td colSpan={7}>` group header immediately before the first row of each new group, and verify visually that today's/yesterday's imports show the relative label and older days show the full date.
- [x] 3.2 Ensure no group headers render when `playlistSort !== 'created_at'`, and verify by switching the table sort to `tvg_name` (or another non-date column) and confirming the table renders as a flat list.
- [x] 3.3 Verify pagination behavior: navigate to a page whose first item continues a group from the previous page, and confirm the group header still renders at the top of that page.

## 4. Mobile Card Rendering

- [x] 4.1 Apply the same grouping logic to the mobile `.mobile-list-card` list, inserting an equivalent group header `<div>` before the first card of each new group, and verify on a mobile viewport that grouping, labeling, non-date-sort suppression, and pagination behave the same as the desktop table (tasks 3.1-3.3).

## 5. Styling

- [x] 5.1 Add a shared, low-emphasis CSS class for the group header (used by both the desktop `<td>` content and the mobile `<div>`) in `frontend/src/index.css`, and verify it renders as a subtle text/hairline divider consistent with the "léger" requirement — no colored badge, dot, or per-row background tint.

## 6. Verification

- [x] 6.1 Run the frontend test suite and confirm it passes, including the new unit tests from tasks 1.1 and 1.2.
- [x] 6.2 Manually exercise the playlist page end-to-end (desktop and mobile) per the `run` workflow: default load (grouped by date), switching sort columns (grouping disabled/re-enabled), and paging through multiple days, confirming behavior matches every scenario in `specs/playlist-import-date-grouping/spec.md`.
