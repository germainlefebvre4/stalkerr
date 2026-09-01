## 1. Derivation helpers

- [ ] 1.1 In `frontend/src/utils/pipelineState.ts`, add `getProcessingStatus(state: string): 'pending' | 'processed'` (`'pending'` only when `state === 'pending'`, `'processed'` otherwise) and verify with a quick unit test or manual check covering all six `state` values.
- [ ] 1.2 In the same file, add `getDownloadStatus(state: string): 'not_downloaded' | 'downloading' | 'organizing' | 'downloaded' | 'failed'` (`'not_downloaded'` for `pending`/`processed`, pass-through otherwise) and verify it against all six `state` values.
- [ ] 1.3 Add badge-class helpers `getProcessingStatusBadgeClass` and `getDownloadStatusBadgeClass` mapping the derived statuses to existing CSS classes (`badge-success`, `badge-progress`, `badge-failed`, `badge-pending`), and a new neutral class for `not_downloaded`.

## 2. Styling

- [ ] 2.1 In `frontend/src/index.css`, add a neutral/gray badge style for the `not_downloaded` status (or confirm `badge-pending`'s existing gray styling is visually distinct enough from `badge-success`/`badge-failed` to reuse as-is) and verify visually in the browser that "not downloaded" cannot be confused with "downloaded" or "failed".

## 3. Desktop table

- [ ] 3.1 In `frontend/src/components/PlaylistItemsTable.tsx` (desktop branch, `state` column), replace the single `getPipelineStateBadgeClass(item.state)` badge with two adjacent badges rendered from `getProcessingStatus`/`getDownloadStatus`, and verify a row whose item is `processed` with a `failed` download shows both badges simultaneously.

## 4. Mobile list card

- [ ] 4.1 In `frontend/src/components/PlaylistItemsTable.tsx` (mobile branch, `mobile-list-card`), replace the single text badge with two small icon indicators (one per status facet), each colored per its derived status and carrying a `title`/`aria-label` with the full status text, and verify in a mobile viewport that the card stays on one line and both icons are legible.
- [ ] 4.2 Verify accessibility by inspecting the rendered DOM (or using a screen reader) to confirm each icon exposes its full status text via `title`/`aria-label`.

## 5. Sidepanel

- [ ] 5.1 In `frontend/src/components/PlaylistTab.tsx`, in the "État du Pipeline d'Ingestion" grid (around the `currentStatus` cell), replace the single status cell with two cells - "Traitement" and "Téléchargement" - each showing a full-text badge derived from `getProcessingStatus`/`getDownloadStatus`.
- [ ] 5.2 Verify that an item with no download attempted (`getDownloadStatus` returns `not_downloaded`) shows an explicit `[ non téléchargé ]` gray badge in the sidepanel, rather than an empty or hidden cell.
- [ ] 5.3 Verify that an item processed successfully with a failed forced download shows `[ processed ]` and `[ failed ]` side by side in the sidepanel.

## 6. i18n

- [ ] 6.1 Add translation keys for the "Traitement"/"Téléchargement" cell labels and the "not downloaded" status label to `frontend/src/locales/fr/playlist.json` and `frontend/src/locales/en/playlist.json` (reusing existing `stateFilter.*` labels for `processed`/`pending`/`downloading`/`organizing`/`downloaded`/`failed` where applicable), and verify both locale files remain valid JSON and the UI renders correctly with each locale selected.

## 7. Verification

- [ ] 7.1 Manually walk through the three scenarios from `specs/playlist-pipeline-state-display/spec.md` and `specs/m3u-playlist-details-sidepanel/spec.md` (processed+failed download, processed+not-downloaded, desktop vs mobile rendering) in a running dev build, confirming each renders as specified.
