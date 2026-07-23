## Why

The Downloads tab currently displays only technical information (URLs, file paths, bytes), making it difficult to verify what content was actually downloaded, identify quality issues, or confirm proper file organization. Users need to see enriched metadata (movie/TV show titles, years, formats, resolutions) to validate downloads and troubleshoot failures effectively.

## What Changes

- Enhance backend API `/api/v1/downloads` to include TMDB metadata and parsed file information
- Create a file parser module to extract technical details (extension, resolution, year detection) from download paths
- Update frontend Downloads tab to display content titles, years, formats, and folder structures with visual indicators for issues
- Implement filtering by status, content type, and detected problems (missing year, invalid format)
- Use GORM Preload to join download_info with processed_lines, movies, and tvshows tables

## Capabilities

### New Capabilities
- `file-metadata-parsing`: Backend parser to extract file extension, resolution, year, and format validity from download paths
- `downloads-enrichment-api`: Enhanced API response including content metadata (title, year, genres) and parsed file information
- `downloads-display-ui`: Frontend UI showing enriched download cards with titles, technical specs, and problem indicators

### Modified Capabilities
<!-- No existing capabilities are being modified, this is net-new functionality -->

## Impact

**New Files**:
- `internal/fileparser/parser.go` - File metadata extraction logic
- `internal/fileparser/parser_test.go` - Parser unit tests
- `internal/api/types.go` - API response types for enriched downloads

**Modified Files**:
- `internal/api/handlers_frontend.go` - Replace `listDownloads` with `listDownloadsEnriched`
- `frontend/src/App.tsx` - Update Downloads tab UI with enriched display

**API Changes**: GET `/api/v1/downloads` returns enriched response with `content` and `file_info` fields

**Database**: No schema changes required (uses existing GORM associations)
