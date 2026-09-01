## Context

See proposal.md - Why. Two existing endpoints already fetch and sort the *entire* monitored catalog in memory before slicing to the requested page (`internal/api/radarr_sonarr_handlers.go`): `listRadarrMonitoredMovies` and `listSonarrMonitoredSeries`. Match computation is deliberately scoped to only the page's entries - for Sonarr this bounds live per-series Sonarr episode-list calls to the page size (see `radarr-sonarr-monitoring-api` - "Pagination is applied before match computation"). This constraint shapes what the Résumé sub-tab can cheaply show.

`RadarrSonarrTab.tsx` currently renders Films and Séries as two stacked `<section>` blocks inside a single `Tabs.Content value="radarr-sonarr"`. `useRadarrSonarr.ts` owns independent loading/error/pagination state per section, fetching only on explicit triggers (tab activation, pagination change, refresh button) per the existing "Manual refresh per section" requirement - the same discipline extends to search and stats.

## Goals / Non-Goals

**Goals:**
- Reorganize the existing Films/Séries content into three nested sub-tabs without changing their existing table/pagination/sidepanel behavior.
- Add server-side title search to both listing endpoints, reusing the existing full-catalog-then-slice pattern.
- Add a full-catalog matched/unmatched stat for Radarr, cheaply, since matching there only touches the local DB.
- Add a total-only stat for Sonarr, avoiding the per-series Sonarr fan-out the existing pagination design explicitly bounds.

**Non-Goals:**
- No Sonarr matched/unmatched aggregate across the full catalog - would require an episode-list call per monitored series, contradicting the existing bounded-fan-out design. Explicitly out of scope per proposal.md.
- No change to the sidepanel/drawer, per-item detail endpoints, or the in-flight `radarr-sonarr-occurrence-detail-drawer` change's work.
- No caching/persistence of search results or stats - both computed fresh per request, consistent with "No caching or persistence of monitoring results".
- No debouncing/UX polish requirements specified here (left to implementation/tasks); the spec only requires that typing filters and clearing restores the full list.

## Decisions

**Nested `Tabs.Root` inside the existing `Tabs.Content`.** `RadarrSonarrTab.tsx` gets its own `Tabs.Root value={activeSubTab}` wrapping three `Tabs.Content`s (résumé/radarr/sonarr), mirroring the top-level pattern already used in `App.tsx` (`segmented-tabs-list`/`segmented-tabs-trigger` classes reused for visual consistency). Alternative considered: three separate top-level tabs in `App.tsx` instead of nesting - rejected because it would flatten "Radarr/Sonarr monitoring" as a single navigational concept into three unrelated top-level entries, losing the grouping the proposal and existing spec (`Dedicated Radarr/Sonarr monitoring tab`) establish.

**Search is server-side, filtering the full monitored list before the pagination slice.** Both handlers already hold the full `monitored`/`allSeries` slice in memory before pagination; inserting a `strings.Contains(strings.ToLower(title), strings.ToLower(search))` filter there costs nothing extra (no new upstream calls) and keeps `total`/pagination consistent with the filtered set. Alternative considered: client-side filtering of the loaded page - rejected (explored and explicitly rejected with the user) because it only filters 20 already-loaded rows, silently hiding matches on other pages.

**Radarr stats: new endpoint running `matcher.MatchMoviesBatch` over the *entire* monitored list, not just a page.** This intentionally departs from the "pagination before match computation" cost discipline for the *listing* endpoints, but only for this dedicated stats endpoint, and only because Radarr matching is local-DB-only (no per-item upstream call) - so scanning the whole catalog is a single DB query, not an N-call fan-out. Alternative considered: derive the aggregate by having the frontend page through and sum `matched` client-side - rejected as slower (N page requests) and not resilient to concurrent catalog changes mid-walk.

**Sonarr stats: total count only, sourced from the existing monitored-series list length.** No new match computation. Alternative considered: full matched/unmatched aggregate like Radarr - rejected per the user's explicit decision during exploration, since it requires one Sonarr episode-list call per monitored series with no local-only shortcut (episode `Monitored` flags aren't persisted locally).

**Reuse `PaginatedResponse`'s `total` field for filtered counts; no separate "result count" field.** Consistent with how pagination already reports `total` for the unfiltered case.

## Risks / Trade-offs

- [Radarr stats endpoint cost grows with catalog size (single DB batch query over all monitored movies)] -> Acceptable: local DB only, no external fan-out, and the existing per-page matcher call already does the same batch-match logic per page.
- [Nested tabs is a new UI pattern in this codebase] -> Mitigated by reusing the existing top-level tab CSS classes rather than introducing a second styling system.
- [Search filtering happening server-side means every keystroke could trigger a request] -> Left to tasks/implementation to debounce; not a spec-level concern since the spec only requires eventual filtering + restore-on-clear.
