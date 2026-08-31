## Context

Today the API server (`cmd/server.go` / `internal/api`) is a pure query/management surface: it never starts a file transfer. All actual downloading happens in the separate `radarr`/`sonarr` CLI commands, run on a schedule (k8s CronJobs), which source their work from Radarr/Sonarr's `/wanted/missing`-style endpoints and always stop after the first successful candidate per movie/episode (see `internal/matcher.FindMovieDownloadCandidates` and `cmd/radarr.go`). Neither the Radarr nor the Sonarr internal ID, nor `movie.Path`/series path, is ever persisted locally — they're only known transiently inside that CLI loop.

There is also an existing, separate `resume-downloads` CLI command/CronJob that scans `DownloadInfo` rows in interruptible states (`pending`, `downloading`, `paused`, `failed`) and resumes them (`internal/downloader/resume_helper.go`). This is the safety net this design leans on for server-restart resilience, instead of building bespoke resilience for the new code path.

See `proposal.md` for the motivation and `specs/force-download-media-occurrence/spec.md` and the sidepanel delta for the behavior contract.

## Goals / Non-Goals

**Goals:**
- Let the API server itself perform a targeted, on-demand, asynchronous download for one `ProcessedLine`, something it has never done before.
- Add the minimal Radarr/Sonarr client surface needed to check "does this media exist at all" and obtain its destination root, independent of the "missing" endpoints.
- Guarantee a forced download's file can never silently collide with a sibling occurrence's file already on disk.
- Keep the automatic `radarr`/`sonarr` CLI pipeline's behavior and file naming completely unchanged.

**Non-Goals:**
- Building a generic job queue or worker pool for the API server; a single background goroutine per request is enough for this manual, low-frequency action.
- Changing how the automatic pipeline selects or orders quality candidates.
- Persisting Radarr/Sonarr internal IDs long-term (movie/series existence is re-checked live on every forced-download request, per the spec's "no pre-check on passive viewing" requirement).
- Reworking `resume-downloads` beyond the one path-resolution fix described below.

## Decisions

### 1. One endpoint, addressed by `ProcessedLine` id
`POST /api/v1/items/:id/force-download`, where `:id` is the same `ProcessedLine` id already used by `/api/v1/items/:id` and `/api/v1/items/:id/override`. This keeps the existing per-occurrence addressing convention and requires no new identifier scheme. The handler synchronously runs eligibility + the live existence check (fast, local DB reads plus one Radarr/Sonarr HTTP call under a short timeout), then hands the actual transfer to a background goroutine and returns `202 Accepted` immediately. It returns a 4xx (see status mapping in tasks) without any side effect when eligibility or the existence check fails.

*Alternative considered*: a generic `/api/v1/downloads` POST that takes a `processed_line_id` body field. Rejected only for consistency with the existing `:id`-scoped item routes; behaviorally equivalent.

### 2. New Radarr/Sonarr client methods, existence-only
- Radarr: `GetMovieByTMDBID(ctx, tmdbID) (*Movie, error)`, calling `GET /api/v3/movie?tmdbId=`. Local `Movie.TMDBID` is always populated, so this is sufficient without a TVDB fallback (unlike the missing-list matching path, which goes the other direction and needs both).
- Sonarr: `GetSeriesByTVDBID(ctx, tvdbID) (*Series, error)` (`GET /api/v3/series?tvdbId=`), then `GetEpisodesBySeriesID(ctx, seriesID) ([]Episode, error)` (`GET /api/v3/episode?seriesId=`) filtered client-side by season/episode number — Sonarr has no direct season+episode-number query, so this mirrors how `matcher.MatchEpisode` already compares `episode.SeasonNumber`/`EpisodeNumber`.
- Both return a distinguishable "not found" outcome (empty result array is not an error) versus a transport/HTTP error, so the handler can treat "not found" and "check failed" the same way (refuse) while still logging them differently.
- `TVShow.TVDBID` is nullable (only backfilled asynchronously, see `tvdb-id-backfill`). When nil, the Sonarr existence check cannot run at all and the request is refused — this is the fail-closed behavior the spec already mandates for an indeterminate result, not a special case to code around.

### 3. Destination naming: extract and extend, don't duplicate
`buildRadarrDestPath`/`buildSonarrDestPath`/`sanitizeFilename` currently live in `cmd/format.go` (package `main`), unusable from `internal/api`. They move into an internal package (e.g. `internal/downloader` or a small new `internal/pathbuilder`) so both the CLI commands and the new handler call the same code; `cmd/format.go`'s existing call sites are updated accordingly. A new variant/parameter adds the resolution suffix (or a unique fallback marker when resolution is `NULL`) and is used **only** by the force-download path — the automatic pipeline keeps calling the unsuffixed form, so its naming is byte-for-byte unchanged.

*Alternative considered*: duplicating a small suffixed-naming helper directly inside `internal/api` instead of touching `cmd/format.go`. Rejected — it would fork the sanitization/path-joining rules and the two would drift apart over time.

### 4. Resume must not recompute a different path for a forced download
`resume_helper.buildBaseDestPath` currently recomputes the destination base path from `line.Movie`/`line.TVShow` metadata first, and only falls back to the already-persisted `download.DownloadPath` when the line has no matched movie/TV show. For a forced download this ordering is wrong: if the server process restarts mid-transfer, the generic `resume-downloads` job would recompute the plain (unsuffixed) name and resume into a path that no longer matches the one the force-download handler chose — reintroducing exactly the collision risk this change exists to prevent.

Fix: reorder `buildBaseDestPath` to prefer a non-empty `download.DownloadPath`-derived base whenever one is already recorded, falling back to recomputing from `line.Movie`/`line.TVShow` only when no path was ever chosen (i.e., the download never got far enough to have one). Concretely, the force-download handler persists the resolution-suffixed base path onto `DownloadInfo` before the transfer starts, exactly as the normal flow already persists it once headers are known — just earlier. This is the one behavioral change to the existing resume path; it is a strict widening (prefer known truth over recomputed convention) and does not change resume's outcome for any download that never had a path recorded yet.

### 5. Eligibility ordering inside the handler
Cheapest/local checks first, so a request that's obviously wrong never reaches Radarr/Sonarr: (1) `ProcessedLine` exists, (2) has `movie_id`/`tv_show_id`, (3) its own `state` isn't already `downloaded`/`downloading`. Only then the live existence check runs. This matches the spec's "no existence check outside an actual trigger" requirement and keeps failed requests cheap.

## Risks / Trade-offs

- **[Risk]** A forced download can still race the automatic pipeline's own run for the same `ProcessedLine` (e.g., a scheduled `radarr` job and a manual click landing at the same moment) → **Mitigation**: both paths already go through `Downloader.Download()`'s existing per-`DownloadInfo` lock (`AcquireLock`/`ReleaseLock`); the second caller is rejected exactly like today's concurrent-download protection. No new locking needed.
- **[Risk]** `TVShow.TVDBID` being nil blocks forced downloads for TV episodes not yet backfilled → **Mitigation**: none beyond documenting; this is the intended fail-closed behavior, and the existing `tvdb-id-backfill` job is what resolves it over time.
- **[Risk]** Extracting `cmd/format.go`'s path builders into an internal package touches code exercised by `cmd/pathutil_test.go` → **Mitigation**: pure move + thin wrapper, existing tests move or get a one-line import update; no behavior change for the unsuffixed callers.
- **[Trade-off]** Running the transfer in a goroutine owned by the HTTP server process (rather than a separate worker) means a server redeploy can interrupt an in-flight forced download → accepted, because `resume-downloads` already exists to pick it back up (once Decision 4's fix is in place) and this is a manual, occasional action, not a high-throughput path.

## Migration Plan

Purely additive: new route, new client methods, one reordering fix in `resume_helper`, and a moved (not removed) path-building helper. No data migration. No feature flag needed — the new endpoint is inert until called, and the sidepanel action only appears for eligible occurrences. Rollback is a plain revert.

## Open Questions

- Exact HTTP status codes for each refusal reason (not matched / already downloaded / already in progress / media not found / existence check failed) — pick a coherent mapping during implementation; it doesn't change the observable behavior the spec describes.
- Exact resolution-suffix string format (e.g. `[1080p]` vs ` - 1080p`) and the fallback marker used when resolution is `NULL` — cosmetic, doesn't affect the anti-collision guarantee.
