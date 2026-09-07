## ADDED Requirements

### Requirement: Downloads Filter Option Sets
The Downloads tab SHALL offer three filter controls — status, type, and problem — each with a fixed set of options mapping to specific backend query values: the status filter maps "Tous" / "Complétés" / "En cours" / "Échoués" to no filter / `completed` / `downloading` / `failed`; the type filter maps "Tous" / "Films" / "Séries" to no filter / `movies` / `tvshows`; the problem filter maps "Aucun" / "Année manquante" / "Année incorrecte" / "Format inconnu" / "Basse qualité" to no filter / `missing_year` / `year_mismatch` / `unknown_format` / `low_quality`.

#### Scenario: Selecting a status filter option applies the mapped backend value
- **WHEN** the user selects "Échoués" in the status filter
- **THEN** the frontend SHALL query downloads with `status=failed`

#### Scenario: Selecting a problem filter option applies the mapped backend value
- **WHEN** the user selects "Année manquante" in the problem filter
- **THEN** the frontend SHALL query downloads with `problem=missing_year`

#### Scenario: "Tous" resets a filter to no constraint
- **WHEN** the user selects "Tous" in the status or type filter
- **THEN** the frontend SHALL query downloads without that filter parameter

### Requirement: Graceful Handling of Missing Enrichment Fields
The Downloads list SHALL render correctly when optional enrichment fields (`content`, `file_info`, and their nested fields) are absent from a download record, without errors or broken layout.

#### Scenario: Download with no content metadata renders without a title
- **WHEN** a download has no `content` object (e.g. an orphan download with no matched Movie/TVShow)
- **THEN** the row/card SHALL render its other fields normally and fall back gracefully in place of the missing title

#### Scenario: Download with no file_info renders without file details
- **WHEN** a download has a null `download_path` and consequently no `file_info`
- **THEN** the row/card SHALL render without file/format details, without erroring
