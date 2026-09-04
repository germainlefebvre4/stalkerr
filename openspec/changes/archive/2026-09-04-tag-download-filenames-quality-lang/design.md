## Context

See proposal.md for motivation. Two existing mechanisms already compute destination paths and already carry resolution/language data per candidate:

- `internal/downloader/destpath.go`: `BuildRadarrDestPath`/`BuildSonarrDestPath` (used by the automatic pipeline, no quality suffix) and `BuildRadarrDestPathWithResolution`/`BuildSonarrDestPathWithResolution` (used only by forced download, resolution-only anti-collision suffix).
- `internal/models.ProcessedLine.Resolution`/`.Language` (`*string`), populated by `internal/processor.setContentType` from `internal/classifier.Classifier.ExtractResolution` and `internal/processor.detectLanguage`.

The automatic pipeline (`internal/scheduler/build.go`) computes each `Item.BaseDestPath` once, at stream-build time, before any of the item's `Candidates` have been attempted. `cmd/download.go`'s `downloadItem` then loops over `item.Candidates` (already quality-ordered) and uses that one fixed `BaseDestPath` for whichever candidate happens to succeed. Forced download (`internal/api/force_download.go`) is simpler: it always targets one known `ProcessedLine` occurrence, so its resolution/language are known before the path is built.

## Goals / Non-Goals

**Goals:**
- Tag the destination filename with the resolution/language/VFQ of the candidate that actually succeeded, for both trigger sources.
- Detect the `VFQ` marker as an independent signal from the existing `VF`/`MULTI`/`VOSTFR` marker.

**Non-Goals:**
- Renaming or migrating any file downloaded before this change.
- Deleting or cleaning up a superseded file when a tier-2 upgrade lands under a different tagged filename (per proposal decision: both files are left in place).
- Changing which candidate is chosen first (aside from the new VF-over-VFQ tie-break) — this is not a re-ranking of quality preference, only a naming and a narrow tie-break change.

## Decisions

### Decision 1: Move automatic-pipeline path finalization from stream-build time to per-candidate attempt time
`Item.BaseDestPath` (currently a single precomputed string) becomes an `Item.BaseDestDir` (the movie/season folder, i.e. what `BuildRadarrDestPath`/`BuildSonarrDestPath` return today) plus a per-candidate suffix computed inside `downloadItem`'s existing `for _, candidate := range item.Candidates` loop, right before each attempt — using that candidate's own `Resolution`/`Language`/VFQ fields. This mirrors what forced download already does per-occurrence, just moved one level down (per-candidate instead of per-item) for the automatic pipeline.

Alternative considered: precompute one tagged path per candidate at stream-build time (a `[]string` parallel to `Candidates`). Rejected — it duplicates path-building work for every candidate whether or not it's ever attempted, whereas building it lazily inside the existing attempt loop costs nothing extra and keeps `Build*DestPath` a pure function of an item's already-known root dir.

### Decision 2: `VFQ` is a separate boolean-ish signal, not a fourth `language` value
Extend `ProcessedLine` with a new nullable column (e.g. `FrenchVariant *string`, value `"VFQ"` or `NULL`) detected by its own regex scan (`\bVFQ\b`) over `tvgName`/`groupTitle`, independent of the existing `languageTokenRe` match for `VF`/`MULTI`/`VOSTFR`. `detectLanguage`'s existing leftmost-match behavior for `VF`/`MULTI`/`VOSTFR` is unchanged.

Alternative considered: add `"VFQ"` as a fifth value of the existing `language` field. Rejected — a title like `Multi.Vfq` needs to be recognized as both `MULTI` (drives the existing language-tier ranking and the `[MULTI]` filename tag) and Québec French (drives the `[VFQ]` filename tag and the new tie-break) at once; a single-valued field can't hold both without either losing the `MULTI` classification or inventing a combinatorial value space (`MULTI_VFQ`, `VF_VFQ`, ...) that `languageRank`/`resolutionOrderSQL` would need to special-case anyway.

### Decision 3: Ranking impact is a tie-break only, not a new tier
`candidateRank`/`resolutionOrderSQL` keep their existing `language`-then-`resolution` ordinal exactly as-is; VFQ is added as a third sort key evaluated only when language and resolution are already equal. This keeps the change additive to `m3u-quality-selection`'s existing ordering guarantees rather than risking a reshuffle of already-relied-upon VF/MULTI/VOSTFR/resolution precedence.

### Decision 4: Forced download keeps its resolution-only anti-collision fallback marker unchanged
`BuildRadarrDestPathWithResolution`/`BuildSonarrDestPathWithResolution` still fall back to a marker unique to the occurrence (its `ProcessedLine` id) when resolution is unknown — this exists to prevent two forced downloads of different nil-resolution occurrences of the same movie from colliding, a risk specific to forced download's arbitrary per-occurrence triggering. The new language/VFQ tags are appended alongside it, but don't replace it. The automatic pipeline has no equivalent risk (at most one candidate is ever downloaded per item per run, and tier-2 upgrades only fire for a strictly-better-ranked candidate) so it gets no such fallback marker.

## Risks / Trade-offs

- **Orphaned files after a tier-2 upgrade** → Accepted trade-off per proposal: the previous, differently-tagged file is left in place rather than deleted. Users doing cleanup will see e.g. both `Movie (2024) [720p][VF].mkv` and `Movie (2024) [1080p][MULTI].mkv` after an upgrade.
- **Existing `DownloadInfo.download_path` records for in-flight/incomplete downloads predate this change** → `resume_helper.go` already prefers a persisted `DownloadPath` over recomputing one, so an incomplete download started before this change resumes into its original (untagged) path rather than suddenly gaining tags mid-transfer. No special handling needed.

## Migration Plan

- Add the `FrenchVariant` (or equivalently named) nullable column to `processed_lines` via a new GORM auto-migration entry, consistent with how `Resolution`/`Language` were added.
- No backfill: the column is populated only for newly-processed M3U entries going forward; existing rows keep it `NULL`, and existing downloaded files are never touched (per proposal).
- No feature flag: this is a pure naming/detection addition with no rollback complexity beyond a standard revert.
