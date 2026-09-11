## 1. Rewrite processor test DB setup

- [x] 1.1 Rewrite `setupTestDB` in `internal/processor/processor_test.go` to open an in-memory SQLite `*gorm.DB` (`gorm.Open(sqlite.Open(":memory:"), ...)`), `AutoMigrate` `models.Movie`, `models.TVShow`, `models.ProcessedLine`, `models.ProcessingLog`, `models.ManualMapping`, `models.FilterConfig`, cap the pool with `sqlDB.SetMaxOpenConns(1)`, then call `database.SetDB(db)` and `config.SetConfig(&config.Config{})` — remove the `config.Load()` / `database.Initialize()` / `TRUNCATE ... CASCADE` calls. Verify by reading the rewritten function back and confirming it compiles (`go vet ./internal/processor/...`). (`FilterConfig` added beyond the original list: `TestProcess_BackfillSkippedWhenSkipTMDB` needs it — see design.md decision 3.)
- [x] 1.2 Remove `teardownTestDB` and its `defer teardownTestDB(t)` call sites from `processor_test.go`, `backfill_metadata_test.go`, `enrich_tvdb_test.go`, and `tmdb_metadata_test.go`. Verify with `grep -rn "teardownTestDB" internal/processor/` returning no matches.

## 2. Verify the fix

- [x] 2.1 Run `go test ./internal/processor/...` with no `STALKEER_DATABASE_*`/`DB_*` env vars set and no Postgres reachable; verify all 27 previously-failing tests pass.
- [x] 2.2 Run the full `make test`; verify it exits 0 with every package `ok`.
- [x] 2.3 Confirm no remaining test needs a real database: `grep -rln "config.Load()" --include=*_test.go .` must return no matches. Verify by running the command and inspecting the (empty) output.

## 3. Clean up CI

- [x] 3.1 Remove the `services.postgres` block and the `STALKEER_DATABASE_*` / `STALKEER_M3U_FILE_PATH` env vars from the `Run tests` step in `.github/workflows/ci.yml`, now that no test depends on them (confirmed by task 2.3). Verify by re-reading the edited step and confirming only `go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...` remains, with no leftover `env:` block for those vars.
- [x] 3.2 Push the branch and confirm the `Go CI` workflow run is green (`gh run list --workflow=ci.yml --limit 1`), ending the current 6-run failure streak.
