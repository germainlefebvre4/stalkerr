## Why

When M3U IPTV playlist titles are heavily malformed or poorly standardized, automatic fuzzing matching algorithms (even with high-tolerance thresholds) fail to resolve correct TMDB metadata. This results in either incorrect TMDB association or completely "unmatched" items in the database.

Introducing an interactive **Manual Override UI** on the React frontend, coupled with a secure **TMDB Search Proxy** and a **Persistent Learning Mapping System** on the Go backend, solves this problem. It allows users to manually search and "force" associations, which are then remembered permanently to automatically resolve identical IPTV titles in future playlist updates.

## What Changes

- **Database Model**: Expand the `ProcessedLine` model with `OverrideBy` and `OverrideAt` columns to log manual overrides.
- **Persistent Mapping Schema**: Introduce a new GORM model/table `ManualMapping` storing the mapping of `(TvgName, GroupTitle) -> TMDB ID/ContentType/Season/Episode` to remember user corrections.
- **Go Backend API**:
  - Add `GET /api/v1/tmdb/search` as a secure proxy to search movies or TV shows on TMDB without exposing the API key to the frontend.
  - Add `POST /api/v1/items/:id/override` to execute a manual override, fetching TMDB metadata, creating the local media record (Movie or TVShow), linking the item, and saving the persistent learning mapping.
- **TMDB Client**: Add plural `SearchMovies` and `SearchTVShows` methods to TMDB client to return complete result slices instead of just the first result.
- **M3U Processor**: Modify the automatic VOD classification pipeline to consult the `manual_mappings` database table first, bypassing TMDB API search entirely for previously corrected titles.
- **React Frontend**:
  - Add an "Association / Correction" column or action button to the **🗒️ Playlist M3U** table.
  - Implement a highly-polished, accessible, and interactive search and override modal dialog (using Radix UI) displaying poster images, release years, overviews, and custom season/episode selectors.

## Capabilities

### New Capabilities
- `tmdb-manual-override`: Covers the manual TMDB override REST API and the interactive frontend search/override modal dialog.

### Modified Capabilities
- `m3u-title-normalization`: Integrate the persistent manual mappings check into the processing pipeline, prioritizing learned matches before fuzzy-matching titles.

## Impact

- **Database**: GORM migration will add two columns to `processed_lines` and create the new `manual_mappings` table.
- **Backend API Router**: Exposes two new API endpoints under `/api/v1`.
- **TMDB Integration**: Minor addition of plural search endpoints in TMDB package.
- **Frontend App**: Expands `App.tsx` with a new override dialog and list integration. No external library additions (uses Radix UI which is already installed).
