## Why

The local API port (`8080`) is duplicated across `Makefile` echo messages, the CLI `--port` flag default in `cmd/server.go`, and `frontend/vite.config.ts`'s dev proxy target, with no single override point. Worse, `config.yml`'s `api.port` (and its `API_PORT` / `STALKEER_API_PORT` env bindings) is read by the config loader but silently ignored by `cmd/server.go`, which always uses its own hardcoded CLI flag default. Today, changing the dev port requires hand-editing `vite.config.ts` directly (visible right now as an uncommitted, undocumented change from `8080` to `8099`). Developers need one place — the Makefile — to control the port for `make dev`.

## What Changes

- Add an `API_PORT` Makefile variable (default `8080`), exported to the `dev`, `dev-backend`, and `dev-frontend` targets; their `echo` messages and the launched binary/dev server both follow it.
- Fix `cmd/server.go` so the server listens on `cfg.API.Port` (sourced from `config.yml`, `API_PORT`, or `STALKEER_API_PORT`) whenever `--port` is not explicitly passed on the CLI, instead of always defaulting to a hardcoded `8080`. **BREAKING**: any deployment that was unknowingly relying on this value being ignored will now see the server honor it.
- `frontend/vite.config.ts`: the dev-server proxy target for `/api/v1` and `/health` reads `process.env.API_PORT` (fallback `8080`) instead of a hardcoded host:port.
- `docker-compose.yml`: align the `server` service's port mapping to `"${API_PORT:-8080}:${API_PORT:-8080}"` (both sides driven by the same variable), so that now that the server actually honors `API_PORT`, the host mapping and the container's internal listen port stay in sync instead of silently diverging.

## Capabilities

### New Capabilities
_None._

### Modified Capabilities
- `local-dev`: the `make dev` workflow (and `dev-backend` / `dev-frontend`) gains a single configurable `API_PORT` Makefile variable that consistently drives the backend's listen port and the frontend dev proxy target.

## Impact

- `Makefile` — new `API_PORT` variable, updated `dev` / `dev-backend` / `dev-frontend` targets.
- `cmd/server.go` — port resolution now falls back to `cfg.API.Port` instead of a hardcoded default.
- `frontend/vite.config.ts` — proxy target reads from environment instead of a literal.
- `docker-compose.yml` — port mapping expression updated for consistency.
- No API surface change, no database migration. Docker/Kubernetes users who never set `API_PORT` see no behavior change (default remains `8080`).
