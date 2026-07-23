## MODIFIED Requirements

### Requirement: Display of file_info metadata
The system SHALL display physical file path and technical metadata for each download in a clean, highly readable, and structured layout. The physical paths (folder name and file name) SHALL be kept in a compact monospace box, while the technical specifications (format, resolution, size, and duration if available) SHALL be extracted and displayed inline outside the box. Crucially, validation status/indicators SHALL be rendered as structured, colored badge chips.

#### Scenario: Rendering compact file info and metadata outside of container
- **WHEN** rendering a download card containing file_info
- **THEN** the system SHALL display folder/file names inside a compact monospace box, display technical specs (Format, Resolution, Size, and optional Duration) inline below it, and display validation status chips (Year validity, Format validity, and Low Quality warning if applicable) as styled badge elements next to the specs.

#### Scenario: Display low-quality warning indicator
- **WHEN** the download's detected resolution is 480p or 360p
- **THEN** the system SHALL display a warning chip (e.g. `⚠️ Basse qualité (<resolution>)`) next to the validation indicators.
