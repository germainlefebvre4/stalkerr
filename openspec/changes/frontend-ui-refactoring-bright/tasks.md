## 1. Config & Environment Setup

- [ ] 1.1 Update `frontend/vite.config.ts` proxies to include `/health` and `/api/v1/stats` to avoid CORS issues in local development.
- [ ] 1.2 Start the local dev server and verify health endpoint responds correctly via proxy.

## 2. Visual Styles Refactoring ("Bright Aura" Theme)

- [ ] 2.1 Overwrite `frontend/src/variables.css` with the new "Bright Aura" light theme variables (soft background, vibrant indigo-to-violet gradients, pure white cards, translucent borders, and soft shadows).
- [ ] 2.2 Update `frontend/src/index.css` to add support for active pulse animations on the health dot and gradient shimmers on progress bars.
- [ ] 2.3 Refactor segmented control tabs styles and modern table hover transitions in `frontend/src/index.css` or `App.tsx`.

## 3. Core Interface and API Integration

- [ ] 3.1 Fetch global database statistics from `GET /api/v1/stats` inside `App.tsx` on mount.
- [ ] 3.2 Implement and render a responsive grid of 4 gorgeous KPI cards underneath the header displaying total playlist items, movies, TV shows, and download success percentage.
- [ ] 3.3 Add "Réinitialiser ↻" buttons on playlist table rows for items in appropriate states.
- [ ] 3.4 Wire the reset button to trigger a `POST` request to `/api/v1/movies/:id/reset` or `/api/v1/tvshows/:id/reset` and refresh data upon success.
- [ ] 3.5 Refactor the main dashboard header to include the pulsing green health indicator dot.

## 4. Interactive Filters Tab & Modals

- [ ] 4.1 Build a brand-new tab called "Filtres de Tri" using Radix Tabs inside `App.tsx`.
- [ ] 4.2 Fetch active filter configurations from `GET /api/v1/filters` on mount and display them in a visually pleasing grid of clean card layout.
- [ ] 4.3 Add a "Supprimer 🗑️" button on each filter card to issue a `DELETE` request to `/api/v1/filters/:id` and update state.
- [ ] 4.4 Design and build a "Créer Filtre" dialog form using Radix UI `Dialog` primitives that collects filter names, attributes, inclusion, and exclusion patterns.
- [ ] 4.5 Wire the filter creation form to issue a `POST` request to `/api/v1/filters` and refresh the card list on success.

## 5. Testing & Verification

- [ ] 5.1 Run `npm run build` inside the `/frontend` directory to verify TypeScript compilations and bundling succeed without errors.
- [ ] 5.2 Validate that the user interface looks bright, pleasant, and modern, and is fully functional.
