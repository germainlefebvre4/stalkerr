## 1. Frontend image

- [x] 1.1 In `frontend/Dockerfile`, change the production stage base image from `nginx:alpine` to `nginxinc/nginx-unprivileged:alpine`
- [x] 1.2 In `frontend/Dockerfile`, change `EXPOSE 80` to `EXPOSE 8080`
- [x] 1.3 In `frontend/nginx.conf`, change `listen 80;` to `listen 8080;`

## 2. Helm chart wiring

- [x] 2.1 In `charts/stalkerr/templates/deployment-frontend.yaml`, change the container's `containerPort` from `80` to `8080`
- [x] 2.2 In `charts/stalkerr/templates/service-frontend.yaml`, change `targetPort` from `80` to `8080` (leave `port: {{ .Values.frontend.service.port }}` unchanged — external Service port stays 80)

## 3. Local dev parity

- [x] 3.1 In `docker-compose.yml`, change the frontend port mapping from `"${FRONTEND_PORT:-5173}:80"` to `"${FRONTEND_PORT:-5173}:8080"`

## 4. Verification

- [x] 4.1 Build the frontend image locally and run it under a non-root UID (e.g. `docker run --user 1000:0 ...` or via `docker compose`) to confirm Nginx starts without permission errors
- [x] 4.2 Run `helm template` (or `helm lint`) on `charts/stalkerr` with `values.test.yaml` and confirm the frontend Deployment/Service render with port `8080` internally and `80` externally
- [ ] 4.3 Deploy to the test cluster and confirm the `stalkerr-frontend` pod reaches `1/1 Running` with no restarts, and that the site is reachable through the existing Service/Ingress on port 80
