## Context

Our current M3U playlist view is hardcoded to 15 items per page and only supports sequential Next/Previous page steps. With scale (up to 100,000 items), this navigation mechanism is highly inefficient and frustrating for users. Furthermore, GORM executes the paginated queries with `ORDER BY created_at DESC`. Without a proper index on `created_at`, the database engine must scan and sort all matching rows, leading to slow response times under heavy datasets.

## Goals / Non-Goals

**Goals:**
- Provide a pleasant pagination UI with interactive page numbers, dynamic ellipses, and first/last page quick jump buttons.
- Enable user-configurable page limits (10, 50, 100 entries per page) persisted in `localStorage`.
- Optimize database query performance for large datasets (100,000+ items) by adding an index on `created_at`.

**Non-Goals:**
- Implement infinite-scroll or virtualized list scrolling (e.g., React Window), as the user explicitly requested a paginated table experience.
- Refactor the backend payload format or introduce cursor-based pagination, which would break compatibility and prevent random page jumps.

## Decisions

### Decision 1: Index on `created_at`
- **Rationale**: An index on the sorted column (`created_at`) allows GORM's `ORDER BY` + `LIMIT` + `OFFSET` queries to execute in O(log N) complexity, avoiding expensive full-table scans and disk filesorts.
- **Alternatives considered**: Index on `(content_type, state, created_at)` composite. This is unnecessary since GORM pre-filters on individual indexes, and a simple index on `created_at` is lighter and sufficient.

### Decision 2: Smart Page Range Array Generator
- **Rationale**: We will implement a custom range calculation helper in React that yields an array like `[1, 2, '...', 41, 42, 43, 44, 45, '...', 1000]`. This avoids pulling third-party heavy pagination dependencies.
- **Alternatives considered**: Standard paginate library. Keeping it vanilla is more lightweight and perfectly customized to our CSS theme.

### Decision 3: LocalStorage Preference Persistence
- **Rationale**: Saving the chosen limit under `stalkeer_playlist_limit` prevents the UI from resetting to default on page refresh or tab switching, ensuring a seamless user experience.

## Risks / Trade-offs

- **[Risk]** Large offsets (e.g., page 2000 with limit 50) can still have some scan cost in PostgreSQL/SQLite.
  - **Mitigation**: The search and filters (Movies/TVShows) are highly effective and narrow down the count significantly. With proper indexing, query execution times will remain extremely low.
- **[Risk]** Parsing error when reading `localStorage`.
  - **Mitigation**: Wrap the retrieval in a robust fallback mechanism defaulting to 50 if the stored value is invalid or missing.
