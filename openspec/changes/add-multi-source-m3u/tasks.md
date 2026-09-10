## 1. Configuration

- [ ] 1.1 Add `M3USourceConfig` (`Name`, `FilePath`, `Download M3UDownloadConfig`) and a `Sources []M3USourceConfig` field on `M3UConfig` in `internal/config/config.go`, with viper bindings/defaults for the list; verify `config_test.go` covers parsing a `m3u.sources` list from YAML.
- [ ] 1.2 Add a resolver on `M3UConfig` that returns the configured `Sources`, or synthesizes one implicit source named `"default"` from the legacy `FilePath`/`Download` fields when `Sources` is empty; verify a unit test for both branches (legacy-only config, and `sources`-configured config where the singular fields are ignored).

## 2. Database schema

- [ ] 2.1 Add `SourceName string` to `models.ProcessedLine` (`gorm:"type:varchar(100);not null;default:'default'"`) and change the `LineHash` unique index to a named composite `uniqueIndex:idx_processed_lines_source_hash` covering `(source_name, line_hash)`; verify `models_test.go` reflects the new field.
- [ ] 2.2 Inspect a running dev database (`\d processed_lines`) to confirm the exact name of the existing unique index on `line_hash`, then add idempotent raw SQL to `internal/database/database.go#runMigrations` (following the existing `DROP COLUMN IF EXISTS` precedent) to `DROP INDEX IF EXISTS` that index and `CREATE UNIQUE INDEX IF NOT EXISTS idx_processed_lines_source_hash ON processed_lines (source_name, line_hash)`; verify by running `migrate` against a database seeded with pre-migration data and confirming the new composite index exists and old rows have `source_name = 'default'`.

## 3. Parser

- [ ] 3.1 Thread a source name into `parser.NewParser`/`NewParserWithLogger` and set it on every returned `models.ProcessedLine.SourceName`; verify with a parser unit test asserting `SourceName` on parsed entries.
- [ ] 3.2 Scope the in-memory `seenHashes` within-run dedup key to `source_name + line_hash` instead of `line_hash` alone; verify a parser test where two different source-name parser instances each successfully parse an entry that would hash identically, while a true duplicate within one parser's run is still skipped.

## 4. Downloader / archiver

- [ ] 4.1 Update `internal/m3udownloader` so a source's download/archive writes under a per-source subdirectory (derived from the source name) rather than a bare configured path, leaving the legacy implicit source's path unchanged (no subdirectory inserted); verify `downloader_test.go`/`archive_test.go` cover the new subdirectory behavior alongside the unchanged legacy-path case.

## 5. CLI: `m3u-download`

- [ ] 5.1 Update `cmd/m3u.go`'s `downloadM3UCmd` to resolve the configured source list and loop over it, downloading/archiving each source independently; catch and log a per-source error without stopping the loop, and exit non-zero only after every source has been attempted if any failed; verify with a test exercising one failing and one succeeding source in the same run.
- [ ] 5.2 Update `listM3UArchivesCmd`/`cleanupM3UArchivesCmd` to operate per-source (iterating the resolved source list and their per-source archive subdirectories) so archive listing/cleanup covers every configured source, not just one path; verify with a test covering multiple sources' archive directories.

## 6. CLI: `process`

- [ ] 6.1 Update `cmd/process.go` to resolve the configured source list and loop over it, processing each source's downloaded file with a parser instance tagged with that source's name; when a source's file is missing, log a warning and skip it without stopping the loop; verify with a test covering one missing source file alongside one present source file in the same run.
- [ ] 6.2 Confirm the existing single positional-argument override (`process <m3u-file>`) still works unchanged for a manual one-off run against an explicit file path, bypassing the configured source list entirely; verify with the existing/updated CLI test for that argument.

## 7. API

- [ ] 7.1 Add `SourceName` to `internal/api/dto.go`'s `ItemResponse` and populate it in `internal/api/handlers.go` alongside the existing `LineHash` mapping; verify an API handler test asserts `source_name` is present in the JSON response.

## 8. Helm chart

- [ ] 8.1 Add a `{{- if .Values.config.m3u.sources }}` `range` block to `charts/stalkerr/templates/configmap.yaml` rendering each source's `file_path`/`download.*` into `config.yml`'s `m3u.sources` list, alongside the existing singular `m3u` block; verify with `helm template` against a values file setting two sources, and against the chart's existing default values (singular-only) to confirm no rendering change there.
- [ ] 8.2 Document the new `config.m3u.sources` shape in `charts/stalkerr/values.yaml` (as a commented-out example, default empty list) and `charts/stalkerr/values.schema.json`; verify `helm lint charts/stalkerr` passes.

## 9. Verification

- [ ] 9.1 Run `go test ./...` and `golangci-lint run` and confirm both pass with the new/updated tests from tasks 1-7.
- [ ] 9.2 Manually configure two sources against sample files under `m3u_playlist/`, run `m3u-download` then `process`, and confirm both sources' entries are ingested with the correct `source_name`, that a deliberately-broken URL for one source does not prevent the other source from downloading/processing, and that an identical entry duplicated across both sample files is kept as two distinct `ProcessedLine` rows.
