## Why

The `stalkerr-frontend` pod is stuck in `CrashLoopBackOff`. The Helm chart forces every container (backend and frontend alike) to run under `securityContext.runAsNonRoot: true` / `runAsUser: 1000` via a shared, global `.Values.securityContext` block. The frontend image is a stock `nginx:alpine`, which is built to start as root and drop privileges internally — it has never been adapted to start directly as a non-root UID. Forced into UID 1000, the nginx master process cannot `mkdir` its own baked-in cache directory (`/var/cache/nginx/client_temp`: `Permission denied`) and exits immediately.

## What Changes

- Replace the frontend base image `nginx:alpine` with `nginxinc/nginx-unprivileged:alpine`, which is purpose-built to run as an arbitrary non-root UID (its writable directories are group-owned by GID 0 with `g+w`, and PID/temp files are relocated under `/tmp`).
- **BREAKING**: the frontend container's internal listen port changes from `80` to `8080` (binding port 80 requires root privileges, which the unprivileged image intentionally does not have). This only affects the container's internal port — the Kubernetes Service and Ingress continue to expose port `80` externally; only `targetPort`/local port-forwarding need updating.
- Update `frontend/nginx.conf` to listen on `8080`.
- Update the Helm chart's frontend deployment/service templates (`containerPort`, `targetPort`) to reflect the new internal port.
- Update `docker-compose.yml` frontend port mapping for local-dev parity.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `helm-chart-frontend-integration`: the frontend deployment must run its Nginx container successfully under the chart's enforced non-root `securityContext` (no crash on startup), and its container port moves from 80 to 8080 internally while the Service's external port stays 80.

## Impact

- `frontend/Dockerfile` — base image change, `EXPOSE` port change.
- `frontend/nginx.conf` — `listen` directive change.
- `charts/stalkerr/templates/deployment-frontend.yaml` — `containerPort` change.
- `charts/stalkerr/templates/service-frontend.yaml` — `targetPort` change.
- `docker-compose.yml` — frontend port mapping change (local dev only).
- No change to `values.yaml` / `values.test.yaml` (`frontend.service.port` stays 80 — it's the external Service port, unaffected).
- No change to `ingress.yaml` (routes to the Service port, unaffected).
