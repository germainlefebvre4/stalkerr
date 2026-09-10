## Purpose

Automate versioning, changelog generation, and Docker Hub publishing for the app directly from Conventional Commits, so cutting a release requires no manual tagging, changelog editing, or image push.

## ADDED Requirements

### Requirement: Conventional Commit validation in CI
Commits merged into `main` SHALL follow the Conventional Commits format. CI SHALL validate this on every pull request and fail the check when it is not met.

#### Scenario: Conforming commit passes
- **WHEN** a pull request's commits (or its title, for a squash merge) follow the Conventional Commits format
- **THEN** the CI check SHALL pass

#### Scenario: Non-conforming commit fails
- **WHEN** a pull request's commits (or title) do not follow the Conventional Commits format
- **THEN** the CI check SHALL fail with a message indicating the expected format

### Requirement: Unified semantic version across the monorepo
The system SHALL automatically compute the next semantic version from Conventional Commits merged into `main` (patch for `fix`, minor for `feat`, major for a breaking change), starting from `1.0.0`, and apply that single version to the backend, the frontend, and the Helm chart together.

#### Scenario: Fix commit bumps patch version
- **WHEN** only `fix` commits have merged since the last release
- **THEN** the next proposed version SHALL increment the patch number

#### Scenario: Feature commit bumps minor version
- **WHEN** at least one `feat` commit has merged since the last release
- **THEN** the next proposed version SHALL increment the minor number

#### Scenario: Breaking change bumps major version
- **WHEN** a merged commit is marked as breaking (`!` after the type/scope, or a `BREAKING CHANGE` footer)
- **THEN** the next proposed version SHALL increment the major number

### Requirement: Release pull request accumulates pending changes
The system SHALL maintain an open release pull request that accumulates the pending version bump and changelog entries as Conventional Commits merge into `main`, updating `CHANGELOG.md`, until that pull request is merged.

#### Scenario: New conventional commit updates the pending release PR
- **WHEN** a new Conventional Commit merges into `main` while a release pull request is open
- **THEN** the release pull request SHALL be updated to include that commit's changelog entry and reflect any resulting version bump

### Requirement: Merging the release PR cuts a tagged release
Merging the release pull request SHALL create a git tag and a GitHub Release at the computed version.

#### Scenario: Release PR merge creates tag and release
- **WHEN** the release pull request is merged into `main`
- **THEN** a git tag matching the computed version SHALL be created and a corresponding GitHub Release SHALL be published

### Requirement: Multi-arch Docker images published on release
When a release is published, the system SHALL build and push `linux/amd64` and `linux/arm64` images for the backend and frontend to Docker Hub under `germainlefebvre4/stalkerr` and `germainlefebvre4/stalkerr-frontend` respectively, tagged with the release version and `latest`.

#### Scenario: Release triggers image publish
- **WHEN** a GitHub Release is published at version `vX.Y.Z`
- **THEN** images `germainlefebvre4/stalkerr:vX.Y.Z`, `germainlefebvre4/stalkerr:latest`, `germainlefebvre4/stalkerr-frontend:vX.Y.Z`, and `germainlefebvre4/stalkerr-frontend:latest` SHALL exist on Docker Hub for both `linux/amd64` and `linux/arm64`

#### Scenario: No release, no publish
- **WHEN** commits merge into `main` without a release being published (only accumulating in the open release pull request)
- **THEN** no image SHALL be pushed to Docker Hub
