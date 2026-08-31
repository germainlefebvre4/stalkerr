## Context

See `proposal.md` - Why/What Changes for motivation and scope.

Relevant existing building blocks this change builds on, without modifying:
- `internal/external/radarr.Client` / `internal/external/sonarr.Client`: today only expose `GetMissingMovies`/`GetMissingEpisodes` (Radarr's `wanted/missing`, Sonarr's `wanted/missing`) plus single-item detail lookups. Neither exposes "all monitored items regardless of file/missing status" as a single call.
- `internal/matcher`: `MatchMovieByTVDB`/`MatchTVShowByTVDB` (and their `...ByTMDB` fallbacks) already implement the TVDB-primary -> TMDB -> fuzzy title/year (or title/season/episode) matching used by the `stalkeer radarr`/`stalkeer sonarr` CLI commands. These functions couple the match lookup to a *download-candidate* query that restricts `ProcessedLine.state IN (processed, failed)` - correct for "what can I still download" but wrong for this feature, which must also recognize already-`downloaded` occurrences as a match.
- `internal/api`: existing proxy-style handler (`searchTMDBProxy`) and pagination helper (`parsePagination`) are the established patterns for a live external-API-backed, paginated endpoint.
- `frontend/src/components/PlaylistTab.tsx`: established patterns for a two-section tab, per-section loading/pagination state, and a Radix `Dialog` sidepanel - reused visually rather than reinvented.

## Goals / Non-Goals

**Goals:**
- Bound the cost of a single request to these new endpoints by page size, not catalog size - in particular the Sonarr per-series episode lookups and the local DB matching lookups.
- Keep the view's match status provably consistent with what the CLI download commands would actually do, by sharing the same underlying matching rules (TVDB -> TMDB -> fuzzy), not a parallel reimplementation.
- Keep this strictly read-only and additive: no new tables, no scheduled job, no change to existing CLI or matcher behavior used by the download path.

**Non-Goals:**
- No local caching/sync table for Radarr/Sonarr data (explicitly rejected in favor of live-on-demand, per proposal).
- No title/text search or filter controls on the new tab in this iteration - pagination only. Revisit if browsing by page proves painful in practice.
- No write actions from this view (no triggering downloads, no editing Radarr/Sonarr monitoring flags) - that already exists elsewhere (force-download from the Playlist sidepanel).
- No support for other *arr applications (Lidarr, Readarr, etc.).

## Decisions

### 1. New client methods fetch the full monitored list once; pagination and matching happen in Stalkeer, not in Radarr/Sonarr
Radarr's movie list and Sonarr's series list endpoints have no server-side pagination (unlike `wanted/missing`, which does). So:
- Add `radarr.Client.GetAllMovies` (or equivalent) hitting `/api/v3/movie`, returning every movie's lightweight metadata (id, title, year, tvdbId, tmdbId, monitored, hasFile, path) in one call.
- Reuse/adjust the existing all-series fetch already inside `sonarr.GetMissingSeries` (it already calls `getSeries` for everything before filtering) to expose an all-monitored-series variant without the missing-episodes filter.
- The handler slices this in-memory list to the requested page (sorted by title) *before* doing any matching work. This is the mechanism that satisfies "avoid big requests" while staying live: the expensive part (DB matching, Sonarr per-series episode calls) only ever runs against `limit` items.

Alternative considered: ask Radarr/Sonarr for pages directly. Rejected - neither's "all items" endpoint supports it, so we'd still need a full fetch to know total count/sort order; doing so explicitly once and slicing ourselves is simpler and no more expensive.

### 2. Batch the matching lookups per page instead of per item
Naively calling the existing per-item matcher functions in a loop over a page would issue O(page size) DB round-trips for movies, and for series, O(page size x episodes per series) DB round-trips for the aggregate - the latter could be thousands of queries for a single page of long-running shows.
- **Movies**: for the current page's Radarr movies, issue one batched DB query (`tvdb_id IN (...) OR tmdb_id IN (...)`) to load all potentially-matching local `Movie` rows at once, index them in memory by TVDB/TMDB id, then only fall back to the existing per-title fuzzy path for movies that missed the batch (small remainder).
- **Series aggregate**: for each series in the page, fetch its monitored episodes from Sonarr (one call per series - unavoidable, bounded by page size per Decision 1), then issue **one** DB query loading all local `TVShow` rows for that series' TVDB id (all seasons/episodes at once), and compute the matched/monitored ratio in memory against the monitored-episode list, instead of one query per episode.
- Both batched lookups and the sidepanel's per-item detail lookup share the same normalization/title-similarity core already in `internal/matcher`, so the state-agnostic variant added for this feature cannot silently drift from the download-path matching rules - only the `ProcessedLine.state` filter differs (dropped entirely here vs. restricted to `processed`/`failed` in the download-candidate functions).

### 3. Fetch a page's Sonarr per-series episode lists concurrently, bounded by page size
Page size already bounds the *count* of extra Sonarr calls (Decision 1), but issuing them sequentially would make a full page's latency scale linearly with page size. Issue them concurrently (bounded by the page size itself, which is already small/capped) to keep response time reasonable without increasing the total call count.

### 4. Match status is independent of playlist pipeline state
"Matched" means a local `Movie`/`TVShow` record exists (via TVDB/TMDB/fuzzy) - it does not require an eligible-for-download (`processed`/`failed`) occurrence. A movie whose only playlist occurrence is already `downloaded` must still show as matched; otherwise the view would misreport already-completed items as gaps. This is why new matcher helpers are needed rather than reusing `FindMovieDownloadCandidates`/`FindTVShowDownloadCandidates` as-is (those remain unchanged for the CLI's use case).

### 5. New dedicated frontend tab, not a Playlist tab extension
Confirmed with the user: a separate "Radarr/Sonarr" tab keeps the existing Playlist tab's filters/state untouched, and lets each section (Films/Séries) own independent loading, pagination, and error state - which the per-section error-isolation requirement needs anyway. Visual patterns (Radix `Tabs`/`Dialog`, drawer sidepanel, per-section pagination) are copied from `PlaylistTab.tsx` rather than introducing a new UI pattern.

### 6. Manual refresh only, per section
No polling interval, no shared "refresh all" action - each section's own explicit trigger (including its initial mount) is the only way it fetches. This keeps the live-fetch cost fully user-driven and matches the "éviter les grosses requêtes" constraint: nothing fetches without an explicit user action.

## Risks / Trade-offs

- **[Risk]** A single very large Radarr/Sonarr catalog (many thousands of items) still means one large "list everything" call before slicing → **Mitigation**: this call returns lightweight metadata only (no per-item enrichment), which is what Radarr/Sonarr's own UIs already do; accepted as the cost of choosing live-over-cached.
- **[Risk]** Sonarr aggregate computation for a page still issues up to `limit` live calls to Sonarr on every page view/refresh → **Mitigation**: bounded by page size (Decision 1), executed concurrently (Decision 3), and only on explicit refresh (Decision 6) - never automatic.
- **[Risk]** Two similar-but-different matching code paths (state-filtered download-candidate vs. state-agnostic monitoring-view) could drift apart over time → **Mitigation**: share the core matching/normalization logic (Decision 2); only the state filter differs, kept as a small, clearly-named difference rather than duplicated matching logic.
- **[Risk]** Without search/filter, browsing a very large catalog to find one title means paging through many pages → **Mitigation**: accepted as a non-goal for this iteration (see Non-Goals); revisit if it proves painful.

## Migration Plan

Purely additive: new backend endpoints, new matcher helpers, new frontend tab. No database migration, no config changes (existing `cfg.Radarr`/`cfg.Sonarr` reused), no changes to existing endpoints or CLI behavior. Deploy as a normal release; rollback is a normal revert with no data cleanup required.

## Open Questions

- Whether a title/text search filter should be added to these sections in a later iteration if pure pagination proves too slow to browse in practice (deferred; does not change this change's specs or approach).
- Whether live-fetch performance in practice (large Radarr/Sonarr libraries, slow Sonarr per-series calls) eventually justifies revisiting the live-vs-cached decision; not a concern this change needs to resolve now.
