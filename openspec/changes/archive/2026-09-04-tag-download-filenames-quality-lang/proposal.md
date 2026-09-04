## Why

Downloaded movie and episode files currently land on disk with a bare title (e.g. `Inside Out 2 (2024).mkv`), with no indication of which resolution or audio language was actually fetched. Users can't tell, from the filename alone, whether they got a 720p or 4K file, or whether the French audio track is France French or Québec French (`VFQ`) — a distinction the source catalog's own titles already carry (e.g. `Multi.Vfq.720P`) but that today's pipeline discards during matching.

## What Changes

- New destination-naming rule: when a download completes, the on-disk filename SHALL include the actually-downloaded candidate's resolution and language, in standard bracket tags (e.g. `Movie Title (2024) [1080p][MULTI][VFQ].mkv`), for whichever of those is known. A tag is simply omitted when that piece of information isn't available.
- This applies uniformly to both the automatic `download` pipeline (Radarr/Sonarr missing-content + tier-2 upgrades) and to a forced download triggered from the UI. Forced download's existing resolution-only anti-collision suffix is superseded by this shared tagging rule.
- New detection: a French audio track that is specifically Québec French (`VFQ`) is now recognized as its own signal, independent of (and combinable with) the existing `VF`/`MULTI`/`VOSTFR` language marker, since a title can carry both (e.g. `Multi.Vfq`).
- New tie-break rule for candidate/upgrade ranking: when two candidates are otherwise equally ranked, a candidate without the Québec French variant is preferred over one with it (France French preferred over Québec French).
- Already-downloaded files are unaffected: this naming rule applies only to downloads that happen from this change forward. No existing file is renamed or re-downloaded to comply.

## Capabilities

### New Capabilities
- `download-filename-quality-tags`: Defines the shared rule for tagging a downloaded file's name with the resolution and language (including the Québec French variant) of the candidate that was actually downloaded, applied consistently across the automatic pipeline and forced downloads.

### Modified Capabilities
- `m3u-quality-selection`: Adds detection and persistence of the Québec French (`VFQ`) variant as a signal independent of the existing `VF`/`MULTI`/`VOSTFR` language marker, and adds a secondary ranking tie-break (France French preferred over Québec French) to the existing candidate ordering.
- `force-download-media-occurrence`: Updates the forced-download destination-naming requirement so the language (and Québec French variant, when detected) is included alongside the existing resolution anti-collision suffix, and removes the now-outdated guarantee that automatic-pipeline naming stays unaffected.

## Impact

- `internal/classifier` / `internal/processor`: language/variant detection regexes.
- `internal/models`: new column on `ProcessedLine` for the French-variant signal.
- `internal/scheduler/rank.go`: new tie-break in candidate ranking.
- `internal/downloader/destpath.go`: destination-path builders extended with a language(+variant) tag alongside the existing resolution tag.
- `internal/scheduler/build.go` and `cmd/download.go`: destination path for the automatic pipeline must be finalized once the actually-attempted candidate is known, rather than once per item before any candidate is tried.
- `internal/api/force_download.go`: pass language/variant into the existing resolution-suffixed path builder.
- A new database migration for the added `ProcessedLine` column.
