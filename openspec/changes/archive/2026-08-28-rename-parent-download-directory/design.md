## Context

Downloaded files live on disk following two conventions built by `internal/downloader/path.go`:
- Movie: `{root}/{Title (Year)}/{Title (Year)}.ext`
- TV show: `{root}/{Series (Year)}/Season NN/{Series (Year)} - SXXEYY.ext`

`internal/models.DownloadInfo.DownloadPath` is the only place the physical path is persisted; `Movie`/`TVShow` rows hold TMDB metadata only, and `TVShow` is one row per episode (there is no series- or season-level entity). The existing `moveMovieFolder`/`moveTVShowFolder` handlers (`internal/api/handlers_frontend.go:271-495`) already contain the two pieces of logic this change needs to reuse: detecting the series root by checking whether a path's parent directory name starts with `"season"` (`:411-415`), and `MoveDir` (`:498+`, rename-first with copy+verify+delete fallback). Both existing handlers are scoped by `Movie.ID`/`TVShow.ID` and move the *entire* folder; this change adds a handler scoped by a single `DownloadInfo.ID` that moves *one file*. See proposal.md - Why/What Changes for the motivation and behavioral contract.

## Goals / Non-Goals

**Goals:**
- Reuse the existing path-detection and `MoveDir`-style move primitives rather than inventing a parallel path-building system.
- Keep the new endpoint fully independent from `Movie`/`TVShow` rows so it cannot hit the id-scope ambiguity that `moveMovieFolder`/`moveTVShowFolder` have (those expect a `Movie.ID`/`TVShow.ID`, but `TVShow` is per-episode).

**Non-Goals:**
- Changing TMDB matching/metadata (`Movie`/`TVShow` rows are never modified by this feature).
- Reworking `moveMovieFolder`/`moveTVShowFolder` or their existing id-scoping behavior — out of scope for this change.
- Bulk/multi-item rename in one request.

## Decisions

**Endpoint scope: `DownloadInfo.ID`, not `Movie`/`TVShow.ID`.**
The Downloads page already has the `DownloadInfo` id for every card (`DownloadEnriched.id`, see `downloads-enrichment-api`), so no new lookup is needed client-side, and the handler can go straight to `db.First(&dl, id)` instead of walking `ProcessedLines`. This also sidesteps the existing move handlers' id-scope mismatch instead of inheriting it.

**Single-file move instead of `MoveDir` on a directory.**
`MoveDir` (`handlers_frontend.go:498+`) moves a whole directory tree. This feature only ever touches one file, so the handler will: `os.MkdirAll` the destination file's directory, then `os.Rename` the file with a copy+verify+delete fallback mirroring `moveFile` in `internal/downloader/downloader.go:610-613` (same fallback shape as `MoveDir`, just for a single file instead of a tree). A small shared helper (e.g. `moveFile(src, dst)`) is reused/extracted where possible rather than duplicating the fallback logic a third time.

**Path reconstruction reuses the existing convention, driven by `new_name` alone.**
`new_name` replaces the whole `{Title (Year)}` segment used by `buildMovieBasePath`/`buildTVShowBasePath` (`internal/downloader/path.go:8-18`) — it is sanitized with the existing `sanitizeFilename` and used both as the new folder name and as the file's basename prefix, exactly like today's generated names. Season number for a TV episode is read off the *current* path (same `"season"`-prefix detection already used in `moveTVShowFolder:411-415`), not from a DB join to `TVShow.Season` — this keeps the handler independent of `Movie`/`TVShow` rows entirely and matches the pattern already proven in the move handlers.

**Cleanup of emptied directories.**
After a successful move, the handler checks whether the original season directory (TV only) and then the original series/movie directory are empty (`os.ReadDir` returns zero entries) and removes each with `os.Remove` (not `RemoveAll`, to avoid ever deleting a non-empty directory through a race). This is a best-effort step after the DB update commits; a failure to clean up is logged but does not fail the request, since the file has already been safely relocated and the DB is consistent.

**Collision handling blocks before any mutation.**
The handler `os.Stat`s the computed destination path first; if it exists, it returns `rename_target_exists` (400) immediately, before any `MkdirAll`, move, or DB write — mirroring the `same_directory` pre-check already done in `moveMovieFolder:326-332`.

**Order of operations: move on disk, then DB update, then cleanup — matching the existing move handlers' order** (`handlers_frontend.go:344-370`). If the DB update fails after a successful disk move, the response uses a distinct `database_update_failed`-style code (same convention as the existing handlers) so the client knows the file moved but the record is stale.

## Risks / Trade-offs

- **[Risk]** A race between the collision `os.Stat` check and the actual move (TOCTOU) could still cause `os.Rename` to overwrite an existing file on some filesystems/OSes. → Mitigation: same residual risk already accepted by the existing `same_directory` check in the move handlers; not worsened by this change, and full atomic-create semantics are out of scope.
- **[Risk]** Deriving the season number from the current path string (rather than a DB field) means a download whose path doesn't already follow the `Season NN` convention (e.g. manually placed files) won't be recognized as a TV episode correctly. → Mitigation: this is the exact same detection already relied on by `moveTVShowFolder`, so behavior stays consistent with the rest of the app; no regression introduced.
- **[Trade-off]** Best-effort cleanup (log-only on failure) means an emptied directory can occasionally be left behind (e.g. permissions issue). Accepted since failing the whole rename over a non-critical cleanup step would be worse UX than a rare leftover empty folder.

## Migration Plan

Additive only: new route `POST /api/v1/downloads/:id/rename` registered alongside the existing `/movies/:id/move` and `/tvshows/:id/move` routes in `internal/api/api.go`; new handler in `internal/api/handlers_frontend.go`; no schema migration (writes to the existing `download_info.download_path` column). No rollback concerns beyond reverting the route/handler/frontend addition.
