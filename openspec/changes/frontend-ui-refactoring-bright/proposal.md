## Why

Currently, the Stalkeer dashboard uses a very basic styling theme and inline styling, which lack modern UI polish, refined animations, and high-quality visual feedback ("Bright Aura" theme). Furthermore, several powerful backend endpoints—such as global database statistics (`/api/v1/stats`), media pipeline reset capabilities (`/api/v1/movies/:id/reset` & `/api/v1/tvshows/:id/reset`), and filter configuration CRUD endpoints (`/api/v1/filters`)—are completely unexposed on the web interface, limiting the dashboard's capabilities and forcing manual database/CLI operations.

## What Changes

- Refactor the styling of the frontend using a refined "Bright Aura" light theme with modern indigo/violet color palettes, sleek card layouts, micro-interactions, smooth CSS transitions, and advanced animations (pulsing indicators, progress shimmers).
- Integrate global database statistics (`/api/v1/stats`) to render a horizontal grid of 4 highly styled visual KPI cards showing total playlist items, movies count, TV shows count, and download success ratios.
- Add "Reset Pipeline ↻" action buttons on playlist item rows to trigger stream resets on the backend, allowing users to easily re-queue media files for ingestion or downloading.
- Build a brand-new tab called "Filtres de Tri" implementing full CRUD management of filter configurations, allowing real-time viewing, creation, and deletion of custom exclusion/inclusion rules directly from the web browser.
- Update `vite.config.ts` proxies to correctly map `/health` and other API v1 endpoints during local development to avoid CORS or routing errors.

## Capabilities

### New Capabilities

- `frontend-filters-management`: A responsive interface to view, create, and delete custom filter configurations, mapping directly to backend REST endpoints.

### Modified Capabilities

- `frontend-ihm-dashboard`: Extend the playlist and monitoring dashboard with high-fidelity visual elements, automatic health indicators, statistics KPI cards, status filters, and pipeline reset actions.

## Impact

- **Affected folders/files**:
  - `frontend/src/variables.css`
  - `frontend/src/index.css`
  - `frontend/src/App.tsx`
  - `frontend/vite.config.ts`
- **APIs and Endpoints consumed**:
  - `GET /health` (via corrected Vite proxy)
  - `GET /api/v1/stats`
  - `GET /api/v1/filters`
  - `POST /api/v1/filters`
  - `DELETE /api/v1/filters/:id`
  - `POST /api/v1/movies/:id/reset`
  - `POST /api/v1/tvshows/:id/reset`
- **Dependencies**: React v19, Radix UI Primitives, CSS variables. No new library dependencies are required.
