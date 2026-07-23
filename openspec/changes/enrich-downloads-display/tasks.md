# Implementation Tasks

## Phase 1: Backend File Parser (Isolated)

- [x] 1.1: Create `internal/fileparser/parser.go`
  - [x] Define FileInfo struct
  - [x] Implement Parse() function with path and tmdbYear parameters
  - [x] Implement extractYear() helper with regex `\b(19|20)\d{2}\b`
  - [x] Implement extractResolution() helper with regex for 1080p, 720p, 4K, etc.
  - [x] Create validExtensions map (.mkv, .mp4, .avi, .mov, .m4v, .wmv, .flv, .webm)
  - [x] Implement year mismatch detection logic
  - [x] Handle nil/empty path cases

- [x] 1.2: Create `internal/fileparser/parser_test.go`
  - [x] Test valid movie path with all fields populated
  - [x] Test valid TV show path
  - [x] Test missing year scenario
  - [x] Test year mismatch scenario  
  - [x] Test ambiguous year (multiple years in path)
  - [x] Test nil/empty paths
  - [x] Test unknown file formats
  - [x] Test case variations (MKV vs mkv)
  - [x] Test resolution variations (4K, 2160p, 1080P)
  - [x] Verify test coverage ≥ 95%

- [x] 1.3: Run parser tests
  ```bash
  go test ./internal/fileparser/... -v -cover
  ```

## Phase 2: Backend API Types

- [x] 2.1: Create `internal/api/types.go`
  - [x] Define DownloadEnrichedResponse struct
  - [x] Define ContentInfo struct
  - [x] Define FileInfo struct (matching fileparser.FileInfo)
  - [x] Add JSON tags to all fields
  - [x] Import time package for UpdatedAt

- [x] 2.2: Verify compilation
  ```bash
  go build ./internal/api/...
  ```

## Phase 3: Backend API Handler

- [x] 3.1: Modify `internal/api/handlers_frontend.go`
  - [x] Create listDownloadsEnriched() function
  - [x] Parse query parameters (status, type, problem)
  - [x] Build GORM query with Preload("ProcessedLines.Movie") and Preload("ProcessedLines.TVShow")
  - [x] Add status filter to WHERE clause
  - [x] Add type filter with JOIN on processed_lines
  - [x] Execute Count() for total
  - [x] Execute Find() with Order, Limit, Offset

- [x] 3.2: Implement enrichDownloadInfo() helper
  - [x] Copy base DownloadInfo fields to response
  - [x] Extract content from first ProcessedLine (if exists)
  - [x] Call fileparser.Parse() with download_path and content year
  - [x] Return DownloadEnrichedResponse

- [x] 3.3: Implement buildContentInfo() helper
  - [x] Check if ProcessedLine.Movie exists
    - [x] Extract movie title, year, genres, duration
    - [x] Set type to "movies"
  - [x] Check if ProcessedLine.TVShow exists
    - [x] Format title as "Title S##E##" if season/episode present
    - [x] Extract tvshow year, genres, season, episode
    - [x] Set type to "tvshows"
  - [x] Fallback to TvgName if no movie/tvshow
  - [x] Include resolution from ProcessedLine

- [x] 3.4: Implement matchesProblem() helper
  - [x] Implement "missing_year" logic: !file_info.has_year_in_path
  - [x] Implement "year_mismatch" logic: file_info.year_mismatch
  - [x] Implement "unknown_format" logic: !file_info.is_valid_format
  - [x] Implement "low_quality" logic: detected_res in ("480p", "360p")
  - [x] Return true for unknown filters (permissive default)

- [x] 3.5: Apply problem filter in listDownloadsEnriched()
  - [x] Loop through enriched downloads
  - [x] Filter using matchesProblem() if problemFilter is set
  - [x] Build filtered result array

## Phase 4: Backend Route Registration

- [x] 4.1: Modify `internal/api/api.go`
  - [x] Update route: `v1.GET("/downloads", s.listDownloadsEnriched)`
  - [x] (Optional) Add backward compat route: `v1.GET("/downloads/simple", s.listDownloads)`

- [x] 4.2: Test compilation
  ```bash
  go build ./cmd/...
  ```

## Phase 5: Backend Manual Testing

- [x] 5.1: Start server
  ```bash
  make run
  ```

- [x] 5.2: Test basic query
  ```bash
  curl http://localhost:8080/api/v1/downloads | jq
  ```
  - [x] Verify content field is populated
  - [x] Verify file_info field is populated

- [x] 5.3: Test status filter
  ```bash
  curl "http://localhost:8080/api/v1/downloads?status=completed" | jq
  ```

- [x] 5.4: Test type filter
  ```bash
  curl "http://localhost:8080/api/v1/downloads?type=movies" | jq
  ```

- [x] 5.5: Test problem filter
  ```bash
  curl "http://localhost:8080/api/v1/downloads?problem=missing_year" | jq
  ```

- [x] 5.6: Verify SQL queries in logs
  - [x] Check for Preload working (no N+1 queries)
  - [x] Verify JOIN is applied for type filter

## Phase 6: Frontend Types

- [x] 6.1: Update `frontend/src/App.tsx` interfaces
  - [x] Replace DownloadInfo interface with DownloadEnriched
  - [x] Add content field with type, title, year, resolution, season, episode, genres, duration
  - [x] Add file_info field with extension, folder_name, file_name, has_year_in_path, year_mismatch, detected_year, detected_resolution, is_valid_format

- [x] 6.2: Update state declarations
  - [x] Change downloads state type to DownloadEnriched[]
  - [x] Add statusFilter state
  - [x] Add typeFilter state
  - [x] Add problemFilter state

- [x] 6.3: Verify TypeScript compilation
  ```bash
  cd frontend && npm run build
  ```

## Phase 7: Frontend UI Implementation

- [x] 7.1: Update fetchDownloads() function
  - [x] Add statusFilter to query params
  - [x] Add typeFilter to query params
  - [x] Add problemFilter to query params
  - [x] Update type cast to PaginatedResponse<DownloadEnriched>

- [x] 7.2: Create filter UI in Downloads tab
  - [x] Add status dropdown (Tous, Complétés, En cours, Échoués)
  - [x] Add type dropdown (Tous, Films, Séries)
  - [x] Add problem dropdown (Aucun, Année manquante, Année incorrecte, Format inconnu, Basse qualité)
  - [x] Wire onChange handlers to state updates
  - [x] Trigger fetchDownloads() on filter change

- [x] 7.3: Enhance download card rendering
  - [x] Display content.title as primary heading (large, bold)
  - [x] Add content type icon (🎬 for movies, 📺 for tvshows)
  - [x] Display year in parentheses from content.year
  - [x] Add file path section with 📂 icon
  - [x] Display folder_name and file_name in code-style formatting
  - [x] Add technical specs line: extension, resolution, file size
  - [x] Add year indicator (✓ if present, ⚠️ if missing)
  - [x] Display year mismatch warning if applicable
  - [x] Show genres with 🎭 icon
  - [x] Render error_message in highlighted error box for failed downloads

- [x] 7.4: Add CSS styling
  - [x] Create .download-card class
  - [x] Create .download-header class
  - [x] Create .file-info class with monospace font
  - [x] Create .error-message class with error styling
  - [x] Create .filters class for filter layout
  - [x] Style select elements for filters

- [x] 7.5: Handle edge cases
  - [x] Graceful fallback when content is null (show URL)
  - [x] Graceful fallback when file_info is null (skip file section)
  - [x] Display partial progress for downloading status
  - [x] Show retry count for failed downloads

## Phase 8: Frontend Testing

- [x] 8.1: Start frontend dev server
  ```bash
  cd frontend && npm run dev
  ```

- [x] 8.2: Visual testing
  - [x] Navigate to Downloads tab
  - [x] Verify downloads display with enriched data
  - [x] Verify title is prominently displayed
  - [x] Verify file technical specs are visible
  - [x] Check icons render correctly

- [x] 8.3: Filter testing
  - [x] Test status filter (completed)
  - [x] Test status filter (failed)
  - [x] Test type filter (movies)
  - [x] Test type filter (tvshows)
  - [x] Test problem filter (missing_year)
  - [x] Test problem filter (year_mismatch)
  - [x] Verify filters can be combined

- [x] 8.4: Edge case testing
  - [x] Test with download that has no content (orphan)
  - [x] Test with download that has no download_path
  - [x] Test with failed download showing error message
  - [x] Test with in-progress download showing progress bar

- [x] 8.5: Responsive testing
  - [x] Test on mobile viewport
  - [x] Test on tablet viewport
  - [x] Verify filters wrap properly

## Phase 9: Integration & Cleanup

- [x] 9.1: End-to-end test
  - [x] Start both backend and frontend
  - [x] Verify full workflow works
  - [x] Test auto-refresh (5 second interval)

- [x] 9.2: Code review
  - [x] Check for TODOs or commented code
  - [x] Verify error handling is comprehensive
  - [x] Verify naming is consistent
  - [x] Check for unused imports

- [x] 9.3: Documentation
  - [x] Update API documentation if exists
  - [x] Add comments to complex logic
  - [x] Document any assumptions or limitations

## Phase 10: Deployment Preparation

- [x] 10.1: Build production frontend
  ```bash
  cd frontend && npm run build
  ```

- [x] 10.2: Build backend binary
  ```bash
  make build
  ```

- [x] 10.3: Test production build
  - [x] Run backend with production frontend
  - [x] Verify Downloads tab works in production mode

- [ ] 10.4: Git commit
  ```bash
  git add -A
  git commit -m "feat: enrich downloads display with TMDB metadata and file parsing"
  ```

## Success Criteria

- [x] Parser tests pass with ≥95% coverage
- [x] Backend API returns enriched data with content and file_info
- [x] Frontend displays content titles prominently
- [x] Filters work correctly (status, type, problem)
- [x] UI gracefully handles missing data
- [x] No TypeScript compilation errors
- [x] No console errors in browser
- [x] Response time < 500ms for 20 results
