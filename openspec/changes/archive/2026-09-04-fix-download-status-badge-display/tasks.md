## 1. Update status label translations

- [x] 1.1 In `frontend/src/locales/fr/downloads.json`, update the `status` block: `completed` → `"✅"`, `downloading` → `"📥"`, `pending` → `"⏳"`, `failed` → `"❌"`, `failedWithRetry` → `"❌ ({{count}}×)"`; leave `retrying` unchanged. Verify by inspecting the file for exact key values.
- [x] 1.2 Apply the same key changes to `frontend/src/locales/en/downloads.json` (translated wording only where a word remains, e.g. none here since all four are emoji-only and the retry one has no word). Verify by inspecting the file for exact key values.

## 2. Verify status badge logic still composes correctly

- [x] 2.1 In `frontend/src/components/DownloadsSummaryList.tsx`, confirm `getStatusInfo` (lines 18-35) still produces the right label per status now that translations are emoji-only — `completed`, `downloading`, and `pending` return the translated string as-is; `failed` still branches on `item.retry_count > 0` between `status.failedWithRetry` (with `{{count}}`) and `status.failed`. No code change expected here since the branching already exists; verify by reading the function against the new JSON values.
- [x] 2.2 Confirm the same `getStatusInfo`/`statusLabel` is used for both the mobile card render (line ~72) and the desktop table render further down in the same file, so the new badge content applies to both views without separate changes. Verify by reading both render branches.

## 3. Fix `.badge` CSS to prevent wrapping

- [x] 3.1 In `frontend/src/index.css`, add `white-space: nowrap;` and `flex-shrink: 0;` to the `.badge` rule (around line 322-334). Verify by inspecting the updated rule.

## 4. Manual verification

- [x] 4.1 Run the frontend dev server, open the Downloads tab on a mobile viewport (or narrow browser window), and find/create a download with a long title. Verify the status badge stays on one line and never shows its emoji and text (or count) stacked vertically, for `completed`, `pending`, `downloading`, and `failed` (with and without retries) statuses.
- [x] 4.2 Verify visually that `pending` (⏳) and `downloading` (📥) now show different emoji and are no longer visually identical.
- [x] 4.3 Verify the desktop table view (viewport at or above the mobile breakpoint) also shows the updated emoji-only/retry-count badges, unchanged in layout otherwise.
