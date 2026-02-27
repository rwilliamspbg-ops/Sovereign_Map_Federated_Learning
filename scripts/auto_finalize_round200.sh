#!/usr/bin/env bash
set -euo pipefail

cd /workspaces/Sovereign_Map_Federated_Learning
OUT_DIR="test-results/round200_live"
mkdir -p "$OUT_DIR"

LOG_FILE="$OUT_DIR/finalize_progress.log"
START_TS=$(date -Iseconds)
echo "[$START_TS] finalizer_started" | tee -a "$LOG_FILE"

MAX_WAIT_SEC=$((60*180))
SLEEP_SEC=15
START_EPOCH=$(date +%s)
LAST_ROUND=-1
RESETS=0

while true; do
  NOW=$(date +%s)
  ELAPSED=$((NOW-START_EPOCH))

  ROUND=$(curl -sS http://localhost:8000/convergence | python3 -c 'import sys, json
try:
 d=json.load(sys.stdin)
 print(int(d.get("current_round",0)))
except Exception:
 print(-1)')
  ACC=$(curl -sS http://localhost:8000/convergence | python3 -c 'import sys, json
try:
 d=json.load(sys.stdin)
 print(d.get("current_accuracy", "n/a"))
except Exception:
 print("n/a")')
  LOSS=$(curl -sS http://localhost:8000/convergence | python3 -c 'import sys, json
try:
 d=json.load(sys.stdin)
 print(d.get("current_loss", "n/a"))
except Exception:
 print("n/a")')

  if [[ "$LAST_ROUND" -ge 0 && "$ROUND" -ge 0 && "$ROUND" -lt "$LAST_ROUND" ]]; then
    RESETS=$((RESETS+1))
  fi
  if [[ "$ROUND" -ge 0 ]]; then
    LAST_ROUND=$ROUND
  fi

  echo "[$(date -Iseconds)] progress round=$ROUND acc=$ACC loss=$LOSS resets=$RESETS elapsed_sec=$ELAPSED" | tee -a "$LOG_FILE"

  if [[ "$ROUND" -ge 200 ]]; then
    echo "[$(date -Iseconds)] status=completed" | tee -a "$LOG_FILE"
    break
  fi

  if [[ "$ELAPSED" -ge "$MAX_WAIT_SEC" ]]; then
    echo "[$(date -Iseconds)] status=timeout" | tee -a "$LOG_FILE"
    break
  fi

  sleep "$SLEEP_SEC"
done

# Final snapshots
curl -sS http://localhost:8000/convergence > "$OUT_DIR/convergence_final.json" || true
curl -sS http://localhost:8000/metrics_summary > "$OUT_DIR/metrics_summary_final.json" || true
docker compose -f docker-compose.full.yml logs --tail=800 backend > "$OUT_DIR/backend_tail.log" 2>&1 || true
docker compose -f docker-compose.full.yml ps > "$OUT_DIR/compose_ps_final.txt" 2>&1 || true

python3 - <<'PY'
import json, pathlib, datetime, subprocess
out = pathlib.Path('/workspaces/Sovereign_Map_Federated_Learning/test-results/round200_live')
summary = {
    'finished_at': datetime.datetime.now(datetime.timezone.utc).isoformat(),
    'final_round': None,
    'final_accuracy': None,
    'final_loss': None,
    'backend_restarts': None,
    'nodes_running': None,
}
try:
    c = json.loads((out/'convergence_final.json').read_text())
    summary['final_round'] = c.get('current_round')
    summary['final_accuracy'] = c.get('current_accuracy')
    summary['final_loss'] = c.get('current_loss')
except Exception:
    pass
try:
    br = subprocess.check_output(['docker','inspect','-f','{{.RestartCount}}','sovereign-backend'], text=True).strip()
    summary['backend_restarts'] = int(br)
except Exception:
    pass
try:
    n = subprocess.check_output("docker ps --filter name='sovereign_map_federated_learning-node-agent-' --filter status=running -q | wc -l", shell=True, text=True).strip()
    summary['nodes_running'] = int(n)
except Exception:
    pass
(out/'completion_snapshot.json').write_text(json.dumps(summary, indent=2))
PY

# Commit artifacts (force-add because test-results may be ignored)
git add -f "$OUT_DIR"
if ! git diff --cached --quiet; then
  git commit -m "Add detailed monitored results for 200-round full test"
  echo "[$(date -Iseconds)] commit_created=$(git rev-parse --short HEAD)" | tee -a "$LOG_FILE"
else
  echo "[$(date -Iseconds)] no_new_changes_to_commit" | tee -a "$LOG_FILE"
fi
