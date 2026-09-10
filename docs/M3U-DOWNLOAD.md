# M3U Playlist Download and Archive Management

This feature enables automatic downloading of M3U playlist files from one or more remote sources with built-in archiving and rotation capabilities.

## Overview

The M3U Download feature provides:

- **Automated Downloads**: Download M3U playlists from HTTP/HTTPS URLs
- **Multi-Source Support**: Configure one or more M3U providers; each is downloaded, archived, and processed independently within the same run
- **Atomic Operations**: Safe, atomic file updates prevent corruption
- **Validation**: Verify M3U format before accepting downloads
- **Archive Management**: Automatic timestamped archiving of playlist versions, per source
- **Rotation**: Keep only the N most recent archives to manage disk space
- **Failure Isolation**: A failing source (download error, missing file) is logged and does not prevent the other configured sources from being attempted
- **Error Handling**: Retry mechanism with circuit breaker for reliability
- **Security**: File size limits, content validation, and HTTPS support

## Configuration

`m3u.sources` is **required** and must be a non-empty list; configuration fails to load if it is absent or empty. Add the following to your `config.yml`:

```yaml
m3u:
  update_interval: 3600
  sources:
    - name: provider-a
      file_path: /path/to/provider-a.m3u
      download:
        enabled: true
        url: "https://provider-a.example.com/playlist.m3u"
        archive_dir: ./m3u_playlist
        retention_count: 5
        max_file_size_mb: 500
        timeout_seconds: 300
        retry_attempts: 3
        # Optional: HTTP Basic Authentication
        auth_username: ""
        auth_password: ""
    - name: provider-b
      file_path: /path/to/provider-b.m3u
      download:
        enabled: true
        url: "https://provider-b.example.com/playlist.m3u"
        archive_dir: ./m3u_playlist
        retention_count: 5
        max_file_size_mb: 500
        timeout_seconds: 300
        retry_attempts: 3
```

A deployment with only one M3U provider still configures it as a one-entry `sources` list.

### Configuration Options

Each entry in `m3u.sources` has:

| Option | Type | Description |
|--------|------|-------------|
| `name` | string | Required, unique identifier for this source |
| `file_path` | string | Where to save this source's downloaded M3U file |

Each entry's `download` block has:

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | boolean | `false` | Enable/disable download for this source |
| `url` | string | `""` | Remote M3U playlist URL (HTTP/HTTPS) |
| `archive_dir` | string | `./m3u_playlist` | Directory to store archived M3U files |
| `retention_count` | integer | `5` | Number of archived files to retain |
| `max_file_size_mb` | integer | `500` | Maximum allowed file size in megabytes |
| `timeout_seconds` | integer | `300` | Download timeout in seconds |
| `retry_attempts` | integer | `3` | Number of retry attempts on failure |
| `auth_username` | string | `""` | HTTP Basic Auth username (optional) |
| `auth_password` | string | `""` | HTTP Basic Auth password (optional) |

Key points:

- **Unique `name` required**: each source needs a unique `name`. It tags every `ProcessedLine` parsed from that source (`source_name` field, also exposed as `source_name` in the `/api/v1/items` API response) and is used to namespace that source's files on disk.
- **Per-source subdirectory, always**: regardless of what `file_path`/`archive_dir` you configure per source, the effective download destination and archive directory automatically get the source's `name` inserted as a subdirectory - e.g. `archive_dir: ./m3u_playlist` with `name: provider-a` becomes `./m3u_playlist/provider-a/`. This guarantees two sources never collide, even if their configured paths look identical.
- **Dedup is scoped per source**: the uniqueness constraint is `(source_name, line_hash)`. Two sources that happen to produce an identical-looking entry (same `tvg_name` + URL) are both kept as separate, retained `ProcessedLine` rows rather than one being silently dropped as a duplicate - see [Database Schema](DATABASE.md#processed_lines) for details. A true duplicate within the *same* source's file is still deduplicated exactly as before.
- **Redundancy, not ranking**: when a movie or episode is available from more than one source, all matching `ProcessedLine` candidates (one per source) feed into the existing quality/language download-fallback ordering unchanged - `source_name` is informational only and never used as a ranking factor. If one source's stream fails to download, the next candidate (possibly from another source) is tried automatically.

### Environment Variables

`m3u.sources` is a structured list and is config-file only - there is no environment variable equivalent for configuring M3U sources. Every other top-level setting (database, logging, TMDB, Radarr/Sonarr, etc.) can still be set via environment variables as usual; only M3U source configuration requires `config.yml`.

## Usage

### Download M3U Playlist

Download the M3U playlist(s) from every configured source:

```bash
stalkeer m3u-download
```

Every configured source is attempted independently in the same run:

1. Download the M3U file from the source's configured URL
2. Validate the M3U format
3. Save atomically to the source's effective `file_path`
4. Create a timestamped archive copy
5. Rotate old archives based on the source's retention settings

A source that fails to download (network error, invalid response) is logged and does not stop the remaining sources from being attempted. The command exits with a non-zero status if at least one source failed, but only after every configured source has been attempted.

**With custom URL:**

```bash
stalkeer m3u-download --url https://example.com/custom-playlist.m3u
```

`--url` only applies when exactly one source is configured. With multiple configured sources it is ambiguous and is silently ignored in favor of each source's own configured URL.

**Without archiving:**

```bash
stalkeer m3u-download --no-archive
```

### List Archived Playlists

View archived M3U files for every configured source:

```bash
stalkeer m3u-list-archives
```

Output example:
```
Archived M3U files for source "provider-a" (./m3u_playlist/provider-a):

Filename                                 Size         Modified
--------------------------------------------------------------------------------
playlist_20260203_230154.054749.m3u      40.96 MB     2026-02-03 23:01:54
playlist_20260203_225902.373353.m3u      40.96 MB     2026-02-03 22:59:02
playlist_20260202_180000.123456.m3u      40.95 MB     2026-02-02 18:00:00
playlist_20260201_120000.000000.m3u      40.94 MB     2026-02-01 12:00:00
playlist_20260131_160000.000000.m3u      40.93 MB     2026-01-31 16:00:00

Total: 5 archived files
```

With multiple configured sources, this lists each source's own archive subdirectory (e.g. `./m3u_playlist/provider-a`, `./m3u_playlist/provider-b`) in turn.

### Clean Up Old Archives

Manually trigger archive rotation for every configured source:

```bash
# Use each source's own configured retention count
stalkeer m3u-cleanup-archives

# Override retention count for every source
stalkeer m3u-cleanup-archives --retention 3
```

## How It Works

### Download Workflow

The following runs independently for each configured source:

1. **Request**: HTTP GET request to the configured URL
2. **Validation**: 
   - Check HTTP status code (200 OK)
   - Verify content type (if provided)
   - Enforce file size limits
3. **Content Validation**: Verify M3U format (`#EXTM3U` header)
4. **Atomic Write**: 
   - Download to temporary file (in the per-source destination directory, created if missing)
   - Validate content
   - Atomic rename to the source's effective `file_path`
5. **Archive**: Create timestamped copy in the source's archive directory
6. **Rotation**: Delete archives beyond that source's retention count

With multiple sources, each source's failure in any of these steps is caught and logged individually - it does not stop the remaining sources' workflow from running.

### Archive Filename Format

Archives use the following naming convention:

```
playlist_YYYYMMDD_HHMMSS.ffffff.m3u
```

Example: `playlist_20260203_230154.054749.m3u`

- `YYYYMMDD`: Date (2026-02-03)
- `HHMMSS`: Time (23:01:54)
- `ffffff`: Microseconds (054749)

### Error Handling

The download system includes multiple layers of error handling:

1. **Retry Mechanism**: Automatically retries transient failures
   - Configurable retry attempts
   - Exponential backoff with jitter
   - Retries network errors and 5xx HTTP errors
   - Skips 4xx errors (client errors)

2. **Circuit Breaker**: Prevents repeated failures
   - Opens after 5 consecutive failures
   - Waits 60 seconds before retry
   - Protects against unreliable remote servers

3. **Validation**: Prevents corruption
   - M3U format validation
   - File size enforcement
   - Atomic file operations

4. **Preservation**: Original file safety
   - Downloads to temporary file first
   - Original file unchanged on failure
   - Only replaced after successful validation

## Security Considerations

### File Size Limits

The `max_file_size_mb` setting prevents disk exhaustion attacks:

```yaml
m3u:
  sources:
    - name: provider-a
      download:
        max_file_size_mb: 500  # Reject files larger than 500MB
```

### HTTPS Support

Use HTTPS URLs for secure downloads:

```yaml
m3u:
  sources:
    - name: provider-a
      download:
        url: "https://example.com/playlist.m3u"  # HTTPS recommended
```

### Authentication

Support for HTTP Basic Authentication:

```yaml
m3u:
  sources:
    - name: provider-a
      download:
        url: "https://example.com/playlist.m3u"
        auth_username: "myuser"
        auth_password: "mypassword"
```

## Troubleshooting

### Download Fails with Timeout

**Problem**: Downloads timeout for large files

**Solution**: Increase timeout setting for the affected source:

```yaml
m3u:
  sources:
    - name: provider-a
      download:
        timeout_seconds: 600  # 10 minutes
```

### Invalid M3U Format Error

**Problem**: Downloaded file fails validation

**Solution**: Check the remote URL returns a valid M3U file:

```bash
curl -I https://example.com/playlist.m3u
# Should return Content-Type: application/vnd.apple.mpegurl or similar
```

### File Size Exceeded Error

**Problem**: File larger than configured limit

**Solution**: Increase the limit or verify URL is correct:

```yaml
m3u:
  sources:
    - name: provider-a
      download:
        max_file_size_mb: 1000  # Increase limit
```

### Authentication Failures

**Problem**: 401 Unauthorized errors

**Solution**: Verify credentials are correct:

```bash
# Test with curl
curl -u username:password https://example.com/playlist.m3u
```

### Archive Directory Permission Error

**Problem**: Cannot create archive directory

**Solution**: Ensure write permissions:

```bash
mkdir -p ./m3u_playlist
chmod 755 ./m3u_playlist
```

### Circuit Breaker Open

**Problem**: "circuit breaker is open" error

**Solution**: Wait 60 seconds or fix the remote URL. The circuit breaker automatically resets after the timeout period.

## Integration with Process Command

After downloading an M3U playlist, process it:

```bash
# Download the latest playlist(s) for every configured source
stalkeer m3u-download

# Process every configured source's downloaded file, each tagged with its own source_name
stalkeer process
```

`process` iterates every configured source and processes each source's file independently, tagging every resulting `ProcessedLine` with that source's `name`. If a source's file is missing (for example, because its `m3u-download` attempt failed), `process` logs a warning, skips it, and still processes the other sources.

A single positional file path (`stalkeer process /path/to/file.m3u`) still works exactly as before: it bypasses the configured source list entirely for a manual one-off run, tagging every resulting line with the `default` source name.

## Best Practices

1. **Regular Downloads**: Set up periodic downloads
   - Running a bare `stalkeer` binary: use a host crontab entry, e.g. `0 2 * * * /usr/local/bin/stalkeer m3u-download`
   - Running via Docker Compose: see [DOCKER-QUICKSTART.md's Next Steps](../DOCKER-QUICKSTART.md#next-steps) for the `cron` profile sidecar (default: daily 23:30) or the equivalent host-crontab recipe
   - Running on Kubernetes: see the Helm chart's [`CronJob` documentation](../charts/stalkerr/README.md)

2. **Monitor Archives**: Check archive count periodically
   ```bash
   stalkeer m3u-list-archives
   ```

3. **Retention Policy**: Adjust based on your needs
   - Keep 5-10 archives for debugging
   - Keep 1-2 for minimal storage

4. **Backup Archive Directory**: Include in backups
   ```bash
   tar -czf m3u-archives-backup.tar.gz ./m3u_playlist
   ```

5. **Use HTTPS**: Always prefer HTTPS URLs for security

6. **Environment Variables**: Store sensitive data (database, API keys) in environment variables where supported

## Future Enhancements

Planned features for future releases:

- **Scheduled Downloads**: Automatic periodic downloads
- **Content Hashing**: Skip redundant downloads when content unchanged
- **Webhook Notifications**: Notify on successful/failed downloads
- **S3 Archive Storage**: Store archives in cloud storage
- **Download Metrics**: Track download history and statistics

## Related Documentation

- [Configuration Management](DEVELOPMENT.md#configuration)
- [Database Schema](DATABASE.md#processed_lines) - `source_name` field and per-source dedup constraint
- [Error Handling](ERROR-HANDLING.md)
- [Logging System](LOGGING.md)

## Support

For issues or questions:

1. Check the [Troubleshooting](#troubleshooting) section
2. Review logs for detailed error messages
3. Open an issue on GitHub
