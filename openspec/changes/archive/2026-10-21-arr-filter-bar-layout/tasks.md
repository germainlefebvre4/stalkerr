# Tasks

## 1. Reset in the hook

- [x] 1.1 Add `resetFilmsFilters` and `resetSeriesFilters` to `frontend/src/hooks/useRadarrSonarr.ts`, each one `patchFilterURLState` call restoring match-status `''`, Suivi `monitored`, Fichier `''` plus `setPage(1)`, and return them from the hook; verify in `useRadarrSonarr.test.ts` that after changing the three filters a reset restores the defaults, resets the page, triggers a single fetch, and leaves the other section and the search term untouched
- [x] 1.2 Add FR/EN keys `filterBar.group.radarr`, `filterBar.group.sonarr`, `filterBar.group.playlist`, `filterBar.correspondance` and `filterBar.reset` to `frontend/src/locales/{fr,en}/radarrSonarr.json`; verify both files stay in sync (same key set) and the existing locale tests, if any, pass

## 2. Shared filter bar component

- [x] 2.1 Reproduce the current layout in the browser (`npm run dev`, Radarr and Sonarr sub-tabs, desktop and mobile widths) and confirm whether the dropdown triggers stretch to full rows because of `.custom-select { width: 100% }`; verify by noting the observation in the task review before the CSS is written
- [x] 2.2 Create `frontend/src/components/ArrFilterBar.tsx`: search line, then a Radarr/Sonarr group (Suivi, Fichier, "Manquants"), a Playlist group (labelled "Correspondance" native select) and a "Réinitialiser" button disabled when `isDefault`; props only, no hook import; verify with `ArrFilterBar.test.tsx` covering group captions, the labelled match-status select, callbacks of every control, the active "Manquants" state, and the reset button enabled/disabled
- [x] 2.3 Add the `.arr-filter-bar`, `.arr-filter-group` and shared `.arr-filter-control` classes to `frontend/src/index.css` (same height/padding/font for triggers, select and buttons, `width: auto` scoped to the bar so `.custom-select` stays untouched elsewhere), plus the mobile media query on the existing breakpoint (full-width search, two-column grid, reset spanning the row); verify visually at desktop and mobile widths in the browser and that the configuration forms using `.custom-select` are unchanged

## 3. Wire it into the tab

- [x] 3.1 Replace the Films and Séries filter markup in `frontend/src/components/RadarrSonarrTab.tsx` with `ArrFilterBar`, deriving `isDefault` per section (match-status `''`, Suivi exactly `{monitored}`, Fichier empty) and passing the reset handlers; verify the inline-styled `<select>`/buttons are gone and `npm run build` type-checks
- [x] 3.2 Update `RadarrSonarrTab.test.tsx` for the new structure (queries by label/role) and add cases for: reset restores the defaults on Films and on Séries without touching the other section, reset leaves the search text, reset disabled at defaults; verify `npm test` passes
- [x] 3.3 Run `npm run lint` and `npm test` in `frontend/`; verify no warning and no failing test, then check in the browser that filtering, "Manquants", URL persistence on refresh and the reset behave as in the spec scenarios on both sections

## Workflow follow-up

- Archive the change after review; the delta in `specs/radarr-sonarr-monitoring-view/spec.md` is then synced into the main spec.
