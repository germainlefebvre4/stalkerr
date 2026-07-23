# ingress-configuration Specification

## Purpose
TBD - created by archiving change helm-chart-stalkeer. Update Purpose after archive.
## Requirements
### Requirement: Standard Ingress resource
The chart SHALL create a standard Kubernetes Ingress resource for exposing the API server.

#### Scenario: Ingress created
- **WHEN** ingress.enabled is true
- **THEN** Ingress resource is created with networking.k8s.io/v1 API
- **THEN** Ingress class name is configurable
- **THEN** Ingress routes traffic to the server service

### Requirement: Host and path configuration
The chart SHALL allow configuration of ingress host and path routing.

#### Scenario: Host configuration
- **WHEN** ingress.hosts is configured
- **THEN** Ingress rule matches specified host (e.g., stalkeer.example.com)
- **THEN** Multiple hosts can be configured
- **THEN** Each host can have multiple paths

#### Scenario: Path routing
- **WHEN** ingress.hosts[].paths is configured
- **THEN** Path type is configurable (Prefix, Exact, ImplementationSpecific)
- **THEN** Path routes to specified service port
- **THEN** Default path is "/" routing to port 8080

### Requirement: TLS configuration
The chart SHALL support TLS termination via Ingress with certificate management.

#### Scenario: TLS enabled
- **WHEN** ingress.tls is configured
- **THEN** Ingress spec includes tls section
- **THEN** TLS secretName references certificate secret
- **THEN** Hosts list matches TLS certificate SANs

#### Scenario: cert-manager integration
- **WHEN** ingress.annotations includes cert-manager.io/cluster-issuer
- **THEN** cert-manager automatically provisions certificates
- **THEN** TLS secret is created and referenced in ingress

### Requirement: Ingress annotations
The chart SHALL allow custom annotations for ingress controller behavior.

#### Scenario: Annotations configured
- **WHEN** ingress.annotations is set
- **THEN** Annotations are applied to Ingress resource
- **THEN** Common annotations (rate limiting, CORS, etc.) are supported

### Requirement: Ingress class selection
The chart SHALL allow specification of ingress controller via ingressClassName.

#### Scenario: Traefik ingress class
- **WHEN** ingress.className is "traefik"
- **THEN** Ingress is processed by Traefik controller
- **THEN** Traefik-specific features are available

#### Scenario: nginx ingress class
- **WHEN** ingress.className is "nginx"
- **THEN** Ingress is processed by nginx controller
- **THEN** nginx-specific annotations can be used

### Requirement: Ingress disabled option
The chart SHALL allow disabling ingress creation for alternative exposure methods.

#### Scenario: No ingress
- **WHEN** ingress.enabled is false
- **THEN** No Ingress resource is created
- **THEN** User can manually create IngressRoute or other exposure method
- **THEN** Service remains accessible within cluster

