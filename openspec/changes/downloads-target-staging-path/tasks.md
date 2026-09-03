## 1. Data model

- [ ] 1.1 Add `TargetPath *string` and `StagingPath *string` (with `json:"target_path,omitempty"` / `json:"staging_path,omitempty"`) to `models.DownloadInfo` in `internal/models/download.go`, and verify `AutoMigrate` adds the two nullable columns on a fresh run (check the generated schema / run the app locally against a scratch DB)

## 2. Downloader: write target_path and staging_path

- [ ] 2.1 In `internal/downloader/state_manager.go`, extend the `downloading` case of `UpdateState` (or add a dedicated setter) to also persist `target_path = opts.BaseDestPath + detectFileExtension(url, "")`, and verify with a unit test that starting a download sets `target_path` to the URL-guessed final path
- [ ] 2.2 In `internal/downloader/downloader.go`, right after `tempPath` is computed (around line 152, before `retry.Do`), persist `staging_path = tempPath` on the `DownloadInfo` record, and verify with a unit test that `staging_path` is set once per `Download()` call and stays stable across retries within that call
- [ ] 2.3 In `updateDownloadInfoCompleted` (`internal/downloader/downloader.go:452`), clear `target_path` and `staging_path` to `nil` in the same update that sets `download_path`, and verify with a unit test that a completed download has both fields nil while `download_path` is set
- [ ] 2.4 Verify with a unit test that a download failing mid-transfer (HTTP/EOF error path, `internal/downloader/downloader.go:205-223`) leaves `target_path` and `staging_path` populated with the values from that attempt, and `download_path` still nil
- [ ] 2.5 Verify with a unit test that a download failing during the move-to-destination step (`internal/downloader/downloader.go:254-273`) leaves `target_path` and `staging_path` populated, and `download_path` still nil

## 3. API exposure

- [ ] 3.1 Add `TargetPath *string` and `StagingPath *string` (`json:"target_path,omitempty"` / `json:"staging_path,omitempty"`) to `DownloadEnrichedResponse` in `internal/api/types.go`
- [ ] 3.2 In `enrichDownloadInfo` (`internal/api/handlers_frontend.go`), copy `dl.TargetPath`/`dl.StagingPath` onto the response, and verify with a unit/integration test that `GET /api/v1/downloads` includes `target_path`/`staging_path` for a `downloading`/`failed` fixture and omits them for a `completed` one
- [ ] 3.3 Verify with a test (or by reading the code) that `rename`, `moveMovieFolder`, `moveTVShowFolder`, and `resume_helper.buildBaseDestPath` are unchanged and do not reference the two new fields

## 4. Frontend

- [ ] 4.1 Add `target_path?: string` and `staging_path?: string` to `DownloadEnriched` in `frontend/src/types.ts`
- [ ] 4.2 In `frontend/src/components/DownloadsTab.tsx`, update the File section render condition and content so that when `download_path` is absent, `target_path` (labeled as planned/not-final) and `staging_path` (labeled as temporary, may no longer exist) are shown if present, falling back to `url` only when none of the three are available; verify by inspecting the rendered sidepanel for a `downloading` and a `failed` fixture item (component test or manual run)
- [ ] 4.3 Add the new label keys to `frontend/src/locales/en/downloads.json` and `frontend/src/locales/fr/downloads.json`, and verify both locales render without missing-key warnings
- [ ] 4.4 Update/add cases in `frontend/src/components/DownloadsTab.test.tsx` covering: in-progress download shows target+staging path, failed download shows target+staging path, completed download shows only `download_path`, download with none of the three falls back to `url`

## 5. Verification

- [ ] 5.1 Run backend tests (`go test ./internal/downloader/... ./internal/api/...`) and frontend tests (`DownloadsTab.test.tsx`) and confirm all pass
- [ ] 5.2 Manually trigger a download that fails with a network error (e.g. point at an unreachable/truncated URL) and confirm the sidepanel shows the planned and temporary paths instead of the raw source URL
