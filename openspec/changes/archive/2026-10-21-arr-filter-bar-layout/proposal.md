# Proposal

## Why

`arr-file-state-axes` added the Suivi and Fichier filters next to the existing match-status filter on the arr-suite Films and Séries sections, and the filter bar is now disorganised: a native `<select>` ("Tous", no label), two dropdowns and a preset button are laid out in one `flex-wrap` row with three different styles. `.custom-select` carries `width: 100%`, so each dropdown tends to take a full row of its own. Suivi, Fichier and "Manquants" form one coherent group, the match-status filter is a different concern (link with the playlist), and nothing shows that, nor offers a way back to the default view once several filters are set.

## What Changes

- Reorganise the filter bar of both the Films and the Séries sections into two levels:
  - the title search field alone on its own line;
  - below it, labelled filter groups: **Radarr/Sonarr** (Suivi, Fichier, "Manquants") and **Playlist** (match-status, now labelled "Correspondance"), followed by a reset button.
- Give every control in the bar the same look and height (the match-status native `<select>` is replaced by the same trigger style as the multi-select dropdowns' or styled alike), and fix the full-width stretching of the dropdown triggers.
- Add a **Réinitialiser** button on both sections that restores Suivi = Surveillé, Fichier = no selection, match-status = "Tous", in one URL update (one fetch), keeping the page at 1. It does not touch the search field. It is disabled when the three filters already hold their default values.
- On mobile, the search field takes the full width and the filter groups wrap into a two-column grid.
- Extract the duplicated Films/Séries filter markup into one shared `ArrFilterBar` component driven by CSS classes instead of inline styles.
- No change to the API, to filter semantics (OR within an axis, AND across axes), to defaults, or to URL parameters.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `radarr-sonarr-monitoring-view`: the filter bar is organised into labelled groups with the search field on its own line, and a "Réinitialiser" control restores the default filters.

## Impact

- Frontend only: `frontend/src/components/RadarrSonarrTab.tsx` (filter markup replaced by `ArrFilterBar`), new `frontend/src/components/ArrFilterBar.tsx`, `frontend/src/hooks/useRadarrSonarr.ts` (`resetFilmsFilters` / `resetSeriesFilters`, same pattern as `setFilmsMissing`), `frontend/src/index.css`, `frontend/src/locales/{fr,en}/radarrSonarr.json`, and the matching tests.
- No backend, API, database or URL schema change. Existing bookmarked URLs keep working.
