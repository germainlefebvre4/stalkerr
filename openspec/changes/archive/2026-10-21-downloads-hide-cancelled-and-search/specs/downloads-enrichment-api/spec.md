# Spec Delta

## MODIFIED Requirements

### Requirement: Status and Type Filtering
`GET /api/v1/downloads` SHALL support `status` and `type` query parameters that restrict the result set to downloads matching the given status or content type respectively. In addition to every concrete download status, `status` SHALL accept the value `active`, which matches every download whose status is not `cancelled`.

#### Scenario: Filtering by status
- **WHEN** a client requests `GET /api/v1/downloads?status=completed`
- **THEN** the response SHALL include only downloads whose status is `completed`

#### Scenario: Filtering by content type
- **WHEN** a client requests `GET /api/v1/downloads?type=movies`
- **THEN** the response SHALL include only downloads whose content type is `movies`

#### Scenario: Filtering by the active pseudo-status
- **WHEN** a client requests `GET /api/v1/downloads?status=active`
- **THEN** the response SHALL include downloads of every status except `cancelled`, and no `cancelled` download

#### Scenario: Omitting status keeps every status
- **WHEN** a client requests `GET /api/v1/downloads` without a `status` parameter
- **THEN** the response SHALL include downloads of every status, including `cancelled`

## ADDED Requirements

### Requirement: Downloads Title Search
`GET /api/v1/downloads` SHALL support a `search` query parameter that restricts the result set to downloads matching the given text, case-insensitively and as a substring, against any of: the linked movie or TV show title, the linked processed line's `tvg_name`, or the file name of the download's path. An empty or absent `search` SHALL apply no constraint. The `search` constraint SHALL combine with `status`, `type` and `problem`, and SHALL be reflected in `total` and `total_pages`.

#### Scenario: Search matches the content title
- **WHEN** a client requests `GET /api/v1/downloads?search=peacemaker` and a download is linked to a TV show titled "Peacemaker"
- **THEN** the response SHALL include that download and SHALL NOT include downloads whose title, `tvg_name` and file name all lack "peacemaker"

#### Scenario: Search is case-insensitive
- **WHEN** a client requests `GET /api/v1/downloads?search=PEACEMAKER`
- **THEN** the response SHALL contain the same downloads as for `search=peacemaker`

#### Scenario: Search falls back to the M3U name and file name
- **WHEN** a download has no linked movie or TV show but its `tvg_name` or download file name contains the search text
- **THEN** the response SHALL include that download

#### Scenario: Search combines with other filters and pagination
- **WHEN** a client requests `GET /api/v1/downloads?search=peacemaker&status=active&limit=20&offset=0`
- **THEN** the response SHALL include only non-cancelled downloads matching the text, and `total` SHALL be the count of that filtered set

#### Scenario: Search does not duplicate a download
- **WHEN** a download has several processed lines that each match the search text
- **THEN** the download SHALL appear exactly once in the response and be counted once in `total`

### Requirement: Downloads Status-Ranked Ordering
`GET /api/v1/downloads` SHALL order downloads first by status rank — `downloading`, `retrying` and `pending` first, then `completed`, then `failed`, then `cancelled` last — and then by `updated_at` descending within a rank. Any other status SHALL rank after `failed` and before `cancelled`. This ordering SHALL be identical whether or not the `problem`, `type` or `search` parameters are set.

#### Scenario: Cancelled downloads sort last
- **WHEN** a client requests `GET /api/v1/downloads` without a `status` filter and a `cancelled` download was updated more recently than a `completed` one
- **THEN** the `completed` download SHALL appear before the `cancelled` download in the response

#### Scenario: In-progress downloads sort first
- **WHEN** the result set contains `downloading`, `completed`, `failed` and `cancelled` downloads
- **THEN** the `downloading` download SHALL appear first and the `cancelled` download last

#### Scenario: Recency orders downloads within a rank
- **WHEN** two `completed` downloads are returned
- **THEN** the one with the more recent `updated_at` SHALL appear first

### Requirement: Hidden Cancelled Count
When `status=active` is requested, the paginated response of `GET /api/v1/downloads` SHALL include an integer field `cancelled_hidden` equal to the number of `cancelled` downloads that match the request's `type`, `problem` and `search` constraints and were excluded by `active`. For any other `status` value, `cancelled_hidden` SHALL be `0`.

#### Scenario: Count of cancelled downloads excluded by active
- **WHEN** a client requests `GET /api/v1/downloads?status=active&search=peacemaker` and 3 matching downloads are `cancelled`
- **THEN** the response SHALL have `cancelled_hidden` equal to `3` and SHALL NOT list those downloads

#### Scenario: No hidden count for explicit statuses
- **WHEN** a client requests `GET /api/v1/downloads?status=cancelled` or omits `status`
- **THEN** the response SHALL have `cancelled_hidden` equal to `0`
