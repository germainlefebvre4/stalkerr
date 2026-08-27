## Why

The Manual Override modal always opens "blind": it never shows the item's current TMDB match (title, season, episode), and the "Saison"/"Épisode" fields stay hidden until a new search result is selected. This is true whether the item was matched automatically by the pipeline, corrected manually before, or corrected via the bulk-associate flow. Users reopening the dialog to check or fix an association have no way to see what it is currently matched to, or - for a bulk batch - what season/episode each checked candidate will actually receive, before confirming.

## What Changes

- Add a persistent "current match" panel in the modal, shown whenever `overrideItemData` already has an associated `movie` or `tvshow` (regardless of search/selection state): current TMDB title, year, and - for TV shows - season/episode, plus who/when it was last overridden (`override_by`/`override_at`) when set.
- Show the season/episode detected from the opened item's own raw title immediately when the modal opens, instead of only after a new TMDB result is selected.
- Extend each row of the bulk candidate list to also show that candidate's own current match (if already associated) and/or its own detected season/episode from its raw title, so the user can verify what a batch confirmation will actually apply before clicking "Forcer l'association".

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `tmdb-manual-override`: the "Frontend Interactive Manual Override Modal Dialog" requirement gains persistent current-match visibility (opened item and bulk candidates) and an unconditional detected season/episode preview.

## Impact

- Frontend: `frontend/src/components/ManualOverrideDialog.tsx` (current-match panel, unconditional detected-S/E preview, candidate row enrichment).
- i18n: new keys under the existing `manualOverride` block in `frontend/src/locales/{en,fr}/dialogs.json`.
- No backend/API changes: all data needed (current `movie`/`tvshow` association, `override_by`, `override_at`) is already present on the `PlaylistItem` objects the frontend already loads via the existing playlist listing endpoint.
- No database schema changes.
