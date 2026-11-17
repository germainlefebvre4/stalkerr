# Tasks

## 1. Scaffolding and per-worktree Postgres

- [x] 1.1 Add `.wt/` to `.gitignore`; verify `git check-ignore -v .wt/state.env` reports the rule and `git status` stays clean after creating `.wt/x`
- [x] 1.2 Create `docker-compose.wt.yml` with a single `postgres` service (same image as `docker-compose.yml`, no `container_name`, named volume, `pg_isready` healthcheck, `127.0.0.1::5432` port, worktree-path label); verify `docker compose -f docker-compose.wt.yml -p stalkeer-wt-test up -d --wait` starts it, `docker compose -p stalkeer-wt-test port postgres 5432` returns a loopback port, and `down -v` leaves no container or volume (`docker ps -a`, `docker volume ls`)
- [x] 1.3 Verify the standard `docker-compose.yml` Postgres (5433) and a `wt` Postgres can run side by side without error

## 2. `scripts/wt.sh`: identity, ports, state

- [x] 2.1 Implement worktree identity (slug = sanitized directory name + short hash of `git rev-parse --show-toplevel`, compose project `stalkeer-wt-<slug>`) and state-file read/write in `.wt/state.env`; verify two different worktree paths yield different slugs and the same path always yields the same slug
- [x] 2.2 Implement deterministic port selection (hash of the path into 20000-29999, probe upward until free and distinct from already-chosen ports); verify with a smoke test that occupies the preferred port and checks a different free port is returned, and that the same worktree gets the same ports when free
- [x] 2.3 Implement pin handling: `WT_API_PIN` / `WT_FRONT_PIN`, no substitution of a busy pinned port (explicit error naming the port), no side effects on failure; verify busy-pinned-port error and both-spellings conflict error from the shell without starting anything

## 3. Makefile targets

- [x] 3.1 Add `wt-up`, `wt-down`, `wt-status`, `wt-logs` targets (documented in the `## ` help comments) that call `scripts/wt.sh`, computing pins from `$(origin API_PORT)`, `$(origin FRONTEND_PORT)` and `$(origin FRONT_PORT)` so the `?= 8080`/`?= 5173` defaults are not treated as pins; verify with `make -n wt-up`, `make -n wt-up FRONT_PORT=3000`, `make -n wt-up API_PORT=8080`, `make -n wt-up FRONT_PORT=3000 API_PORT=8080` that the right pin variables are passed in each of the four forms (and none for the plain form)
- [x] 3.2 Verify `make dev`, `make dev-backend` and `make dev-frontend` behavior is unchanged (`make -n dev` output identical to before the change)

## 4. Start sequence

- [x] 4.1 Implement `wt.sh up`: start the Postgres project, wait for healthy, read its Docker-assigned port; verify the port and healthy state are written to `.wt/state.env`
- [x] 4.2 Implement the scrubbed backend environment (unset `DATABASE_URL`, `STALKEER_DATABASE_*`, `STALKEER_API_PORT`, `STALKEER_METRICS_PORT`, `PG*`; set `DB_*` and `API_PORT`; free `STALKEER_METRICS_PORT` only when `STALKEER_METRICS_ENABLED=true`), build the binary, start it under `setsid` with `WT_INSTANCE=<slug>`, logging to `.wt/api.log`; verify GET `/health` on the chosen API port succeeds in a worktree that has no `config.yml`, even when the caller exports `DB_HOST=elsewhere` and `DATABASE_URL`
- [x] 4.3 Install frontend dependencies when `frontend/node_modules` is absent (`npm ci`), skip when present, then start Vite under `setsid` with `API_PORT` and `FRONTEND_PORT` set, logging to `.wt/frontend.log`; verify the frontend URL answers 200 and that a request to `<front>/api/v1/...` reaches the API through the proxy
- [x] 4.4 Implement the seed step: after the API is healthy, pipe `scripts/seed-test-data.sql` into `psql -v ON_ERROR_STOP=1` inside the instance container; honor `WT_SEED=0`; verify the seeded rows are present (`SELECT count(*) FROM download_info` > 0) by default and absent with `WT_SEED=0`, and that a database pointed to by exported `DB_*`/`config.yml` is left untouched
- [x] 4.5 Validate `scripts/seed-test-data.sql` on a fresh AutoMigrate'd schema (extension, `ON CONFLICT` targets, columns); fix it if it fails; verify it runs to completion with `ON_ERROR_STOP=1` and that the dashboard lists the seeded downloads
- [x] 4.6 Implement readiness timeouts, failure cleanup and `trap` handling: on any component failure or interruption, stop everything started, print the failing component and its log path, exit non-zero; verify by forcing an API crash (e.g. invalid pinned API port already in use / broken binary) that no process or container remains and the exit code is non-zero
- [x] 4.7 Implement idempotency: second `wt-up` with an unchanged or absent pin prints URLs and exits 0 without a second instance; with a differing pin it fails telling to run `wt-down`, leaving the running instance untouched; verify both cases and that exactly one API process exists afterwards

## 5. Teardown, status, logs

- [x] 5.1 Implement `wt.sh down`: SIGTERM then SIGKILL on both process groups, `docker compose -p <project> down -v --remove-orphans`, removal of `.wt/` state; verify no process, container, volume or listening port remains and a following `wt-up` yields a fresh (re-seeded) database
- [x] 5.2 Implement teardown without state (compose project from the path, `/proc/*/environ` `WT_INSTANCE` scan for leftover processes) and the "nothing running" success path; verify by deleting `.wt/` after `wt-up` then running `wt-down`, and by running `wt-down` on a clean worktree (exit 0, message)
- [x] 5.3 Implement `wt.sh status` (URLs, DB host port, per-component running/healthy; "no instance" message) and `wt.sh logs`; verify output in running and stopped states

## 6. Parallel validation and documentation

- [x] 6.1 Add `scripts/wt-smoke.sh` automating the key spec scenarios (four port forms, busy pinned port, idempotent `wt-up`, `wt-down` completeness, seed on/off, no `config.yml` needed, `git status` clean) and document how to run it; verify it passes locally
- [x] 6.2 Integration check: bring up instances in two real worktrees simultaneously (different ports and databases, data created in one invisible in the other), tear one down and confirm the other is unaffected, then tear it down; verify via `make wt-status` in both and `docker ps`
- [x] 6.3 Document the `wt-*` workflow in the local-development section of `README.md` and in `docs/DEVELOPMENT.md` (four `wt-up` forms, `wt-down` destroys data, `WT_SEED`, prerequisites, orphan cleanup when a worktree is deleted without `wt-down`); verify every documented command runs as written
- [x] 6.4 Run `openspec validate worktree-dev-environments --strict` and verify it passes

## Workflow follow-up

- Archive the change after the project's review requirements are satisfied.
- Verify the archived result (the `local-dev` main spec gains the new requirements).
