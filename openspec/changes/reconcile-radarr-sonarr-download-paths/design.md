## Context

See `proposal.md` for motivation and `specs/sonarr-series-path-routing/spec.md`, `specs/api-media-management/spec.md`, `specs/downloads-details-sidepanel/spec.md` for the behavior contract. This section only covers the implementation-shape facts needed to justify the decisions below.

Tracing the existing code confirms `download_path`'s structure precisely matches what the rename feature already assumes:
- **Movie**: `download_path` = `{movie.Path}/{fileBase}.ext`, i.e. Radarr's `Path` for that movie *is* the file's immediate parent directory. `filepath.Dir(download_path)` already recovers it (`handlers_frontend.go`, used by `renameFolderName`/`computeRenameDestination`).
- **TV episode**: `download_path` = `{series.Path}/Season NN/{fileBase}.ext`. `detectTVSeasonPath(download_path).SeriesRoot` (`filepath.Dir(filepath.Dir(download_path))`) already recovers `series.Path`.

Both extraction functions already exist and are already unit-tested (they back the "Rename Target Folder Name" behavior in `downloads-enrichment-api`). Reconciliation therefore reduces to: recover the currently-stored root with these same functions, compare its basename (and parent) against Radarr's/Sonarr's current `movie.Path`/`series.Path`, and — when they differ — invoke the same move-and-rename-DB-record primitive the manual "Renommer" endpoint (`renameDownload` → `computeRenameDestination` → `moveSingleFile` → update `download_path`) already uses, except the new folder name/parent comes from Radarr/Sonarr instead of user free text.

## Goals / Non-Goals

**Goals:**
- Keep `download_path` for Radarr/Sonarr-matched completed downloads converging back to reality without requiring the user to redo a manual rename through the app.
- Reuse the existing rename primitive and its tested collision/cleanup handling rather than building a second file-move code path.
- Add exactly one new full-library fetch per Radarr/Sonarr to the scheduled `download` run (not a per-item call), and exactly one extra call (a single-item lookup) per on-demand resync request.

**Non-Goals:**
- Reconciling downloads with no confirmed Radarr/Sonarr match (population 2, per proposal.md) — explicitly out of scope.
- Detecting or correcting drift in anything other than the root directory (e.g. a user manually renaming just the file's own basename without touching the root is not this change's concern).
- Real-time/live reconciliation on every `GET /api/v1/downloads` read — the enrichment endpoint's `< 500ms` budget stays untouched; correction happens at the two write-triggering points described in the proposal.

## Decisions

**Reuse the rename primitive instead of writing a parallel move path.** Both the scheduled and on-demand mechanisms compute a target root name/parent from Radarr's/Sonarr's `Path` and hand it to the same `computeRenameDestination`/`moveSingleFile`/DB-update sequence already exercised by `POST /api/v1/downloads/:id/rename`. Alternative considered: a bespoke "sync path" function duplicating the move+collision+cleanup logic — rejected, since it would duplicate already-tested behavior (empty-directory cleanup, collision detection) for no behavioral difference.

**Scheduled reconciliation is a lookup against an in-memory map, not a per-item API call.** The `download` run's existing `GetMissingMovies`/`GetMissingEpisodes` calls only return content not yet downloaded, so they cannot answer "where does an already-completed item live now" — reconciliation therefore adds one new `GetAllMovies`/`GetAllMonitoredSeries` full-library fetch per run, once each, not gated by item count. Building a `map[tmdbID]Path` (movies) and `map[tvdbID]Path` (series) from that response, then iterating completed downloads once, keeps the added cost at exactly one extra call per Radarr/Sonarr per run regardless of library or download-history size. Alternative considered: a per-download `GetMovieDetails`/`GetSeriesDetails` call during the run — rejected as needlessly expensive and rate-limit-prone at library scale, and unnecessary since a single full-library fetch already contains everything needed for the lookup.

**On-demand resync uses a single live lookup, not the full library fetch.** `GetMovieByTMDBID`/`GetSeriesByTVDBID` (already present in the Radarr/Sonarr clients) fetch just the one entry the user asked to fix, keeping the manual action fast and independent of library size.

**Reconciliation eligibility is decided by presence in Radarr's/Sonarr's response, not by a stored flag.** No new "is this Radarr/Sonarr-managed" column is introduced. Both mechanisms treat "lookup returned nothing" as the population-2 signal, computed fresh each time. Alternative considered: persisting a `radarr_managed`/`sonarr_managed` boolean on `Movie`/`TVShow` at match time — rejected as an extra piece of state that could itself drift (e.g. a movie later removed from Radarr) when the live lookup already gives the correct answer for free at the only two moments it's needed.

**Movies and TV shows share one algorithm shape, parameterized by root-extraction function.** Movie uses `filepath.Dir`; TV uses `detectTVSeasonPath(...).SeriesRoot`. Both then follow the identical "compare basename+parent of extracted root to Radarr/Sonarr Path, move if different" logic, avoiding two independently-maintained implementations.

## Risks / Trade-offs

- **[Risk] Radarr/Sonarr API unreachable during a scheduled `download` run** → the run already tolerates fetch failures for its primary scheduling purpose; reconciliation SHALL be skipped for that run (not retried mid-run) and simply re-attempted on the next scheduled run, same as any other transient Radarr/Sonarr outage the run already handles.
- **[Risk] A move succeeds on disk but the subsequent `download_path` DB update fails** → already a handled case in the existing `renameDownload` handler (it returns a distinct error indicating the file moved but the DB didn't update); the reconciliation path SHALL reuse that same error handling rather than inventing new semantics.
- **[Risk] Concurrent scheduled reconciliation and an on-demand resync race on the same download** → both ultimately call the same move-and-update sequence; the on-demand endpoint SHALL re-read the `download_info` row immediately before moving (as the existing rename endpoint already does) so a same-run collision surfaces as the existing `"rename_target_exists"`/`"rename_failed"` outcomes rather than corrupting state.
- **[Trade-off] Scheduled reconciliation has a delay** (up to one scheduled run's interval) before a manually-fixed item stops showing a stale path/error — accepted, since the on-demand action exists precisely to cover the "I want this fixed now" case.

## Migration Plan

No database schema change and no new configuration. The change ships as: (1) a new step in the existing scheduled `download` run, active from the next deploy's first scheduled run onward; (2) a new endpoint and sidepanel action, usable immediately after deploy. No backfill or data migration is required — the first scheduled run after deploy naturally reconciles any drift that already exists.
