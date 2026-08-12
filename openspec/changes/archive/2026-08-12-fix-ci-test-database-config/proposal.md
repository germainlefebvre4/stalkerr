## Why

The `Go CI` "Test" job has failed on every run since at least 2026-08-06, independent of any feature work (confirmed on PR #6 and on prior runs of `main`). Two of the affected packages, `internal/downloader` and `internal/processor`, fail entirely because their test setup helpers misconfigure the database connection rather than because of any application defect. This blocks merging any PR, including unrelated feature work, since the `Test` check never goes green.

Once the database setup was fixed, fixing it unmasked further `internal/downloader` failures that were previously hidden behind the DB error (the DB error fired first, or the test was skipped outright for "database not reachable"). Those turned out to be two real, pre-existing production bugs in `internal/downloader/downloader.go` — this proposal's scope was extended (2026-08-07) to cover them too, since the `Test` job cannot go green without fixing them as well.

## What Changes

- `internal/downloader/downloader_test.go`: `setupTestDB` builds an in-memory SQLite `*gorm.DB` but then calls `database.Initialize()`, which discards it and opens a real Postgres connection using an unloaded (zero-value) `config.Get()`, producing an empty DSN (`host= port=0 user= password=`) and failing every test that touches the database. Change it to call the existing `database.SetDB(db)` helper so the in-memory database is actually installed as the package-global instance.
- `internal/downloader/downloader_test.go`: `TestNew/with_zero_values_uses_defaults` expects a default HTTP timeout of `300 * time.Second`, but `downloader.New` in `internal/downloader/downloader.go:54` defaults to `600 * time.Second`. Align the two so the default is defined once and asserted consistently (test expectation is the one aligned to the current documented default of 10 minutes, since that is the intended production behavior; the test's expected value is corrected to match).
- `internal/processor/processor_test.go`: `setupTestDB` falls back to hardcoded local dev values (`DB_HOST`, `DB_PORT=5433`, `DB_USER=stalkerr`, `DB_PASSWORD=stalkerr`, `DB_NAME=stalkerr`) whenever the bare `DB_*` env vars are unset. Because `config.bindEnvWithAlternatives` treats any non-empty alternative env var as an explicit `viper.Set()` override (highest precedence), this silently overwrites the real CI-provided `STALKEER_DATABASE_*` values (`localhost:5432`, `stalkeer`, `stalkeer_test`), pointing every processor test at a nonexistent Postgres instance on port 5433. Remove this hardcoded fallback so `setupTestDB` relies solely on `config.Load()` picking up whatever `STALKEER_DATABASE_*` (or already-set `DB_*`) values are present in the environment, matching how `internal/downloader` and other passing packages behave.
- **(scope extension) `internal/downloader/downloader.go` — extension detection bug**: `detectFileExtension` calls `filepath.Ext(url)` on the raw URL string instead of parsing it first. For a URL with no path (e.g. a bare `http://host:port`, as produced by `httptest.NewServer`), `filepath.Ext` treats the last `.` in the host's dotted IP as a file extension, e.g. returning `.1:38713` for `http://127.0.0.1:38713`. Fix: parse the URL with `net/url` and run `filepath.Ext` on `u.Path` instead of the raw string.
- **(scope extension) `internal/downloader/downloader.go` — retry bug**: in `downloadFileWithResume`, a non-2xx HTTP status on a non-resume request returns a plain `fmt.Errorf("unexpected status code: %d", ...)`. `apperrors.IsRetryable` only recognizes `*apperrors.AppError` with specific codes, so this plain error is never retried — even for transient 5xx responses — and `retry.Do` gives up after a single attempt. Fix: classify 5xx responses as a retryable `apperrors.AppError` (`apperrors.CodeServiceUnavailable`) so `retry.Do` actually retries them, while 4xx responses remain non-retryable (unchanged behavior, still covered by `TestDownload_HTTPErrors`).
- **(scope extension) `internal/downloader/downloader_test.go` — test expectation corrections**: `TestDownload_Success`, `TestDownload_Retry`, and `TestDownload_CreatesDestinationDirectory` assert the downloaded file ends up at exactly `BaseDestPath` with no extension appended. That was never actually achievable: `Download()` always appends the extension `detectFileExtension` resolves (defaulting to `.mkv` when neither the URL path nor `Content-Type` gives a hint, which is the case for all three tests' bodies). These tests could not have passed as written even before the CI outage; their expected paths are corrected to include the actually-appended extension.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
(none — these are bug fixes restoring already-intended behavior — extension detection from the real URL path, and retrying transient 5xx responses — not new or changed product requirements; no existing spec documents this internal `downloader` behavior)

## Impact

- `internal/downloader/downloader_test.go` (test helper, one test's default-timeout expectation, three tests' expected destination paths)
- `internal/downloader/downloader.go` (production code: `detectFileExtension` URL parsing, `downloadFileWithResume` status-code error classification)
- `internal/processor/processor_test.go` (test helper)
- `internal/database/database.go` behavior is unchanged.
- Expected outcome: the `Test` CI job's `internal/downloader` and `internal/processor` packages pass again, unblocking the `Test` check on PR #6 and future PRs. The `Lint` job failure (golangci-lint v2 config schema) is out of scope — already being handled separately.
