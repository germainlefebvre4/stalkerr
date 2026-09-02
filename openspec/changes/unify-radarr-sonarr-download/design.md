## Context

See `proposal.md` - Why. Relevant current-state facts that shape this design:

- `cmd/radarr.go` and `cmd/sonarr.go` each fetch missing content, match it against the DB, then download **sequentially, one item at a time** — the `--parallel` flag they expose is parsed but never wired to any concurrency primitive.
- `internal/downloader/parallel.go` already provides a working N-worker pool (`ParallelDownloader.DownloadBatch`) consumed by `stalkeer resume-downloads`, but that consumer just flattens all incomplete downloads into one job list with no grouping — any two workers can end up processing two episodes of the same season simultaneously, or draining a whole show before touching anything else, purely by accident of list order.
- Sonarr's missing-episodes endpoint is already sorted `series.sortTitle, season ASC, episode ASC`, so per-series episode order is already correct; nothing currently derives "earliest incomplete season" as a schedulable unit or coordinates that with Radarr's movie list.
- `internal/downloader.StateManager` already has DB-backed per-download locks (`AcquireLock`/`ReleaseLock`, `CleanupStaleLocks`) used by the resume path; the new scheduler can reuse this rather than invent its own locking.
- `downloads.max_parallel` (config) is read independently by each of today's four cronjobs; because they never run concurrently with each other in practice (offset schedules), the shared-budget problem hasn't surfaced yet, but it's the whole point of this change.

## Goals / Non-Goals

**Goals:**
- One process, one worker pool, one concurrency knob shared across Radarr- and Sonarr-originated work.
- Deterministic *invariants* (season order, season completion) with non-deterministic *selection* (which stream starts next).
- Reuse existing building blocks (`ParallelDownloader`'s worker-pool shape, `StateManager` locks, `matcher` queries, `ResumeHelper` concepts) rather than rewriting download mechanics.

**Non-Goals:**
- Not changing how a single file is downloaded (retry, resume-via-Range, disk-space checks, atomic rename) — that machinery in `internal/downloader/downloader.go` is unaffected.
- Not changing `m3u-download` or `process` cronjobs.
- Not adding a UI/API surface for the scheduler; it remains a batch CLI command.
- Not preserving `--series-id`-style single-series targeting from the CLI (per proposal, explicitly dropped).

## Decisions

### 1. New `internal/scheduler` package, built on top of `internal/downloader`
Rather than growing `internal/downloader` (already large: downloader, parallel, resume, state manager, destpath...) or cramming stream logic into `cmd/download.go`, add `internal/scheduler` with:
- `Stream` — an ordered list of download-able items (movie: 1 item; series-season: N episodes) plus a `Tier` (1 = missing, 2 = force/upgrade) and a source key (series ID, or movie ID) used to enforce "one claimable stream per series."
- `Scheduler` — holds the mutable pool of claimable streams behind a mutex, exposes `ClaimNext() (*Stream, bool)` (does the tier-weighted, then uniform-random draw) and `Release(seriesID)` (called after a series-season stream fully drains, to compute and expose that series' *next* season as claimable, per the ascending-order rule).
- The worker loop itself (N goroutines calling `ClaimNext`, draining, calling `Release`, looping until no streams remain) lives in `cmd/download.go` or a small `scheduler.Run(ctx, n)` helper — thin either way, since the interesting logic is claim/release, not the loop.

Alternative considered: extend `ParallelDownloader` to accept a "job source" interface instead of a flat `[]DownloadJob`. Rejected because `ParallelDownloader` is also used, unmodified, by `resume-downloads`, and conflating "flat batch of independent jobs" with "stateful stream claiming" would complicate both call sites for no shared benefit.

### 2. Stream construction happens once, up front, per run
At the start of a run: fetch missing movies (Radarr) and missing episodes (Sonarr), fetch incomplete/interrupted downloads (`StateManager.GetIncompleteDownloads`), fetch force/upgrade candidates (previously-downloaded content still eligible per existing `--force` semantics). Build the full stream set before starting workers, rather than re-querying mid-run.
- Simpler to reason about and test (pure function: inputs -> `[]Stream`).
- Matches current behavior (today's commands also fetch once per invocation); newly-missing content that appears mid-run is naturally picked up by the *next* scheduled run, same as today.
- Incomplete downloads are merged into their owning stream (matched by series+season or movie) at construction time, at the front of that stream's item list, so "resume" is just "this item happens to already have a download record" rather than a separate code path.

### 3. Tier-weighted random draw, not strict priority
`ClaimNext` draws a uniform random number; with probability `p` (config, default proposed 0.1) it draws from tier 2 if tier 2 has any claimable stream, otherwise (or with probability `1-p`) it draws uniformly at random from tier 1. This is O(1) per draw given tier 1/tier 2 kept as separate slices with swap-remove. Alternative considered — interleave tier 2 items into tier 1 at a fixed ratio (e.g., "every 8th pick") — rejected as it produces a visible, predictable pattern rather than the "un peu d'aléatoire" the user asked for, and complicates resuming a partially-consumed ratio across a run that ends early (concurrencyPolicy: Forbid could kill a run mid-way on the next schedule tick... actually it just skips the next tick; the run itself completes or fails, no mid-run kill by k8s).

### 4. "Earliest incomplete season" is recomputed from remaining items, not cached
A series' claimable season is derived at construction time (and again at `Release`) as `min(season number)` among that series' episodes still in a non-terminal state. No separate "current season pointer" is persisted — it falls out naturally from which episodes are still missing/incomplete, so there's no extra state to keep in sync with the DB.

### 5. `--force` becomes tier classification, not a CLI flag
Today's `--force` (re-download already-downloaded content) becomes: at construction time, always compute tier-2 candidates (already-downloaded content whose files could be upgraded — same eligibility rule as today's `--force`), and let the weighted draw decide if/when they're attempted. There is no `--force` flag on `stalkeer download`; every run considers both tiers. A `--dry-run` and `--verbose` flag are kept (same meaning as today).

### 6. CLI surface: single `stalkeer download` command replaces `radarr`/`sonarr`
`cmd/download.go` replaces `cmd/radarr.go` and `cmd/sonarr.go`. Flags kept: `--dry-run`, `--limit` (caps total streams considered, applied before the random draw so it still respects season-order/tier rules on the subset), `--parallel` (overrides `downloads.max_parallel` for this run), `--verbose`. Flags dropped: `--force` (see above), `--series-id` (per proposal, no replacement).

## Risks / Trade-offs

- **[Risk]** A mutex-guarded scheduler shared by N goroutines is new concurrent code, easy to get subtly wrong (e.g., a series double-claimed). → Mitigation: keep `Scheduler`'s state mutation (claim/release) behind a single mutex with no I/O inside the critical section; cover with concurrent-access unit tests (N workers hammering `ClaimNext`/`Release` against a fixed stream set, asserting no double-claim and full drain).
- **[Risk]** Accepted trade-off from earlier discussion: strict one-worker-per-series means a run with many series each missing only 1-2 episodes can under-use the configured parallelism. → Mitigation: none needed — explicitly accepted, and the IPTV backend's 2-3 concurrent limit means "full utilization" was never the binding constraint anyway.
- **[Risk]** Folding force/upgrade into every run (vs. a once-daily separate force cronjob) means tier-2 work competes for the same run's time budget as tier-1; on a very large backlog, tier-2 throughput could be lower than today's dedicated force runs. → Mitigation: `p` is a config knob, tunable after observing real run behavior; can be revisited without a spec change (implementation detail).
- **[Risk]** Removing `--series-id` removes a manual debugging affordance. → Mitigation: accepted per user decision; `--verbose` plus direct DB/log inspection remains available.
- **[Trade-off]** Building all streams up front (Decision 2) means a run's total work is fixed at start; a long-running run won't pick up content that becomes missing partway through. Acceptable since this matches today's per-invocation behavior and the cronjob repeats regularly.

## Migration Plan

1. Implement `internal/scheduler`, `cmd/download.go`; remove `cmd/radarr.go`, `cmd/sonarr.go`.
2. Add `charts/stalkerr/templates/cronjob-download.yaml`; remove the four `cronjob-radarr-sync*.yaml` / `cronjob-sonarr-sync*.yaml` templates; update `values.yaml` (`jobs.download` replacing `jobs.radarrSync`/`jobs.sonarrSync`).
3. Deploy: a Helm upgrade naturally deletes the four old CronJob resources (no longer templated) and creates the one new one; no data migration needed (same DB models, same `ProcessedLine`/`DownloadInfo` tables).
4. Rollback: revert the Helm chart change and the binary; the old four-cronjob shape and old CLI commands come back unchanged since no schema changed.

## Open Questions

- Exact default value for the tier-2 draw probability `p` (proposal/design assume ~0.1; final value is a tuning detail, not a spec-affecting decision) - to settle during implementation by observing a real run's tier-2 throughput.
