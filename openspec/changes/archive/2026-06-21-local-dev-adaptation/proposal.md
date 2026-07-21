## Why

With the addition of the modern React 19 Frontend Web Dashboard (IHM), the local development experience needs to be updated. Currently, there are no Makefile targets to install frontend dependencies, compile frontend code, or launch both backend and frontend concurrently. Updating the Makefile and README.md allows developers to start, run, and develop on Stalkeer in a highly efficient and standardized manner.

## What Changes

- Add native frontend targets to the `Makefile` (`front-install`, `front-dev`, `front-build`, `front-lint`, `dev`).
- Update the project's root `README.md` to introduce the new Web Dashboard, list its features, and provide clear step-by-step local development setup guidelines.

## Capabilities

### New Capabilities

- `local-dev-commands`: Makefile targets allowing developers to install dependencies, run development servers, lint code, and build static assets for both frontend and backend.
- `local-dev-documentation`: Detailed README.md documentation covering prerequisites, environment setup, and instructions for running both the API server and the React dashboard locally.

### Modified Capabilities

<!-- No existing capabilities are being modified -->

## Impact

- **New Files**: No new files are introduced; we only extend existing root administrative files.
- **Modified Files**:
  - `Makefile` (adding frontend and hybrid orchestration targets).
  - `README.md` (adding dashboard features and updated quickstart guide).
