## Context

See `proposal.md` - Why for motivation. Relevant current state:

- Zero git tags exist; there is no `CHANGELOG.md`.
- `Dockerfile` and `Makefile` already compute `-ldflags "-X main.version=... -X main.commit=... -X main.date=..."`, but `cmd/main.go` never declares matching `version`/`commit`/`date` package vars, so the flags bind to nothing. `cmd/version.go` prints a hardcoded `"Stalkeer v0.1.0"`.
- `GET /api/v1/system/status` (`internal/api/system_status.go`) and `SystemStatusDialog.tsx` already exist and fetch-on-open; this is the agreed home for version display.
- Docker Hub already has live, actively-pulled repos `germainlefebvre4/stalkeer` and `germainlefebvre4/stalkeer-frontend` (~2954 / ~1879 pulls), populated by manually running `make docker-build-push`. `docker-compose.yml` and the Helm chart (`charts/stalkerr/`) instead reference the name `stalkerr` — but only as a local build tag/label, since compose always builds from the local Dockerfile (`build: *stalkerr-build`) rather than pulling. The decision (confirmed with the user) is to canonicalize the *published* name on `stalkerr`, matching compose/Helm, and let the old `stalkeer`-named repos go stale.
- The Go module (`github.com/glefebvre/stalkeer`) and binary name stay `stalkeer` — only the published Docker image name changes to `stalkerr`. The Makefile's image tag is currently derived from `BINARY_NAME`; it needs its own `IMAGE_NAME` variable so it can diverge from the binary name.
- Frontend `package.json` is at `1.0.0`; Helm chart `Chart.yaml` has `version: 0.1.0` / `appVersion: "1.0.0"`. Confirmed decision: one unified semver for the whole monorepo, bootstrapped at `1.0.0`, applied to backend, frontend, and Helm chart together.
- No existing CI step builds or pushes Docker images; `.github/workflows/ci.yml` only runs `go test` and a plain `go build`.

## Goals / Non-Goals

**Goals:**
- Fully automate: commit -> release PR -> tag/GitHub Release -> multi-arch Docker Hub publish, with zero manual `make docker-*` steps for tagged releases.
- Make the running app's version/commit/date self-reporting, end to end (build -> CLI -> API -> UI).
- Keep backend, frontend, and Helm chart versions from drifting apart.

**Non-Goals:**
- Renaming the Go module, binary, repository, or any internal package names (only the *published Docker image* name changes).
- Retiring or redirecting the existing `germainlefebvre4/stalkeer` Docker Hub repos - they're simply left as-is and stop receiving new pushes.
- Changing how `docker-compose.yml` builds locally (it already builds from source, not from a registry pull).
- A frontend-only or independently-releasable frontend build; the frontend continues to be versioned and released together with the backend.

## Decisions

### release-please for version/changelog/tag automation
Use `googleapis/release-please-action`, configured as a single ("simple") release-please component at the repo root, with a manifest tracking the current version. Rationale: it's the de facto standard for Conventional-Commits-driven releases, needs no separate commit-lint tool to compute bumps (it parses Conventional Commits itself), and supports `extra-files` generic updaters for bumping version strings in files it doesn't natively understand (the Helm chart, `frontend/package.json`).

Alternative considered: `semantic-release`. Similar capability, but release-please's PR-based flow (a standing "release PR" the user reviews and merges) gives an explicit human gate before a release/tag/publish is cut, which fits better than semantic-release's typical fully-automatic tag-on-merge-to-main behavior.

### One release-please component, not per-package
Rejected a multi-package manifest (independent backend/frontend/Helm versions) per the confirmed decision: the app is presented and consumed as one unit, and the Makefile already tags both images with the same `$(VERSION)` today.

### Commit-message linting: PR-level check, not a commit-msg hook
Validate Conventional Commits format in CI (e.g. a PR-title/commits lint action) rather than a local `commit-msg` git hook. Rationale: a CI check is enforced for every contributor regardless of local tooling, and doesn't require every clone to install a hook.

### Docker publish workflow triggered by release-please's output
A separate workflow (or a second job in the same workflow) watches `release-please-action`'s `release_created` output and, only when true, runs the multi-arch build/push. This keeps "did we cut a release" and "did we publish images" as distinct, inspectable steps, and avoids pushing images for every merge to `main` (only for actual releases).

### Multi-arch via `docker/setup-qemu-action` + `docker/build-push-action` with buildx
Standard approach for cross-building `linux/arm64` on GitHub's `ubuntu-latest` (amd64) runners. Both the backend and frontend Dockerfiles are built this way; the backend Dockerfile already accepts `VERSION`/`COMMIT`/`DATE` build args, so the workflow passes the release tag, resolved commit SHA, and build timestamp straight through.

### Version metadata: backend is the single source of truth
The frontend does not get its own build-time-injected version. It only ever displays what `GET /api/v1/system/status` reports. Rationale: this was an explicit decision (extend the System Status dialog, sourced from the API) - it also avoids a second place where version/commit/date could drift out of sync between the two containers if they're ever deployed at slightly different times.

### Makefile: decouple image name from binary name
Add `IMAGE_NAME ?= stalkerr` (distinct from `BINARY_NAME := stalkeer`), and point `docker-build-versioned`/`docker-push` at `$(REGISTRY)/$(IMAGE_NAME)` and `$(REGISTRY)/$(IMAGE_NAME)-frontend`. These manual targets remain available for local/manual use; they're just no longer the only way images get published.

## Risks / Trade-offs

- **[Risk]** Abandoning the `germainlefebvre4/stalkeer` repos silently breaks anyone currently pulling `latest` from them (real pull counts exist: ~2954/~1879). -> **Mitigation**: none automated in this change (no redirect/deprecation notice is in scope); call this out clearly in the release notes / README once the new images are live. Purely an operational/communication follow-up, not a spec requirement.
- **[Risk]** release-please's Conventional-Commit parsing is strict about type/format; existing history is only loosely conventional (capitalized descriptions, no `!`/`BREAKING CHANGE` markers yet). -> **Mitigation**: this doesn't affect adoption - release-please only parses commits *after* it's introduced; past history doesn't need to conform. The new CI lint check catches format drift going forward.
- **[Risk]** Multi-arch builds roughly double CI build time (QEMU-emulated `arm64`). -> **Mitigation**: acceptable for a release-triggered (not every-merge) workflow; can add native `arm64` runners later if it becomes a bottleneck.
- **[Trade-off]** A human still has to merge the release PR to cut a release (not fully automatic on every merge to `main`). This is intentional - it's the confirmed release-please-style gate, not an oversight.

## Migration Plan

1. Add the Conventional Commit CI lint check first (no behavior change to releases yet).
2. Add the release-please workflow + manifest, bootstrapped at `1.0.0`, with `extra-files` updaters for `charts/stalkerr/Chart.yaml` and `frontend/package.json`. Merging its first release PR creates tag `v1.0.0` and a GitHub Release - no Docker publish wired yet, so this is a safe dry run.
3. Add the Docker publish workflow gated on `release_created`, targeting `germainlefebvre4/stalkerr` / `germainlefebvre4/stalkerr-frontend`, multi-arch. Verify manually against the `v1.0.0` release (or wait for the next release PR to merge).
4. Wire `cmd/main.go` version/commit/date vars, fix `cmd/version.go`, extend the system-status API response and the frontend dialog.
5. Update the Makefile's `IMAGE_NAME`/registry targets to match the new canonical name.

No rollback complexity beyond reverting the workflow files - nothing here is destructive to existing deploys (`docker-compose.yml` still builds locally either way).
