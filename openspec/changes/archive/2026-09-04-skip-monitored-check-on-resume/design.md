## Context

See `proposal.md` for motivation. Two existing facts shape the approach:

- `mergeIncompleteDownloads` (`internal/scheduler/build.go`) runs inside `BuildStreams`, called from `cmd/download.go`. It only ever sees local DB state (`StateManager.GetIncompleteDownloads`, `models.DownloadInfo.ProcessedLines[0].Movie`/`TVShow`) — it has no Radarr/Sonarr client access of its own beyond what `BuildDeps` already exposes (`RadarrClient`/`SonarrClient` interfaces limited to `GetMissingMovies`/`GetMissingEpisodes`/`GetMovieByTMDBID`/`GetSeriesByTVDBID`/`GetSeriesDetails` — none of which return the full library in one call).
- `cmd/download.go`'s `reconcileDownloadPaths` (added by `reconcile-radarr-sonarr-download-paths`) already fetches the entire Radarr/Sonarr library once per run — before `scheduler.BuildStreams` is even called — and already builds `map[tmdbID]string` (movie path) / `map[tvdbID]string` (series path) from the response. Both `radarr.Movie` and `sonarr.Series` also carry a `Monitored bool` field.
- Correction found during implementation: `sonarr.GetAllMonitoredSeries` (the call `reconcileDownloadPaths` used for series) already filters its result to `Monitored == true` client-side before returning. A series absent from that response is therefore indistinguishable between "not in Sonarr at all" and "in Sonarr but unmonitored" — exactly the distinction this change needs. `radarr.GetAllMovies` has no equivalent problem; it already returns every movie regardless of monitored status. Fixed by adding `sonarr.GetAllSeries` (an unfiltered mirror of `GetAllMovies`) and switching `reconcileDownloadPaths` to call it instead — `GetAllMonitoredSeries` itself is untouched (still used, filtered, by `listSonarrMonitoredSeries`). This also means path reconciliation's own series-path map now covers unmonitored-but-present series too, a superset of its previous behavior; `reconcile-radarr-sonarr-download-paths`'s eligibility spec matches purely on TMDB/TVDB id presence in Radarr's/Sonarr's own data, never on monitored status, so this is not a spec violation, just a broader match than before.

## Goals / Non-Goals

**Goals:**
- Stop `mergeIncompleteDownloads` from resuming an incomplete movie/series download once that movie/series is confirmed unmonitored in Radarr/Sonarr.
- Add zero new Radarr/Sonarr API calls: reuse the exact library fetch `reconcileDownloadPaths` already performs, rather than a second full-library fetch or a per-item lookup at merge time.
- Keep `BuildDeps` backward compatible: existing callers/tests that don't populate the new field keep today's resume-everything behavior unchanged.

**Non-Goals:**
- Tier-2 (already-downloaded, upgrade-eligible) streams (`buildTier2MovieStreams`/`buildTier2SeriesStreams`) also don't check monitored status today — that is a separate, pre-existing gap and out of scope for this change (per proposal.md's stated scope).
- No change to how an incomplete download is detected, to stale-lock cleanup, or to any other `BuildStreams` step.
- No change to `reconcile-radarr-sonarr-download-paths`'s own path-correction behavior — this only adds a second consumer of the same fetched data.

## Decisions

**Thread the already-fetched monitored status through `BuildDeps` as a nil-safe `map[int]bool`, keyed by TMDBID (movies) / TVDBID (series).** `reconcileDownloadPaths` is extended to also return `movieMonitored map[int]bool` / `seriesMonitored map[int]bool` built from the same `[]radarr.Movie`/`[]sonarr.Series` slices it already holds (`true` when `Monitored`, entry present only when the item is in the fetched library). `cmd/download.go`'s `Run` passes these into `scheduler.BuildDeps` as new fields (e.g. `MonitoredMovieTMDBIDs`, `MonitoredSeriesTVDBIDs`), consumed only by `mergeIncompleteDownloads`. A nil map (the zero value, what every existing test and any caller that doesn't set the field gets) behaves identically to an empty one: every lookup misses.
Alternative considered: a per-item live lookup (`GetMovieByTMDBID`/`GetSeriesByTVDBID`) inside `mergeIncompleteDownloads` at merge time — rejected because the proposal commits to zero new API calls, incomplete-download counts aside; it would also double the number of network attempts against an already-unreachable Radarr/Sonarr instance within the same run, muddying the "unreachable → still resume" guarantee.
Alternative considered: pass the raw `[]radarr.Movie`/`[]sonarr.Series` slices into `BuildDeps` and let `mergeIncompleteDownloads` build its own lookup — rejected as duplicated map-building already done once for path reconciliation, and it would pull `radarr`/`sonarr` client package types one layer deeper into the scheduler than needed for a single boolean.

**A map miss is always "unconfirmed → resume", never "confirmed unmonitored → skip"; only an explicit `false` blocks resumption.** This covers both: the movie/series isn't in the fetched library at all (e.g. a TMDB-only match with no real Radarr/Sonarr entry — the same "population 2" case `reconcile-radarr-sonarr-download-paths` already carves out), and the fetch failed entirely this run (empty map, same as any other transient Radarr/Sonarr outage the run already tolerates elsewhere). Both produce a miss with no special-casing required.

## Risks / Trade-offs

- **[Risk] Radarr/Sonarr fetch fails this run** → the monitored maps stay empty, so `mergeIncompleteDownloads` resumes every incomplete download exactly as it did before this change — accepted, and safer than the alternative (a transient outage silently dropping a download the user still wants).
- **[Trade-off] A movie/series unmonitored *after* this run's library fetch already happened** (a race within the same run) won't be caught until the next scheduled run — accepted, consistent with `reconcile-radarr-sonarr-download-paths`'s identical up-to-one-run-interval delay for its own corrections.

## Migration Plan

No database schema change and no new configuration. Ships as: a new nil-safe `BuildDeps` field (existing tests/callers unaffected), `reconcileDownloadPaths` returning the two additional maps it can already build for free, and a check inside `mergeIncompleteDownloads`. Active from the next deploy's first scheduled `download` run onward; no backfill needed.
