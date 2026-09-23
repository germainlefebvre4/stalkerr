# Design

## Context

See `proposal.md` for motivation and the production evidence (25 movies, 43 TV-episode rows, ~465 GB of duplicate completed downloads).

Two independent gaps combine to produce the observed duplicates, and a third gap hides the aftermath:

1. `mergeIncompleteDownloads` (`internal/scheduler/build.go`) resumes any `DownloadInfo` whose status is one of pending/downloading/paused/failed/retrying/policy-stopped, without ever checking whether the movie/episode it belongs to already has a *different* `downloaded` `ProcessedLine`. When it has to synthesize a brand-new stream (because tier-1 already skipped building one — precisely *because* the movie is already downloaded), it builds the destination from the local `Movie.TMDBTitle`/`TVShow.TMDBTitle` and an empty root path, instead of the live Radarr/Sonarr lookup tier-1 and tier-2 both use.
2. `buildTier2MovieStreams`/`buildTier2SeriesStreams` already gate correctly on "already downloaded" plus a strict quality-rank comparison — they are not part of the duplicate-download bug — but they still source the destination *filename's* title from the local `TMDBTitle` while correctly sourcing the destination *directory* from a live lookup, so an accepted upgrade can rename the file to a different language/title than the original.
3. Path reconciliation (`internal/api/resync_path.go`) only ever touches a completed download when its root directory has drifted from Radarr's/Sonarr's current path. When the root already matches, no check of any kind happens — including whether the file still exists. Every duplicate we inspected in production had already been reduced to one physical file by something outside stalkeer (Radarr's own import, or manual cleanup), leaving 1-4 `completed` `download_info` rows per movie permanently pointing at files that no longer exist, with zero signal anywhere in the system.

## Goals / Non-Goals

**Goals:**
- Stop `mergeIncompleteDownloads` from resuming/synthesizing a stream for a movie/episode that already has a different downloaded occurrence.
- Make every automatic-download destination filename (tier-1, tier-2, resume-synthesis) consistently use the live Radarr/Sonarr title, matching what tier-1 already does.
- Surface (not silently absorb) a completed download whose file is gone even when its root directory hasn't moved, via both the scheduled pass and the on-demand endpoint.

**Non-Goals:**
- Deleting the file a tier-2 upgrade supersedes (proposal.md, Out of scope).
- Detecting/preventing the silent `os.Rename` overwrite when two downloads compute an identical destination filename (proposal.md, Out of scope) — largely mitigated as a side effect of the goals above, but not directly addressed.
- Backfilling or cleaning up the ~25 movies / 43 episode rows already affected in production. This change fixes the mechanism going forward; remediating already-orphaned `download_info` rows is an operational follow-up, not a spec change (see Open Questions).

## Decisions

### D1: `mergeIncompleteDownloads` gains the same already-downloaded check, and cancels what it skips

Before resuming or synthesizing a stream for an incomplete `DownloadInfo`, call the same `findDownloadedLine` helper `buildTier1MovieStreams`/`buildTier1SeriesStreams` already use, keyed by the specific movie/episode identity the *sibling* occurrence belongs to (movie_id for a movie; the per-episode `tv_show_id` row for TV, since each `TVShow` row is already one specific episode). If a `downloaded` line already exists for that identity, this incomplete `DownloadInfo` is not resumed this run.

Rather than leaving the skipped `DownloadInfo` sitting in an "incomplete" status forever — re-fetched and re-skipped by `GetIncompleteDownloads` on every future run — transition it to `DownloadStatusCancelled` (the same terminal status the existing `cancel-download-occurrence` capability already uses), so it naturally drops out of the incomplete-downloads query going forward.

**Alternative considered**: skip without changing status. Rejected — this repeats the exact same no-op check forever for the same stale row and leaves a misleading "still eligible" row in download history/monitoring views indefinitely.

### D2: Resume-synthesis reuses the run's existing single Radarr/Sonarr library fetch, no new API calls

`cmd/download.go`'s `reconcileDownloadPaths` already performs one `GetAllMovies`/`GetAllSeries` call per run and builds `movieTMDBPaths`/`seriesTVDBPaths` (path only) plus `movieMonitored`/`seriesMonitored` from it, passing the monitored maps into `BuildDeps` — but not the paths, and not titles.

Decision: extend that same single fetch to also capture each movie's/series' `Title` alongside its `Path` (both already present on the `radarr.Movie`/`sonarr.Series` structs returned by that one call), and pass this combined lookup (keyed the same way as the existing monitored maps: `Movie.TMDBID` / `TVShow.TVDBID`) into `BuildDeps`. `mergeIncompleteDownloads`' synthesis branch uses it for both the destination directory and the filename title, exactly mirroring what a tier-1 stream for that same movie/series would have used — with zero additional Radarr/Sonarr calls, since the data is already being fetched today.

**Alternative considered**: give `mergeIncompleteDownloads` its own live per-item lookup (`GetMovieByTMDBID`/`GetSeriesByTVDBID`, as tier-2 already does). Rejected in favor of reusing the existing run-wide fetch — synthesis can in principle fire for several incomplete downloads in one run, and reusing already-fetched data avoids a proportional number of extra API calls for no benefit.

**Degrade path**: if the library fetch fails this run (Radarr/Sonarr unreachable), `reconcileDownloadPaths` already tolerates this today (logs a warning, proceeds with an empty map). Synthesis simply falls back to the existing local-title/empty-path behavior in that case — unchanged from today's behavior, not a new failure mode.

### D3: A new `file_missing` reconciliation outcome, checked only on the "root already matches" path

In `reconcileDownloadPath` (`internal/api/resync_path.go`), when `computeReconciledPath` reports `changed=false` (root already matches), add an `os.Stat` on the stored `download_path` before returning:
- File present → `OutcomeAlreadyUpToDate`, exactly as today.
- File absent → new `OutcomeFileMissing ReconcileOutcome = "file_missing"`, database left untouched.

This runs in both the scheduled pass (`ReconcileScheduledDownloadPaths`, logged as a WARN distinguishable from today's `rename_failed`) and the on-demand endpoint (`resyncDownloadPath`, returned as `200 OK` with `{"status": "file_missing"}` — a report, not a failure, matching how `already_up_to_date` is already a 200 today).

**Alternative considered**: fold this into the existing `rename_failed` outcome instead of adding a new value. Rejected — collapsing "I tried to move the file and failed" and "the root is fine but the file is simply gone" into one bucket is the exact ambiguity that let 6 movies' worth of `rename_failed` warnings go uninvestigated for weeks in production; the two situations need distinguishable signals to actually get acted on.

### D4: Tier-2 filename title comes from the lookup it already performs

`buildTier2MovieStreams`/`buildTier2SeriesStreams` already fetch `radarrMovie`/`series` live (for the destination directory, per the existing `sonarr-series-path-routing` requirements). Change only which field of that already-fetched struct feeds the filename: `radarrMovie.Title`/`series.Title` instead of `movie.TMDBTitle`/`ep.TMDBTitle`. No new call, no change to the existing "already downloaded + strictly better candidate" gating.

## Risks / Trade-offs

- **[Risk]** Cancelling a stale incomplete `DownloadInfo` (D1) removes a retry path for the case where the "sibling" already-downloaded occurrence's file later disappears. → **Mitigation**: this is exactly what D3 covers — if the sibling's file goes missing, scheduled reconciliation now surfaces `file_missing` so a human (or a future, separate change) can act, instead of `mergeIncompleteDownloads` silently re-downloading a second copy the way it does today.
- **[Risk]** Changing tier-2's filename title source (D4) means an already-downloaded item with a pre-existing local-language filename that later receives a tier-2 upgrade will produce a file named in whatever language Radarr/Sonarr's live title is — a third title variant appearing on that one item's next upgrade. → **Mitigation**: accepted as a one-time transitional artifact; nothing breaks (the movie's root folder, tracked separately by path reconciliation, is unaffected), and it naturally stops recurring since all paths now agree on the same title source going forward.
- **[Risk]** D2's combined path+title map adds a small amount of data to `BuildDeps` that only the rare incomplete-download-synthesis path uses on most runs. → **Mitigation**: negligible — the data is already fetched and held in memory for the monitored-status maps; this only threads two more fields per entry through the same structure.

## Migration Plan

- No database schema change.
- No feature flag: both fixes are corrections to existing intended behavior (per the already-in-spec tier-1-precedence rule) and an additive, non-breaking API outcome — safe to deploy directly.
- No backfill of already-affected production rows is included (see Non-Goals). Operators can use the existing `resync-path` endpoint once `file_missing` starts surfacing to find and manually resolve stale records for the movies/episodes already affected.

## Open Questions

- Should a follow-up change add an automated or UI-driven way to review/clean up `download_info` rows the scheduled pass now reports as `file_missing`, rather than requiring an operator to notice the warning and act manually? Deferrable — doesn't change this change's specs, approach, or tasks.
