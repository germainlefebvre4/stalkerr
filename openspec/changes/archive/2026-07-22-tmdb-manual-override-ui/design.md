## Context

The current automatic enrichment pipeline works extremely well for standard, well-formatted titles. However, many IPTV playlist titles contain severe formatting issues (e.g. typos, strange suffixes, or no year/season/episode cues) that cause the automatic matcher to fail or misidentify titles.

Currently, we have:
- A `ProcessedLine` GORM model that maps to the `processed_lines` database table.
- A `Classifier` that parses resolutions, seasons, and episodes.
- A `tmdb.Client` that performs automated fuzzy searches.
- A simple React dashboard built using Vite, TailwindCSS, and Radix UI primitives.

By adding a **Manual Override Dialog** in the UI, users can resolve unmatched or misclassified titles. By coupling this with a **Persistent Learning Mapping system**, any override performed once is saved permanently, meaning identical IPTV titles in future M3U imports will automatically resolve correctly without repeating the manual step.

## Goals / Non-Goals

**Goals:**
- Provide a secure backend search proxy API for TMDB movies and series.
- Provide a manual override REST endpoint to link VOD items with custom TMDB entities.
- Implement a persistent mapping system to automatically match identical raw IPTV titles in future imports.
- Maintain atomic consistency and avoid duplicate code by refactoring the enrichment process.
- Implement an interactive, polished Radix UI modal dialog for searching and associating.

**Non-Goals:**
- Changing existing automatic fuzzy matching logic thresholds.
- Implementing automated bulk corrections or bulk matching suggestions (only item-by-item manual matching).
- Integrating other metadata providers besides TMDB.

## Decisions

### D1: Strict separation of search modes (movies vs tvshows)
- **Choice**: Separate TMDB search proxy query types strictly on the UI and backend levels.
- **Rationale**: TMDB API has distinct search endpoints for movies (`/search/movie`) and TV shows (`/search/tv`). Separating them ensures cleaner search results, prevents mixed types, and allows the UI to render movie-specific features (like release years) or TV-specific features (like seasons/episodes) cleanly.
- **Alternative considered**: Use `/search/multi` which searches everything. This was rejected because the results are noisy, harder to filter, and would require more complex conditional parsing in Go.

### D2: Persistent mapping via a dedicated `manual_mappings` database table
- **Choice**: Introduce a separate `manual_mappings` GORM model/table.
- **Rationale**: Storing the user's manual correction `(TvgName, GroupTitle) -> TMDBID` in a separate table creates a decoupled "dictionary of associations" that is highly performant to query. Checking this table during M3U processing is a simple database lookup that bypasses TMDB entirely.
- **Alternative considered**: Storing the learning association on the `ProcessedLine` itself. This was rejected because `ProcessedLine` records can be hard-deleted or pruned by the cleanup process (e.g., if a channel is removed from the playlist), which would lose the learned mapping for future playlist versions. A dedicated mapping table survives list prunes.

### D3: Refactoring the enrichment logic to be modular (DRY)
- **Choice**: Refactor `enrichMovie` and `enrichTVShow` in the processor to accept a TMDB ID directly.
- **Rationale**: Splitting the automated matching into: (1) Find TMDB ID (automated search OR manual mapping lookup) and (2) Fetch and link metadata (common for both auto and manual flows) allows us to reuse the exact GORM database logic without copy-pasting code, keeping our codebase clean and maintainable.

### D4: Safe TMDB client initialization in the API Server
- **Choice**: Initialize a `tmdb.Client` on the `Server` struct in `NewServer()` if TMDB configuration is enabled and a key is set.
- **Rationale**: Keeping the TMDB client as a single reusable singleton on the `Server` struct is memory-efficient, shares the internal client cache and rate-limit tracker safely, and simplifies tests (mocking or pointing to local httptest servers is trivial).

## Risks / Trade-offs

- **[Risk]**: TMDB Rate Limiting if users spam the interactive search.
  - **Mitigation**: Implement search debouncing (300ms) on the React frontend, and leverage the TMDB client's built-in rate-limiting and cache mechanisms.
- **[Risk]**: Incorrect Season/Episode manually specified for a TV Show.
  - **Mitigation**: Perform input validation on the backend to ensure season and episode inputs are non-negative, and allow users to execute a new override if they make a mistake.
- **[Risk]**: Stale manual mappings if a movie or TV show's entry changes on TMDB (ID deletion).
  - **Mitigation**: Very low risk, as TMDB IDs are highly stable. If it happens, the user can re-trigger a manual override which will update the mapping to the new correct TMDB ID.
