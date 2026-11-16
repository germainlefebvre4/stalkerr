# Proposal

## Why

Dependabot PR #69 bumps `react` to 19.3.1 and `@types/react` but leaves `react-dom` at 19.2.8. `react-dom` refuses to load when its version differs from `react`, so 31 of the 35 frontend test suites crash at import and the `Test` check fails. The current `.github/dependabot.yml` has no grouping for npm, so the same split can happen on every React release.

## What Changes

- Unblock PR #69 by bumping `react-dom` to 19.3.1 and `@types/react-dom` to its latest 19.3.x on the PR branch, regenerating `frontend/package-lock.json`, and confirming `npm run lint`, `npm run build` and `npm test` pass in `frontend/`.
- Add a Dependabot `groups` entry for the npm ecosystem in `/frontend` so that `react`, `react-dom`, `@types/react` and `@types/react-dom` are always proposed in one PR.
- Document this grouping as a requirement so it is not removed by accident.

## Capabilities

### New Capabilities
- `dependabot-update-grouping`: rules for how Dependabot groups interdependent dependencies (starting with the React runtime and its type packages) into a single update PR.

### Modified Capabilities
<!-- None: frontend-ci requirements are unchanged; the CI job already catches this failure. -->

## Impact

- `.github/dependabot.yml` (new `groups` block under the `npm` `/frontend` entry).
- `frontend/package.json` and `frontend/package-lock.json` (on PR #69's branch: `react-dom`, `@types/react-dom`).
- No runtime code or API change. The `<ViewTransition />` API from React 19.3 is available but not adopted by this change.
- Side effect: pushing a commit to Dependabot's branch stops Dependabot from auto-rebasing PR #69.
