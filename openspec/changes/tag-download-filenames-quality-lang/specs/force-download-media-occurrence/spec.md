## MODIFIED Requirements

### Requirement: Forced download destination naming avoids overwriting siblings
When a forced download completes, the system SHALL name the destination file so that it includes the occurrence's detected resolution as a distinguishing suffix (for example, a file named `Title (Year) [1080p].ext`), so that it cannot silently replace a sibling occurrence's file already present in the same movie/show destination folder. When the occurrence has no detected resolution, the system SHALL still apply a distinguishing marker unique to that download rather than omit the suffix. In addition, the destination filename SHALL include the occurrence's detected language and, when present, its Québec French variant, as tags following the resolution suffix (for example, `Title (Year) [1080p][MULTI][VFQ].ext`), per the `download-filename-quality-tags` capability. This naming rule SHALL apply only to files produced by a forced download.

#### Scenario: Forced HD download does not overwrite an existing SD file
- **WHEN** a movie already has a downloaded SD file in its Radarr destination folder and a forced download of a different, higher-resolution occurrence of the same movie completes
- **THEN** the new file SHALL be written under a resolution-suffixed name distinct from the existing SD file, and the existing SD file SHALL remain intact

#### Scenario: Forced download with no detected resolution still avoids collision
- **WHEN** a forced download completes for an occurrence with no detected resolution
- **THEN** the system SHALL still write the file under a name that cannot collide with a sibling occurrence's file in the same destination folder

#### Scenario: Forced download filename also carries language and variant tags
- **WHEN** a forced download completes for an occurrence with a detected language and a detected Québec French variant
- **THEN** the destination filename SHALL include the resolution suffix as well as the language and Québec French variant tags

#### Scenario: Automatic pipeline naming is unaffected
- **WHEN** the existing automatic `radarr`/`sonarr`-sourced `download` command completes a download
- **THEN** the destination file name SHALL NOT gain the forced-download-specific anti-collision fallback marker described by this requirement; it follows the shared `download-filename-quality-tags` tagging rule instead (see that capability), which now applies resolution/language/Québec French variant tags to automatic-pipeline downloads too
