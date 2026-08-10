## Context

See `proposal.md` - Why. Two existing write paths already fetch full TMDB detail + external-ID responses but discard most of the payload before persistence:
- `internal/processor/processor.go` (`enrichMovieWithTMDBID` / TV show equivalent), called during `stalkeer process` ingestion.
- `internal/api/handlers_frontend.go` (`POST /api/v1/items/:id/override` handler), called on manual override.

Both currently keep only `TMDBID, TVDBID, TMDBTitle, TMDBYear, TMDBGenres` (+ `Duration` or `Season`/`Episode`) in a restricted `attrs` map before `FirstOrCreate`. `poster_path`, `overview`, and external IDs (`imdb_id`, `tvdb_id`) are present in the in-memory API response at that point but never mapped.

A precedent for a backfill-style command already exists: `cmd/enrich_tvdb.go` (`stalkeer enrich-tvdb`), which queries `Movie`/`TVShow` rows with `tmdb_id` set and `tvdb_id` null, dedupes by `tmdb_id`, and calls `GetMovieExternalIDs`/`GetTVShowExternalIDs`. It is a separate, manually-invoked command (`openspec/specs/tvdb-id-backfill`). This change deliberately does **not** follow that pattern for its own backfill — the backfill here is silent and automatic, embedded inside `process`, not a new manual command.

The TMDB client already has request throttling and response caching (`openspec/specs/tmdb-rate-limiter`): default 4 req/s, in-memory cache, `Retry-After` compliance on 429.

## Goals / Non-Goals

**Goals:**
- Persist `poster_path`, `overview`, `imdb_id` (new columns) and expose `tvdb_id` (already stored) on `Movie`/`TVShow`, at zero additional TMDB API cost for the two existing write paths.
- Backfill these fields on records that predate this change, automatically and idempotently, every time `stalkeer process` runs, with no dedicated flag.
- Surface poster, synopsis, and TMDB/IMDB/TheTVDB links in the playlist sidepanel.

**Non-Goals:**
- No periodic refresh of already-backfilled records (no re-fetch once `poster_path` is set) — this is explicitly a one-time-per-record backfill, not a sync mechanism for data that may change on TMDB's side over time.
- No `vote_average`/`backdrop_path` or any other TMDB field beyond the five listed above.
- No new manual CLI command (unlike `enrich-tvdb`'s pattern) — the backfill has no `--dry-run`/`--limit`/`--verbose` flags and cannot be disabled independently of `--skip-tmdb`.
- No change to the TMDB search proxy or the manual override modal's search UX.

## Decisions

**Extend `attrs` in both existing write paths instead of adding a third write path.** Both `processor.go` and the override handler already hold the full `MovieDetails`/`TVShowDetails` and `ExternalIDs` structs in memory when they build their restricted `attrs` map. Adding the four new keys to those two existing maps is a minimal, low-risk change that guarantees no behavioral drift between the two paths.

**Backfill lives inside `processor.Process()`, not as a new `cmd/*.go` command.** This was an explicit product decision (see conversation): unlike `enrich-tvdb`, this backfill must run automatically whenever `stalkeer process` executes, with no separate invocation and no flag to control it. Implementation-wise this means a new function in the `processor` package (structurally similar to `EnrichMissingTVDBIDs`: query, dedupe by `tmdb_id`, fetch, update, per-record error handling) invoked once per `Process()` call, guarded by the same `SkipTMDB` option and TMDB-enabled config check already used for the main enrichment flow.

**No control flag for the backfill.** Confirmed acceptable: the lookup query is cheap when nothing needs backfilling (the common case after the first run), and when it does have work, it is throttled by the existing rate limiter like any other TMDB usage. On a large legacy library, the first post-upgrade `process` run will take longer (order of minutes, bounded by rate limit × record count) — this is accepted as a one-time cost.

**Frontend link construction happens client-side.** `https://www.themoviedb.org/{movie|tv}/{tmdb_id}`, `https://www.imdb.com/title/{imdb_id}`, and `https://thetvdb.com/?tab=series&id={tvdb_id}` are built directly in `PlaylistTab.tsx` from fields already present in the API payload — no backend endpoint returns pre-formed URLs, avoiding a redundant round-trip and keeping URL-construction logic (and any future changes to it) in one place.

## Risks / Trade-offs

- **First run after upgrade is slower** on libraries with many legacy TMDB-matched items, since the backfill runs inline with normal ingestion → Mitigated by per-record error isolation (a failed record doesn't abort the run) and reuse of the existing rate limiter's caching (repeated TV show rows sharing a `tmdb_id` cost one request, not N).
- **Stale metadata never refreshes** — if TMDB later updates a poster or synopsis, the stored copy will not follow → Accepted as a non-goal; can be revisited later as a separate change if it becomes a problem.
- **Sidepanel must handle partially-populated records** during the window between deploying this change and the first backfilling `process` run (or for records whose TMDB lookup keeps failing) → Mitigated by the spec's explicit "gracefully omits missing metadata" scenario; frontend renders conditionally per field.

## Migration Plan

1. Add nullable columns to `Movie`/`TVShow` (GORM `AutoMigrate` handles this automatically on next startup — no manual SQL migration needed, consistent with existing patterns in `internal/database/database.go`).
2. Extend `attrs` in `processor.go` and the override handler to persist the new fields going forward.
3. Add the backfill routine to `processor.Process()`, gated by the existing TMDB-enabled/`SkipTMDB` checks.
4. Extend `MovieResponse`/`TVShowResponse` and the frontend types/sidepanel.
5. Rollback: the new columns are additive and nullable; reverting the code change leaves them unused but harmless. No data migration needs reversing.
