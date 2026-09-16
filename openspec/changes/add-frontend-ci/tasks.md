## 1. Fix the release-blocking build error

- [x] 1.1 Remove `value: null` from the `radarr.api_key` fixture in `frontend/src/components/SettingsGroupCard.test.tsx` (keep `is_set: true`) and verify `npm run build` (in `frontend/`) exits 0
- [x] 1.2 Verify `npm test` (in `frontend/`) still passes for `SettingsGroupCard.test.tsx` after the fixture change

## 2. Add the Frontend CI workflow

- [x] 2.1 Create `.github/workflows/frontend-ci.yml` triggered on `push`/`pull_request` to `main`/`develop` with `paths: ["frontend/**", ".github/workflows/frontend-ci.yml"]`
- [x] 2.2 Add a job that checks out the repo, sets up Node 24 (`actions/setup-node@v4`), and runs `npm ci`, `npm run lint`, `npm run build`, `npm test` inside `frontend/`, and verify the workflow YAML is valid (`gh workflow view frontend-ci.yml` or equivalent lint/parse)
- [ ] 2.3 Push a branch with only a `frontend/` change and verify the new workflow run appears and passes in the Actions tab

## 3. Scope the existing Go CI workflow

- [ ] 3.1 Add a `paths`/`paths-ignore` filter to `.github/workflows/ci.yml` so it does not run on frontend-only changes, and verify a frontend-only test branch no longer triggers the Go CI workflow (filter added; remote push verification pending, see note below)
- [ ] 3.2 Verify a backend-only change still triggers Go CI as before (no regression)

## 4. Validate the fix against the original failure

- [x] 4.1 Confirm locally (or via the new workflow's run log) that `npm run build` in `frontend/` no longer reproduces `TS2322` from the original failing run (34988595152)
