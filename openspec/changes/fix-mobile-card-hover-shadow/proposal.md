## Why

On mobile, tapping anywhere inside a tab's content area (any of Home, Downloads, Playlist, Filters, Logs, Errors, Radarr/Sonarr) makes a card-style drop shadow and a slight upward lift appear and stay stuck on screen until the user taps elsewhere. This comes from the `.card:hover` rule in `frontend/src/index.css`, which applies `box-shadow` and `transform: translateY(-2px)` on `:hover`. Touch browsers apply `:hover` on tap and only clear it on a later tap elsewhere ("sticky hover"), so the effect that is meant as a desktop mouse-hover affordance shows up as an unwanted, persistent shadow on mobile. This also violates the existing `frontend-responsive-layout` requirement that mobile tab content render with no box-shadow: `.card:hover`'s box-shadow (specificity 0,2,0) currently overrides the mobile `.tab-panel { box-shadow: none; }` reset (specificity 0,1,0) regardless of source order, so the "no box-shadow" rule silently fails to hold whenever the tap-triggered hover state is active.

## What Changes

- Scope the `.card:hover` lift/shadow effect (`frontend/src/index.css`) to pointer devices that support real hovering, using an `@media (hover: hover) and (pointer: fine)` guard, so it no longer triggers or sticks on touch taps.
- No visual change on desktop with a mouse: the hover lift/shadow still shows exactly as before.
- No change to the existing mobile full-bleed tab content styling (`.tab-panel` background/border/shadow removal) — this change makes that existing "no box-shadow on mobile" guarantee hold reliably even while a tap-triggered hover state is active, closing the gap between the two rules.
- Frontend-only, CSS-only: no component, markup, or API changes.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `frontend-responsive-layout`: the "Full-Bleed Tab Content on Mobile" requirement's "no box-shadow" guarantee is strengthened to explicitly hold even when a touch tap leaves an element in a stuck `:hover` state, not just in the element's default/static state.

## Impact

- `frontend/src/index.css`: the `.card:hover` rule (currently unconditional) gains a `hover`/`pointer` media guard.
