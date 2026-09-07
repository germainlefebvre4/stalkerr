## MODIFIED Requirements

### Requirement: Expanding a Group Lists Its Underlying Items
Clicking a group row (movie, TV show, or unmatched pseudo-group) in the "Films & Séries" view SHALL expand an inline section listing that group's underlying items, rendered with the same columns, state badges, and per-item actions (association/correction, pipeline reset) as the "Items" view's table, independently paginated from the group's own pagination. Clicking one of those underlying item rows (anywhere other than its action buttons) SHALL open the same details sidepanel used by the "Items" view for that item, without collapsing the expanded group.

#### Scenario: Expanding a TV show with many episodes
- **WHEN** the user clicks a TV-show group row whose 100 episodes span 5 seasons
- **THEN** the system SHALL expand an inline item list scoped to that show's episodes, showing a first page of items with its own pagination controls rather than all 100 at once.

#### Scenario: Expanding a movie group
- **WHEN** the user clicks a movie group row
- **THEN** the system SHALL expand an inline item list scoped to items linked to that movie, with the same columns and actions as the Items view.

#### Scenario: Expanding an unmatched pseudo-group
- **WHEN** the user clicks the "unmatched TV shows" pseudo-group row
- **THEN** the system SHALL expand an inline item list scoped to TV-show items with no linked TV show record.

#### Scenario: Clicking an underlying item opens its details sidepanel
- **WHEN** the user clicks an item row inside an expanded group's item list
- **THEN** the system SHALL open the details sidepanel for that item, showing the same TMDB metadata, pipeline state badges, and ingestion provenance as when opened from the "Items" view, and the expanded group SHALL remain expanded.

#### Scenario: Row-level action buttons do not open the sidepanel
- **WHEN** the user clicks the "Associate"/"Correct" or "Reset" action button on an item row inside an expanded group's item list
- **THEN** the system SHALL perform that action without opening the details sidepanel.
