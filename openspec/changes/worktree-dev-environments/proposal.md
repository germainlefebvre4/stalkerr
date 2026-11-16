# Proposal

## Why

Developing and validating several features in parallel with `git worktree` is blocked by shared, hard-coded resources: `make dev` runs in the foreground on fixed ports (8080/5173), and the only database is the single Postgres from `docker-compose.yml` (fixed `container_name`, fixed host port 5433), so two worktrees would share and corrupt each other's schema and data. A developer cannot bring up an isolated, fully working Stalkeer instance per worktree with one command.

## What Changes

- Add `make wt-up`: starts a self-contained instance (Postgres + Go API + Vite dev server) for the current worktree, in the background, and prints its URLs.
  - API and frontend ports are **dynamic by default** (free ports, stable for a given worktree), and can each be pinned independently: `FRONTEND_PORT` / `API_PORT` (with `FRONT_PORT` accepted as an alias for `FRONTEND_PORT`).
  - Each worktree gets its **own Postgres container** (dedicated compose project, dynamic host port bound to loopback, dedicated volume).
  - The backend is configured purely through environment variables (`DB_*`, `API_PORT`); no `config.yml` is generated or copied, so no secret is duplicated across worktrees.
  - The database is seeded with `scripts/seed-test-data.sql` on every `wt-up` (opt out with `WT_SEED=0`), so the dashboard is not empty.
  - Frontend dependencies are installed on first use if missing.
- Add `make wt-down`: stops the processes and **destroys** the worktree's Postgres container and volume. It works even if the local state file is missing.
- Add `make wt-status` (ports, URLs, health) and `make wt-logs`.
- Per-worktree runtime state (ports, PIDs, logs) lives in a git-ignored `.wt/` directory.
- Document the workflow in the README / `docs/DEVELOPMENT.md`.

No change to Go code or `frontend/vite.config.ts`: both already honor `API_PORT` / `FRONTEND_PORT`.

Out of scope: a `wt-prune` target to clean up orphaned containers of deleted worktrees, keeping data across `wt-down`, cloning an existing database, and enriching the seed data for newer features.

## Capabilities

### New Capabilities
<!-- None: the behavior extends the existing local-dev tooling capability. -->

### Modified Capabilities
- `local-dev`: adds requirements for isolated per-worktree instances (`wt-up` / `wt-down` / `wt-status` / `wt-logs`), dynamic-or-pinned port resolution with the `FRONT_PORT` alias, per-worktree Postgres isolation, per-worktree seeding, and git-ignored runtime state.

## Impact

- **Makefile**: new `wt-*` targets; the existing `API_PORT ?= 8080` / `export` defaults must not leak into `wt-*` resolution (dynamic vs. pinned must be distinguishable).
- **New files**: `scripts/wt.sh` (orchestration), `docker-compose.wt.yml` (Postgres only, no fixed `container_name`/host port).
- **`.gitignore`**: ignore `.wt/`.
- **Docs**: README local-development section, `docs/DEVELOPMENT.md`.
- **Requirements on the host**: Docker with the `docker compose` v2 plugin (the legacy `docker-compose` binary is not assumed), `go`, `node`/`npm`, `psql` (seed) — or `psql` executed inside the container.
- **Safety**: `seed-test-data.sql` begins with `DELETE FROM` on core tables; the target database must only ever be derived from the worktree's own state, never from `config.yml` or the caller's shell environment.
