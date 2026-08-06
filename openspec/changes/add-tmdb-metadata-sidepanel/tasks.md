## 1. Data model

- [ ] 1.1 Add nullable `PosterPath`, `Overview`, `IMDBID` fields to `Movie` and `TVShow` in `internal/models/media.go` (with appropriate `gorm` tags, e.g. `Overview *string` as `type:text`)
- [ ] 1.2 Verify `TVDBID` is already present on both models (no change needed there beyond exposure below)
- [ ] 1.3 Confirm GORM `AutoMigrate` in `internal/database/database.go` picks up the new columns on next startup (no manual migration file required)

## 2. Backend persistence — existing write paths

- [ ] 2.1 In `internal/processor/processor.go`, extend the `attrs` map built in the movie enrichment path (`enrichMovieWithTMDBID` or equivalent) to include `poster_path`, `overview`, `imdb_id`, `tvdb_id` from the already-fetched `GetMovieDetails`/`GetMovieExternalIDs` responses
- [ ] 2.2 Apply the same extension to the TV show enrichment path (`GetTVShowDetails`/`GetTVShowExternalIDs`)
- [ ] 2.3 In `internal/api/handlers_frontend.go`, extend the equivalent `attrs` map in the `POST /api/v1/items/:id/override` handler (movie branch) with the same four fields
- [ ] 2.4 Apply the same extension to the TV show branch of the override handler
- [ ] 2.5 Add/update backend tests covering that both write paths persist the new fields when TMDB returns them, and leave them null when TMDB details/external IDs are unavailable

## 3. Backend backfill routine

- [ ] 3.1 Add a new function in the `processor` package (e.g. `BackfillRichMetadata`) that queries `Movie`/`TVShow` rows where `tmdb_id` is set and `poster_path` is null, deduplicating `TVShow` rows by `tmdb_id` before calling TMDB (mirroring `EnrichMissingTVDBIDs`'s structure)
- [ ] 3.2 For each unique record/`tmdb_id` needing backfill, fetch `GetMovieDetails`/`GetTVShowDetails` + `GetMovieExternalIDs`/`GetTVShowExternalIDs` and update `poster_path`, `overview`, `imdb_id`, `tvdb_id`
- [ ] 3.3 On a per-record TMDB error, log a warning, leave the record unchanged, and continue with the next record without failing the overall run
- [ ] 3.4 Invoke this backfill once per `processor.Process()` call, guarded by the same TMDB-enabled/`SkipTMDB` checks used by the main enrichment flow (skip entirely if TMDB is disabled or `--skip-tmdb` was passed)
- [ ] 3.5 Ensure the backfill reuses the existing rate-limited TMDB client (no separate client instance, no bypass of `RequestsPerSecond`)
- [ ] 3.6 Add tests: legacy record gets backfilled, no-op when nothing needs backfilling, TV show dedup by `tmdb_id`, per-record failure isolation, skipped when TMDB disabled/`--skip-tmdb`

## 4. API response

- [ ] 4.1 Add `PosterPath`, `Overview`, `IMDBID`, `TVDBID` fields to `MovieResponse` and `TVShowResponse` in `internal/api/dto.go`, using `omitempty`/pointer types so absent values serialize as null/omitted
- [ ] 4.2 Update the model→DTO mapping code (wherever `MovieResponse`/`TVShowResponse` are constructed, e.g. in `handlers_frontend.go`) to populate the four new fields
- [ ] 4.3 Update/extend `internal/api/handlers_frontend_test.go` (and `handlers_filters_test.go` if relevant) to assert the new fields appear in item list/detail responses

## 5. Frontend

- [ ] 5.1 Extend `MovieResponse`/`TVShowResponse` types in `frontend/src/types.ts` with `poster_path`, `overview`, `imdb_id`, `tvdb_id` (optional/nullable)
- [ ] 5.2 In `frontend/src/components/PlaylistTab.tsx`'s sidepanel, render a poster thumbnail (TMDB image CDN URL built from `poster_path`) when present
- [ ] 5.3 Render the `overview` synopsis text when present
- [ ] 5.4 Build and render links to the TMDB, IMDB, and TheTVDB pages from `tmdb_id` (+ `content_type` for movie/tv path), `imdb_id`, and `tvdb_id` respectively, each opening in a new tab, only when the corresponding ID is present
- [ ] 5.5 Ensure the panel renders cleanly with no broken image/empty link when poster/overview/IDs are absent (not yet backfilled)

## 6. Verification

- [ ] 6.1 Run backend test suite (`go test ./...`)
- [ ] 6.2 Manually verify: process a playlist entry, confirm poster/overview/links appear in the sidepanel
- [ ] 6.3 Manually verify: run `stalkeer process` against an existing database with pre-change `Movie`/`TVShow` rows lacking the new fields, confirm they get backfilled and the sidepanel reflects it afterward
