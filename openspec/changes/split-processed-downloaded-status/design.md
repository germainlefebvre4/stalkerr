## Context

See `proposal.md` - Why. Today the Playlist page renders a single badge from `PlaylistItem.state` (`frontend/src/utils/pipelineState.ts`, consumed by `PlaylistItemsTable.tsx` and `PlaylistTab.tsx`'s sidepanel). `state` is `ProcessedLine.State` on the backend, a single enum column (`pending`/`processed`/`downloading`/`organizing`/`downloaded`/`failed`) that many backend consumers already depend on as-is: the matcher (`internal/matcher/matcher.go`), stats (`by_state` in `getStats`), force-download eligibility (`internal/api/force_download.go`), and maintenance/reset logic (`internal/database/maintenance.go`). `/api/v1/items` does not currently join `DownloadInfo` at all.

Because processing (TMDB matching) always completes before any download is attempted, `state` already encodes both facts in one monotonic value: any state other than `pending` implies the item was successfully processed. This lets us derive both display facets from the single existing field without touching the backend.

## Goals / Non-Goals

**Goals:**
- Show processing status and download status as two independent, simultaneously-visible indicators on the Playlist table, mobile list card, and sidepanel.
- Keep the mobile list card compact (icons, not two full text badges).
- Zero backend/API/data-model change.

**Non-Goals:**
- Cleaning up `ProcessedLine.State` itself (e.g. splitting it into two backend columns, or removing its download-phase values) - out of scope; too many existing consumers (matcher, stats, eligibility, maintenance) depend on its current values, and no backend behavior needs to change to satisfy this proposal.
- Changing force-download eligibility rules or any other logic keyed on `state`.
- Changing the Downloads tab (`DownloadInfo`-based), which already treats download status independently.

## Decisions

**Derive both statuses on the frontend from the existing `state` string; no new API field.**
Add two pure functions to `frontend/src/utils/pipelineState.ts`:
- `getProcessingStatus(state): 'pending' | 'processed'` - `'pending'` only when `state === 'pending'`, `'processed'` otherwise.
- `getDownloadStatus(state): 'not_downloaded' | 'downloading' | 'organizing' | 'downloaded' | 'failed'` - `'not_downloaded'` when `state` is `'pending'` or `'processed'`, otherwise pass through (`downloading`/`organizing`/`downloaded`/`failed`).

Alternative considered: preload `DownloadInfo` on `/api/v1/items` (mirroring the join already used by `listDownloadsEnriched` in `internal/api/handlers_frontend.go`) and expose a dedicated `download_status`/`error_message` field. Rejected for this change: it would add a backend join, a new response field, and migration surface for information the existing `state` field already determines unambiguously for display purposes. Revisit only if a future need arises for data `state` cannot express (e.g. showing the download `error_message` inline in the table, which today only the Downloads tab needs).

**Two badge classes, one new neutral style.**
Reuse existing CSS badge classes (`badge-success`, `badge-progress`, `badge-failed`, `badge-pending`) for `processed`/`downloading`/`organizing`/`failed`/`downloaded`. Add one new class (or reuse `badge-pending`'s neutral gray styling) for `not_downloaded`, used by both the sidepanel's explicit `[ non téléchargé ]` badge and, implicitly, the mobile icon's neutral color.

**Desktop table:** keep one "Pipeline state" column, render two small badges side by side inside it (processing badge, download badge), replacing the single badge.

**Mobile list card:** replace the single text badge with two small icon indicators (one per facet), each carrying a `title`/`aria-label` with the full status text for accessibility, keeping the card's single-line layout.

**Sidepanel:** in the "État du Pipeline d'Ingestion" grid, replace the single "État actuel" cell with two cells, "Traitement" and "Téléchargement", each a full-text badge - including the explicit `[ non téléchargé ]` badge when `getDownloadStatus` returns `not_downloaded`.

## Risks / Trade-offs

- **Assumption that processing always precedes downloading** → if that pipeline invariant ever changes (e.g. a future path reaches `downloading`/`failed` without passing through `processed`), the derived processing status would misreport "processed". Mitigation: this is already guaranteed by the current pipeline (matching happens in the processor before any download is triggered, and force-download requires an existing match); no code change needed now, just a documented assumption for future maintainers.
- **Two elements instead of one in constrained mobile width** → mitigated by using icon-only indicators (no text) on mobile, as already decided.
- **Missed i18n keys** → new labels needed in both `frontend/src/locales/fr/playlist.json` and `en/playlist.json`; enumerate exact keys in tasks.md to avoid a partial rollout.

## Migration Plan

Frontend-only change, no data migration, no backend deploy coordination needed. Ship as a normal frontend release.
