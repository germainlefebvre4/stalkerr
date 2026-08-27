## ADDED Requirements

### Requirement: Playlist Pagination Bar Is Compact on Mobile

Below the mobile breakpoint, the Playlist pagination bar SHALL render the first/previous/next/last controls (`<<`, `<`, `>`, `>>`) together with a `Page X / Y` text indicator in place of the numbered page-button list, and SHALL NOT render the "go to page" input/button group. The bar SHALL fit within the viewport width without requiring horizontal scrolling, regardless of the total number of pages.

#### Scenario: Compact pagination bar renders on mobile
- **WHEN** the viewport is narrower than the mobile breakpoint, the Playlist view is active, and there is more than one page of results
- **THEN** the frontend SHALL render the first/previous/next/last controls and a `Page X / Y` text indicator, SHALL NOT render individual numbered page buttons or the "go to page" input/button group, and the pagination bar SHALL require no horizontal scrolling to view in full.

#### Scenario: Desktop pagination bar is unchanged
- **WHEN** the viewport is at or above the mobile breakpoint
- **THEN** the frontend SHALL continue to render the full pagination bar as defined by `frontend-ihm-dashboard` (numbered page buttons with ellipsis, first/last jump buttons, and the "go to page" input), unchanged from current desktop behavior.

#### Scenario: Navigation controls remain functional in the compact mobile bar
- **WHEN** the viewport is narrower than the mobile breakpoint and the user taps the next-page (`>`) control while not on the last page
- **THEN** the frontend SHALL navigate to the next page, fetch/display its items, and update the `Page X / Y` indicator to reflect the new current page.
