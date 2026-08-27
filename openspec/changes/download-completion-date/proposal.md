## Why

The Downloads tab shows technical specs (format, resolution, size, duration) and status badges for each download, but no timestamp. Users have no way to tell when a completed download actually finished without checking logs or the database. The Playlist tab already shows this kind of date (`downloaded_at`) for its items; Downloads should offer the same visibility for consistency.

## What Changes

- Expose the download's completion timestamp (`completed_at`) through the enriched downloads API response, which currently only exposes `updated_at`.
- Render the completion date on each download card in the Downloads tab, inline in the existing technical specs row (alongside Format • Resolution • Size • Duration), formatted with the same locale-aware `formatDate` utility used by the Playlist tab.
- The date only appears when the download has actually completed (`completed_at` is set); for pending/downloading/failed/retrying downloads, the specs row is unchanged (no placeholder date, no extra bullet).
- Applies identically on mobile and desktop — the Downloads tab is a single responsive component (CSS breakpoint at 768px governs density), not separate implementations.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `downloads-display-ui`: adds a requirement that completed download cards display their completion date inline in the technical specs row.

## Impact

- Backend: `internal/api/types.go` (`DownloadEnrichedResponse`), `internal/api/handlers_frontend.go` (`enrichDownloadInfo`).
- Frontend: `frontend/src/types.ts` (`DownloadEnriched`), `frontend/src/components/DownloadsTab.tsx`, `frontend/src/locales/{fr,en}/downloads.json`.
- No database schema change — `DownloadInfo.CompletedAt` already exists and is already populated on completion (`internal/downloader/downloader.go`).
