## Why

The release-triggered "Publish backend image" job takes ~10 minutes, and "Publish frontend image" ~4m30s, almost entirely because `docker/build-push-action` cross-builds `linux/arm64` under QEMU emulation on the `ubuntu-latest` (amd64) runner. Log inspection of run `34750849335` shows the backend's `go build` step alone takes 529s emulated vs 56.5s native (~9x), and the frontend's `npm install`/`npm run build` take 5-11x longer emulated. Separately, the backend Dockerfile hardcodes `GOARCH=amd64` regardless of target platform, so the published `linux/arm64` manifest likely already ships an amd64 binary — the arm64 build is both slow and not producing a correct arm64 artifact today. Dropping `linux/arm64` now removes ~9 minutes from the release pipeline immediately; a correct multi-arch setup (native cross-compilation via `--platform=$BUILDPLATFORM` + `TARGETARCH`, or native ARM runners) can be reintroduced later as its own change.

## What Changes

- Modify `docker-publish-backend` and `docker-publish-frontend` jobs in `.github/workflows/release-please.yml`: `platforms: linux/amd64,linux/arm64` -> `platforms: linux/amd64`.
- Remove the now-unneeded `docker/setup-qemu-action@v3` step from both jobs.
- **BREAKING**: future releases stop publishing `linux/arm64` images/tags for `germainlefebvre4/stalkerr` and `germainlefebvre4/stalkerr-frontend`. Previously published arm64 tags remain on Docker Hub but stop receiving updates.
- Document the reasoning and a concrete re-introduction path (native cross-compilation or native ARM runners) in `design.md` as a follow-up, out of scope for this change.

## Capabilities

### New Capabilities
(none)

### Modified Capabilities
- `release-automation`: the "Multi-arch Docker images published on release" requirement changes from amd64+arm64 to amd64-only, temporarily.

## Impact

- `.github/workflows/release-please.yml` (both Docker publish jobs).
- Docker Hub: `germainlefebvre4/stalkerr` and `germainlefebvre4/stalkerr-frontend` stop getting new `linux/arm64` layers on release; existing arm64 tags become stale rather than being removed.
- Anyone currently pulling these images on real arm64 hardware keeps whatever was last published (already suspected broken for the backend, per the `GOARCH=amd64` hardcoding found during investigation) and gets no further updates until arm64 is reintroduced correctly.
