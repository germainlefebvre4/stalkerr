## Context

`RunItemsDialog` (the run's items dialog on the Logs/Processing page) already renders `PlaylistItemsTable` with genuine `PlaylistItem[]` data fetched via `api.getRunItems`, but never passes `onRowClick`, so rows are inert there today. The Playlist page wires `onRowClick={setSelectedItem}` on the same table and renders `MediaOccurrenceDrawer` from local `selectedItem` state (see proposal.md - Why).

`RunItemsDialog` and `MediaOccurrenceDrawer` are each an independent Radix `Dialog.Root`. `MediaOccurrenceDrawer` already exposes `withOverlay` and `modal` props for exactly this situation — `RadarrSonarrTab`'s mobile view sidesteps the same problem differently, by mounting `MediaOccurrenceDrawerBody` alone (no `Dialog.Root` at all) instead of the full drawer shell.

## Goals / Non-Goals

**Goals:**
- Clicking an item row inside the run's items dialog opens the same sidepanel content and behavior as on the Playlist page.
- No visible double overlay or competing focus trap when the sidepanel is open on top of the run's items dialog.

**Non-Goals:**
- Changing `MediaOccurrenceDrawer`'s content, props contract, or its usage on the Playlist or Radarr/Sonarr pages.
- Introducing a shared/global selection store — selection stays page-local state, consistent with every existing consumer of the drawer.

## Decisions

**Nest the full `MediaOccurrenceDrawer` dialog shell with `withOverlay={false}` and `modal={false}`, rather than embedding `MediaOccurrenceDrawerBody` inline.**

Alternatives considered:
- *Stack two fully modal `Dialog.Root`s unchanged (no prop changes)* — simplest diff, but Radix mounts a second overlay and a second focus trap, which fights the parent dialog for focus and visibly darkens the screen twice.
- *Embed `MediaOccurrenceDrawerBody` directly inside `RunItemsDialog`'s own content (the `RadarrSonarrTab` mobile pattern)* — avoids nested dialogs entirely, but the result is no longer literally "the same sidepanel" (no slide-in transition, no independent open/close affordance), which is the behavior being asked for.
- *Chosen: keep the full drawer shell but disable its own overlay/focus-trap via the props it already exposes* — preserves the drawer's visual identity and transition, while the parent `RunItemsDialog` remains the single active modal/focus scope. No new component surface is introduced; both props already exist on `MediaOccurrenceDrawer` for this reason.

**Selection state lives in `RunItemsDialog`, mirroring `PlaylistTab`.**

A local `const [selectedItem, setSelectedItem] = useState<PlaylistItem | null>(null)` inside `RunItemsDialog`, cleared via the drawer's `onOpenChange`, matches the existing pattern exactly and needs no new abstraction.

## Risks / Trade-offs

- [Disabling `modal`/`withOverlay` on the drawer could subtly change close-on-outside-click or escape-key behavior compared to the Playlist page's fully-modal drawer] → Manually verify closing via the drawer's own close button, Escape, and outside-click all still work and leave `RunItemsDialog` open, before considering this done.
- [`PlaylistItemsTable`'s per-row action buttons (association/correction, pipeline reset) already call `e.stopPropagation()` to avoid triggering `onRowClick` — this is existing behavior, not something this change introduces, but worth re-confirming still holds once `onRowClick` is wired in this new caller] → Covered by the existing "Correcting an item from within the run's item dialog" scenario; no new stopPropagation logic needed.
