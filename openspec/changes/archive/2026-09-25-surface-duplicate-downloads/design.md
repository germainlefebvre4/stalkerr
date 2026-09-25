# Design

## Context

See `proposal.md` for motivation. Relevant existing pieces this design builds on:

- `internal/scheduler/build.go`'s `findDownloadedLine(db, column, id)` already queries `ProcessedLine` by `movie_id`/`tv_show_id` where `state = "downloaded"` — the same identity keys a duplicate-group query needs.
- `internal/api/resync_path.go` already establishes the "per-item on-demand action + scheduled report-only pass" shape this change reuses (its `file_missing` outcome is the closest precedent for "detected but not yet mutated").
- `internal/api/cancel_download.go`'s `cancelDownload` already transitions `download_info.status` and every `processed_lines.state` sharing that `download_info_id` to `cancelled` in one transaction — the exact target state this change reuses, but that handler explicitly refuses `completed` downloads (409), so it cannot be called directly; the new resolve action needs its own transition, following the same transaction shape.
- `internal/api/force_download.go`'s `persistForceDownloadPath` is the single place a forced `DownloadInfo` is created or reused — the one place that needs to write the new forced-origin marker.
- `cmd/download.go`'s `reconcileDownloadPaths` already runs once per scheduled `download` run and is where the existing per-item `file_missing` check lives; continuous duplicate detection hooks into this same function.

## Goals / Non-Goals

**Goals:**
- Reuse existing terminal states and action shapes (`cancelled`, the resync-path/cancel handler pattern) rather than introducing new ones.
- Keep continuous detection free of new external (Radarr/Sonarr) calls — it is a pure database query.
- Keep the resolve action's blast radius provably limited to one duplicate group.

**Non-Goals:**
- Backfilling the forced-origin marker onto `DownloadInfo` rows created before this change ships (see Risks).
- Any UI/API change to `cancel-download-occurrence` itself — the new resolve action is separate code, not an extension of that handler's eligibility rules.
- Deciding the exact wording/layout of the sidepanel's diagnostic section beyond what the spec requires (data to show, not visual design) — left for implementation/UI iteration.

## Decisions

### D1: Forced origin is a new boolean column, set once, at write time

`DownloadInfo` gains `IsForced bool` (`gorm:"default:false;not null"`). `persistForceDownloadPath` sets it to `true` on both the create branch and the reuse-existing-record branch. GORM `AutoMigrate` adds the column with its default on existing rows — no manual migration file, matching this project's existing schema-change convention.

**Alternative considered**: infer forced origin from the destination filename's resolution/marker suffix. Rejected — `download-filename-quality-tags` now applies resolution/language tags to automatic-pipeline downloads too, so the filename shape no longer reliably distinguishes the two paths. An explicit column is cheap and removes the ambiguity permanently.

### D2: One shared duplicate-group query, called from both the scheduled pass and the API

A single function (e.g. `FindDuplicateDownloadGroups(db) ([]DuplicateGroup, error)`) implements the Duplicate Group Definition:
1. Group `processed_lines` by `movie_id` (movies) and separately by `tv_show_id` (TV episodes — each `TVShow` row already represents one specific episode, consistent with `findDownloadedLine`'s existing key), filtering to rows whose `state = "downloaded"` and whose associated `download_info.status = "completed"`.
2. Exclude any row whose `download_info.is_forced = true` from the grouping entirely (an excluded forced row never counts toward a group, per the spec's exclusion requirement).
3. Within each remaining group of 2+, compare `download_info.download_path` values; a group only qualifies as a duplicate if at least two distinct path values are present.

This lives in a small new file, not `internal/scheduler` — `internal/api/handlers_frontend.go` already imports `internal/database`/`internal/models` directly and doesn't currently import `internal/scheduler`, and introducing that dependency only for this one query would be a heavier coupling than needed. Placing it alongside the models/database layer (e.g. `internal/database/duplicates.go`) keeps it importable from both `cmd/download.go` and `internal/api` without new cross-package edges.

No Radarr/Sonarr call is needed — this is pure local-database grouping, unlike path reconciliation.

**Alternative considered**: two separate queries (one for the scheduled pass, one for the API). Rejected — the Duplicate Group Definition is one rule; duplicating it risks the two call sites drifting out of sync the way the pre-fix title-sourcing bug did.

### D3: Continuous detection is report-only; the resolve action is the only path that mutates data

`reconcileDownloadPaths` (in `cmd/download.go`) calls the shared query once per run, after its existing per-item path/file checks, and logs each returned group as a distinguishable WARN (same shape as the existing `file_missing` log line: structured fields identifying the movie/episode and each sibling's id/path). It never calls the resolve logic itself. `GET /api/v1/downloads` calls the same shared query per request to compute the `duplicate_download` problem match and the new sibling-detail field, mirroring how the three existing `problem` values are already computed at request time (no caching layer exists for any of them today, so this is consistent, not a new pattern).

The resolve action lives in a new handler, `POST /api/v1/downloads/:id/resolve-duplicate` (new file `internal/api/duplicate_download.go`, same directory as `resync_path.go`/`cancel_download.go`): given one occurrence's `download_info` id, it looks up that occurrence's duplicate group via the shared query, then for each occurrence in the group does `os.Stat` on its `download_path` and, only for those whose file is missing, applies the same transaction shape `cancelDownload` uses (set `download_info.status = "cancelled"` and every associated `processed_lines.state = "cancelled"`). Returns per-occurrence outcome (`cancelled` / `unchanged`). If the requested id doesn't belong to any duplicate group, it responds without changing anything (`404`/`not_a_duplicate`, mirroring the existing `not_found`/`not_managed_by_radarr_sonarr` style used by `resync-path`).

### D4: Sibling detail on `DownloadEnrichedResponse`

A new optional field, e.g. `duplicate_siblings []DuplicateSiblingInfo` (`download_info_id`, `download_path`, `completed_at`), populated only for rows the shared query places in a group. Omitted (not null/empty array) when the row isn't part of a group, consistent with how `file_info`/`target_path`/`staging_path` are already omitted rather than null.

## Risks / Trade-offs

- **[Risk]** Forced downloads created before this change ships have no `is_forced` marker and will look automatic. If one of them happens to share a movie/episode with a genuine automatic duplicate, it will be included in the group instead of excluded. → **Mitigation**: narrow blast radius — this only affects the historical backlog existing at deploy time, self-corrects as each affected group is resolved via the new action (a resolve only cancels occurrences whose file is actually gone, so a mistakenly-included still-present forced download is left unchanged, not cancelled), and matches this project's established convention of not backfilling historical data (see the original `fix-duplicate-media-downloads` change's Non-Goals).
- **[Risk]** If every occurrence in a group has its file missing, resolving the group cancels all of them, leaving the movie/episode with no `downloaded` line — it becomes eligible for tier-1 again on the next scheduled run and may be re-downloaded. → **Mitigation**: this is the same re-triggering behavior `cancel-download-occurrence` already accepts for any cancelled occurrence; not special-cased here. If unwanted, the operator simply doesn't trigger the resolve action for that group.
- **[Risk]** Computing the duplicate-group query on every `GET /api/v1/downloads` request (not only when `problem=duplicate_download` is requested, since sibling detail must appear for any matching row in any view) adds one grouped query per request. → **Mitigation**: the query is a straightforward indexed `GROUP BY` over `movie_id`/`tv_show_id` restricted to `state = downloaded`, expected to be small relative to the full `processed_lines` table; consistent with the existing enrichment endpoint's per-request compute-everything approach, which already has no caching layer for any of the three existing problem values.

## Migration Plan

- Schema: `download_info.is_forced` added via `AutoMigrate`, defaults to `false` on existing rows (including historical forced downloads — see Risks). No manual migration file, no backfill.
- No feature flag: detection is additive (new `problem` value, new response field, new log line) and the resolve action is a new, explicitly-triggered endpoint — nothing existing changes behavior by default. Safe to deploy directly, consistent with the precedent set by `fix-duplicate-media-downloads`.
- No cleanup of the existing ~25/43 historical duplicate rows is performed by this change itself; they become visible and resolvable through the new UI/endpoint once deployed, which is the change's actual purpose.
