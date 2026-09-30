# Proposal

## Why

`surface-duplicate-downloads` (archived 2026-09-25) made duplicated movie/episode downloads visible in the Erreurs tab and added a "Resolve duplicate" action. That action only reconciles the database with the filesystem: it cancels occurrences whose file is already gone and, by design, never touches a file. In practice this leaves the operator stuck. On a real case (Star Wars 1977), resolving cancelled the one occurrence whose file had already been removed, but the two remaining files (`Star Wars (1977).mkv` and `La Guerre des étoiles (1977).mp4`) both exist on disk and in Radarr, so the group is still flagged as a duplicate and nothing in the product can finish the job. The operator has to delete the extra file by hand, then hope Radarr/Sonarr notice, then run Resolve again.

## What Changes

- Add an operator-triggered "keep one, delete the others" action for a duplicate group: the operator picks which occurrence's file to keep; the system deletes every other distinct file of the group from disk, cancels the corresponding occurrences in the database, then asks Radarr/Sonarr to refresh and rescan the movie/series so their library state matches disk.
- Add a read-only "duplicate candidates" endpoint that describes the group's distinct files on demand (path, whether it still exists, size, latest completion date, and whether its name matches the live Radarr/Sonarr canonical name) and flags one existing file as suggested. The suggestion prefers the file whose name matches the canonical name, then the most recent; the operator can always choose another. It is computed on demand, not on every `GET /api/v1/downloads`, so no live Radarr/Sonarr call is added to the list view.
- The Erreurs sidepanel lists those files with the suggested one pre-selected, and requires an explicit confirmation before the destructive action runs.
- Add `RefreshMovie` (Radarr) and `RefreshSeries` (Sonarr) client methods — the "Refresh & Scan" command — and use them for this action. The existing `RescanMovie`/`RescanSeries` used after downloads are unchanged.
- **BREAKING (spec-level, deliberate)**: the `duplicate-download-detection` requirement that no resolve action may delete, move or modify a file on disk is relaxed for this one new, explicitly confirmed action. The existing "Resolve duplicate" (verify files) action remains database-only and keeps its guarantee.
- Deletion is bounded by safety rules: only paths already recorded as `download_path` of an occurrence in the identified group are deletable; the kept file is never deleted (including when several occurrences share its path); only regular files are removed (no directories, no symlink targets); files are removed before any database change and an unexpected removal failure aborts before the database is touched.
- A failed Radarr/Sonarr refresh never undoes the deletion; it is reported in the response so the operator can retry the refresh manually.

**Out of scope**: automatic (unattended) deletion of duplicates; deleting anything not recorded as a group occurrence's `download_path`; bulk resolution of several groups at once; removing the now-empty duplicate folder beyond the existing empty-directory cleanup helper.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `duplicate-download-detection`: adds the duplicate-candidates read endpoint, the "Keep One, Delete Others" action and its endpoint, its deletion safety rules, and the on-demand Radarr/Sonarr "Refresh & Scan" that follows it; scopes the existing "never modifies a file on disk" guarantee to the verify-only Manual Resolve Action.
- `downloads-errors-view`: the duplicate sidepanel gains a file picker with a suggested selection and a confirmation step for the keep-one action, and reflects per-occurrence and refresh outcomes; the "no other corrective action" carve-out is extended to cover this action.

## Impact

- `internal/api/duplicate_download.go`: new handler (e.g. `POST /api/v1/downloads/:id/keep-duplicate`, where `:id` is the `download_info` id of the occurrence to keep), deletion safety helpers, DB transaction reusing `cancelDuplicateOccurrence`'s shape; route registration in `internal/api/api.go`.
- `internal/external/radarr/radarr.go`, `internal/external/sonarr/sonarr.go`: new `RefreshMovie`/`RefreshSeries` methods (command names and payload fields to be confirmed against Radarr/Sonarr sources).
- `internal/api/duplicate_download.go`: also hosts the new read-only candidates handler (e.g. `GET /api/v1/downloads/:id/duplicate-candidates`); `GET /api/v1/downloads` and its `duplicate_siblings` field are unchanged.
- Frontend: `ErrorsSidepanel.tsx`, `services/api.ts`, `types.ts`, `locales/{en,fr}` for the picker, confirmation, and outcomes.
- Deployment: the API process needs read-write access to the media volume. The Helm `deployment.yaml` already mounts `/media` read-write; `docker-compose.yml`'s API service does not visibly mount it and would need a volume.
- No schema change; no change to detection, to the verify-only resolve action, or to the scheduled reconciliation pass.
