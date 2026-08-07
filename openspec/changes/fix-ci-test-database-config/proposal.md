## Why

The `Go CI` "Test" job has failed on every run since at least 2026-08-06, independent of any feature work (confirmed on PR #6 and on prior runs of `main`). Two of the affected packages, `internal/downloader` and `internal/processor`, fail entirely because their test setup helpers misconfigure the database connection rather than because of any application defect. This blocks merging any PR, including unrelated feature work, since the `Test` check never goes green.

## What Changes

- `internal/downloader/downloader_test.go`: `setupTestDB` builds an in-memory SQLite `*gorm.DB` but then calls `database.Initialize()`, which discards it and opens a real Postgres connection using an unloaded (zero-value) `config.Get()`, producing an empty DSN (`host= port=0 user= password=`) and failing every test that touches the database. Change it to call the existing `database.SetDB(db)` helper so the in-memory database is actually installed as the package-global instance.
- `internal/downloader/downloader_test.go`: `TestNew/with_zero_values_uses_defaults` expects a default HTTP timeout of `300 * time.Second`, but `downloader.New` in `internal/downloader/downloader.go:54` defaults to `600 * time.Second`. Align the two so the default is defined once and asserted consistently (test expectation is the one aligned to the current documented default of 10 minutes, since that is the intended production behavior; the test's expected value is corrected to match).
- `internal/processor/processor_test.go`: `setupTestDB` falls back to hardcoded local dev values (`DB_HOST`, `DB_PORT=5433`, `DB_USER=stalkerr`, `DB_PASSWORD=stalkerr`, `DB_NAME=stalkerr`) whenever the bare `DB_*` env vars are unset. Because `config.bindEnvWithAlternatives` treats any non-empty alternative env var as an explicit `viper.Set()` override (highest precedence), this silently overwrites the real CI-provided `STALKEER_DATABASE_*` values (`localhost:5432`, `stalkeer`, `stalkeer_test`), pointing every processor test at a nonexistent Postgres instance on port 5433. Remove this hardcoded fallback so `setupTestDB` relies solely on `config.Load()` picking up whatever `STALKEER_DATABASE_*` (or already-set `DB_*`) values are present in the environment, matching how `internal/downloader` and other passing packages behave.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
(none — this is a test-infrastructure/CI fix; no application or API behavior changes)

## Impact

- `internal/downloader/downloader_test.go` (test helper + one test's expected value)
- `internal/processor/processor_test.go` (test helper)
- No production code paths change; `internal/downloader/downloader.go` and `internal/database/database.go` behavior is unchanged (only the test's expectation is corrected to match `downloader.go`'s existing default).
- Expected outcome: the `Test` CI job's `internal/downloader` and `internal/processor` packages pass again, unblocking the `Test` check on PR #6 and future PRs. The `Lint` job failure (golangci-lint v2 config schema) is out of scope — already being handled separately.
