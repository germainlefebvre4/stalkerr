## 1. Wiring

- [x] 1.1 Add a `playlist: PlaylistItem[]` prop to `ManualOverrideDialogProps` in `frontend/src/components/ManualOverrideDialog.tsx`, and pass `playlist` into `<ManualOverrideDialog />` from `frontend/src/App.tsx`; verify the app still builds/typechecks (`npm run build` or `tsc --noEmit` in `frontend/`).

## 2. Candidate detection

- [x] 2.1 Extract the existing raw-title cleaning logic (currently inline in the `useEffect` at `ManualOverrideDialog.tsx:36-44`) into a small pure helper function usable on any raw title string, and use it for both the search query and candidate matching.
- [x] 2.2 Compute the candidate list whenever `overrideMediaType === 'tvshow'` and `selectedResult` is set: every `playlist` entry with `content_type === 'tvshows'` other than `overrideItemData`, each carrying a derived `preChecked` flag (cleaned title equals the opened item's cleaned title) and a mutable `checked` state initialized from `preChecked`. Verify with a quick manual check: opening the dialog on a `Breaking Bad S01E01` item on a page that also has `Breaking Bad S01E02`/`S02E01` and an unrelated `tvshows` entry shows the two matching episodes pre-checked and the unrelated one unchecked.

## 3. Candidate list UI

- [x] 3.1 Render the candidate list below the season/episode fields, only when `overrideMediaType === 'tvshow'` and `selectedResult` is set and at least one candidate exists: one row per candidate with its raw `tvg_name` and a checkbox reflecting/controlling its `checked` state.
- [x] 3.2 Verify checkbox interaction: clicking an unchecked candidate checks it and clicking a checked one unchecks it, independent of the `preChecked` heuristic (manual test: uncheck a pre-checked candidate, check an unchecked one, confirm both states persist until submit).

## 4. Batch submission

- [x] 4.1 Replace the body of `handleForceOverride` so that, on click, it builds the ordered list of target item ids: the opened item first, then every candidate with `checked === true`.
- [x] 4.2 For the opened item, keep sending the existing payload (`tmdb_id`, `type`, and the dialog's `season`/`episode` fields). For each checked candidate, send the same `tmdb_id`/`type` with `season: null, episode: null` so the backend auto-extracts them from that candidate's own title.
- [x] 4.3 Send the requests sequentially (`await` each `api.forceOverride(...)` in turn, not `Promise.all`), collecting a per-item success/failure result and continuing past individual failures instead of aborting the batch.
- [x] 4.4 While the batch is in flight, show per-item progress (e.g. "3/12...") in place of the current single `isSubmittingOverride` spinner state.
- [x] 4.5 After the batch completes: if every request succeeded, call `onSuccess` with an aggregated message (e.g. "12/12 associés") and close the dialog. If at least one failed, keep the dialog open, list which item(s) failed (by raw title) and the translated error for each, and still call `onSuccess`'s playlist-refresh behavior (or `fetchPlaylist` directly) so the successful subset is reflected immediately. Verify by temporarily forcing one candidate's request to fail (e.g. an invalid id) and confirming the dialog reports "N/M associés" with the failing title named, stays open, and the table shows the successful ones updated.

## 5. i18n

- [x] 5.1 Add new keys under the existing `manualOverride` block in `frontend/src/locales/en/dialogs.json` and `frontend/src/locales/fr/dialogs.json` for: candidate list section title, per-candidate checkbox label pattern, batch progress text, aggregated success message, and partial-failure message (naming the failing item). Verify both locale files remain valid JSON (e.g. `node -e "JSON.parse(require('fs').readFileSync('frontend/src/locales/en/dialogs.json'))"` for each file) and that switching the app language renders the new strings in both languages.

## 6. Manual verification

- [x] 6.1 Run through the full scenario from design.md end to end against a real playlist: open the dialog on one episode of a multi-season series, confirm candidates from the same page pre-select correctly, deselect a false positive, manually add a missed episode, submit, and confirm the resulting playlist rows all show the correct TV show association and that pipeline `state` is left untouched for every affected row.
