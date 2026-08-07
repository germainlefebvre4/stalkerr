## 1. Backend port resolution

- [x] 1.1 In `cmd/server.go`, reorder `serverCmd.Run` so `config.Load()` runs before the port is resolved, then use `cfg.API.Port` unless `cmd.Flags().Changed("port")` is true (in which case use the explicit `--port` value).
- [x] 1.2 Update the log lines and `server.Run(...)` call in `cmd/server.go` to use the resolved port value.

## 2. Makefile API_PORT variable

- [x] 2.1 Add an `API_PORT ?= 8080` variable near the top of the `Makefile` and export it.
- [x] 2.2 Update the `dev` target's `echo` message and invocation to use `$(API_PORT)` instead of the hardcoded `8080`.
- [x] 2.3 Update the `dev-backend` target's `echo` message to use `$(API_PORT)`.
- [x] 2.4 Update the `dev-frontend` target's `echo` message to use `$(API_PORT)`.

## 3. Frontend dev proxy target

- [x] 3.1 Update `frontend/vite.config.ts` so the `/api/v1` and `/health` proxy targets read `` `http://localhost:${process.env.API_PORT || 8080}` `` instead of the hardcoded `http://localhost:8080`.

## 4. Docker Compose port mapping

- [x] 4.1 In `docker-compose.yml`, change the `server` service's port mapping from `"${API_PORT:-8080}:8080"` to `"${API_PORT:-8080}:${API_PORT:-8080}"`.

## 5. Verification

- [x] 5.1 Run `go build ./...` to confirm `cmd/server.go` compiles.
- [x] 5.2 Run `make dev-backend` with default and overridden (`API_PORT=9090 make dev-backend`) settings to confirm the server listens on the expected port (via startup log message).
- [x] 5.3 Run `cd frontend && npm run build` (or lint) to confirm `vite.config.ts` remains valid TypeScript.
