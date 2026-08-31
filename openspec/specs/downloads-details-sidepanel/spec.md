# Downloads Details Sidepanel Specification

## Purpose

Provides the Downloads tab's detail sidepanel: the click-triggered Radix UI drawer that shows a single download's full status, file, technical, and validation information, kept live-synced with the existing polling refresh, and hosting the Move/Rename actions that no longer live on the summary row.

## Requirements

### Requirement: Download Item Details Sidepanel
The Downloads tab SHALL open a sliding sidepanel (drawer), based on Radix UI `Dialog`, when the user clicks or taps any download's summary row (capability `downloads-display-ui`). The sidepanel SHALL display, for the selected download:
- A status section: the status badge and, when the status is `downloading` or `retrying`, a progress bar with percentage and downloaded/total size.
- A file section: the folder and file names (from `file_info`) and the full `download_path`, or the source `url` when no local path is available yet.
- A technical specifications section: format (extension), detected resolution, total file size, duration (when known), and completion date (when `completed_at` is set).
- A validation section: the same badge chips previously shown inline (year OK/missing/mismatch, format OK/unknown, low-quality warning for 480p/360p).
- Genres, when `content.genres` is present.

The sidepanel SHALL omit any of the above sections or fields that have no data, without rendering an empty placeholder or a layout gap.

#### Scenario: Open sidepanel from a table row or mobile card
- **WHEN** the user clicks a download's row (desktop table) or taps its card (mobile)
- **THEN** the frontend SHALL open the sidepanel for that download, showing its status, file, technical, and validation information

#### Scenario: Sidepanel omits sections with no data
- **WHEN** the selected download has no `file_info` and no `content.genres`
- **THEN** the sidepanel SHALL omit the file, technical specifications, validation, and genres sections entirely, without showing empty boxes or placeholders

### Requirement: Live-Synced Sidepanel Selection
While the sidepanel is open, the Downloads tab SHALL keep it synchronized with the existing 5-second polling refresh: the displayed download SHALL be looked up by id from the current downloads list on every refresh, rather than frozen at the moment the user opened the panel. If the selected download's id is no longer present in the current (possibly filtered) list, the sidepanel SHALL close automatically.

#### Scenario: Progress updates live while the sidepanel is open
- **WHEN** the user opens the sidepanel for a `downloading` item and the next 5-second poll reports increased `bytes_downloaded` or a changed `status`
- **THEN** the sidepanel SHALL reflect the updated progress and/or status without the user needing to close and reopen it

#### Scenario: Sidepanel closes when its item leaves the list
- **WHEN** the sidepanel is open for a download and a subsequent poll's result (e.g. after a filter change) no longer includes that download's id
- **THEN** the frontend SHALL close the sidepanel automatically instead of showing stale or empty content

### Requirement: Sidepanel Error Message Gated on Failed Status
The sidepanel SHALL render its error message section only when the selected download's current `status` is `failed`. A non-empty `error_message` alone SHALL NOT be sufficient to render it, since `error_message` may be a value carried over from a prior attempt on the same record.

#### Scenario: Completed download with a leftover error_message shows no error section
- **WHEN** the sidepanel is open for a download whose `status` is `completed` and whose `error_message` is non-empty (leftover from an earlier failed attempt)
- **THEN** the sidepanel SHALL NOT render an error message section

#### Scenario: Failed download shows the error message
- **WHEN** the sidepanel is open for a download whose `status` is `failed` and whose `error_message` is non-empty
- **THEN** the sidepanel SHALL render the error message section with that message

#### Scenario: In-progress download with a leftover error_message shows no error section
- **WHEN** the sidepanel is open for a download whose `status` is `downloading` or `retrying` and whose `error_message` is non-empty (leftover from an earlier failed attempt)
- **THEN** the sidepanel SHALL NOT render an error message section

### Requirement: Move and Rename Actions in the Sidepanel
For a completed download, the sidepanel SHALL offer the "Déplacer" and "Renommer" actions (previously available inline on the card) instead of, or in addition to, wherever they are otherwise exposed. The "Renommer" action SHALL open a dialog pre-filled with the item's `rename_folder_name` (the series root folder name for a TV episode, or the movie's own folder name — never a season subdirectory's name) as an editable free-text field, plus an optional field for a different destination root. Submitting the dialog SHALL call `POST /api/v1/downloads/:id/rename` scoped to that single download's id, and SHALL only update that download's displayed folder/path — every other download in the list SHALL remain unchanged. For a download that is not completed, the sidepanel SHALL NOT offer an active "Renommer" action.

#### Scenario: Open the rename dialog pre-filled with the current name
- **WHEN** the user clicks "Renommer" in the sidepanel of a completed download
- **THEN** the system SHALL open a dialog with the folder name field pre-filled with the item's `rename_folder_name`, and the destination root field empty

#### Scenario: Rename dialog pre-fills the series name for a TV episode, not the season folder
- **WHEN** the user clicks "Renommer" in the sidepanel of a completed TV episode download whose file lives under a `Season NN` folder
- **THEN** the system SHALL pre-fill the dialog with the series' own folder name (e.g. `Breaking Bad (2008)`), and SHALL NOT pre-fill it with the season folder's name (e.g. `Season 01`)

#### Scenario: Rename in place
- **WHEN** the user edits the folder name field only and submits
- **THEN** the system SHALL call `POST /api/v1/downloads/:id/rename` with `{"new_name": <edited value>}`, and on success SHALL update only that download's displayed folder path

#### Scenario: Rename and move in one action
- **WHEN** the user edits both the folder name field and the destination root field and submits
- **THEN** the system SHALL call `POST /api/v1/downloads/:id/rename` with both `new_name` and `destination_parent_dir`, and on success SHALL update only that download's displayed folder path

#### Scenario: Rename blocked by a naming collision
- **WHEN** the rename request fails with the `"rename_target_exists"` error code
- **THEN** the system SHALL keep the dialog open and display a translatable error message indicating the destination already exists, without altering any download

#### Scenario: Siblings unaffected by a rename
- **WHEN** a rename succeeds for one download from its sidepanel
- **THEN** the system SHALL leave every other download's displayed folder path, status, and data unchanged, without requiring a full page reload

#### Scenario: Rename unavailable for a non-completed download
- **WHEN** the user opens the sidepanel for a download whose `status` is not `completed`
- **THEN** the sidepanel SHALL NOT offer an active "Renommer" action
