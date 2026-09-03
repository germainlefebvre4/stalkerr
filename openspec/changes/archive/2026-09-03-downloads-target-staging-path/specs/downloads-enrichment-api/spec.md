## ADDED Requirements

### Requirement: Target and Staging Path Fields
`GET /api/v1/downloads` SHALL include, on `DownloadEnrichedResponse`, two additional optional fields: `target_path` (the intended final destination path for the download, computed before completion) and `staging_path` (the path of the temporary file the current or most recent attempt writes to while transferring). Both fields SHALL be omitted from the JSON response when their underlying `DownloadInfo` column is null, consistent with how `download_path` is already omitted.

#### Scenario: In-progress download exposes target and staging paths
- **WHEN** a download's `status` is `downloading` or `retrying` and its `DownloadInfo.TargetPath` and `DownloadInfo.StagingPath` columns are set
- **THEN** the enriched response SHALL include `target_path` and `staging_path` with those values

#### Scenario: Failed download still exposes target and staging paths
- **WHEN** a download's `status` is `failed` and its `DownloadInfo.TargetPath` and `DownloadInfo.StagingPath` columns are set from the attempt that failed
- **THEN** the enriched response SHALL include `target_path` and `staging_path` with those values, independent of `error_message` content

#### Scenario: Completed download omits target and staging paths
- **WHEN** a download's `status` is `completed` and `DownloadInfo.TargetPath`/`DownloadInfo.StagingPath` have been cleared to null on completion
- **THEN** the enriched response SHALL omit `target_path` and `staging_path`, leaving `download_path` as the sole location field

#### Scenario: Pending download omits target and staging paths
- **WHEN** a download's `status` is `pending` (not yet attempted, `DownloadInfo.TargetPath` and `DownloadInfo.StagingPath` still null)
- **THEN** the enriched response SHALL omit `target_path` and `staging_path`
