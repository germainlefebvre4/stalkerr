## Why

On the Radarr/Sonarr monitoring tab's Résumé sub-tab, the "Films (Radarr)" and "Séries (Sonarr)" summary cards show aggregate stats but give no way to jump to the corresponding detailed sub-tab (Radarr / Sonarr). Users currently have to notice and click the separate sub-tab trigger themselves, even though the card they're already looking at is the natural entry point.

## What Changes

- Make the "Films (Radarr)" summary card on the Résumé sub-tab clickable: clicking it switches the active sub-tab to Radarr.
- Make the "Séries (Sonarr)" summary card on the Résumé sub-tab clickable: clicking it switches the active sub-tab to Sonarr.
- Add a visual affordance (hover/cursor feedback) so both cards read as interactive, consistent with existing clickable-row styling elsewhere in the app.
- Scope is limited to the Résumé sub-tab's own cards on the Radarr/Sonarr monitoring tab; the similar-looking summary cards on the Home dashboard are out of scope and remain non-interactive.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `radarr-sonarr-monitoring-view`: the Résumé sub-tab's Radarr and Sonarr cards gain a new requirement to navigate to their respective detailed sub-tab when clicked.

## Impact

- Frontend only: `frontend/src/components/RadarrSonarrTab.tsx` (Résumé sub-tab card markup, wiring `onClick` to the existing `setActiveSubTab` from `useRadarrSonarrView()`).
- Styling: `frontend/src/index.css` (new clickable-card affordance, reusing the pattern already used by `.clickable-row`).
- No backend, API, or persistence changes; sub-tab selection already persists via the existing `?subtab=` URL state mechanism.
