# build-version-metadata Specification

## Purpose

Embed the release version, commit, and build date into the running binary at build time and expose them via the CLI, so any running instance can be identified without cross-referencing deploy logs.

## Requirements

### Requirement: Version metadata embedded at build time
The system SHALL embed the semantic version, commit SHA, and build date into the binary at build time via linker flags, defaulting to placeholder values (e.g. `dev`/`unknown`) when built without them.

#### Scenario: Release build embeds real metadata
- **WHEN** the binary is built with version/commit/date linker flags set (as in the release Docker image build)
- **THEN** the binary SHALL report that exact version, commit, and date

#### Scenario: Local build without flags falls back to defaults
- **WHEN** the binary is built without those linker flags (e.g. a plain local `go build`)
- **THEN** the binary SHALL report a default placeholder value instead of failing or reporting stale data

### Requirement: `version` CLI command reports build metadata
The `stalkeer version` command SHALL print the embedded version, commit, and build date.

#### Scenario: Running the version command
- **WHEN** a user runs `stalkeer version`
- **THEN** the command SHALL print the binary's embedded version, commit SHA, and build date
