## 1. Season grouping

- [x] 1.1 In `frontend/src/components/RadarrSonarrTab.tsx`, replace the `expandedEpisode: { season, episode } | null` state (currently line 88) with two independent states, `expandedSeason: number | null` and `expandedEpisode: number | null`; update `openSeries`/`closeDrawer` to reset both. Verify `tsc --noEmit` (or `npm run build`) passes in `frontend/`.
- [x] 1.2 Add a `useMemo` that groups `seriesDetail.episodes` into `{ season, matchedCount, totalCount, episodes }[]`, sorted by season then episode number, defensively (not assuming API order). Verify with a quick manual check (or a focused unit test) that a fixture episode list groups into the expected per-season counts.
- [x] 1.3 Render the Séries sidepanel as season sections (rotating chevron, "Saison N" header row showing the matched/total ratio), collapsed by default, using `expandedSeason` as a single-open accordion; selecting a season sets `expandedSeason` and resets `expandedEpisode` to `null`. Verify manually that only one season is expanded at a time and each header shows its own ratio.

## 2. Episode selection clarity and misclick fix

- [x] 2.1 Add an active/expanded visual state to the episode row (conditional class, e.g. `clickable-row--active`, plus a chevron indicator matching the season header's) applied when `expandedEpisode` matches that row. Verify visually that the open episode is distinguishable from collapsed ones at a glance.
- [x] 2.2 Restructure the nested occurrences block (currently lines 507-540) into a framed container — distinct background, border, and indentation from the episode rows above and below it, mirroring `PlaylistGroupedView`'s `renderExpandedSection` treatment — while keeping each occurrence row individually clickable to open the media detail drawer. Verify manually that the boundary between the last occurrence row and the next episode row is visually unambiguous.
- [x] 2.3 Verify manually (desktop and touch/mobile emulation) that clicking an occurrence row inside an expanded episode reliably opens that occurrence's detail rather than toggling the next episode.

## 3. Mobile sub-tab switcher

- [x] 3.1 Give the Résumé/Radarr/Sonarr `Tabs.List` (`RadarrSonarrTab.tsx`, currently line 165) its own class alongside `segmented-tabs-list` (e.g. `radarr-sonarr-subtabs`). Verify the class is present without changing desktop appearance.
- [x] 3.2 Add a mobile-breakpoint override in `frontend/src/index.css` for `.radarr-sonarr-subtabs` that keeps it visible and usable (e.g. horizontally scrollable) on mobile widths, without altering the existing `.segmented-tabs-list` mobile-hide rule used by `App.tsx`'s top-level navigation. Verify manually on a mobile-width viewport that Résumé/Radarr/Sonarr can be switched, and that the app-level bottom tab bar is unaffected.

## 4. Mobile-adapted sidepanel tables

- [x] 4.1 Add an `isMobile` branch for the movie occurrences list (currently lines 446-465) rendering as `mobile-list-card` rows instead of the 3-column table. Verify visually on a mobile-width viewport that no horizontal scrolling is needed to read a row.
- [x] 4.2 Add an `isMobile` branch for the season/episode list (from task 1.3) rendering as `mobile-list-card` rows, preserving the season/episode accordion behavior from tasks 1 and 2. Verify visually on a mobile-width viewport.
- [x] 4.3 Add an `isMobile` branch for an expanded episode's nested occurrences (from task 2.2) rendering as `mobile-list-card` rows. Verify visually on a mobile-width viewport.

## 5. Localization

- [x] 5.1 Add new strings needed for season grouping (e.g. season label, season ratio) to `frontend/src/locales/en/radarrSonarr.json` and `frontend/src/locales/fr/radarrSonarr.json`. Verify the keys are used from `RadarrSonarrTab.tsx` and `npm run build` passes.

## 6. Tests and verification

- [x] 6.1 Update `frontend/src/components/RadarrSonarrTab.test.tsx` to cover: season grouping produces correct per-season matched/total counts, season-level accordion allows only one open season, episode-level accordion allows only one open episode with a visible active state, mobile sub-tab switcher renders and is operable, mobile sidepanel lists render as cards instead of tables. Verify `npm test` passes in `frontend/`.
- [x] 6.2 Manual end-to-end pass in the browser (desktop and mobile viewport) confirming: season grouping and stats read clearly, the open episode is unambiguous, no misclicks occur near an expanded episode's occurrences, the mobile sub-tab switcher works, and mobile sidepanel tables render as list cards without horizontal scroll. Note: coordinate with `radarr-sonarr-tab-persistence-icons` if it has landed in the meantime, since both changes touch the same `Tabs.List`/`Tabs.Trigger` region.
