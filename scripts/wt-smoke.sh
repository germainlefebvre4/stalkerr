#!/usr/bin/env bash
# wt-smoke.sh - end-to-end smoke test of `make wt-*` in the CURRENT worktree.
# Starts and destroys a real instance (Docker, go build, npm); takes a few minutes.
# Run:  scripts/wt-smoke.sh        (any existing instance of this worktree is destroyed)
set -uo pipefail
cd "$(git rev-parse --show-toplevel)" || exit 1

PASS=0 FAIL=0
ok()   { PASS=$((PASS + 1)); printf '  ok   %s\n' "$1"; }
bad()  { FAIL=$((FAIL + 1)); printf '  FAIL %s\n' "$1"; }
check() { # description command...
  local d="$1"; shift
  if "$@" >/dev/null 2>&1; then ok "$d"; else bad "$d"; fi
}
st() { sed -n "s/^$1=//p" .wt/state.env; }
psql_wt() { docker compose -f docker-compose.wt.yml -p "$(st WT_PROJECT)" exec -T postgres psql -U postgres -d stalkeer -Atc "$1"; }
leftovers() {
  local slug d n=0
  slug="$(scripts/wt.sh plan | sed -n 's/^slug=//p')"
  for d in /proc/[0-9]*; do
    { tr '\0' '\n' <"$d/environ" | grep -qx "WT_INSTANCE=$slug"; } 2>/dev/null && n=$((n + 1))
  done
  echo "$n"
}

echo "== clean slate"
make wt-down >/dev/null 2>&1
[ ! -f config.yml ] && ok "no config.yml in this worktree (instance must not need one)" || echo "  note: config.yml present; 'no config.yml needed' not exercised"

echo "== port forms (make -n)"
check "plain form pins nothing"       bash -c "make -n wt-up | grep -q \"WT_API_PIN='' WT_FRONT_PIN='' WT_FRONT_ALIAS_PIN=''\""
check "FRONT_PORT pins the frontend"  bash -c "make -n wt-up FRONT_PORT=3000 | grep -q \"WT_API_PIN='' .*WT_FRONT_ALIAS_PIN='3000'\""
check "API_PORT pins the API"         bash -c "make -n wt-up API_PORT=8080 | grep -q \"WT_API_PIN='8080'\""
check "both pins"                     bash -c "make -n wt-up FRONT_PORT=3000 API_PORT=8080 | grep -q \"WT_API_PIN='8080' .*WT_FRONT_ALIAS_PIN='3000'\""

echo "== port resolution (no side effects)"
a="$(scripts/wt.sh plan)"; b="$(scripts/wt.sh plan)"
[ "$a" = "$b" ] && ok "same worktree, same ports" || bad "ports not deterministic"
api="$(sed -n 's/^api=//p' <<<"$a")"; front="$(sed -n 's/^front=//p' <<<"$a")"
[ "$api" != "$front" ] && ok "API and frontend ports differ" || bad "API == frontend port"
check "conflicting FRONTEND_PORT/FRONT_PORT rejected" bash -c '! WT_FRONT_PIN=3000 WT_FRONT_ALIAS_PIN=4000 scripts/wt.sh plan'
python3 - "$api" <<'PY' &
import socket, sys, time
s = socket.socket(); s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
s.bind(("127.0.0.1", int(sys.argv[1]))); s.listen(); time.sleep(8)
PY
holder=$!; sleep 1
alt="$(scripts/wt.sh plan | sed -n 's/^api=//p')"
[ -n "$alt" ] && [ "$alt" != "$api" ] && ok "busy preferred port -> another free port" || bad "busy port not avoided"
check "busy pinned port is refused" bash -c "! WT_API_PIN=$api scripts/wt.sh plan"
wait "$holder" 2>/dev/null

echo "== wt-up (default, seeded)"
make wt-up >/tmp/wt-smoke-up.log 2>&1 && ok "wt-up succeeds without config.yml" || { bad "wt-up failed (see /tmp/wt-smoke-up.log)"; cat /tmp/wt-smoke-up.log; }
if [ -f .wt/state.env ]; then
  ap="$(st API_PORT)"; fp="$(st FRONT_PORT)"
  check "API /health answers"             curl -fsS "http://localhost:$ap/health"
  check "frontend answers"                curl -fsS "http://localhost:$fp/"
  check "frontend proxies /api/v1 to API"  curl -fsS "http://localhost:$fp/api/v1/downloads"
  [ "$(psql_wt 'select count(*) from download_info')" -gt 0 ] 2>/dev/null && ok "seed rows present" || bad "no seed rows"
  [ -z "$(git status --porcelain --untracked-files=all -- .wt)" ] && ok "git status clean of .wt" || bad ".wt shows in git status"
  pids_before="$(pgrep -fc "^./bin/stalkeer server" || true)"
  make wt-up >/dev/null 2>&1 && ok "second wt-up is a no-op (exit 0)" || bad "second wt-up failed"
  [ "$(pgrep -fc "^./bin/stalkeer server" || true)" = "$pids_before" ] && ok "still a single API process" || bad "extra API process"
  make wt-up API_PORT=$((ap + 1)) >/dev/null 2>&1 && bad "different pin accepted" || ok "different pin refused while running"
  check "make wt-status reports healthy"   bash -c 'make -s wt-status | grep -q healthy'
  check "make wt-logs prints logs"         env WT_FOLLOW=0 make -s wt-logs
fi

echo "== wt-down"
make wt-down >/dev/null 2>&1 && ok "wt-down succeeds" || bad "wt-down failed"
[ "$(leftovers)" = 0 ] && ok "no process left" || bad "processes left"
[ -z "$(docker ps -aq --filter "label=com.docker.compose.project=$(scripts/wt.sh plan | sed -n 's/^project=//p')")" ] && ok "no container left" || bad "container left"
[ -z "$(docker volume ls -q | grep "^$(scripts/wt.sh plan | sed -n 's/^project=//p')_")" ] && ok "no volume left" || bad "volume left"
[ ! -d .wt ] && ok ".wt removed" || bad ".wt still present"
make wt-down 2>&1 | grep -q "nothing running" && ok "wt-down with nothing running reports it" || bad "no 'nothing running' message"

echo "== wt-up WT_SEED=0, then wt-down without state"
WT_SEED=0 make wt-up >/dev/null 2>&1 && ok "wt-up WT_SEED=0 succeeds" || bad "wt-up WT_SEED=0 failed"
[ -f .wt/state.env ] && [ "$(psql_wt 'select count(*) from download_info')" = 0 ] && ok "no seed rows with WT_SEED=0" || bad "seed rows present with WT_SEED=0"
rm -rf .wt
make wt-down >/dev/null 2>&1
[ "$(leftovers)" = 0 ] && [ -z "$(docker volume ls -q | grep "^$(scripts/wt.sh plan | sed -n 's/^project=//p')_")" ] \
  && ok "wt-down works without .wt/ state" || bad "teardown without state incomplete"

echo "== interruption and metrics"
scripts/wt.sh up >/dev/null 2>&1 &
pid=$!; sleep 4; kill -TERM "$pid" 2>/dev/null; wait "$pid" 2>/dev/null
sleep 1
[ "$(leftovers)" = 0 ] && [ -z "$(docker ps -aq --filter "label=com.docker.compose.project=$(scripts/wt.sh plan | sed -n 's/^project=//p')")" ] \
  && ok "interrupting wt-up (SIGTERM) leaves nothing behind" || bad "interrupted wt-up left resources behind"
make wt-down >/dev/null 2>&1
STALKEER_METRICS_ENABLED=true make wt-up >/dev/null 2>&1 && ok "wt-up with metrics enabled" || bad "wt-up with metrics failed"
[ -f .wt/state.env ] && check "metrics answers on its own port" curl -fsS "http://localhost:$(st METRICS_PORT)/metrics"
make wt-down >/dev/null 2>&1

echo
echo "passed: $PASS  failed: $FAIL"
[ "$FAIL" = 0 ]
