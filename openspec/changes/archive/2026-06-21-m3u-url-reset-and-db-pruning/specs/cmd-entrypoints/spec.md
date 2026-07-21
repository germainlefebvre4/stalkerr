## ADDED Requirements

### Requirement: Database Pruning CLI Interface
The CLI SHALL provide a `db-prune` command under `stalkeer db-prune` to manually invoke database cleaning.

#### Scenario: db-prune CLI flags registered
- **WHEN** running `stalkeer db-prune --help`
- **THEN** the output displays `--dry-run` and `--hard` flags with descriptions.

### Requirement: Media Reset CLI Interface
The CLI SHALL provide a `reset` command with nested subcommands `movie` and `tvshow` to surgically reset elements.

#### Scenario: reset movie CLI command
- **WHEN** running `stalkeer reset movie --id 42`
- **THEN** the system executes a surgical reset for movie ID 42 and prints a success message.

#### Scenario: reset tvshow CLI command
- **WHEN** running `stalkeer reset tvshow --id 12`
- **THEN** the system executes a surgical reset for TV Show ID 12 and prints a success message.
