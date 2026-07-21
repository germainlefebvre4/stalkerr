## Why

Currently, Stalkeer operates as a background and CLI-driven application, making it difficult for users to track M3U playlist items, monitor ongoing downloads, view background processing logs, or re-organize media files without direct command-line or database queries. Providing a decoupled, modern React 19 web frontend with a dedicated, extended backend REST API solves these visibility and management gaps in a user-friendly manner.

## What Changes

- Create a brand new decoupled frontend application inside the `/frontend` directory using React 19, Radix UI, TypeScript, and a modern Light Theme.
- Extend the backend REST API with new dedicated endpoints for querying processing logs, tracking download states, and managing media file structures.
- Implement folder-level moving capabilities to safely transfer the entire parent directory of a movie or TV show (not single files or seasons) to a new parent path, updating database records accordingly.
- Enhance the Helm Chart and Docker configurations to package and deploy the frontend as a separate, light-weight container service alongside the existing backend API.

## Capabilities

### New Capabilities

- `api-processing-logs`: Dedicated API endpoints to retrieve, query, and monitor historical and active background execution logs from the `processing_logs` table.
- `api-downloads-monitoring`: API endpoints to query tracking information and dynamic progress for active, pending, completed, or failed downloads from the `download_info` table.
- `api-media-management`: API endpoint to safely move the complete parent directory of an entire movie or TV show from its current location to a new base parent path on the file system, adjusting all associated database records.
- `frontend-ihm-dashboard`: A responsive, modern light-themed dashboard built with React 19 and Radix UI primitives, facilitating real-time playlist exploration, download tracking, log viewing, and folder-level re-organizations.
- `helm-chart-frontend-integration`: Multi-pod Helm chart configurations to deploy, configure, and route the decoupled frontend application using a dedicated Nginx container, Kubernetes service, and Ingress paths.

### Modified Capabilities

<!-- No existing capabilities are being modified -->

## Impact

- **New Folders/Files**:
  - `/frontend`: React 19 + Radix UI + TypeScript frontend SPA workspace.
  - `/frontend/Dockerfile`: Multi-stage compilation and Nginx-based deployment container for the SPA.
  - `internal/api/handlers_logs.go`, `internal/api/handlers_downloads.go`, `internal/api/handlers_media.go`: Modular backend API handlers.
  - `charts/stalkerr/templates/deployment-frontend.yaml`, `charts/stalkerr/templates/service-frontend.yaml`: New Helm templates for the frontend.
- **Modified Files**:
  - `internal/api/api.go`: Registering new endpoints and path queries.
  - `charts/stalkerr/values.yaml`, `charts/stalkerr/templates/ingress.yaml`: Updating Helm values and adding ingress routes for the frontend.
  - `docker-compose.yml`: Adding the `frontend` container service to the development profiles.
- **Dependencies**:
  - Frontend: `react@19`, `react-dom@19`, `@radix-ui/*` primitives, `typescript`, `vite`.
  - Backend: No new external library dependencies (uses Go standard library and Gin/GORM).
- **Migration**: No database migration is required since we utilize existing `processing_logs` and `download_info` tables; only directory structure moves require file-system updates.
