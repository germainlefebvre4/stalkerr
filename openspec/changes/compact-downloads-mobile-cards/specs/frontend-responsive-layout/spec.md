## MODIFIED Requirements

### Requirement: Responsive Card Density

Below the mobile breakpoint, the KPI cards, filter cards, and download cards SHALL reduce their padding and font sizes and ensure all interactive elements (buttons, badges) meet a minimum touch target size of `44px` in height. KPI cards and filter cards SHALL preserve all information currently shown on desktop. Download cards SHALL show their essential fields always visible (per the Mobile Download Card Progressive Disclosure requirement), with secondary technical detail collapsed by default and reachable per-card via tap.

#### Scenario: Download card remains fully readable and tappable on a narrow viewport
- **WHEN** the viewport is narrower than the mobile breakpoint and the Downloads view is active
- **THEN** the frontend SHALL render each download card with its essential fields (title, year, status badge, `download_path`, size/progress line, error message when failed, and the "Déplacer" button when applicable) visible without horizontal overflow, any action button (e.g. "Déplacer") SHALL have a tappable height of at least `44px`, and the card's expand/collapse control SHALL also meet the `44px` minimum tappable height.

## ADDED Requirements

### Requirement: Mobile Download Card Progressive Disclosure

Below the mobile breakpoint, each download card SHALL always show its essential fields — content title and year, status badge, the full `download_path`, a size/progress line, the error message when the download has failed, and the "Déplacer" button when the download is completed — and SHALL collapse all other desktop fields (format, resolution, duration, the year/format/low-quality validation badges, genres) into a per-card expandable section that is hidden by default. This applies uniformly across all download statuses (pending, downloading, retrying, completed, failed). Desktop layout (viewport at or above the mobile breakpoint) is unaffected and continues to show all fields always visible, unchanged.

#### Scenario: Essential fields are always visible on a collapsed mobile download card
- **WHEN** the viewport is narrower than the mobile breakpoint and a download card is in its default (collapsed) state
- **THEN** the frontend SHALL display the content title and year, the status badge, the full `download_path`, the size/progress line, the error message (if the download has failed), and the "Déplacer" button (if the download is completed), without requiring the user to tap the card.

#### Scenario: download_path replaces the raw URL and folder/file box on mobile
- **WHEN** the viewport is narrower than the mobile breakpoint
- **THEN** the frontend SHALL render the download's full `download_path` as a single line in place of the desktop's separate raw URL line and boxed folder-name/file-name display.

#### Scenario: Downloading or retrying items show progress instead of a duplicate size
- **WHEN** the viewport is narrower than the mobile breakpoint and a download's status is "downloading" or "retrying"
- **THEN** the frontend SHALL show only the `downloaded / total (percent)` progress line and progress bar as the card's size/progress line, and SHALL NOT additionally render a separate fixed file-size value alongside it.

#### Scenario: Completed or failed items show a fixed size
- **WHEN** the viewport is narrower than the mobile breakpoint and a download's status is "completed" or "failed"
- **THEN** the frontend SHALL show the download's file size as the card's size/progress line.

#### Scenario: Tapping a card reveals its secondary details
- **WHEN** the viewport is narrower than the mobile breakpoint and the user taps a collapsed download card (or its expand control)
- **THEN** the frontend SHALL expand that card in place to reveal its format, resolution, duration, validation badges (year, format, low quality), and genres, without navigating away from the Downloads list.

#### Scenario: Each card's expanded state is independent
- **WHEN** the viewport is narrower than the mobile breakpoint and the user expands one download card among several displayed
- **THEN** the frontend SHALL leave all other download cards in their current (collapsed or expanded) state, unaffected by that action.

#### Scenario: Desktop download cards are unaffected
- **WHEN** the viewport is at or above the mobile breakpoint
- **THEN** the frontend SHALL continue to render each download card with all fields always visible and no expand/collapse control, unchanged from current desktop behavior.
