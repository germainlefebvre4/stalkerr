## Context

See `proposal.md - Why` for the motivating bug reports. Three existing pieces of machinery this design builds on:

- `internal/matcher/matcher.go` already ranks candidates purely by `resolutionOrderSQL`, a SQL `CASE` expression used in `ORDER BY`.
- `internal/processor/processor.go` already recognizes `VF`/`VOSTFR`/`MULTI` tokens (via `qualitySuffixRe`) when normalizing titles for matching, but only strips them — it never persists which one was found.
- `internal/api/force_download.go` already demonstrates the "live Radarr/Sonarr lookup by TMDB/TVDB ID, independent of the missing list" pattern (`GetMovieByTMDBID`) that tier-2 needs; `GetSeriesByTVDBID` exists on the Sonarr client but is unused.

`internal/scheduler/build.go` is the integration point: `buildTier2MovieStreams`/`buildTier2SeriesStreams` currently gate only on `downloaded > 0 && len(candidates) > 0`, and compute destination paths locally instead of through Radarr/Sonarr.

## Goals / Non-Goals

**Goals:**
- Persist a language classification on `ProcessedLine` and fold it into candidate ranking ahead of resolution.
- Make tier-2 stream construction compare the best untried candidate against the already-downloaded one, using that same ranking, and only proceed when it is strictly better.
- Make tier-2 destination paths come from a live Radarr/Sonarr lookup, matching tier-1 and `force_download.go`.

**Non-Goals:**
- Backfilling `language` for pre-existing `ProcessedLine` rows via a dedicated migration command — `NULL` is a valid, correctly-ranked value (see Decisions), so existing rows are simply re-classified the next time they are re-ingested, if ever.
- Changing the manual force-download flow (`force-download-media-occurrence`) — it already documents and relies on ignoring sibling candidates entirely; out of scope here.
- Surfacing `language` in the frontend UI — this change is scoped to the automatic download pipeline's decision logic.

## Decisions

### Language detection reuses the existing token recognition, doesn't duplicate it
`processor.go`'s `qualitySuffixRe` already matches `VF`/`VOSTFR`/`MULTI` (among other tokens) against the same text used for title normalization. Rather than writing a second, independent regex against `LineContent`/`TvgName`/`GroupTitle`, the processor captures which language token (if any) matched during that same pass and stores it on the `ProcessedLine` being built, before the token is stripped for title comparison. This keeps a single source of truth for "what language token did we see" instead of two regexes that could drift apart.

### Ranking stays SQL-driven, extended to a compound key
`resolutionOrderSQL` becomes a compound `ORDER BY` clause: a language `CASE` expression first, the existing resolution `CASE` second, then `created_at DESC`. This mirrors the existing architecture (`FindMovieDownloadCandidates`/`FindTVShowDownloadCandidates` already just return SQL-ordered rows) rather than fetching unordered rows and sorting in Go, which would require call-site changes beyond `matcher.go`.

Language `CASE` order: `VF` (1) → `MULTI` (2) → `NULL` (3) → `VOSTFR` (4), per the proposal's stated preference (unmarked entries are assumed French, being a French-IPTV catalog).

### Tier-2 comparison is a small Go ranking function, not a second SQL query
Gating tier-2 construction requires comparing two specific rows (the already-downloaded one vs. the best untried one), not just ordering a list. `buildTier2MovieStreams`/`buildTier2SeriesStreams` already load the downloaded `ProcessedLine` (via `countDownloaded`, extended to fetch the row instead of just a count) and the untried candidates (via the existing `FindMovieDownloadCandidates`/`FindTVShowDownloadCandidates`, already SQL-ordered so `candidates[0]` is the best untried one). A small Go helper (`(language, resolution) -> ordinal`) reuses the same preference order as the SQL `CASE` expressions and compares `candidates[0]`'s ordinal against the downloaded row's ordinal. Keeping this comparison in Go avoids maintaining the ranking logic in two SQL dialects of the same rule.

### Tier-2 path lookup happens after the strictly-better gate, not before
`GetMovieByTMDBID`/`GetSeriesByTVDBID` are only called for a movie/series-season that has already passed the strictly-better check. Ordering it this way means the fix for cause n°1 (fewer tier-2 candidates overall) directly reduces the number of extra Radarr/Sonarr calls introduced by the fix for cause n°3, instead of the two changes compounding costs independently.

### A failed live lookup drops that one candidate for this run, not the whole run
Consistent with the existing pattern in `buildTier1MovieStreams` (`matcher.MatchMovieByTVDB` error → `continue`), a `GetMovieByTMDBID`/`GetSeriesByTVDBID` error for one tier-2 candidate is logged and that candidate is skipped for this run rather than aborting `BuildStreams` entirely. A "no match found" result (movie no longer in Radarr) falls back to the existing `cfg.Downloads.MoviesPath`/`TVShowsPath`-based path, same as tier-1's empty-path fallback.

## Risks / Trade-offs

- **[Risk]** Extending `RadarrClient`/`SonarrClient` (the interfaces `BuildDeps` depends on) with `GetMovieByTMDBID`/`GetSeriesByTVDBID` touches every existing fake/mock implementing those interfaces in tests. → **Mitigation**: both methods already exist on the concrete `radarr.Client`/`sonarr.Client`; only the narrow interfaces and their test fakes need the new method added.
- **[Risk]** Compound language+resolution ordering changes candidate selection for tier-1 too (not just tier-2), since both tiers share `FindMovieDownloadCandidates`/`FindTVShowDownloadCandidates`. → **Mitigation**: this is intentional — a movie being downloaded for the first time should also prefer `VF` over `VOSTFR` at the same resolution tier; call this out explicitly so it isn't mistaken for scope creep during review.
- **[Trade-off]** Existing `ProcessedLine` rows keep `language = NULL` until reprocessed, so recently-ingested-but-not-yet-reclassified rows rank as "unspecified" (between `MULTI` and `VOSTFR`) rather than their true language until the M3U line is reprocessed. Accepted per the Non-Goals above.

## Migration Plan

- Add `Language *string` to `models.ProcessedLine` with a `gorm` tag; the existing `db.AutoMigrate(...)` call in `internal/database/database.go` adds the nullable column on next startup — no manual migration script needed, consistent with how `resolution` was added.
- No rollback beyond reverting the code change; the added column is additive and nullable, so it is safe to leave in place even if the change is reverted.
