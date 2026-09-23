# Proposal

## Why

Movies and TV episodes that already have a successful download are being re-downloaded through two paths that never check whether the work is already done: resuming an "incomplete" `DownloadInfo` for a different playlist occurrence of the same title, and the tier-2 quality-upgrade path. Both paths also name the destination file from the local database's `TMDBTitle` (configured in French on this deployment) instead of the live Radarr/Sonarr title used everywhere else, so the extra copies land under a differently-spelled filename in the same Radarr/Sonarr-managed folder — e.g. `Star Wars (1977).mkv` next to `La Guerre des étoiles (1977).mkv` and `.mp4`, or an entire TV season duplicated with filenames differing only by French vs. English colon-spacing (`Star Wars _ The Clone Wars` vs. `Star Wars_ The Clone Wars`).

Live inspection of the production database found 25 movies and 43 TV-episode rows currently affected, totalling ~465 GB of duplicate completed downloads. In every sampled case, whatever eventually reduced the folder back to one file (Radarr's own import, or manual cleanup) did so without stalkeer's knowledge: the scheduled and on-demand path reconciliation only checks whether a completed download's file still exists when its recorded root directory has drifted from Radarr's/Sonarr's current path. When the root already matches, no existence check happens at all, so a `download_info` row stays `completed` forever pointing at a file that is actually gone — invisible to every check in the system, including the `rename_failed` warning that at least flags the cases where the root *did* drift.

This is causing real, ongoing storage and bandwidth waste and produces confusing, duplicate-looking libraries, and it will keep recurring on any movie/series whose IPTV catalog carries more than one playlist entry for the same title.

## What Changes

- **Bug fix, no spec-level behavior change**: `mergeIncompleteDownloads` currently synthesizes or resumes a download stream for an incomplete `DownloadInfo` without checking whether the movie/episode already has a *different* `downloaded` `ProcessedLine`. This already violates the existing `media-download-scheduling` requirement that a work unit "qualifies for tier-1 classification only while the local database has no `downloaded` `ProcessedLine` for it." The fix makes it respect that existing rule; no requirement text changes. (The tier-2 upgrade builders already call this same check and a strict quality-rank comparison before building a stream — they are not part of this bug.)
- Automatic downloads (tier-1, tier-2, and the incomplete-download resume path) will consistently source the *filename's* title from the same live Radarr/Sonarr lookup already used for the destination *directory*, instead of falling back to the local database's `TMDBTitle`/`TVShow.TMDBTitle` when synthesizing a resume stream without a pre-built tier-1 stream to attach to.
- Scheduled and on-demand path reconciliation will verify a completed download's file still exists on disk even when its recorded root directory already matches Radarr's/Sonarr's current path, and will report a distinct outcome when the file is missing instead of silently doing nothing.

**Out of scope** (related, but not addressed by this change — noted for a follow-up):
- A tier-2 quality upgrade does not delete the file it supersedes; the superseded file is left on disk untracked. Not fixed here.
- Two completed downloads that happen to compute the exact same destination filename overwrite each other silently (`os.Rename` semantics) instead of being detected as a collision. Fixing the duplicate-download root cause above removes the main way this occurs today, but the underlying silent-overwrite behavior itself is not addressed here.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `sonarr-series-path-routing`: adds a requirement that the filename's title portion (not just the destination directory) is sourced from the live Radarr/Sonarr lookup for every automatic trigger path, including the incomplete-download resume path when it has to synthesize its own stream; adds a requirement that both scheduled and on-demand reconciliation verify file existence even when the root directory needs no correction, with a distinct reported outcome when the file is missing.
- `api-media-management`: the `POST /api/v1/downloads/:id/resync-path` endpoint's outcome contract gains the new "file missing" outcome so its response stays consistent with the scheduled reconciliation pass.

## Impact

- `internal/scheduler/build.go`: `mergeIncompleteDownloads` — add the missing already-downloaded check, and source its resume-synthesis destination (directory and filename) from a live Radarr/Sonarr lookup; `buildTier2MovieStreams`/`buildTier2SeriesStreams` — change only which title they pass for the filename (their already-downloaded/rank checks are untouched).
- `internal/downloader/destpath.go`: destination-path builders already accept a title parameter; call sites change which title they pass.
- `internal/api/resync_path.go`: `reconcileDownloadPath`/`applyPathCorrection` gain a file-existence check on the already-up-to-date path; `ReconcileOutcome` gains a new value.
- `cmd/download.go`: wiring for the live-lookup title source in the resume-synthesis path.
- No database schema change expected (existing `download_info.status`/`download_path` columns are sufficient); no breaking API change (new outcome value is additive).
