# Design

## Context

- `.github/dependabot.yml` declares three ecosystems (gomod, npm `/frontend`, github-actions), each weekly, with no `groups`. Without grouping, Dependabot may open one PR per package, or group only the packages it considers related, as it did on PR #69.
- `react-dom` checks at import time that its version equals `react`'s. A split bump therefore breaks every test suite that imports React, not just one.
- `frontend/package.json` declares `react` and `react-dom` with caret ranges (`^19.2.8`), and `@types/react` / `@types/react-dom` separately. `frontend-ci.yml` already runs lint, build and tests, so it detected the problem. Detection is not the issue; prevention is.
- PR #69 comes from Dependabot's branch `dependabot/npm_and_yarn/frontend/multi-7f19880bf6`.

## Goals / Non-Goals

**Goals:**
- Make React runtime and type packages always move together in Dependabot PRs.
- Get PR #69 green with the smallest possible diff.

**Non-Goals:**
- Grouping other npm dependencies (vitest, vite, eslint, etc.).
- Changing Go or GitHub Actions update behavior.
- Adopting React 19.3 features such as `<ViewTransition />`.
- Pinning exact React versions (`react` and `react-dom` keep their caret ranges).

## Decisions

### D1: Use Dependabot `groups` with name-based `patterns`

Add under the npm `/frontend` entry:

```yaml
groups:
  react:
    patterns:
      - "react"
      - "react-dom"
      - "@types/react"
      - "@types/react-dom"
```

Exact names, not `react*`, so unrelated packages such as `react-router` or `react-i18next` are not pulled into the group.

*Alternatives:*
- `react*` wildcard: simpler, but would capture ecosystem libraries with independent release cycles.
- npm `overrides` to force a single version: hides the mismatch instead of preventing it and fights the lockfile.
- Exact version pins: prevents drift but stops caret-range updates and adds manual work on every release.

### D2: Fix PR #69 by committing to its branch (not by closing it)

Check out the Dependabot branch, bump `react-dom` and `@types/react-dom` to the same release line as `react` (`19.3.x`), regenerate the lockfile with `npm install` in `frontend/`, and verify with `npm run lint && npm run build && npm test` before pushing.

*Alternatives:*
- Close #69 after D1 and let Dependabot recreate a grouped PR: cleaner history, no human commit on a bot branch, but slower and relies on Dependabot's next run. It remains a valid fallback if the bump fails.
- `@dependabot recreate` after merging D1: same as above, can be triggered on demand.

### D3: Order of work

Merge D1 first only if the fallback of D2 is chosen. Otherwise the two are independent and can ship in any order; D1 only affects future PRs.

## Risks / Trade-offs

- [Pushing to the Dependabot branch disables its auto-rebase] → Merge #69 promptly after CI is green, or fall back to closing and recreating.
- [React 19.3.x may introduce failures beyond the version mismatch (types, test utilities)] → Run the full local suite before pushing and read the first real failures rather than assuming the bump is enough. If it is not, stay on 19.2.x by closing #69.
- [`@types/react-dom` latest is 19.3.0 while `react-dom` is 19.3.1, so type packages may lag the runtime] → Acceptable: type packages only need to share the `19.x` line. The grouping requires them to move together, not to share a patch number.
- [Grouping delays a React update until all four packages have a release] → Acceptable; a delayed update is cheaper than a broken PR.
