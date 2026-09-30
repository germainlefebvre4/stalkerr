# Design

## Context

See `proposal.md` for motivation and `specs/` for the behavior contract. Existing pieces this builds on:

- `internal/api/duplicate_download.go` already holds `resolveDuplicateDownload`, `findGroupContaining` and `cancelDuplicateOccurrence` (the transaction that sets `download_info.status` and `processed_lines.state` to `cancelled`). The new handlers live in the same file/package and reuse all three.
- `database.FindDuplicateDownloadGroups` is the single definition of a duplicate group. After the action, cancelled occurrences no longer match its `completed`/`downloaded` filters, so a group with only one distinct path left stops being reported without any change to detection.
- `resyncDownloadPath` (`internal/api/resync_path.go`) shows how an API handler builds a Radarr/Sonarr client from `settings.Effective()` with `existenceCheckTimeout`, `retry.Config{MaxAttempts: 1}` and the server's circuit breakers, and how it maps a `ProcessedLine` to a Radarr movie (`GetMovieByTMDBID`) or Sonarr series (`GetSeriesByTVDBID`). The new code follows that shape.
- `RescanMovie`/`RescanSeries` (used by `cmd/download.go` after downloads) post to `/api/v3/command`. Refresh & Scan is the same transport with a different command name.
- The API pod mounts `/media` read-write in `charts/stalkerr/templates/deployment.yaml`; `docker-compose.yml`'s API service shows no media volume.

## Goals / Non-Goals

**Goals:**
- Let an operator finish a duplicate cleanup in one confirmed action, with Radarr/Sonarr left consistent with disk.
- Make the destructive step as constrained as possible: only files stalkeer itself recorded for this group, only regular files, never the kept one.
- Keep the list endpoint cheap: no live Radarr/Sonarr call is added to `GET /api/v1/downloads`.

**Non-Goals:**
- Retrying a failed Radarr/Sonarr refresh automatically, or queueing it.
- Quality-based suggestion (would need Radarr/Sonarr movie-file/episode-file endpoints the clients do not have).
- Cleaning the parent folder beyond the existing `removeDirIfEmpty` helper; a leftover empty folder is left to Radarr/Sonarr.

## Decisions

### D1: Two new endpoints, keyed like the existing resolve endpoint

- `GET /api/v1/downloads/:id/duplicate-candidates` — read-only, describes the group.
- `POST /api/v1/downloads/:id/keep-duplicate` — `:id` is the `download_info` id of the occurrence to keep; no request body.

`:id` identifies the kept occurrence rather than passing a body with a second id: the group is derived from it exactly as `resolve-duplicate` does, there is a single source of truth for "which file", and there is no id/body mismatch to validate. The existing `resolve-duplicate` endpoint and its response are untouched.

**Alternative considered**: extend `resolve-duplicate` with an optional `keep` body. Rejected — it would make one endpoint both non-destructive and destructive depending on a field; keeping them separate lets the verify-only endpoint keep its "never touches a file" guarantee absolute and simple to audit.

### D2: Candidate suggestion is computed on demand, from local data plus one library lookup

The candidates handler groups the group's occurrences by `download_path`, then per path does `os.Stat` (exists, size) and takes the max `CompletedAt`. It then performs at most one Radarr/Sonarr lookup (`GetMovieByTMDBID` / `GetSeriesByTVDBID`, same as `resyncDownloadPath`) to get the live title.

Canonical-name match is a normalized comparison of the file's base name with the live title: for movies, the base name (minus extension, case-folded, punctuation-normalized) must start with `"<Title> (<Year>)"`; for episodes, it must contain the normalized series title. Quality/language tags added by `download-filename-quality-tags` are ignored by the prefix/contains rule. This is a heuristic used only to order a suggestion the operator confirms, so a false negative is harmless.

Suggestion order: existing file matching canonical name → existing file with the latest completion → none. A lookup error of any kind (not configured, unreachable, not found) omits the canonical fields (`matches_library_name` is omitted, not `false`) and the endpoint returns 200.

**Alternative considered**: compute the suggestion inside `GET /api/v1/downloads` alongside `duplicate_siblings`. Rejected — the list is rendered on every Downloads/Errors view and would add a live external call and N `os.Stat` calls per duplicate row.

**Alternative considered**: suggest by quality. Rejected for now — needs new client methods for Radarr/Sonarr file listings, and canonical name already captures the observed failure mode (French-titled artifact vs. canonical file).

### D3: Keep-one execution order

1. Resolve the group containing `:id` (404 `not_a_duplicate` otherwise).
2. `os.Stat` the kept path; if it does not exist → refuse with 409, nothing touched. This is the guard that stops "keep a ghost file" from deleting everything.
3. Preflight: for every other distinct path, `os.Lstat`. Absent → mark `already_missing`. Regular file → mark `to_delete`. Anything else (directory, symlink, device, ...) → refuse the whole request with 409, nothing touched.
4. For each distinct non-kept path: if `to_delete`, `os.Remove`; treat `ErrNotExist` as `already_missing`; any other error → stop, respond 500 with what already succeeded, leave this path's occurrences unchanged. Only after a successful removal (or already-missing) call `cancelDuplicateOccurrence` for each occurrence with that path.
5. If any file was deleted or any occurrence cancelled, request Refresh & Scan (D4) once.
6. Respond 200 with per-occurrence `outcome` (`kept` / `deleted` / `cancelled_file_missing`) and `refresh` outcome (`requested` / `failed` / `not_configured` / `not_found_in_library`).

The kept path is compared by cleaned path equality against each candidate before any removal, so occurrences that share the kept path are never treated as removal targets, and removing one path never affects another. Per-path "delete then cancel" (rather than delete everything, then cancel everything) keeps the database consistent with disk after a partial failure without needing a compensating transaction across the filesystem.

Only paths taken from the group's own `download_path` values are ever passed to `os.Remove`; nothing is derived from request input, so there is no path-traversal input to sanitize. Paths are `filepath.Clean`-ed before comparison.

**Alternative considered**: a single DB transaction wrapping all removals. Rejected — file removal is not transactional; either the DB would claim cancelled while a file survives, or a rollback would leave files deleted with the DB still saying `completed`.

### D4: Refresh & Scan is a new client method pair, best-effort, mirrored from `RescanMovie`/`RescanSeries`

`radarr.Client.RefreshMovie(ctx, movieID)` posts a `RefreshMovie` command to `/api/v3/command` for the given movie; `sonarr.Client.RefreshSeries(ctx, seriesID)` posts `RefreshSeries` for the given series. The exact command names and payload field names (the movie id may be a single id or an id list in Radarr's command model) are confirmed against the Radarr/Sonarr sources during implementation (as the existing `rescanMovieCommand` comment did), and pinned by client tests against a fake server, mirroring `TestRescanSeries`.

They are separate methods rather than a parameter on the existing rescan ones so the post-download path (which wants only a lightweight disk rescan) keeps its exact behavior and the on-demand path uses the operator-equivalent "Refresh & Scan". Both use `httpclient.Post` with the same retry/circuit-breaker handling.

The handler needs the Radarr movie / Sonarr series id: it comes from the same lookup as D2 (`Movie.ID` / `Series.ID`). Failure mapping: service not configured → `not_configured` (no call); lookup returns nil → `not_found_in_library`; lookup or command error → `failed`; success → `requested`. None of these change the HTTP status (200), because the deletion already happened.

**Alternative considered**: reuse `RescanMovie`/`RescanSeries`. Rejected — the operator asked for Refresh & Scan (the UI command), which is the broader, more stable contract (metadata refresh plus disk scan), and it stays correct even if a plain rescan's coverage changes.

### D5: Frontend flow

`ErrorsSidepanel.tsx` keeps the existing "verify files" button. A second control opens a picker: it calls `GET .../duplicate-candidates`, renders one radio row per distinct file (path, size, date, a "matches library name" mark, a "missing" state that disables selection), pre-selects the suggested row, and on submit shows a confirmation naming the kept file and the files to be deleted. Confirming calls `POST .../keep-duplicate` for the selected row's `download_info_id` (any of its occurrence ids), then renders the per-occurrence outcomes and the refresh outcome, and refreshes the Errors list. New i18n keys go in `locales/{en,fr}`. Existing `resolveDuplicateDownload` in `services/api.ts` is left alone; two new functions and types are added.

## Risks / Trade-offs

- **[Risk] Irreversible deletion of the wrong file** → Mitigation: operator picks the file, explicit confirmation naming both sides, only group-recorded paths, kept file re-verified to exist at request time, dirs/symlinks refused.
- **[Risk] The list is stale when the operator confirms** (another process changed the group) → Mitigation: the server re-derives the group from `:id` at request time and validates against the disk itself; the client's earlier view is never trusted for the delete set.
- **[Risk] Kept-path shared with an unrelated, non-group file** (a path recorded by another download that is not in this group) → Mitigation: deletion targets are exactly the group's non-kept paths; a path also used by an out-of-group occurrence is not expected (paths are per-file) but if the kept path equals one, nothing is deleted since the kept path is excluded from targets.
- **[Risk] Refresh is heavier than a rescan** → Mitigation: one call per operator action, rare and manual.
- **[Risk] Radarr/Sonarr re-imports or re-grabs the removed release** (the movie becomes "missing" in its view until refresh completes) → Mitigation: the action refreshes immediately; if Radarr's own automation re-searches, that is Radarr's configured behavior, outside this change. Noted for operators in docs.
- **[Risk] docker-compose users cannot see or delete media from the API container** → Mitigation: documented as a deployment note; the action fails safe (kept file `os.Stat` fails → 409, nothing touched).
- **[Trade-off]** Suggestion ignores quality; the operator sees size and date and can override.

## Migration Plan

- No schema change, no data migration, no feature flag: both endpoints are new and only run on explicit operator request.
- Deployment note (docs): the API service needs read-write access to the media volume (`/media`), as the Helm chart already provides.
- Rollback: revert the change; already-deleted files and cancelled occurrences are not restored (same as any manual cleanup), and `resolve-duplicate` remains functional.

## Open Questions

- Whether to also expose the refresh result as a toast in addition to the sidepanel message — a pure UI detail, decidable during implementation.
