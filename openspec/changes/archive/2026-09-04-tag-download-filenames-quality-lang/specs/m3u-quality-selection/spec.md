## ADDED Requirements

### Requirement: Québec French variant is persisted independently of language
The processor SHALL detect when an M3U entry's title or group indicates that its French audio track is specifically Québec French (marker `VFQ`), and SHALL persist this as a signal independent of, and combinable with, the `language` field (`VF`/`MULTI`/`VOSTFR`/`NULL`). This detection SHALL NOT replace or suppress detection of the entry's `VF`/`MULTI`/`VOSTFR` marker when one is also present.

#### Scenario: VFQ detected alongside a MULTI marker
- **WHEN** an M3U entry's title contains both a `MULTI` marker and a `VFQ` marker (for example `Multi.Vfq.720P`)
- **THEN** the resulting `ProcessedLine` SHALL have `language = "MULTI"` and SHALL also record the Québec French variant

#### Scenario: VFQ detected with no other language marker
- **WHEN** an M3U entry's title contains a `VFQ` marker but no `VF`/`MULTI`/`VOSTFR` marker
- **THEN** the resulting `ProcessedLine` SHALL record the Québec French variant, independent of the value of `language`

#### Scenario: No VFQ marker present
- **WHEN** an M3U entry's title or group contains no `VFQ` marker
- **THEN** the resulting `ProcessedLine` SHALL NOT record the Québec French variant

## MODIFIED Requirements

### Requirement: Quality-ordered candidate list for download
When selecting a URL to download a given movie or TV episode, the system SHALL return all eligible `ProcessedLine` records for that content ordered first by language preference, then by resolution preference within the same language tier, then by the Québec French variant tie-break within the same language-and-resolution tier, then by recency within the same language, resolution, and variant tier.

Language preference order (ascending priority): `"VF"` (1) → `"MULTI"` (2) → `NULL` (3) → `"VOSTFR"` (4). A candidate with no detected language marker is assumed to be French, consistent with this being a French-language IPTV catalog, and ranks above `VOSTFR` but below an explicitly-tagged `VF`/`MULTI` candidate.

Resolution preference order (ascending priority, applied within a language tier): `720p` (1) → `1080p` (2) → `4K` (3) → `480p` (4) → `NULL` (5).

Québec French variant tie-break (applied within the same language-and-resolution tier): a candidate without the Québec French variant SHALL be preferred over an otherwise-equally-ranked candidate with the Québec French variant.

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

#### Scenario: France French preferred over Québec French at the same language and resolution
- **WHEN** a movie has a `VF` `720p` candidate without the Québec French variant and a `VF` `720p` candidate with the Québec French variant
- **THEN** the candidate without the Québec French variant SHALL be returned first
