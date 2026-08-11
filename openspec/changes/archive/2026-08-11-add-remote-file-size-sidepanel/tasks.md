## 1. Data model & migration

- [x] 1.1 Add `RemoteFileSize *int64` and `RemoteFileSizeCheckedAt *time.Time` (both nullable, `json:"remote_file_size,omitempty"` / `json:"remote_file_size_checked_at,omitempty"`) to `ProcessedLine` in `internal/models/processed_line.go`
- [x] 1.2 Verify GORM auto-migration adds the two nullable columns to `processed_lines` without touching existing rows (manual check against a dev DB or existing migration test pattern)

## 2. Remote file size probing

- [x] 2.1 Add a config value for the per-request probe timeout and the per-run probe cap (mirroring the `m3u.download.max_file_size_mb` pattern in `internal/config/config.go`: struct field, `viper.BindEnv`, `viper.SetDefault`), with a sane default for the cap (e.g. 200) and timeout (e.g. a few seconds)
- [x] 2.2 Create `internal/processor/remote_file_size.go` implementing a `BackfillRemoteFileSize(db *gorm.DB, log *logger.Logger, cfg ...) (*RemoteFileSizeBackfillStats, error)` function, following the shape of `internal/processor/backfill_metadata.go` (batched query, per-row error isolation, stats struct)
- [x] 2.3 Implement the probe: HTTP `HEAD` on `line_url` reading `Content-Length`, with bounded timeout
- [x] 2.4 Implement the fallback probe: `GET` with `Range: bytes=0-0`, reading the total size from `Content-Range`, used when `HEAD` fails or returns no usable `Content-Length`
- [x] 2.5 Query eligible rows: `content_type IN (movies, tvshows, uncategorized)`, `line_url IS NOT NULL`, `remote_file_size_checked_at IS NULL`, batched/limited to the configured per-run cap
- [x] 2.6 On probe completion (success or failure), always set `remote_file_size_checked_at`; set `remote_file_size` only on success — ensuring a line is never re-probed on a later run
- [x] 2.7 Log and continue (never abort the loop or fail the caller) on any single-row probe error or timeout
- [x] 2.8 Add unit tests for `internal/processor/remote_file_size.go` covering: HEAD success, HEAD failure + GET-range fallback success, both failing, channels excluded from the query, per-run cap respected, already-checked rows skipped

## 3. Wire into `process`

- [x] 3.1 Call `BackfillRemoteFileSize` from `processor.Process()` in `internal/processor/processor.go`, alongside the existing `BackfillRichMetadata` call
- [x] 3.2 Add `FileSizeBackfilled` / `FileSizeBackfillErrors` (or equivalent) fields to the processor's stats struct
- [x] 3.3 Print the new stats in `cmd/process.go`, following the existing `MetadataBackfilled` / `MetadataBackfillErrors` output pattern

## 4. API exposure

- [x] 4.1 Add `RemoteFileSize *int64` (`json:"remote_file_size,omitempty"`) to `ItemResponse` in `internal/api/dto.go`
- [x] 4.2 Populate it from `ProcessedLine.RemoteFileSize` wherever `ItemResponse` is built in `internal/api/handlers_frontend.go` (and `internal/api/handlers.go` if items are also mapped there)
- [x] 4.3 Update/extend `internal/api/handlers_frontend_test.go` to cover `remote_file_size` presence/absence in the response

## 5. Frontend

- [x] 5.1 Add `remote_file_size?: number` to `PlaylistItem` in `frontend/src/types.ts`
- [x] 5.2 In `frontend/src/components/PlaylistTab.tsx`, add remote file size display to the sidepanel (Section 3, M3U provenance area, or alongside Section 5's raw URL block): formatted human-readable size (MB/GB) when `remote_file_size` is present; an "unavailable" state when absent for a VOD item; nothing at all when `content_type === 'channels'`
- [x] 5.3 Add the new i18n keys (`drawer.remoteFileSize`, `drawer.remoteFileSizeUnavailable`, or similar, following the existing `drawer.*` icon-prefixed naming convention) to `frontend/src/locales/en/playlist.json` and `frontend/src/locales/fr/playlist.json`

## 6. Verification

- [x] 6.1 Run backend test suite (`go test ./...`) and confirm new/existing tests pass
- [x] 6.2 Run frontend build/lint/tests for the modified files
- [x] 6.3 Manually run `stalkeer process` against a dev playlist with real VOD entries and confirm `remote_file_size` / `remote_file_size_checked_at` populate as expected, and that the sidepanel renders the three states (known size, unavailable, omitted for channels)
