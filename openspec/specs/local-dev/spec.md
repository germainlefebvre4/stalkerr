# local-dev Specification

## Purpose
Defines the local development tooling for running the frontend and backend together — Makefile targets, a unified hybrid dev command, configurable API port resolution kept consistent across the Go server, Vite proxy, and docker-compose, and the developer-facing documentation.

## Requirements

### Requirement: Frontend Make targets
The system SHALL provide Makefile targets to install, run, build, and lint the frontend application.

#### Scenario: Running frontend commands
- **WHEN** a developer runs `make front-install`, `make front-dev`, `make front-build`, or `make front-lint`
- **THEN** the system SHALL execute the corresponding npm action inside the `/frontend` directory.

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

### Requirement: Isolated instance per worktree
The system SHALL provide `make wt-up` to start a self-contained Stalkeer instance (database, API server, frontend dev server) for the current git worktree, in the background, so that several worktrees can run instances at the same time without interfering with each other.

#### Scenario: Starting an instance
- **WHEN** a developer runs `make wt-up` in a worktree
- **THEN** a dedicated database, the Go API server and the Vite development server SHALL be started and the command SHALL return once they are reachable
- **THEN** the command SHALL print the frontend URL and the API URL of that instance

#### Scenario: Two worktrees in parallel
- **WHEN** `make wt-up` has been run in two different worktrees of the same repository
- **THEN** both instances SHALL run simultaneously, each with its own ports and its own database, and neither SHALL read or modify the other's data

#### Scenario: Instance already running
- **WHEN** a developer runs `make wt-up` in a worktree whose instance is already running with the same port settings
- **THEN** the system SHALL NOT start a second instance and SHALL print the URLs of the running one

#### Scenario: Instance running with different pinned ports
- **WHEN** a developer runs `make wt-up` with pinned ports that differ from those of the instance already running in that worktree
- **THEN** the command SHALL fail with a message telling the developer to run `make wt-down` first, and the running instance SHALL be left untouched

### Requirement: Dynamic ports by default
Without an explicit port, `make wt-up` SHALL select a free TCP port for the API and a free TCP port for the frontend. Selection SHALL be deterministic for a given worktree, so that the same ports are reused whenever they are still free, and SHALL fall back to another free port when they are not.

#### Scenario: No port provided
- **WHEN** a developer runs `make wt-up` with neither `API_PORT` nor `FRONTEND_PORT` (nor `FRONT_PORT`) set
- **THEN** the API and the frontend SHALL each listen on a free port, and these ports SHALL differ from each other
- **THEN** the default values `8080` and `5173` used by `make dev` SHALL NOT be implied

#### Scenario: Ports reused across restarts
- **WHEN** a developer runs `make wt-down` then `make wt-up` in the same worktree and the previously used ports are still free
- **THEN** the instance SHALL come up on the same API and frontend ports as before

#### Scenario: Previously used port now taken
- **WHEN** the port that would normally be chosen for a worktree is occupied by another process
- **THEN** the system SHALL select another free port instead of failing

#### Scenario: Different worktrees get different ports
- **WHEN** `make wt-up` is run in two worktrees with no explicit port
- **THEN** the two instances SHALL NOT share any port

### Requirement: Pinned ports
A developer SHALL be able to pin the frontend port with `FRONTEND_PORT` (alias `FRONT_PORT`) and the API port with `API_PORT`, each independently of the other. A port that is not pinned SHALL remain dynamic.

#### Scenario: Pinned frontend port only
- **WHEN** a developer runs `make wt-up FRONT_PORT=3000`
- **THEN** the frontend SHALL listen on port 3000 and the API SHALL listen on a dynamically selected free port
- **THEN** the frontend proxy SHALL target that API port

#### Scenario: Pinned API port only
- **WHEN** a developer runs `make wt-up API_PORT=8080`
- **THEN** the API SHALL listen on port 8080 and the frontend SHALL listen on a dynamically selected free port

#### Scenario: Both ports pinned
- **WHEN** a developer runs `make wt-up FRONT_PORT=3000 API_PORT=8080`
- **THEN** the frontend SHALL listen on port 3000 and the API on port 8080

#### Scenario: Alias equivalence
- **WHEN** a developer runs `make wt-up FRONTEND_PORT=3000`
- **THEN** the result SHALL be identical to `make wt-up FRONT_PORT=3000`

#### Scenario: Both spellings provided
- **WHEN** both `FRONTEND_PORT` and `FRONT_PORT` are provided with different values
- **THEN** the command SHALL fail with an explicit conflict message and SHALL NOT start anything

#### Scenario: Pinned port already in use
- **WHEN** a pinned port is occupied by another process
- **THEN** the command SHALL fail with an error naming the port and SHALL NOT silently use another one, and no partial instance SHALL remain running

### Requirement: No metrics port collision
When Prometheus metrics are enabled for an instance (`STALKEER_METRICS_ENABLED=true`), `make wt-up` SHALL give that instance its own free metrics port so that two worktree instances never contend for the default metrics port.

#### Scenario: Metrics enabled in two worktrees
- **WHEN** `make wt-up` is run in two worktrees with metrics enabled
- **THEN** both instances SHALL start and each SHALL expose its metrics on a distinct port

#### Scenario: Metrics disabled
- **WHEN** metrics are not enabled
- **THEN** no metrics port SHALL be opened by the instance

### Requirement: Dedicated database per worktree
Each worktree instance SHALL use its own PostgreSQL instance, isolated from the repository's `docker-compose.yml` services and from the instances of other worktrees, reachable only from the local machine on a dynamically selected host port.

#### Scenario: Database isolation
- **WHEN** an instance is running in worktree A and another in worktree B
- **THEN** each SHALL connect to a distinct PostgreSQL instance and the data of one SHALL NOT be visible from the other

#### Scenario: No collision with the standard development stack
- **WHEN** the standard `docker-compose.yml` Postgres is running on its usual host port
- **THEN** `make wt-up` SHALL start normally without stopping, reusing or modifying that Postgres

#### Scenario: Local access only
- **WHEN** an instance database is running
- **THEN** its host port SHALL be bound to the loopback interface only

#### Scenario: Schema creation
- **WHEN** the API server of a fresh instance starts
- **THEN** it SHALL create the schema through its normal startup migrations, without any manual step

### Requirement: Environment-only instance configuration
The instance SHALL be configured exclusively through process environment variables (database connection and API port), and `make wt-up` SHALL NOT create, copy or symlink a `config.yml` nor copy any credential or API key from another location.

#### Scenario: Worktree without config.yml
- **WHEN** `make wt-up` is run in a freshly created worktree that has no `config.yml`
- **THEN** the instance SHALL start successfully

#### Scenario: No secret propagation
- **WHEN** `make wt-up` completes
- **THEN** no file containing the integration keys or passwords of another checkout SHALL have been created in the worktree

### Requirement: Test data seeding on start
`make wt-up` SHALL load the repository's test data seed into the freshly created instance database, so the dashboard is not empty, unless the developer sets `WT_SEED=0`.

#### Scenario: Default seeding
- **WHEN** a developer runs `make wt-up`
- **THEN** after the command returns, the instance database SHALL contain the seeded test records

#### Scenario: Seeding disabled
- **WHEN** a developer runs `make wt-up WT_SEED=0`
- **THEN** the instance database SHALL contain no seeded records

#### Scenario: Seeding targets only the instance database
- **WHEN** the caller's shell exports `DB_HOST`, `DB_PORT`, `DB_NAME` or `STALKEER_DATABASE_*` pointing to another database, or a `config.yml` points to one
- **THEN** the seed SHALL still be applied only to the worktree's own database and the other database SHALL NOT be modified

### Requirement: Instance teardown
`make wt-down` SHALL stop the API server and the frontend dev server of the current worktree and SHALL destroy its dedicated database, including its stored data. It SHALL NOT affect instances of other worktrees.

#### Scenario: Full teardown
- **WHEN** a developer runs `make wt-down` in a worktree with a running instance
- **THEN** the API and frontend processes SHALL be stopped, the instance database container and its volume SHALL be removed, and the instance ports SHALL be released

#### Scenario: Data is not kept
- **WHEN** a developer runs `make wt-down` then `make wt-up` in the same worktree
- **THEN** the database SHALL be a fresh one (data created in the previous run SHALL be gone)

#### Scenario: Other worktrees untouched
- **WHEN** `make wt-down` is run in worktree A while worktree B has a running instance
- **THEN** the instance of worktree B SHALL keep running unchanged

#### Scenario: Teardown without local state
- **WHEN** the worktree's runtime state directory has been deleted and a developer runs `make wt-down`
- **THEN** the dedicated database container and volume of that worktree SHALL still be removed and any process left from that instance SHALL be stopped

#### Scenario: Nothing to tear down
- **WHEN** a developer runs `make wt-down` in a worktree with no instance
- **THEN** the command SHALL succeed and report that nothing was running

### Requirement: Failed start leaves nothing behind
If any component of `make wt-up` fails to start or to become reachable, the command SHALL fail with a non-zero status, report which component failed and where its log is, and release everything it had started.

#### Scenario: API fails to start
- **WHEN** the API server exits during startup
- **THEN** `make wt-up` SHALL exit non-zero, print the failing component and the path of its log file, and SHALL stop the database and frontend processes it had started

### Requirement: Instance inspection
The system SHALL provide `make wt-status` to show the state of the current worktree's instance and `make wt-logs` to display its logs.

#### Scenario: Status of a running instance
- **WHEN** a developer runs `make wt-status` in a worktree with a running instance
- **THEN** the output SHALL list the frontend URL, the API URL, the database host port, and whether each component is running and healthy

#### Scenario: Status without instance
- **WHEN** a developer runs `make wt-status` in a worktree with no instance
- **THEN** the output SHALL state that no instance is running

#### Scenario: Reading logs
- **WHEN** a developer runs `make wt-logs`
- **THEN** the system SHALL show the API and frontend logs of that worktree's instance

### Requirement: Frontend dependencies on demand
`make wt-up` SHALL install the frontend dependencies of the worktree when they are missing, and SHALL NOT reinstall them when already present.

#### Scenario: Fresh worktree
- **WHEN** `make wt-up` is run in a worktree where `frontend/node_modules` does not exist
- **THEN** the frontend dependencies SHALL be installed before the frontend dev server starts

#### Scenario: Dependencies already installed
- **WHEN** `frontend/node_modules` already exists
- **THEN** no installation SHALL be performed

### Requirement: Runtime state excluded from version control
Per-worktree runtime state (ports, process identifiers, logs) SHALL be stored under a `.wt/` directory that is ignored by git.

#### Scenario: Clean working tree
- **WHEN** an instance is running and a developer runs `git status`
- **THEN** no file created by `make wt-up` SHALL appear as untracked or modified

### Requirement: Worktree workflow documentation
The project documentation SHALL describe the `wt-*` targets, the dynamic and pinned port forms, the per-worktree database and seeding behavior, the fact that `wt-down` destroys the instance data, and the host prerequisites.

#### Scenario: Reading the documentation
- **WHEN** a developer reads the local development section of `README.md` or `docs/DEVELOPMENT.md`
- **THEN** they SHALL find the four `make wt-up` forms (dynamic, `FRONT_PORT`, `API_PORT`, both), `make wt-down`, and the `WT_SEED` option
