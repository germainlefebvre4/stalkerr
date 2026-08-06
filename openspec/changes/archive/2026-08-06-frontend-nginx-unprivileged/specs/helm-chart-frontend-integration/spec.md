## MODIFIED Requirements

### Requirement: Independent Frontend Service in Helm Chart
The Helm Chart SHALL provide distinct deployment and service templates to compile, deploy, and scale the frontend application independently from the Go API backend. The frontend container SHALL start successfully under the chart's enforced non-root Pod/container `securityContext` (`runAsNonRoot: true`, `runAsUser` set to a non-zero UID), without requiring elevated privileges.

#### Scenario: Enable and deploy frontend templates
- **WHEN** a user installs the Helm chart with `frontend.enabled=true`
- **THEN** Kubernetes SHALL deploy a separate replica set for the frontend (using Nginx serving the compiled SPA) and expose it via a dedicated `ClusterIP` Service.

#### Scenario: Frontend container starts under enforced non-root security context
- **WHEN** the frontend pod is scheduled with the chart's default `securityContext` (`runAsNonRoot: true`, `runAsUser: 1000`, no matching `runAsGroup`)
- **THEN** the Nginx process SHALL start and reach a `Ready` state without permission errors on its cache, temp, or PID directories, and SHALL NOT enter `CrashLoopBackOff`.

#### Scenario: Frontend Service exposes port 80 while the container listens on 8080
- **WHEN** a client or the Ingress controller sends traffic to the frontend Service
- **THEN** the Service SHALL continue to accept traffic on port `80` and forward it to the container's internal port `8080`, so external consumers observe no change in the exposed port.
