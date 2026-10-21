## 1. Makefile

- [ ] 1.1 Add `FRONTEND_PORT ?= 5173` and `export FRONTEND_PORT` next to `API_PORT` in `Makefile`
- [ ] 1.2 Replace the hardcoded `5173` in the `dev` and `dev-frontend` echo lines with `$(FRONTEND_PORT)`

## 2. Vite config

- [ ] 2.1 In `frontend/vite.config.ts`, add `const frontendPort = Number(process.env.FRONTEND_PORT) || 5173;` and use it for `server.port`
- [ ] 2.2 Add `strictPort: true` to `server`

## 3. Documentation

- [ ] 3.1 Add `FRONTEND_PORT=5173` after `API_PORT=8080` in `.env.example`
- [ ] 3.2 Update README "Run Both Servers Concurrently" to mention the `FRONTEND_PORT` override (e.g. `make dev FRONTEND_PORT=3000`) and the strict-port failure

## 4. Verification

- [ ] 4.1 Run `make dev FRONTEND_PORT=3000` and confirm Vite listens on 3000 and the echo shows 3000
- [ ] 4.2 Occupy the port and confirm Vite fails instead of incrementing
- [ ] 4.3 Run `make dev` with no override and confirm 5173 is used
