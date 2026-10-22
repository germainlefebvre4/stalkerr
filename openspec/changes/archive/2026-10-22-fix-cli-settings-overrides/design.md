# Design

## Context

`settings.Effective()` copies `config.Get()` and applies rows of `settings_overrides` read through `database.Get()`. When `database.Get()` is `nil`, `loadOverridesMap()` returns an empty map without any log. The API server is unaffected: it connects the database at startup and its handlers call `Effective()` per request.

Each CLI command, by contrast, takes a single snapshot at the top of its `Run`:

```
config.Load() -> cfg := settings.Effective() -> logger(cfg) -> database.Initialize()
                         ^ db == nil here: 0 override
```

Constraints that shape the fix:
- `database.Initialize()` reads `config.Get()` (bootstrap fields, never overridable), so it does not depend on `Effective()`. The database can be connected first without a cycle.
- `database.Initialize()` logs through `logger.AppLogger()`, while `logging.app.level`, `logging.database.level` and `logging.format` are overridable, hence stored in the database. This is a logger -> database -> logger dependency.
- Several commands (`cleanup`, `config`, `m3u download`) do not use the database today.
- See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- Every CLI command applies stored overrides before using any overridable field.
- The failure mode (overrides skipped) is never silent again.
- One shared bootstrap sequence instead of per-command ordering, so new commands cannot reintroduce the bug by getting the order wrong.

**Non-Goals:**
- Changing the server's per-request `Effective()` usage.
- Changing the Helm chart, the Secret layout or the `RADARR_API_KEY` / `SONARR_API_KEY` env injection.
- Adding new overridable fields or touching the bootstrap/overridable split.
- Making overrides hot-reload inside a long-running command: one resolution per run is enough.

## Decisions

### D1. A shared bootstrap helper in `cmd/`

Add one helper used by every command, replacing the ad-hoc `Effective()` call:

```
config.Load()                          (unchanged, done by the command)
loadEffectiveConfig(requireDB bool):
  1. logger.InitializeLoggers(...) from config.Get()   bootstrap logger
  2. database.Initialize()
       fail  + requireDB  -> return error (command exits non-zero)
       fail  + !requireDB -> Warn "overrides not applied", fall back to config.Get()-based Effective()
  3. cfg := settings.Effective()
  4. re-initialise loggers from cfg (level + format)
  5. return cfg
```

Why a helper rather than reordering each command: the order is subtle (logger / DB / effective / logger again), and ten commands each re-implementing it is how the bug appeared. A single function is also the one place to unit test.

Alternative considered: make `settings.Effective()` connect the database itself. Rejected: `internal/settings` would own a connection lifecycle (close, retry, tests with `SetDB`) and hide a side effect behind a getter.

### D2. `requireDB` per command

- `requireDB = true`: `download`, `process`, `resume_downloads`, `enrich_tvdb`, `db_prune`, `reset`, `migrate`. They already need the database and already exit when it fails, so behavior is unchanged except for ordering.
- `requireDB = false` (best effort): `cleanup`, `config`, `m3u download`. They must keep working when the database is unreachable (for example `cleanup` on a node, `config` for diagnostics).

`m3u download` and the other `m3u` subcommands are to be audited during implementation: if a subcommand already reaches the database through a helper it joins the first group.

`migrate` needs the `settings_overrides` table; on a fresh database it may not exist yet. `Effective()` already logs and skips a failed read, so the table being absent must not abort `migrate` (verify in tests).

### D3. Explicit warning in `Effective()` when `database.Get()` is nil

`Effective()` logs one `Warn` ("settings overrides not applied: database not initialised") when called without a connection, and keeps returning the file/env/default configuration. Rationale: it is the safety net for any future caller, and it turns this bug class into a visible log line. It does not change the signature, so the many existing callers and tests keep compiling.

Alternative considered: return `(cfg, error)`. Rejected for now: widespread signature change for a diagnostic that a warning already provides.

### D4. Logger re-initialisation

The bootstrap logger uses `config.Get()` levels and format. After `Effective()`, loggers are rebuilt with the effective levels. Messages emitted by `database.Initialize()` therefore use bootstrap levels, which is acceptable.

## Risks / Trade-offs

- [Behavior change for existing deployments: jobs start honoring overrides that were silently ignored, e.g. a different URL or a disabled integration] -> Document in the changelog; this is the intended behavior per the existing `app-settings` spec.
- [`config` command now opens a database connection] -> Best effort only; on failure it warns and prints file/env values, as before.
- [Bootstrap log level differs from the overridden one for a few startup lines] -> Accepted (D4).
- [Tests that call `Effective()` without a database now emit a warning] -> Warning only; assert on it rather than suppress it.
- [Missing `settings_overrides` table on first `migrate`] -> `Effective()` already tolerates a failed read; add a test.

## Migration Plan

No data migration. Ship the code; the next scheduled run of each CronJob uses the stored overrides. Rollback is reverting the commit. Operators who relied on file/env values being used despite stored overrides must clear those overrides from the UI.
