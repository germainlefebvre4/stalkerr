# Tasks

## 1. Group React packages in Dependabot (prevention)

- [x] 1.1 Add a `groups.react` block with the patterns `react`, `react-dom`, `@types/react`, `@types/react-dom` under the npm `/frontend` entry of `.github/dependabot.yml`; verify the file is valid YAML (`python3 -c 'import yaml,sys; yaml.safe_load(open(".github/dependabot.yml"))'`) and that the `gomod` and `github-actions` entries are unchanged in `git diff`
- [ ] 1.2 Open a PR for this change with a Conventional Commits title (e.g. `ci(deps): group react packages in dependabot`); verify the `Lint PR title` check passes and, after merge, that GitHub's Dependabot page (Insights > Dependency graph > Dependabot) reports no configuration error

## 2. Unblock PR #69 (react-dom mismatch)

- [x] 2.1 Check out branch `dependabot/npm_and_yarn/frontend/multi-7f19880bf6` and, in `frontend/`, bump `react-dom` and `@types/react-dom` to the `19.3.x` line with `npm install react-dom@19.3.1 @types/react-dom@19.3.0`; verify `npm ls react react-dom` shows both at 19.3.1 with no `invalid` or `UNMET` entries
- [x] 2.2 Run `npm run lint`, `npm run build` and `npm test` in `frontend/`; verify all three exit 0 and the test run reports 35 passing suites. If a failure unrelated to the version mismatch appears, stop and decide between fixing it and closing #69 (see design D2 alternatives)
- [ ] 2.3 Commit only `frontend/package.json` and `frontend/package-lock.json` (e.g. `fix(deps): bump react-dom and @types/react-dom with react`) and push to the PR branch; verify the `Test` check on #69 turns green and `git diff main --stat` lists only those two files
- [ ] 2.4 Merge #69 once green; verify `main` has `react` and `react-dom` at the same version in `frontend/package.json`

## 3. Integration check

- [ ] 3.1 After both groups are merged, wait for the next weekly Dependabot run (or trigger it from the Dependabot page) and verify any React-related update arrives as a single grouped PR containing the four packages together

## Workflow follow-up

- Archive the change after #69 and the Dependabot configuration PR are merged.
- Verify the new `dependabot-update-grouping` spec appears in `openspec list --specs` after archive.
