## 1. Films: split the movie occurrence table's state column

- [x] 1.1 In `frontend/src/components/RadarrSonarrTab.tsx`, replace the `getPipelineStateBadgeClass` import from `utils/pipelineState` with `getProcessingStatus`, `getDownloadStatus`, `getProcessingStatusBadgeClass`, `getDownloadStatusBadgeClass`; verify with `npx tsc --noEmit` (a leftover unused `getPipelineStateBadgeClass` import would fail the lint step in 3.1, not `tsc`).
- [x] 1.2 In the matched movie's "Occurrences" table (around line 340), replace the single `<th>{t('drawer.occurrenceState')}</th>` header with two headers: `{tPlaylist('drawer.processingStatus')}` and `{tPlaylist('drawer.downloadStatus')}`.
- [x] 1.3 Replace the single state `<td>` in that table's row with two `<td>` cells: one badge from `getProcessingStatusBadgeClass(getProcessingStatus(occ.state))` showing `getProcessingStatus(occ.state)`, and one badge from `getDownloadStatusBadgeClass(getDownloadStatus(occ.state))` showing `getDownloadStatus(occ.state) === 'not_downloaded' ? tPlaylist('pipelineStatus.notDownloaded') : getDownloadStatus(occ.state)` - mirroring `MediaOccurrenceDrawerBody`'s Pipeline State section exactly; verify with `npx tsc --noEmit`.

## 2. Séries: split the expanded episode's occurrence table's state column

- [x] 2.1 In the nested occurrence table shown when an episode row is expanded (around line 408), apply the identical two-header / two-badge change from 1.2-1.3 to this table's header row and `occ` mapping. Leave the episode list's own "État" column (Matched/No match, around line 379) untouched.

## 3. Verification

- [x] 3.1 Run `npx tsc --noEmit`, `npm run lint`, and `npm test` in `frontend/` and confirm all green, excluding the pre-existing unrelated `react-hooks/set-state-in-effect` lint errors already present on `main` (`DownloadsTab.tsx`, `ManualOverrideDialog.tsx`, `MediaOccurrenceDrawer.tsx`).
- [x] 3.2 Manually launch the app, open a matched movie with at least one occurrence, and confirm the occurrence row shows two distinct badges (processing, download) with vocabulary matching the detail drawer opened from the same row (e.g. "Processed" / "Downloading", or "Not Downloaded" when applicable); expand a matched episode and confirm its nested occurrence row shows the same two-badge split; confirm the episode list's own Matched/No match column is unchanged.
