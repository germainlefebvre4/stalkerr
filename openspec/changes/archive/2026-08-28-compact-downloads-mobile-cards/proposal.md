## Why

On mobile, each Downloads card currently renders every piece of desktop information at full length (raw URL, a separate folder/file monospace box, inline technical specs, three validation badges, genres), stacking into 7+ blocks per card. This makes it hard to scan multiple downloads at once on a phone screen. The user wants the mobile Downloads view pared down to the essentials, with secondary technical detail available on demand rather than always on screen.

## What Changes

- On mobile viewports only, the Downloads card SHALL show, always visible: content title + year, status badge, the "Déplacer" button (when applicable), the full `download_path`, and a size/progress line (file size when completed/failed, `downloaded / total (percent)` + progress bar when downloading/retrying — no duplicate size shown alongside progress), and the error message when failed.
- The mobile card SHALL replace today's two separate elements (the raw `item.url` secondary line, and the boxed `folder_name`/`file_name` monospace display) with a single line showing the full `download_path`.
- Format/resolution/duration, the three validation badges (year, format, low quality), and genres SHALL be hidden by default on mobile and revealed by tapping the card, which expands an accordion section in place. Each card's expanded/collapsed state SHALL be independent of other cards.
- Desktop layout (viewport ≥ 768px) is unchanged: all information remains always visible, no accordion.
- This applies uniformly to all download statuses (completed, downloading, retrying, failed, pending).
- **BREAKING**: none (mobile-only display change; no API or data contract changes). This modifies the existing `frontend-responsive-layout` requirement that currently mandates all desktop information stay always-visible on mobile Downloads cards — that requirement is superseded by the progressive-disclosure behavior described above.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `frontend-responsive-layout`: The "Responsive Card Density" requirement's mobile Downloads card behavior changes from "all desktop information always visible" to "essential fields always visible, secondary technical detail collapsed by default and revealed per-card via tap," and the mobile card's path display changes from raw URL + separate folder/file box to a single `download_path` line.

## Impact

- `frontend/src/components/DownloadsTab.tsx`: card markup restructured for mobile — add per-card expand/collapse state, reorder always-visible vs. collapsible fields, swap URL/folder-file display for `download_path`.
- `frontend/src/index.css`: new mobile-only styles for the collapsed/expanded accordion sections on `.download-card` (chevron affordance, transition, spacing).
- No backend/API changes — `download_path` is already present on `DownloadEnriched`.
