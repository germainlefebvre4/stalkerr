## 1. Locales

- [x] 1.1 In `frontend/src/locales/en/downloads.json` and `frontend/src/locales/fr/downloads.json`, add a text-only label per status next to the existing emoji-only `status.*` keys (keep the emoji keys unchanged): `completedText`, `downloadingText`, `pendingText`, `failedText` (e.g. EN: "Completed"/"Downloading"/"Pending"/"Failed", matching the wording already used in `errors.json`'s combined `status.*` keys; FR: "Complété"/"En cours"/"En attente"/"Échec", matching `errors.json`). Verify by running the frontend test suite (`npm test` in `frontend/`) with no missing-translation warnings.
- [x] 1.2 Verify `status.retrying` and `status.cancelled` are left untouched in both locale files (out of scope — unchanged on every viewport).

## 2. Downloads list (`DownloadsSummaryList.tsx`)

- [x] 2.1 Update `getStatusInfo` (or the row-building logic that calls it) to accept the current `isMobile` flag and return a composed label: emoji alone on mobile (current behavior, unchanged), emoji + space + text label on desktop, using the new `*Text` keys from task 1.1. Keep the `failedWithRetry` case's mobile form (`❌ (3×)`) and add its desktop form (`❌ Échec (3×)`, i.e. emoji + text + count).
- [x] 2.2 Verify: with a viewport ≥ 768px, each status badge in the desktop table shows emoji + text (e.g. `✅ Complété`); with a viewport < 768px, the mobile card badge shows the emoji alone, unchanged from before. Add/update assertions in `DownloadsSummaryList.test.tsx` covering both viewport cases for at least `completed`, `failed` (no retry), and `failed` (with retry).
- [x] 2.3 Verify the badge still renders on a single line without wrapping when the desktop text is added, for a download with a long title (existing `white-space: nowrap` / `flex-shrink: 0` rules on `.badge` in `index.css` should already cover this — confirm visually via the dev server, no CSS change expected).

## 3. Downloads details sidepanel (`DownloadsTab.tsx`)

- [x] 3.1 Import and call `useIsMobile` (from `frontend/src/hooks/useMediaQuery`), matching the pattern already used in `DownloadsSummaryList.tsx`, `HomeTab.tsx`, and `RadarrSonarrTab.tsx`.
- [x] 3.2 Apply the same emoji-only (mobile) / emoji+text (desktop) composition used in task 2.1 to the drawer's status-section badge (`drawerData.statusLabel`), including the `failedWithRetry` desktop form.
- [x] 3.3 Verify: opening the sidepanel on desktop shows the status badge with emoji + text; opening it on mobile shows the emoji alone. Add/update assertions in `DownloadsTab.test.tsx` covering both viewport cases for at least `completed` and `failed` (with and without retries).

## 4. Manual verification

- [x] 4.1 Run the app (`/run` or the project's existing dev workflow) and check the Downloads tab and its sidepanel at both a desktop width (≥768px) and a mobile width (<768px) for each status: `completed`, `downloading`, `pending`, `failed` (no retry), `failed` (with retries), `retrying`, `cancelled` — confirm desktop shows emoji+text (except `cancelled`, text-only as before, and `retrying`, emoji+word as before) and mobile shows emoji only (except `cancelled`, unchanged).
