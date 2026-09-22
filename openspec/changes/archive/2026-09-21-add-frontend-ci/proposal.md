## Why

The v1.2.0 release pipeline failed at `docker-publish-frontend` (run [34988595152](https://github.com/germainlefebvre4/stalkerr/actions/runs/34988595152/job/104447119559)): `npm run build` (`tsc && vite build`) failed with `TS2322` in `SettingsGroupCard.test.tsx`. The frontend has no CI job at all — `ci.yml` only tests/builds the Go backend — so `tsc`, `eslint`, and `vitest` never run on a pull request. The first time the frontend is actually built is inside the release-time Docker image build, which runs only after `release-please` has already cut the tag and GitHub Release. A broken frontend build is therefore caught after the release exists, not before merge, leaving a release with no corresponding Docker image.

## What Changes

- Add a "Frontend CI" GitHub Actions workflow that runs on pull requests and pushes touching `frontend/**`, executing `npm ci`, `npm run lint`, `npm run build` (typecheck + Vite build), and `npm test`.
- Fix the pre-existing type error in `frontend/src/components/SettingsGroupCard.test.tsx` (`value: null` on a sensitive field) so the frontend build passes; sensitive fields already use `is_set` and must omit `value` rather than set it to `null`.
- Scope the existing Go CI workflow to backend paths (and this new workflow to frontend paths) so unrelated changes don't trigger the other stack's checks.

## Capabilities

### New Capabilities
- `frontend-ci`: Pull-request/push CI validation for the frontend (lint, typecheck+build, unit tests) before code reaches `main`.

### Modified Capabilities
(none — `release-automation`'s Docker publish behavior is unchanged; this change only adds an earlier, pre-merge gate)

## Impact

- New file: `.github/workflows/frontend-ci.yml` (or a new job appended to `ci.yml` — decided in design.md).
- Modified: `.github/workflows/ci.yml` (path filters), `frontend/src/components/SettingsGroupCard.test.tsx` (bug fix).
- No changes to `frontend/Dockerfile` or `release-please.yml`; release-time Docker builds remain the final safety net, not the primary gate.
