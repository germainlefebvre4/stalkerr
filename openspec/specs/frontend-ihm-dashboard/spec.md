# frontend-ihm-dashboard Specification

## Purpose
TBD - created by archiving change frontend-ihm-react-radix. Update Purpose after archive.
## Requirements
### Requirement: Playlist View with Filtering
The frontend SHALL render a comprehensive M3U playlist item table allowing searching by name and filtering by content type (`movie` or `tvshow`).

#### Scenario: Filter items by Content Type and Search Title
- **WHEN** the user selects the "Movies" filter tab and types "Matrix" in the search input
- **THEN** the frontend SHALL fetch and display only playlist items of type `movies` whose names contain "Matrix".

### Requirement: Real-time Monitoring Dashboard
The frontend SHALL present dedicated tabs or tables to display active downloads (incorporating file-size progress bars) and execution logs, using polling to automatically keep content fresh.

#### Scenario: Display active download progress bar and logs status
- **WHEN** the user navigates to the "Downloads" view
- **THEN** the frontend SHALL render active progress bars indicating the percentage of completion for elements in the `downloading` state, and automatically re-fetch the data every 10 seconds.

### Requirement: Modern Light-Themed Radix UI Dialog for File Re-organization
The frontend SHALL offer a modal dialog powered by Radix UI `Dialog` primitives that enables moving entire movie or TV show parent directories to a target directory.

#### Scenario: Confirm folder move from completed items
- **WHEN** the user clicks the "Déplacer ⇄" button on a completed item, selects a destination path, and confirms
- **THEN** the frontend SHALL issue a `POST` request to the media management move endpoint, display a success toast upon success, and refresh the UI state.

