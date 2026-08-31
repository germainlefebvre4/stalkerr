## Why

On the Downloads page, renaming a TV episode's parent folder pre-fills the dialog with the season directory's name (e.g. `Season 01`) instead of the series' own folder name (e.g. `Breaking Bad (2008)`), because the dialog reuses the generic `file_info.folder_name` field (the file's immediate parent directory) as its pre-fill source. The rename endpoint itself already correctly targets the series root when a `Season NN` pattern is detected, so a user who submits that pre-filled value ends up renaming the whole series folder to the season's name, which looks like "the rename only affected the season directory" instead of the tvshow's parent directory. Movies are unaffected because a movie file's immediate parent directory already is the correct rename target.

## What Changes

- Add a new field to the downloads enrichment API response that reports the correct folder name to pre-fill for a rename: the series root folder name for a TV episode (reusing the existing `detectTVSeasonPath` detection already used by the rename endpoint), or the same value as `file_info.folder_name` for a movie or any non-TV-season download.
- Update the Downloads tab's rename dialog to pre-fill from this new field instead of `file_info.folder_name`.
- No change to the `POST /api/v1/downloads/:id/rename` endpoint's own path-resolution logic (`computeRenameDestination`/`detectTVSeasonPath`), which is already correct and tested.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `downloads-enrichment-api`: `DownloadEnrichedResponse` gains a field carrying the correct rename-target folder name, computed per download using the existing TV-season detection.
- `downloads-display-ui`: the "Rename Download Item" requirement's pre-fill source changes from `file_info.folder_name` to the new field.

## Impact

- Backend: `internal/api/types.go` (or wherever `DownloadEnrichedResponse`/`FileInfo` are declared), `internal/api/handlers_frontend.go` (enrichment loop, reuse of `detectTVSeasonPath`).
- Frontend: `frontend/src/types.ts` (`DownloadEnriched` type), `frontend/src/App.tsx` (`openRenameDialog`), `frontend/src/components/RenameFolderDialog.tsx` is unaffected (still just reads `renameItem.folderName`).
- No database schema change, no change to the rename endpoint's request/response contract.
