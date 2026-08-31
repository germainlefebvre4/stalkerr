## 1. Backend: rename target folder name

- [x] 1.1 Extract the series-root-name lookup out of `detectTVSeasonPath` (`internal/api/handlers_frontend.go`) into a small reusable helper, and have `computeRenameDestination` call it unchanged, verifying `go test ./internal/api/... -run TestComputeRenameDestination|TestDetectTVSeasonPath|TestRenameDownload` still passes.
- [x] 1.2 Add `RenameFolderName *string \`json:"rename_folder_name,omitempty"\`` to `DownloadEnrichedResponse` (`internal/api/dto.go` or `internal/api/types.go`, wherever the struct lives) and populate it in `listDownloadsEnriched`'s enrichment loop: series root name (via the helper from 1.1) for a TV episode under a `Season NN` folder, otherwise the same value as `file_info.folder_name`, and omitted when `download_path` is null.
- [x] 1.3 Add/extend a Go test in `internal/api/handlers_frontend_test.go` covering: a movie download reports `rename_folder_name` equal to `file_info.folder_name`; a TV episode under `Season 01` reports the series folder name in `rename_folder_name` while `file_info.folder_name` stays `"Season 01"`; a download with a null `download_path` omits `rename_folder_name`. Verify with `go test ./internal/api/...`.

## 2. Frontend: use the new field for rename pre-fill

- [x] 2.1 Add `rename_folder_name?: string` to the `DownloadEnriched` type in `frontend/src/types.ts`, alongside `file_info`.
- [x] 2.2 Update `openRenameDialog` in `frontend/src/App.tsx` to pre-fill `folderName` from `item.rename_folder_name` (falling back to the existing `item.file_info?.folder_name || filename` when `rename_folder_name` is absent), and verify by inspecting the diff that `RenameFolderDialog.tsx` itself needs no change.
- [x] 2.3 Update or add a frontend test (e.g. in `frontend/src/components/DownloadsTab.test.tsx` or a new test for the rename flow) asserting that opening the rename dialog for a TV episode item pre-fills the series folder name rather than the season folder name. Verify with the frontend test runner.

## 3. Manual verification

- [ ] 3.1 Run the app, open the Downloads tab, click "Renommer" on a completed TV episode download, and confirm the dialog pre-fills the series' folder name (not `Season NN`); repeat for a completed movie download and confirm no change in behavior.
