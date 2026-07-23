## Architecture

### Data Flow

```
Frontend Request
    ↓
GET /api/v1/downloads?status=X&type=Y&problem=Z
    ↓
listDownloadsEnriched()
    ├─ Query DB with GORM Preload
    │  (ProcessedLines.Movie, ProcessedLines.TVShow)
    ├─ For each DownloadInfo:
    │  ├─ Extract content from first ProcessedLine
    │  ├─ Parse file metadata from download_path
    │  └─ Build enriched response
    └─ Filter by "problem" (post-processing)
    ↓
JSON Response (DownloadEnrichedResponse[])
    ↓
Frontend renders enriched cards
```

### Database Relations (Existing)

```
download_info 1─────┐
                    │
            ┌───────┴──────────┐
            │ processed_lines  │
            └───────┬──────────┘
                    │
        ┌───────────┼───────────┐
        │           │           │
    ┌───▼────┐  ┌───▼──────┐  ...
    │ movies │  │ tvshows  │
    └────────┘  └──────────┘
```

### Component Design

#### 1. File Parser (Pure Logic)

**Location**: `internal/fileparser/parser.go`

**Purpose**: Extract technical metadata from file paths with no external dependencies

**Interface**:
```go
func Parse(downloadPath string, tmdbYear *int) *FileInfo
```

**Input**: Download path + optional TMDB year for validation
**Output**: Structured file metadata

**Parsing Logic**:
- Extract folder name, file name, extension
- Detect year using regex `\b(19|20)\d{2}\b`
- Detect resolution using regex `(?i)\b(2160p|4K|1080p|720p|480p|360p)\b`
- Validate extension against known video formats (mkv, mp4, avi, mov, m4v, wmv, flv, webm)
- Compare detected year with TMDB year (flag mismatch)

**Edge Cases**:
- Null/empty paths → return nil
- Multiple years in path (e.g., "2001.A.Space.Odyssey.1968") → take first match
- No year → `has_year_in_path = false`
- Unknown extension → `is_valid_format = false`

#### 2. API Handler Enhancement

**Location**: `internal/api/handlers_frontend.go`

**New Handler**: `listDownloadsEnriched()`

**Query Strategy** (GORM-based for codebase consistency):
```go
db.Model(&models.DownloadInfo{}).
   Preload("ProcessedLines.Movie").
   Preload("ProcessedLines.TVShow").
   Where("status = ?", status).
   Joins("JOIN processed_lines...").  // For content_type filter
   Order("updated_at desc").
   Limit(limit).Offset(offset).
   Find(&downloads)
```

**Enrichment Logic**:
1. For each DownloadInfo, take first ProcessedLine
2. If ProcessedLine.Movie exists → extract movie metadata
3. If ProcessedLine.TVShow exists → format as "Title S##E##"
4. Parse download_path using fileparser
5. Build DownloadEnrichedResponse

**Filtering**:
- `status`: DB filter (WHERE clause)
- `type`: DB filter (JOIN + WHERE)
- `problem`: Post-processing filter in Go

**Problem Filters**:
- `missing_year`: `!file_info.has_year_in_path`
- `year_mismatch`: `file_info.year_mismatch`
- `unknown_format`: `!file_info.is_valid_format`
- `low_quality`: `detected_resolution in (480p, 360p)`

#### 3. Frontend UI Enhancement

**Location**: `frontend/src/App.tsx`

**Display Priority**:
1. **Title** (large, bold) - from content.title
2. **Folder/File path** (secondary)
3. **Technical specs** (format, resolution, size)
4. **Status** with badges
5. **Problem indicators** (⚠️ for issues)

**Card Layout**:
```
┌────────────────────────────────────────────┐
│ 🎬 La Cité de Dieu (2002)      ✅ Complété │
├────────────────────────────────────────────┤
│ 📂 La.Cite.de.Dieu.2002.1080p.BluRay/     │
│ 📹 MKV • 1080p • 4.2 GB • Année ✓         │
│ 🎭 Crime, Drama                            │
└────────────────────────────────────────────┘
```

**Filter UI**: Dropdowns for status, type, problem (standard HTML select)

### API Response Schema

```typescript
{
  id: number;
  url: string;
  status: string;
  download_path?: string;
  file_size?: number;
  bytes_downloaded?: number;
  total_bytes?: number;
  retry_count: number;
  error_message?: string;
  updated_at: string;
  
  content?: {
    type: 'movies' | 'tvshows' | 'channels' | 'uncategorized';
    title: string;
    year?: number;
    resolution?: string;
    season?: number;
    episode?: number;
    genres?: string;
    duration?: number;
  };
  
  file_info?: {
    extension: string;
    folder_name: string;
    file_name: string;
    has_year_in_path: boolean;
    year_mismatch: boolean;
    detected_year?: number;
    detected_resolution?: string;
    is_valid_format: boolean;
  };
}
```

## Design Decisions

### 1. GORM Preload vs Raw SQL

**Decision**: Use GORM Preload

**Rationale**: 
- Consistent with existing codebase patterns (`listItems`, `listProcessingLogs`)
- Type-safe
- Easier to test
- Acceptable performance for paginated queries

**Trade-off**: Loads all ProcessedLines per DownloadInfo (memory overhead), but mitigated by:
- Pagination limits results
- Most downloads have 1-2 ProcessedLines
- Can optimize later if needed

### 2. First Match Strategy for ProcessedLines

**Decision**: Use first ProcessedLine (by created_at) when multiple exist

**Rationale**:
- Simpler than aggregation
- Most downloads have single ProcessedLine
- Edge case (multiple lines) is acceptable degradation

### 3. Backend vs Frontend Parsing

**Decision**: Parse in backend

**Rationale**:
- Centralized logic (reusable for stats, reports)
- Testable in isolation
- Consistent parsing across all clients

### 4. Problem Filter Post-Processing

**Decision**: Apply "problem" filters in Go after DB query

**Rationale**:
- Parsing logic can't be expressed in SQL easily
- Total count reflects DB state (before filter)
- Acceptable UX: "3 results with issues (150 total)"

**Trade-off**: Pagination may return fewer than `limit` items when problem filter is active

### 5. Year Detection Strategy

**Decision**: Regex `\b(19|20)\d{2}\b` on full path, take first match

**Rationale**:
- Covers 1900-2099 range
- Simple and fast
- Edge cases (like "2001.A.Space.Odyssey.1968") are rare

**Mitigation**: year_mismatch flag alerts user to ambiguous cases

## Testing Strategy

### Unit Tests
- File parser: 100% coverage required
  - Valid paths (movies, TV shows)
  - Missing year, missing resolution
  - Edge cases (multiple years, ambiguous formats)
  - Null/empty inputs

### Integration Tests
- API endpoint: Manual curl testing
  - With/without filters
  - Verify GORM generates expected queries (check logs)
  - Verify pagination works

### Frontend Testing
- Manual browser testing
  - Verify enriched display renders correctly
  - Test filters (status, type, problem)
  - Verify graceful degradation (missing data)

## Performance Considerations

**Expected Load**: <1000 downloads in production

**Query Performance**:
- GORM Preload with pagination = acceptable
- Indexes already exist on relevant columns

**Optimization Path** (if needed later):
- Add materialized view for common queries
- Cache parsed file_info in DB column
- Implement cursor-based pagination

**Current Decision**: Start with simple approach, optimize based on real metrics
