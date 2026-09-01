## Why

The Radarr/Sonarr monitoring tab's occurrence tables (a matched movie's occurrences, and an expanded episode's occurrences) show a single "État" badge built from the raw pipeline `state` value (`pending`, `processed`, `downloading`, `organizing`, `downloaded`, `failed`). That value actually encodes two independent dimensions - processing and download - which the shared media detail drawer already separates into distinct "Traitement" / "Téléchargement" badges once opened. Collapsing them into one badge in the occurrence list hides which of the two is actually in progress or has failed (e.g. "processed" vs "downloading" looks like a single ambiguous state instead of "processing done, download in progress").

## What Changes

- Films section: the matched movie's "Occurrences" table gains a third column, splitting the single "État" column into "Traitement" (processing status) and "Téléchargement" (download status), computed from the existing `state` field via the same `getProcessingStatus`/`getDownloadStatus` helpers already used by the shared media detail drawer.
- Séries section: the identical split is applied to the nested occurrence table shown when an episode row is expanded.
- Reuse the existing `playlist` namespace vocabulary (`drawer.processingStatus`, `drawer.downloadStatus`, `pipelineStatus.notDownloaded`) for the two new column headers and badge text, instead of introducing new `radarrSonarr`-scoped strings, so the sidepanel's vocabulary matches the detail drawer opened from the same row.
- The episode list's own "État" column (Matched / No match, a playlist-match indicator unrelated to processing/download) is unchanged.
- No backend or API changes: the split is a pure frontend decomposition of the existing `OccurrenceResponse.state` field.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `radarr-sonarr-monitoring-view`: the "Sidepanel shows matched playlist occurrences for a selected item" requirement (and the occurrence rendering implied by "Selecting an occurrence opens the full media detail drawer" and "Séries episode rows expand to reveal their occurrences") is extended so that each occurrence row's pipeline state is shown as separate processing and download statuses rather than a single combined state.

## Impact

- `frontend/src/components/RadarrSonarrTab.tsx`: both occurrence tables (movie occurrences, expanded episode occurrences) gain a processing-status column and a download-status column in place of the single state column; imports `getProcessingStatus`, `getDownloadStatus`, `getProcessingStatusBadgeClass`, `getDownloadStatusBadgeClass` from `utils/pipelineState.ts` (replacing the now-unused `getPipelineStateBadgeClass` import) and reads the two new column labels via the existing `tPlaylist` translator already used in this file.
- No changes to `frontend/src/locales/*/radarrSonarr.json` (no new strings needed) or to any backend code.
