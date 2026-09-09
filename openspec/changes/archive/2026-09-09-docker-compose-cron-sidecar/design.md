## Context

See `proposal.md` - Why for the motivation. Relevant current state:

- `docker-compose.yml` already defines one-shot job containers (`profiles: [stalkerr]`, `restart: no`) meant to be triggered externally, sharing YAML anchors for build args, volumes, environment, and resource limits (`x-stalkerr-build`, `x-stalkerr-volumes`, `x-stalkerr-environment`, `x-stalkerr-resources`).
- The Helm chart already solves the equivalent problem for Kubernetes with three `CronJob` resources (`charts/stalkerr/templates/cronjob-*.yaml`) whose schedules live in `values.yaml` (`30 23 * * *`, `0 0 * * *`, `0 */2 * * *`) and which set a `timeZone` on the CronJob.
- No process inside the Stalkeer Go binary or its container image currently triggers anything on a schedule; `internal/scheduler` is an in-memory tiered stream picker used *within* a `download` run, not a cron.

## Goals / Non-Goals

**Goals:**
- Give Docker Compose deployments an opt-in way to run `m3u-download` / `process` / `download` on the same default cadence as Helm, without adding scheduling logic to the Go application.
- Keep the change purely at the orchestration layer (compose file + docs), consistent with fully committing to external-only scheduling.
- Preserve today's default `docker-compose up -d --profile stalkerr` behavior for anyone not opting into scheduling.

**Non-Goals:**
- Building a scheduler inside Stalkeer itself (explicitly rejected during exploration - see the `compose-scheduled-jobs` capability's Purpose in `specs/compose-scheduled-jobs/spec.md`).
- Supporting arbitrary/user-defined schedules beyond simple compose-level overrides of the three existing cron expressions - no new UI, API, or config-file-driven scheduling.
- Swarm or other multi-host orchestration - single-host Docker Compose only, matching the project's existing Compose scope.

## Decisions

### Sidecar tool: Ofelia (`mcuadros/ofelia`), not a custom cron image
Ofelia is a purpose-built Docker job scheduler driven entirely by container labels, with no image to build or crontab file to maintain - job definitions live next to the services they schedule, mirroring how the Helm chart keeps schedules in `values.yaml` next to the CronJob templates.

Alternative considered: a minimal custom image (busybox/alpine + `docker` CLI + `crond`) running `docker compose run --rm --profile stalkerr <service>` on a crontab. Rejected: it requires bundling and maintaining the Docker CLI (and Compose plugin) inside a bespoke image, and would still need `docker.sock` mounted, i.e. it carries the exact same trade-off as Ofelia (see Risks) for strictly more code to own.

### One-shot job services fixed first, then scheduled
`sonarr-sync` / `radarr-sync` (invoke removed `sonarr`/`radarr` commands) and `resume-downloads` (obsolete, resume is automatic inside `download`) are removed from `docker-compose.yml`; a `download` one-shot service is added. Ofelia's job definitions target only the corrected set (`process`, `download`, `m3u-download`), so there is nothing left to schedule that doesn't already work when run manually via `docker compose run --rm --profile stalkerr <service>`.

### Scheduling is opt-in via a new `cron` profile
The Ofelia service is placed in its own Compose profile (`cron`), separate from `stalkerr` and `development`. Starting the stack the way the README already documents (`docker-compose up -d` / `--profile stalkerr`) is unaffected; a user opts in with `docker-compose --profile stalkerr --profile cron up -d`. This directly satisfies the "Cron sidecar is optional" scenario in the spec and avoids granting Docker socket access (see Risks) to anyone who didn't ask for it.

### Default schedule and timezone mirror the Helm chart exactly
Ofelia job labels reuse the same three cron expressions as `charts/stalkerr/values.yaml` (`m3u-download`: `30 23 * * *`, `process`: `0 0 * * *`, `download`: `0 */2 * * *`), and the sidecar reads the same `TZ` environment variable already present in `x-stalkerr-environment` (defaulting to `Etc/UTC`), matching the Helm chart's `timeZone` field. This is a direct instance of the parity goal in the proposal, not a new decision to design from scratch.

### Concurrency: enable Ofelia's `no-overlap` middleware per job
Ofelia does not prevent overlapping runs of the same job by default - it requires the `no-overlap: true` label per job (`middlewares/overlap.go`; opt-in, not automatic). Each `job-run` label set in `docker-compose.yml` sets `ofelia.job-run.<job>.no-overlap: "true"`, which satisfies the spec's "No overlapping scheduled runs" requirement with a one-line label per job rather than any custom locking logic - the same property the Helm CronJobs get from `concurrencyPolicy`.

## Risks / Trade-offs

- **[Risk] Triggering one-shot containers from a sidecar requires Docker socket access** (`/var/run/docker.sock`), which is effectively root-equivalent access to the host. This is inherent to any container-based scheduler triggering sibling containers under Compose (Ofelia or a custom crond image alike) - it is not something a different tool choice avoids.
  → **Mitigation**: the sidecar lives behind its own opt-in `cron` profile (never started by default); `README.md`/`DOCKER-QUICKSTART.md` state the socket-access trade-off explicitly and document the host-crontab-plus-`docker compose run` recipe as a lower-privilege alternative for users who prefer not to grant it.
- **[Risk] Duplication between compose service definitions and Ofelia job labels.** Ofelia's `job-run` labels need their own image/volume/environment/network values; they can't directly reference the existing `x-stalkerr-volumes`/`x-stalkerr-environment` YAML anchors used by the `process`/`download`/`m3u-download` services, so the two must be kept in sync by hand when either changes.
  → **Mitigation**: keep the job labels minimal (reuse the same image tag and env vars already exported into the compose environment) and note the duplication explicitly in a comment above the Ofelia service definition, so future edits to the one-shot services remember to check the labels too.
- **[Risk] BREAKING removal of `sonarr-sync`/`radarr-sync`/`resume-downloads` services** could affect anyone with an external script or crontab entry still referencing them (they've been non-functional since the CLI unification, but the service names still existed).
  → **Mitigation**: called out explicitly as **BREAKING** in the proposal; `resume-downloads`/`sonarr`/`radarr` have been documented as replaced by `download` since the CLI unification, so this only removes already-dead entry points.

## Migration Plan

No data migration. Deployment-level changes only:
1. Ship the corrected `docker-compose.yml` (fixed one-shot services + optional `cron`-profile Ofelia service) and the config/docs cleanup together in one release.
2. Existing users on the default `stalkerr` profile see no behavior change other than the three removed/added service names.
3. Users who want scheduling opt in explicitly by adding the `cron` profile; nothing runs automatically as a result of this change alone.
4. Rollback is a plain revert of `docker-compose.yml`, `internal/config/config.go`, `config.yml.example`, and the two doc files - no database or running-state impact.

## Open Questions

- Exact Ofelia label syntax for wiring `network`/`volume`/`environment` per job (e.g. whether to reuse the existing `stalkerr-network` name or derive it from `COMPOSE_PROJECT_NAME`) is an implementation detail to finalize while writing the compose file - it won't change the chosen approach or the spec.
