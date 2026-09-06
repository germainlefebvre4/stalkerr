## MODIFIED Requirements

### Requirement: Expanding a Group Lists Its Underlying Items
Clicking a group row (movie, TV show, or unmatched pseudo-group) in the "Films & Séries" view SHALL expand an inline section listing only the underlying items attributed to that group's most recent processing run (the run reported by the group's run-attribution field, capability `processing-run-items`), rendered with the same columns, state badges, and per-item actions (association/correction, pipeline reset) as the "Items" view's table, independently paginated from the group's own pagination. When the group's run-attribution field has no value (its most recent contributing item predates run attribution being recorded), the expanded section SHALL instead list all of that group's underlying items, unscoped by run.

#### Scenario: Expanding a TV show with many episodes
- **WHEN** the user clicks a TV-show group row whose 100 episodes span 5 seasons
- **THEN** the system SHALL expand an inline item list scoped to that show's episodes from its most recent processing run, showing a first page of items with its own pagination controls rather than all 100 at once.

#### Scenario: Expanding a movie group
- **WHEN** the user clicks a movie group row
- **THEN** the system SHALL expand an inline item list scoped to items linked to that movie and attributed to the group's most recent processing run, with the same columns and actions as the Items view.

#### Scenario: Expanding an unmatched pseudo-group
- **WHEN** the user clicks the "unmatched TV shows" pseudo-group row
- **THEN** the system SHALL expand an inline item list scoped to TV-show items with no linked TV show record, attributed to the pseudo-group's most recent processing run.

#### Scenario: Older items from earlier runs are hidden
- **WHEN** a movie group's most recent processing run added one new item last night, and that movie also has items linked from processing runs on previous nights
- **THEN** the expanded item list SHALL show only last night's item, not the items from previous runs.

#### Scenario: A group predating run attribution lists all its items
- **WHEN** a group's most recent contributing item predates run attribution being recorded, so the group has no run-attribution value
- **THEN** the expanded item list SHALL show all of that group's underlying items, matching the behavior before run-scoping was introduced, rather than an empty list.

## ADDED Requirements

### Requirement: Grouped Row Reports Its Most Recent Processing Run
Each movie, TV-show, or unmatched pseudo-group entry returned by `GET /api/v1/items/grouped` SHALL include the `processing_logs` id attributed to the same contributing item whose timestamp is reported as the group's `latest_activity`, or no value when that item predates run attribution being recorded.

#### Scenario: A group's reported run matches its most recent item's run
- **WHEN** a movie group's most recent contributing item was created by processing run `42`
- **THEN** the group entry SHALL report `42` as its run-attribution value.

#### Scenario: A group predating run attribution reports no run
- **WHEN** a group's most recent contributing item predates run attribution being recorded
- **THEN** the group entry SHALL report no run-attribution value, rather than a fabricated one.
