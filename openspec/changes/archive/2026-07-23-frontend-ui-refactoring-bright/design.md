## Context

The current Stalkeer frontend provides basic status visibility but suffers from plain styling, inline-heavy styling blocks, and lacks several key administrative features. To deliver a visually delightful and functionally exhaustive administration dashboard, we are executing a complete visual refactoring under the "Bright Aura" light theme. We will also integrate previously unused backend REST endpoints—such as API stats, item reset actions, and filter configuration CRUD.

## Goals / Non-Goals

**Goals:**
- Implement a stunning, modern light theme ("Bright Aura") using a pure CSS variables approach with standard CSS files, adhering strictly to the "Vanilla CSS" mandate.
- Deliver visual polish including capsule-style tab lists (segmented controls), translucent borders, responsive layouts, micro-interactions (hover transitions, card elevations), and smooth animations (health pulse, progress bar shimmer).
- Fully expose and integrate backend statistics (`GET /api/v1/stats`) to display dashboard KPI counters.
- Expose pipeline reset endpoints (`POST /api/v1/movies/:id/reset` & `POST /api/v1/tvshows/:id/reset`) through intuitive row-level UI buttons.
- Deliver a brand-new, fully functional "Filtres de Tri" tab to create, list, and delete custom backend filters (`/api/v1/filters` GET/POST/DELETE) with accessible Radix UI dialog forms.
- Fix local development proxy routing in `vite.config.ts`.

**Non-Goals:**
- Introducing heavy UI frameworks like TailwindCSS or bulky CSS-in-JS libraries.
- Rewriting the Go backend router or data schemas (leveraging existing robust controllers and model structs).
- Adding complex front-end client-side routers (maintaining the high-performance and lightweight single-page layout).

## Decisions

### 1. Refactoring CSS variables into "Bright Aura" Light Theme
- **Rationale**: Utilizing a centralized, modern CSS variables system in `variables.css` combined with scoped animations in `index.css` is clean, lightweight, highly readable, and conforms with the workspace's existing structure.
- **Alternatives Considered**: Installing TailwindCSS. Rejected because Tailwind adds significant dependency footprint, requires build-time compilation configurations, and contradicts the local directive to avoid Tailwind unless explicitly requested.
- **Implementation**: Redefine CSS variables in `variables.css` to use soft backgrounds (`#f4f6fc`), clean pure-white card surfaces, an indigo-to-violet gradient accent, refined pastel status indicators, and large modern border-radii (`--radius-lg: 18px`).

### 2. Segmented Capsule Controls for Tabs
- **Rationale**: Segmented controls (capsule lists with sliding active backgrounds) provide a vastly superior, modern tactile feel compared to simple underlines.
- **Implementation**: Style Radix Tabs list into a capsule background (`#e2e8f0` / `#f1f5f9` with `--radius-md`), where the active tab trigger is rendered as a clean elevated white card with a subtle shadow and 0.2s ease-in-out transition.

### 3. Integrating `api/v1/stats` for Dashboard KPI Cards
- **Rationale**: Visualizing global statistics immediately captures the user's attention and provides immediate operational state clarity (number of items, movies vs tvshows ratio, download success metrics).
- **Implementation**: Fetch statistics on mount (and in tandem with tabs polling). Display a responsive grid of 4 cards. Use a gorgeous custom icon or styled typography for each KPI block.
- **Data Shape**:
  ```json
  {
    "total_items": 3480,
    "by_content_type": { "movies": 1120, "tvshows": 940, "channels": 100, "uncategorized": 1320 },
    "by_state": { "processed": 3000, "pending": 50, "downloading": 10, "downloaded": 418, "failed": 2 }
  }
  ```

### 4. Interactive Filter Administration ("Filtres de Tri" Tab)
- **Rationale**: Managing filters through the web UI eliminates the need for manual DB entries or editing configuration files, allowing rapid runtime rule testing.
- **Implementation**: Build a clean CRUD interface. Render existing filters as individual grid cards showing names, targets (e.g. `group_title`), inclusion, and exclusion patterns. Use Radix UI `Dialog` to implement the "Create Filter" form, passing list fields (comma-separated or multiple inputs) cleanly to the backend payload structure.

### 5. Add Proxy support for `/health` and `/api/v1/stats` in Vite
- **Rationale**: For local development to run smoothly, Vite must proxy all API-related requests to the Gin backend. Currently, `/health` and stats were prone to CORS or 404 errors during development.
- **Implementation**: Update `vite.config.ts` to include the health and filters routes under proxies.

## Risks / Trade-offs

- **[Risk] State and Polling Race Conditions** → Polling multiple endpoints (playlist, downloads, stats) can saturate the network or trigger race conditions on state updates.
  - *Mitigation*: Consolidate polling intervals. Poll downloads/stats only when their respective tabs or views are active. Use cleanup hooks in `useEffect` to safely cancel intervals on tab switch or unmount.
- **[Risk] Form Validation for Regex Filters** → Submitting malformed regex patterns to `/api/v1/filters` can cause backend crashes or runtime database parsing failures.
  - *Mitigation*: Ensure basic validation in the frontend form (e.g., checking that the filter name is present and at least one inclusion or exclusion pattern is defined).
- **[Risk] Heavy Render Reflows during Animations** → Frequent shimmer or pulse updates can cause layout thrashing.
  - *Mitigation*: Use GPU-accelerated CSS properties (`transform`, `opacity`) for all `@keyframes` animations, and restrict progress-bar transition durations to smooth, performant values.
