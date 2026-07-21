## Context

Stalkeer has evolved from a pure backend/CLI program into a multi-pod, decoupled system featuring a rich React 19 + Radix UI frontend and an extended Gin API server. The existing `Makefile` and `README.md` only refer to building and running the Go binary.

We need to add commands to:
1. Support installing frontend dependencies (`npm install` in `/frontend`).
2. Run Vite dev server in `/frontend`.
3. Support hybrid parallel development starting both GIN server on `:8080` and Vite dev server on `:5173` with full SIGINT cleanup of both processes.
4. Support linting/building the frontend.

Additionally, the `README.md` must be updated to showcase the modern Light theme dashboard features and explain the prerequisite steps and starting commands.

## Goals / Non-Goals

**Goals:**
- Provide developer-friendly make commands for all frontend operations.
- Provide a single unifed `make dev` command starting both frontend and backend and cleaning up gracefully on exit.
- Document dashboard features, ports, and dev workflows in root `README.md`.

**Non-Goals:**
- Modifying production Go compiler flags.
- Changing K8s deploy workflows (handled in separate Helm change).

## Decisions

### 1. Unix Process Trap for Parallel Dev Orchestration
- **Rationale**: When starting GIN server and Vite concurrently inside a single terminal using `make dev`, a developer expect to terminate both processes cleanly when they press `Ctrl+C`. 
- **Implementation**:
  ```makefile
  dev: build
  	(trap 'kill 0' SIGINT; ./bin/$(BINARY_NAME) server & cd frontend && npm run dev)
  ```
  This unifies process groups under the parent terminal and kills all background children when SIGINT is caught.

### 2. Streamlining Root Prerequisite Documentation
- **Rationale**: To maximize dev velocity, prerequisites should separate Go requirements and Node requirements clearly, indicating the default ports:
  - Go GIN API: `http://localhost:8080`
  - Vite dev server (and SPA): `http://localhost:5173`

## Risks / Trade-offs

- **[Risk] Port 8080 / 5173 Collisions** → Ports may be occupied by other local services.
  - *Mitigation*: Ensure Vite config has proxy settings ready, and document how developers can pass custom environment variables if needed.
