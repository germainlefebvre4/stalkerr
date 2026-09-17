## 1. Résumé cards navigation

- [x] 1.1 In `frontend/src/components/RadarrSonarrTab.tsx`, wire the Films (Radarr) card on the Résumé sub-tab to call `setActiveSubTab('radarr')` on click, and verify clicking it switches the active sub-tab to Radarr (URL `?subtab=radarr`).
- [x] 1.2 Wire the Séries (Sonarr) card on the Résumé sub-tab to call `setActiveSubTab('sonarr')` on click, and verify clicking it switches the active sub-tab to Sonarr (URL `?subtab=sonarr`).
- [x] 1.3 Add a clickable-card affordance style (cursor pointer + hover feedback, reusing the pattern from `.clickable-row` in `frontend/src/index.css`) to both cards, and verify visually that hovering either card shows the affordance.

## 2. Verification

- [x] 2.1 Manually test in the running app: from the Résumé sub-tab, click each card and confirm it lands on the correct detailed sub-tab with its content visible; confirm the Home dashboard's similar-looking cards remain non-interactive (out of scope).
