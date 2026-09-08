## MODIFIED Requirements

### Requirement: Download Item Details Sidepanel
The Downloads tab SHALL open a sliding sidepanel (drawer), based on Radix UI `Dialog`, when the user clicks or taps any download's summary row (capability `downloads-display-ui`). The sidepanel SHALL display, for the selected download:
- A status section: the status badge — following the same viewport-dependent emoji/text rule as the list's status badge (capability `downloads-display-ui`: emoji only on mobile, emoji plus text label on desktop) — and, when the status is `downloading` or `retrying`, a progress bar with percentage and downloaded/total size.
- A file section: the folder and file names (from `file_info`) and the full `download_path` when the download is `completed`. When `download_path` is not yet available, the file section SHALL instead show, when present, the planned final location (`target_path`) labeled distinctly as a planned/not-yet-final location, and the temporary file location (`staging_path`) labeled distinctly as a temporary file that may no longer exist on disk. When none of `download_path`, `target_path`, or `staging_path` are available, the file section SHALL fall back to the source `url`.
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

#### Scenario: In-progress download shows the planned location and the temporary file
- **WHEN** the sidepanel is open for a download whose `status` is `downloading` or `retrying`, with `target_path` and `staging_path` both present and `download_path` absent
- **THEN** the file section SHALL show the planned final location from `target_path` (labeled as planned, not yet final) and the current temporary file location from `staging_path` (labeled as temporary), instead of falling back to the source `url`

#### Scenario: Failed download shows where the file was going and where the attempt was writing
- **WHEN** the sidepanel is open for a download whose `status` is `failed`, with `target_path` and `staging_path` present and `download_path` absent
- **THEN** the file section SHALL show `target_path` and `staging_path` with the same distinct labels used for an in-progress download, giving the user the intended destination and the temporary write location for troubleshooting the failure

#### Scenario: Staging path is shown without implying the file still exists
- **WHEN** the file section renders `staging_path` for a `failed` download
- **THEN** the label SHALL indicate the temporary file may already have been removed, and the sidepanel SHALL NOT offer any action (e.g. open/browse) that assumes the file is still present at that path

#### Scenario: Completed download shows only the final path
- **WHEN** the sidepanel is open for a download whose `status` is `completed`
- **THEN** the file section SHALL show only `download_path` as before; `target_path` and `staging_path` SHALL be absent by this point and SHALL NOT be rendered

#### Scenario: Sidepanel status badge shows emoji only on mobile
- **WHEN** the sidepanel is open on a viewport narrower than the mobile breakpoint
- **THEN** the status section's badge SHALL display only the status's emoji, with no text label, matching the list's mobile status badge

#### Scenario: Sidepanel status badge shows emoji and text on desktop
- **WHEN** the sidepanel is open on a viewport at or above the mobile breakpoint
- **THEN** the status section's badge SHALL display the status's emoji followed by its text label, matching the list's desktop status badge
