## Context

See `proposal.md` - Why. Two existing facts drive this design:

- `internal/processor` discards lines that fail `filter.Manager.ShouldProcess` before they ever reach the database: `processed_lines` only ever contains what the *currently active* filter already let through. It cannot be used to preview a different pattern's effect on previously-excluded content.
- Every M3U source already keeps its most recently downloaded raw file on disk, independent of the database: `m3udownloader.SourcePaths(source.FilePath, source.Download.ArchiveDir, source.Name)` resolves that source's archive directory, and `m3udownloader.NewArchiveManager(archiveDir, log).GetLatestArchive()` returns the newest one. Re-parsing that file (`internal/parser`, pure, no DB/network access) gives the full, unfiltered catalog for free - no re-download needed, and no risk of hitting a slow or rate-limited playlist URL on every keystroke.

There is also a directly analogous existing feature to follow: `integration-connectivity-test` (`POST /api/v1/settings/integrations/test`) already establishes the "stateless, caller-supplied-values-only, never touches saved config" pattern this change needs, both on the backend (a throwaway check, `Breaker: nil`-style isolation from the real pipeline) and on the frontend (`SettingsGroupCard`'s `testState: idle/testing/ok/ko` + inline result). This change reuses that shape rather than inventing a new one.

An unrelated existing package, `internal/dryrun` (+ `POST /api/v1/dryrun`), already does something *called* "dry run" but is a different, unused feature: full pipeline analysis (parse + filter-from-config-only + classify + dedupe) of a server-side file path, with no support for caller-supplied candidate patterns and no frontend caller. It is not reused or modified by this change.

## Goals / Non-Goals

**Goals:**
- Preview a filter's effect against the true unfiltered catalog (matches and exclusions alike), not just against already-accepted database rows.
- Keep the check cheap enough to run on demand from the UI: no network fetch, bounded response size, single-pass parse.
- Let the same evaluation logic serve both the aggregate summary and the line-level content search described in the proposal.

**Non-Goals:**
- Triggering or scheduling a live download of the source. If no archive exists, the endpoint reports that; it does not fall back to fetching one.
- Reconciling or fixing the existing `internal/dryrun` package - it is left as-is.
- Changing how filters are actually applied during real ingestion (`filter.Manager`, `processor.go`) - this is a read-only preview path alongside it.
- Aggregating across multiple sources in one request - the user picks one source per test (per prior decision).

## Decisions

### Reuse `settings.EffectiveSources()` to resolve and validate `source_name`
`internal/settings.EffectiveSources()` already merges origin `config.yml` sources with runtime overrides by name - exactly the "does this source exist" check the endpoint needs for input validation, with no new resolution logic to write or keep in sync.

### New exported, Manager-independent pattern evaluator in `internal/filter`
Add `filter.CompilePatterns(includePatterns, excludePatterns []string) (*CompiledFilter, error)` and a `(*CompiledFilter) Matches(value string) bool`, extracting the existing exclude-then-include semantics currently private inside `Manager.Matches`/`loadFilterSet`. The dry-run handler compiles the caller-supplied patterns directly with this, entirely independent of `filter.NewManager()`, `LoadFromConfig`, or `LoadFromDatabase`. This guarantees the exact same match semantics as production filtering (single source of truth) while making it structurally impossible for a dry-run request to read or mutate the loaded Manager singleton used by real ingestion.

Alternative considered: instantiate a throwaway `filter.Manager` and call `LoadFromConfig`/inject patterns into it. Rejected - `Manager` is designed around loading *all* configured filters for every attribute from config/DB, not a single ad-hoc pattern set for one attribute; forcing caller-supplied values through it would need more workarounds than extracting the small matching primitive it's already built on.

### Parse the archive fresh on every request; no caching
Each dry-run request re-parses the source's latest archive file with `parser.NewParser(path, sourceName).Parse()`. Archive files are already local and parsing is a single linear pass (the same parse the real `process` pipeline already pays on every run), so an explicit cache is not worth the invalidation complexity for an on-demand, user-initiated action. If the archive is later found to be large enough to make this noticeably slow, that is worth revisiting, but is not assumed up front.

### Request accepts patterns the same shape the create dialog already edits
`CreateFilterDialog` holds include/exclude patterns as one free-typed, comma-separated string per attribute (`newFilterIncludes`/`newFilterExcludes`), matching how `FiltersSection` displays them (`joinPatterns` = `patterns.join(', ')`). The dry-run request takes `include_patterns`/`exclude_patterns` as that same single string per field and splits on `,` (trimming whitespace, dropping empties) before compiling - mirroring the form field exactly, so "what you typed" and "what got tested" are visibly the same text.

Note for awareness (not addressed by this change): the *save* path (`POST /api/v1/filters`) currently stores this same raw comma-joined string directly into `FilterConfig.IncludePatterns`, while `filter.Manager.LoadFromDatabase` reads it back expecting a JSON array (`json.Unmarshal`). That mismatch is pre-existing and orthogonal to dry-run testing (which never reads or writes `FilterConfig` rows); it is called out here only so it isn't mistaken for something this change should also fix.

### Content search reuses the same parsed line set and compiled patterns
Both response modes (aggregate summary, content search) run off one parse + one compiled pattern set per request; `search` only changes which projection of the results is returned (top-20-by-count distinct values vs. up to 100 matching individual lines). No second parse or duplicated matching logic.

## Risks / Trade-offs

- **[Risk] Large archives make every request pay a full parse + full-file scan.** → Mitigation: parsing is already proven at this scale by the real `process` pipeline running the same parser; the dry-run response itself stays small (bounded top-20/100-cap lists) regardless of archive size.
- **[Risk] No archive yet (fresh install, never downloaded) makes the feature unusable for a first-time filter.** → Mitigation: explicit, distinct "no archive available" response (per spec) rather than a generic error, so the frontend can point the user at running a download first instead of showing a confusing failure.
- **[Risk] A user relies on the dry-run result as a guarantee, then the live archive changes before they save.** → Mitigation: inherent to any preview-then-save flow; out of scope to solve here (no locking/versioning), but worth a UI note that the test reflects the last download, not necessarily what happens next run.

## Open Questions

None - the pattern-parsing convention question above is a note for future awareness, not something this change's specs, approach, or tasks depend on resolving.
