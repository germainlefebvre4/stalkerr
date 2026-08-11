## Why

Users reviewing a playlist entry in the M3U sidepanel can see TMDB metadata, pipeline state, and raw ingestion details, but have no way to know how large the remote media file behind `line_url` is before downloading it. Surfacing the remote file size helps users judge stream quality/legitimacy and anticipate download time and disk usage, without waiting for an actual download attempt.

## What Changes

- Add nullable `remote_file_size` (bytes) and `remote_file_size_checked_at` columns to `ProcessedLine`, so a size lookup is attempted at most once per line.
- Add an automatic backfill step to `stalkeer process` that probes `line_url` via `HEAD` (falling back to a ranged `GET` when `HEAD` is unsupported or returns no `Content-Length`) for movie/TV show/uncategorized entries whose size has never been checked, and persists the result (or records the attempt even on failure/unavailability, so it is never retried).
- Bound the backfill to a configurable number of lines per `process` run, so a large pre-existing backlog drains over multiple daily runs instead of extending a single run indefinitely.
- Exclude `channels` (live streams) from size probing entirely — they have no fixed file size.
- Expose `remote_file_size` on `ItemResponse` (API) and `PlaylistItem` (frontend types).
- Display the remote file size in the playlist sidepanel (human-readable, e.g. `4.5 GB`), or an "unavailable" state when checked but not obtained; omit the field entirely for channels.

## Capabilities

### New Capabilities
- `remote-file-size-storage`: New `remote_file_size` / `remote_file_size_checked_at` fields on `ProcessedLine`, and their exposure through the items API.
- `remote-file-size-backfill`: Automatic, bounded, single-attempt backfill of remote file sizes for VOD playlist entries, run as part of `stalkeer process`.

### Modified Capabilities
- `m3u-playlist-details-sidepanel`: The sidepanel gains a requirement to display the remote file size (or its unavailability) for VOD entries.

## Impact

- **Backend**: `internal/models/processed_line.go` (new columns), `internal/processor/processor.go` and a new backfill file (probing logic, wired into `Process()`), `internal/api/dto.go` / `internal/api/handlers_frontend.go` (API exposure), `cmd/process.go` (stats output, configurable per-run limit).
- **Frontend**: `frontend/src/types.ts` (`PlaylistItem.remote_file_size`), `frontend/src/components/PlaylistTab.tsx` (sidepanel display), locale files (`frontend/src/locales/*/playlist.json`).
- **Database**: auto-migration adds two nullable columns to `processed_lines`; no data loss to existing rows.
- **Network**: introduces outbound `HEAD`/ranged `GET` requests from the backend to IPTV-provided URLs during `process` runs — bounded per run and best-effort (failures are logged, never fail the overall command).
