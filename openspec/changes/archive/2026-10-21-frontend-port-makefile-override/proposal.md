## Why

`API_PORT` can already be overridden for local development (`make dev API_PORT=9090`), but the Vite frontend port is hardcoded to `5173` in `frontend/vite.config.ts` and in the Makefile startup messages. Developers running several instances, or with `5173` already taken, cannot change it without editing files.

## What Changes

- Add `FRONTEND_PORT ?= 5173` and `export FRONTEND_PORT` to the Makefile, mirroring `API_PORT`.
- Use `$(FRONTEND_PORT)` in the startup messages of `make dev` and `make dev-frontend`.
- Read `process.env.FRONTEND_PORT` (default `5173`) in `frontend/vite.config.ts` for `server.port`.
- Set `server.strictPort: true` so Vite fails instead of silently moving to the next free port, which keeps the printed URL accurate.
- Document `FRONTEND_PORT=5173` in `.env.example` next to `API_PORT`.
- Update the README "Run Both Servers Concurrently" section to mention the override.
- **BREAKING** (minor): if port `5173` is already in use, `make dev` now fails instead of falling back to `5174`.

Out of scope: `vite preview` (4173), the nginx container port (8080), `docker-compose.yml` (already uses `FRONTEND_PORT`), and making the Makefile read `.env`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `local-dev`: the "Unified Hybrid Development Command" requirement gains a configurable frontend port and strict-port behavior.

## Impact

- `Makefile`, `frontend/vite.config.ts`, `.env.example`, `README.md`
- `openspec/specs/local-dev/spec.md` (via delta)
- No backend, API, or dependency changes.
