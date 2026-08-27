## 1. Shared season/episode extraction helper

- [x] 1.1 Extract the existing `S(\d+)[\s-]*E(\d+)` season/episode detection (currently inline in the `overrideItemData` effect of `ManualOverrideDialog.tsx`) into a small local helper function `extractSeasonEpisode(tvgName: string): { season: number | null; episode: number | null }`, and verify the existing single-item season/episode pre-population (Scenario: "Season and episode separated by a space in the raw title") still passes unchanged.

## 2. Current-match panel for the opened item

- [x] 2.1 Add a "current match" info block, rendered whenever `overrideItemData.tvshow || overrideItemData.movie` is set, showing TMDB title, year, and (for `tvshow`) season/episode, plus `override_by`/`override_at` when present; verify it renders for an item opened directly from a row that already has `movie`/`tvshow` populated, independent of `selectedResult`.
- [x] 2.2 Verify the panel stays visible and unchanged while the user types a new search query or selects a different TMDB result (it must not be gated by `selectedResult`).
- [x] 2.3 Add the new i18n keys for this panel's labels (e.g. current title/year/season-episode/override-by/override-at) to `frontend/src/locales/en/dialogs.json` and `frontend/src/locales/fr/dialogs.json` under the existing `manualOverride` block, and verify both locales render without missing-key fallbacks.

## 3. Unconditional detected season/episode preview

- [x] 3.1 Change the render condition of the "Saison"/"Épisode" input fields (currently `overrideMediaType === 'tvshow' && selectedResult`) to drop the `selectedResult` requirement so the fields show, pre-populated via the helper from task 1.1, as soon as `"tvshow"` mode is active for the opened item; verify the fields are visible and pre-filled before any search result is selected (Scenario: "Detected season/episode preview shown before any result is selected").
- [x] 3.2 Verify the values submitted by "Forcer l'association" for the opened item are unchanged by this visibility change (still read from the same `overrideSeason`/`overrideEpisode` state).

## 4. Bulk candidate row match-state preview

- [x] 4.1 In the `OverrideCandidate` construction (the `useEffect` building `candidates` from `playlist`), compute per candidate either its existing `tvshow` season/episode (if already associated) or the season/episode detected from `candidate.item.tvg_name` via the task 1.1 helper (if not), and store it on the candidate.
- [x] 4.2 Render this season/episode preview inline in each candidate row next to its raw title, distinguishing (via label/i18n text) "current" vs "detected" so the user can tell whether a candidate is already matched or will be matched based on title detection; verify against Scenarios "Bulk candidate row previews its own season/episode before confirming" and "Bulk candidate row shows its existing match when already associated".
- [x] 4.3 Verify a candidate with neither an existing association nor a detectable season/episode renders without a preview (no guessed value), consistent with the existing "leave empty rather than guess" rule.
- [x] 4.4 Add the new i18n keys for the candidate row preview labels to `frontend/src/locales/en/dialogs.json` and `frontend/src/locales/fr/dialogs.json`.

## 5. Regression check

- [x] 5.1 Re-run/verify the existing bulk-associate scenarios (pre-check heuristic, deselect false positive, manual add, partial batch failure) still hold with the new preview and panel present, since none of steps 10-13 of the requirement (submit payload, batch summary, refresh/close behavior) change in this proposal.
