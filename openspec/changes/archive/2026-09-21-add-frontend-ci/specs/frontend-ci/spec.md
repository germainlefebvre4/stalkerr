## Purpose

Validate the frontend (types, lint, build, unit tests) on every pull request and push, so a broken frontend build is caught before merge instead of at release time.

## ADDED Requirements

### Requirement: Frontend checks run in CI on relevant changes
CI SHALL run a frontend validation job on every pull request and push to `main`/`develop` that changes a file under `frontend/**` (or the workflow file itself). The job SHALL install dependencies, then run lint, typecheck+build, and unit tests, failing the check if any step fails.

#### Scenario: Frontend change triggers the checks
- **WHEN** a pull request or push modifies a file under `frontend/**`
- **THEN** CI SHALL run `npm ci`, `npm run lint`, `npm run build`, and `npm test` in the `frontend/` directory
- **AND** the check SHALL fail if any of those commands exits non-zero

#### Scenario: Backend-only change does not trigger frontend checks
- **WHEN** a pull request or push changes only backend Go files (no file under `frontend/**`)
- **THEN** the frontend CI job SHALL NOT run

### Requirement: Build failures are caught before merge, not at release
A pull request introducing a type error, lint violation, or failing unit test in `frontend/` SHALL be blocked from being merge-ready by a failing CI check, rather than only surfacing when `frontend/Dockerfile`'s `npm run build` runs during release-time image publishing.

#### Scenario: Type error in a frontend file fails CI on the PR
- **WHEN** a pull request introduces a TypeScript type error anywhere under `frontend/src` (including test files)
- **THEN** the frontend CI check on that pull request SHALL fail with the `tsc` error
- **AND** the same error SHALL NOT first appear during the release `docker-publish-frontend` job
