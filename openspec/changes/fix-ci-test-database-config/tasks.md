## 1. Fix `internal/downloader` test database setup

- [x] 1.1 In `internal/downloader/downloader_test.go`, replace the `database.Initialize()` call in `setupTestDB` with `database.SetDB(db)` so the in-memory SQLite database created in that function is actually installed as the package-global instance.
- [x] 1.2 Reconcile the default HTTP timeout: `internal/downloader/downloader.go:54` defaults to `600 * time.Second`; update the `wantTimeout` expectation in `TestNew/with_zero_values_uses_defaults` (in `internal/downloader/downloader_test.go`) from `300 * time.Second` to `600 * time.Second` to match.

## 2. Fix `internal/downloader` production bugs (scope extension)

- [x] 2.1 Fix `detectFileExtension` in `internal/downloader/downloader.go` to parse the URL with `net/url` and run `filepath.Ext` on the parsed path instead of the raw URL string, so a path-less URL (e.g. `http://127.0.0.1:38713`) doesn't produce a garbage extension like `.1:38713` from the host's dotted IP.
- [x] 2.2 Fix the retry classification in `downloadFileWithResume` (`internal/downloader/downloader.go`): wrap non-2xx status codes as a retryable `apperrors.AppError` (`apperrors.CodeServiceUnavailable`) for 5xx responses so `apperrors.IsRetryable` lets `retry.Do` actually retry transient server errors; keep 4xx responses non-retryable (unchanged, still covered by `TestDownload_HTTPErrors`).
- [x] 2.3 Run `go test ./internal/downloader/... -run . -v`, observe the exact resulting failures in `TestDownload_Success`, `TestDownload_Retry`, `TestDownload_CreatesDestinationDirectory` (their expected destination paths don't account for the extension `Download()` always appends), and correct those tests' expected paths to include the actually-appended extension.
- [x] 2.4 Run `go test ./internal/downloader/... -run . -v` again and confirm the full package passes (`TestNew`, `TestDownload_Success`, `TestDownload_WithDatabaseTracking`, `TestDownload_Retry`, `TestDownload_DatabaseStateOnFailure`, `TestDownload_CreatesDestinationDirectory`, `TestDownload_RetryCountIncrements`, and everything else previously passing still passes).

## 3. Fix `internal/processor` test database setup

- [x] 3.1 In `internal/processor/processor_test.go`, remove the hardcoded `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` fallback block in `setupTestDB` (the `if os.Getenv("DB_*") == "" { os.Setenv(...) }` calls) so it no longer overrides whatever `STALKEER_DATABASE_*` (or `DB_*`) values are already present in the environment.
- [x] 3.2 Confirm `config.Load()` is still called before `database.Initialize()` in `setupTestDB` (it already is) so the real environment-provided database config is picked up unmodified.
- [x] 3.3 Run `go test ./internal/processor/... -run . -v` locally against a local Postgres matching the CI service (`localhost:5432`, user/db as set in `.github/workflows/ci.yml`) and confirm all previously-failing tests (`TestEnrichMissingTVDBIDs_*`, `TestNewProcessor`, `TestProcessBasic`, `TestProcessWithLimit`, `TestProcessDuplicates`, `TestProcessWithForce`, `TestProcessingLogCreation`, `TestProcessWithManualMapping`) now pass.

## 4. Verify in CI

- [x] 4.1 Push the fix and confirm the `Test` job's `internal/downloader` and `internal/processor` packages report `ok` on PR #6 (the `Lint` job failure is out of scope - tracked separately).
- [x] 4.2 Confirm no other package's tests regressed as a side effect (full `go test ./...` green aside from the already-tracked `Lint` issue). Verified locally, package by package (excluding the gitignored local `data/` docker volume, which `go test ./...` cannot even list due to filesystem permissions on this machine - unrelated to CI, which has no such directory): all packages pass. `internal/config`'s `TestLoad_WithDefaults` only failed when run with `STALKEER_DATABASE_PORT=5433` exported for the local Postgres in this sandbox - a self-inflicted collision with this local run's env, not a code regression; it passes cleanly on its own or with the CI's actual `STALKEER_DATABASE_PORT=5432`.
