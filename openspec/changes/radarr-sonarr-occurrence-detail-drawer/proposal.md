## Why

The Radarr/Sonarr monitoring tab's sidepanel currently lists playlist occurrences as bare `resolution | state` rows (or, for series, per-episode `matched`/`no match` rows with no occurrence detail at all) - there is no way to tell which specific M3U entry a row corresponds to, or to see its TMDB metadata, raw line, stream URL, or trigger a force-download for it. The Playlist tab already has a full media detail drawer that does exactly this; reusing it here removes the guesswork.

## What Changes

- Extract the Playlist tab's existing item-detail drawer (TMDB metadata, pipeline state, force-download action, M3U provenance, raw line/URL) into a shared component driven by a `PlaylistItem`/`ItemResponse`, so both the Playlist tab and the Radarr/Sonarr tab render the identical drawer.
- Add a frontend API client method for the existing (unchanged) `GET /api/v1/items/:id` endpoint, and use it to fetch the full item when an occurrence row is clicked.
- Films section: each occurrence row in the movie sidepanel becomes clickable and opens the shared media detail drawer for that occurrence, including the force-download action.
- Séries section: each monitored-episode row becomes expandable, revealing its own occurrence rows (mirroring the Films occurrence list) before drilling into the shared detail drawer.
- Desktop layout: the media detail drawer opens as a second panel immediately to the left of the existing occurrences sidepanel (both visible at once, sharing one overlay) rather than replacing it.
- Mobile layout: the media detail view replaces the occurrences sidepanel's content in place (no side-by-side panels), with a way back to the occurrence list.
- Reformat the season/episode label from `S1E1` to `S01 E01` (zero-padded, space-separated) in the Séries sidepanel.
- Supersede the "no write actions from this view" non-goal recorded in the archived `radarr-sonarr-monitoring-view` change's design: force-download is now intentionally reachable from this view via the reused drawer, since the view now surfaces the exact occurrence a user would want to act on.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `radarr-sonarr-monitoring-view`: the "Sidepanel shows matched playlist occurrences for a selected item" requirement is extended so that selecting an occurrence (movie or, after expanding an episode, series) opens the full shared media detail drawer - stacked beside the occurrences sidepanel on desktop, replacing it on mobile - instead of only ever showing a bare resolution/state row. A new requirement covers series episode rows expanding to reveal their own occurrences before that drill-down, including the `S01 E01` label format.

## Impact

- `frontend/src/components/RadarrSonarrTab.tsx`: occurrence rows (movies) and episode rows (series) become interactive; episode rows gain expand/collapse state; selecting an occurrence opens the shared drawer.
- `frontend/src/components/PlaylistTab.tsx`: its inline item-detail drawer JSX is extracted into a new shared component (no behavior change for the Playlist tab itself).
- New shared component (e.g. `frontend/src/components/MediaOccurrenceDrawer.tsx`) encapsulating the extracted drawer, taking a `PlaylistItem`/`ItemResponse`-shaped item and an `onOpenChange` handler; owns its own copy-to-clipboard and force-download state.
- `frontend/src/services/api.ts`: new `getItem(id)` method wrapping the existing `GET /api/v1/items/:id` endpoint (no backend changes).
- `frontend/src/index.css`: new drawer positioning variant for the secondary (left-hand) panel on desktop, and mobile in-place replacement behavior.
- `frontend/src/locales/{fr,en}/radarrSonarr.json`: updated/added strings for the expandable episode rows and the drill-down entry points.
- No backend changes: `GET /api/v1/items/:id` (`internal/api/handlers.go`) is reused as-is.
