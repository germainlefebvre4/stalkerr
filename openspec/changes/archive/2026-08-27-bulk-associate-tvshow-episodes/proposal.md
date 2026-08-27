## Why

Associating a `tvshows` playlist entry to its TMDB series today requires opening the Manual Override dialog and confirming the association one item at a time. When a whole season (or several seasons) of the same series needs the same correction, this becomes repetitive, and there is currently no way to apply one TMDB series choice to several playlist entries at once, nor to safely back out of an over-eager match.

## What Changes

- Extend the Manual Override dialog: when the target type is "Série TV" and a TMDB result has been selected, display the other `tvshows` entries from the currently loaded playlist page as a checkable candidate list.
- Pre-check candidates whose cleaned title (same cleaning already applied to the search query) matches the selected item's cleaned title; leave the rest unchecked. Every entry in the list remains freely checkable/uncheckable, so a false-positive suggestion can be deselected and an entry the heuristic missed can still be selected manually.
- On confirmation, apply the chosen TMDB series to every checked entry by calling the existing `POST /api/v1/items/:id/override` endpoint once per checked item (same `tmdb_id`/`type`, season/episode omitted so each item's own season/episode is auto-extracted from its own title as it already is today).
- Report per-item outcome once the batch completes (e.g. "11/12 associated", with the failing item identifiable) instead of a single pass/fail toast, since a multi-item batch can partially fail.
- No change to what a single override call does: pipeline `State` is left untouched, exactly as for today's one-at-a-time override.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `tmdb-manual-override`: the "Frontend Interactive Manual Override Modal Dialog" requirement gains candidate suggestion, multi-select, and sequential bulk-apply behavior for `tvshow` associations.

## Impact

- Frontend: `frontend/src/components/ManualOverrideDialog.tsx` (candidate list UI, batch submit loop, batch result reporting), `frontend/src/App.tsx` (pass the current `playlist` page data into the dialog).
- i18n: new keys under the existing `manualOverride` block in `frontend/src/locales/{en,fr}/dialogs.json`.
- No backend/API changes: reuses `POST /api/v1/items/:id/override` as-is, including its existing per-item season/episode auto-extraction and TMDB response caching.
- No database schema changes.
