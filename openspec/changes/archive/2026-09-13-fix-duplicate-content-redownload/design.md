## Context

See `proposal.md` - Why. Relevant current-state facts that shape this design:

- `buildTier1MovieStreams` (`internal/scheduler/build.go`) adds a movie to this run's stream set purely because Radarr's `GetMissingMovies` still lists it; it never checks whether the local DB already has a `downloaded` `ProcessedLine` for that movie. `buildTier1SeriesStreams` does the same per-episode from Sonarr's `GetMissingEpisodes`.
- `findDownloadedLine(db, column, id)` already exists and is used by `buildTier2MovieStreams`/`buildTier2SeriesStreams` to find the best-ranked already-downloaded `ProcessedLine` for a movie (`column="movie_id"`) or episode (`column="tv_show_id"`) — the same lookup the tier-1 guard needs, just to test presence rather than rank it.
- `Stream.SeriesID` already carries the originating Sonarr series ID for tier-1 series streams (0 for movie/tier-2 streams). Nothing analogous exists for a movie's Radarr ID today; `Stream.SourceKey` for a movie is `"movie:%d"` using the *local* DB `Movie.ID`, not Radarr's ID.
- `cmd/download.go` already collects a run's successfully-changed destination folders (`downloadStats.changedPaths`) and, after the worker pool finishes, sends one deduplicated Jellyfin rescan per run (`notifyJellyfin`). That plumbing is a template but not a direct fit: Jellyfin notification is deliberately batched at run-end because it's not racing anything; here the whole point is to shrink the window before the *next* run, so end-of-run batching would give back most of the window on a long run.
- `radarrFullClient`/`sonarrFullClient` in `cmd/download.go` are already built with a per-run circuit breaker (`newRunBreaker()`) and the configured retry policy, per `radarr-sonarr-resilience`. New client calls added here reuse those, no new resilience mechanism needed.
- Radarr and Sonarr both expose a `/api/v3/command` endpoint for asynchronous commands. Verified against each project's source (`RescanMovieCommand`/`RescanSeriesCommand`, both under `MediaFiles/Commands`) during implementation: Radarr's `RescanMovie` takes a singular `movieId` (`{"name":"RescanMovie","movieId":id}`), not a `movieIds` array as originally assumed here; Sonarr's `RescanSeries` takes `{"name":"RescanSeries","seriesId":id}` as originally assumed. Command-name matching is case-insensitive and the endpoint responds `201 Created`.

## Goals / Non-Goals

**Goals:**
- Make the local "already downloaded" guard (spec: `media-download-scheduling`) the thing that actually stops the duplicate download, independent of Radarr/Sonarr ever catching up.
- Make the Radarr/Sonarr rescan notification (spec: `radarr-sonarr-post-download-rescan`) close the status-lag window as fast as practical within a run, so Radarr's/Sonarr's own UI and wanted lists reflect reality quickly too — a UX/consistency improvement layered on top of the guard above, not a substitute for it.
- Cover both the Radarr/movie path and the Sonarr/series path symmetrically, per the user's explicit request.

**Non-Goals:**
- Not changing how a single file is transferred (retry, resume, disk-space checks) - unaffected.
- Not deduplicating or cleaning up the duplicate files already sitting on disk for Vaiana/Pat'Patrouille - a manual follow-up, out of scope for this change.
- Not adding a full Radarr/Sonarr "download client" integration (import, history, quality profile push) - just enough of the commands API to trigger a rescan.
- Not changing `m3u-download` or `process` cronjobs, or the resume-downloads command.

## Decisions

### 1. Local guard: reuse `findDownloadedLine`, check it before building candidates
In `buildTier1MovieStreams`, right after matching `dbMovie` and before calling `matcher.FindMovieDownloadCandidates`, call `findDownloadedLine(deps.DB, "movie_id", dbMovie.ID)`; a non-nil result means the movie already completed a download locally, so `continue` (skip it entirely for tier-1) instead of building a stream. In `buildTier1SeriesStreams`, do the same per-episode with `findDownloadedLine(deps.DB, "tv_show_id", dbShow.ID)`, placed before the existing `FindTVShowDownloadCandidates`/empty-candidates check - so an already-downloaded episode is skipped without ever creating (or leaving empty) a `seasonMap` entry for a season whose *other* episodes are still genuinely missing.

Alternative considered: pre-load the full set of downloaded movie/episode IDs once (a single query) instead of one `findDownloadedLine` call per candidate movie/episode. Rejected for this change: `buildTier2*Streams` already does one such call per movie/episode in the same run, so this doesn't introduce a new cost shape, just reuses the existing one - batching both call sites is a separate, optional optimization that doesn't change observable behavior and isn't needed to fix the bug.

### 2. Stream gains a resolved Radarr movie ID; Sonarr already has one
Add `Stream.RadarrMovieID int` (0 = unknown/not applicable), populated from the `radarr.Movie` already fetched in `buildTier1MovieStreams` (tier-1) and in `buildTier2MovieStreams` (tier-2, which already calls `deps.Radarr.GetMovieByTMDBID`). `Stream.SeriesID` already serves this role for tier-1 series streams; populate it the same way in `buildTier2SeriesStreams` (which already calls `deps.Sonarr.GetSeriesByTVDBID`) so tier-2 series completions can also be rescanned. Streams synthesized by `mergeIncompleteDownloads` (resumed crashed downloads, built without a fresh live lookup) leave these fields at 0 - see Risks.

### 3. Notify per completed item, deduplicated within the run - not batched at run-end
`downloadStats` gains two guarded sets, `notifiedMovieIDs`/`notifiedSeriesIDs` (`map[int]struct{}`), checked and updated under the existing `stats.mu` at the same point `recordItem` is called. When an item succeeds and its owning `Stream.RadarrMovieID`/`SeriesID` is non-zero and not already in the corresponding set, the appropriate rescan call is issued immediately (from the worker goroutine, synchronously, bounded by a short timeout) and the ID is added to the set so a second episode of the same series completing later in the run does not re-trigger it.

Alternative considered: mirror the Jellyfin pattern exactly (collect during the run, fire deduplicated notifications once after `runDownloadWorkerPool` returns). Rejected: the entire motivation for this capability is to shrink the time between "file written" and "Radarr/Sonarr knows" as much as possible; a `parallel`-worker run can take a long time end-to-end, so waiting for every stream to drain before notifying about the first movie that finished minutes ago would reintroduce most of the gap this change exists to close. Per-item notification with in-run dedup satisfies the spec's "exactly once per run" requirement while notifying as early as possible.

### 4. New Radarr/Sonarr client methods, routed through the existing per-run breaker
Add `radarr.Client.RescanMovie(ctx, movieID int) error` and `sonarr.Client.RescanSeries(ctx, seriesID int) error`, each a thin POST to `/api/v3/command` using the client's existing `newRequest`/retry/breaker plumbing (no new resilience mechanism - `radarr-sonarr-resilience` already covers "every call to that service's API"). Both are best-effort: on error, `cmd/download.go` logs and moves on, never touching `downloadStats.Failed` or the run's exit status.

## Risks / Trade-offs

- **[Risk]** Resumed streams synthesized by `mergeIncompleteDownloads` (crash-recovered downloads) don't have a fresh Radarr/Sonarr lookup, so `RadarrMovieID`/`SeriesID` stay 0 and no rescan fires for them. → Mitigation: none needed beyond today's behavior - Radarr's/Sonarr's own periodic scan still eventually catches these, exactly as before this change; only the common path improves.
- **[Risk]** `RescanMovie`/`RescanSeries` command names/payloads were inferred from Radarr/Sonarr's v3 API and not verified against this deployment's exact version at design time. → Mitigation: verified against Radarr's/Sonarr's own source during implementation (see Context) rather than this deployment's live instances (not reachable from the implementation environment); this deployment runs `:latest` images with no version pin, and the command's C#/JSON contract has been stable for years, so this is treated as sufficiently confirmed.
- **[Trade-off]** The rescan notification is genuinely best-effort/non-blocking; if Radarr/Sonarr queues the command slowly, a very tight next tick could in theory still see "missing". → Not a correctness gap: Decision 1's local guard is what actually prevents the duplicate download, independent of whether/when Radarr/Sonarr's status updates. The rescan is a latency/UX improvement layered on top.
- **[Risk]** Extra Radarr/Sonarr API calls (up to one per distinct completed movie/series per run) add minor load. → Mitigation: bounded by the existing per-run circuit breaker; volume is proportional to a run's own completed-download count, which is already small (a handful of movies/episodes per 2h run in this deployment).

## Migration Plan

1. Implement the `internal/scheduler/build.go` guard (Decision 1) and the `Stream.RadarrMovieID`/`SeriesID` population (Decision 2) - these alone stop the duplicate downloads and can ship independently of the rescan notification if desired.
2. Add `RescanMovie`/`RescanSeries` to the Radarr/Sonarr clients, the `downloadStats` tracking, and the `cmd/download.go` call sites (Decisions 3-4).
3. Deploy via the normal image/Helm release; no schema or config changes, no new CRDs or PVCs.
4. Rollback: revert the image; no data migration to undo since nothing persisted changes shape (only `Stream`, an in-memory type, gained fields).

## Open Questions

- Resolved during implementation: see Context - Radarr's `RescanMovie` command takes a singular `movieId` (not a `movieIds` array), Sonarr's `RescanSeries` takes `seriesId`, confirmed against each project's source rather than this deployment's live instances (unreachable from the implementation environment, and unpinned `:latest` images anyway).
