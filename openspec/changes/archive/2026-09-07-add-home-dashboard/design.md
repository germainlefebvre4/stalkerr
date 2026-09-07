## Context

See `proposal.md` - Why. Two technical constraints shape this design:

- **Dual database drivers**: `internal/testutil/helpers.go` runs the test suite against in-memory SQLite (`gorm.io/driver/sqlite`), while production uses `gorm.io/driver/postgres`. `internal/database/database.go`'s `runMigrations()` relies exclusively on GORM `AutoMigrate` (no hand-written SQL migration files, aside from two idempotent `ALTER TABLE ... DROP COLUMN IF EXISTS` statements). Any new `ProcessingLog` column must be a type `AutoMigrate` can create identically on both drivers.
- **Statistics already computed, not persisted**: `internal/processor/processor.go`'s `Statistics` struct already tracks `Movies`, `TVShows`, `TMDBMatched`, `TMDBNotFound` in memory during a run, and `saveBatch` (`processor.go:643-689`) already distinguishes a newly-created `ProcessedLine` (`gorm.ErrRecordNotFound` branch) from an updated one, across possibly-many `saveBatch` calls per run (one per full batch, plus a final partial batch — `processor.go:140-218`). Only `ItemCount` reaches `updateProcessingLog` (`processor.go:692-701`) today.
- **`processing-logs` has no separate DTO**: `listProcessingLogs` (`internal/api/handlers_frontend.go:27-58`) serializes `models.ProcessingLog` directly to JSON. New fields on the model are automatically part of the API response with no handler changes beyond the query itself.

## Goals / Non-Goals

**Goals:**
- Persist per-run statistics additively, without a destructive migration or a separate table.
- Keep the Home tab's data fetching composed from existing, already-polled hooks/endpoints wherever possible, rather than inventing a new aggregate "dashboard" backend endpoint.
- Make "no run yet" and "run predates this feature" clearly distinguishable from "run legitimately processed zero of something," per the specs.

**Non-Goals:**
- Backfilling statistics for `processing_logs` entries that predate this change (they remain permanently absent/null; see Open Questions).
- A history/trend view across multiple past runs — the Home tab shows only the *most recent* run, matching the proposal's scope.
- Real-time push updates for the Home tab; it follows the same polling pattern already used elsewhere in the app (`useHealthAndStats`'s 10s interval).
- Changing how items are attributed to a run (`processing_log_id`, capability `processing-run-items`) — this design only adds run-level aggregates alongside that existing mechanism.

## Decisions

### 1. New nullable columns on `ProcessingLog`, not a separate table
`MoviesCount`, `TVShowsCount`, `NewItemsCount`, `TMDBMatchedCount`, `TMDBUnmatchedCount` are added as `*int` fields (pointer, so `nil` on pre-migration rows, satisfying the "absent vs. zero" requirement). They are 1:1 with a run and small in number — a separate table would only add a join for no benefit.

**Alternative considered**: a `run_statistics` child table. Rejected — no scenario needs these fields independently of their parent run, and it would complicate the "absent for pre-migration rows" semantics (an outer join returning no row vs. a null column reads less naturally than the latter).

### 2. Group title list stored as a JSON-encoded `TEXT` column via a custom `sql.Scanner`/`driver.Valuer` type
Postgres' native `text[]` array type has no equivalent `AutoMigrate` can create on SQLite, which the test suite depends on. Instead, `GroupTitles` is declared as a small named type (e.g. `type StringList []string`) implementing `Scan`/`Value` to marshal to/from a JSON array stored in a `TEXT`/`jsonb`-agnostic column. This keeps:
- Storage portable across both drivers (a plain string column).
- The Go field and the JSON API response a real `[]string` / JSON array — callers never see the serialization detail, only `GroupTitles []string` (nil when absent, `[]` when empty per the "empty group title list" scenario).

**Alternative considered**: Postgres `jsonb` type directly (`gorm.io/datatypes`). Rejected for the same SQLite-portability reason — `datatypes.JSON` maps to `jsonb` on Postgres but needs driver-specific handling GORM doesn't uniformly resolve for SQLite; a hand-rolled `Scanner`/`Valuer` over `TEXT` avoids the extra dependency and works identically on both drivers, which is all this needs (no in-database querying of the list is required by any spec scenario).

### 3. Extend `Statistics` in place; accumulate across `saveBatch` calls
Add `NewItems int` and `GroupTitles map[string]struct{}` (a set, for the dedup requirement) to `processor.Statistics`. Increment `NewItems` in the `saveBatch` create branch (`processor.go:665-669`, alongside the existing `stats.Movies`/`stats.TVShows` increments at `processor.go:674-685`) and add each processed line's `GroupTitle` to the set in the same loop. `TMDBMatched`/`TMDBNotFound` already accumulate correctly across calls (`enrichMovie`/`enrichTVShow`); no change needed there beyond persisting the existing counters. `updateProcessingLog` converts the set to a sorted `[]string` and writes all new fields, mirroring how it already writes `ItemCount`.

**Alternative considered**: deriving these stats after the fact by querying `processed_lines WHERE processing_log_id = ?` (the "Option A" discussed during exploration). Rejected per explicit user decision (Option B): it would recompute at read time, cost more on runs touching many items, and "new item" cannot be reconstructed from `processing_log_id` alone (attribution updates on forced re-processing too, so a query-time derivation can't tell created from updated after the fact).

### 4. Home tab composes existing hooks/endpoints; downloads/errors counts use `limit=1` requests
The Home tab does not need a new aggregate backend endpoint. It reuses:
- `useHealthAndStats` (`/api/v1/stats`) for the catalog overview — already polled, no change.
- The existing `/api/v1/processing-logs?limit=1&offset=0` (no new query params) for the last-run summary — the newest entry is first by `Order("created_at desc")` already.
- `/api/v1/radarr-sonarr/stats` for the Radarr/Sonarr summary — same shape `useRadarrSonarr`'s stats fetch already consumes.
- `/api/v1/downloads?limit=1` and `/api/v1/downloads?problem=missing_year,year_mismatch,unknown_format&limit=1` for the downloads/errors *counts*: `PaginatedResponse.Total` is populated regardless of `limit`, so a `limit=1` request is enough to read a count without fetching the full item list. This avoids adding a dedicated counts endpoint.

Because the Home tab is now the default tab (mounted on first load, unlike today's tab-gated hooks such as `useDownloads(activeTab === 'downloads')`), these Home-tab fetches are gated on `activeTab === 'home'` the same way, not fetched unconditionally at the `App` level — consistent with how every other non-always-on tab already behaves.

**Alternative considered**: a single `GET /api/v1/dashboard` aggregate endpoint. Rejected for now — it would duplicate logic already living in four separate handlers and couples their evolution; the `limit=1` trick gets the same data with zero new backend surface. Worth reconsidering later only if Home-tab load latency (four to five parallel requests) becomes a measured problem.

### 5. Default tab and mobile nav wiring
- `VALID_TABS`, `TAB_URL_SCHEMA.tab.default`, and `readInitialActiveTab`'s fallback in `App.tsx` change from `'playlist'` to `'home'`.
- `MOBILE_FALLBACK_TAB` (`App.tsx:25`) changes from `'playlist'` to `'home'`, and the existing narrowing-viewport fallback `useEffect` (`App.tsx:149-153`, currently only checking `activeTab === 'errors'`) is extended to also check `activeTab === 'filters'`, mirroring the pattern already in place for `'errors'` rather than introducing a new mechanism.
- The mobile `tabs[]` array (`App.tsx:235-241`) drops the `filters` entry and gains a `home` entry first; the desktop `Tabs.List` gains a `home` trigger first and keeps the rest unchanged (including `filters`, which remains desktop-only-in-the-bottom-bar but still selectable from the desktop segmented tabs).

## Risks / Trade-offs

- **[Historical runs show "unavailable" statistics forever]** → Explicitly scoped as a non-goal; the Home tab renders a clear placeholder (per `frontend-home-dashboard` spec) rather than a misleading zero. A backfill script remains possible later (see Open Questions) without any spec or schema change.
- **[Group title list not queryable in SQL]** → Acceptable: no spec scenario requires filtering/searching by a run's group titles, only displaying them. If that need appears later, revisit storage (Decision 2's alternative).
- **[Home tab issues four to five parallel requests on every load]** → Same order of magnitude as today's `App.tsx` already issues on mount (stats, and whichever tab is active); mitigated by gating on `activeTab === 'home'` (Decision 4) so it doesn't add load when another tab is active, and by reusing already-polled hooks instead of adding new poll loops.
- **[Removing the KPI banner is a visible, immediate change for existing users]** → Called out as **BREAKING** in the proposal; mitigated by Home becoming the default landing tab, so the same numbers are the first thing users see, just relocated.
- **[Net tab count stays the same on mobile (5 before, 5 after: Playlist/Filters/Logs/Downloads/Arr-Suite -> Home/Playlist/Logs/Downloads/Arr-Suite)]** → No new overflow risk on narrow viewports; no additional responsive work needed for the bottom bar itself.

## Migration Plan

1. **Backend**: add the new `ProcessingLog` fields (Decisions 1-2), extend `Statistics` and `saveBatch`/`updateProcessingLog` (Decision 3). Deploy runs `AutoMigrate`, which adds the new nullable columns — additive and non-destructive; no manual SQL, no data backfill step required for correctness (absent is the intended state for old rows).
2. **Frontend**: ship the Home tab, nav changes, and KPI-banner removal together (single release) — this is an internal self-hosted dashboard with no staged-rollout/feature-flag precedent in the codebase (the earlier Erreurs-tab addition shipped the same way), so no flag is introduced here either.
3. **Rollback**: revert both commits. The added columns are harmless if left in place with old backend code (simply unused); no down-migration is needed given they're additive-only.

## Open Questions

- Should a one-off backfill script derive statistics for historical `processing_logs` rows from `processed_lines.processing_log_id` (capability `processing-run-items`), so "Last run" isn't permanently blank for installations upgrading from before this change? This doesn't affect the specs, the chosen approach, or the task breakdown for this change — it can be answered later as a separate, optional follow-up change.
