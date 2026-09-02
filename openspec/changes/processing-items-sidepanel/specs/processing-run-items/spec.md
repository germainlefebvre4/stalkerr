## MODIFIED Requirements

### Requirement: Inspecting a Processing Run's Items From the Logs Table
The processing-logs table (Logs/Processing page) SHALL make each row clickable. Clicking a row SHALL open a dialog listing the items attributed to that run, rendered with the same columns, state badges, and per-item actions (association/correction, pipeline reset) as the Playlist Items view, paginated independently of the logs table's own pagination. Within that dialog, clicking one of the listed items SHALL open the same item-detail sidepanel available from the Playlist Items view, scoped to that single item, nested within the run's items dialog without introducing a competing modal overlay or focus trap.

#### Scenario: Clicking a completed run's row lists its items
- **WHEN** a user clicks a row of a completed processing run reporting `item_count=17`
- **THEN** the system SHALL open a dialog showing that run's 17 items, paginated if they exceed one page

#### Scenario: Clicking an in-progress run's row lists items processed so far
- **WHEN** a user clicks a row of a processing run whose status is `in_progress`
- **THEN** the system SHALL open a dialog showing the items attributed to that run that have been saved so far, without requiring the run to complete first

#### Scenario: Clicking a pre-migration run's row shows an empty state
- **WHEN** a user clicks a row of a processing run that predates run attribution being recorded (no items carry its id)
- **THEN** the system SHALL open the dialog showing the same empty state used elsewhere for a run with no items, rather than an error

#### Scenario: Correcting an item from within the run's item dialog
- **WHEN** a user uses the association/correction action on an item shown inside a run's item dialog
- **THEN** the system SHALL apply the correction the same way it does from the Playlist Items view

#### Scenario: Clicking an item within the run's item dialog opens its detail sidepanel
- **WHEN** a user clicks an item row inside the run's items dialog
- **THEN** the system SHALL open the item-detail sidepanel for that item, showing the same TMDB metadata, pipeline-state badges, and raw ingestion details as when opened from the Playlist Items view

#### Scenario: Closing the item detail sidepanel keeps the run's item dialog open
- **WHEN** a user closes the item-detail sidepanel opened from within the run's items dialog
- **THEN** the run's items dialog SHALL remain open, still showing the run's item list
