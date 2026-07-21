# Downloads Enrichment API

## Description

Enhanced backend API endpoint that returns download information enriched with TMDB metadata and parsed file information. Replaces the simple `listDownloads` handler with a comprehensive view suitable for quality validation and troubleshooting.

## Requirements

### Functional Requirements

**GIVEN** a request to `GET /api/v1/downloads`  
**WHEN** the endpoint is called  
**THEN** the system SHALL:
- Query download_info table with GORM Preload for ProcessedLines.Movie and ProcessedLines.TVShow
- For each download, extract content metadata from the first ProcessedLine
- Parse file metadata using the fileparser package
- Return paginated results with enriched DownloadEnrichedResponse objects

**GIVEN** a request with `?status=completed` query parameter  
**WHEN** the endpoint is called  
**THEN** the system SHALL filter results WHERE status = 'completed'

**GIVEN** a request with `?type=movies` query parameter  
**WHEN** the endpoint is called  
**THEN** the system SHALL filter results WHERE content_type = 'movies' (via JOIN)

**GIVEN** a request with `?problem=missing_year` query parameter  
**WHEN** the endpoint is called  
**THEN** the system SHALL:
- Execute the base query without problem filter
- Apply post-processing filter in Go to include only downloads where file_info.has_year_in_path is false
- Return filtered results with total reflecting pre-filter count

**GIVEN** a DownloadInfo with multiple ProcessedLines  
**WHEN** building the enriched response  
**THEN** the system SHALL use the first ProcessedLine (ordered by created_at ASC)

**GIVEN** a ProcessedLine with a Movie association  
**WHEN** building content metadata  
**THEN** the system SHALL extract:
- type: "movies"
- title: movie.tmdb_title
- year: movie.tmdb_year
- genres: movie.tmdb_genres
- duration: movie.duration
- resolution: processed_line.resolution

**GIVEN** a ProcessedLine with a TVShow association  
**WHEN** building content metadata  
**THEN** the system SHALL:
- Format title as "{tmdb_title} S{season:02d}E{episode:02d}" if season/episode present
- Extract type: "tvshows"
- Extract year: tvshow.tmdb_year
- Extract genres: tvshow.tmdb_genres
- Extract season and episode numbers

**GIVEN** a DownloadInfo with no ProcessedLines (orphan)  
**WHEN** building the enriched response  
**THEN** content SHALL be nil

**GIVEN** a DownloadInfo with null download_path  
**WHEN** building the enriched response  
**THEN** file_info SHALL be nil

### Problem Filters

The system SHALL support these problem filter values:

- `missing_year`: Include only downloads where `!file_info.has_year_in_path`
- `year_mismatch`: Include only downloads where `file_info.year_mismatch`
- `unknown_format`: Include only downloads where `!file_info.is_valid_format`
- `low_quality`: Include only downloads where `detected_resolution in ("480p", "360p")`

### Non-Functional Requirements

- Response time SHALL be < 500ms for 20 results with full enrichment
- The handler SHALL use GORM patterns consistent with existing codebase
- The handler SHALL handle database errors gracefully (500 status with error message)
- Pagination SHALL use limit/offset pattern consistent with other endpoints

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
