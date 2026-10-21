# Design

## Context

See `proposal.md` (Why / What Changes). Current state, confirmed in the code:

- `RadarrSonarrTab.tsx` renders the Films filter bar (`:432-472`) and the Séries one (`:560-600`) as two near-identical copies: one `flex-wrap` row holding the search `<input>`, an unlabelled native `<select>` (match-status, inline-styled), two `MultiSelectDropdown`s (Suivi, Fichier) and the "Manquants" `.btn-secondary`.
- `MultiSelectDropdown` triggers use `className="custom-select multi-select-trigger"`. `.custom-select` is shared with other forms and sets `width: 100%` and `padding: 0.65rem 1rem`; the native `<select>` uses inline `padding: 0.5rem 1rem` and a `--text-secondary` colour. In a `flex-wrap` row the `width: 100%` makes each trigger want a full line.
- `useRadarrSonarr` keeps the three filters per section in one `useURLState` instance. Each setter calls `patchFilterURLState(...)` then resets its page; `setFilmsMissing`/`setSeriesMissing` already patch two keys at once so a single fetch fires. Defaults: `filmsFilter` `''`, `filmsStatus` `monitored`, `filmsFile` `''` (same for series).
- The search term is not in the URL: it is local state (`filmsSearchInput`, debounced into `filmsSearch`).
- `useIsMobile()` exists (`hooks/useMediaQuery.ts`) and `RadarrSonarrTab` already uses it.

## Goals / Non-Goals

**Goals:**
- One shared, presentational filter bar used by Films and Séries, with a clear two-level layout and consistent control styling.
- A reset that returns to the section's default view in one URL patch.

**Non-Goals:**
- No change to filter semantics, defaults, URL keys, API, or to the Suivi/Fichier trigger text ("Suivi: Surveillé").
- No new dropdown/chip component; `MultiSelectDropdown` is reused as is.
- No change to the Résumé sub-tab or to tables, badges and sidepanels.

## Decisions

### 1. A presentational `ArrFilterBar` component, props-only
`ArrFilterBar` takes the values and callbacks the section already has (search input value/onChange/placeholder, match-status value/onChange, Suivi and Fichier sets/onChange/options, the Manquants active flag and handler, the reset handler and a `isDefault` flag, plus the group label for the service). It owns no filter state and does not import the hook.
**Why**: removes the Films/Séries copy-paste and keeps the hook/URL logic where it is. **Alternative**: a generic config-driven filter builder — rejected, two call sites do not justify it.

### 2. Layout: search line, then a row of labelled groups
```
[ Rechercher ........................................................ ]
Radarr                                   Playlist
[Suivi v] [Fichier v] [Manquants]  |     [Correspondance: Tous v]        [Reinitialiser]
```
The search sits alone on the first line (full width). The second line holds the Radarr (or Sonarr, per section) group and the Playlist group, each with a small caption above its controls, and "Réinitialiser" pushed to the end of the row. The row uses `flex-wrap` with a gap so narrow desktop widths wrap group by group, never control by control stretched to full width. Styling goes into CSS classes (`.arr-filter-bar`, `.arr-filter-group`, ...) in `index.css`, replacing the inline styles.

### 3. Fix the stretching locally, not by changing `.custom-select`
`.custom-select` is used outside this bar, so its `width: 100%` stays. The bar scopes `width: auto` on its triggers and gives every control (trigger, match-status select, Manquants, Réinitialiser) the same `min-height`/padding/font-weight through a shared `.arr-filter-control` class.
**Alternative**: edit `.custom-select` globally — rejected, risk of regressions on the configuration forms.

### 4. Match-status stays a native `<select>`, restyled and labelled
It is single-select with three values; a native control keeps the keyboard/mobile behavior and the existing tests. It gets the shared control style and a visible "Correspondance" caption, which fixes the unlabelled "Tous".
**Alternative**: render it with a Radix menu like the other two — rejected for this visual-only change: new interaction code for no behavior gain.

### 5. Mobile: full-width search, two-column grid
Below the existing mobile breakpoint the bar becomes a grid with two columns; the search spans both columns, group captions stay above their controls, and "Réinitialiser" spans the row. This is done with a CSS media query on the same breakpoint as `useIsMobile` rather than a JS branch, so the component has no mobile prop.

### 6. Reset: one URL patch per section, search untouched
The hook gains `resetFilmsFilters` and `resetSeriesFilters`, shaped like `setFilmsMissing`: one `patchFilterURLState({ <x>Filter: '', <x>Status: DEFAULT_SUIVI_STATUS, <x>File: '' })` followed by `setPage(1)`, so exactly one fetch fires. The search term is not reset: it is local component state outside the URL, it has its own line, and clearing it is one gesture.
`isDefault` is derived in `RadarrSonarrTab` from the current values (`filter === ''`, Suivi equal to exactly `{monitored}`, Fichier empty) and passed down; the button is `disabled` when true, which also avoids a pointless refetch.
**Alternative**: reset the search too — rejected, would require lifting `*SearchInput` or a ref callback and contradicts treating search as a query, not a filter. Easy to add later if wanted.

### 7. i18n
New keys in `radarrSonarr.json` (FR/EN) for the group captions (`filterBar.group.radarr`, `.sonarr`, `.playlist`), the match-status caption (`filterBar.correspondance`) and the reset button (`filterBar.reset`). Existing `filterStatus.*` keys keep their texts.

## Risks / Trade-offs

- [Existing `RadarrSonarrTab.test.tsx` queries may depend on DOM order or on the unlabelled select] → update tests to query by label/role; the new caption makes `getByLabelText` possible.
- [Group captions add vertical space] → small captions, one line; the search/filters were already two visual rows once wrapped.
- [A wide trigger text such as "Fichier: Téléchargé, Partiel, Absent" can still be long] → acceptable; triggers keep `white-space: nowrap` and the group wraps instead of squeezing.
- [Reset disabled when at defaults may confuse a user who sees Suivi = Surveillé as an active filter] → intentional: Surveillé is the default view, so there is nothing to reset to.

## Open Questions

- Visual confirmation of the `width: 100%` diagnosis in a browser is done during implementation (task 1.1); it does not change the approach, since the scoped `width: auto` is harmless either way.
