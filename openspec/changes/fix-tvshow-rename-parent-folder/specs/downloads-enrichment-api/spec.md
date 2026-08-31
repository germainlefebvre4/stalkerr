## ADDED Requirements

### Requirement: Rename Target Folder Name
The `GET /api/v1/downloads` enrichment endpoint SHALL include, for every download with a non-null `download_path`, a `rename_folder_name` field carrying the folder name that a rename operation should treat as the current name of the download's parent folder. When the download's path follows the `.../<series folder>/Season NN/<file>` convention, `rename_folder_name` SHALL be the `<series folder>` name (the same series root the `POST /api/v1/downloads/:id/rename` endpoint targets). For every other download (including movies), `rename_folder_name` SHALL equal `file_info.folder_name`. When `download_path` is null, `rename_folder_name` SHALL be omitted, consistent with `file_info` being nil in that case.

#### Scenario: Movie download reports its own folder as the rename target
- **WHEN** a movie download's path is `/media/movies/Interstellar (2014)/Interstellar (2014).mkv`
- **THEN** the enriched response SHALL include `"rename_folder_name": "Interstellar (2014)"`, equal to `file_info.folder_name`

#### Scenario: TV episode reports the series root as the rename target, not the season folder
- **WHEN** a TV episode download's path is `/media/tvshows/Breaking Bad (2008)/Season 01/Breaking Bad (2008) - S01E01.mkv`
- **THEN** the enriched response SHALL include `"rename_folder_name": "Breaking Bad (2008)"`, distinct from `file_info.folder_name` which SHALL remain `"Season 01"`

#### Scenario: Orphan or incomplete download omits the rename target
- **WHEN** a download has a null `download_path`
- **THEN** the enriched response SHALL omit `rename_folder_name`, consistent with `file_info` being nil
