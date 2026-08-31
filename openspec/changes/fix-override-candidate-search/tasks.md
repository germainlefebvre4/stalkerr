## 1. Candidate fetch

- [ ] 1.1 Add a typed API helper in `frontend/src/services/api.ts` (or reuse `api.getPlaylist` directly) that fetches TV show items filtered by `content_type=tvshows` and `tvg_name=<cleaned title>` with a fixed generous limit, and verify it compiles and returns `PaginatedResponse<PlaylistItem>`.
- [ ] 1.2 In `ManualOverrideDialog.tsx`, replace the `useEffect` at lines 106-130 that filters the `playlist` prop with one that calls the new candidate fetch (keyed on `overrideItemData`, `overrideMediaType === 'tvshow'`, `selectedResult`), building the same `OverrideCandidate[]` shape (`preChecked`/`checked`/`preview`) from the fetch response instead of from `playlist`.
- [ ] 1.3 Remove the `playlist` prop from `ManualOverrideDialogProps` and its usage in `ManualOverrideDialog.tsx`.
- [ ] 1.4 Remove the `playlist={playlist}` wiring from the `<ManualOverrideDialog />` usage in `App.tsx`.

## 2. Verification

- [ ] 2.1 Manually verify in the running app: open the Processing page, click a completed run's row, open the override modal for a TV show episode whose sibling episodes are NOT on the Playlist tab's currently loaded page, select the correct TMDB show, and confirm the candidate list shows those sibling episodes pre-checked.
- [ ] 2.2 Manually verify the existing single-item override and full-batch-confirm flows (steps 10-13 of the `tmdb-manual-override` requirement) still work unchanged: submit succeeds, modal closes, playlist refreshes; partial-failure summary still reports per-item failures.
- [ ] 2.3 Run the frontend test suite (`npm test` or project's configured command) and confirm no regressions in existing `ManualOverrideDialog`/override-related tests.
