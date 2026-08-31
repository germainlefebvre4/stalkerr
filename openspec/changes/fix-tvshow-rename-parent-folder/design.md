## Context

`internal/api/handlers_frontend.go` already has a correct, tested TV-season detector, `detectTVSeasonPath(downloadPath string) (*tvSeasonPathInfo, bool)`, used by `computeRenameDestination` to resolve the series root when a download's parent directory matches `Season NN`. The enrichment endpoint (`listDownloadsEnriched`, same package) currently derives `file_info.folder_name` only via `fileparser.Parse`, which takes the file's immediate parent directory basename — correct for movies, wrong (season folder) for TV episodes. See proposal.md - Why.

## Goals / Non-Goals

**Goals:**
- Give the frontend a single field to pre-fill the rename dialog that is always correct, for both movies and TV episodes.
- Reuse the existing, already-tested `detectTVSeasonPath` rather than reimplementing season detection in a second place (Go or TypeScript).

**Non-Goals:**
- Changing the rename endpoint's own resolution logic (`computeRenameDestination`) — it is already correct.
- Changing what `file_info.folder_name` means or how it's displayed elsewhere in the Downloads list (it stays the literal immediate parent directory, used for the plain folder/file display line).

## Decisions

**New field name and placement: `rename_folder_name` as a top-level field on `DownloadEnrichedResponse`, sibling to `file_info`.**
Alternative considered: nest it inside `FileInfo` (e.g. `file_info.rename_folder_name`). Rejected because `FileInfo` is produced entirely by the generic, context-free `fileparser.Parse` (also used by `file-metadata-parsing` elsewhere); mixing in a value that depends on TV/movie-specific path conventions would blur that boundary. A top-level field on the enrichment response keeps `fileparser` untouched and makes the rename-specific computation visible as what it is.

**Computation: extract the series-root-name lookup already inside `detectTVSeasonPath` into a small helper reused by both `computeRenameDestination` and the enrichment loop, rather than duplicating the `Season NN` prefix/parse logic.**
Alternative considered: duplicate the check inline in the enrichment loop. Rejected — the whole point of this fix is that the season-detection logic must only exist once so it can't drift out of sync with the rename endpoint again.

**Omission semantics mirror `file_info`: `rename_folder_name` is omitted (not empty string) when `download_path` is null.**
Keeps the contract consistent with the existing `file_info,omitempty` behavior and avoids the frontend having to special-case an empty string vs. absent field.

## Risks / Trade-offs

- [Risk] Frontend forgets to switch and keeps reading `file_info.folder_name` → Mitigation: `downloads-display-ui` spec's `MODIFIED Requirements` makes the new source the normative pre-fill contract, and `RenameFolderDialog.tsx`'s existing behavior (pre-fill from `renameItem.folderName`) is unchanged — only the value `App.tsx` passes into `folderName` changes, so no dialog-side edits are needed.
- [Risk] A download whose path happens to have an immediate parent directory literally named `Season <N>` but is not actually a TV episode (e.g. a movie oddly placed in a folder called "Season 01") would be misclassified → Mitigation: this is the same heuristic the rename endpoint itself already relies on (`detectTVSeasonPath`); staying consistent with it is preferable to introducing a second, different heuristic.
