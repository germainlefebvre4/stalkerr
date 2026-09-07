## Why

Scheduled path reconciliation (`ReconcileScheduledDownloadPaths`, `cmd/download.go`) compares a completed download's stored root directory against Radarr's/Sonarr's current `movie.Path`/`series.Path` using plain string equality, with no normalization. When that `Path` happens to carry a trailing directory separator (confirmed in production Sonarr data for 7 of 97 series — a pre-existing data artifact in Sonarr's own root-folder configuration, not something Stalkeer controls), the equality check always reports a mismatch even though the two paths are identical. The system then attempts to "correct" the path by moving the file onto itself, always finds the destination already occupied (it's the same file), and logs a `rename_target_exists` warning. Because this outcome is never persisted, it is compared and warned about again on every single scheduled run — currently producing over 200 identical warnings per run, one for every completed episode of the 7 affected series, with no way for the condition to ever resolve.

## What Changes

- Normalize both sides of the root-directory comparison (`computeReconciledPath` in `internal/api/resync_path.go`) so a trailing path separator on Radarr's/Sonarr's reported `Path` no longer causes a false "path has drifted" result.
- Apply the same normalization to both the scheduled reconciliation pass and the on-demand `resync-path` endpoint, since both call the same shared comparison function.
- No change to behavior when the roots are genuinely different (a real rename/move is still detected and corrected as before).

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `sonarr-series-path-routing`: the "Series root unchanged" and "Movie root unchanged" scenarios (scheduled reconciliation), and the "On-demand lookup finds the path already correct" scenario, currently rely on an implicit exact-string match between the stored root and Radarr's/Sonarr's current `Path`. This is being tightened to explicitly require normalized comparison, so a trailing separator on the Radarr/Sonarr-reported `Path` is treated as "unchanged" rather than "drifted."

## Impact

- **Code**: `internal/api/resync_path.go` (`computeReconciledPath`, used by both `ReconcileScheduledDownloadPaths` and `resyncDownloadPath`).
- **Tests**: `internal/api/resync_path_test.go` gains coverage for a target root with a trailing separator.
- **Operational**: eliminates the ~200 recurring `rename_target_exists` warnings per scheduled run for the 7 currently-affected series; no data migration needed, since affected `download_path` values are already correct on disk and in the DB — only the comparison logic was wrong.
- **Out of scope**: the separate, smaller cluster of `rename_failed` warnings on a handful of movies (caused by title drift between Stalkeer's stored path and Radarr's current title, compounded by duplicate downloads across different playlist quality/language variants) is a distinct problem, explicitly not addressed by this change.
