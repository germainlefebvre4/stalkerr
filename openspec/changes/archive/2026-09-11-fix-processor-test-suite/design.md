## Context

See `proposal.md` - Why. `internal/processor`'s `setupTestDB`/`teardownTestDB` (defined once in `processor_test.go`, reused by `backfill_metadata_test.go`, `enrich_tvdb_test.go`, `tmdb_metadata_test.go` — 27 tests total) are the only test setup in the repo still calling `config.Load()` + `database.Initialize()` against a real Postgres. Every other package (`internal/api`, `internal/downloader`, `internal/scheduler`, `cmd/process_test.go`) already uses an in-memory SQLite `*gorm.DB` injected via `database.SetDB`, with `config.SetConfig(...)` supplying a minimal config instead of loading one. `internal/processor`'s own production code (`processor.go`, `backfill_metadata.go`, `enrich_tvdb.go`, `tmdb_metadata.go`) contains no Postgres-specific SQL (no `ILIKE`, `JSONB`, `ON CONFLICT`) and reads no fields from `config.Get()`, so it is not tied to Postgres or to any config value.

## Goals / Non-Goals

**Goals:**
- Make `internal/processor`'s tests self-contained (no live database, no env vars, no config file needed).
- Reuse the exact pattern already proven in `cmd/process_test.go` rather than inventing a new one.
- Remove the now-dead Postgres service from CI once nothing in the suite needs it.

**Non-Goals:**
- Changing `internal/processor`'s production code or its public API (e.g., not refactoring `NewProcessor` to accept an injected `*gorm.DB` — out of scope, larger blast radius, not needed to fix the failure).
- Adding new test coverage or behavior assertions — this only replaces *how* existing tests reach a database, not *what* they assert.
- Changing the `m3u.sources` validation or reintroducing an env-var path for it (that hardening in `5822d8b` was intentional; see Option B considered/rejected below).

## Decisions

**1. Fix via `database.SetDB` + `config.SetConfig`, not constructor injection.**
`internal/processor`'s functions (`NewProcessor`, etc.) call the package-level `database.Get()` singleton internally, unlike `internal/matcher`/`internal/api` which take `db *gorm.DB` as an explicit parameter. Changing `internal/processor`'s signatures to match would touch every call site (`cmd/process.go`, etc.) for no behavioral gain. `database.SetDB` already exists precisely for this: swap the singleton before the test runs, leave production code untouched. This is also exactly what `cmd/process_test.go:19-46` already does for the same package's higher-level orchestration tests.

**2. In-memory SQLite, one connection per test, `SetMaxOpenConns(1)`.**
A `sqlite.Open(":memory:")` database is private per connection; without capping the pool at 1, GORM can open a second connection mid-test and silently see an empty, un-migrated database (a known footgun already documented as a comment in `cmd/process_test.go`). Every helper we mirror does this; the rewritten `setupTestDB` must too.

**3. Migrate `Movie`, `TVShow`, `ProcessedLine`, `ProcessingLog`, `ManualMapping`, and `FilterConfig`.**
The first five match the current `TRUNCATE TABLE processed_lines, processing_logs, movies, tvshows, manual_mappings` list being removed — no table gets silently dropped from test setup. `FilterConfig` is additionally required: `TestProcess_BackfillSkippedWhenSkipTMDB` calls `filter.NewManager().LoadAll()` directly, which queries `filter_configs` — a table the old Postgres-backed setup provided implicitly via `database.Initialize()`'s full production migration (`internal/database/database.go`), but that the TRUNCATE list (a cleanup list, not a schema inventory) omitted. Discovered when task 2.1's verification run left this one test failing with `no such table: filter_configs` after an initial migration of only the first five models. `cmd/process_test.go`'s `setupProcessTestDB` already migrates `FilterConfig` (among others) for the same reason.

**4. Drop `teardownTestDB` (and its call sites) rather than keep a Postgres-shaped cleanup.**
No other in-memory-SQLite test in the repo tears down between tests — each test gets its own fresh `:memory:` instance, so explicit truncation/close is redundant. Keeping a no-op `teardownTestDB` was considered (smaller diff: 1 function body vs. removing 27 `defer` call sites) but leaves dead code and an inconsistent pattern; removing it matches convention. Implementation should do a straight removal since it's a mechanical, low-risk edit across call sites that all look identical.

**5. `config.SetConfig(&config.Config{})` — empty config, no `config.Load()`.**
Confirmed no code path under `internal/processor` reads `config.Get()`, so an empty struct is sufficient (same as `cmd/process_test.go`).

**6. Also remove the Postgres `services:` block from `.github/workflows/ci.yml` in this change, not a follow-up.**
Once this fix lands, no test in the entire suite touches a real Postgres connection — confirmed by grepping for `config.Load()` usage outside `cmd/*.go` production entrypoints. Leaving a live but unused Postgres service in CI is misleading (implies some test still needs it) and costs CI startup time for nothing. Doing it in the same change keeps "why the service was removed" attached to the commit that made it unnecessary, instead of relying on someone rediscovering that later.

**Alternatives considered (rejected):**
- *Give `m3u.sources` an env-var binding for tests* (original Option B): would partially undo the deliberate hardening from `5822d8b`, and doesn't fix the more fundamental issue that this package's tests would still depend on a live Postgres reachable from `localhost` in every dev environment and in CI.
- *Drop a `config.yml` fixture discoverable from `internal/processor/`'s test working directory* (original Option C): keeps a real-Postgres dependency for no reason once SQLite works fine for every assertion these tests make; also doesn't fix CI's stale `STALKEER_M3U_FILE_PATH` on its own.
- *Makefile-only env export* (original Option D): doesn't address the `m3u.sources` failure at all; CI would stay red.

## Risks / Trade-offs

- **[Risk]** Some assertion in the 27 tests implicitly relies on Postgres-only behavior (e.g., case-insensitive collation, specific error text from a Postgres driver) → **Mitigation:** run the full `internal/processor` suite locally against SQLite after the rewrite and inspect any failure; none is expected given the earlier grep found no dialect-specific SQL, but this is the one assumption that needs empirical confirmation, not just static analysis.
- **[Risk]** Removing the CI Postgres service could be premature if some other, currently-passing test relies on it in a way this exploration missed → **Mitigation:** the “Impact” task should re-grep for `config.Load()` / `database.Initialize()` usage across `_test.go` files right before touching `ci.yml`, as a final confirmation gate rather than relying solely on this design's earlier finding.

## Migration Plan

1. Rewrite `setupTestDB`/`teardownTestDB` in `internal/processor/processor_test.go`; remove `teardownTestDB` and its `defer` call sites in all 4 test files.
2. Run `go test ./internal/processor/...` locally (no env vars, no Postgres) to confirm all 27 tests pass.
3. Run the full `make test` to confirm no regression elsewhere.
4. Re-confirm no other `_test.go` still needs Postgres, then remove the `services:` Postgres block and the `STALKEER_DATABASE_*` / `STALKEER_M3U_FILE_PATH` env vars from the `Run tests` step in `.github/workflows/ci.yml`.
5. No rollback complexity: this is test-only + CI-config-only, fully reversible with a revert, no data migration involved.
