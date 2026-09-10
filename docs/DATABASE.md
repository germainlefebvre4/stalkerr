# Database Schema Documentation

This document describes the database schema for the Stalkeer application.

## Overview

The schema implements a polymorphic design where M3U playlist lines are stored in `processed_lines` with relationships to content-specific tables (`movies`, `tvshows`, `channels`, `uncategorized`). This approach:

- Preserves original M3U line content
- Enables TMDB integration for normalized metadata
- Supports deduplication via content hashing
- Tracks processing state and version history

## Tables

### processed_lines

Stores original M3U playlist lines with polymorphic relationships to content types.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | INTEGER | PRIMARY KEY | Unique identifier |
| `line_content` | TEXT | NOT NULL | Original M3U EXTINF line |
| `line_url` | TEXT | NULLABLE | Stream URL from M3U |
| `source_name` | VARCHAR(100) | NOT NULL, DEFAULT 'default', part of composite unique | Name of the configured M3U source this line was parsed from (see [M3U Download](M3U-DOWNLOAD.md#configuration-options)) |
| `line_hash` | VARCHAR(64) | NOT NULL, part of composite unique | SHA-256 hash for deduplication, scoped per source |
| `line_number` | INTEGER | NOT NULL, DEFAULT 0 | Line number in the source M3U file |
| `tvg_name` | VARCHAR(255) | NOT NULL | Original TVG name from M3U |
| `group_title` | VARCHAR(255) | NOT NULL | Original group title from M3U |
| `processed_at` | TIMESTAMP | NOT NULL | Processing timestamp |
| `content_type` | VARCHAR(20) | NOT NULL | Content category (movies/tvshows/channels/uncategorized) |
| `resolution` | VARCHAR(10) | NULLABLE | Detected video resolution (e.g. 1080p) |
| `language` | VARCHAR(10) | NULLABLE | Detected audio/subtitle language |
| `french_variant` | VARCHAR(10) | NULLABLE | Detected French audio variant (VF/VFF/VOSTFR, etc.) |
| `channel_id` | INTEGER | FOREIGN KEY, NULLABLE | Reference to channels |
| `movie_id` | INTEGER | FOREIGN KEY, NULLABLE | Reference to movies |
| `tvshow_id` | INTEGER | FOREIGN KEY, NULLABLE | Reference to tvshows |
| `uncategorized_id` | INTEGER | FOREIGN KEY, NULLABLE | Reference to uncategorized |
| `download_info_id` | INTEGER | FOREIGN KEY, NULLABLE | Reference to download tracking (`download_info`) |
| `processing_log_id` | INTEGER | FOREIGN KEY, NULLABLE | Reference to the processing run (`processing_logs`) that created this line |
| `state` | VARCHAR(50) | NOT NULL, DEFAULT 'processed' | Processing state |
| `override_by` | VARCHAR(50) | NULLABLE | Identifier of who/what manually overrode this line's match |
| `override_at` | TIMESTAMP | NULLABLE | Timestamp of the manual override |
| `downloaded_at` | TIMESTAMP | NULLABLE | Timestamp the download completed |
| `created_at` | TIMESTAMP | NOT NULL | Record creation time |
| `updated_at` | TIMESTAMP | NOT NULL | Record update time |

**Indexes:**
- `idx_processed_lines_source_hash` (unique) on `(source_name, line_hash)` - replaces the old single-column unique index on `line_hash`, so identical-looking entries from two different sources are both retained instead of one being rejected as a duplicate
- `idx_processed_lines_content` on `(content_type, state)`
- `idx_processed_lines_m3u` on `(group_title, tvg_name)`
- `idx_processed_lines_download` on `download_info_id`
- `idx_processed_lines_created_at` on `created_at`
- individual indexes on `channel_id`, `movie_id`, `tvshow_id`, `uncategorized_id`, `processing_log_id`

**Content Types:**
- `movies` - Movie content
- `tvshows` - TV show episodes
- `channels` - Live TV channels
- `uncategorized` - Unclassified content

**Processing States:**
- `processed` - Successfully parsed and categorized
- `pending` - Awaiting processing
- `downloading` - Currently being downloaded
- `organizing` - Download complete, being moved/renamed into its final destination
- `downloaded` - Download completed
- `failed` - Processing or download failed
- `cancelled` - Manually cancelled or exhausted its retry budget

---

### movies

Stores movie metadata from TMDB with deduplication by title and year.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | INTEGER | PRIMARY KEY | Unique identifier |
| `tmdb_id` | INTEGER | NOT NULL | TMDB database ID |
| `tvdb_id` | INTEGER | NULLABLE | TVDB database ID (from TMDB external_ids) |
| `tmdb_title` | VARCHAR(255) | NOT NULL | Movie title from TMDB |
| `tmdb_year` | INTEGER | NOT NULL | Release year from TMDB |
| `tmdb_genres` | TEXT | NULLABLE | Genres as JSON array |
| `duration` | INTEGER | NULLABLE | Duration in minutes |
| `poster_path` | VARCHAR(255) | NULLABLE | TMDB poster path |
| `overview` | TEXT | NULLABLE | TMDB synopsis |
| `imdb_id` | VARCHAR(20) | NULLABLE | IMDb ID (from TMDB external_ids) |
| `created_at` | TIMESTAMP | NOT NULL | Record creation time |
| `updated_at` | TIMESTAMP | NOT NULL | Record update time |

**Indexes:**
- `idx_movies_tmdb` on `tmdb_id`
- `idx_movies_tvdb` on `tvdb_id`

**Unique Constraints:**
- `idx_movies_unique` composite unique index on `(tmdb_title, tmdb_year)` - Prevents duplicate movies

**Foreign Keys:**
- `processed_lines.movie_id` → `movies.id` (CASCADE)

---

### tvshows

Stores TV show metadata from TMDB with season/episode information.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | INTEGER | PRIMARY KEY | Unique identifier |
| `tmdb_id` | INTEGER | NOT NULL | TMDB database ID |
| `tvdb_id` | INTEGER | NULLABLE | TVDB database ID (from TMDB external_ids) |
| `tmdb_title` | VARCHAR(255) | NOT NULL | Show title from TMDB |
| `tmdb_year` | INTEGER | NOT NULL | First air date year |
| `tmdb_genres` | TEXT | NULLABLE | Genres as JSON array |
| `season` | INTEGER | NULLABLE | Season number |
| `episode` | INTEGER | NULLABLE | Episode number |
| `poster_path` | VARCHAR(255) | NULLABLE | TMDB poster path |
| `overview` | TEXT | NULLABLE | TMDB synopsis |
| `imdb_id` | VARCHAR(20) | NULLABLE | IMDb ID (from TMDB external_ids) |
| `created_at` | TIMESTAMP | NOT NULL | Record creation time |
| `updated_at` | TIMESTAMP | NOT NULL | Record update time |

**Indexes:**
- `idx_tvshows_tmdb` on `tmdb_id`
- `idx_tvshows_tvdb` on `tvdb_id`
- `idx_tvshows_season_episode` on `(season, episode)`

Note: unlike `movies`, `tvshows` currently has no composite unique constraint on `(tmdb_title, tmdb_year, season, episode)` at the database level; deduplication for TV episodes is handled in application logic.

**Foreign Keys:**
- `processed_lines.tvshow_id` → `tvshows.id` (CASCADE)

---

### channels

Stores live TV channel metadata.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | INTEGER | PRIMARY KEY | Unique identifier |
| `name` | VARCHAR(255) | NOT NULL | Channel name |
| `logo` | TEXT | NULLABLE | Channel logo URL |
| `group_title` | VARCHAR(255) | NOT NULL, INDEXED | Original group title from M3U |
| `created_at` | TIMESTAMP | NOT NULL | Record creation time |
| `updated_at` | TIMESTAMP | NOT NULL | Record update time |

**Foreign Keys:**
- `processed_lines.channel_id` → `channels.id`

---

### uncategorized

Stores content that could not be classified as a movie, TV show, or channel.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | INTEGER | PRIMARY KEY | Unique identifier |
| `title` | VARCHAR(255) | NOT NULL | Raw title extracted from the M3U line |
| `group_title` | VARCHAR(255) | NOT NULL, INDEXED | Original group title from M3U |
| `created_at` | TIMESTAMP | NOT NULL | Record creation time |
| `updated_at` | TIMESTAMP | NOT NULL | Record update time |

**Foreign Keys:**
- `processed_lines.uncategorized_id` → `uncategorized.id`

---

### download_info

Tracks per-line download progress and resumability.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | INTEGER | PRIMARY KEY | Unique identifier |
| `url` | TEXT | INDEXED | Source URL of the download |
| `status` | VARCHAR(50) | NOT NULL, INDEXED | `pending`, `downloading`, `paused`, `completed`, `failed`, `retrying`, `cancelled` |
| `download_path` | TEXT | NULLABLE | Final local path once downloaded |
| `target_path` | TEXT | NULLABLE | Planned final destination for the current/last attempt (cleared on completion) |
| `staging_path` | TEXT | NULLABLE | Temp file location for the current/last attempt (cleared on completion) |
| `file_size` | BIGINT | NULLABLE | Total file size, if known upfront |
| `bytes_downloaded` | BIGINT | DEFAULT 0 | Bytes downloaded so far |
| `total_bytes` | BIGINT | NULLABLE | Expected total file size |
| `resume_token` | VARCHAR(255) | NULLABLE | Server-specific resume identifier (ETag, etc.) |
| `retry_count` | INTEGER | NOT NULL, DEFAULT 0 | Number of retry attempts |
| `last_retry_at` | TIMESTAMP | NULLABLE | Timestamp of the last retry attempt |
| `locked_at` | TIMESTAMP | NULLABLE, INDEXED | Lock timestamp to prevent concurrent downloads |
| `locked_by` | VARCHAR(100) | NULLABLE | Instance/process that acquired the lock |
| `started_at` | TIMESTAMP | NULLABLE | When the download started |
| `completed_at` | TIMESTAMP | NULLABLE | When the download completed |
| `error_message` | TEXT | NULLABLE | Last error message, if any |
| `created_at` | TIMESTAMP | NOT NULL | Record creation time |
| `updated_at` | TIMESTAMP | NOT NULL, INDEXED | Record update time |

**Foreign Keys:**
- `processed_lines.download_info_id` → `download_info.id`

See [DOWNLOAD-TRACKING.md](DOWNLOAD-TRACKING.md) and [RESUME-DOWNLOADS.md](RESUME-DOWNLOADS.md) for behavior details.

---

### processing_logs

Records metrics for each processing run (see [`internal/models/log.go`](../internal/models/log.go)).

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | INTEGER | PRIMARY KEY | Unique identifier |
| `action` | VARCHAR(100) | NOT NULL | Action performed (e.g. `process`, `download`) |
| `item_count` | INTEGER | NOT NULL, DEFAULT 0 | Number of items handled |
| `status` | VARCHAR(50) | NOT NULL | `success`, `failed`, or `in_progress` |
| `started_at` | TIMESTAMP | NOT NULL | Run start time |
| `completed_at` | TIMESTAMP | NULLABLE | Run completion time |
| `error_message` | TEXT | NULLABLE | Error message, if any |
| `movies_count` | INTEGER | NULLABLE | Movies processed |
| `tv_shows_count` | INTEGER | NULLABLE | TV shows processed |
| `new_items_count` | INTEGER | NULLABLE | New items created |
| `tmdb_matched_count` | INTEGER | NULLABLE | Items matched via TMDB |
| `tmdb_unmatched_count` | INTEGER | NULLABLE | Items not matched via TMDB |
| `metadata_backfilled_count` | INTEGER | NULLABLE | Items updated by the run's rich-metadata backfill step (poster/overview/external IDs on legacy records) |
| `metadata_backfill_errors_count` | INTEGER | NULLABLE | Items on which the rich-metadata backfill step errored |
| `group_titles` | JSON | NOT NULL (nullable value) | List of group titles involved in the run |
| `created_at` | TIMESTAMP | NOT NULL | Record creation time |
| `updated_at` | TIMESTAMP | NOT NULL | Record update time |

**Foreign Keys:**
- `processed_lines.processing_log_id` → `processing_logs.id`

Note: `metadata_backfilled_count`/`metadata_backfill_errors_count` are `NULL` on rows created before this column existed, not `0` - a `NULL` means "not tracked", `0` means "tracked and nothing changed".

---

### job_runs

Records a durable run history entry for the standalone `resume-downloads` and `enrich-tvdb` CLI commands, whose outcomes were previously only logged and lost on process exit (see [`internal/models/job_run.go`](../internal/models/job_run.go)). Written unconditionally, independent of whether Prometheus metrics exposition is enabled.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | INTEGER | PRIMARY KEY | Unique identifier |
| `action` | VARCHAR(100) | NOT NULL | `resume-downloads` or `enrich-tvdb` |
| `status` | VARCHAR(50) | NOT NULL | `success`, `failed`, or `in_progress` |
| `started_at` | TIMESTAMP | NOT NULL | Run start time |
| `completed_at` | TIMESTAMP | NULLABLE | Run completion time |
| `succeeded_count` | INTEGER | NOT NULL, DEFAULT 0 | Items the invocation succeeded on |
| `failed_count` | INTEGER | NOT NULL, DEFAULT 0 | Items the invocation failed on |
| `skipped_count` | INTEGER | NOT NULL, DEFAULT 0 | Items the invocation skipped |
| `error_message` | TEXT | NULLABLE | Error message, if any |
| `created_at` | TIMESTAMP | NOT NULL | Record creation time |
| `updated_at` | TIMESTAMP | NOT NULL | Record update time |

A row is created with `status=in_progress` when the invocation starts and finalized (`success`/`failed`, with counts accumulated up to that point) when it completes or crashes.

---

### filter_configs

Runtime filters manageable via the `/api/v1/filters` endpoints, in addition to the file-based filters in `config.yml`.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | INTEGER | PRIMARY KEY | Unique identifier |
| `name` | VARCHAR(255) | NOT NULL, UNIQUE | Filter name |
| `attribute` | VARCHAR(50) | NOT NULL | `group_title` or `tvg_name` |
| `include_patterns` | TEXT | NULLABLE | JSON array of include regex patterns |
| `exclude_patterns` | TEXT | NULLABLE | JSON array of exclude regex patterns |
| `is_runtime` | BOOLEAN | NOT NULL, DEFAULT true, INDEXED | Whether the filter is runtime-only (vs. seeded from config) |
| `created_at` | TIMESTAMP | NOT NULL | Record creation time |
| `updated_at` | TIMESTAMP | NOT NULL | Record update time |

---

### manual_mappings

Persistent manual overrides mapping a malformed raw M3U title to a specific TMDB entry.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | INTEGER | PRIMARY KEY | Unique identifier |
| `tvg_name` | VARCHAR(255) | NOT NULL, part of composite unique | Raw TVG name from M3U |
| `group_title` | VARCHAR(255) | NOT NULL, part of composite unique | Raw group title from M3U |
| `content_type` | VARCHAR(20) | NOT NULL | `movies` or `tvshows` |
| `tmdb_id` | INTEGER | NOT NULL | Target TMDB ID |
| `season` | INTEGER | NULLABLE | Season number, for TV mappings |
| `episode` | INTEGER | NULLABLE | Episode number, for TV mappings |
| `created_at` | TIMESTAMP | NOT NULL | Record creation time |
| `updated_at` | TIMESTAMP | NOT NULL | Record update time |

**Unique Constraints:**
- `idx_manual_mappings_unique` composite unique index on `(tvg_name, group_title)`

---

## Entity Relationships

```
processed_lines
    ├── channel_id → channels.id
    ├── movie_id → movies.id
    ├── tvshow_id → tvshows.id
    ├── uncategorized_id → uncategorized.id
    ├── download_info_id → download_info.id
    └── processing_log_id → processing_logs.id

movies
    └── processed_lines[] (one-to-many)

tvshows
    └── processed_lines[] (one-to-many)

channels
    └── processed_lines[] (one-to-many)

uncategorized
    └── processed_lines[] (one-to-many)

download_info
    └── processed_lines[] (one-to-many)
```

## Design Principles

### Polymorphic Relationships

The `processed_lines` table uses nullable foreign keys to establish relationships with different content types. Only one foreign key should be populated per record based on `content_type`.

### Data Separation

- **M3U-specific data**: Stored in `processed_lines` (original line content, URLs, parsing metadata)
- **Normalized metadata**: Stored in content-specific tables (`movies`, `tvshows`) from TMDB
- API responses include both via `raw` attribute containing the `processed_lines` data

### Deduplication Strategy

1. **M3U lines**: SHA-256 hash (`line_hash`) prevents duplicate playlist entries, scoped per `source_name` - a true duplicate within the same source's file is rejected, but two different sources producing an identical-looking entry are both kept as separate rows (redundancy across M3U providers is intentional, see [M3U Download](M3U-DOWNLOAD.md#configuration-options))
2. **Movies**: Unique constraint on `(tmdb_title, tmdb_year)` prevents duplicate TMDB entries
3. **TV Shows**: Deduplication of `(tmdb_title, tmdb_year, season, episode)` is enforced in application logic (no DB-level composite unique constraint)
4. **Manual mappings**: Unique constraint on `(tvg_name, group_title)` prevents duplicate overrides

### Manual Overrides

The `override_by` and `override_at` fields in `processed_lines` track when a line's match was manually overridden (via `POST /api/v1/items/:id/override`), maintaining an audit trail.

## Migration Strategy

GORM AutoMigrate handles schema creation and updates. For production:

1. Use migration tools (e.g., golang-migrate)
2. Version control all schema changes
3. Test migrations in staging environment
4. Backup data before migration

## Query Optimization

### Recommended Queries

**Find all movies from a specific group:**
```sql
SELECT m.* FROM movies m
JOIN processed_lines pl ON pl.movie_id = m.id
WHERE pl.group_title = 'VOD - Movies'
AND pl.state = 'processed';
```

**Find TV show episodes by season:**
```sql
SELECT t.* FROM tvshows t
JOIN processed_lines pl ON pl.tvshow_id = t.id
WHERE t.tmdb_title = 'Breaking Bad'
AND t.season = 1
ORDER BY t.episode;
```

**Check for duplicate lines within the same source:**
```sql
SELECT source_name, line_hash, COUNT(*) FROM processed_lines
GROUP BY source_name, line_hash
HAVING COUNT(*) > 1;
```

**Find redundant candidates for the same content across sources (by hash, ignoring source):**
```sql
SELECT line_hash, COUNT(DISTINCT source_name) AS source_count, array_agg(DISTINCT source_name) AS sources
FROM processed_lines
GROUP BY line_hash
HAVING COUNT(DISTINCT source_name) > 1;
```

## Best Practices

1. **Always use transactions** for operations affecting multiple tables
2. **Populate content_type** before setting foreign keys
3. **Update line_hash** when line_content changes
4. **Check for existing TMDB entries** before creating new movies/tvshows
5. **Use prepared statements** to prevent SQL injection
6. **Index foreign keys** for optimal join performance

## Future Enhancements

Potential schema additions:

- Full-text search indexes on titles
- Materialized views for statistics
- Database-level composite unique constraint on `tvshows (tmdb_title, tmdb_year, season, episode)`
