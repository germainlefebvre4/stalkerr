## 1. Dedup tier-2 streams against tier-1

- [x] 1.1 In `internal/scheduler/build.go`, after `mergeIncompleteDownloads` runs (so any tier-1 stream it synthesizes is included), filter tier-2 streams against the existing typed `movieByID` and `tvdbSeasonStreams` maps (not raw `SourceKey` string equality — the tier-1/tier-2 series `SourceKey` formats use different ID spaces, see design.md): drop any tier-2 movie stream whose `SourceKey`'s movie ID is a key in `movieByID`, and drop any tier-2 series stream whose `(tvdbID, season)` — formatted as tier-2's own `series:tvdb:<id>:season:<n>` — matches a key in `tvdbSeasonStreams`, before `BuildStreams` returns.
- [x] 1.2 Verify with `go build ./...` that the package still compiles.

## 2. Test coverage

- [x] 2.1 Add a unit/integration test in `internal/scheduler/build_test.go` reproducing the reported bug: a movie present in Radarr's missing list AND with a `downloaded` `ProcessedLine` plus another eligible candidate in the DB - verify `BuildStreams` returns exactly one stream with `SourceKey` `movie:<id>` (the tier-1 one), not two.
- [x] 2.2 Add the equivalent test for the series-season case: a series-season with a missing episode per Sonarr AND a `downloaded` episode plus another eligible candidate in the same season - verify `BuildStreams` returns exactly one stream for that `SourceKey`.
- [x] 2.3 Add a test covering the `mergeIncompleteDownloads`-synthesis ordering case from design.md: a movie with an existing tier-2 stream (already downloaded, upgrade-eligible) that also has a stuck/incomplete `DownloadInfo` for a *different* candidate not covered by tier-1 or tier-2 fetch - verify the synthesized tier-1 stream wins and the pre-existing tier-2 stream for the same `SourceKey` is dropped.
- [x] 2.4 Run `go test ./internal/scheduler/...` and verify all tests pass, including the existing `TestBuildStreams_TierClassification` and `TestBuildStreams_FullIntegration`.

## 3. Verification

- [x] 3.1 Run `go test ./...` and verify the full suite still passes (no regression in other packages consuming `scheduler.Stream`/`BuildStreams`). `internal/processor` fails with pre-existing, unrelated "database.user is required" config errors, confirmed present before this change (`git stash` + rerun) - not a regression. All other packages, including `internal/scheduler`, pass.
- [x] 3.2 Manually confirm via `go run . download --dry-run` (or equivalent) against a dataset with a known missing+already-downloaded movie that only one entry is printed for it in the dry-run plan. Confirmed by user: dry-run output shows exactly one entry per `SourceKey` (`series:tvdb:73838`, `movie:14`), no duplicates.
