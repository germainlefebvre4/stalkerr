## Context

See `proposal.md` - Why. Two independent pieces of existing code are involved:

- `frontend/src/hooks/useDownloads.ts` currently fetches on tab activation and every 5s (`setInterval(fetchDownloads, 5000)`), always with a hardcoded `limit=20` and no `offset` — it discards the `total`/`total_pages` fields the backend already returns.
- `internal/api/handlers_frontend.go`'s `listDownloadsEnriched` already supports `limit`/`offset` at the DB level (`query.Limit(limit).Offset(offset)`), but when `problem` is set it filters the *already-paginated* page in Go, after `Count()` has run — so `total`/`total_pages` reflect the status/type-matched count, not the problem-matched count. This was a deliberate original design choice: the problem-filter criteria (`has_year_in_path`, `year_mismatch`, `is_valid_format`, `detected_resolution`) come from parsing `download_path` at request time (`fileparser` package), not from stored columns, so they can't be pushed into the SQL `WHERE` clause without a schema change.

## Goals / Non-Goals

**Goals:**
- Remove background polling; the list only refreshes on tab activation, manual refresh, page change, filter change, or items-per-page change.
- Add page navigation and an items-per-page selector to the Downloads tab, reusing the existing `<Pagination />` component and the `limit`/`offset`/`total`/`total_pages` contract already returned by the API.
- Make `total`/`total_pages` accurate when `problem` is active, without a database migration.

**Non-Goals:**
- Persisting problem-filter criteria as DB columns for SQL-level filtering (rejected: larger scope, needs a migration + backfill + sync-on-enrichment logic; current data volume doesn't need it — see Risks).
- Changing polling/refresh behavior on any other tab (Playlist, etc.).
- Changing the set of available problem/status/type filter values, or sidepanel behavior.

## Decisions

**1. Drop `setInterval`, keep the existing manual-refresh wiring.**
`fetchDownloads` already backs both the initial `Promise.resolve().then(fetchDownloads)` call and the refresh button's `onClick`. Removing the `setInterval` effect (and its `clearInterval` cleanup) at `useDownloads.ts:39-44` is the entire change on the polling side — no new function needed, the refresh button keeps calling the same `fetchDownloads`.

**2. Reuse `<Pagination />` and the page/limit/total pattern already used by Playlist, not a bespoke control.**
`frontend/src/components/Pagination.tsx` is already generic (`total`, `page`, `setPage`, `limit`, `setLimit`, `limitOptions` props) and already used identically in `PlaylistTab.tsx` and `RunItemsDialog.tsx`. `useDownloads.ts` adopts the same shape those callers use: `downloadsPage`/`setDownloadsPage`, `downloadsLimit`/`setDownloadsLimit` (default 20, `limitOptions={[20, 50, 100]}` to keep the current default as the first/smallest option), `downloadsTotal` from the response. `api.getDownloads` gains an `offset` parameter following the exact `(page - 1) * limit` convention already used by `getRunItems`/`getGroupedPlaylist`.
Alternative considered: a bespoke prev/next-only control (simpler, less code) — rejected because it would diverge from the established pattern for no benefit, and users already expect `<Pagination />`'s page-jump/limit-select behavior from the Playlist tab.

**3. Page/limit reset to page 1 on any filter or limit change; page is not preserved across a manual refresh.**
Matches the existing (pre-change) documented behavior "resetting to page 1 on change" for filters, extended to the new `limit` control. A manual refresh click re-fetches the *current* page/limit/filters as-is (no reset) — refresh means "get fresh data for what I'm looking at", not "start over".

**4. Fix the `problem`-filter total by filtering before pagination in application code (Option A), not by persisting derived columns (Option B).**
When `problem` is set, `listDownloadsEnriched` fetches the full status/type-matched set (no `Limit`/`Offset` at the DB level), builds `file_info` for each row via the existing `fileparser` logic, applies the problem predicate, sets `total`/`total_pages` from the filtered slice's length, and then slices `[offset:offset+limit]` in Go before building the enriched response. When `problem` is not set, the existing DB-level `Limit`/`Offset`/`Count` path is unchanged.
Alternatives considered:
- **Option B (persist derived flags as DB columns + backfill + recompute on enrichment)**: correct and scales better, but is materially larger scope (migration, backfill job, keeping columns in sync whenever a download's file is renamed/parsed) for a filter that is one of several optional query parameters on an admin-facing tab. Deferred; revisit if downloads volume or query latency becomes a real problem.
- **Leave uncorrected, document only**: rejected per proposal decision — visible pagination (this change) makes the inconsistency user-facing (wrong last-page number, "total: 200" that never fills 200 items when filtered), unlike today where only page 1/limit 20 is ever shown.

## Risks / Trade-offs

- **[Risk]** The problem-filter path now loads the entire status/type-matched set into memory per request (instead of just one page) → **Mitigation**: bounded by `status`/`type` filters, which are typically applied together with `problem` in practice, and by realistic download-table sizes for a self-hosted single-user/small-team tracker; not bounded by the full unfiltered table. If this ever becomes measurable, Option B (persisted columns) is the escalation path, not a workaround here.
- **[Risk]** `total`/`total_pages` semantics change for any existing consumer of `GET /api/v1/downloads?problem=...` that relied on the old pre-filter count → **Mitigation**: the only frontend consumer is the Downloads tab itself (per `Impact` in proposal.md), and the new behavior is what a pagination UI needs to be usable; this is called out as a spec-level BREAKING change in the proposal.
- **[Trade-off]** Reusing `<Pagination />`'s `playlist` i18n namespace (`useTranslation('playlist')`, hardcoded inside the component) means Downloads-tab pagination labels are sourced from `locales/*/playlist.json` rather than `downloads.json` → accepted as-is, consistent with how `RunItemsDialog` already reuses the same component/namespace outside the Playlist tab; not worth a namespace refactor for this change.

## Migration Plan

No database migration. Deploy is a normal frontend + backend release:
1. Backend: `listDownloadsEnriched` change (behind no flag — `problem`-filtered requests immediately return corrected pagination metadata).
2. Frontend: `useDownloads.ts`/`api.ts`/`DownloadsTab.tsx` changes ship together (frontend already expects the `PaginatedResponse` shape the backend already returns, so no version skew concern between old frontend / new backend or vice versa — worst case with an old frontend hitting the new backend, it still only reads `data.data`, ignoring the now-correct `total`).
Rollback: revert both changes independently if needed; they are not interdependent (removing polling and fixing pagination-count can each be reverted alone without breaking the other).
