## 1. VFQ detection and persistence

- [x] 1.1 Add a `FrenchVariant *string` (varchar) column to `models.ProcessedLine`, relying on the existing `db.AutoMigrate` call in `internal/database/database.go` to create it — verify with a fresh local DB init that the column exists (e.g. `\d processed_lines` or an equivalent query)
- [x] 1.2 Add a `\bVFQ\b` detection regex in `internal/processor/processor.go`, scanned independently over `tvgName` then `groupTitle` (same fallback order as `detectLanguage`), and set `line.FrenchVariant` in `setContentType` alongside the existing `line.Language` assignment, without altering `detectLanguage`'s own `VF`/`MULTI`/`VOSTFR` matching
- [x] 1.3 Add processor unit tests covering: `VFQ` alone, `MULTI` + `VFQ` together (e.g. `Multi.Vfq.720P`), `VF` + `VFQ` together, and no `VFQ` marker — verify `line.Language` and `line.FrenchVariant` are set independently per `m3u-quality-selection`'s new ADDED requirement scenarios

## 2. Ranking tie-break

- [x] 2.1 Add a `frenchVariantRank` (or equivalent) helper in `internal/scheduler/rank.go` and fold it into `candidateRank` as a third key evaluated after language and resolution, matching the design's "tie-break only" decision
- [x] 2.2 Extend `matcher.resolutionOrderSQL` (`internal/matcher/matcher.go`) with the same VFQ tie-break as a trailing `ORDER BY` clause, keeping the existing language/resolution/recency ordering unchanged ahead of it
- [x] 2.3 Add/extend tests in `internal/scheduler/rank_test.go` (new) and `internal/matcher/matcher_test.go` proving: two candidates equal on language+resolution sort the non-VFQ one first, and the tie-break has no effect when language or resolution already differ — verify against `m3u-quality-selection`'s "France French preferred over Québec French" scenario

## 3. Shared filename tag builder

- [x] 3.1 Add a shared tag-building helper in `internal/downloader/destpath.go` (e.g. `qualityTags(resolution, language, frenchVariant *string) string`) producing `[resolution][language][VFQ]` with each bracket omitted when its input is nil/empty, per `download-filename-quality-tags`'s requirements
- [x] 3.2 Extend `BuildRadarrDestPathWithResolution`/`BuildSonarrDestPathWithResolution` to accept and append the language/VFQ tags via the new helper, keeping the existing `fallbackMarker` anti-collision behavior for unknown resolution unchanged (design Decision 4)
- [x] 3.3 Export the tag-building helper (`QualityTags`) so the automatic pipeline can append quality tags without the anti-collision fallback marker (design Decision 4). Implemented as an exported function rather than movie/series-specific `Build*WithTags` wrappers: `cmd/download.go`'s `downloadItem` only has `item.BaseDestDir` (an already-fully-built root path) at attempt time, not the raw `moviePath`/`title`/`year` components a `Build*` wrapper would need, so it appends `QualityTags(...)` directly to `BaseDestDir` (see task 4.3) instead of recomputing the root path per candidate
- [x] 3.4 Add/extend `internal/downloader/destpath_test.go` covering: resolution-only, language-only, resolution+language, resolution+language+VFQ, VFQ-only (no language), and all-nil (no tags at all, base path unchanged) for both the forced-download and automatic-pipeline builders

## 4. Automatic pipeline: tag the actually-downloaded candidate

- [x] 4.1 In `internal/scheduler/types.go`, change `Item.BaseDestPath` to `Item.BaseDestDir` (the untagged movie/season folder root), updating its doc comment to reflect the new per-candidate tagging step that now happens in `cmd/download.go`
- [x] 4.2 Update all four call sites in `internal/scheduler/build.go` (`buildTier1MovieStreams`, `buildTier1SeriesStreams`, `buildTier2MovieStreams`, `buildTier2SeriesStreams`) to populate `Item.BaseDestDir` from `BuildRadarrDestPath`/`BuildSonarrDestPath` as before, without a quality suffix
- [x] 4.3 In `cmd/download.go`'s `downloadItem`, compute each candidate's tagged destination path (`item.BaseDestDir` + tags from the new helper, using that candidate's own `Resolution`/`Language`/`FrenchVariant`) inside the existing `for j, candidate := range item.Candidates` loop, right before calling `dl.Download`, replacing the current single `item.BaseDestPath` usage
- [x] 4.4 Update `internal/scheduler/build_test.go` and `internal/scheduler/scheduler_test.go` assertions that reference `BaseDestPath` to reference `BaseDestDir` instead, verifying it stays untagged at build time
- [x] 4.5 Add/extend `cmd/download_test.go` coverage proving the on-disk filename for a completed automatic-pipeline download includes the succeeding candidate's resolution/language/VFQ tags, and that a failed-then-succeeded candidate sequence tags the file with the *succeeding* candidate's info, not the failed one's

## 5. Forced download: add language/VFQ tags

- [x] 5.1 Update `resolveForceDownloadMoviePath`/`resolveForceDownloadEpisodePath` in `internal/api/force_download.go` to pass `item.Language` and `item.FrenchVariant` into the extended `BuildRadarrDestPathWithResolution`/`BuildSonarrDestPathWithResolution` from task 3.2
- [x] 5.2 Update `internal/api/force_download_test.go` to cover a forced download whose occurrence has a known language and VFQ variant, asserting the resulting `download_path`/filename carries all three tags, per `force-download-media-occurrence`'s modified requirement

## 6. Resume path consistency

- [x] 6.1 Confirm (with a test in `internal/downloader/resume_helper_test.go` if not already covered) that `buildBaseDestPath`'s persisted-`DownloadPath` preference means a resumed download keeps whatever tags (or lack thereof) its original attempt was recorded with, and that its metadata-recomputation fallback path (`buildMovieBasePath`/`buildTVShowBasePath`) intentionally stays untagged, per design's "existing `DownloadInfo` records predate this change" risk note

## 7. Validation

- [x] 7.1 Run the full Go test suite (`go test ./...`) and confirm it passes
- [x] 7.2 Run `openspec validate tag-download-filenames-quality-lang --strict` and confirm it passes
