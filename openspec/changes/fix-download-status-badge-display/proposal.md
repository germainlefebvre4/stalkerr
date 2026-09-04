## Why

On mobile, the Downloads tab's status badge wraps onto two lines when the download's title is long, pushing the badge's text below its emoji (`.badge` has no `white-space: nowrap` / `flex-shrink: 0` guard, unlike the title). Separately, the "En cours" (downloading) and "En attente" (pending) statuses currently share the same ⏳ emoji, so they are only distinguishable by their text label, and the "Échec" (failed) badge always shows the word "Échec" even when the retry count is the only variable information. Cleaning up the badge content (dropping redundant words, giving "downloading" its own emoji) removes the risk of ever needing to wrap in the first place, and the CSS fix guarantees it going forward.

## What Changes

- Give the `downloading` status its own emoji (📥) so it's visually distinct from `pending` (⏳) without relying on text.
- Drop the word label from the `completed`, `pending`, `downloading`, and `failed` (no-retry) status badges, showing the emoji alone.
- Change the `failed`-with-retries badge from `❌ Échec ({{count}}×)` to `❌ ({{count}}×)`, dropping the word "Échec" but keeping the retry count.
- Leave the `retrying` badge (`🔄 Réessai`) unchanged — out of scope.
- Fix `.badge` CSS so its content never wraps onto a second line regardless of available width (`white-space: nowrap`, `flex-shrink: 0`), as a layout guarantee independent of the label content change above.
- Applies to both the mobile list-card and desktop table renderings of the Downloads tab, since both consume the same status label.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `downloads-display-ui`: the status badge's label content changes (emoji-only for completed/pending/downloading/failed, retry count without the "Échec" word), and the badge is now required to never wrap its content onto multiple lines.

## Impact

- `frontend/src/locales/fr/downloads.json` and `frontend/src/locales/en/downloads.json`: status label strings updated.
- `frontend/src/components/DownloadsSummaryList.tsx`: `getStatusInfo` (or equivalent) needs to build the failed+retry label from a separate count value instead of a single pre-composed translated string, since the word "Échec" is dropped but the count is kept.
- `frontend/src/index.css`: `.badge` rule (`white-space: nowrap; flex-shrink: 0;`).
- No backend/API changes.
