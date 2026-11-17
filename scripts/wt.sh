#!/usr/bin/env bash
# wt.sh - isolated Stalkeer dev instance per git worktree (Postgres + API + Vite).
#
# Usage: scripts/wt.sh <up|down|status|logs|plan>
#   (normally invoked through `make wt-up | wt-down | wt-status | wt-logs`)
#
# Environment (set by the Makefile, see `## wt-up` help):
#   WT_API_PIN         pinned API port (unset/empty => dynamic)
#   WT_FRONT_PIN       pinned frontend port from FRONTEND_PORT (unset/empty => dynamic)
#   WT_FRONT_ALIAS_PIN pinned frontend port from the FRONT_PORT alias
#   WT_SEED=0          skip loading scripts/seed-test-data.sql
#   WT_FOLLOW=0        `logs`: print the tail and exit instead of following
#   WT_READY_TIMEOUT   seconds to wait for each component (default 90)
#
# Everything about the instance (ports, process groups, logs) lives in .wt/
# (git-ignored). The database is a dedicated compose project created from
# docker-compose.wt.yml; `down` destroys it together with its volume.

set -uo pipefail

ROOT="$(git rev-parse --show-toplevel 2>/dev/null)" || {
  echo "wt: not inside a git worktree" >&2
  exit 1
}
cd "$ROOT" || exit 1

STATE_DIR="$ROOT/.wt"
STATE_FILE="$STATE_DIR/state.env"
COMPOSE_FILE="$ROOT/docker-compose.wt.yml"
READY_TIMEOUT="${WT_READY_TIMEOUT:-90}"
PORT_MIN=20000
PORT_SPAN=10000

DB_USER_WT=postgres
DB_PASSWORD_WT=postgres
DB_NAME_WT=stalkeer

# ---------------------------------------------------------------- identity ---

hash_num() { printf '%s' "$1" | cksum | cut -d' ' -f1; }

make_slug() {
  local base short
  base="$(basename "$ROOT" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//' | cut -c1-30)"
  base="${base%-}"
  [ -n "$base" ] || base="wt"
  short="$(printf '%06x' $(($(hash_num "$ROOT") % 16777216)))"
  printf '%s-%s' "$base" "$short"
}

SLUG="$(make_slug)"
PROJECT="stalkeer-wt-$SLUG"

log() { printf 'wt: %s\n' "$*"; }
err() { printf 'wt: error: %s\n' "$*" >&2; }

# ------------------------------------------------------------------- state ---

state_get() { # key -> value ("" if absent)
  [ -f "$STATE_FILE" ] || return 0
  sed -n "s/^$1=//p" "$STATE_FILE" | tail -n1
}

state_set() { # key value
  mkdir -p "$STATE_DIR"
  touch "$STATE_FILE"
  if grep -q "^$1=" "$STATE_FILE" 2>/dev/null; then
    sed -i "s|^$1=.*|$1=$2|" "$STATE_FILE"
  else
    printf '%s=%s\n' "$1" "$2" >>"$STATE_FILE"
  fi
}

# ------------------------------------------------------------------- ports ---

port_in_use() {
  local p="$1"
  if command -v ss >/dev/null 2>&1; then
    ss -ltnH 2>/dev/null | awk '{print $4}' | grep -Eq "[:.]${p}\$"
  else
    (exec 3<>"/dev/tcp/127.0.0.1/$p") 2>/dev/null
  fi
}

# pick_port <seed> <already-chosen ports...>: deterministic start, probe upward.
pick_port() {
  local seed="$1"; shift
  local p i taken
  p=$((PORT_MIN + $(hash_num "$ROOT:$seed") % PORT_SPAN))
  for ((i = 0; i < PORT_SPAN; i++)); do
    taken=0
    for c in "$@"; do [ "$c" = "$p" ] && taken=1; done
    if [ "$taken" = 0 ] && ! port_in_use "$p"; then
      printf '%s' "$p"
      return 0
    fi
    p=$((PORT_MIN + (p - PORT_MIN + 1) % PORT_SPAN))
  done
  return 1
}

valid_port() { [[ "$1" =~ ^[0-9]+$ ]] && [ "$1" -ge 1 ] && [ "$1" -le 65535 ]; }

# resolve_ports: sets API_PORT_R FRONT_PORT_R METRICS_PORT_R, API_PINNED FRONT_PINNED.
# No side effects; returns non-zero with an explicit message on any problem.
resolve_ports() {
  local api_pin="${WT_API_PIN:-}" front_pin="${WT_FRONT_PIN:-}" alias_pin="${WT_FRONT_ALIAS_PIN:-}"

  if [ -n "$front_pin" ] && [ -n "$alias_pin" ] && [ "$front_pin" != "$alias_pin" ]; then
    err "conflicting frontend ports: FRONTEND_PORT=$front_pin and FRONT_PORT=$alias_pin (use only one)"
    return 1
  fi
  [ -n "$front_pin" ] || front_pin="$alias_pin"

  local v
  for v in "API_PORT:$api_pin" "FRONTEND_PORT:$front_pin"; do
    if [ -n "${v#*:}" ] && ! valid_port "${v#*:}"; then
      err "invalid ${v%%:*} '${v#*:}' (expected 1-65535)"
      return 1
    fi
  done
  if [ -n "$api_pin" ] && [ "$api_pin" = "$front_pin" ]; then
    err "API_PORT and frontend port are both pinned to $api_pin"
    return 1
  fi

  API_PINNED=0 FRONT_PINNED=0
  if [ -n "$api_pin" ]; then
    API_PINNED=1
    if port_in_use "$api_pin"; then err "pinned API port $api_pin is already in use"; return 1; fi
    API_PORT_R="$api_pin"
  fi
  if [ -n "$front_pin" ]; then
    FRONT_PINNED=1
    if port_in_use "$front_pin"; then err "pinned frontend port $front_pin is already in use"; return 1; fi
    FRONT_PORT_R="$front_pin"
  fi
  if [ "$API_PINNED" = 0 ]; then
    API_PORT_R="$(pick_port api "${front_pin:-}")" || { err "no free port for the API"; return 1; }
  fi
  if [ "$FRONT_PINNED" = 0 ]; then
    FRONT_PORT_R="$(pick_port front "$API_PORT_R")" || { err "no free port for the frontend"; return 1; }
  fi
  METRICS_PORT_R=""
  if [ "${STALKEER_METRICS_ENABLED:-}" = "true" ]; then
    METRICS_PORT_R="$(pick_port metrics "$API_PORT_R" "$FRONT_PORT_R")" || { err "no free port for metrics"; return 1; }
  fi
  return 0
}

# ---------------------------------------------------------------- processes --

alive() { [ -n "${1:-}" ] && kill -0 "$1" 2>/dev/null; }

# pids of every process carrying this instance's WT_INSTANCE marker (Linux).
marker_pids() {
  local d pid
  for d in /proc/[0-9]*; do
    pid="${d#/proc/}"
    [ "$pid" = "$$" ] && continue
    if { tr '\0' '\n' <"$d/environ" | grep -qx "WT_INSTANCE=$SLUG"; } 2>/dev/null; then
      printf '%s\n' "$pid"
    fi
  done
}

kill_instance_processes() {
  local pids pid pgid
  pids="$(marker_pids)"
  # also the recorded groups, if their leader still carries our marker (PID reuse guard)
  for pgid in "$(state_get API_PGID)" "$(state_get FRONT_PGID)"; do
    [ -n "$pgid" ] || continue
    if { tr '\0' '\n' <"/proc/$pgid/environ" | grep -qx "WT_INSTANCE=$SLUG"; } 2>/dev/null; then
      kill -TERM -- "-$pgid" 2>/dev/null || true
    fi
  done
  for pid in $pids; do kill -TERM "$pid" 2>/dev/null || true; done
  local i
  for ((i = 0; i < 50; i++)); do
    [ -z "$(marker_pids)" ] && break
    sleep 0.2
  done
  for pid in $(marker_pids); do kill -KILL "$pid" 2>/dev/null || true; done
  for pgid in "$(state_get API_PGID)" "$(state_get FRONT_PGID)"; do
    [ -n "$pgid" ] && kill -KILL -- "-$pgid" 2>/dev/null || true
  done
  [ -n "$pids" ]
}

compose() { docker compose -f "$COMPOSE_FILE" -p "$PROJECT" "$@"; }

db_containers() { compose ps -aq 2>/dev/null; }

# teardown <keep_logs 0|1> -> returns 0 if something was found/removed
teardown() {
  local keep_logs="${1:-0}" found=1
  [ -f "$STATE_FILE" ] && found=0
  kill_instance_processes && found=0
  [ -n "$(db_containers)" ] && found=0
  compose down -v --remove-orphans >/dev/null 2>&1 || true
  if [ "$keep_logs" = 1 ]; then
    rm -f "$STATE_FILE"
  else
    rm -rf "$STATE_DIR"
  fi
  return $found
}

# -------------------------------------------------------------------- probes -

http_ok() { curl -fsS -o /dev/null --max-time 3 "$1" 2>/dev/null; }

wait_http() { # url pid component logfile
  local url="$1" pid="$2" name="$3" logf="$4" i
  for ((i = 0; i < READY_TIMEOUT * 2; i++)); do
    if http_ok "$url"; then return 0; fi
    if ! alive "$pid"; then
      err "$name exited during startup (log: $logf)"
      return 1
    fi
    sleep 0.5
  done
  err "$name did not become ready within ${READY_TIMEOUT}s (log: $logf)"
  return 1
}

db_port() {
  compose port postgres 5432 2>/dev/null | sed -E 's/.*:([0-9]+)$/\1/'
}

instance_running() {
  local ap fp
  ap="$(state_get API_PORT)"; fp="$(state_get FRONT_PORT)"
  [ -n "$ap" ] && [ -n "$fp" ] || return 1
  alive "$(state_get API_PGID)" && alive "$(state_get FRONT_PGID)" \
    && http_ok "http://localhost:$ap/health" && http_ok "http://localhost:$fp/"
}

print_urls() {
  log "frontend: http://localhost:$(state_get FRONT_PORT)"
  log "API:      http://localhost:$(state_get API_PORT)"
  log "database: 127.0.0.1:$(state_get DB_PORT) (user $DB_USER_WT, db $DB_NAME_WT)"
  log "data is ephemeral: 'make wt-down' destroys the database and its volume"
}

# ----------------------------------------------------------------------- up --

STARTED=0
fail_up() { # message
  err "$1"
  err "cleaning up..."
  teardown 1 >/dev/null 2>&1
  err "wt-up failed; logs kept in $STATE_DIR/"
  trap - INT TERM
  exit 1
}

on_signal() {
  err "interrupted"
  teardown 1 >/dev/null 2>&1
  exit 130
}

check_prereqs() {
  local missing=0 b
  for b in git go npm curl; do
    command -v "$b" >/dev/null 2>&1 || { err "missing prerequisite: $b"; missing=1; }
  done
  docker compose version >/dev/null 2>&1 || { err "missing prerequisite: docker compose (v2 plugin)"; missing=1; }
  [ -f "$COMPOSE_FILE" ] || { err "missing $COMPOSE_FILE"; missing=1; }
  return $missing
}

scrub_env() {
  local v
  unset DATABASE_URL PGPASSFILE
  for v in $(compgen -e); do
    case "$v" in
      PG*|STALKEER_DATABASE_*|STALKEER_API_PORT|STALKEER_METRICS_PORT|DB_*|API_PORT|FRONTEND_PORT) unset "$v" ;;
    esac
  done
}

cmd_up() {
  check_prereqs || exit 1

  # Idempotency (D9)
  if instance_running; then
    local differs=0 f_pin
    f_pin="${WT_FRONT_PIN:-${WT_FRONT_ALIAS_PIN:-}}"
    if [ -n "${WT_FRONT_PIN:-}" ] && [ -n "${WT_FRONT_ALIAS_PIN:-}" ] && [ "$WT_FRONT_PIN" != "$WT_FRONT_ALIAS_PIN" ]; then
      err "conflicting frontend ports: FRONTEND_PORT=$WT_FRONT_PIN and FRONT_PORT=$WT_FRONT_ALIAS_PIN (use only one)"
      exit 1
    fi
    [ -n "${WT_API_PIN:-}" ] && [ "$WT_API_PIN" != "$(state_get API_PORT)" ] && differs=1
    [ -n "$f_pin" ] && [ "$f_pin" != "$(state_get FRONT_PORT)" ] && differs=1
    if [ "$differs" = 1 ]; then
      err "an instance is already running on different ports; run 'make wt-down' first"
      print_urls
      exit 1
    fi
    log "instance already running"
    print_urls
    exit 0
  fi

  # Port resolution happens before any side effect (incl. stale cleanup).
  # A stale instance (partial state/containers) is cleaned first so its own
  # ports are not counted as "in use".
  if [ -f "$STATE_FILE" ] || [ -n "$(db_containers)" ] || [ -n "$(marker_pids)" ]; then
    log "cleaning up a stale instance"
    teardown 0 >/dev/null 2>&1
  fi

  resolve_ports || exit 1

  mkdir -p "$STATE_DIR"
  trap on_signal INT TERM
  STARTED=1
  state_set WT_SLUG "$SLUG"
  state_set WT_PROJECT "$PROJECT"
  state_set API_PORT "$API_PORT_R"
  state_set FRONT_PORT "$FRONT_PORT_R"
  state_set API_PINNED "$API_PINNED"
  state_set FRONT_PINNED "$FRONT_PINNED"
  [ -n "$METRICS_PORT_R" ] && state_set METRICS_PORT "$METRICS_PORT_R"

  # 1. Database
  log "starting PostgreSQL ($PROJECT)..."
  WT_PATH="$ROOT" compose up -d --wait --wait-timeout "$READY_TIMEOUT" >"$STATE_DIR/db.log" 2>&1 \
    || fail_up "PostgreSQL failed to become healthy (log: $STATE_DIR/db.log)"
  local dbp
  dbp="$(db_port)"
  [ -n "$dbp" ] || fail_up "could not read the PostgreSQL host port"
  state_set DB_PORT "$dbp"
  state_set DB_HEALTHY 1

  # 2. Backend
  log "building backend..."
  make --no-print-directory build >"$STATE_DIR/build.log" 2>&1 \
    || fail_up "backend build failed (log: $STATE_DIR/build.log)"

  log "starting API on :$API_PORT_R..."
  local apid
  (
    scrub_env
    export DB_HOST=127.0.0.1 DB_PORT="$dbp" DB_USER="$DB_USER_WT" DB_PASSWORD="$DB_PASSWORD_WT" \
      DB_NAME="$DB_NAME_WT" DB_SSLMODE=disable API_PORT="$API_PORT_R" WT_INSTANCE="$SLUG"
    if [ -n "$METRICS_PORT_R" ]; then export STALKEER_METRICS_ENABLED=true STALKEER_METRICS_PORT="$METRICS_PORT_R"; fi
    exec setsid ./bin/stalkeer server >>"$STATE_DIR/api.log" 2>&1
  ) &
  apid=$!
  state_set API_PGID "$apid"
  wait_http "http://localhost:$API_PORT_R/health" "$apid" "API" "$STATE_DIR/api.log" \
    || fail_up "API failed to start"

  # 3. Seed
  if [ "${WT_SEED:-1}" != "0" ]; then
    log "seeding test data..."
    if ! compose exec -T postgres psql -U "$DB_USER_WT" -d "$DB_NAME_WT" -v ON_ERROR_STOP=1 -q \
      <scripts/seed-test-data.sql >"$STATE_DIR/seed.log" 2>&1; then
      fail_up "seeding failed (log: $STATE_DIR/seed.log); retry with WT_SEED=0 to skip it"
    fi
  else
    log "seeding skipped (WT_SEED=0)"
  fi

  # 4. Frontend
  if [ ! -d frontend/node_modules ]; then
    log "installing frontend dependencies (npm ci)..."
    (cd frontend && npm ci) >"$STATE_DIR/npm.log" 2>&1 \
      || fail_up "npm ci failed (log: $STATE_DIR/npm.log)"
  fi
  log "starting frontend on :$FRONT_PORT_R..."
  local fpid
  (
    cd frontend || exit 1
    scrub_env
    export API_PORT="$API_PORT_R" FRONTEND_PORT="$FRONT_PORT_R" WT_INSTANCE="$SLUG"
    exec setsid npm run dev >>"$STATE_DIR/frontend.log" 2>&1
  ) &
  fpid=$!
  state_set FRONT_PGID "$fpid"
  wait_http "http://localhost:$FRONT_PORT_R/" "$fpid" "frontend" "$STATE_DIR/frontend.log" \
    || fail_up "frontend failed to start"

  trap - INT TERM
  STARTED=0
  log "instance ready"
  print_urls
  [ -n "$METRICS_PORT_R" ] && log "metrics:  http://localhost:$METRICS_PORT_R/metrics"
  return 0
}

# --------------------------------------------------------------------- down --

cmd_down() {
  if teardown 0; then
    log "instance stopped and destroyed ($PROJECT)"
  else
    log "nothing running for this worktree"
  fi
}

# ------------------------------------------------------------------- status --

comp() { # name running healthy
  printf '  %-9s %s, %s\n' "$1" "$2" "$3"
}

cmd_status() {
  if [ ! -f "$STATE_FILE" ] && [ -z "$(db_containers)" ] && [ -z "$(marker_pids)" ]; then
    log "no instance running for this worktree"
    return 0
  fi
  local ap fp dp
  ap="$(state_get API_PORT)"; fp="$(state_get FRONT_PORT)"; dp="$(state_get DB_PORT)"
  [ -n "$dp" ] || dp="$(db_port)"
  log "worktree: $ROOT"
  log "instance: $PROJECT"
  echo "  frontend: http://localhost:${fp:-?}"
  echo "  API:      http://localhost:${ap:-?}"
  echo "  database: 127.0.0.1:${dp:-?}"
  local r h
  if [ -n "$(compose ps -q --status running postgres 2>/dev/null)" ]; then r=running; else r=stopped; fi
  if compose exec -T postgres pg_isready -U "$DB_USER_WT" -d "$DB_NAME_WT" >/dev/null 2>&1; then h=healthy; else h=unhealthy; fi
  comp database "$r" "$h"
  if alive "$(state_get API_PGID)"; then r=running; else r=stopped; fi
  if [ -n "$ap" ] && http_ok "http://localhost:$ap/health"; then h=healthy; else h=unhealthy; fi
  comp API "$r" "$h"
  if alive "$(state_get FRONT_PGID)"; then r=running; else r=stopped; fi
  if [ -n "$fp" ] && http_ok "http://localhost:$fp/"; then h=healthy; else h=unhealthy; fi
  comp frontend "$r" "$h"
}

# --------------------------------------------------------------------- logs --

cmd_logs() {
  local files=() f
  for f in "$STATE_DIR/api.log" "$STATE_DIR/frontend.log"; do [ -f "$f" ] && files+=("$f"); done
  if [ ${#files[@]} -eq 0 ]; then
    log "no logs for this worktree (no instance started)"
    return 0
  fi
  if [ "${WT_FOLLOW:-1}" = "0" ]; then
    tail -n 100 "${files[@]}"
  else
    tail -n 100 -F "${files[@]}"
  fi
}

# --------------------------------------------------------------------- plan --
# Resolve identity and ports without starting anything (used by the smoke test).
cmd_plan() {
  resolve_ports || exit 1
  printf 'slug=%s\nproject=%s\napi=%s\nfront=%s\nmetrics=%s\n' \
    "$SLUG" "$PROJECT" "$API_PORT_R" "$FRONT_PORT_R" "$METRICS_PORT_R"
}

case "${1:-}" in
  up) cmd_up ;;
  down) cmd_down ;;
  status) cmd_status ;;
  logs) cmd_logs ;;
  plan) cmd_plan ;;
  *) echo "usage: $0 {up|down|status|logs}" >&2; exit 2 ;;
esac
