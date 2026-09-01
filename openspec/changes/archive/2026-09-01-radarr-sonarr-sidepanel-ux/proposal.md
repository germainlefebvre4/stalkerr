## Why

The Sonarr sidepanel renders every monitored episode of a series as a single flat table, with no season grouping or per-season stats, so a series with several seasons becomes a long wall of near-identical rows. Selecting an episode gives no visible "this one is open" state beyond the nested occurrences table appearing below it, and that nested table shares the same clickable-row styling and sits directly against the next episode row, so users frequently click the next episode instead of an occurrence inside it. Separately, on mobile viewports the Radarr/Sonarr tab's own Résumé/Radarr/Sonarr sub-tab switcher is hidden by a CSS rule shared with the app's top-level navigation (which has a bottom-bar replacement) but has no mobile replacement of its own, so mobile users get stuck on whichever sub-tab loaded first; and every table shown inside the sidepanel (movie occurrences, episode list, episode occurrences) always renders as the fixed 3-column desktop table, which is too wide for a mobile drawer and forces horizontal squeeze/scroll.

## What Changes

- Sonarr sidepanel episodes are grouped into collapsible season sections (collapsed by default), each showing a matched/total episode ratio.
- Season sections and episode rows both behave as single-open accordions (opening one closes any other at the same level), consistent with the episode-level behavior already in place.
- The currently-expanded episode row is visually distinguished from other rows (not just by the nested table appearing below it).
- The nested occurrences table shown inside an expanded episode is visually separated from the episode row above and the next episode row below (indentation/spacing/distinct styling, not the same `clickable-row` treatment), to stop accidental clicks on the next episode.
- On mobile viewports, the Radarr/Sonarr tab's Résumé/Radarr/Sonarr sub-tab switcher becomes reachable again (fixes a bug where it is hidden with no replacement control).
- On mobile viewports, the sidepanel's tables (movie occurrences, episode list, episode occurrences) render as a mobile-appropriate list layout instead of the fixed 3-column desktop table.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `radarr-sonarr-monitoring-view`: extends "Radarr/Sonarr tab organizes content into three sub-tabs" so the sub-tab switcher remains usable on mobile; extends "Selecting a series shows per-episode breakdown" so episodes are grouped by season with per-season stats; extends "Séries episode rows expand to reveal their occurrences" so the expanded episode is visually distinguishable and its nested occurrences table cannot be mistaken for adjacent episode rows; extends "Sidepanel shows matched playlist occurrences for a selected item" so sidepanel tables adapt to mobile viewports.

## Impact

- `frontend/src/components/RadarrSonarrTab.tsx`: group episodes by season (client-side, from the already-fetched `SonarrSeriesEpisodesResponse`), add collapsible season sections with matched/total stats, add a visible selected-state to the active episode row, restructure the nested occurrences table markup to avoid visual adjacency with the next episode row, add a mobile-reachable sub-tab switcher, add mobile-appropriate rendering for the sidepanel's occurrence/episode tables (mirroring the existing `mobile-list-card` pattern already used for the top-level Films/Séries lists).
- `frontend/src/index.css`: new/adjusted styles for season group headers, the selected-episode state, the nested-occurrences container, and the mobile sub-tab control.
- `frontend/src/locales/en/radarrSonarr.json`, `frontend/src/locales/fr/radarrSonarr.json`: new strings for season grouping (label, stats).
- No backend or API changes; all data needed for season grouping is already present in the existing `SonarrSeriesEpisodesResponse` payload.
- Coordination note: `openspec/changes/radarr-sonarr-tab-persistence-icons` (in progress, 0/7 tasks done) also modifies the "three sub-tabs" requirement and the same `Tabs.List`/`Tabs.Trigger` markup in `RadarrSonarrTab.tsx` (URL persistence + solution icons). Both changes touch overlapping lines; implement/merge with that in mind to avoid conflicting edits.
