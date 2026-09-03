## Context

See `proposal.md` - Why. Relevant current state:

- `DownloadInfo.DownloadPath` (`internal/models/download.go`) is written exactly once, in `updateDownloadInfoCompleted` (`internal/downloader/downloader.go:452`), after both the HTTP transfer and the move-to-final-destination succeed. It stays `nil` for `pending`, `downloading`, `retrying`, and `failed` states.
- Three backend code paths treat `DownloadPath != nil` as a proxy for "the file physically exists on disk, download is done": `POST /api/v1/downloads/:id/rename` (guards only on non-nil `DownloadPath`, no `status` check), `moveMovieFolder`/`moveTVShowFolder` (collect every sibling `DownloadInfo` with a non-nil `DownloadPath` into the batch being moved on disk), and `resume_helper.buildBaseDestPath` (prefers an existing `DownloadPath` to keep the target filename stable across retries).
- `Download()` (`internal/downloader/downloader.go:87`) receives `opts.BaseDestPath` (final folder + filename, without extension) as an input - it is computed by the caller (`scheduler/build.go`, `force_download.go`) before `Download()` is ever invoked. `detectFileExtension(url, contentType)` (line 575) tries the URL path first and only falls back to a Content-Type map or a `.mkv` default - so a best-guess extension is available immediately, without waiting for an HTTP response.
- The actual bytes land in a per-`Download()`-call temp file (`tempDownloadDir/download.tmp`, `tempDownloadDir` named with a fresh UUID), computed once before entering the retry loop and reused across retries within that call. `defer os.RemoveAll(tempDownloadDir)` deletes it synchronously when `Download()` returns, on both the success and the failure paths.

## Goals / Non-Goals

**Goals:**
- Make the intended final location and the current temporary-write location visible through the API and the sidepanel while a download is `downloading`, `retrying`, or `failed`.
- Do this without changing the meaning of `download_path` or touching any of the three code paths that currently key off its presence.

**Non-Goals:**
- Do not make the temporary file survive past `Download()` returning on failure (no retention/cleanup policy change). `staging_path` may point to a file that no longer exists by the time it is read.
- Do not add any action (browse/open/resume-from) built on `staging_path` or `target_path` - they are display-only.
- Do not change `rename`, `moveMovieFolder`/`moveTVShowFolder`, or `resume_helper.buildBaseDestPath` behavior or guards.

## Decisions

**New, separate nullable columns instead of writing `download_path` earlier.**
Reusing `download_path` (writing it as soon as the destination is known) was the initially discussed option but was dropped: it would make `DownloadPath != nil` true for in-progress and failed downloads too, which three existing code paths interpret as "file exists, safe to rename/move/reuse as resume target" without any additional `status` check. Introducing `target_path`/`staging_path` as new fields that nothing else reads keeps `download_path`'s existing invariant intact and needs no changes to rename/move/resume code.

**`target_path` is written once, at the start of the attempt, using the URL-guessed extension.**
`opts.BaseDestPath` and a best-effort extension (`detectFileExtension(opts.URL, "")`, URL-only since no response exists yet) are both available before the HTTP call starts. Writing `target_path` at that point (alongside the existing `status -> downloading` transition in `state_manager.go:155`) covers the common case where the URL path carries the real extension. Edge case: if the Content-Type-based extension used at completion differs from the URL-guessed one, `target_path` and the eventual `download_path` differ by extension only, for the (short) duration before completion - acceptable since `target_path` is explicitly labeled as a planned location in the UI, not a promise.

**`staging_path` is written once per `Download()` call, right after the temp file path is computed.**
The temp path is already stable across retries within one `Download()` invocation (computed once, before `retry.Do`), so one write covers the whole attempt, including any retries.

**Both fields are cleared to `nil` on successful completion, in the same update as `download_path` is set.**
Keeps the sidepanel's file section unambiguous once a download is done: only `download_path` remains meaningful, matching the existing `downloads-details-sidepanel` requirement's "completed" scenarios unchanged. Alternative considered: leave them populated as history - rejected, since it would force the frontend to reason about three path fields simultaneously for a completed item with no benefit (the value was diagnosing an in-progress or failed state).

**On failure, `target_path`/`staging_path` are left as-is (not cleared).**
This is the primary payoff of the change: after a failure, the sidepanel can still show where the file was headed and where the failed attempt was writing, even though the temp file itself is already gone (`defer os.RemoveAll`). The UI must label `staging_path` in a way that does not imply the file is still retrievable.

**Migration via `AutoMigrate`.**
The project already runs GORM `AutoMigrate` (`internal/database/database.go`); two new nullable columns need no hand-written migration.

## Risks / Trade-offs

- [`staging_path` looks like a real, browsable file path but the file is deleted by the time most readers see it] -> Mitigate in the UI copy only (explicit "temporary, may already be gone" label); no backend guarantee is added or implied.
- [`target_path`'s extension can be wrong if URL and Content-Type disagree] -> Label it explicitly as a planned/not-final location in the sidepanel; `download_path` remains authoritative once set.
- [Two more nullable columns on a frequently-updated table] -> Both are simple `*string` columns, written at most twice per download attempt (once at start, once cleared at completion); no measurable impact expected given the existing `bytes_downloaded`/`total_bytes` progress columns are already updated far more frequently.
