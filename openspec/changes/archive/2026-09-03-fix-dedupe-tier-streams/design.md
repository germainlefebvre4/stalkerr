## Context

`BuildStreams` (`internal/scheduler/build.go`) assembles the full stream set for a run in five sequential steps, appending each step's output to a flat `[]*Stream`:

1. `buildTier1MovieStreams` - movies Radarr reports missing
2. `buildTier1SeriesStreams` - series-seasons Sonarr reports missing
3. `buildTier2MovieStreams` - movies with a `downloaded` `ProcessedLine` plus another eligible candidate
4. `buildTier2SeriesStreams` - series-seasons with a `downloaded` `ProcessedLine` plus another eligible candidate
5. `mergeIncompleteDownloads` - attaches in-progress/interrupted downloads to an existing stream from steps 1-4, or **synthesizes a brand-new tier-1 stream** when the owning movie/season wasn't already covered by step 1 or 2

Steps 3 and 4 run independently of steps 1 and 2 and of each other's `SourceKey`s, so the same `SourceKey` (`movie:<id>` or `series:tvdb:<id>:season:<n>`) can appear twice in the final set - once from a tier-1 step, once from a tier-2 step. See proposal.md for why that's a problem (duplicate download, destination-path overwrite).

## Goals / Non-Goals

**Goals:**
- Guarantee `BuildStreams` never returns two streams for the same `SourceKey` in one run.
- Cover streams synthesized by `mergeIncompleteDownloads`, not just the output of steps 1-4, since that step can itself introduce a fresh tier-1 `SourceKey` after tier-2 streams have already been built.

**Non-Goals:**
- Not changing tier-2 eligibility rules (which movies/seasons qualify for upgrade) - only suppressing a tier-2 stream when a tier-1 stream for the same `SourceKey` already exists this run.
- Not changing `NewScheduler`'s tier-selection probability logic or `ApplyLimit`'s work-unit limiting - the fix happens earlier, at `BuildStreams`, so every downstream consumer sees an already-deduplicated set.

## Decisions

**Where to filter: at the end of `BuildStreams`, after all tier-1 sources (including `mergeIncompleteDownloads`) are known.**
`mergeIncompleteDownloads` currently runs last specifically because it needs the maps built by steps 1 and 2 (`movieByID`, `tvdbSeasonStreams`) to decide whether to attach to an existing stream or synthesize a new one. Reordering it earlier would require larger surgery for no benefit. Instead, filter the tier-2 streams from steps 3 and 4 against `movieByID` and `tvdbSeasonStreams` *after* step 5 has finished mutating them (so any tier-1 stream synthesized by `mergeIncompleteDownloads` is included), then return. This keeps step ordering untouched and puts the guarantee in one place, right before `BuildStreams` returns.

**How to match: via the existing typed maps (`movieByID`, `tvdbSeasonStreams`), not via raw `SourceKey` string equality.**
Tier-1 and tier-2 `SourceKey` strings are not directly comparable for the series case: `buildTier1SeriesStreams` sets `SourceKey` to `series:<sonarrSeriesID>` and shares one key across every season of a series (needed for season-order queueing), while `buildTier2SeriesStreams` sets `SourceKey` to `series:tvdb:<tvdbID>:season:<season>`, one key per season, using Sonarr's *TVDB* ID rather than its internal series ID — two different ID spaces. A literal `SourceKey`-set membership check would silently never match for series, defeating the fix for exactly the case the proposal describes.
Instead, dedup on the typed identities `BuildStreams` already has in hand:
- Movies: `movieByID` (keyed by the local DB `Movie.ID`) — this key space is already identical to tier-2 movie streams' `SourceKey` (`movie:<id>`), so membership can be checked directly.
- Series: `tvdbSeasonStreams` (keyed by `tvdbSeasonKey{tvdbID, season}`) — this is exactly the `(tvdbID, season)` pair tier-2 encodes in its `SourceKey`. Format each key as `series:tvdb:<tvdbID>:season:<season>` (tier-2's own format) to build the comparison set, then check tier-2 streams' `SourceKey` against it.
Both maps are already mutated in place by `mergeIncompleteDownloads` when it synthesizes a new tier-1 stream, so reading them after step 5 gives the complete, correctly-typed tier-1 set with no new API calls, no `SourceKey` format changes, and no string parsing of tier-2 keys.

Alternatives considered:
- *Filter tier-2 immediately after building it (skip building tier-1 map lookups from steps 1-2 only)*: rejected - would miss the case where `mergeIncompleteDownloads` synthesizes a new tier-1 stream for a `SourceKey` that a tier-2 stream from step 3/4 already claimed, since that synthesis happens after tier-2 is built.
- *Dedupe downstream, in `NewScheduler` or `ApplyLimit`*: rejected - `ApplyLimit` is a no-op whenever `--limit` isn't passed (the common case), so it can't be the enforcement point; `NewScheduler` would need the same tier-1-wins logic duplicated away from where the streams are actually classified, and every other `BuildStreams` caller (e.g. dry-run) would still see the duplicate in its output.

**Precedence rule: tier-1 always wins.** If a `SourceKey` has both a tier-1 and a tier-2 candidate stream, the tier-1 stream is kept and the tier-2 stream is dropped entirely (not merged, not converted). Tier-1 already represents "still has missing content" for that work unit, which is a strict superset of concern compared to "already downloaded, upgrade eligible" - draining the tier-1 stream first is always correct, and a tier-2 upgrade opportunity that's still valid will reappear on a later run once the work unit is no longer classified tier-1.

## Risks / Trade-offs

- **[Risk]** A movie/season that is only *transiently* misclassified tier-1 (e.g., a momentary Radarr/Sonarr API inconsistency) has its tier-2 upgrade suppressed for that one run. → **Mitigation**: no data or opportunity is lost - the tier-2 candidate is simply re-evaluated on the next run; correctness (never overwriting a completed download mid-run) outweighs a one-run delay to an upgrade.
- **[Risk]** Silently dropping the tier-2 stream could look like "it just didn't get picked" from the logs. → **Mitigation**: none needed beyond what's already specced; this is an internal scheduling detail, not user-facing behavior that needs its own log line for this change.

## Migration Plan

None required: no data model change, no config surface change, no coordination with other services. The fix ships as a normal code change to `internal/scheduler/build.go` and takes effect on the next `download` run.
