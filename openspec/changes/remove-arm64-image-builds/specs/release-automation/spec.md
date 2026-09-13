## MODIFIED Requirements

### Requirement: Multi-arch Docker images published on release
When a release is published, the system SHALL build and push `linux/amd64` images for the backend and frontend to Docker Hub under `germainlefebvre4/stalkerr` and `germainlefebvre4/stalkerr-frontend` respectively, tagged with the release version and `latest`. `linux/arm64` publishing is temporarily suspended (see design.md for the reintroduction plan) and is not required by this version of the requirement.

#### Scenario: Release triggers image publish
- **WHEN** a GitHub Release is published at version `vX.Y.Z`
- **THEN** images `germainlefebvre4/stalkerr:vX.Y.Z`, `germainlefebvre4/stalkerr:latest`, `germainlefebvre4/stalkerr-frontend:vX.Y.Z`, and `germainlefebvre4/stalkerr-frontend:latest` SHALL exist on Docker Hub for `linux/amd64`

#### Scenario: No release, no publish
- **WHEN** commits merge into `main` without a release being published (only accumulating in the open release pull request)
- **THEN** no image SHALL be pushed to Docker Hub
