## Context

See `proposal.md` - Why. Four independent pieces of code currently hardcode or ignore the API port: `Makefile` (display only), `cmd/server.go` (real source of truth today, via a Cobra flag default), `frontend/vite.config.ts` (dev proxy target), and `docker-compose.yml` (already parameterized on the host side only). `config.yml`'s `api.port` and its `API_PORT` / `STALKEER_API_PORT` env bindings already exist in `internal/config/config.go` (`bindEnvWithAlternatives("api.port", "API_PORT")`, default `8080` via `viper.SetDefault`) but are never consulted by `cmd/server.go`.

## Goals / Non-Goals

**Goals:**
- One override point (`API_PORT` Makefile variable) drives both the backend port and the frontend dev proxy target for the `make dev` family of targets.
- Make the existing (currently dead) `config.yml` / env-based port configuration actually take effect, without breaking any current default behavior.
- Keep `docker-compose.yml`'s host↔container port mapping consistent once the server starts honoring `API_PORT`.

**Non-Goals:**
- No changes to the Helm chart (`charts/stalkerr/`) - it hardcodes container port `8080` in a separate deployment path not exercised by `make dev`, and was explicitly left out of scope.
- No change to the production/admin port (`8081`) or any other service's port.
- No new config file format or new env var name - reuse the existing `API_PORT` / `STALKEER_API_PORT` names already bound in `internal/config/config.go`.

## Decisions

**Port precedence in `cmd/server.go`: explicit `--port` flag > `cfg.API.Port` > hardcoded `8080`.**
Today, `port, _ := cmd.Flags().GetInt("port")` is read *before* `config.Load()` even runs, so the flag's static default (`8080`) always wins regardless of config/env. The fix reorders this: call `config.Load()` first, then check `cmd.Flags().Changed("port")`. If the user passed `--port` explicitly, use it (highest precedence - an explicit CLI arg should never be silently overridden by a config file). Otherwise use `cfg.API.Port`, which viper has already resolved from `config.yml` / `API_PORT` / `STALKEER_API_PORT` / the `8080` default. This makes the Cobra flag's own static default value irrelevant in practice (kept only as the value `GetInt` would return if `Changed` were somehow unreliable), since `cfg.API.Port` already carries the same `8080` fallback via `viper.SetDefault`.

**Makefile propagates via the existing `API_PORT` env var name, not a new one.**
Because `config.go` already binds `API_PORT` (alongside `STALKEER_API_PORT`), the Makefile only needs to `export API_PORT` (default `?= 8080`) before invoking `./bin/stalkeer server` and `npm run dev` - no new plumbing, no explicit `--port` flag needs to be passed by the Makefile at all. This also means anyone already using `API_PORT=9090 make dev` today (if they guessed at it) gets the behavior they'd expect, for free.

**`vite.config.ts` reads `process.env.API_PORT` directly (no `VITE_` prefix).**
`vite.config.ts` executes in Node/Vite's config-loading context, not in client bundle code - the `VITE_` prefix requirement only applies to variables exposed to `import.meta.env` in client code. Reading `process.env.API_PORT` directly keeps the same variable name used everywhere else (Makefile, docker-compose, config.go), avoiding a second name for the same concept.

**`docker-compose.yml` mapping changes from `"${API_PORT:-8080}:8080"` to `"${API_PORT:-8080}:${API_PORT:-8080}"`.**
Alternative considered: leave docker-compose untouched and accept that setting `API_PORT` for compose would now break the mapping (since the container would listen on the overridden port while the mapping's right-hand side stays `8080`). Rejected per user decision - fixing the backend precedence bug must not introduce a new, more subtle bug in the docker path. Symmetric interpolation keeps default behavior (`8080:8080`) identical to today's `8080:8080` and keeps any override consistent on both sides.

## Risks / Trade-offs

- [Someone was relying on `config.yml`'s `api.port` / `API_PORT` being ignored, expecting the CLI flag default to always win] → Extremely unlikely (the current behavior is an unannounced bug, not a documented feature); flagged as **BREAKING** in the proposal for visibility.
- [A user's shell already has a stray `API_PORT` env var set for an unrelated reason, now silently changing the server's port] → Same class of risk as any `AutomaticEnv`-based config today (e.g. `DB_HOST`); consistent with existing project convention, not a new risk class.
