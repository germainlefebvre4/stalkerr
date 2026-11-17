# Design

## Context

See `proposal.md` for motivation. Observed constraints in the current code:

- `API_PORT` and `FRONTEND_PORT` are already honored end to end: Makefile (`?=` + `export`), Go (`bindEnvWithAlternatives("api.port", "API_PORT")`; `--port` only wins if explicitly passed), `frontend/vite.config.ts` (listen port with `strictPort: true`, `/api/v1` and `/health` proxied to `localhost:$API_PORT`). No Go or Vite change is needed.
- The backend supports PostgreSQL only and runs `AutoMigrate` at startup. It needs just `database.user` and `database.dbname`; `m3u.sources` may be empty. It therefore boots with **no `config.yml`**, from `DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME/DB_SSLMODE` env vars, which take precedence over a config file.
- `docker-compose.yml` uses fixed `container_name`s and a fixed host port (5433), so it cannot be started twice; its `docker-compose` v1 invocation in the Makefile is not available on this machine (only `docker compose` v2).
- Because of `API_PORT ?= 8080` + `export API_PORT` in the Makefile, `API_PORT` is always defined inside recipes: a naive `wt-up` could not tell "pinned to 8080" from "not provided".
- `scripts/seed-test-data.sql` begins with `DELETE FROM download_info / processed_lines / movies / tvshows`, requires `pgcrypto`, and was written for the downloads-enrichment feature (about 15 downloads). It sets no `source_name` (column default applies).
- Data paths (`./data/downloads`, `./m3u_playlist`) are relative, hence already per-worktree.

## Goals / Non-Goals

**Goals:**
- One command to bring up, and one to destroy, a fully isolated instance per worktree.
- Pinned or dynamic ports, independently for API and frontend.
- Nothing that can touch another checkout's database or config.

**Non-Goals:**
- Orphan cleanup for deleted worktrees (`wt-prune`), keeping data across `wt-down`, cloning an existing database, enriching the seed for newer features.
- Containerizing the Go API or the Vite server (they run as host processes).
- Touching `docker-compose.yml`, Helm charts or production images.
- Native Windows support.

## Decisions

### D1. Thin Makefile targets, logic in `scripts/wt.sh`
`wt-up`, `wt-down`, `wt-status`, `wt-logs` are one-line Makefile targets calling `scripts/wt.sh <command>`. Process management, port probing and readiness polling are awkward and unreadable in Make recipes.
*Alternative*: pure Makefile logic. Rejected: shell state does not survive between recipe lines, and error handling/cleanup (trap) is poor.

### D2. Telling "pinned" from "default" in Make
The Makefile computes pins from `$(origin ...)`: a port is pinned only if its origin is `command line`, `environment` or `override`; the `file` origin (the `?= 8080` default) means "not provided". Pins are passed to the script as `WT_API_PIN` / `WT_FRONT_PIN`; the script ignores the ambient `API_PORT`/`FRONTEND_PORT` and exports the resolved values itself to child processes. `FRONT_PORT` is read the same way and mapped onto the frontend pin; both spellings set to different values is an error.
*Consequence*: an `API_PORT` exported in the caller's shell counts as a pin (explicit intent). That matches the existing `make dev` behavior.
*Alternative*: rename variables to `WT_*`. Rejected: the requested UX is `make wt-up FRONT_PORT=3000 API_PORT=8080`.

### D3. Port selection: deterministic hash + probe
API and frontend ports are derived from a hash of the worktree top-level path (`git rev-parse --show-toplevel`) into the range 20000-29999 (below the Linux ephemeral range, away from the common dev ports), then probed upward until a port is free and not already chosen for this instance. Same worktree gives the same ports as long as they are free, so bookmarks survive `wt-down`/`wt-up`.
A pinned port is never substituted: if it is busy, fail.
*Alternative*: bind to port 0 and read the port back. Rejected for API/Vite: the ports must be known before launch (Vite proxy target depends on the API port) and neither server supports "port 0, then report". A small race between probe and bind remains; Vite's `strictPort` and the API's listen error surface it as a clear startup failure.

### D4. Postgres per worktree via a dedicated compose file
`docker-compose.wt.yml` defines only `postgres` (same image as the main compose), with no `container_name`, a named volume, a healthcheck, and `ports: ["127.0.0.1::5432"]` so Docker itself assigns a free host port (no probe race); the script reads it with `docker compose port`. The compose project is named `stalkeer-wt-<slug>`, where `<slug>` is the sanitized worktree directory name plus a short hash of its full path (unique and valid as a compose project name). Credentials are fixed, local-only development values (user `postgres`, database `stalkeer`) because the port is loopback-bound and the database is throwaway. The container is labeled with the worktree path to help manual orphan cleanup. The service uses `network_mode: bridge` (Docker's default bridge) so no per-project network is created: with many worktree stacks on one machine Docker can run out of address pools, and the database only needs its published loopback port.
`wt-down` runs `docker compose -p <project> down -v --remove-orphans`, derived from the path only, so it works without `.wt/`.
*Alternatives*: a shared Postgres with one database per worktree (lighter, but not autonomous and depends on a Postgres already running); reusing `docker-compose.yml` with overrides (fixed `container_name`s make it impossible).

### D5. Configuration purely by environment, scrubbed
The script launches the API with an explicit environment: `DB_HOST=127.0.0.1`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE=disable`, `API_PORT`, and when `STALKEER_METRICS_ENABLED=true` a free `STALKEER_METRICS_PORT`. It first removes variables that could redirect the database or ports (`DATABASE_URL`, `STALKEER_DATABASE_*`, `STALKEER_API_PORT`, `STALKEER_METRICS_PORT`, `PG*`). A `config.yml` that happens to exist in the worktree is not touched; the explicit env vars take precedence over it. No `config.yml` is generated or copied: nothing sensitive is duplicated.
*Alternative*: generate a `config.yml` from the example. Rejected: not needed for boot, adds a file to keep in sync and a temptation to copy secrets.

### D6. Background processes with process groups
The API (`bin/stalkeer server`, built in the worktree via the existing `build` target) and Vite (`npm run dev` in `frontend/`) are each started with `setsid`, logging to `.wt/api.log` and `.wt/frontend.log`. `.wt/state.env` records slug, ports, pin flags and process-group ids. Teardown sends SIGTERM to each group, then SIGKILL after a timeout, so the `npm -> node` children die too.
Each process is also started with `WT_INSTANCE=<slug>` in its environment: if `.wt/state.env` is lost, `wt-down` finds leftovers by scanning `/proc/*/environ` for that marker (Linux). Readiness: database healthcheck, API `/health`, frontend HTTP 200, each with a timeout; on any failure the script stops everything it started, prints the failing component and its log path, and exits non-zero (a shell `trap` handles interruption too).

### D7. Seeding through the container, after the API is healthy
The seed is piped into `psql` **inside the instance's container** (`docker compose -p <project> exec -T postgres psql ... -v ON_ERROR_STOP=1`). That targets only the worktree's database by construction (ambient `PG*`/`DB_*` variables and `config.yml` are irrelevant), and removes the need for a host `psql`. It runs once the API is healthy, since the API creates the schema at startup. Because `wt-down` always destroys the database, every `wt-up` starts empty and is seeded; `WT_SEED=0` skips it. A seed failure fails `wt-up` (with cleanup) and names `WT_SEED=0` as the escape hatch.
*Alternative*: reuse `make seed-test-data`. Rejected: it reads `config.yml` and hard-codes port 5433, i.e. it would hit the wrong database.

### D8. `docker compose` v2 only
The script uses `docker compose` (plugin). The legacy `docker-compose` binary is not assumed.

### D9. Idempotency and pin conflicts
`wt-up` on a worktree whose instance is healthy prints the URLs and exits 0. If pins were given and differ from the running ports, it fails and asks for `wt-down` first rather than restarting silently (which would destroy data).

## Risks / Trade-offs

- [Deleting a worktree without `wt-down` leaves an orphaned Postgres container and volume] -> label the container with its worktree path, document `docker compose ls` / `docker volume ls` cleanup, and keep `wt-prune` as a possible follow-up.
- [`wt-down` destroys data by design; a long validation session can be lost] -> documented prominently; `wt-up` prints that the data is ephemeral; a `KEEP_DATA` option is a cheap later addition.
- [Seed starts with `DELETE FROM` on core tables] -> database coordinates come only from the instance's own container/state, never from the environment or `config.yml` (D5, D7).
- [Seed may be stale against the current schema or too thin for the feature under validation] -> a task validates it on a fresh migrated database; richer seeds are out of scope.
- [Probe-then-bind race on API/frontend ports] -> failure is explicit (strict port, listen error) and the start is cleaned up; the DB port is race-free (Docker-assigned).
- [`/proc` fallback for teardown without state is Linux-specific] -> primary path uses the state file; on other systems the fallback is limited to the compose project.
- [Resource use: one Postgres + one Go process + one Node process per worktree] -> acceptable for a handful of worktrees; `wt-down` frees everything.
- [Hash collisions between worktrees on ports] -> probing guarantees distinct free ports at start time.

## Migration Plan

Purely additive developer tooling; `make dev` and `docker-compose.yml` are unchanged. Rollback = remove the `wt-*` targets, `scripts/wt.sh`, `docker-compose.wt.yml` and the `.gitignore` entry.

## Open Questions

- Whether to extend the seed with doubles, errors and job runs to exercise newer UI states (separate change).
- Exact timeouts for readiness polling (to be tuned during implementation; do not affect specs).
