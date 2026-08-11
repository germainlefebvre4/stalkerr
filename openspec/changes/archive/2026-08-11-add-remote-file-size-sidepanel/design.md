## Context

See `proposal.md` - Why. `ProcessedLine` (`internal/models/processed_line.go`) has no notion of remote file size today. The closest existing pattern is `internal/downloader/resume.go:CheckServerSupport`, which does a one-off `HEAD` before an actual download to check `Accept-Ranges`/`Content-Length`, but it's only invoked at download time, never exposed via API/UI, and never persisted.

The backfill for TMDB rich metadata (`internal/processor/backfill_metadata.go`, wired into `Process()` and reported in `cmd/process.go`) is the direct structural precedent for "automatically complete missing data during `process`, batched, tolerant of individual failures." That backfill targets a fast, reliable API (TMDB, itself rate-limited). Probing arbitrary IPTV-provided URLs is a different risk profile: slow, sometimes unreachable, sometimes actively hostile to `HEAD` requests, and a provider could throttle/ban an IP that hammers it. The design below reuses the backfill *shape* but adds guardrails specific to that risk (bounded timeouts, bounded per-run volume, single-attempt semantics).

## Goals / Non-Goals

**Goals:**
- Reuse the `BackfillRichMetadata`-style batched backfill shape, wired into the same `Process()` call site, for consistency and minimal new plumbing.
- Never let a single unresponsive/hostile IPTV endpoint stall or fail a `process` run.
- Never re-probe a line once an attempt (successful or not) has been recorded, per the confirmed "compute once" decision.
- Keep the per-run probing volume bounded and configurable, so the initial backlog (potentially thousands of pre-existing lines) drains over multiple daily `process` runs instead of extending one run indefinitely.

**Non-Goals:**
- No periodic refresh/re-check of a line's remote file size once checked (explicitly out of scope per the confirmed decision).
- No isolation into a separate command/CronJob — explicitly grafted into `process`, at the accepted cost of coupling the main ingestion run to IPTV probing reliability (mitigated by the guardrails above).
- No size probing for `channels` (live streams, no fixed size).
- No UI-triggered on-demand re-probe or manual "refresh size" action in this change.

## Decisions

**Two new nullable columns on `ProcessedLine`: `remote_file_size *int64` and `remote_file_size_checked_at *time.Time`.**
A single `remote_file_size` column can't distinguish "never attempted" from "attempted, unavailable" — both are `NULL`. Since the backfill query selects on "never attempted" and never retries, that distinction has to be persisted, or a permanently-unreachable line would be re-probed on every `process` run forever, defeating the "compute once" decision and continuing to hammer a hostile provider. Alternative considered: a sentinel value like `-1` for "checked, unavailable" — rejected because it overloads a byte-count field with non-numeric meaning and complicates API/UI consumers that would need to special-case it.

**Probing: `HEAD` first, ranged `GET` (`Range: bytes=0-0`) fallback, reading `Content-Length`/`Content-Range`.**
Many IPTV/Xtream-style endpoints reject or mishandle `HEAD`. A minimal ranged `GET` is the standard fallback to obtain the resource's total size without downloading it, mirroring the intent of `internal/downloader/resume.go:CheckServerSupport` (which only checks `Accept-Ranges`, not size, and isn't reused directly since it doesn't return a byte count). New code lives in the processor package rather than extending `resume.go`, since the downloader's resume support is about resumable partial downloads, a different concern than a one-off size probe.

**New backfill function grafted into `Process()`, mirroring `BackfillRichMetadata`.**
Confirmed by the user: coupling to the main `process` pipeline is accepted, in exchange for simplicity (no new CronJob/command/scheduling to maintain). The guardrails below exist specifically to bound the blast radius of that coupling.

**Per-run cap on probed lines, configurable with a sane default.**
Confirmed by the user. Mirrors the existing `batchSize`/`limit` flag pattern already used by `process` (`cmd/process.go`). A default in the low hundreds keeps a run's added duration predictable even against a large pre-existing backlog, which drains across subsequent daily runs.

**Bounded per-request timeout (short, e.g. single-digit seconds) on both the `HEAD` and fallback `GET`.**
Without this, one slow/unresponsive IPTV host could stall the entire backfill step (and therefore the `process` run) regardless of the per-run row cap.

**Scope restricted to `content_type IN (movies, tvshows, uncategorized)`.**
Confirmed by the user: `channels` are excluded entirely from the backfill query, never probed, size never displayed for them.

## Risks / Trade-offs

- **[Risk] A `process` run's duration becomes variable/unpredictable due to network probing of third-party servers.** → Mitigated by the per-run row cap and the per-request timeout; worst case added duration is bounded by `cap × timeout`.
- **[Risk] Probing many IPTV URLs from the same server IP in quick succession could trigger provider-side rate limiting or bans, potentially affecting the actual playlist download itself.** → Mitigated by the per-run cap (the flagged trade-off of grafting into `process` instead of an isolated job); probing happens sequentially rather than in a concurrent burst, in the same spirit as the existing TMDB rate-limited client.
- **[Risk] A file's size is captured once and never refreshed; if an IPTV provider replaces a file with a re-encoded version, the displayed size becomes stale.** → Accepted per the confirmed "compute once" decision; out of scope for this change.
- **[Trade-off] Coupling a network-dependent, best-effort feature into the core ingestion pipeline (`process`) rather than isolating it.** → Accepted by the user; resilience requirements (per-line failure isolation, bounded timeouts, single-attempt semantics) exist specifically to contain this trade-off's downside.

## Migration Plan

- Auto-migration adds the two nullable columns to `processed_lines`; existing rows are unaffected (`remote_file_size_checked_at` starts `NULL` for all of them, meaning the entire existing backlog becomes eligible for backfill starting from the first `process` run after deploy).
- No backward-incompatible API change: `remote_file_size` is a new optional field on `ItemResponse`/`PlaylistItem`, additive only.
- Rollback: reverting the code leaves the two extra columns in place (harmless, unused) unless a down-migration is explicitly authored; no data migration is required either way since the columns are purely additive.
