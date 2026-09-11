## 1. Conventional Commit enforcement

- [x] 1.1 Add a CI workflow (or job) that lints PR titles/commits against the Conventional Commits format and verify it fails on a deliberately malformed PR title and passes on a conforming one
- [x] 1.2 Document the enforced format expectation (link to the existing `.github/skills/git-commit/SKILL.md` convention) so contributors know why the check exists

## 2. release-please setup

- [x] 2.1 Add `release-please-config.json` configured as a single root-level "simple" component, bootstrapped at `1.0.0`, and a matching `.release-please-manifest.json`
- [x] 2.2 Add `extra-files` updaters in the release-please config for `charts/stalkerr/Chart.yaml` (`version` and `appVersion` fields) and `frontend/package.json` (`version` field), and verify a manual dry run (`release-please release-pr --dry-run` or equivalent) shows all three files staged for the same version
- [x] 2.3 Add `.github/workflows/release-please.yml` running `googleapis/release-please-action` on push to `main`, and verify it opens a release PR titled with the next version after a `feat`/`fix` commit lands on `main`
- [ ] 2.4 Verify merging the release PR creates a git tag `v1.0.0`, publishes a GitHub Release, and updates `CHANGELOG.md`

## 3. Docker Hub multi-arch publish

- [x] 3.1 Add a Docker publish workflow (or job) gated on the release-please action's `release_created` output, using `docker/setup-qemu-action` + `docker/setup-buildx-action` + `docker/build-push-action`
- [x] 3.2 Configure the backend build to push `germainlefebvre4/stalkerr:vX.Y.Z` and `germainlefebvre4/stalkerr:latest` for `linux/amd64,linux/arm64`, passing `VERSION`/`COMMIT`/`DATE` build args from the release tag, resolved commit SHA, and workflow run timestamp
- [x] 3.3 Configure the frontend build to push `germainlefebvre4/stalkerr-frontend:vX.Y.Z` and `germainlefebvre4/stalkerr-frontend:latest` for the same two platforms
- [ ] 3.4 Add `DOCKERHUB_USERNAME`/`DOCKERHUB_TOKEN` as GitHub Actions repository secrets and wire them into the `docker/login-action` step
- [ ] 3.5 Verify end-to-end by merging a release PR (or manually dispatching the workflow against an existing tag) and confirming both images and both architectures appear on Docker Hub under the new tag

## 4. Build version metadata

- [x] 4.1 Add `var version, commit, date string` (defaulting to `"dev"`/`"unknown"`) to `cmd/main.go`, matching the existing `-X main.version/commit/date` ldflags in `Dockerfile`/`Makefile`, and verify `go build` (no ldflags) still runs and reports the defaults
- [x] 4.2 Rewrite `cmd/version.go` to print the actual `version`, `commit`, and `date` values instead of the hardcoded `"Stalkeer v0.1.0"` string, and verify `./bin/stalkeer version` reflects flags passed via `make build VERSION=... `
- [x] 4.3 Add/adjust a unit test for the `version` command's output format

## 5. Expose version metadata through the API

- [x] 5.1 Extend the `SystemStatusResponse` struct and `getSystemStatus` handler in `internal/api/system_status.go` to include version, commit, and date fields sourced from the `cmd`-level build metadata, and verify a test asserts these fields are present in the JSON response
- [x] 5.2 Verify `GET /api/v1/system/status` returns the fields end-to-end when the binary is built with version ldflags set

## 6. Display version metadata in the frontend

- [x] 6.1 Add `version`, `commit`, `date` fields to `SystemStatusResponse` in `frontend/src/types.ts`
- [x] 6.2 Extend `SystemStatusDialog.tsx` to render the version/commit/date, add the corresponding `dialogs` i18n keys, and verify the existing `SystemStatusDialog.test.tsx` is updated/passing for the new content
- [x] 6.3 Manually run the app (`make dev` or equivalent) and confirm the System Status dialog shows version/commit/date matching the running build

## 7. Align image naming

- [x] 7.1 Add an `IMAGE_NAME ?= stalkerr` variable to the `Makefile`, decoupled from `BINARY_NAME`, and repoint `docker-build-versioned`/`docker-push` targets (backend and `-frontend`) at `$(REGISTRY)/$(IMAGE_NAME)`
- [x] 7.2 Verify `make docker-build-versioned VERSION=test` produces images tagged `germainlefebvre4/stalkerr:test` and `germainlefebvre4/stalkerr-frontend:test` locally
- [x] 7.3 Confirm `docker-compose.yml` and `charts/stalkerr/` already reference `stalkerr` consistently (no changes expected there; note explicitly if any drift is found)
