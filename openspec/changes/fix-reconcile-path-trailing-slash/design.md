## Context

See `proposal.md` for motivation. The relevant code is `computeReconciledPath` in `internal/api/resync_path.go`:

```go
func computeReconciledPath(root downloadRoot, targetRoot string) (newPath string, changed bool) {
	if targetRoot == "" || targetRoot == root.Root {
		return "", false
	}
	return filepath.Join(targetRoot, root.SubPath), true
}
```

`root.Root` is always produced by `extractDownloadRoot` via `filepath.Dir` (movie) or `detectTVSeasonPath`'s `filepath.Dir(filepath.Dir(...))` (TV) — both of which already return a clean path with no trailing separator. `targetRoot` comes straight from the Radarr/Sonarr API response (`movie.Path` / `series.Path`) with no processing. Production data confirms 7 of 97 Sonarr series currently have a `Path` ending in `/` (a pre-existing artifact of how those series' root folders were configured in Sonarr itself); Radarr's `movie.Path` values in the same environment have no such artifact today, but nothing in the Radarr API contract guarantees that will always hold.

## Goals / Non-Goals

**Goals:**
- Make the root-directory comparison treat a trailing separator as insignificant, on either input, without weakening detection of an actual root change.
- Keep the fix confined to the comparison itself; the shared move/DB-update logic in `applyPathCorrection` is unaffected and already correct.

**Non-Goals:**
- Fixing the trailing slash at its source in Sonarr's own configuration — out of the user's/Stalkeer's control, and the comparison should be robust regardless of whether Radarr/Sonarr ever supplies one.
- Handling any other form of path drift (case differences, symlinks, `..` segments) — no evidence of these in production data; not worth the added complexity now.
- The separate `rename_failed` movie cluster (title drift + duplicate downloads) — explicitly out of scope per proposal.md.

## Decisions

**Normalize with `filepath.Clean`, applied to `targetRoot` (and, for defense-in-depth, `root.Root`) immediately before the equality check in `computeReconciledPath`.** `filepath.Clean` is already what `filepath.Dir` itself uses internally to produce `root.Root`, so cleaning `targetRoot` the same way guarantees both sides are compared in the same normal form. It's a pure string operation (no filesystem access), already imported in this file, and handles the trailing-separator case along with the other insignificant variations `Clean` normalizes (double separators, `./` segments), for negligible extra cost on a per-run, not per-item, string.

Alternative considered: `strings.TrimRight(targetRoot, "/")` — rejected as narrower (only handles the one case seen in production) and reimplements a subset of what `filepath.Clean` already guarantees correct.

Alternative considered: normalize once when building the `map[tmdbID]Path`/`map[tvdbID]Path` lookups in `cmd/download.go`, instead of inside `computeReconciledPath`. Rejected: `computeReconciledPath` is the single shared choke point already used by both the scheduled pass and the on-demand `resync-path` endpoint (which builds its `targetRoot` from a separate live lookup, not the map). Normalizing there fixes both call sites at once instead of requiring the same care to be applied independently wherever a `targetRoot` is produced.

**Use `filepath.Join(targetRoot, root.SubPath)` unchanged for the corrected path.** Once `targetRoot` is compared in clean form, `filepath.Join` (which already calls `Clean` on its result) naturally produces the same clean path whether or not the original `targetRoot` had a trailing separator — no further change needed there.

## Risks / Trade-offs

- **[Risk] `filepath.Clean` also silently absorbs other malformed input** (e.g., a `Path` with a redundant `../` segment) that might have been worth surfacing as an error instead of treated as equal. → Accepted: no such case has been observed in production data, `filepath.Clean`'s normalization rules are the same ones Go's own standard library and this codebase's existing `filepath.Dir` calls already rely on implicitly, and treating it as a non-issue keeps the fix minimal and low-risk.
- **[Trade-off] The fix does not address the `rename_failed` movie cluster** — accepted per proposal.md's scope; that cluster needs its own investigation and is tracked separately.

## Migration Plan

No schema change, no data migration. Ships as a single logic change to `computeReconciledPath`; takes effect on the next scheduled `download` run after deploy, at which point the 7 currently-affected series stop being compared as "drifted" and the recurring warnings for them stop immediately (no backlog to clear, since their stored `download_path` values are already correct).
