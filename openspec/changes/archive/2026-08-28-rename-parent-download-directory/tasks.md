## 1. Backend: single-file move primitive

- [x] 1.1 Add a `moveSingleFile(src, dst string) error` helper in `internal/api/handlers_frontend.go` (or a shared file) that creates the destination directory, tries `os.Rename`, and falls back to copy+verify+delete — mirroring the fallback shape of `moveFile` in `internal/downloader/downloader.go:610-613` — and verify with a unit test covering both the rename-success path and the cross-device copy-fallback path (e.g. by forcing `os.Rename` to fail).
- [x] 1.2 Add a helper that detects whether a `download_info.download_path` sits under a `Season NN`-style directory (reusing the same prefix check as `moveTVShowFolder` in `internal/api/handlers_frontend.go:411-415`) and returns the series root, season dir, and season number; verify with unit tests for a TV path and a movie path (no season).

## 2. Backend: rename endpoint

- [x] 2.1 Add `RenameDownloadRequest` request struct (`new_name string`, `destination_parent_dir *string`) and implement `renameDownload(c *gin.Context)` in `internal/api/handlers_frontend.go`: load `DownloadInfo` by `:id`, return `not_found`/`404` if missing and a validation error if `download_path` is empty (not yet completed).
- [x] 2.2 In `renameDownload`, compute the new destination path using `sanitizeFilename(new_name)` following the existing `buildMovieBasePath`/`buildTVShowBasePath` conventions (`internal/downloader/path.go:8-18`) — reconstructing `Season NN` only when task 1.2 detected a TV path — and default the destination root to the current library root when `destination_parent_dir` is omitted; verify by unit-testing path computation for a movie, a TV episode with no destination root, and a TV episode with a destination root.
- [x] 2.3 In `renameDownload`, `os.Stat` the computed destination first and return `400` with `ErrorResponse.error = "rename_target_exists"` without touching disk or DB if it already exists; verify with a test that pre-creates the destination file and asserts no mutation occurs.
- [x] 2.4 In `renameDownload`, on a clear destination: move the single file with the task 1.1 helper, then update only the targeted `download_info.download_path` in a DB transaction, returning `rename_failed`/`database_update_failed` (mirroring `moveMovieFolder`'s error codes) on the respective failure; verify with a test asserting only the targeted `download_info` row changes and sibling rows are untouched.
- [x] 2.5 After a successful move, remove the original season directory (TV only) and then the original series/movie directory if each is left empty (`os.ReadDir` + `os.Remove`, best-effort/log-only on error); verify with a test that an emptied season+series directory pair is removed, and a test that a directory still containing sibling episodes is left intact.
- [x] 2.6 Register `POST /api/v1/downloads/:id/rename` in `internal/api/api.go` next to the existing `/movies/:id/move` and `/tvshows/:id/move` routes, and verify with an integration test hitting the route end-to-end (movie case and TV case).

## 3. Frontend: API client and dialog

- [x] 3.1 Add `api.renameDownload(id, { new_name, destination_parent_dir? })` in `frontend/src/services/api.ts` calling the new endpoint, and add/extend the `DownloadEnriched`-adjacent TypeScript types as needed; verify with `tsc --noEmit` passing.
- [x] 3.2 Create `frontend/src/components/RenameFolderDialog.tsx` modeled on `MoveFolderDialog.tsx`: a text field pre-filled with `file_info.folder_name`, an optional destination-root field, submit/cancel actions, and inline error display (translated) when the response error is `rename_target_exists` vs. other failures; verify by rendering the component in isolation (existing test setup, if any) or manual browser check per task 5.
- [x] 3.3 Add French and English translation strings for the new dialog's labels, helper text, and the `rename_target_exists`/generic-failure error messages in `frontend/src/locales/{fr,en}/dialogs.json`; verify by checking both locale files contain matching keys.

## 4. Frontend: wire the action into the Downloads page

- [x] 4.1 Add a "Renommer" action button per completed item in `frontend/src/components/DownloadsTab.tsx` next to the existing "Move" button, and open/close state wiring in `frontend/src/App.tsx` (mirroring `openMoveDialog`/`isMoveOpen`) scoped to the clicked item's `DownloadInfo.id`; verify by confirming the button renders only for completed downloads, matching the existing "Move" button's visibility rule.
- [x] 4.2 On successful rename, update only the affected card's local state (`file_info.folder_name`/`download_path`) without refetching or re-rendering unrelated cards; verify by manually renaming one item on a Downloads page with multiple items from the same series and confirming sibling cards remain visually unchanged.

## 5. Manual verification

- [x] 5.1 Start the app locally (per repo's `run` workflow), rename a movie's folder, and confirm the file moved on disk to `{root}/{new name}/{new name}.ext` and the Downloads card reflects the new path.
- [x] 5.2 Rename a single TV episode that has sibling episodes downloaded in the same series/season directory, and confirm only that episode's file moved to `{root}/{new name}/Season NN/...`, siblings remain untouched on disk and in the UI, and the original directory still exists (not empty).
- [x] 5.3 Rename the last remaining episode in a season directory and confirm the emptied season directory (and series directory, if also emptied) is removed from disk.
- [x] 5.4 Attempt a rename that collides with an existing directory/file and confirm the UI shows the translated `rename_target_exists` error and no file was moved.
