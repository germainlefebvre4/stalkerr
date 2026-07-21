## 1. Backend API Extension

- [x] 1.1 Implement processing logs endpoint `GET /api/v1/processing-logs` with pagination and status filtering
- [x] 1.2 Implement downloads monitoring endpoint `GET /api/v1/downloads` with pagination and status filtering
- [x] 1.3 Export `MoveFile` helper or create folder-moving logic supporting cross-filesystem fallbacks in Go
- [x] 1.4 Implement media management endpoints `POST /api/v1/movies/:id/move` and `POST /api/v1/tvshows/:id/move` with folder-moving logic and database transaction updates
- [x] 1.5 Implement configured root paths endpoint `GET /api/v1/config/paths`
- [x] 1.6 Register new routes in `internal/api/api.go` and write comprehensive tests for new handlers

## 2. Frontend Project Setup

- [x] 2.1 Initialize Vite + React 19 + TypeScript + CSS Modules template in the `/frontend` directory
- [x] 2.2 Add Radix UI primitive dependencies (`@radix-ui/react-dialog`, `@radix-ui/react-progress`, `@radix-ui/react-select`, `@radix-ui/react-tabs`)
- [x] 2.3 Configure CSS variables for the modern Light Theme design system
- [x] 2.4 Setup base layout, client-side API routing helper, and basic fetch/polling custom hooks

## 3. Frontend Views Implementation

- [x] 3.1 Build `PlaylistManager` view with search and content-type filtering
- [x] 3.2 Build `ProcessingLogs` view with auto-refreshing logs and expandable error blocks
- [x] 3.3 Build `DownloadTracker` view rendering active downloading progress bars
- [x] 3.4 Build `MoveModal` utilizing Radix UI Dialog to display target parent selections and trigger directory moves

## 4. Docker & Helm Chart Integration

- [x] 4.1 Write `/frontend/Dockerfile` for multi-stage building and Nginx static hosting
- [x] 4.2 Create frontend Deployment and Service templates in `charts/stalkerr/templates`
- [x] 4.3 Update Helm values (`values.yaml`, `values.schema.json`) with frontend options
- [x] 4.4 Update Helm Ingress template to route `/` to the frontend and `/api/v1/*` to the backend API

## 5. End-to-End Validation

- [x] 5.1 Add docker-compose configuration for the new frontend service
- [x] 5.2 Validate API responses and file movement robustness under simulation
- [x] 5.3 Verify theme, layout responsiveness, and accessibility compliance of all Radix UI modals
