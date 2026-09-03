## ADDED Requirements

### Requirement: Tier-2 stream requires a strictly better candidate
A tier-2 stream SHALL only be built for a movie or series-season when its best untried candidate — ranked by the `m3u-quality-selection` ordering (language, then resolution) — is strictly better than the candidate already downloaded for that same movie or series-season. A leftover untried candidate that is equal to or worse than the already-downloaded candidate SHALL NOT cause a tier-2 stream to be built.

#### Scenario: Untried candidate with a worse language is not offered as an upgrade
- **WHEN** a movie's already-downloaded candidate has `language = "VF"` and its only untried candidate has `language = "VOSTFR"`
- **THEN** no tier-2 stream SHALL be built for that movie

#### Scenario: Untried candidate with a worse resolution in the same language is not offered as an upgrade
- **WHEN** a movie's already-downloaded candidate has `language = "VF"`, `resolution = "720p"` and its only untried candidate has `language = "VF"`, `resolution = "4K"`
- **THEN** no tier-2 stream SHALL be built for that movie, since `720p` already outranks `4K` in the resolution preference order

#### Scenario: Untried candidate with a strictly better language triggers an upgrade
- **WHEN** a movie's already-downloaded candidate has `language = "MULTI"` and an untried candidate has `language = "VF"`
- **THEN** a tier-2 stream SHALL be built for that movie, targeting the `VF` candidate

#### Scenario: Untried candidate with a strictly better resolution in the same language triggers an upgrade
- **WHEN** a movie's already-downloaded candidate has `language = "VF"`, `resolution = "4K"` and an untried candidate has `language = "VF"`, `resolution = "720p"`
- **THEN** a tier-2 stream SHALL be built for that movie, targeting the `720p` candidate

#### Scenario: A satisfied tier-2 upgrade replaces the original file rather than duplicating it
- **WHEN** a tier-2 stream is built and its download completes successfully
- **THEN** the resulting file SHALL be written to the same destination path as the movie's or episode's original download, replacing it, rather than being written to a different path alongside it
