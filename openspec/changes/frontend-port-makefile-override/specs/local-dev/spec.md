## MODIFIED Requirements

### Requirement: Unified Hybrid Development Command
The system SHALL provide a unified make command to compile and launch the Go backend server alongside the React frontend SPA in parallel, with the API port configurable via a single `API_PORT` variable and the frontend port configurable via a single `FRONTEND_PORT` variable.

#### Scenario: Running make dev
- **WHEN** a developer runs `make dev`
- **THEN** the system SHALL compile the Go binary, start the API server, start the Vite development server, and gracefully terminate both servers upon receiving a SIGINT signal.

#### Scenario: Overriding the API port
- **WHEN** a developer runs `make dev API_PORT=9090` (or sets `API_PORT` before invoking `make dev-backend` / `make dev-frontend`)
- **THEN** the Go API server SHALL listen on port 9090
- **THEN** the Vite development server's proxy for `/api/v1` and `/health` SHALL target `http://localhost:9090`
- **THEN** the printed startup messages SHALL reflect port 9090 instead of the default

#### Scenario: Default port unchanged
- **WHEN** a developer runs `make dev` without setting `API_PORT`
- **THEN** the API server and Vite proxy SHALL both use port 8080, matching current behavior

#### Scenario: Overriding the frontend port
- **WHEN** a developer runs `make dev FRONTEND_PORT=3000` (or sets `FRONTEND_PORT` before invoking `make dev-frontend` or `npm run dev`)
- **THEN** the Vite development server SHALL listen on port 3000
- **THEN** the printed startup messages SHALL show `http://localhost:3000` for the frontend dashboard

#### Scenario: Default frontend port unchanged
- **WHEN** a developer runs `make dev` without setting `FRONTEND_PORT`
- **THEN** the Vite development server SHALL listen on port 5173, matching current behavior

#### Scenario: Frontend port already in use
- **WHEN** the configured frontend port is already occupied by another process
- **THEN** the Vite development server SHALL fail with an error instead of falling back to another port
