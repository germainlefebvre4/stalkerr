# Implementation Tasks

## Phase 1: Backend File Parser (Isolated)

- [ ] 1.1: Create `internal/fileparser/parser.go`
  - [ ] Define FileInfo struct
  - [ ] Implement Parse() function with path and tmdbYear parameters
  - [ ] Implement extractYear() helper with regex `\b(19|20)\d{2}\b`
  - [ ] Implement extractResolution() helper with regex for 1080p, 720p, 4K, etc.
  - [ ] Create validExtensions map (.mkv, .mp4, .avi, .mov, .m4v, .wmv, .flv, .webm)
  - [ ] Implement year mismatch detection logic
  - [ ] Handle nil/empty path cases

- [ ] 1.2: Create `internal/fileparser/parser_test.go`
  - [ ] Test valid movie path with all fields populated
  - [ ] Test valid TV show path
  - [ ] Test missing year scenario
  - [ ] Test year mismatch scenario  
  - [ ] Test ambiguous year (multiple years in path)
  - [ ] Test nil/empty paths
  - [ ] Test unknown file formats
  - [ ] Test case variations (MKV vs mkv)
  - [ ] Test resolution variations (4K, 2160p, 1080P)
  - [ ] Verify test coverage ≥ 95%

- [ ] 1.3: Run parser tests
  ```bash
  go test ./internal/fileparser/... -v -cover
  ```

## Phase 2: Backend API Types

- [ ] 2.1: Create `internal/api/types.go`
  - [ ] Define DownloadEnrichedResponse struct
  - [ ] Define ContentInfo struct
  - [ ] Define FileInfo struct (matching fileparser.FileInfo)
  - [ ] Add JSON tags to all fields
  - [ ] Import time package for UpdatedAt

- [ ] 2.2: Verify compilation
  ```bash
  go build ./internal/api/...
  ```

## Phase 3: Backend API Handler

- [ ] 3.1: Modify `internal/api/handlers_frontend.go`
  - [ ] Create listDownloadsEnriched() function
  - [ ] Parse query parameters (status, type, problem)
  - [ ] Build GORM query with Preload("ProcessedLines.Movie") and Preload("ProcessedLines.TVShow")
  - [ ] Add status filter to WHERE clause
  - [ ] Add type filter with JOIN on processed_lines
  - [ ] Execute Count() for total
  - [ ] Execute Find() with Order, Limit, Offset

- [ ] 3.2: Implement enrichDownloadInfo() helper
  - [ ] Copy base DownloadInfo fields to response
  - [ ] Extract content from first ProcessedLine (if exists)
  - [ ] Call fileparser.Parse() with download_path and content year
  - [ ] Return DownloadEnrichedResponse

- [ ] 3.3: Implement buildContentInfo() helper
  - [ ] Check if ProcessedLine.Movie exists
    - [ ] Extract movie title, year, genres, duration
    - [ ] Set type to "movies"
  - [ ] Check if ProcessedLine.TVShow exists
    - [ ] Format title as "Title S##E##" if season/episode present
    - [ ] Extract tvshow year, genres, season, episode
    - [ ] Set type to "tvshows"
  - [ ] Fallback to TvgName if no movie/tvshow
  - [ ] Include resolution from ProcessedLine

- [ ] 3.4: Implement matchesProblem() helper
  - [ ] Implement "missing_year" logic: !file_info.has_year_in_path
  - [ ] Implement "year_mismatch" logic: file_info.year_mismatch
  - [ ] Implement "unknown_format" logic: !file_info.is_valid_format
  - [ ] Implement "low_quality" logic: detected_res in ("480p", "360p")
  - [ ] Return true for unknown filters (permissive default)

- [ ] 3.5: Apply problem filter in listDownloadsEnriched()
  - [ ] Loop through enriched downloads
  - [ ] Filter using matchesProblem() if problemFilter is set
  - [ ] Build filtered result array

## Phase 4: Backend Route Registration

- [ ] 4.1: Modify `internal/api/api.go`
  - [ ] Update route: `v1.GET("/downloads", s.listDownloadsEnriched)`
  - [ ] (Optional) Add backward compat route: `v1.GET("/downloads/simple", s.listDownloads)`

- [ ] 4.2: Test compilation
  ```bash
  go build ./cmd/...
  ```

## Phase 5: Backend Manual Testing

- [ ] 5.1: Start server
  ```bash
  make run
  ```

- [ ] 5.2: Test basic query
  ```bash
  curl http://localhost:8080/api/v1/downloads | jq
  ```
  - [ ] Verify content field is populated
  - [ ] Verify file_info field is populated

- [ ] 5.3: Test status filter
  ```bash
  curl "http://localhost:8080/api/v1/downloads?status=completed" | jq
  ```

- [ ] 5.4: Test type filter
  ```bash
  curl "http://localhost:8080/api/v1/downloads?type=movies" | jq
  ```

- [ ] 5.5: Test problem filter
  ```bash
  curl "http://localhost:8080/api/v1/downloads?problem=missing_year" | jq
  ```

- [ ] 5.6: Verify SQL queries in logs
  - [ ] Check for Preload working (no N+1 queries)
  - [ ] Verify JOIN is applied for type filter

## Phase 6: Frontend Types

- [ ] 6.1: Update `frontend/src/App.tsx` interfaces
  - [ ] Replace DownloadInfo interface with DownloadEnriched
  - [ ] Add content field with type, title, year, resolution, season, episode, genres, duration
  - [ ] Add file_info field with extension, folder_name, file_name, has_year_in_path, year_mismatch, detected_year, detected_resolution, is_valid_format

- [ ] 6.2: Update state declarations
  - [ ] Change downloads state type to DownloadEnriched[]
  - [ ] Add statusFilter state
  - [ ] Add typeFilter state
  - [ ] Add problemFilter state

- [ ] 6.3: Verify TypeScript compilation
  ```bash
  cd frontend && npm run build
  ```

## Phase 7: Frontend UI Implementation

- [ ] 7.1: Update fetchDownloads() function
  - [ ] Add statusFilter to query params
  - [ ] Add typeFilter to query params
  - [ ] Add problemFilter to query params
  - [ ] Update type cast to PaginatedResponse<DownloadEnriched>

- [ ] 7.2: Create filter UI in Downloads tab
  - [ ] Add status dropdown (Tous, Complétés, En cours, Échoués)
  - [ ] Add type dropdown (Tous, Films, Séries)
  - [ ] Add problem dropdown (Aucun, Année manquante, Année incorrecte, Format inconnu, Basse qualité)
  - [ ] Wire onChange handlers to state updates
  - [ ] Trigger fetchDownloads() on filter change

- [ ] 7.3: Enhance download card rendering
  - [ ] Display content.title as primary heading (large, bold)
  - [ ] Add content type icon (🎬 for movies, 📺 for tvshows)
  - [ ] Display year in parentheses from content.year
  - [ ] Add file path section with 📂 icon
  - [ ] Display folder_name and file_name in code-style formatting
  - [ ] Add technical specs line: extension, resolution, file size
  - [ ] Add year indicator (✓ if present, ⚠️ if missing)
  - [ ] Display year mismatch warning if applicable
  - [ ] Show genres with 🎭 icon
  - [ ] Render error_message in highlighted error box for failed downloads

- [ ] 7.4: Add CSS styling
  - [ ] Create .download-card class
  - [ ] Create .download-header class
  - [ ] Create .file-info class with monospace font
  - [ ] Create .error-message class with error styling
  - [ ] Create .filters class for filter layout
  - [ ] Style select elements for filters

- [ ] 7.5: Handle edge cases
  - [ ] Graceful fallback when content is null (show URL)
  - [ ] Graceful fallback when file_info is null (skip file section)
  - [ ] Display partial progress for downloading status
  - [ ] Show retry count for failed downloads

## Phase 8: Frontend Testing

- [ ] 8.1: Start frontend dev server
  ```bash
  cd frontend && npm run dev
  ```

- [ ] 8.2: Visual testing
  - [ ] Navigate to Downloads tab
  - [ ] Verify downloads display with enriched data
  - [ ] Verify title is prominently displayed
  - [ ] Verify file technical specs are visible
  - [ ] Check icons render correctly

- [ ] 8.3: Filter testing
  - [ ] Test status filter (completed)
  - [ ] Test status filter (failed)
  - [ ] Test type filter (movies)
  - [ ] Test type filter (tvshows)
  - [ ] Test problem filter (missing_year)
  - [ ] Test problem filter (year_mismatch)
  - [ ] Verify filters can be combined

- [ ] 8.4: Edge case testing
  - [ ] Test with download that has no content (orphan)
  - [ ] Test with download that has no download_path
  - [ ] Test with failed download showing error message
  - [ ] Test with in-progress download showing progress bar

- [ ] 8.5: Responsive testing
  - [ ] Test on mobile viewport
  - [ ] Test on tablet viewport
  - [ ] Verify filters wrap properly

## Phase 9: Integration & Cleanup

- [ ] 9.1: End-to-end test
  - [ ] Start both backend and frontend
  - [ ] Verify full workflow works
  - [ ] Test auto-refresh (5 second interval)

- [ ] 9.2: Code review
  - [ ] Check for TODOs or commented code
  - [ ] Verify error handling is comprehensive
  - [ ] Verify naming is consistent
  - [ ] Check for unused imports

- [ ] 9.3: Documentation
  - [ ] Update API documentation if exists
  - [ ] Add comments to complex logic
  - [ ] Document any assumptions or limitations

## Phase 10: Deployment Preparation

- [ ] 10.1: Build production frontend
  ```bash
  cd frontend && npm run build
  ```

- [ ] 10.2: Build backend binary
  ```bash
  make build
  ```

- [ ] 10.3: Test production build
  - [ ] Run backend with production frontend
  - [ ] Verify Downloads tab works in production mode

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
