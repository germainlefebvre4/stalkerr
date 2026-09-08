## Why

On the Downloads tab, the status badge (list row/card and the details sidepanel) currently shows only an emoji, with no text, on every viewport — including desktop, where there is ample room and every other tab (Home, Radarr/Sonarr, Playlist) shows a text label next to its status badge. This makes the Downloads status harder to scan on desktop than the rest of the app, since the meaning of each emoji must be memorized rather than read.

## What Changes

- On desktop (viewport ≥ the existing `768px` mobile breakpoint), the Downloads status badge — in the list (table row) and in the details sidepanel's status section — SHALL show its emoji **and** a text label (e.g. `✅ Complété`), instead of the emoji alone.
- On mobile (viewport < `768px`), the Downloads status badge — in the mobile list card and in the details sidepanel — keeps showing the emoji alone, unchanged from current behavior.
- The `failed` status with a non-zero retry count keeps its existing "emoji + retry count" form (e.g. `❌ (3×)`) on mobile; on desktop it gains the "Échec"/"Failed" text alongside the emoji and count (e.g. `❌ Échec (3×)`).
- The `retrying` status (already emoji + word on every viewport today) is unchanged.
- No emoji glyphs change. The `✅`/`❌` pair itself is not being replaced — this change only adds a text label alongside the emoji on desktop.
- Frontend-only: no API or data model changes.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `downloads-display-ui`: the status badge in the Downloads list (table row) changes from emoji-only on every viewport to emoji-only on mobile / emoji+text on desktop.
- `downloads-details-sidepanel`: the status badge in the sidepanel's status section follows the same desktop (emoji+text) / mobile (emoji-only) rule as the list.

## Impact

- `frontend/src/components/DownloadsSummaryList.tsx`: status label building (`getStatusInfo`) and rendering (desktop table row, mobile card) need a responsive text/no-text split, using the existing `useIsMobile` hook already imported there.
- `frontend/src/components/DownloadsTab.tsx`: the sidepanel's status badge (currently a single unconditional render) needs the same responsive split; needs to import `useIsMobile`.
- `frontend/src/locales/en/downloads.json` and `frontend/src/locales/fr/downloads.json`: the `status.*` keys need a text label available for each status, separate from the emoji, so it can be shown only on desktop.
