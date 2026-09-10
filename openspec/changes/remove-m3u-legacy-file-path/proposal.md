## Why

`m3u.sources` (added same-day, currently `HEAD`) already fully supersedes the singular `m3u.file_path` / `m3u.download.*` block for every real deployment — this repo's own local `config.yml` has already been hand-migrated to a one-entry `sources` list. Keeping both shapes forces an either/or resolver (`ResolvedSources()` / `UsesImplicitSource()`) that three CLI/API code paths (`dryrun`, `db-prune`, `config` display, and the API's dry-run endpoint) never actually adopted — they still read `cfg.M3U.FilePath` directly and silently ignore `m3u.sources` if configured. It also forces the Helm chart to keep `file_path` as a required field even for sources-only users. Collapsing to a single supported shape removes that inconsistency and the associated fallback logic.

## What Changes

- **BREAKING**: Remove the singular `m3u.file_path` / `m3u.download.*` configuration fields and the `ResolvedSources()` / `UsesImplicitSource()` fallback resolver. `m3u.sources` becomes the only supported way to configure M3U ingestion and must be a non-empty list.
- **BREAKING**: Remove the `STALKEER_M3U_FILE_PATH` / `M3U_FILE_PATH` and top-level `m3u.download.*` environment variable bindings. `m3u.sources` remains config-file only, as it is today.
- **BREAKING**: Remove the "legacy implicit source keeps its exact configured path" special case. Every configured source — including a lone one — always gets a `<name>/` subdirectory inserted into its download destination and archive directory.
- Wire `dryrun` (CLI) and the API's dry-run endpoint to drop their implicit default file path; both already accept an explicit file path (CLI positional arg / JSON `file_path`) and will now require one.
- Wire `config` (CLI display command) to list every configured source's name and file path instead of a single `M3U File Path` line.
- Wire `db-prune` to loop over every configured source and prune each source's stale `processed_lines` against that source's own current file content, instead of reading a single implicit "active" file.
- Update the example config (`config.yml.example`) and the Helm chart's example values (`values.yaml`, `values.test.yaml`, `values.schema.json`, `templates/configmap.yaml`) to show/require `sources` only.
- Update documentation (`docs/M3U-DOWNLOAD.md`, `docs/DEVELOPMENT.md`, `charts/stalkerr/README.md`) to describe `sources` as the only configuration shape.
- Remove or rewrite tests that exercise the legacy fallback path (config resolver, downloader path resolution, parser's implicit source tag) that no longer apply.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `m3u-multi-source`: remove the "legacy fallback" requirement and its scenarios (config resolution no longer has an implicit-source branch; `m3u.sources` is required and non-empty). Simplify the per-source subdirectory and processing requirements to drop legacy-branch wording, since every source is now handled uniformly.
- `configuration-management`: the "ConfigMap renders multiple M3U sources" requirement changes from an additive, opt-in list (rendered alongside the singular block) to the only supported shape; the ConfigMap scenario listing "m3u settings (file path, download config)" updates to describe the `sources` list.
- `db-prune`: the pruning requirement changes from operating against "the currently active M3U playlist file" (one implicit file) to evaluating every configured source independently against its own current file.

## Impact

- **Code**: `internal/config/config.go` (struct fields, resolver, env bindings), `cmd/analyze.go`, `cmd/db_prune.go`, `cmd/config_cmd.go`, `internal/api/handlers.go`, `internal/m3udownloader` (drop the `implicit` parameter from `SourcePaths`).
- **Tests**: `internal/config/config_test.go`, `cmd/m3u_test.go`, `internal/m3udownloader/downloader_test.go`, `internal/parser/parser_test.go` (legacy-path cases removed/rewritten).
- **Config/deployment**: `config.yml.example`; `charts/stalkerr/{values.yaml,values.test.yaml,values.schema.json,templates/configmap.yaml,README.md}`.
- **Docs**: `docs/M3U-DOWNLOAD.md` (substantial rewrite), `docs/DEVELOPMENT.md`.
- **Database**: no schema change — the `source_name` column and its composite unique index already exist from the multi-source change and are unaffected.
- **Local deployment**: this repo's `config.yml` (git-ignored, not tracked) already uses `sources` and requires no changes.
