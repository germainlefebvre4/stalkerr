# local-dev Specification

## Purpose
TBD - created by archiving change local-dev-adaptation. Update Purpose after archive.
## Requirements
### Requirement: Frontend Make targets
The system SHALL provide Makefile targets to install, run, build, and lint the frontend application.

#### Scenario: Running frontend commands
- **WHEN** a developer runs `make front-install`, `make front-dev`, `make front-build`, or `make front-lint`
- **THEN** the system SHALL execute the corresponding npm action inside the `/frontend` directory.

### Requirement: Unified Hybrid Development Command
The system SHALL provide a unified make command to compile and launch the Go backend server alongside the React frontend SPA in parallel, with the API port configurable via a single `API_PORT` variable.

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

### Requirement: Backend Port Resolution Honors Configuration
The API server SHALL resolve its listen port from, in order of precedence: an explicitly-passed `--port`/`-p` CLI flag, then `api.port` as configured via `config.yml` or the `API_PORT` / `STALKEER_API_PORT` environment variables, then a default of `8080`.

#### Scenario: No explicit CLI flag, port set via environment
- **WHEN** the server is started without `--port` and the `API_PORT` (or `STALKEER_API_PORT`) environment variable is set to `9090`
- **THEN** the server SHALL listen on port 9090

#### Scenario: No explicit CLI flag, port set via config file
- **WHEN** the server is started without `--port` and `config.yml` sets `api.port: 9090`
- **THEN** the server SHALL listen on port 9090

#### Scenario: Explicit CLI flag takes precedence
- **WHEN** the server is started with `--port 9091` while `API_PORT=9090` is also set
- **THEN** the server SHALL listen on port 9091

### Requirement: Docker Compose Port Mapping Stays Consistent
`docker-compose.yml` SHALL derive both the host-side and container-side API port from the same `API_PORT` variable, so overriding it does not desynchronize the port mapping from the port the server actually listens on.

#### Scenario: Overriding API_PORT for docker-compose
- **WHEN** a developer runs `API_PORT=9090 docker-compose up server`
- **THEN** the host SHALL expose port 9090
- **THEN** the container's internal listen port SHALL also be 9090, matching the mapping

#### Scenario: Default docker-compose port unchanged
- **WHEN** `API_PORT` is not set
- **THEN** the mapping SHALL default to `8080:8080`, matching current behavior

### Requirement: Dashboard and Dev Setup Documentation
The system's main `README.md` SHALL describe the Web Dashboard's features and list step-by-step local development setup instructions.

#### Scenario: Read development instructions
- **WHEN** a user reads the root `README.md`
- **THEN** they SHALL find the detailed prerequisites, install commands, ports specification, and instructions to launch both GIN server and Vite frontend.

