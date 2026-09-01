## Why

On the Playlist page, "processing" (a video was enriched via TMDB matching) and "downloading" (a video's file was fetched) are two independent concerns, but today they are collapsed into a single `ProcessedLine.State` badge that overwrites itself as the item moves through the pipeline. Once a download starts, the fact that the item was already `processed` is no longer visible — a failed forced download shows only a red "failed" badge, with no indication that enrichment itself succeeded. Users cannot tell "not yet downloaded" apart from "download failed" apart from "still processing" at a glance, and the Playlist and its sidepanel need to surface both facts independently and simultaneously.

## What Changes

- Derive two independent display values from the existing `ProcessedLine.State` field: a **processing status** (`pending` / `processed`) and a **download status** (`not_downloaded` / `downloading` / `organizing` / `downloaded` / `failed`), computed purely on the frontend (any state other than `pending` implies processing succeeded, since matching happens before any download is attempted). No backend or API change: `state` keeps its current values and semantics for every existing backend consumer (matcher, stats, force-download eligibility, maintenance).
- Playlist table (desktop) and mobile list card: render two independent badges/icons instead of one combined badge.
  - Desktop: two compact text badges side by side (processing, download).
  - Mobile: two small icon indicators (colored per status) with an accessible label (`title`/`aria-label`), replacing the single text badge — keeps the list card compact.
- Playlist item sidepanel (drawer): replace the single "État actuel" badge with two full-text badges shown side by side — "Traitement" and "Téléchargement" — including an explicit gray `[ non téléchargé ]` badge when no download has been attempted yet (rather than hiding that field).
- Add a distinct visual style for "not downloaded" (neutral/gray) so it is never confused with "downloaded" (green) or "failed" (red).

## Capabilities

### New Capabilities
- `playlist-pipeline-state-display`: independent processing-status and download-status badges/icons on the Playlist table and mobile list card.

### Modified Capabilities
- `m3u-playlist-details-sidepanel`: the sidepanel's pipeline status display changes from one combined badge to two independent full-text badges (processing, download), including an explicit "not downloaded" state.

## Impact

Frontend only:
- `frontend/src/utils/pipelineState.ts` (new derivation helpers for processing/download status)
- `frontend/src/components/PlaylistItemsTable.tsx` (desktop two-badge column, mobile two-icon indicator)
- `frontend/src/components/PlaylistTab.tsx` (sidepanel "État du Pipeline d'Ingestion" section)
- `frontend/src/index.css` (new/adjusted badge classes, notably a neutral "not downloaded" style)
- `frontend/src/locales/{fr,en}/playlist.json` (new labels: download status values, "not downloaded")

No backend/API changes, no data migration.
