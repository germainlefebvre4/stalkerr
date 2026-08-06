## Context

The frontend pod (`stalkerr-frontend`) crash-loops with:
```
mkdir() "/var/cache/nginx/client_temp" failed (13: Permission denied)
```
The chart's `.Values.securityContext` (`runAsNonRoot: true`, `runAsUser: 1000`) is shared globally between the Go backend container and the Nginx frontend container (`charts/stalkerr/templates/deployment-frontend.yaml`). The frontend image (`frontend/Dockerfile`) is a stock `FROM nginx:alpine`: it expects to start as root and drop privileges to the `nginx` user internally, and its baked-in `/var/cache/nginx` directory is owned `root:root`. Forced to start directly as UID 1000, the master process cannot create its own cache subdirectories. See `proposal.md` for the full symptom analysis.

## Goals / Non-Goals

**Goals:**
- Make the frontend container start reliably under the chart's existing non-root enforcement, with no chart-wide security regression (no relaxing `runAsNonRoot`/`allowPrivilegeEscalation`/`capabilities.drop`).
- Keep the change confined to the frontend image and its immediate port wiring — no change to the shared `securityContext` values used by the backend.

**Non-Goals:**
- Reworking the shared `securityContext` model to give frontend and backend independent security contexts (would be a larger chart refactor, not needed here).
- Enabling `readOnlyRootFilesystem` for the frontend (out of scope; not required by this fix).

## Decisions

**Decision: switch to `nginxinc/nginx-unprivileged:alpine` instead of patching the existing `nginx:alpine` image (piste A), adding chart-side `emptyDir` volumes (piste C), or relaxing `runAsNonRoot` (piste D).**

Rationale: `nginxinc/nginx-unprivileged` is the officially maintained image built exactly for this scenario. Verified directly from its Dockerfile (not assumed):
- `/var/cache/nginx` and `/etc/nginx` are `chown $UID:0` + `chmod g+w` at build time — group-owned by GID 0 (root group) with group-write access.
- Its base `nginx.conf` relocates the PID file and all temp paths (client/proxy/fastcgi/uwsgi/scgi) under `/tmp`, and removes the now-meaningless `user nginx;` directive (this is exactly the warning line seen in the crash logs).
- It listens on `8080` by default, since binding `<1024` requires `CAP_NET_BIND_SERVICE`, which a non-root, no-added-capabilities container does not have (and shouldn't be granted just for this).

Why this works under our specific chart configuration: Kubernetes sets `runAsUser: 1000` without a matching `runAsGroup`. Because UID 1000 has no `/etc/passwd` entry in this image (only UID 101 `nginx` does), the container runtime falls back to primary **GID 0** — the same mechanism OpenShift's arbitrary-UID support relies on. GID 0 plus the image's `g+w` permissions on the cache/config directories means UID 1000 can write there. Belt-and-suspenders: even if that fallback didn't apply, the PID/temp files now live under `/tmp`, which is writable regardless (root filesystem is not read-only in this chart: `securityContext.readOnlyRootFilesystem: false`).

Alternatives considered:
- *Piste A (patch current Dockerfile)*: fixes it too, but means hand-maintaining chown/temp-path patches that the unprivileged image already gets for free from upstream.
- *Piste C (chart-side emptyDir volumes)*: works without touching the image, but doesn't fix the underlying image assumption (root startup) and adds volume-mount complexity for no real benefit here.
- *Piste D (drop non-root enforcement for frontend)*: fastest but weakens the security posture the chart otherwise applies consistently across containers.

## Risks / Trade-offs

- **[Risk]** The fix's reliability depends on the *implicit* GID-0 fallback (no `runAsGroup` set). If `runAsGroup: 1000` is later added to `.Values.securityContext` for stricter posture, the GID-0 write path breaks and the same class of failure could resurface for anything still writing outside `/tmp`. → **Mitigation**: this design doc records the dependency explicitly; the `/tmp` relocation of PID/temp paths (done by the base image, not by us) is the actual safety net and holds regardless of GID, as long as `readOnlyRootFilesystem` stays `false` for this container.
- **[Risk]** Port change (80 → 8080 internally) is a breaking change to anything that assumed the frontend container listens on 80 directly (e.g., local `kubectl port-forward` scripts, direct container debugging). → **Mitigation**: the externally-facing Service port (`frontend.service.port`) and Ingress routing are untouched — only `targetPort`/`containerPort`/`EXPOSE` change, which are internal wiring documented in the proposal's Impact section.
- **[Trade-off]** Pinning to `nginxinc/nginx-unprivileged:alpine` (a third-party-maintained, NGINX-Inc-published image) instead of the official `nginx:alpine` adds a dependency on that image's maintenance cadence. Accepted: it's an actively maintained, purpose-built image from the same publisher (NGINX Inc.) as the base image already in use.

## Migration Plan

1. Update `frontend/Dockerfile`, `frontend/nginx.conf`, and the chart templates/`docker-compose.yml` (see `tasks.md`).
2. Rebuild and push the frontend image under its existing repository (`germainlefebvre4/stalkeer-frontend`) — no new image name needed, the base image change is transparent to consumers of that tag.
3. Roll out via the existing Helm release process; since this only changes the frontend Deployment/Service, it is independently rollout-able from the backend.
4. Rollback: revert the image tag and the two port-mapping template edits; no data or schema migration is involved (stateless frontend).
