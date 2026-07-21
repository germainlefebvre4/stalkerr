## Context

Stalkeer coordinates the download, processing, and curation of media streams from an M3U playlist. Currently, these activities (represented in database tables such as `processed_lines`, `processing_logs`, and `download_info`) are only inspectable or manageable through backend CLI tools or raw database queries. 

To make the application more accessible and support user-friendly directory re-organizations, we are introducing a responsive React 19 web frontend with a dedicated REST API. Per user specifications, the frontend and backend must be physically decoupled in deployment, utilize Radix UI accessible primitives, feature an elegant modern Light theme, and allow moving the complete parent directory of films or TV shows (rather than individual files or seasons).

## Goals / Non-Goals

**Goals:**
- Provide a completely decoupled React 19 + Radix UI + TypeScript frontend.
- Implement an elegant, modern Light-themed dashboard.
- Create secure API endpoints to retrieve processing logs, track download progress, and retrieve configurations.
- Develop a robust file-system reorganization endpoint that moves the complete parent directory of a movie or TV show to a new target parent path and updates all linked database references to prevent out-of-sync paths.
- Add Helm chart deployment and routing configurations for the separate frontend container.

**Non-Goals:**
- Allowing individual file moves or single-season folder moves (the unit of reorganization is the complete work).
- Implementing user access control or authentication on the frontend in this release (assumes private home server network).
- Replacing the existing Go server router (continues utilizing Gin and GORM).

## Decisions

### 1. Separate Frontend Deployment with Nginx (Option B)
- **Rationale**: To enforce physical decoupling between frontend and backend, the frontend is packaged as a distinct Docker image utilizing Nginx to serve static files. 
- **Alternative Considered**: Embedding frontend static files in the Go binary using `go:embed`. While simpler, it runs in the same process/port, violating the explicit requirement to distinguish the frontend and backend deployments.
- **Implementation**: Write a lightweight Multi-Stage `Dockerfile` inside `/frontend` that compiles the React 19 SPA with Node 22 and packages it in an Alpine-based Nginx image.

### 2. Move Granularity at the Work Parent Directory Level
- **Rationale**: Movies and TV shows must reside in structured root paths for downstream indexers (Plex, Jellyfin, Sonarr, Radarr) to index them properly. Moving a single file or an individual season folder breaks metadata linkage. Thus, the move operation is performed on the parent directory of the entire movie or TV show.
- **Folder Derivation**:
  - Movie: `filepath.Dir(download_path)` of any associated download.
  - TV Show: `filepath.Dir(filepath.Dir(download_path))` to jump up from `Season XX/` to the main series folder.
- **Database Alignment**: Update the `download_path` columns of all associated `download_info` records in a single database transaction.

### 3. React 19, Radix UI Primitives, and modern Light Theme CSS Modules
- **Rationale**: Radix UI offers accessible, unstyled primitives (*Dialog*, *Select*, *Progress*), allowing us to implement a tailored, elegant custom Light theme using CSS Modules and native CSS variables without being locked into or bloated by TailwindCSS.
- **Color Palette**:
  - Base Background: `#f8f9fa` (soft off-white)
  - Card/Dialog Surface: `#ffffff` (pure white)
  - Primary Slate: `#1e293b` (slate-800 for high readability typography)
  - Primary Accent: `#2563eb` (cobalt blue for active controls and links)
  - Secondary Gray: `#64748b` (slate-500)
  - Borders: `#e2e8f0` (thin, clean slate-200 lines)

### 4. Helm Chart Integration via Multi-Pod and Ingress Routing
- **Rationale**: To seamlessly host both containers on a single domain and avoid CORS constraints, we add a `deployment-frontend.yaml` and a `service-frontend.yaml` to the Helm chart.
- **Ingress Configuration**: Route `/api/v1/*` paths to the backend container, and route all other paths (`/`) to the frontend Nginx container.

## Risks / Trade-offs

- **[Risk] Cross-Filesystem Moves** → Move operations across different mount points or physical disks can fail under `os.Rename` with an `EXDEV` error.
  - *Mitigation*: Leverage the existing fallback pattern in `internal/downloader/downloader.go` (copy all folder contents recursively, verify file sizes/counts, and only then delete the source folder).
- **[Risk] Path Out-of-Sync** → Moving directories can cause database records (`download_path` in `download_info`) to point to stale paths.
  - *Mitigation*: Execute the folder movement and database updates in a single unified API handler within a database transaction. If any part of the move or database save fails, rollback and keep the original state.
