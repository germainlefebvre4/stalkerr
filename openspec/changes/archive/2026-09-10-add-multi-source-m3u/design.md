## Context

See proposal.md - Why. Today every layer of the M3U pipeline is single-source: `M3UConfig` (`internal/config/config.go`) has one `FilePath` and one `Download` block; `internal/m3udownloader` downloads/archives one file; `internal/parser` computes `LineHash = sha256(tvg_name+url)` and enforces a single global unique index on it; `internal/processor` has no notion of provenance; `cmd/m3u.go` / `cmd/process.go` each operate on one file; the Helm chart (`charts/stalkerr`) renders one `m3u` block into its ConfigMap and runs one `m3u-download` CronJob and one `process` CronJob (`scheduled-jobs`, `compose-scheduled-jobs` - unchanged by this design, since the fan-out happens inside the binary, not at the orchestration layer).

Movies/TV shows already converge across duplicate entries via `tmdb_id`-based `FirstOrCreate` (`internal/processor/processor.go`), and the existing `m3u-quality-selection` capability already orders every `ProcessedLine` candidate for a piece of content by language/resolution and falls back through them on download failure. Multi-source redundancy plugs into that existing machinery rather than replacing it - the design below is deliberately small.

## Goals / Non-Goals

**Goals:**
- Let one run of `m3u-download` / `process` handle N configured sources, each isolated on failure.
- Keep every source's streams as distinct, retained `ProcessedLine` candidates (never silently dropped as a "duplicate" of another source's stream).
- Zero-migration backward compatibility for existing single-source configs.

**Non-Goals:**
- Per-source priority/quality weighting in the download-fallback ordering (`source_name` is informational only - see `m3u-multi-source` spec, "Redundant candidates" requirement).
- TV channel support (channels are not yet grouped into deduplicated entities at all in the current codebase - a pre-existing gap, out of scope here).
- Any new frontend page, or surfacing `source_name` in the existing playlist item drawer (API/logs only for v1).
- Per-source scheduling (different cadence per source) or dynamic source management via API - sources are config-file-driven and take effect on the next CLI run, exactly like the current single-source config does.

## Decisions

### Config shape: `m3u.sources` list alongside the existing singular block
Add `Sources []M3USourceConfig` to `M3UConfig`, each entry shaped like today's singular fields plus a required `Name`. Resolution rule: if `Sources` is non-empty, it is used exclusively; otherwise the singular `FilePath`/`Download` fields are wrapped into one implicit source (name `"default"`). This is a plain "either/or", not a merge - simpler to reason about and document than mixing both, and it matches how the user already runs a single subscription today (no partial-migration state to support).

**Alternative considered**: merge singular config as an always-present extra source. Rejected - forces every existing user to think about a rename/interaction the moment they add a second source, for no real benefit.

### Provenance: flat `source_name` column, not a `sources` table
`ProcessedLine` gains `SourceName string` (`gorm:"type:varchar(100);not null;default:'default'"`). No new `Source` model/table.

**Rationale**: sources are managed exclusively via config.yml (per proposal scope - no management UI, no per-source enable/disable API). A `sources` table would exist purely to be joined against a string that already lives in config, with no CRUD surface to justify it. The `default:'default'` column default means GORM's `AutoMigrate` backfills existing rows in the same `ALTER TABLE ADD COLUMN` statement - no separate backfill step needed.

**Alternative considered**: a `sources` table with a `source_id` FK, enabling future per-source health/stats rows. Rejected for v1 as speculative - nothing in this change's scope reads or writes such metadata; can be introduced later without touching the spec above (`source_name` could migrate to `source_id` behind the same field name in the API/stats surface if ever needed).

### Dedup scope: composite unique index `(source_name, line_hash)`
Replace the existing single-column unique index on `line_hash` with a composite unique index on `(source_name, line_hash)`, and update the parser's in-memory `seenHashes` dedup (used for within-run duplicate skipping) to key on `source_name + line_hash` as well.

**Rationale**: this is the one change required to make "keep redundancy" safe rather than accidental. Today two sources happen not to collide because their URLs differ, but nothing guarantees that; scoping the constraint per source removes the landmine without weakening intra-source dedup (a true duplicate within one source's file is still rejected exactly as today).

### Per-source failure isolation in `m3u-download` / `process`
Both commands loop over the resolved source list. Each iteration's errors are caught and logged individually; the loop always continues to the next source. Final process exit code is non-zero if any source failed, computed only after every source has been attempted. This mirrors the existing single-source error-handling style (log + `os.Exit(1)`) but moves the exit decision to after the full loop.

**Rationale**: the whole point of "redundancy" is that one dead subscription shouldn't take the others down with it in the same scheduled run - fail-fast on the first source would defeat that.

### Per-source archive subdirectory
Each source downloads/archives under `<configured base path>/<source name>/...` instead of a bare file path. For the legacy implicit single source, the path is unchanged from today (no subdirectory inserted), so existing deployments see no on-disk layout change unless they opt into `m3u.sources`.

### Migration mechanics: extend the existing raw-SQL post-`AutoMigrate` step
`internal/database/database.go#runMigrations` already runs idempotent raw SQL after `AutoMigrate` (see the existing `DROP COLUMN IF EXISTS` statements for a precedent). Add, in the same style:
```sql
DROP INDEX IF EXISTS idx_processed_lines_hash;
DROP INDEX IF EXISTS idx_processed_lines_line_hash;
CREATE UNIQUE INDEX IF NOT EXISTS idx_processed_lines_source_hash ON processed_lines (source_name, line_hash);
```
The two `DROP INDEX IF EXISTS` statements defensively cover either index-naming convention GORM may have produced for the original bare `uniqueIndex` tag (the exact existing name should be confirmed against a running database with `\d processed_lines` before finalizing this statement, but the idempotent `IF EXISTS` form is safe either way). The column addition itself (`source_name` with a SQL-level default) is left to `AutoMigrate`, consistent with how every other `ProcessedLine` column is managed.

### Helm ConfigMap: additive `range` block
`charts/stalkerr/templates/configmap.yaml` gains a conditional `{{- if .Values.config.m3u.sources }}` block that ranges over the list and renders each entry's `file_path`/`download.*`, alongside (not replacing) the existing singular block's template lines. No new CronJob/PVC template - the existing `m3u-download`/`process` CronJobs already mount the whole M3U PVC and invoke the same two commands, which now internally loop over sources.

## Risks / Trade-offs

- **[Risk]** `AutoMigrate` never drops an index once created, so the schema change is not "free" the way a plain new column is. → **Mitigation**: explicit idempotent `DROP INDEX IF EXISTS` / `CREATE UNIQUE INDEX IF NOT EXISTS` in `runMigrations`, following the project's existing precedent for this kind of change.
- **[Risk]** Misconfiguring the same subscription twice under two different source names now bypasses dedup entirely (both copies kept as "redundant" candidates) instead of being silently deduplicated. → **Mitigation**: accepted trade-off of choosing redundancy over dedup, matches the explicit product decision; not something the system can distinguish from genuine redundancy.
- **[Risk]** Sequential per-source download/process increases total run time roughly linearly with source count. → **Mitigation**: acceptable at the expected scale (a handful of personal IPTV subscriptions); parallelizing sources is explicitly a non-goal for v1.
- **[Risk]** Multiple sources' credentials (`auth_username`/`auth_password`) continue to be rendered as plaintext in the Helm ConfigMap rather than the Secret - a pre-existing pattern for the single-source case, now multiplied by source count. → **Mitigation**: out of scope to fix here (would be a pre-existing-behavior change unrelated to this proposal), but worth the chart maintainer's awareness; flagged rather than silently carried forward unremarked.

## Migration Plan

1. Ship the `source_name` column + composite index migration in `runMigrations()` - safe to deploy ahead of the config/CLI changes, since a single-implicit-source deployment simply gets `source_name = "default"` on every row and behaves identically.
2. Ship the config/CLI/downloader/parser/processor changes together (they are mutually dependent).
3. Ship the Helm chart `configmap.yaml` addition last, as purely additive and opt-in.
4. Rollback: all three steps are additive/backward-compatible; reverting the binary while the migrated schema remains in place is safe (the extra column and composite index are harmless to an old binary that never reads/writes `source_name`, since Postgres still enforces uniqueness correctly for the `source_name = "default"` case single-source deployments produce).
