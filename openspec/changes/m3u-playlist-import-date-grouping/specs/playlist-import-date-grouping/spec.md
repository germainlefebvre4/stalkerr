## Purpose

Defines how the M3U playlist items table (desktop table and mobile card list) visually groups items by their import (`created_at`) day, so users can quickly identify recent import activity without reading every row's date.

## ADDED Requirements

### Requirement: Date-Group Separators When Sorted By Import Date
When the playlist items view is sorted by `created_at` (ascending or descending), the system SHALL render a lightweight, low-emphasis group header above the first item of each distinct calendar day present in the currently displayed page, using the viewer's local time zone to determine calendar day boundaries. The system SHALL NOT add any per-item marker, badge, dot, or background tint to individual rows or cards as part of this grouping.

#### Scenario: Items from the same day are grouped under one header
- **WHEN** the playlist items table is sorted by `created_at` descending and the current page contains three items imported today followed by two items imported yesterday
- **THEN** the system SHALL render one group header above the first of the three items imported today, and a second group header above the first of the two items imported yesterday, with no separator between the other same-day items.

#### Scenario: Mobile card list groups items the same way
- **WHEN** the playlist items view is displayed on a mobile viewport, sorted by `created_at`, and the current page spans two distinct import days
- **THEN** the system SHALL render an equivalent group header between the mobile cards at the boundary between the two days, without adding a marker to individual cards.

### Requirement: Relative Labels for Recent Days, Full Date Beyond
The system SHALL label a date-group header for the current calendar day as "Today" (localized), for the immediately preceding calendar day as "Yesterday" (localized), and for any earlier calendar day using the same full date format already used in the item's date cell (e.g. `DD/MM/YYYY`).

#### Scenario: Today's group uses a relative label
- **WHEN** a date-group header corresponds to the viewer's current calendar day
- **THEN** the system SHALL display the localized equivalent of "Today" (e.g. "Aujourd'hui" in French) as the header label.

#### Scenario: Yesterday's group uses a relative label
- **WHEN** a date-group header corresponds to the calendar day immediately before the viewer's current day
- **THEN** the system SHALL display the localized equivalent of "Yesterday" (e.g. "Hier" in French) as the header label.

#### Scenario: Older groups use the full date
- **WHEN** a date-group header corresponds to any calendar day earlier than yesterday
- **THEN** the system SHALL display the full date in the same format used elsewhere in the table's date columns, with no relative wording.

### Requirement: Grouping Disabled for Non-Date Sorts
When the playlist items view is sorted by any field other than `created_at`, the system SHALL render the items as a flat list with no date-group headers, since items are no longer contiguous by import day.

#### Scenario: Sorting by media name hides group headers
- **WHEN** the user changes the playlist items sort to `tvg_name` (or any field other than `created_at`)
- **THEN** the system SHALL remove all date-group headers from the table and render items as a plain list.

#### Scenario: Switching back to date sort restores grouping
- **WHEN** the user changes the playlist items sort from a non-date field back to `created_at`
- **THEN** the system SHALL resume rendering date-group headers according to the other requirements in this capability.

### Requirement: Group Header Repeats At Top Of Each Page
When paginating through playlist items sorted by `created_at`, the system SHALL render the applicable date-group header above the first item of every page, including when that page's first item continues a group whose header already appeared on a previous page.

#### Scenario: A page that continues a group still shows its header
- **WHEN** the last item of one page and the first item of the next page both belong to the "Yesterday" group
- **THEN** the system SHALL render the "Yesterday" group header above the first item of the next page, even though the group did not start on that page.
