## Why

Real-world IPTV usage often involves subscribing to more than one provider at once (complementary catalogs, or redundancy against a single provider going down). Today `stalkeer` can only download, archive, and process one M3U playlist (single `m3u.file_path` / `m3u.download.url`), so a second subscription can only be ingested by manually overwriting the configured file or by running commands with ad-hoc file paths — with no tracking of which subscription a given stream came from, and a global unique-hash constraint that risks silently dropping a line if two providers ever produce a similar `tvg_name`+`url` pair.

## What Changes

- Add an optional `m3u.sources: []` list to configuration, each entry carrying its own `name`, `file_path`, and `download` block (same shape as today's singular `m3u.download`). The existing singular `m3u.file_path` / `m3u.download.*` keys remain fully supported as a zero-migration fallback used when `sources` is empty or absent.
- `m3u-download` and `process` CLI commands loop over every configured source (or the single implicit legacy source) in one run. A failure on one source (download error, missing file) is logged and does **not** abort the other sources — no change to the number or schedule of CronJobs/Compose jobs, the looping happens inside the existing single job.
- Each source downloads/archives to its own subdirectory under the existing M3U storage path, so archives from different providers never collide or overwrite each other.
- `ProcessedLine` gains a flat `source_name` string column (default `"default"` for pre-existing rows via migration). The dedup uniqueness constraint moves from a global `line_hash` unique index to a composite `(source_name, line_hash)` unique index, so identical-looking lines from two different providers are both kept rather than one silently rejected.
- Movies/TV shows already converge across sources for free (existing `tmdb_id`-based `FirstOrCreate` matching), and the existing quality/language candidate ordering and download-fallback loop (`m3u-quality-selection`) already treats every `ProcessedLine` for a piece of content as an interchangeable candidate — so redundant streams from a second provider are automatically usable as fallback candidates with no changes to that selection logic. `source_name` is informational only; it is not a ranking dimension.
- The Helm chart's ConfigMap gains the ability to render a `m3u.sources` list (in addition to the existing singular block) when the chart user configures more than one source.
- Out of scope for this change: TV channel ingestion (channels are not yet grouped into deduplicated entities in the current codebase — a pre-existing gap, unrelated to this change), any new frontend page or UI to manage sources, and any per-source priority/quality weighting. Source provenance is exposed via the API/stats and logs only, not in the existing playlist item sidepanel/drawer.

## Capabilities

### New Capabilities
- `m3u-multi-source`: configuring, downloading, archiving, and processing more than one M3U playlist source in a single run, with per-source failure isolation and per-source dedup scoping.

### Modified Capabilities
- `configuration-management`: the Helm chart's ConfigMap SHALL also render an `m3u.sources` list when the chart user configures multiple M3U sources, in addition to the existing singular `m3u` block.

## Impact

- `internal/config`: `M3UConfig` gains a `Sources []M3USourceConfig` field; loader resolves the effective source list (explicit sources, or one implicit source from the legacy singular fields).
- `internal/parser`: hash/dedup scoping changes from global to per-source; `ProcessedLine` gains `source_name`.
- `internal/processor`, `internal/m3udownloader`: source-aware iteration, per-source archive subdirectories, per-source error isolation.
- `internal/models`, database migration: new `source_name` column on `processed_lines`, replacement of the unique index on `line_hash`.
- `cmd/m3u.go`, `cmd/process.go`: loop over configured sources instead of a single file/URL.
- `internal/api/handlers.go`: stats/dry-run responses can report source where relevant (no new endpoints required).
- `charts/stalkerr/templates/configmap.yaml`, `charts/stalkerr/values.yaml`: render `m3u.sources` list when configured; no new CronJob templates.
- No frontend changes.
