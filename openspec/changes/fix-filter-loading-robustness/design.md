## Context

See proposal.md - Why. Current mechanics, for reference:

- `models.FilterConfig.IncludePatterns`/`ExcludePatterns` are `*string` columns documented as "JSON array" but every writer (`createFilter`, `updateFilter`, and the original `CreateFilterDialog.tsx`) has only ever written a raw comma-separated string (e.g. `"FRENCH, VFF"`), matching how the same field is displayed back to the user everywhere in the UI (`FiltersTab.tsx` renders `filter.include_patterns` directly as text).
- `filter.Manager.LoadFromDatabase()` is the only place that ever assumed the JSON-array format (`json.Unmarshal([]byte(*dbFilter.IncludePatterns), &includePatterns)`), so it has always failed on real data.
- `filter.Manager.LoadFromConfig()` loads `group_title` then `tvg_name` in sequence and returns immediately on the first `regexp.Compile` failure; `LoadAll()` returns that same error without ever calling `LoadFromDatabase()`.
- `internal/processor/processor.go`'s `NewProcessor()` treats any `LoadAll()` error as non-fatal (logs a warning, proceeds with whatever was loaded before the failure). `internal/dryrun/dryrun.go`'s `Analyze()` treats a `LoadFromConfig()` error as fatal (returns an error to the caller).

## Goals / Non-Goals

**Goals:**
- Make a runtime filter created through the existing API actually take effect during real M3U processing.
- Make one bad pattern (in config.yml or in a DB row) degrade only the specific attribute/row it belongs to, never take down unrelated filters.
- Catch invalid regex patterns at write time (API), not silently at read time deep in the pipeline.
- Make loading failures observable (attribute + pattern in the log) instead of a generic, scope-free warning.

**Non-Goals:**
- Not changing the comma-separated pattern convention itself (e.g., switching to a JSON array on the wire or in the UI) — the fix aligns the reader to the format every writer already uses, not the other way around.
- Not changing `Matches()`'s override semantics (single active runtime override per attribute, from `filter-override-policy`) — this change is about loading reliability, not override policy.
- Not adding a data migration — see Migration Plan below for why none is needed.
- Not changing `internal/dryrun/dryrun.go`'s or `internal/processor/processor.go`'s public interfaces; only the underlying `filter` package changes, so both callers keep working, and their existing error/non-error handling both become far less likely to trigger for the reasons described here.

## Decisions

**Fix the reader (`LoadFromDatabase`), not the writer.** The comma-separated string is what `createFilter`/`updateFilter` write, what the DB actually contains, and what the frontend displays as-is (`FiltersTab.tsx`, `CreateFilterDialog.tsx`'s "Load current config"). Changing the writer to emit JSON arrays would mean also changing the frontend's display/prefill logic and would leave any pre-existing (currently-broken, but still comma-formatted) row still broken. Fixing the reader to parse the comma-separated convention makes every writer — past and present — correct with no other code path touched.

**Introduce one shared parsing helper, `filter.ParsePatternList(raw string) []string`,** used by both `LoadFromDatabase()` (to turn a stored value into compiled patterns) and the new validation step in `createFilter`/`updateFilter` (to turn a submitted value into patterns to validate). Splits on `,`, trims whitespace per token, and drops empty tokens (e.g. from a trailing comma). One function is the single source of truth for "how a stored pattern list is interpreted," instead of validation and loading each re-deriving the split logic and risking drift.

**Per-attribute isolation happens inside `loadFilterSet`'s caller, not by changing `Matches()`.** `LoadFromConfig()` and `LoadFromDatabase()` each wrap their per-attribute (or per-row) `loadFilterSet` call: on a compile error, log it (attribute, raw pattern, error) at ERROR level via `internal/logger` and continue to the next attribute/row instead of returning early. Neither function returns an error for a pattern-compile failure anymore — only for conditions unrelated to pattern content (e.g., the existing "database not initialized" check in `LoadFromDatabase`). A skipped attribute simply has no filter loaded for it, which `Matches()` already treats as "no filters for this attribute → allow all" — the existing, well-understood permissive fallback, not a new code path.

**Validate patterns at `createFilter`/`updateFilter` using the existing `filter.ValidatePattern()`,** looped over `filter.ParsePatternList()` for both `include_patterns` and `exclude_patterns`. On the first invalid token, respond `400` with `ErrorResponse{Error: "invalid_pattern", Message: "<which pattern, which attribute>"}` and do not persist. This is deliberately the same regex engine (`regexp.Compile`, RE2) the processing pipeline itself will later use, so "valid at save time" reliably means "valid at load time."

**Alternative considered — reject the change to `LoadFromDatabase` and instead fix `createFilter`/`updateFilter` to JSON-encode on write:** rejected because it does nothing for any row already in the database under the comma-separated convention (there is no code path today that would have produced valid JSON), and it would require the frontend's display/prefill logic to be rewritten to consume the new shape, which is a strictly larger blast radius for the same outcome.

**Alternative considered — keep `LoadFromConfig`/`LoadAll` fail-fast, just with a better message:** rejected because a single typo in `config.yml` (as we found: `tvg_name.include_patterns: ["*"]`) would still silently disable the entire runtime-override system for every attribute — improving the message doesn't fix the blast radius, which is the actual problem.

## Risks / Trade-offs

- **[Risk] A pattern that used to be silently ignored (because the whole load failed) now silently falls back to "allow all" for just its own attribute, which is still a silent behavior change** → Mitigation: the new logging (attribute + pattern + error, at ERROR level) makes this visible in logs going forward, and pattern validation at creation time means new invalid patterns should never reach this path at all — only pre-existing bad data (like the `config.yml` typo) can still trigger it, and it's now confined to one attribute instead of everything.
- **[Risk] Existing `filter_configs` rows written before this fix, if any contain patterns that happen to also fail regex compilation (not just the format mismatch), will still fail to load** → Mitigation: they'll be logged per the per-row isolation above and skipped, same as any other invalid pattern; this is strictly better than today (total load failure) and requires no migration since it's handled at read time.
- **[Trade-off] Introducing `internal/logger` as a dependency of `internal/filter`** → acceptable; `logger` has no dependency back on `filter`, so no import cycle, and it's already the project's standard logging path (used by `processor`, `dryrun`, `api`, etc.).

## Migration Plan

No data migration. `LoadFromDatabase()`'s new parser (`filter.ParsePatternList`) reads the comma-separated format that has always been the actual on-disk format — every existing row, however it was created, becomes correctly parseable the moment this ships, with no backfill step. Rows containing patterns that are separately invalid as regex (not just previously misread) are logged and skipped, same as new invalid data would be — this is a behavior improvement, not a breaking read for any row that was ever going to work correctly.

Rollout is a standard deploy: no schema change, no new environment variable, no feature flag — the corrected behavior is on for every request as soon as the binary updates.
