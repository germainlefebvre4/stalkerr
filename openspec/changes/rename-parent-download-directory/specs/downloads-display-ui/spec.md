## ADDED Requirements

### Requirement: Rename Download Item
On each completed download card, the system SHALL provide a "Renommer" action, available independently of the existing "Move" action, that opens a dialog pre-filled with the item's current parent folder name (from `file_info.folder_name`) as an editable free-text field, plus an optional field for a different destination root. Submitting the dialog SHALL call `POST /api/v1/downloads/:id/rename` scoped to that single download's id, and SHALL only update the affected card's displayed folder/path — every other visible download card SHALL remain unchanged.

#### Scenario: Open the rename dialog pre-filled with the current name
- **WHEN** the user clicks "Renommer" on a completed download card
- **THEN** the system SHALL open a dialog with the folder name field pre-filled with the current parent folder name, and the destination root field empty.

#### Scenario: Rename in place
- **WHEN** the user edits the folder name field only and submits
- **THEN** the system SHALL call `POST /api/v1/downloads/:id/rename` with `{"new_name": <edited value>}`, and on success SHALL update only that card's displayed folder path.

#### Scenario: Rename and move in one action
- **WHEN** the user edits both the folder name field and the destination root field and submits
- **THEN** the system SHALL call `POST /api/v1/downloads/:id/rename` with both `new_name` and `destination_parent_dir`, and on success SHALL update only that card's displayed folder path.

#### Scenario: Rename blocked by a naming collision
- **WHEN** the rename request fails with the `"rename_target_exists"` error code
- **THEN** the system SHALL keep the dialog open and display a translatable error message indicating the destination already exists, without altering any download card.

#### Scenario: Siblings unaffected by a rename
- **WHEN** a rename succeeds for one download card
- **THEN** the system SHALL leave every other download card's displayed folder path, status, and data unchanged, without requiring a full page reload.
