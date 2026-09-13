## 1. Backend publish job

- [ ] 1.1 In `.github/workflows/release-please.yml`, change `docker-publish-backend`'s `platforms:` from `linux/amd64,linux/arm64` to `linux/amd64`
- [ ] 1.2 Remove the `docker/setup-qemu-action@v3` step from `docker-publish-backend` and verify `docker/setup-buildx-action@v3` alone still succeeds for a single-platform build

## 2. Frontend publish job

- [ ] 2.1 In `.github/workflows/release-please.yml`, change `docker-publish-frontend`'s `platforms:` from `linux/amd64,linux/arm64` to `linux/amd64`
- [ ] 2.2 Remove the `docker/setup-qemu-action@v3` step from `docker-publish-frontend`

## 3. Verification

- [ ] 3.1 Trigger (or wait for) the next release-please release and confirm both publish jobs complete in roughly native-build time (backend ~1 min instead of ~10 min, frontend well under its previous ~4m30s), with no `linux/arm64` entry in the pushed manifest (`docker manifest inspect germainlefebvre4/stalkerr:<tag>` / `...-frontend:<tag>` shows only `linux/amd64`)
- [ ] 3.2 Confirm `germainlefebvre4/stalkerr:latest` and `germainlefebvre4/stalkerr-frontend:latest` still pull and run correctly on an amd64 host after the change
