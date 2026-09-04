## Purpose

Defines how a downloaded movie or TV episode's on-disk filename is tagged with the resolution and language of the candidate that was actually downloaded, so the file itself communicates its quality and audio language without needing to consult the catalog.

## ADDED Requirements

### Requirement: Destination filename carries a resolution tag
When a download completes, the destination filename SHALL include the downloaded candidate's resolution as a bracketed tag using the same standard values already used elsewhere in the system (`4K`, `1080p`, `720p`, `480p`), placed immediately after the title/year (or title/season/episode) portion of the filename.

#### Scenario: Resolution tag reflects the downloaded candidate
- **WHEN** a download completes for a candidate with resolution `1080p`
- **THEN** the destination filename SHALL contain the tag `[1080p]`

#### Scenario: Resolution unknown omits the tag
- **WHEN** a download completes for a candidate with no detected resolution
- **THEN** the destination filename SHALL contain no resolution tag

### Requirement: Destination filename carries a language tag
When a download completes for a candidate with a detected language marker (`VF`, `MULTI`, or `VOSTFR`), the destination filename SHALL include that marker as a bracketed tag following the resolution tag (if any).

#### Scenario: Language tag reflects the downloaded candidate
- **WHEN** a download completes for a candidate with language `MULTI`
- **THEN** the destination filename SHALL contain the tag `[MULTI]`

#### Scenario: Language unknown omits the tag
- **WHEN** a download completes for a candidate with no detected language marker
- **THEN** the destination filename SHALL contain no language tag

### Requirement: Destination filename carries a Québec French variant tag
When a download completes for a candidate whose French audio track was detected as Québec French (`VFQ`), the destination filename SHALL include a `[VFQ]` tag following the resolution and language tags (if any), regardless of whether a `VF`/`MULTI`/`VOSTFR` language tag is also present.

#### Scenario: VFQ tag alongside a MULTI language tag
- **WHEN** a download completes for a candidate detected as `MULTI` language whose French track is Québec French
- **THEN** the destination filename SHALL contain both the tag `[MULTI]` and the tag `[VFQ]`

#### Scenario: VFQ tag with no other language marker
- **WHEN** a download completes for a candidate whose only detected language signal is the Québec French variant
- **THEN** the destination filename SHALL contain the tag `[VFQ]` with no other language tag

#### Scenario: Not Québec French omits the tag
- **WHEN** a download completes for a candidate with no detected Québec French signal
- **THEN** the destination filename SHALL contain no `[VFQ]` tag

### Requirement: Tagging applies uniformly across trigger sources
The resolution, language, and Québec French variant tagging rules SHALL apply the same way regardless of whether the download was triggered by the automatic pipeline (missing content or a tier-2 quality upgrade) or by a user-initiated forced download from the UI.

#### Scenario: Automatic pipeline download is tagged
- **WHEN** the automatic `download` pipeline completes a download for a candidate with known resolution and language
- **THEN** the destination filename SHALL carry the same resolution/language/variant tags it would carry had the same candidate been forced from the UI

#### Scenario: Forced download is tagged
- **WHEN** a user-forced download completes for a candidate with known resolution and language
- **THEN** the destination filename SHALL carry the corresponding resolution/language/variant tags

### Requirement: Existing downloads are not retroactively renamed
This tagging rule SHALL apply only to downloads that complete after this capability takes effect. A file downloaded before this capability existed SHALL NOT be renamed, moved, or re-downloaded as a result of this capability.

#### Scenario: Pre-existing file is left untouched
- **WHEN** a movie or episode already has a successfully downloaded file on disk from before this capability existed
- **THEN** that file's name SHALL remain unchanged unless a new download is separately triggered for it
