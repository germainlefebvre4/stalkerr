## Context

`BackfillRemoteFileSize` (`internal/processor/remote_file_size.go`) probes lines strictly sequentially, one HTTP HEAD (falling back to a ranged GET) at a time, capped at `per_run_cap` (default 200) per nightly `process` run. See proposal.md - Why for the observed throughput gap (~703 newly eligible lines/night vs 200 probed/night) and the 442 lines permanently stuck after a single failed probe. `remote_file_size_checked_at` is currently the sole gate for eligibility (`IS NULL`); there is no column tracking retry state beyond it. The table has ~190k eligible rows today and grows by ~700/night; the eligibility query has no supporting index.

## Goals / Non-Goals

**Goals:**
- Let the backfill sustainably outpace nightly ingestion so the "never checked" backlog shrinks over time instead of growing.
- Recover file sizes for lines that failed a probe due to a transient condition, without re-probing lines that already succeeded or that failed very recently.
- Keep the change entirely internal to the backfill/storage layer: no API or frontend changes, no new required config (existing defaults keep working if unset).

**Non-Goals:**
- Guaranteeing the backlog clears by a specific date — throughput tuning (cap/concurrency values) is an operational knob, not a spec-level commitment.
- Per-line retry counters, exponential backoff, or a dead-letter mechanism for lines that fail repeatedly across many cooldown cycles. A flat cooldown is enough given the observed failure rate (0.23% of probed lines).
- Changing the HEAD/range-GET probing strategy itself (out of scope; unrelated to the throughput problem).

## Decisions

**Concurrency via a bounded worker pool, not one goroutine per row.** `BackfillRemoteFileSize` will fan the selected rows out to a fixed-size pool (config `remote_file_size.concurrency`, default 10) instead of spawning unbounded goroutines. A semaphore channel (or `errgroup.SetLimit`) is enough — no new dependency needed since `errgroup` semantics can be replicated with a buffered channel and `sync.WaitGroup`. Alternative considered: increase `per_run_cap` alone without concurrency. Rejected — sequential probing at 200/run already costs meaningful wall time when many lines hit the 5s timeout; raising the cap without concurrency would multiply worst-case run duration linearly instead of dividing it by the pool size.

**Raise `per_run_cap` default from 200 to 2000.** At concurrency 10 with the existing 5s per-request timeout, worst case (every line times out on both HEAD and range-GET) is `(2000/10) * 10s` ≈ 33 min added to a nightly run; observed average is far lower since 99.5%+ of probes resolve well under the timeout. Net backlog reduction: `2000 - ~703 (nightly ingestion) ≈ 1300/night`, clearing the current ~190k backlog in roughly 5-6 months while keeping pace with ongoing ingestion indefinitely after that. Both `concurrency` and `per_run_cap` remain operator-tunable; the numbers above are starting defaults, not spec commitments (see Non-Goals).

**Retry gate: flat cooldown on `remote_file_size_checked_at`, no new column.** Eligibility becomes `remote_file_size IS NULL AND (remote_file_size_checked_at IS NULL OR remote_file_size_checked_at < now() - cooldown)`, with `remote_file_size_checked_at` continuing to double as "last attempt time" (it's already updated on every probe, success or failure). Config `remote_file_size.retry_cooldown_hours`, default 168 (7 days). Alternative considered: a dedicated `remote_file_size_attempts` counter with capped retries. Rejected as premature — failure rate is 0.23% and mostly transient (see proposal.md); a flat cooldown is simpler and the per_run_cap already bounds how much probe budget stale failures can consume relative to the much larger never-checked backlog.

**Add a composite index to support the new eligibility query.** Add `idx_processed_lines_remote_file_size_probe` on `(content_type, remote_file_size, remote_file_size_checked_at)` via GORM auto-migration. The current query has no supporting index and does a sequential scan over a table already at ~190k rows and growing; the new predicate (an `OR` with a timestamp comparison) makes an unindexed scan more expensive, not less.

## Risks / Trade-offs

- [Higher concurrency increases simultaneous outbound connections to third-party IPTV servers] → Mitigation: concurrency default (10) is modest and configurable per-deployment; per-request timeout is unchanged, so total open connections at any instant is bounded by `concurrency`, not `per_run_cap`.
- [Raising `per_run_cap` extends nightly `process` run duration, worst case ~33 min added] → Mitigation: worst case only occurs if most probes fail, which is not the observed pattern (0.23% failure rate); both knobs are independently tunable if the nightly window is tight.
- [Cooldown-based retry can still spin forever on a permanently dead URL (e.g., removed from the provider's catalog), spending probe budget every cooldown cycle indefinitely] → Mitigation: accepted per Non-Goals — the volume of such lines (442 today, 0.23%) is negligible against `per_run_cap`, and it self-corrects if a line's `content_type` changes or the row is pruned by existing catalog maintenance.
- [New composite index adds write overhead on every `ProcessedLine` update] → Mitigation: `remote_file_size`/`remote_file_size_checked_at` updates already happen once per line during backfill, not on the hot ingestion path; index maintenance cost there is negligible.

## Migration Plan

- Config: add `remote_file_size.concurrency` (default 10) and `remote_file_size.retry_cooldown_hours` (default 168) to `internal/config/config.go` defaults; raise `remote_file_size.per_run_cap` default to 2000. No changes required to deployed `config.yml` files — existing installs pick up new defaults automatically via Viper's default-value fallback, same as today.
- Schema: the new composite index is created by GORM auto-migration on next startup, same mechanism already used for the `remote_file_size` columns themselves (see `remote-file-size-storage` spec) — no manual migration step.
- Data: no backfill/migration needed for existing rows. The 442 already-failed lines naturally become eligible on the first run after deploy (their `remote_file_size_checked_at` values already predate the 7-day cooldown), and are absorbed into the same eligibility pool as the rest of the backlog — no special-casing required.
- Rollback: reverting the binary/config restores the old sequential, single-attempt behavior immediately; no data cleanup needed since no destructive change was made to existing columns.

## Open Questions

- Should `per_run_cap`/`concurrency` defaults be revisited once real nightly run durations are observed post-deploy? Deferred — doesn't change the spec, approach, or task breakdown; it's a config value adjustable without a new change.
