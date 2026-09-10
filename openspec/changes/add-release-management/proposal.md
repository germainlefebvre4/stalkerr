## Why

Releases today are unversioned and manually published: there are zero git tags, `CHANGELOG.md` doesn't exist, the CLI `version` command prints a hardcoded `"Stalkeer v0.1.0"` string, and Docker images are pushed ad hoc via `make docker-build-push` whenever someone remembers to run it. The build already threads `-X main.version/commit/date` ldflags through the Dockerfile and Makefile, but no `cmd` package variable exists to receive them, so that metadata is silently discarded and nothing running in production can be traced back to a commit. Conventional Commits + semantic-release automation removes the manual versioning/tagging/publishing steps, and surfacing the resulting version/commit/date in the app closes the loop so a running instance can be identified at a glance.

## What Changes

- Enforce Conventional Commits on `main` via a CI check (commit/PR-title lint).
- Add a release-please workflow + manifest that tracks one unified semantic version (bootstrapped at `1.0.0`) across the backend, frontend, and Helm chart, generating `CHANGELOG.md` and a tagged GitHub Release from a release PR.
- Add a Docker publish workflow, triggered by each release-please release, that builds multi-arch (`linux/amd64` + `linux/arm64`) backend and frontend images via buildx and pushes them to Docker Hub under `germainlefebvre4/stalkerr` and `germainlefebvre4/stalkerr-frontend`, tagged with the release version and `latest`. This replaces the manual `make docker-build-push` flow for tagged releases.
- Wire the existing but currently-orphaned `-X main.version/commit/date` ldflags into real `cmd` package variables, and fix the `version` CLI command to print the actual injected build metadata instead of a hardcoded string.
- Extend `GET /api/v1/system/status` to report the running build's version, commit, and date.
- Extend the frontend System Status dialog to display version, commit, and date from that endpoint.
- Decouple the published Docker image name (`stalkerr`) from the Go module/binary name (`stalkeer`) in the Makefile, so release automation and the existing `docker-compose.yml`/Helm chart naming agree.

## Capabilities

### New Capabilities
- `release-automation`: CI enforces Conventional Commits; release-please automates a single unified semantic version, `CHANGELOG.md`, and tagged GitHub Release across backend, frontend, and Helm chart; a Docker publish workflow builds and pushes multi-arch backend/frontend images to Docker Hub on each release.
- `build-version-metadata`: The binary embeds version, commit, and build date at build time via ldflags, and the `stalkeer version` CLI command reports them.

### Modified Capabilities
- `system-status-api`: The aggregated status response gains version/commit/date fields sourced from the build metadata.
- `system-status-view`: The System Status dialog displays the version/commit/date reported by the API.

## Impact

- **CI**: new `.github/workflows/release-please.yml` and a Docker publish workflow (buildx + QEMU for multi-arch); Docker Hub credentials added as repository secrets.
- **Backend**: `cmd/main.go` (or a new file) gains `version`/`commit`/`date` package vars; `cmd/version.go` reads them; `internal/api/system_status.go` response struct and handler gain the new fields.
- **Frontend**: `frontend/src/types.ts` (`SystemStatusResponse`), `frontend/src/components/SystemStatusDialog.tsx`, and the `dialogs` i18n namespace gain the version/commit/date display.
- **Build/release tooling**: `Makefile` gains an image-name variable decoupled from `BINARY_NAME`; `charts/stalkerr/Chart.yaml` (`version`, `appVersion`) and `frontend/package.json` (`version`) are kept in sync by the release automation's file updaters.
- **Consumer-facing**: Docker Hub publishing moves from the existing ad hoc `germainlefebvre4/stalkeer` / `germainlefebvre4/stalkeer-frontend` repositories (already live, ~2954 / ~1879 pulls) to `germainlefebvre4/stalkerr` / `germainlefebvre4/stalkerr-frontend`. Anyone currently pulling the `stalkeer`-named images stops receiving new tags after this change ships.
