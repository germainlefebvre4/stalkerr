## 1. Backend: problem-filter pagination accuracy

- [x] 1.1 In `internal/api/handlers_frontend.go` (`listDownloadsEnriched`), when `problem` is set: fetch the full status/type-matched set (no DB-level `Limit`/`Offset`), build `file_info` for each row via the existing enrichment/`fileparser` logic, apply the problem predicate, then set `total`/`total_pages` from the filtered slice's length and slice `[offset:offset+limit]` before building the enriched response. Leave the existing DB-level `Limit`/`Offset`/`Count` path unchanged when `problem` is not set.
- [x] 1.2 Extend `TestListDownloadsEnriched` in `internal/api/handlers_frontend_test.go` with cases asserting `total`/`total_pages`/`data` are correct when `problem` is combined with pagination (e.g. more problem-matching rows than one page, and a page beyond the first) — verify with `go test ./internal/api/... -run TestListDownloadsEnriched -v`.
- [x] 1.3 Run the full backend suite and verify it passes: `make test`.

## 2. Frontend: remove auto-refresh polling

- [x] 2.1 In `frontend/src/hooks/useDownloads.ts`, remove the `setInterval(fetchDownloads, 5000)` effect and its `clearInterval` cleanup (lines ~39-44), keeping the initial `fetchDownloads` call on tab activation.
- [x] 2.2 Update/add a test in `frontend/src/hooks/useDownloads.test.ts` (or wherever the hook is tested) asserting no repeated fetch occurs after the initial one when no user action is taken (e.g. advance fake timers past 5s and assert the fetch mock was called only once) — verify with `npx vitest run useDownloads`.

## 3. Frontend: pagination state and API wiring

- [x] 3.1 In `frontend/src/services/api.ts`, add an `offset` parameter to `getDownloads`, appended to the URL as `&offset=${offset}` following the existing `getRunItems`/`getGroupedPlaylist` convention.
- [x] 3.2 In `frontend/src/hooks/useDownloads.ts`, add `downloadsPage`/`setDownloadsPage`, `downloadsLimit`/`setDownloadsLimit` (default `20`), and `downloadsTotal` state; compute `offset = (downloadsPage - 1) * downloadsLimit` when calling `api.getDownloads`, and set `downloadsTotal` from the response's `total`. Reset `downloadsPage` to `1` whenever `statusFilter`, `typeFilter`, `problemFilter`, or `downloadsLimit` changes. Export the new state/setters from the hook.
- [x] 3.3 Wire the new state/setters through `App.tsx` into `DownloadsTab` props, matching how `PlaylistTab` receives its pagination props.

## 4. Frontend: pagination UI

- [x] 4.1 In `frontend/src/components/DownloadsTab.tsx`, render `<Pagination total={downloadsTotal} page={downloadsPage} setPage={setDownloadsPage} limit={downloadsLimit} setLimit={setDownloadsLimit} limitOptions={[10, 50, 100]} />` below `DownloadsSummaryList`, mirroring the usage in `PlaylistTab.tsx`.
- [x] 4.2 Update/add tests in `frontend/src/components/DownloadsTab.test.tsx` asserting the pagination controls render and that changing page/limit triggers the corresponding fetch with the expected `offset`/`limit` — verify with `npx vitest run DownloadsTab`.

## 5. Verification

- [x] 5.1 Run the full frontend test suite and verify it passes: `cd frontend && npm test` (or `npx vitest run` from `frontend/`).
- [x] 5.2 Manually run the app (`run` skill or equivalent), open the Downloads tab, and verify: no network request to `/api/v1/downloads` fires without user action for at least 15s; the manual refresh button still re-fetches; page navigation and the items-per-page selector both work and match the displayed `total`; applying the "problem" filter and paginating shows a consistent, non-repeating set of results across pages.
