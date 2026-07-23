## 1. Frontend Active Tab Persistence

- [x] 1.1 Update `frontend/src/App.tsx` to read the initial `activeTab` from `localStorage` ("stalkeer_active_tab") or default to "playlist".
- [x] 1.2 Add a `useEffect` hook in `frontend/src/App.tsx` to write the updated `activeTab` value to `localStorage` whenever it changes.

## 2. Compact Downloads Tab Refactoring

- [x] 2.1 Adjust `download-card` styling in `frontend/src/components/DownloadsTab.tsx` to have a smaller padding of `1rem` and smaller gaps of `0.5rem`.
- [x] 2.2 Refactor the `.file-info` container in `frontend/src/components/DownloadsTab.tsx` to strictly contain the folder and file name with a reduced padding of `0.5rem 0.75rem`.
- [x] 2.3 Extract technical values (`Format`, `Resolution`, `Size`) and display them on a clean, styled inline row below the `.file-info` box.
- [x] 2.4 Add support for `item.content.duration` (in minutes) s'il est présent, et l'afficher inline dans la même ligne technique avec l'icône `🕒`.
- [x] 2.5 Compute `isLowQuality` on the frontend (if resolution is 480p or 360p) and render a structured chip `⚠️ Basse qualité (<resolution>)` using `badge badge-pending`.
- [x] 2.6 Render validation status chips (Year validity: `✅ Année OK` or `⚠️ Année manquante`/`⚠️ Année incorrecte`, Format validity: `✅ Format OK` or `⚠️ Format inconnu`) as structured chips next to each other to the right of the technical specs line.

## 3. Verification & Validation

- [x] 3.1 Run `npm run build` inside the `frontend` folder to ensure compilation and linting succeed without error.
- [x] 3.2 Visually verify that the layout is responsive, compact, and that tab state is correctly restored on page reload.
