## Context

See proposal.md - Why. Today `M3UConfig` (`internal/config/config.go`) carries both the singular `FilePath`/`Download` fields and the `Sources []M3USourceConfig` list, resolved by `ResolvedSources()` (either/or, not a merge) and `UsesImplicitSource()`. Four call sites go through that resolver correctly (`cmd/m3u.go`'s three subcommands and `cmd/process.go`), but three do not: `cmd/analyze.go` (`dryrun`), `cmd/db_prune.go` (`db-prune`), `cmd/config_cmd.go` (`config` display), plus `internal/api/handlers.go`'s `executeDryRun`. Those four read `cfg.M3U.FilePath` directly and are blind to `m3u.sources`.

`db-prune` is the one non-trivial case: it has no source-selection flag today and queries `processed_lines` keyed only on `line_hash` (not scoped by `source_name`, even though that column and its composite unique index already exist from the multi-source change). The user has confirmed the intended behavior: `db-prune` should loop over every configured source and prune each source's stale lines against that source's own current file, mirroring the failure-isolation loop pattern already used by `m3u-download`/`process`.

## Goals / Non-Goals

**Goals:**
- Single supported M3U configuration shape: `m3u.sources`, required and non-empty.
- Every CLI/API code path that currently reads `cfg.M3U.FilePath` is migrated to the sources list, with no remaining reference to the removed fields.
- `db-prune` becomes source-aware and source-scoped, closing the gap left by the original multi-source change.

**Non-Goals:**
- No database schema change. The `source_name` column and its `(source_name, line_hash)` composite unique index already exist and are unaffected.
- No new environment-variable mechanism for configuring multiple sources. `m3u.sources` remains config-file only, as it is today; the singular block's env bindings (`STALKEER_M3U_FILE_PATH`/`M3U_FILE_PATH`, `m3u.download.*`) are simply removed.
- No change to per-source dedup scoping, download-fallback ordering, or the frontend — those are unaffected by removing the legacy config shape.
- No change to the `process <file>` positional-argument bypass (`cmd/process.go`) or its `"default"` source-name tagging — that is a separate, still-supported one-off manual mode, not the `m3u.file_path` config field.

## Decisions

### Config: delete `FilePath`/`Download` from `M3UConfig`, delete the resolver
Remove `M3UConfig.FilePath` and `M3UConfig.Download` (lines 37/39 today), and delete `ResolvedSources()` / `UsesImplicitSource()` entirely. Every caller that used `cfg.M3U.ResolvedSources()` switches to reading `cfg.M3U.Sources` directly, since it is now guaranteed non-empty by validation (see next decision). `M3USourceConfig` and `M3UDownloadConfig` are unchanged.

**Alternative considered**: keep `ResolvedSources()` as a thin wrapper that just returns `cfg.M3U.Sources` (to minimize call-site churn). Rejected — the whole point is to remove the either/or concept; keeping a same-named function around invites someone to reintroduce fallback logic later, and the call-site churn is small (4 files).

### Validation: reject empty/missing `sources` at config load
Add a check in `internal/config/config.go`'s existing `validate()` function (replacing the current `// m3u.file_path is optional - can be provided via CLI` comment): `if len(cfg.M3U.Sources) == 0 { return fmt.Errorf("m3u.sources must be a non-empty list") }`. This surfaces as a `Load()` error, which every CLI command already handles the same way (`fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err); os.Exit(1)`), so no new error-handling path is needed.

**Alternative considered**: validate lazily at each call site (e.g., `dryrun`, `db-prune`) instead of centrally. Rejected — `m3u-download`/`process` already implicitly require a non-empty list to do anything useful; validating once at load time gives a single, early, consistent error message instead of N slightly different ones.

### `SourcePaths`: drop the `implicit` parameter
`internal/m3udownloader.SourcePaths(filePath, archiveDir, name, implicit bool)` drops `implicit`. Every source (there is no more "the implicit one") always gets its `<name>/` subdirectory inserted, matching the behavior already documented for named sources. Every call site (`cmd/m3u.go`, `cmd/process.go`) drops its paired `implicit := cfg.M3U.UsesImplicitSource()` line.

### `dryrun` / API `executeDryRun`: drop the implicit default, keep the explicit override
Both already accept an explicit file path (CLI positional arg / JSON `file_path`) as an override of the configured default. Remove the `cfg.M3U.FilePath` default assignment; when no explicit path is given, these commands SHALL require one and error clearly ("a file path is required") instead of falling back to config. This is a pure simplification — no source list to choose from with sane semantics, so no attempt is made to guess which configured source's file to default to.

**Alternative considered**: default to the first configured source's file path (`cfg.M3U.Sources[0].FilePath`). Rejected — silently picking "the first source" as a stand-in default is more surprising than requiring an explicit path, especially once there are 2+ sources with no inherent ordering semantics.

### `config` (display command): loop over sources
Replace the single `fmt.Printf("\nM3U File Path: %s\n", cfg.M3U.FilePath)` line with a loop printing each configured source's `name` and `file_path`.

### `db-prune`: loop over every configured source, scope every query by `source_name`
Restructure `cmd/db_prune.go` to loop over `cfg.M3U.Sources`. For each source:
1. Parse that source's `file_path` with `parser.NewParserWithLogger(source.FilePath, source.Name, log)` to collect that source's current `line_hash` set (as today, just scoped to one source instead of "the" file).
2. If the source's file does not exist, or parses to zero entries, log a warning and skip that source (do not abort the whole command) — this mirrors `m3u-download`/`process`'s existing per-source failure isolation, and specifically avoids one broken/missing source's file wiping out a different, healthy source's history.
3. Run the existing soft/hard/dry-run queries with an added `AND source_name = ?` clause (or `WHERE source_name = ? AND line_hash NOT IN (?)`), so a source's stale-line detection only ever considers that source's own current file content, never another source's.
4. Accumulate dry-run counts and post-prune orphan-cleanup effects across all sources for the final printed summary; exit non-zero if at least one source could not be pruned (its file missing/empty), consistent with `m3u-download`'s exit-code convention.

**Alternative considered**: require an explicit `--source <name>` flag and prune one source per invocation. Rejected per the user's stated preference — looping over all sources in one run keeps `db-prune`'s UX consistent with `m3u-download`/`process`, and cron/CI callers don't need to be updated to loop themselves.

**Alternative considered**: prune against the union of all sources' hashes without scoping by `source_name` (i.e., a line is stale only if its hash is absent from *every* source's file). Rejected — this reintroduces exactly the cross-source ambiguity the multi-source change's composite `(source_name, line_hash)` index was designed to eliminate; two sources can legitimately share a `line_hash` for unrelated entries, and unscoped pruning could keep a line alive because a different, unrelated source happens to still contain a matching hash.

### Helm chart: `sources` required, singular block removed
- `values.schema.json`: change the `m3u` object's `required` from `["file_path"]` to `["sources"]`, and add `"minItems": 1` to the `sources` array schema.
- `values.yaml` / `values.test.yaml`: remove the singular `file_path`/`download` keys from `config.m3u`; keep only `sources` (already present, uncomment/adopt as the sole example).
- `templates/configmap.yaml`: remove the unconditional singular-block render (lines 26-46 today); always render the `{{- range .Values.config.m3u.sources }}` block unconditionally instead of behind `{{- if .Values.config.m3u.sources }}` (schema now guarantees it is non-empty).
- `charts/stalkerr/README.md`: replace the singular-only `m3u.file_path`/`download.*` values table with the `sources` shape.

## Risks / Trade-offs

- **[Risk]** Any deployment still relying on `m3u.file_path` (config file or the `STALKEER_M3U_FILE_PATH`/`M3U_FILE_PATH` env vars) fails to start after this change, with no automatic migration. → **Mitigation**: this is an intentional **BREAKING** change (see proposal), documented in `docs/M3U-DOWNLOAD.md` with the exact one-entry-list replacement shape; this repo's own local `config.yml` is already migrated.
- **[Risk]** Existing single-source deployments that migrate their config will see their on-disk download/archive path gain a `<name>/` subdirectory, requiring either a manual file move or a fresh `m3u-download` run. → **Mitigation**: called out explicitly in the spec's REMOVED-requirement migration note and in the docs rewrite; this repo's local config already went through this and left a comment recording it.
- **[Risk]** `db-prune`'s new per-source loop is new, previously-untested code (the old single-file path had test coverage; the multi-source loop does not yet). → **Mitigation**: tasks.md includes new tests mirroring `TestDownloadAllSources_OneFailsOthersStillRun`'s failure-isolation style, adapted for pruning.

## Migration Plan

1. Land the config/CLI/API/downloader changes together (`internal/config`, `cmd/analyze.go`, `cmd/db_prune.go`, `cmd/config_cmd.go`, `internal/api/handlers.go`, `internal/m3udownloader`) — they are mutually dependent (deleting the struct fields breaks every remaining direct reference at compile time, so this cannot be split into independently-shippable steps the way the original additive multi-source change was).
2. Update tests alongside the code they cover, in the same change.
3. Update `config.yml.example` and the Helm chart files.
4. Update `docs/M3U-DOWNLOAD.md`, `docs/DEVELOPMENT.md`, and `charts/stalkerr/README.md` last, once the final shape of every command's behavior is settled.
5. **Rollback**: reverting this change is safe at the code level (no DB migration to undo). Any config file already converted to `sources` (like this repo's) continues to work unchanged if the change is rolled back, since the old binary still supports both shapes.
