## 1. Backend: clear stale error_message on new attempt

- [x] 1.1 In `internal/downloader/state_manager.go`, update `StateManager.UpdateState`'s `case models.DownloadStatusDownloading` branch to also set `updates["error_message"] = nil`, and verify with a unit test that a `DownloadInfo` with `status=failed` + non-empty `error_message` has `error_message` set to `NULL` in the DB immediately after `UpdateState(ctx, id, DownloadStatusDownloading, nil)` is called.
- [x] 1.2 Add/extend a unit test in `internal/downloader` covering a full failed-then-retried-successfully cycle (`downloading` -> `failed` with error -> `downloading` -> `completed`) and assert the final `DownloadInfo` row has `status="completed"` and `error_message=nil`.
- [x] 1.3 Add a unit test asserting a first-ever attempt (fresh `DownloadInfo`, `error_message` already nil) transitioning to `downloading` leaves `error_message` nil (no regression/no-op case).

## 2. Frontend: gate the error banner on failed status

- [x] 2.1 In `frontend/src/components/DownloadsTab.tsx`, change the error banner render condition at the `item.error_message &&` check (around line 325) to `item.status === 'failed' && item.error_message &&`, and verify by reading the diff that no other conditional in the file still renders the banner unconditionally.
- [x] 2.2 Run the frontend test suite (`npm test` or project equivalent in `frontend/`) and confirm no existing DownloadsTab test regresses; add/update a test that renders a card with `status: 'completed'` and a non-empty `error_message` and asserts the error banner is absent.
- [x] 2.3 Add/update a test that renders a card with `status: 'failed'` and a non-empty `error_message` and asserts the error banner is still present (no regression on the primary failed-download use case).

## 3. Verification

- [x] 3.1 Run `go build ./...` and `go test ./internal/downloader/...` and confirm both succeed. (`./...` fails on an unrelated, pre-existing permission-denied `data/postgres/18/docker` directory unrelated to this change; built `./cmd/... ./internal/... ./webserver/... ./m3u_playlist/...` instead, which succeeded. `go test ./internal/downloader/...` passed.)
- [x] 3.2 Run the frontend build (`npm run build` in `frontend/`) and confirm it succeeds with the updated `DownloadsTab.tsx`.
- [x] 3.3 Manually reproduce the original scenario locally (or via a DB fixture): a `DownloadInfo` row with `status='failed'`, `error_message` set, then trigger a new download attempt on the same `ProcessedLine` and confirm the Downloads tab shows the card as completed with no error banner once it succeeds. (Reproduced via DB fixture rather than a live app run: the project only supports Postgres, not sqlite, for a real run, and the local `data/postgres` directory is permission-denied in this environment. `TestUpdateState_FailedThenRetriedSuccessfully` in `internal/downloader/state_manager_test.go` reproduces the exact DB-level scenario — same `DownloadInfo` row: `failed` with `error_message` set -> `downloading` -> `completed` -> asserts `error_message` is nil in the DB. `DownloadsTab error banner` in `frontend/src/components/DownloadsTab.test.tsx` reproduces the resulting UI state — a card with `status: 'completed'` and a leftover `error_message` -> asserts the banner is absent. Together these confirm the fix end-to-end.)
