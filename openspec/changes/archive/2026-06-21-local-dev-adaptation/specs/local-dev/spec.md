## ADDED Requirements

### Requirement: Frontend Make targets
The system SHALL provide Makefile targets to install, run, build, and lint the frontend application.

#### Scenario: Running frontend commands
- **WHEN** a developer runs `make front-install`, `make front-dev`, `make front-build`, or `make front-lint`
- **THEN** the system SHALL execute the corresponding npm action inside the `/frontend` directory.

### Requirement: Unified Hybrid Development Command
The system SHALL provide a unified make command to compile and launch the Go backend server alongside the React frontend SPA in parallel.

#### Scenario: Running make dev
- **WHEN** a developer runs `make dev`
- **THEN** the system SHALL compile the Go binary, start the API server, start the Vite development server, and gracefully terminate both servers upon receiving a SIGINT signal.

### Requirement: Dashboard and Dev Setup Documentation
The system's main `README.md` SHALL describe the Web Dashboard's features and list step-by-step local development setup instructions.

#### Scenario: Read development instructions
- **WHEN** a user reads the root `README.md`
- **THEN** they SHALL find the detailed prerequisites, install commands, ports specification, and instructions to launch both GIN server and Vite frontend.
