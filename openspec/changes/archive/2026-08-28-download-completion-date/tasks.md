## 1. Backend: expose completion date

- [x] 1.1 Add `CompletedAt *time.Time` (json: `completed_at,omitempty`) to `DownloadEnrichedResponse` in `internal/api/types.go` and verify the struct compiles
- [x] 1.2 Populate `resp.CompletedAt = dl.CompletedAt` in `enrichDownloadInfo` (`internal/api/handlers_frontend.go`) and verify `GET /api/v1/downloads` returns `completed_at` for a completed download and omits it for a pending/failed one (existing handler test or manual curl)

## 2. Frontend: render completion date

- [x] 2.1 Add `completed_at?: string` to the `DownloadEnriched` interface in `frontend/src/types.ts`
- [x] 2.2 In `DownloadsTab.tsx`, destructure `i18n` from `useTranslation('downloads')` (alongside `t`), matching the pattern already used in `PlaylistTab.tsx`
- [x] 2.3 In the technical specs row, conditionally render the formatted completion date (via `formatDate(item.completed_at, i18n.language)` from `frontend/src/utils/date.ts`) after Duration, following the existing `•`-separated bullet pattern, only when `item.completed_at` is present
- [x] 2.4 Add the new label key (e.g. `completedAt`) to `frontend/src/locales/fr/downloads.json` and `frontend/src/locales/en/downloads.json`

## 3. Verification

- [x] 3.1 Run the frontend against a mix of completed/pending/failed downloads and verify: completed cards show the date inline with the other specs, non-completed cards show no date and no stray separator
- [x] 3.2 Resize the viewport below the mobile breakpoint (768px) and verify the completion date still renders in full on the mobile card layout, per `frontend-responsive-layout`'s "preserve all desktop information" requirement
- [x] 3.3 Run existing backend tests (`internal/api/handlers_frontend_test.go`) and frontend build/typecheck to confirm no regressions
