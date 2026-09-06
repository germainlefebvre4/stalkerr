## 1. Backend: report each group's run attribution

- [x] 1.1 Add `MAX(pl.processing_log_id) AS latest_processing_log_id` to each of the four UNION arms in `internal/api/handlers_grouped.go` (movies, tvshows, unmatched movies, unmatched tvshows) and verify `go build ./...` succeeds.
- [x] 1.2 Add `LatestProcessingLogID *uint` to `itemGroupRow` and `latest_processing_log_id` (nullable) to `ItemGroupResponse` (`internal/api/dto.go`), wiring it through `toResponse()`, and verify with a unit/integration test that `GET /api/v1/items/grouped` returns the expected run id for a group whose contributing items span two different `processing_log_id`s (asserting it reports the newer one, matching `latest_activity`).
- [x] 1.3 Add a test covering a group whose only contributing item has a null `processing_log_id`, verifying the response's `latest_processing_log_id` is absent/null rather than a fabricated value.

## 2. Frontend: scope the expand request to that run

- [x] 2.1 Add `latest_processing_log_id?: number` to `MediaGroupItem` (`frontend/src/types.ts`).
- [x] 2.2 Update `api.getGroupItems` (`frontend/src/services/api.ts`) to accept the group's `latest_processing_log_id` and, when present, append `&processing_log_id=<id>` to the `/api/v1/items` request; when absent, send the request unchanged (no run filter), and verify with a unit test covering both branches.
- [x] 2.3 Update `PlaylistGroupedView.tsx`'s call site to pass the expanding group (already available as `expandedGroup`) through to `getGroupItems` unchanged in shape, and verify the existing expand/pagination behavior still compiles and passes existing component tests.

## 3. Verification

- [x] 3.1 Add/extend an integration test that seeds one movie with items from two different processing runs and asserts expanding that movie's group (via the full `grouped` → `items` request pair) returns only the items from the more recent run.
- [x] 3.2 Manually verify in the running app: a movie/show processed again tonight shows only tonight's item(s) on expand, while a movie/show untouched tonight but still visible under an older date-group header still expands to its own last-run items (not a mix of all history).
