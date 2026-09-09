## Why

Kubernetes users get periodic M3U download/process/download runs out of the box via three well-tuned Helm `CronJob` resources (`charts/stalkerr/templates/cronjob-*.yaml`, schedules defined in `values.yaml`). Docker Compose users get none of that: `DOCKER-QUICKSTART.md` tells them to "Set up scheduled processing" with no concrete recipe, and `docker-compose.yml`'s one-shot job services are stale — `sonarr-sync`/`radarr-sync` invoke the `sonarr`/`radarr` CLI commands that were removed when they were unified into `download`, `resume-downloads` is now obsolete (resume is automatic inside `download`), and there is no `download` service at all. On top of that, `m3u.download.schedule_enabled`/`interval_hours` have sat in `internal/config/config.go` and `config.yml.example` since February, explicitly documented in `docs/M3U-DOWNLOAD.md` as "(future feature)", parsed and defaulted but never consumed anywhere — implying Stalkeer might schedule itself when it never has. This change fully commits to external-only scheduling: it gives Compose deployments the same scheduling parity Helm already has, and removes the config surface that falsely suggests otherwise.

## What Changes

- Fix the stale one-shot job services in `docker-compose.yml`: remove `sonarr-sync` and `radarr-sync` (call removed CLI commands `sonarr`/`radarr`), remove `resume-downloads` (obsolete now that resume runs automatically inside `download`), and add a `download` one-shot service — so the `stalkerr` profile's job services match the current CLI surface (`process`, `download`, `m3u-download`), mirroring what Helm already schedules.
- Add an optional cron sidecar service to `docker-compose.yml` that triggers `process`, `download`, and `m3u-download` on the same default cadence as Helm's `values.yaml` (m3u-download daily 23:30, process daily 00:00, download every 2 hours), via `docker compose run --rm --profile stalkerr <service>` — giving Compose deployments scheduling parity with Helm without adding any scheduling logic to the Go binary itself.
- **BREAKING**: Remove the unused `m3u.download.schedule_enabled` / `interval_hours` config fields — including their struct fields, `viper.BindEnv`/`viper.SetDefault` calls in `internal/config/config.go`, the example block in `config.yml.example`, and the "(future feature)" section in `docs/M3U-DOWNLOAD.md`. These fields have never been implemented; anyone with them set today has them silently ignored, so removal changes no runtime behavior, only the documented config surface.
- Update `README.md` and `DOCKER-QUICKSTART.md` to replace the vague "Set up scheduled processing" step with the concrete recipe: use the Compose cron sidecar (or an equivalent host crontab calling `docker compose run --rm --profile stalkerr <service>`), matching the guidance Kubernetes users already get for the Helm chart.

## Capabilities

### New Capabilities
- `compose-scheduled-jobs`: Docker Compose cron-driven scheduling for the M3U download/process/download pipeline — corrected one-shot job service definitions matching the current CLI, plus an optional cron sidecar reproducing Helm's default schedule, giving Compose deployments parity with the existing Helm `scheduled-jobs` capability.

### Modified Capabilities
(none — `scheduled-jobs` and `configuration-management` describe Helm/Kubernetes-only behavior and are unaffected; `local-dev`'s Docker Compose requirements concern API port mapping only and are unaffected by this change)

## Impact

- **Code/config**: `docker-compose.yml` (fixed one-shot services + new cron sidecar service), `internal/config/config.go` (remove dead fields, bindings, defaults), `config.yml.example`, `docs/M3U-DOWNLOAD.md`, `README.md`, `DOCKER-QUICKSTART.md`.
- **No Go application behavior change** beyond deleting dead config-parsing code: no new Go source files, no API changes, no DB migrations, no changes to the `download`/`process`/`m3u-download` commands themselves.
- Users who currently set `m3u.download.schedule_enabled`/`interval_hours` in their `config.yml` (even though it has always done nothing) will find those keys no longer documented; the values remain harmlessly ignored by Viper's loose YAML parsing.
