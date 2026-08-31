## REMOVED Requirements

### Functional Requirements
**Reason**: The Downloads tab no longer renders each download as a self-contained detailed card. Fetching, auto-refresh, loading state, and filter-dropdown behavior are unchanged and are re-stated under the new "Download List Summary Row" requirement below; the card-specific rendering rules (title/icon/badge layout, inline technical specs, validation chips, genres, error section, year-mismatch tooltip) are superseded by the summary row (row content) and the sidepanel (full detail).
**Migration**: See "Requirement: Download List Summary Row" below for the summary row content, and capability `downloads-details-sidepanel` for the full detail previously shown inline on the card.

**GIVEN** the Downloads tab is active
**WHEN** the component loads
**THEN** the system SHALL:
- Fetch enriched downloads from GET /api/v1/downloads
- Display each download as a card with content title, technical specs, and status
- Auto-refresh every 5 seconds
- Show loading state during initial fetch

**GIVEN** a download with content metadata
**WHEN** rendering the download card
**THEN** the system SHALL display:
- Content title (large, bold) from content.title
- Year in parentheses from content.year
- Content type icon (🎬 for movies, 📺 for tvshows)
- Status badge (✅ Complété, ❌ Échec, ⏳ En cours)

**GIVEN** a download with file_info metadata
**WHEN** rendering the download card
**THEN** the system SHALL:
- Display the folder/file names inside a compact monospace box.
- Extract and display technical specifications (Format, Resolution, Size, and optional Duration in minutes with 🕒 icon) inline below the monospace box.
- Render validation status/indicators as structured, colored badge chips (Year validity: `✅ Année OK` or `⚠️ Année manquante`/`⚠️ Année incorrecte`, Format validity: `✅ Format OK` or `⚠️ Format inconnu`) next to the technical specs line.

**GIVEN** the download's detected resolution is 480p or 360p
**WHEN** rendering the download card
**THEN** the system SHALL display a warning chip `⚠️ Basse qualité (<resolution>)` using `badge badge-pending` next to the validation indicators.

**GIVEN** a download with genres
**WHEN** rendering the download card
**THEN** the system SHALL display genres with 🎭 icon

**GIVEN** a failed download
**WHEN** rendering the download card
**THEN** the system SHALL:
- Show error message in a highlighted error section
- Display retry count if > 0
- Show partial download progress if available

**GIVEN** a download with year_mismatch flag
**WHEN** rendering the download card
**THEN** the system SHALL show ⚠️ indicator with tooltip "Année dans le path différente de TMDB"

**GIVEN** filter dropdowns (status, type, problem)
**WHEN** a filter is changed
**THEN** the system SHALL:
- Update query parameters
- Re-fetch downloads with new filters
- Reset to page 1

### Display Priority
**Reason**: This priority ordering described the stacked layout of a single detailed card (title, folder path, technical specs, badges, genres). That layout no longer exists; the summary row shows only title and status/progress, and the sidepanel now owns its own section ordering.
**Migration**: See capability `downloads-details-sidepanel` for the sidepanel's section ordering.

1. **Title** (most prominent) - 1.25rem, bold
2. **Folder path** (secondary) - 0.85rem, code style
3. **Technical specs** (tertiary) - 0.8rem, inline
4. **Status/badges** (corner) - badge component
5. **Genres** (optional) - 0.8rem, muted

### Requirement: Completion Date on Download Cards
**Reason**: Completion date is no longer shown on a card's technical specs row (there is no card); it moves into the sidepanel's technical specifications section.
**Migration**: See capability `downloads-details-sidepanel`, requirement "Sidepanel Technical Specifications".

**GIVEN** a download whose `completed_at` timestamp is set
**WHEN** rendering the download card
**THEN** the system SHALL display the formatted completion date inline in the technical specs row, alongside Format, Resolution, Size, and Duration, using the same locale-aware date formatting as the Playlist tab's `downloaded_at` column.

**GIVEN** a download whose `completed_at` timestamp is not set (pending, downloading, failed, or retrying)
**WHEN** rendering the download card
**THEN** the system SHALL NOT display a completion date or a placeholder for one in the technical specs row.

**GIVEN** the Downloads tab rendered on a viewport narrower than the mobile breakpoint
**WHEN** a download card shows its completion date
**THEN** the system SHALL display the same completion date, in the same technical specs row, as on desktop, without truncation or omission.

#### Scenario: Completed download shows its completion date
- **WHEN** a download's `completed_at` is `2026-08-27T14:33:00Z` and the technical specs row is rendered
- **THEN** the row SHALL include the formatted completion date alongside Format, Resolution, Size, and Duration

#### Scenario: In-progress download shows no completion date
- **WHEN** a download's `completed_at` is absent and the technical specs row is rendered
- **THEN** the row SHALL omit any completion date element or placeholder, leaving the existing Format/Resolution/Size/Duration content unchanged

#### Scenario: Completion date renders identically on mobile
- **WHEN** the Downloads tab is rendered on a viewport narrower than the mobile breakpoint and a download card has a completion date
- **THEN** the card SHALL display the same completion date in the technical specs row as it would on desktop

### Requirement: Rename Download Item
**Reason**: The "Renommer" action is no longer available directly on the list; it is triggered from the sidepanel instead.
**Migration**: See capability `downloads-details-sidepanel`, requirement "Move and Rename Actions in the Sidepanel".

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

### Requirement: Error Banner Gated on Failed Status
**Reason**: The error banner is no longer rendered inline on a card; it moves into the sidepanel's "Erreur" section, gated by the same rule.
**Migration**: See capability `downloads-details-sidepanel`, requirement "Sidepanel Error Message Gated on Failed Status".

The Downloads tab SHALL render a download card's error banner only when that download's current `status` is `failed`. The presence of a non-empty `error_message` alone SHALL NOT be sufficient to render the banner, since `error_message` may be a value carried over from a prior attempt on the same record.

#### Scenario: Completed download with a leftover error_message does not show the banner
- **WHEN** a download card's `status` is `completed` and its `error_message` is non-empty (leftover from an earlier failed attempt)
- **THEN** the system SHALL NOT render the error banner on that card

#### Scenario: Failed download still shows the banner
- **WHEN** a download card's `status` is `failed` and its `error_message` is non-empty
- **THEN** the system SHALL render the error banner with the error message, as before

#### Scenario: In-progress download with a leftover error_message does not show the banner
- **WHEN** a download card's `status` is `downloading` or `retrying` and its `error_message` is non-empty (leftover from an earlier failed attempt)
- **THEN** the system SHALL NOT render the error banner on that card

## ADDED Requirements

### Requirement: Download List Summary Row
The Downloads tab SHALL fetch enriched downloads from `GET /api/v1/downloads`, auto-refresh every 5 seconds, and show a loading state during the initial fetch, same as before. Instead of a detailed card, each download SHALL render as a compact summary row: on viewports at or above the mobile breakpoint, a table row; below it, a list card matching the Playlist tab's `mobile-list-card` pattern. Each summary row SHALL display, at minimum: the content type icon (🎬/📺/🔗) and title (falling back to the file name or URL when no `content.title` is available) with the year in parentheses when known, the status badge (same statuses/labels as before), and a compact progress indicator (percentage and/or size) for downloads whose status is `downloading` or `retrying`. The summary row SHALL NOT render file paths, technical specification chips, validation badges, genres, or error messages inline — that detail is available in the sidepanel (capability `downloads-details-sidepanel`). Each summary row SHALL be clickable/tappable to open that sidepanel for the corresponding download.

The status, type, and problem filter dropdowns SHALL keep updating query parameters, re-fetching downloads with the new filters, and resetting to page 1 on change, unchanged from before.

#### Scenario: Desktop table row shows only summary content
- **WHEN** the Downloads tab is rendered on a viewport at or above the mobile breakpoint
- **THEN** each download SHALL appear as one table row showing its type icon, title, year, status badge, and (for `downloading`/`retrying` items) a compact progress indicator, with no inline file path, technical specs, validation badges, genres, or error message

#### Scenario: Mobile card shows only summary content
- **WHEN** the Downloads tab is rendered on a viewport narrower than the mobile breakpoint
- **THEN** each download SHALL appear as one list card showing its type icon, title, year, and status badge, matching the visual pattern used by the Playlist tab's mobile list cards

#### Scenario: Missing content title falls back gracefully
- **WHEN** a download has no `content.title`
- **THEN** the summary row SHALL display the download's file name (derived from `download_path`) or, failing that, its `url`, instead of leaving the title blank

#### Scenario: Clicking a row opens the sidepanel
- **WHEN** the user clicks (or taps) a download's summary row
- **THEN** the frontend SHALL open the details sidepanel for that download (capability `downloads-details-sidepanel`)
