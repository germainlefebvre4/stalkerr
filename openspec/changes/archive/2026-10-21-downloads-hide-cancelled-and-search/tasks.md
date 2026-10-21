# Tasks

## 1. Backend: filters, search and ordering

- [x] 1.1 In `listDownloadsEnriched` (`internal/api/handlers_frontend.go`), replace the `JOIN processed_lines ... Distinct()` type filter with an `EXISTS` subquery on `processed_lines`; verify existing `type` filter tests in `handlers_frontend_test.go` still pass
- [x] 1.2 Accept `status=active` (`status <> 'cancelled'`) and leave an omitted `status` unconstrained; add tests for `active`, a concrete status, and no status
- [x] 1.3 Add the `search` parameter (case-insensitive substring via `getLikeOp`, escaped `%`/`_`, matching movie/TV show `tmdb_title`, `tvg_name`, and `download_path`) as an `EXISTS`/column predicate; add tests for title match, case-insensitivity, `tvg_name`/file-name fallback, combination with `status`/`type`/pagination, and no duplicate rows when several processed lines match
- [x] 1.4 Apply the status-rank `ORDER BY CASE ... , download_info.updated_at DESC` on both the SQL-paginated path and the `problem` Go path; add tests that cancelled sorts last, in-progress first, and recency orders within a rank, plus a test that every `models.DownloadStatus` constant is ranked
- [x] 1.4a Verify the rank/order SQL runs on PostgreSQL semantics (no `SELECT DISTINCT` left in the query) by inspecting the generated SQL in a test (`db.ToSQL` / dry run) and confirming it contains no `DISTINCT`

## 2. Backend: hidden cancelled count

- [x] 2.1 Add a downloads-specific response type (embedding `PaginatedResponse` fields plus `cancelled_hidden int64`) in `internal/api/dto.go` so other paginated endpoints are unchanged; verify other handlers' JSON is byte-identical in existing tests
- [x] 2.2 Compute `cancelled_hidden` when `status=active` (second `COUNT` with the same `type`/`search` predicates and `status = 'cancelled'`; apply `matchesProblem` on the Go path when `problem` is set), `0` otherwise; add tests for active with/without search, explicit `cancelled`, omitted status, and with `problem`
- [x] 2.3 Update `docs/` API documentation for the new `status=active`, `search` and `cancelled_hidden` contract and confirm the documented `curl` example returns the described fields

## 3. Frontend: state, API client and types

- [x] 3.1 In `frontend/src/hooks/useDownloads.ts`, change the `dlStatus` default to `'active'` (`'all'` for Tous, mapped to no `status` param), add `dlSearch` to the URL schema with a 300 ms debounced setter that resets `dlPage` to 1, and read `searchName` as a fallback alias when `dlSearch` is absent; extend `useDownloads.test.ts` for default, `dlStatus=all`, debounce, page reset and alias
- [x] 3.2 Extend `api.getDownloads` in `frontend/src/services/api.ts` with a `search` argument (URL-encoded) and surface `cancelled_hidden` in the returned type in `frontend/src/types.ts`; verify via the hook/API tests that the request URL contains `status=active&search=...`
- [x] 3.3 Expose `cancelledHidden` from `useDownloads` and wire the new props through `App.tsx` to `DownloadsTab`; verify `tsc` and existing `App` tests pass

## 4. Frontend: Downloads tab UI

- [x] 4.1 In `DownloadsTab.tsx`, update the status `<select>` options to Actifs / Tous / Complétés / En cours / En attente / Réessai / Échoués / Annulés with the mapped values, and add the search input; extend `DownloadsTab.test.tsx` for option values, default selection and typing in the search input
- [x] 4.2 Add the "N annulés masqués" indicator (shown only when status is `active` and the count is > 0; activating it sets the status to `cancelled` while keeping other filters); add tests for shown, hidden-at-zero, hidden-when-not-active, and activation behavior
- [x] 4.3 Add FR and EN strings (`statusFilter.active`, `pending`, `retrying`, `search.placeholder`, `cancelledHidden` with plural count) to `frontend/src/locales/{fr,en}/downloads.json`; verify the i18n key-parity test or lint passes and the UI shows localized text

## 5. Integration and release notes

- [x] 5.1 Run `go test ./internal/api/...` and the frontend test suite and confirm all pass
- [x] 5.2 Manually verify in the running app that `?tab=downloads&searchName=peacemaker` shows only non-cancelled Peacemaker downloads, the hidden counter appears, "Tous" lists cancelled rows last, and a refresh keeps each choice
- [x] 5.3 Note the behavioral default change (Downloads opens on "Actifs") in the project's change documentation, and verify the entry is present

## Workflow follow-up

- Archive the change after review requirements are satisfied.
- Verify the archived result: `openspec/specs/downloads-enrichment-api` and `openspec/specs/downloads-display-ui` reflect the deltas.
