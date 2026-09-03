## ADDED Requirements

### Requirement: Language is persisted on ProcessedLine
The processor SHALL store the detected language of each M3U entry in the `language` field of `ProcessedLine`. Valid values are `"VF"`, `"MULTI"`, `"VOSTFR"`, or `NULL` when no language marker can be detected.

#### Scenario: Language stored for a VF entry
- **WHEN** an M3U entry's title or group contains a `VF` marker
- **THEN** the resulting `ProcessedLine` SHALL have `language = "VF"`

#### Scenario: Language stored for a MULTI entry
- **WHEN** an M3U entry's title or group contains a `MULTI` marker
- **THEN** the resulting `ProcessedLine` SHALL have `language = "MULTI"`

#### Scenario: Language stored for a VOSTFR entry
- **WHEN** an M3U entry's title or group contains a `VOSTFR` marker
- **THEN** the resulting `ProcessedLine` SHALL have `language = "VOSTFR"`

#### Scenario: No language marker yields NULL
- **WHEN** an M3U entry's title or group contains no recognized language marker
- **THEN** the resulting `ProcessedLine` SHALL have `language = NULL`

## MODIFIED Requirements

### Requirement: Quality-ordered candidate list for download
When selecting a URL to download a given movie or TV episode, the system SHALL return all eligible `ProcessedLine` records for that content ordered first by language preference, then by resolution preference within the same language tier, then by recency within the same language-and-resolution tier.

Language preference order (ascending priority): `"VF"` (1) → `"MULTI"` (2) → `NULL` (3) → `"VOSTFR"` (4). A candidate with no detected language marker is assumed to be French, consistent with this being a French-language IPTV catalog, and ranks above `VOSTFR` but below an explicitly-tagged `VF`/`MULTI` candidate.

Resolution preference order (ascending priority, applied within a language tier): `720p` (1) → `1080p` (2) → `4K` (3) → `480p` (4) → `NULL` (5).

Eligible candidates are `ProcessedLine` records with `state IN ('processed', 'failed')`.

#### Scenario: VF candidate is selected before MULTI regardless of resolution
- **WHEN** a movie has a `720p` `MULTI` candidate and a `1080p` `VF` candidate
- **THEN** the `VF` candidate SHALL be returned before the `MULTI` candidate

#### Scenario: Unmarked language ranks between MULTI and VOSTFR
- **WHEN** a movie has a `VOSTFR` candidate and a candidate with no detected language
- **THEN** the candidate with no detected language SHALL be returned before the `VOSTFR` candidate

#### Scenario: 720p candidate is selected first when available
- **WHEN** a movie has `VF` candidates at `720p`, `1080p`, and `4K` (same language tier)
- **THEN** the `720p` entry SHALL be returned as the first candidate

#### Scenario: Falls back to 1080p when no 720p exists
- **WHEN** a movie has `VF` candidates at `1080p` and `4K`, but none at `720p` (same language tier)
- **THEN** the `1080p` entry SHALL be returned as the first candidate

#### Scenario: NULL resolution candidates sorted last
- **WHEN** a movie has a `VF` candidate at `1080p` and another `VF` candidate with `NULL` resolution (same language tier)
- **THEN** the `1080p` entry SHALL precede the `NULL`-resolution entry

#### Scenario: Most recent entry preferred within same quality tier
- **WHEN** a movie has two `VF` `720p` `ProcessedLine` entries added at different times
- **THEN** the one with the later `created_at` SHALL be returned first
