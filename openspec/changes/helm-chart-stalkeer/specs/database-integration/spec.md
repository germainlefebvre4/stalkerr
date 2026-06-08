## ADDED Requirements

### Requirement: PostgreSQL subchart dependency
The chart SHALL declare Bitnami PostgreSQL as an optional subchart dependency.

#### Scenario: PostgreSQL dependency declared
- **WHEN** Chart.yaml is parsed
- **THEN** PostgreSQL dependency is listed with version, repository, and condition
- **THEN** Dependency condition is postgresql.enabled
- **THEN** Repository is oci://registry-1.docker.io/bitnamicharts

### Requirement: PostgreSQL conditional deployment
The chart SHALL deploy PostgreSQL only when explicitly enabled via values.

#### Scenario: PostgreSQL enabled
- **WHEN** postgresql.enabled is true
- **THEN** PostgreSQL pods are deployed via subchart
- **THEN** PostgreSQL service is created
- **THEN** Database credentials are auto-generated if not provided

#### Scenario: PostgreSQL disabled
- **WHEN** postgresql.enabled is false
- **THEN** No PostgreSQL resources are created
- **THEN** User must provide external database connection details

### Requirement: Automatic connection URL generation
The chart SHALL automatically generate database connection URL when using internal PostgreSQL.

#### Scenario: Internal database URL
- **WHEN** postgresql.enabled is true
- **THEN** DATABASE_URL environment variable is auto-populated
- **THEN** URL format is postgres://<username>:<password>@<service>:<port>/<database>
- **THEN** Service name references PostgreSQL subchart service

### Requirement: External database configuration
The chart SHALL allow configuration of external database connection when PostgreSQL is disabled.

#### Scenario: External database
- **WHEN** postgresql.enabled is false
- **THEN** config.database.host is required
- **THEN** config.database.port, user, name, sslmode are configurable
- **THEN** Database password comes from secrets section

### Requirement: Database credentials management
The chart SHALL manage database credentials via Secret resource.

#### Scenario: Credentials in Secret
- **WHEN** Secret is created
- **THEN** Database password is stored as data.database-password
- **THEN** Password is base64 encoded
- **THEN** existingSecret option allows referencing external secret

### Requirement: Database persistence
The chart SHALL configure persistent storage for PostgreSQL data when enabled.

#### Scenario: PostgreSQL persistence
- **WHEN** postgresql.enabled is true
- **THEN** postgresql.primary.persistence.enabled defaults to true
- **THEN** PVC size is configurable (default 10Gi)
- **THEN** Storage class is configurable
