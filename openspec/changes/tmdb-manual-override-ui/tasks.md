## 1. Database Schema Updates

- [ ] 1.1 Update GORM model `ProcessedLine` in `internal/models/processed_line.go` to add `OverrideBy` and `OverrideAt` fields
- [ ] 1.2 Create the new GORM model `ManualMapping` in a new file `internal/models/manual_mapping.go`
- [ ] 1.3 Update `internal/database/database.go` to auto-migrate the new `ManualMapping` model at startup
- [ ] 1.4 Write a unit test in `internal/models/models_test.go` to verify model validation and table creation for `ManualMapping`

## 2. TMDB Client Additions

- [ ] 2.1 Implement `SearchMovies` method in `internal/external/tmdb/tmdb.go` to return plural results
- [ ] 2.2 Implement `SearchTVShows` method in `internal/external/tmdb/tmdb.go` to return plural results
- [ ] 2.3 Write unit tests in `internal/external/tmdb/tmdb_test.go` to verify plural search results are correctly retrieved and parsed

## 3. Go Backend REST API Implementation

- [ ] 3.1 Update `ItemResponse` in `internal/api/dto.go` to include `OverrideBy` and `OverrideAt` fields, and update `toItemResponse` in `internal/api/handlers.go`
- [ ] 3.2 Define `OverrideItemRequest` and proxy search structures in `internal/api/dto.go`
- [ ] 3.3 Register the new routes `/api/v1/tmdb/search` (GET) and `/api/v1/items/:id/override` (POST) in `internal/api/api.go`
- [ ] 3.4 Implement `searchTMDBProxy` handler in `internal/api/handlers_frontend.go`
- [ ] 3.5 Implement `overrideItem` handler in `internal/api/handlers_frontend.go` to safely execute manual overrides and save `ManualMapping` records
- [ ] 3.6 Create and run unit tests in `internal/api/handlers_frontend_test.go` for both endpoints (success, errors, disabled config)

## 4. M3U Processor Pipeline Updates

- [ ] 4.1 Refactor `enrichMovie` in `internal/processor/processor.go` to decouple TMDB API search from GORM model upsert and item linking
- [ ] 4.2 Refactor `enrichTVShow` in `internal/processor/processor.go` to decouple TMDB API search from GORM model upsert and item linking
- [ ] 4.3 Update `setContentType` in `internal/processor/processor.go` to query the `manual_mappings` table before performing any normal enrichment
- [ ] 4.4 Write a unit test in `internal/processor/processor_test.go` to verify that the processing pipeline skips TMDB searches and applies correct associations when a manual mapping exists

## 5. React Frontend IHM Implementation

- [ ] 5.1 Update DTO interfaces in `frontend/src/App.tsx` to include override fields
- [ ] 5.2 Add an action/edit column to the M3U Playlist table in `frontend/src/App.tsx` displaying search/correction buttons
- [ ] 5.3 Implement the Radix UI Dialog modal for TMDB manual override in `frontend/src/App.tsx`
- [ ] 5.4 Implement title cleaning, automatic search, and media type selection (Film vs Série TV) in the modal
- [ ] 5.5 Render poster images, titles, years, ratings, and overviews for TMDB search results with selection highlights
- [ ] 5.6 Render optional Season and Episode input fields for series, and implement form submission with state updates

## 6. Testing & Validation

- [ ] 6.1 Run all backend Go tests (`go test ./...`) to ensure no regressions
- [ ] 6.2 Compile the frontend using Vite (`npm run build` or similar) to ensure TypeScript and linter pass cleanly
- [ ] 6.3 Verify the complete manual override, automatic mapping check, and frontend list updates end-to-end
