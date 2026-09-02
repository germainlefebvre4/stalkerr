## Why

The "Track Details" sidepanel only exposes a "Force Download" action, buried mid-panel in its own section. Correcting an item's TMDB match today requires leaving the sidepanel and using a separate row-level "Associate" button on the Playlist table (or, from the Radarr/Sonarr tab, isn't possible at all). Users reviewing a track's details have no way to fix a wrong/missing TMDB match without abandoning the panel they're already looking at.

## What Changes

- Add an "Associate" action to the Track Details sidepanel, triggering the same manual-override flow (TMDB search dialog) already used by the Playlist table's "Associate"/"Correct" button.
- Introduce a single action bar at the top of the sidepanel body (just under the header), holding both "Associate" and "Force Download" side by side.
- Move the existing "Force Download" button out of its own mid-panel section and into this action bar; its eligibility hints and status badges (ineligible reason, queued acknowledgement, error) move with it, directly under the action bar.
- Wire the Associate action through both places the sidepanel is used: the Playlist tab and the Radarr/Sonarr tab (desktop and mobile detail views), so the action is available regardless of entry point.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `m3u-playlist-details-sidepanel`: adds an "Associate" action to the sidepanel and restructures the sidepanel layout so "Associate" and "Force Download" live together in one action bar near the top, instead of "Force Download" sitting alone in a lower section.
- `tmdb-manual-override`: the manual-override modal can now also be opened from the Track Details sidepanel's "Associate" action (Playlist tab and Radarr/Sonarr tab), in addition to the existing Playlist table row button.
- `radarr-sonarr-monitoring-view`: the shared media detail drawer opened from this view also exposes the "Associate" action, consistent with the Playlist tab's drawer.

## Impact

- Frontend: `frontend/src/components/MediaOccurrenceDrawer.tsx` (layout restructure, new Associate button, new `onOpenOverride` prop), `frontend/src/components/PlaylistTab.tsx` (wire existing `onOpenOverride` prop into the drawer), `frontend/src/components/RadarrSonarrTab.tsx` (accept and forward an override handler to both its mobile and desktop drawer usages), `frontend/src/App.tsx` (pass `handleOpenOverride` down to `RadarrSonarrTab`).
- No backend changes: reuses the existing `POST /api/v1/items/:id/override` endpoint and `ManualOverrideDialog` component as-is.
- i18n: new translation keys for the sidepanel's "Associate" button label/title (English and French locale files).
