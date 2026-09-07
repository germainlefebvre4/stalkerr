# Downloads Enrichment API

## Purpose

Enhanced backend API endpoint that returns download information enriched with TMDB metadata and parsed file information. Replaces the simple `listDownloads` handler with a comprehensive view suitable for quality validation and troubleshooting.

## Requirements

### Requirement: Enriched Downloads Listing
`GET /api/v1/downloads` SHALL return paginated download records enriched with content metadata (from the associated Movie or TVShow via the download's first ProcessedLine) and parsed file metadata (via the fileparser package), replacing the plain download listing.

#### Scenario: Enriched listing includes content and file metadata
- **WHEN** a client requests `GET /api/v1/downloads`
- **THEN** each returned download SHALL include a `content` object built from its first ProcessedLine's associated Movie or TVShow (ordered by `created_at` ascending) when one exists, and a `file_info` object parsed from its `download_path` when non-null

### Requirement: Status and Type Filtering
`GET /api/v1/downloads` SHALL support `status` and `type` query parameters that restrict the result set to downloads matching the given status or content type respectively.

#### Scenario: Filtering by status
- **WHEN** a client requests `GET /api/v1/downloads?status=completed`
- **THEN** the response SHALL include only downloads whose status is `completed`

#### Scenario: Filtering by content type
- **WHEN** a client requests `GET /api/v1/downloads?type=movies`
- **THEN** the response SHALL include only downloads whose content type is `movies`

### Requirement: Detected-Problem Filter Values
The `problem` query parameter on `GET /api/v1/downloads` SHALL support the values `missing_year` (`file_info.has_year_in_path` is false), `year_mismatch` (`file_info.year_mismatch` is true), `unknown_format` (`file_info.is_valid_format` is false), and `low_quality` (detected resolution is `480p` or `360p`), each restricting the result set to downloads matching that condition.

#### Scenario: Filtering by missing_year
- **WHEN** a client requests `GET /api/v1/downloads?problem=missing_year`
- **THEN** the response SHALL include only downloads where `file_info.has_year_in_path` is false

#### Scenario: Filtering by low_quality
- **WHEN** a client requests `GET /api/v1/downloads?problem=low_quality`
- **THEN** the response SHALL include only downloads whose detected resolution is `480p` or `360p`

### Requirement: Content Metadata Field Mapping
When building `content` from a Movie association, the system SHALL set `type` to `movies` and populate title, year, genres, duration, and resolution from the movie and processed line. When building `content` from a TVShow association, the system SHALL set `type` to `tvshows`, format the title as `{title} S{season:02d}E{episode:02d}` when season/episode are present, and populate year, genres, season, and episode.

#### Scenario: Movie content metadata
- **WHEN** a download's first ProcessedLine has a Movie association
- **THEN** `content.type` SHALL be `movies` and SHALL include title, year, genres, duration, and resolution from that movie/processed line

#### Scenario: TV show content metadata includes formatted episode title
- **WHEN** a download's first ProcessedLine has a TVShow association with season 5 and episode 14
- **THEN** `content.type` SHALL be `tvshows` and `content.title` SHALL be formatted as `{tmdb_title} S05E14`

### Requirement: Missing Association Handling
When a download has no ProcessedLines, `content` SHALL be nil. When a download has a null `download_path`, `file_info` SHALL be nil.

#### Scenario: Orphan download has no content
- **WHEN** a download has no ProcessedLines
- **THEN** `content` SHALL be nil in the response

#### Scenario: Download with no path has no file_info
- **WHEN** a download's `download_path` is null
- **THEN** `file_info` SHALL be nil in the response

### Requirement: Database Error Response
When a database error occurs while serving `GET /api/v1/downloads`, the system SHALL respond with HTTP 500 and a JSON body identifying the error (`error`/`message` fields).

#### Scenario: Database failure returns 500 with error body
- **WHEN** the database query for `GET /api/v1/downloads` fails
- **THEN** the response SHALL have status 500 and a JSON body of the form `{"error": "database_error", "message": "..."}`

### Requirement: Response Time Budget
`GET /api/v1/downloads` SHALL respond in under 500ms for a page of 20 fully-enriched results under normal load.

#### Scenario: Standard page responds within budget
- **WHEN** a client requests a page of 20 results with full enrichment under normal load
- **THEN** the response SHALL be returned in under 500ms

### Requirement: Multi-Value Problem Filter
The `problem` query parameter on `GET /api/v1/downloads` SHALL accept a comma-separated list of two or more of the existing problem values (`missing_year`, `year_mismatch`, `unknown_format`, `low_quality`), in addition to continuing to accept a single value as before. When multiple values are given, the system SHALL include a download in the filtered result if it matches at least one of the listed values (logical OR across the listed problems).

#### Scenario: Combined filter returns the union of matching downloads
- **WHEN** a client requests `GET /api/v1/downloads?problem=missing_year,unknown_format`
- **THEN** the response SHALL include every download where `!file_info.has_year_in_path` OR `!file_info.is_valid_format`, without requiring both conditions to hold

#### Scenario: Single value keeps its existing exact-match behavior
- **WHEN** a client requests `GET /api/v1/downloads?problem=missing_year`
- **THEN** the response SHALL behave exactly as before this change, including only downloads where `!file_info.has_year_in_path`

### Requirement: Problem Filter Pagination Accuracy
When the `problem` query parameter is set on `GET /api/v1/downloads` — whether to a single value or a comma-separated list of values — the system SHALL apply the problem filter (matching any listed value) to the full status/type-matched result set before computing `total` and `total_pages` and before slicing the requested `limit`/`offset` page, so that pagination metadata and the returned page reflect the actual filtered result count. This supersedes the previously documented behavior of reporting the pre-filter count as `total` when `problem` is set.

#### Scenario: Total reflects the problem-filtered count
- **WHEN** a client requests `GET /api/v1/downloads?problem=missing_year&limit=20&offset=0` and, of the downloads matching the active `status`/`type` filters, 45 have `!file_info.has_year_in_path`
- **THEN** the response SHALL report `total: 45`, `total_pages` computed from 45 and the given `limit`, and `data` containing the first 20 of those 45 filtered downloads

#### Scenario: Later pages stay consistent with the filtered total
- **WHEN** a client requests `GET /api/v1/downloads?problem=missing_year&limit=20&offset=20` under the same conditions as the previous scenario
- **THEN** the response SHALL return items 21–40 of the problem-filtered set (not of the pre-filter set), consistent with the `total` reported for `offset=0`

#### Scenario: Total reflects the combined-filter count for a multi-value request
- **WHEN** a client requests `GET /api/v1/downloads?problem=missing_year,unknown_format&limit=20&offset=0` and, of the downloads matching the active `status`/`type` filters, 60 have `!file_info.has_year_in_path` or `!file_info.is_valid_format`
- **THEN** the response SHALL report `total: 60`, `total_pages` computed from 60 and the given `limit`, and `data` containing the first 20 of those 60 filtered downloads

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

## API Contract

### Request

```
GET /api/v1/downloads?limit=20&offset=0&status=completed&type=movies&problem=missing_year
```

**Query Parameters**:
- `limit` (optional, default: 20): Number of results per page
- `offset` (optional, default: 0): Pagination offset
- `status` (optional): Filter by download status (pending, downloading, completed, failed, retrying)
- `type` (optional): Filter by content type (movies, tvshows, channels, uncategorized)
- `problem` (optional): Filter by detected issues (missing_year, year_mismatch, unknown_format, low_quality)

### Response

```json
{
  "data": [
    {
      "id": 123,
      "url": "https://example.com/stream.m3u8",
      "status": "completed",
      "download_path": "/media/movies/La.Cite.de.Dieu.2002.1080p.BluRay/video.mkv",
      "file_size": 4516241408,
      "bytes_downloaded": 4516241408,
      "total_bytes": 4516241408,
      "retry_count": 0,
      "error_message": null,
      "updated_at": "2026-07-20T15:30:00Z",
      
      "content": {
        "type": "movies",
        "title": "La Cité de Dieu",
        "year": 2002,
        "resolution": "1080p",
        "genres": "Crime, Drama",
        "duration": 130
      },
      
      "file_info": {
        "extension": ".mkv",
        "folder_name": "La.Cite.de.Dieu.2002.1080p.BluRay",
        "file_name": "video.mkv",
        "has_year_in_path": true,
        "year_mismatch": false,
        "detected_year": 2002,
        "detected_resolution": "1080p",
        "is_valid_format": true
      }
    }
  ],
  "total": 150,
  "limit": 20,
  "offset": 0,
  "total_pages": 8
}
```

### Error Responses

```json
{
  "error": "database_error",
  "message": "failed to fetch downloads"
}
```

## Implementation Details

**Implementation Notes**:
- The handler SHALL use GORM patterns consistent with the existing codebase.
- Pagination SHALL use the limit/offset pattern consistent with other endpoints.

**File Location**: `internal/api/handlers_frontend.go`

**Handler Name**: `listDownloadsEnriched(c *gin.Context)`

**GORM Query Pattern**:
```go
query := db.Model(&models.DownloadInfo{}).
    Preload("ProcessedLines.Movie").
    Preload("ProcessedLines.TVShow")

if status != "" {
    query = query.Where("status = ?", status)
}

if contentType != "" {
    query = query.Joins("JOIN processed_lines pl ON pl.download_info_id = download_info.id").
        Where("pl.content_type = ?", contentType).
        Distinct()
}

query.Order("updated_at desc").Limit(limit).Offset(offset).Find(&downloads)
```

**Enrichment Flow**:
```go
for _, dl := range downloads {
    resp := DownloadEnrichedResponse{
        // Copy base fields
        ID: dl.ID,
        URL: dl.URL,
        // ... other fields
    }
    
    // Extract content from first ProcessedLine
    if len(dl.ProcessedLines) > 0 {
        resp.Content = buildContentInfo(dl.ProcessedLines[0])
    }
    
    // Parse file metadata
    if dl.DownloadPath != nil {
        resp.FileInfo = fileparser.Parse(*dl.DownloadPath, contentYear)
    }
    
    enriched = append(enriched, resp)
}
```

**Type Definitions** (in `internal/api/types.go`):
```go
type DownloadEnrichedResponse struct {
    ID              uint       `json:"id"`
    URL             string     `json:"url"`
    Status          string     `json:"status"`
    DownloadPath    *string    `json:"download_path,omitempty"`
    FileSize        *int64     `json:"file_size,omitempty"`
    BytesDownloaded *int64     `json:"bytes_downloaded,omitempty"`
    TotalBytes      *int64     `json:"total_bytes,omitempty"`
    RetryCount      int        `json:"retry_count"`
    ErrorMessage    *string    `json:"error_message,omitempty"`
    UpdatedAt       time.Time  `json:"updated_at"`
    Content         *ContentInfo `json:"content,omitempty"`
    FileInfo        *FileInfo    `json:"file_info,omitempty"`
}

type ContentInfo struct {
    Type       string  `json:"type"`
    Title      string  `json:"title"`
    Year       *int    `json:"year,omitempty"`
    Resolution *string `json:"resolution,omitempty"`
    Season     *int    `json:"season,omitempty"`
    Episode    *int    `json:"episode,omitempty"`
    Genres     *string `json:"genres,omitempty"`
    Duration   *int    `json:"duration,omitempty"`
}

type FileInfo struct {
    Extension       string  `json:"extension"`
    FolderName      string  `json:"folder_name"`
    FileName        string  `json:"file_name"`
    HasYearInPath   bool    `json:"has_year_in_path"`
    YearMismatch    bool    `json:"year_mismatch"`
    DetectedYear    *int    `json:"detected_year,omitempty"`
    DetectedRes     *string `json:"detected_resolution,omitempty"`
    IsValidFormat   bool    `json:"is_valid_format"`
}
```

## Testing

**Manual Tests** (curl):
```bash
# Basic query
curl http://localhost:8080/api/v1/downloads

# With filters
curl "http://localhost:8080/api/v1/downloads?status=completed&type=movies"

# With problem filter
curl "http://localhost:8080/api/v1/downloads?problem=missing_year"

# Pagination
curl "http://localhost:8080/api/v1/downloads?limit=10&offset=20"
```

**Validation**:
- Verify content metadata is populated for movies/tvshows
- Verify file_info is populated when download_path exists
- Verify filtering works correctly
- Verify pagination calculates total_pages correctly
- Check SQL query logs for N+1 issues (should see Preload working)

## Integration Points

**Dependencies**:
- `internal/fileparser` package (for Parse function)
- `internal/models` package (DownloadInfo, ProcessedLine, Movie, TVShow)
- `internal/database` package (Get() function)

**Route Registration** (in `internal/api/api.go`):
```go
v1.GET("/downloads", s.listDownloadsEnriched)
```

## Performance Notes

**Expected Performance**:
- Query time: ~50-100ms for 20 results with Preload
- Parsing time: ~1ms per download
- Total response time: <500ms

**Optimization Opportunities** (future):
- Add materialized view for common queries
- Cache parsed file_info in database
- Implement cursor-based pagination for very large datasets

**Current Approach**: Optimize for simplicity and correctness; performance is acceptable for expected load (<1000 downloads)
