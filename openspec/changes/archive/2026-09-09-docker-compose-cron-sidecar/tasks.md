## 1. Fix stale Docker Compose one-shot job services

- [x] 1.1 Remove the `sonarr-sync` and `radarr-sync` services from `docker-compose.yml` and verify no other file in the repo still references them (`grep -rn "sonarr-sync\|radarr-sync" .`)
- [x] 1.2 Remove the obsolete `resume-downloads` service from `docker-compose.yml` and verify `grep -n "resume-downloads" docker-compose.yml` no longer matches
- [x] 1.3 Add a `download` one-shot service to `docker-compose.yml` (command `["./stalkeer", "download", "--config", "/app/config/config.yml"]`, `restart: no`, `profiles: [stalkerr]`), matching the shape of the existing `process`/`m3u-download` services, and verify `docker compose --profile stalkerr config --services` lists exactly `process`, `download`, and `m3u-download` as job services (plus `server`, `postgres`)
- [x] 1.4 Verify `docker compose --profile stalkerr config` renders without errors

## 2. Add the opt-in cron sidecar

- [x] 2.1 Add an `ofelia` service to `docker-compose.yml` under a new `cron` profile, using a pinned `mcuadros/ofelia` image tag, with `/var/run/docker.sock` mounted so it can launch one-off job containers
- [x] 2.2 Add Ofelia `job-run` labels for `m3u-download` (`30 23 * * *`), `process` (`0 0 * * *`), and `download` (`0 */2 * * *`), each referencing the `germainlefebvre4/stalkerr:${VERSION:-dev}` image and the same command as the corresponding one-shot service
- [x] 2.3 Wire the same volumes and environment (including `TZ`) used by `x-stalkerr-volumes`/`x-stalkerr-environment` into the Ofelia job labels, and the `stalkerr-network` network, so scheduled runs behave the same as `docker compose run --rm --profile stalkerr <service>`
- [x] 2.4 Verify `docker-compose --profile stalkerr --profile cron up -d ofelia` starts and `docker compose logs ofelia` shows all three jobs registered with the expected schedules
- [x] 2.5 Verify `docker-compose up -d --profile stalkerr` (without the `cron` profile) does not start the `ofelia` service, confirming scheduling stays opt-in

## 3. Remove the dead in-app scheduling config

- [x] 3.1 Remove the `ScheduleEnabled`/`IntervalHours` fields (and their `viper.BindEnv`/`viper.SetDefault` calls) from the M3U download config in `internal/config/config.go`, and verify `go build ./...` succeeds
- [x] 3.2 Remove the `schedule_enabled`/`interval_hours` example lines from `config.yml.example`
- [x] 3.3 Remove the "(future feature)" scheduled-downloads section from `docs/M3U-DOWNLOAD.md`, and verify no remaining references anywhere (`grep -rn "schedule_enabled\|interval_hours" internal config.yml.example docs`)
- [x] 3.4 Run `go test ./internal/config/...` and confirm existing config tests still pass unmodified

## 4. Document the Compose scheduling recipe

- [x] 4.1 Replace the vague "Set up scheduled processing" step in `DOCKER-QUICKSTART.md`'s Next Steps with a concrete recipe covering both the `cron` profile and the host-crontab-plus-`docker compose run --rm --profile stalkerr <service>` alternative for users who don't want to grant Docker socket access
- [x] 4.2 Add a short "Scheduled Execution" note to `README.md`'s Docker deployment section, cross-referencing the Helm chart's `CronJob` documentation so both deployment paths point to equivalent guidance
- [x] 4.3 Check `docs/M3U-DOWNLOAD.md`'s remaining cron-related guidance (e.g. "Regular Downloads: Set up a cron job for periodic downloads") and point it at the same recipe rather than leaving a dangling/duplicate reference

## 5. End-to-end verification

- [x] 5.1 Run a full local `docker-compose --profile stalkerr --profile cron up -d` cycle against a test config; confirm `download`, `process`, and `m3u-download` each still run successfully via `docker compose run --rm --profile stalkerr <service>`
- [x] 5.2 Run `openspec validate docker-compose-cron-sidecar --strict` and confirm it passes before the change is considered ready to archive
