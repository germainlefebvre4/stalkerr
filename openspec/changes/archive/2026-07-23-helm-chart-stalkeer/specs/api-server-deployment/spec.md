## ADDED Requirements

### Requirement: Server deployment resource
The chart SHALL create a Kubernetes Deployment for the Stalkeer API server with configurable replica count.

#### Scenario: Deployment created
- **WHEN** server.enabled is true
- **THEN** Deployment resource is created with name derived from chart fullname
- **THEN** Deployment uses configured image repository and tag
- **THEN** Deployment spec includes replica count from values

### Requirement: Health check configuration
The chart SHALL configure liveness and readiness probes for the server container.

#### Scenario: Health probes configured
- **WHEN** Deployment is created
- **THEN** Liveness probe checks /health endpoint on port 8080
- **THEN** Readiness probe checks /health endpoint on port 8080
- **THEN** Probe timing (initialDelaySeconds, periodSeconds, timeoutSeconds) is configurable

### Requirement: Resource limits
The chart SHALL allow configuration of CPU and memory limits/requests for the server deployment.

#### Scenario: Resource configuration
- **WHEN** server.resources is configured in values
- **THEN** Deployment includes resource limits and requests
- **THEN** Default values provide reasonable limits (2 CPU, 2Gi memory)

### Requirement: Environment configuration injection
The chart SHALL inject configuration from ConfigMap and Secret as environment variables.

#### Scenario: Environment variables set
- **WHEN** Deployment is created
- **THEN** ConfigMap values are available as environment variables
- **THEN** Secret values are available as environment variables
- **THEN** Database URL is auto-populated when postgresql.enabled is true

### Requirement: Port exposure
The chart SHALL configure container ports for API (8080) and admin (8081) endpoints.

#### Scenario: Ports configured
- **WHEN** Deployment is created
- **THEN** Container exposes port 8080 (API)
- **THEN** Container exposes port 8081 (admin)
- **THEN** Ports are named for service reference

### Requirement: Service resource
The chart SHALL create a Kubernetes Service to expose the server deployment.

#### Scenario: Service created
- **WHEN** server.enabled is true
- **THEN** Service resource is created with type ClusterIP
- **THEN** Service maps port 80 to container port 8080
- **THEN** Service selector matches deployment labels
