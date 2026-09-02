## 1. Scheduler core (`internal/scheduler`)

- [ ] 1.1 Define `Stream` (ordered items, `Tier`, source key: series ID or movie ID) and `Item` types; verify with a table-driven unit test constructing streams from fixture data.
- [ ] 1.2 Implement `Scheduler.ClaimNext()` with mutex-guarded tier-weighted draw (probability `p` from config, default 0.1) then uniform-random pick within the chosen tier; verify with a unit test asserting the empirical draw ratio over many iterations falls within tolerance of `p`, and that tier 1 is always chosen when tier 2 is empty.
- [ ] 1.3 Implement `Scheduler.Release(seriesID)` to compute and re-expose the series' next-earliest incomplete season as claimable once its current season stream is fully drained; verify with a unit test asserting season N+1 is not claimable until season N is fully released as complete.
- [ ] 1.4 Add a concurrent-access unit test: N goroutines calling `ClaimNext`/drain/`Release` in a loop against a fixed stream set, asserting no stream is claimed by two workers simultaneously and every stream is eventually drained exactly once.
- [ ] 1.5 Implement episode-ascending ordering within a season stream and movie-as-single-item streams; verify with a unit test on out-of-order input episodes.

## 2. Stream construction

- [ ] 2.1 Implement fetch-and-match for missing movies (Radarr) and missing episodes (Sonarr) reusing `internal/matcher` and the existing Radarr/Sonarr clients, returning tier-1 candidates; verify with a unit test using mocked API responses.
- [ ] 2.2 Implement tier-2 candidate construction (already-downloaded content eligible for re-download/upgrade, same eligibility as today's `--force`); verify with a unit test covering an already-downloaded item being classified tier 2 and a never-downloaded item being classified tier 1.
- [ ] 2.3 Merge `StateManager.GetIncompleteDownloads` results into their owning stream (matched by series+season or movie), placed ahead of not-yet-attempted items in that stream; verify with a unit test asserting a resumed item is attempted before a fresh item in the same stream.
- [ ] 2.4 Call `StateManager.CleanupStaleLocks` at the start of stream construction, before any stream is built; verify with a unit test/integration test that a stale lock is cleared before the corresponding item becomes claimable.
- [ ] 2.5 Wire construction into a single `BuildStreams(ctx, cfg)` entry point returning the full stream set for a run; verify with an integration-style test combining 2.1-2.4 against a seeded test DB.

## 3. CLI command

- [ ] 3.1 Create `cmd/download.go` with `stalkeer download` (flags: `--dry-run`, `--limit`, `--parallel`, `--verbose`; no `--force`, no `--series-id`), self-registering via `init()` per `cmd-entrypoints` conventions; verify `stalkeer download --help` lists exactly these flags.
- [ ] 3.2 Implement the worker-pool run loop (N goroutines calling `ClaimNext`/drain/`Release` until the scheduler reports no streams remain), with per-item download via the existing `internal/downloader.Downloader`; verify with an integration test using a small fixed stream set and a mock HTTP download server, asserting all items complete and summary stats are correct.
- [ ] 3.3 Implement `--dry-run` output (list planned streams/items without downloading) and `--limit` (caps total streams considered before the random draw); verify with unit tests for each flag.
- [ ] 3.4 Remove `cmd/radarr.go` and `cmd/sonarr.go` (and their now-orphaned tests, if any); verify `stalkeer radarr`/`stalkeer sonarr` no longer exist (`stalkeer --help` omits them) and `go build ./...` succeeds.
- [ ] 3.5 Add/adjust `downloads.force_tier_probability` (or equivalent name) to `internal/config` and `config.yml.example`; verify config loads with a default value when unset.

## 4. Helm chart

- [ ] 4.1 Add `charts/stalkerr/templates/cronjob-download.yaml` (command `["./stalkeer", "download", "--config", "/app/config/config.yaml"]`, both Radarr and Sonarr API key secrets, m3u/media/temp volumes); verify with `helm template` showing the rendered CronJob.
- [ ] 4.2 Remove `cronjob-radarr-sync.yaml`, `cronjob-radarr-sync-force.yaml`, `cronjob-sonarr-sync.yaml`, `cronjob-sonarr-sync-force.yaml`; verify `helm template` no longer renders any `radarr-sync*`/`sonarr-sync*` CronJob.
- [ ] 4.3 Update `values.yaml`: replace `jobs.radarrSync`/`jobs.sonarrSync` (and their `forceSync` sub-keys) with a single `jobs.download` block (schedule, concurrencyPolicy, resources, history limits); verify `helm lint` passes.
- [ ] 4.4 Update chart `NOTES.txt` / README references to the old four cronjobs, if any, to reference the single `download` cronjob instead; verify by grepping the chart directory for `radarr-sync`/`sonarr-sync` and finding no remaining references.

## 5. Spec-driven verification

- [ ] 5.1 Run through every scenario in `specs/media-download-scheduling/spec.md` against the implemented scheduler (unit/integration tests already cover most; add any missing scenario coverage) and confirm each passes.
- [ ] 5.2 Confirm the modified `scheduled-jobs`, `radarr-movie-matching`, and `m3u-quality-selection` scenarios still hold against the new command (matching/quality-fallback behavior unchanged, cronjob shape updated); verify via `helm template` (for scheduled-jobs) and the relevant existing matcher/quality-selection unit tests (updated to reference the unified command where they assert on it).
