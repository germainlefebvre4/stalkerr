## Why

On mobile, the Playlist tab's advanced filter grid (media name, group, TMDB enrichment, pipeline state) always renders above the table/card list, pushing it below the fold and forcing a scroll before any content is visible — undermining the table-visibility gains already made by `mobile-responsive-ui`.

## What Changes

- On mobile only (`useIsMobile()`), the advanced filter grid (media name search, group search, TMDB enrichment, pipeline state) renders inside a collapsible section behind a tappable header, collapsed by default so the table/card list is visible without scrolling past it.
- The content-type pills (Tous/Films/Séries) stay outside the collapsible section and remain always visible on mobile, unchanged.
- The collapsible header shows a count of active advanced filters (i.e. those not at their default/"all" value) so the user can tell filters are applied even while the section is collapsed.
- The collapsed/expanded state is not persisted (no `localStorage`/URL entry); it always starts collapsed when the Playlist tab mounts on mobile, regardless of whether a filter is already active from a restored URL.
- Desktop layout and behavior are unchanged — the advanced filter grid keeps rendering inline, uncollapsible.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `frontend-ihm-dashboard`: on mobile viewports, the Playlist advanced filter grid is rendered inside a collapsed-by-default disclosure with an active-filter count, instead of always being fully visible.

## Impact

- Affected files: `frontend/src/components/PlaylistTab.tsx` (new collapsible state and markup), `frontend/src/locales/{en,fr}/playlist.json` (new labels for the disclosure header), `frontend/src/index.css` (styling for the collapsible header/section, scoped to the existing mobile breakpoint).
- Frontend-only change: no API/backend contract changes.
- No change to desktop rendering, to the content-type pills, or to how filter values themselves are stored/persisted (URL/localStorage behavior for filter values is untouched — only the visibility of the advanced grid's container changes).
