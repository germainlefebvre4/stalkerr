## Why

`make test` and CI (last 6 runs) both fail entirely inside `internal/processor`: its tests are the only ones in the repo that call `config.Load()` + a real Postgres via `database.Initialize()`, instead of the in-memory SQLite pattern (`database.SetDB` + `testutil`-style setup) used by every other package. A recent config hardening (`5822d8b`, removing the legacy single-file `m3u_file_path` fallback) made `m3u.sources` a config-file-only, non-empty-list requirement with no environment-variable equivalent, which this package's test setup cannot satisfy. The result: 27 tests failing locally and in CI, blocking a trustworthy `make test`.

## What Changes

- Rewrite the shared `setupTestDB`/`teardownTestDB` helpers in `internal/processor/processor_test.go` to spin up an in-memory SQLite `*gorm.DB` (auto-migrating `Movie`, `TVShow`, `ProcessedLine`, `ProcessingLog`, `ManualMapping`), inject it via `database.SetDB`, and set an empty `config.SetConfig(&config.Config{})` — mirroring the existing pattern in `cmd/process_test.go`. No individual test body changes required; all 27 tests route through these two helpers.
- Remove the now-unneeded `config.Load()` / `database.Initialize()` / Postgres `TRUNCATE ... CASCADE` calls from that setup path.
- Remove the now-unused Postgres `services:` block and its `STALKEER_DATABASE_*` / stale `STALKEER_M3U_FILE_PATH` env vars from `.github/workflows/ci.yml`, since no test in the suite will exercise a real Postgres connection anymore.
- No production code changes; this only touches test setup and CI configuration.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
(none — this is a test-infrastructure and CI fix with no spec-level behavior change; see `skip_specs: true` in `.openspec.yaml`)

## Impact

- `internal/processor/processor_test.go` (setup/teardown helpers only)
- `.github/workflows/ci.yml` (drop obsolete Postgres service + env vars)
- No API, schema, or runtime behavior changes; no impact on production code paths
- Net effect: `make test` and CI both pass without requiring a live Postgres instance for the test suite
