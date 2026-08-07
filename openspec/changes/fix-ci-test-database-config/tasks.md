## 1. Fix `internal/downloader` test database setup

- [x] 1.1 In `internal/downloader/downloader_test.go`, replace the `database.Initialize()` call in `setupTestDB` with `database.SetDB(db)` so the in-memory SQLite database created in that function is actually installed as the package-global instance.
- [x] 1.2 Reconcile the default HTTP timeout: `internal/downloader/downloader.go:54` defaults to `600 * time.Second`; update the `wantTimeout` expectation in `TestNew/with_zero_values_uses_defaults` (in `internal/downloader/downloader_test.go`) from `300 * time.Second` to `600 * time.Second` to match.
- [ ] 1.3 Run `go test ./internal/downloader/... -run . -v` locally and confirm all previously-failing tests (`TestNew`, `TestDownload_Success`, `TestDownload_WithDatabaseTracking`, `TestDownload_Retry`, `TestDownload_DatabaseStateOnFailure`, `TestDownload_CreatesDestinationDirectory`) now pass.

## 2. Fix `internal/processor` test database setup

- [ ] 2.1 In `internal/processor/processor_test.go`, remove the hardcoded `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` fallback block in `setupTestDB` (the `if os.Getenv("DB_*") == "" { os.Setenv(...) }` calls) so it no longer overrides whatever `STALKEER_DATABASE_*` (or `DB_*`) values are already present in the environment.
- [ ] 2.2 Confirm `config.Load()` is still called before `database.Initialize()` in `setupTestDB` (it already is) so the real environment-provided database config is picked up unmodified.
- [ ] 2.3 Run `go test ./internal/processor/... -run . -v` locally against a local Postgres matching the CI service (`localhost:5432`, user/db as set in `.github/workflows/ci.yml`) and confirm all previously-failing tests (`TestEnrichMissingTVDBIDs_*`, `TestNewProcessor`, `TestProcessBasic`, `TestProcessWithLimit`, `TestProcessDuplicates`, `TestProcessWithForce`, `TestProcessingLogCreation`, `TestProcessWithManualMapping`) now pass.

## 3. Verify in CI

- [ ] 3.1 Push the fix and confirm the `Test` job's `internal/downloader` and `internal/processor` packages report `ok` on PR #6 (the `Lint` job failure is out of scope - tracked separately).
- [ ] 3.2 Confirm no other package's tests regressed as a side effect (full `go test ./...` green aside from the already-tracked `Lint` issue).
