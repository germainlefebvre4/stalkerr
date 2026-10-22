# Tasks

## 1. Safety net in `settings.Effective()`

- [x] 1.1 Log a `Warn` in `internal/settings/settings.go` when `Effective()` runs with `database.Get() == nil`, keeping the returned config unchanged; verify with a unit test in `internal/settings` that asserts the warning and the file/env fallback
- [x] 1.2 Add a unit test showing that, with a database set via `database.SetDB`, a stored `radarr.api_key` override is applied by `Effective()`; run `go test ./internal/settings/...`

## 2. Shared CLI bootstrap helper

- [x] 2.1 Add `loadEffectiveConfig(requireDB bool)` in `cmd/` implementing: bootstrap logger from `config.Get()`, `database.Initialize()`, `settings.Effective()`, logger re-initialisation with the effective levels/format; verify with a table-driven test covering `requireDB` true/false with a reachable and an unreachable database
- [x] 2.2 Cover in that test the best-effort path: unreachable database + `requireDB=false` returns the file/env config and logs the "overrides not applied" warning; and `requireDB=true` returns an error

## 3. Migrate the commands that already use the database

- [x] 3.1 Switch `cmd/download.go` to the helper (`requireDB=true`) and remove its late `database.Initialize()`; add a test in `cmd/download_test.go` where a stored `radarr.api_key` override is the key sent in `X-Api-Key` to a stub Radarr server while the file/env key is different (reproduces the reported 401)
- [x] 3.2 Switch `cmd/process.go`, `cmd/resume_downloads.go`, `cmd/enrich_tvdb.go`, `cmd/db_prune.go`, `cmd/reset.go` to the helper (`requireDB=true`); verify with `go build ./...` and the existing `cmd` tests
- [x] 3.3 Switch `cmd/migrate.go` to the helper and verify that a missing `settings_overrides` table on a fresh database does not abort the migration (test)

## 4. Migrate the commands that must work without a database

- [x] 4.1 Audit `cmd/m3u.go` subcommands: determine whether each one reaches the database (directly or through a helper) and choose `requireDB` accordingly; record the outcome in the commit message
- [x] 4.2 Switch `cmd/m3u.go`, `cmd/cleanup.go` and `cmd/config_cmd.go` to the helper with `requireDB=false`; verify that `cleanup` and `config` still run with an unreachable database (warning logged, exit code 0) and that `m3u` uses overridden sources when the database is reachable (test using `EffectiveSources`)

## 5. Documentation and final checks

- [x] 5.1 Add a `CHANGELOG.md` entry describing that CLI jobs now honor settings overrides, and note in `README.md` (configuration section) that stored overrides take precedence over file/env values for jobs as well as the server; verify the text matches the spec's "Command-Line Commands Apply Stored Overrides" requirement
- [x] 5.2 Run `go vet ./... && go test ./...` and confirm no remaining `settings.Effective()` call precedes a database connection in `cmd/` (`grep -n "settings.Effective()" cmd/*.go` shows only the helper)
- [x] 5.3 On a cluster or local `docker compose`, run the `download` command with an outdated file/env Radarr key and a valid override stored via the UI, and confirm the run no longer fails with a 401
