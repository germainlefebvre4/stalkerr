# processing-run-items Specification

## Purpose

Links each processed playlist item to the import run (`processing_logs` entry) that last created or touched it, and lets users inspect exactly which items a given run added or processed by clicking that run's row in the processing-logs table.

## Requirements

### Requirement: Item Attribution to Processing Run
Each `processed_lines` row SHALL record which `processing_logs` entry last created or updated it. When an import run creates a new item or updates an existing item (e.g. on a forced re-process), the system SHALL set that item's run attribution to the current run's id, replacing any prior attribution.

#### Scenario: A newly created item is attributed to the run that created it
- **WHEN** an import run processes an M3U line that has never been seen before and creates a new `processed_lines` row
- **THEN** that row's run attribution SHALL be set to the id of the currently executing `processing_logs` entry

#### Scenario: A re-processed item is re-attributed to the newer run
- **WHEN** an import run is executed with the force option and updates a `processed_lines` row that was originally created by an earlier run
- **THEN** that row's run attribution SHALL be updated to the id of the currently executing run, no longer reflecting the earlier run

#### Scenario: A skipped duplicate keeps its original attribution
- **WHEN** an import run encounters a line whose hash already exists and is skipped as a duplicate (not forced)
- **THEN** that item's run attribution SHALL remain unchanged from whichever run last created or updated it

### Requirement: Item Listing Supports Filtering by Processing Run
`GET /api/v1/items` SHALL accept an optional `processing_log_id` (integer) query parameter, restricting results to items whose run attribution matches the given `processing_logs` id. This filter SHALL combine with the endpoint's existing filters (`content_type`, `state`, `group_title`, `tvg_name`, `tmdb_enriched`, `movie_id`, `tmdb_id`) and existing pagination/sorting parameters.

#### Scenario: Filtering items by processing_log_id
- **WHEN** a client requests `GET /api/v1/items?processing_log_id=42`
- **THEN** the system SHALL return only items whose run attribution is `42`, using the same pagination and sort semantics as an unfiltered request

#### Scenario: The count of a run-scoped listing matches the run's reported item count
- **WHEN** a client requests `GET /api/v1/items?processing_log_id=42` for a completed run whose `processing_logs` entry reports `item_count=17`
- **THEN** the response's total count SHALL be `17`

#### Scenario: An unknown or not-yet-attributed processing_log_id returns an empty list
- **WHEN** a client requests `GET /api/v1/items?processing_log_id=<id>` for a `processing_logs` entry that predates run attribution being recorded, or for an id that does not exist
- **THEN** the system SHALL return a `200 OK` response with an empty item list, not an error

### Requirement: Inspecting a Processing Run's Items From the Logs Table
The processing-logs table (Logs/Processing page) SHALL make each row clickable. Clicking a row SHALL open a dialog listing the items attributed to that run, rendered with the same columns, state badges, and per-item actions (association/correction, pipeline reset) as the Playlist Items view, paginated independently of the logs table's own pagination.

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
