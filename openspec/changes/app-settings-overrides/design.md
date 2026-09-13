## Context

See `proposal.md` for motivation. Relevant current-state constraints:

- `internal/config` is a viper-based singleton (`config.Get()`), loaded once via `config.Load()`. It has no knowledge of the database.
- `internal/database` already imports `internal/config` (for the DSN), so `internal/config` cannot import `internal/database` without a cycle.
- Radarr/Sonarr handlers and most CLI commands (`cmd/process.go`, `cmd/download.go`, `cmd/m3u.go`, ...) already call `config.Get()` fresh on every invocation. The TMDB client, the `Downloader`'s timeout/retry/min-file-size, and the metrics listener are instead captured once in `api.NewServer()` / `cmd/server.go` at process start.
- The existing `filter_configs` table (GORM, `AutoMigrate`d in `internal/database/database.go`) is the precedent this change generalizes: a DB row keyed by a small fixed key (`attribute`), read alongside the file-origin value, with the DB row winning when present.

## Goals / Non-Goals

**Goals:**
- One consistent place ("effective config") that merges file/env config with stored overrides, usable identically by the API server and every CLI command.
- Per-field override and origin for Tier-1 (applicative) config; read-only, origin-labeled exposure of Tier-0 (bootstrap) config.
- A dedicated, list-shaped mechanism for M3U sources, keyed by name, mirroring the filter override pattern.
- The app boots with zero Tier-1 configuration anywhere.

**Non-Goals:**
- Changing the Helm chart's own config delivery/validation (`configuration-management`, `values.schema.json`). It keeps requiring explicit values; a chart that also tolerates zero config is a separate later change.
- Distinguishing "file" from "env" as separate origins (decided: two buckets, "Interface" vs "Config").
- Field-level override granularity for M3U sources (decided: whole-block per `name`, like filters).
- Making every Tier-1 field apply with zero restart. A small, explicitly-flagged subset stays restart-required (see Decisions).
- Encryption at rest for secrets stored as overrides — left to the existing Postgres/infra posture.

## Decisions

### A new `internal/settings` package hosts the merge
`internal/settings` imports both `internal/config` and `internal/database` (no cycle, since neither imports it back). It exposes:
- `Effective() *config.Config` — same shape as `config.Get()`, with every Tier-1 field's stored override layered on top, and its M3U source list replaced by the resolved effective list (file + `m3u_source_configs`).
- CRUD + origin functions for Tier-1 field overrides and for M3U source rows.

Every call site that currently reads `config.Get()` for a Tier-1 concern (Radarr/Sonarr/TMDB/Jellyfin/Notifications/Downloads/logging handlers, and the equivalent CLI commands) switches to `settings.Effective()`, so API-triggered and cron/CLI-triggered runs stay behaviorally identical. Tier-0 fields (`database.*`, `api.port`, `metrics.*`) keep reading `config.Get()` directly, since they are never overridden.

**Alternative considered**: teach `internal/config` itself to read the DB (rejected — the import cycle), or push the merge into `internal/api` only (rejected — CLI commands would silently ignore UI overrides, diverging from what the API server does for the same job).

### Two storage shapes: generic key/value for scalars, a dedicated table for M3U sources
- `settings_overrides` (new): `key TEXT PRIMARY KEY` (dotted path, e.g. `"radarr.api_key"`), `value TEXT` (JSON-encoded scalar), `updated_at`. One row per overridden field — matches the "per field" granularity decided during exploration, and needs no schema change when a new overridable field is added later.
- `m3u_source_configs` (new): mirrors `filter_configs` — `name` (unique), the full source shape (`file_path` + every `download.*` field), `is_runtime`, timestamps. A row for a given `name` replaces that source's entire block; a `name` with no row falls back to the file-defined entry (see `m3u-source-overrides` spec).

**Alternative considered**: one DB column per Tier-1 field, or one table per domain (Radarr, Sonarr, ...) — rejected as needless schema churn for ~25 scalar fields that all follow the same read/write/mask/origin logic.

### An explicit, hand-written field registry — not reflection
`internal/settings` holds a static registry mapping each overridable dotted key to: its Go type, whether it is sensitive (API keys, tokens, the DB password is Tier-0 so excluded), and whether it is restart-required (see below). Reads and writes go through this registry rather than generic struct reflection, so an invalid or unknown key is rejected at the API boundary instead of silently no-op'ing, and adding a new overridable field is a one-line, compile-checked registry entry next to a `config.go` field.

### Restart-required subset
Only the fields actually baked into a long-lived object at process start are flagged `restart_required` in the registry: `tmdb.api_key`, `tmdb.language`, `tmdb.requests_per_second` (the TMDB client), and `downloads.timeout`, `downloads.retry_attempts`, `downloads.min_file_size_mb` (the `Downloader` instance built in `api.NewServer()`). Every other Tier-1 field is already read fresh per call (Radarr, Sonarr, Jellyfin, Notifications, most `downloads.*` path fields) and applies immediately through `settings.Effective()`.

`logging.app.level`, `logging.database.level`, and `logging.format` are a special case: also process-global (`logger.InitializeLoggersWithFormat`), but cheap to reapply. Writing an override for any of them re-invokes that initializer immediately, so logging is live despite being global state — not flagged `restart_required`.

### Bootstrap fields are read-only, not merely unenforced
The settings API rejects a write attempt for any Tier-0 key (`database.*`, `api.port`, `metrics.port`, `metrics.enabled`) rather than silently ignoring it, so a misdirected client request fails loudly instead of appearing to succeed.

### Sensitive field masking
The settings API never returns a sensitive field's raw value (override or config-resolved) — only whether one is set. The frontend's input for such a field starts empty/masked and only submits an override when the user types a new value (see `frontend-app-settings-management`).

## Risks / Trade-offs

- **Secrets move into the application database.** Radarr/Sonarr/TMDB API keys and the ntfy token, once overridden, live in Postgres (and its backups) rather than only in the K8s Secret/env. → Mitigated by masking at the API layer; DB-level encryption at rest is explicitly out of scope for this change.
- **M3U source rename orphaning.** Renaming a source's `name` in `config.yml` after it has a stored override orphans that override (it becomes a runtime-only source under the old name). → Accepted, matches the existing filter override's implicit assumption that keys are stable; documented in the spec rather than solved with rename-tracking.
- **Registry/struct drift.** A new `config.Config` field added later without a matching registry entry silently stays file/env-only (not overridable), which may surprise a contributor. → Mitigated with a unit test that reflects over `config.Config` and asserts every scalar leaf field is either in the registry or on an explicit exclusion list (Tier-0 fields, list-shaped fields).
- **"Config" bucket hides file-vs-env.** An operator cannot tell from the UI whether a "Config"-origin value came from `config.yml` or an env var. → Accepted per the two-bucket decision; revisiting this later does not require reworking the override mechanism, only the origin label.
- **Restart-required fields could confuse users if missed.** → Mitigated by the frontend spec's explicit "restart required" indicator, driven by the same registry flag the backend uses, so frontend and backend cannot disagree on which fields need it.

## Migration Plan

1. Add `settings_overrides` and `m3u_source_configs` to the existing `AutoMigrate` call in `internal/database/database.go`. Additive only — no existing table or column changes.
2. Relax `internal/config.validate()` to stop failing on an empty `m3u.sources`; bootstrap validation (`database.user`, `database.dbname`) is unchanged.
3. Add `internal/settings` (registry, `Effective()`, override CRUD, M3U source CRUD).
4. Switch Tier-1 call sites (API handlers, CLI commands) from `config.Get()` to `settings.Effective()`.
5. Add API endpoints: settings origin/effective/CRUD, M3U sources origin/effective/CRUD (mirroring `/api/v1/filters`).
6. Add frontend sections/components/hooks and locale strings.

Rollback is a plain redeploy of the previous version: every change is additive (new tables, new endpoints, relaxed validation), and with zero overrides stored, `settings.Effective()` returns exactly what `config.Get()` already returns today — existing deployments with full file/env config see no behavior change until an operator actually sets an override.

## Open Questions

- Whether `downloads.movies_path` / `downloads.tvshows_path` should be flagged `restart_required` out of caution (an in-flight download could be mid-move when the path changes) even though they are read fresh per operation today. Does not change the mechanism — only a registry flag — so can be settled during implementation.
- Exact REST route names/shapes for the new endpoints beyond mirroring `/api/v1/filters`'s existing conventions.
