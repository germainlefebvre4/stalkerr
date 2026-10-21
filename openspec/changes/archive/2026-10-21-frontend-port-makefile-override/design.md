## Context

`API_PORT` was made configurable in the archived change `2026-08-07-configurable-dev-api-port`: the Makefile declares `API_PORT ?= 8080` and exports it, and `vite.config.ts` reads `process.env.API_PORT`. The frontend port is still hardcoded (`server.port: 5173`; `5173` in the Makefile `echo` lines of `dev` and `dev-frontend`). `docker-compose.yml` already uses `${FRONTEND_PORT:-5173}` for the container mapping, so the variable name exists.

## Goals / Non-Goals

**Goals:**
- Override the Vite dev port with `FRONTEND_PORT`, symmetrically to `API_PORT`.
- Keep the URL printed by make accurate.

**Non-Goals:**
- `vite preview` port, nginx container port, docker-compose changes.
- Having the Makefile read `.env` (also true for `API_PORT`; a separate concern).

## Decisions

- **Reuse the name `FRONTEND_PORT`**: already used by docker-compose; avoids a second name for the same concept.
- **Read `process.env.FRONTEND_PORT` in `vite.config.ts`** rather than passing `--port` from make: symmetric with `apiPort`, and `FRONTEND_PORT=3000 npm run dev` works without make. The value is converted with `Number(...)`, since Vite expects a number.
- **`strictPort: true`**: Vite otherwise increments to the next free port, making the Makefile `echo` lie once the port is configurable. This matches the Go backend, which fails on an occupied port. Alternative considered: strict only when `FRONTEND_PORT` is explicitly set; rejected for added complexity and inconsistent behavior.

## Risks / Trade-offs

- A second `make dev` (or any process on 5173) now fails instead of starting on 5174 → intended; the error is explicit and the override is documented.
