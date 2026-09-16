## Context

See proposal.md - Why. Today `.github/workflows/ci.yml` ("Go CI") only runs `go test`/`go build`, triggered on push/PR to `main`/`develop` with no path filter. There is no frontend equivalent. `frontend/package.json` already exposes `lint`, `build` (`tsc && vite build`), and `test` (`vitest run`) scripts, so this change only needs to wire CI to call them — no new tooling to introduce.

## Goals / Non-Goals

**Goals:**
- Run lint, typecheck+build, and unit tests for `frontend/` on every PR/push that touches it, before merge.
- Keep the Go and frontend checks independent so a change to one stack doesn't wait on or fail because of the other.

**Non-Goals:**
- Changing the release pipeline (`release-please.yml`) or the Docker image build itself — release-time build stays as the final safety net.
- Adding new lint rules, tests, or CI steps beyond what `package.json` already defines.
- Multi-version/matrix testing (single Node version is enough, matching the Go job's single-version approach).

## Decisions

- **New `frontend-ci.yml` workflow, not a job appended to `ci.yml`.** `ci.yml` is named "Go CI" and scoped to the backend; keeping frontend checks in their own file mirrors the existing separation of concerns (`commitlint.yml` is already its own file) and lets each workflow's trigger `paths:` filter stay simple and file-specific, rather than one workflow with mixed per-job path conditions.
- **Path-filter both workflows** (`paths: ["frontend/**"]` on the new workflow; `paths-ignore: ["frontend/**"]`, or an explicit `paths:` list of backend paths, on `ci.yml`) so a pure-frontend PR doesn't run Go tests and vice versa.
- **Node version: pin to the same major as `frontend/Dockerfile`'s builder stage (Node 24-alpine)** via `actions/setup-node@v4` with `node-version: 24`, so CI and the release Docker build type-check under the same runtime.
- **Fix the failing test inline in this change** (remove `value: null` from the sensitive-field fixture in `SettingsGroupCard.test.tsx`, since sensitive fields already model "no value" via `is_set` alone) rather than widening `SettingsField.value` to accept `null` — the type already correctly requires sensitive fields to omit `value`, and the test was the thing not respecting the contract.

## Risks / Trade-offs

- [Two workflow files to keep in sync on shared concerns like checkout/action versions] → low risk given they already diverge today (different runtimes); no shared logic to duplicate beyond `actions/checkout@v7`.
- [`npm ci` cache not configured] → acceptable for now; can add `actions/setup-node`'s built-in npm cache later as a fast-follow, not required to fix the release-blocking gap.
