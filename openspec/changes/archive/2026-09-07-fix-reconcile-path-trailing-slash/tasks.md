## 1. Core fix

- [x] 1.1 In `computeReconciledPath` (`internal/api/resync_path.go`), normalize `targetRoot` and `root.Root` with `filepath.Clean` before the equality check, so a trailing path separator on either side no longer triggers a false "changed" result. Verify by inspection that the existing `filepath.Join(targetRoot, root.SubPath)` return value is unaffected for the already-passing no-trailing-separator cases.

## 2. Tests

- [x] 2.1 Add `TestComputeReconciledPath_TargetRootTrailingSeparatorIsNoOp` in `internal/api/resync_path_test.go`, alongside the existing `TestComputeReconciledPath_*` tests: assert `computeReconciledPath` returns `changed == false` when `targetRoot` is `root.Root` plus a trailing `/`.
- [x] 2.2 Add `TestComputeReconciledPath_TrailingSeparatorDoesNotMaskRealDrift` in the same file: assert that when `targetRoot` (with or without a trailing separator) points at a genuinely different directory than `root.Root`, `computeReconciledPath` still returns `changed == true` with the correct cleaned destination path — confirming the fix doesn't weaken real-drift detection.
- [x] 2.3 Add `TestReconcileScheduledDownloadPaths_TrailingSeparatorNoOp` alongside `TestReconcileScheduledDownloadPaths_CorrectsMatchedRow`: seed a completed download whose stored root matches a fake Sonarr/Radarr library fixture's `Path` except for a trailing separator, run `ReconcileScheduledDownloadPaths`, and assert `download_path` is left unchanged (no move attempted, no `rename_target_exists`/`rename_failed` outcome).

## 3. Verification

- [x] 3.1 Run `go test ./internal/api/...` and confirm all tests pass, including the new ones from section 2.
- [x] 3.2 Run the full backend test suite (`go test ./...`) and confirm no regressions elsewhere.
