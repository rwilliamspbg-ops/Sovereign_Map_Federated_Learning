#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

TS="$(date +%Y%m%d-%H%M%S)"
OUT_DIR="test-results/bft-scope-200-${TS}"
mkdir -p "$OUT_DIR"

echo "[INFO] Output directory: $OUT_DIR"

export NUM_ROUNDS=200
export MIN_FIT_CLIENTS=1
export MIN_AVAILABLE_CLIENTS=1
export ROUND_TIMEOUT_SECONDS=1200

# Phase schedule: round_threshold:honest:byzantine
PHASES=(
  "1:10:0"
  "41:8:2"
  "81:6:4"
  "121:5:5"
  "161:4:6"
)

phase_idx=0
current_phase="${PHASES[$phase_idx]}"
IFS=':' read -r phase_round honest byz <<< "$current_phase"

apply_scale() {
  local honest_nodes="$1"
  local byz_nodes="$2"
  echo "[INFO] Scaling honest=${honest_nodes}, byzantine=${byz_nodes}"
  docker compose -f docker-compose.full.yml up -d --scale node-agent="$honest_nodes" --scale node-agent-byzantine="$byz_nodes"
}

# Clean start

docker compose -f docker-compose.full.yml down --remove-orphans || true

# Start non-client services

docker compose -f docker-compose.full.yml up -d --build backend prometheus grafana alertmanager

# Initial phase scale
apply_scale "$honest" "$byz"

echo "[INFO] Starting hook capture"

# Hook log
HOOK_LOG="$OUT_DIR/hooks.log"
PHASE_LOG="$OUT_DIR/phase-events.log"

echo "ts,round,accuracy,loss,rounds_total,backend_restarts,node_honest_running,node_byz_running,backend_errors_15s,alertmanager_status" > "$OUT_DIR/hooks.csv"

echo "$(date -Iseconds) phase_start round>=${phase_round} honest=${honest} byz=${byz}" | tee -a "$PHASE_LOG"

START_EPOCH=$(date +%s)
LAST_ROUND=0
RESETS=0
MAX_WAIT_SEC=$((90*60))

while true; do
  now=$(date +%s)
  if (( now - START_EPOCH > MAX_WAIT_SEC )); then
    echo "[WARN] Timeout reached" | tee -a "$PHASE_LOG"
    break
  fi

  conv_json=$(curl -sS http://localhost:8000/convergence || echo '{}')
  round=$(python3 -c 'import json,sys
try:
 d=json.loads(sys.stdin.read() or "{}")
 print(int(d.get("current_round",0)))
except Exception:
 print(0)' <<< "$conv_json")

  acc=$(python3 -c 'import json,sys
try:
 d=json.loads(sys.stdin.read() or "{}")
 print(d.get("current_accuracy",""))
except Exception:
 print("")' <<< "$conv_json")

  loss=$(python3 -c 'import json,sys
try:
 d=json.loads(sys.stdin.read() or "{}")
 print(d.get("current_loss",""))
except Exception:
 print("")' <<< "$conv_json")

  if (( round < LAST_ROUND )); then
    RESETS=$((RESETS+1))
  fi
  LAST_ROUND=$round

  rounds_total=$(curl -sS http://localhost:8000/metrics | awk '/sovereignmap_fl_rounds_total /{print $2}' | tail -n1)
  backend_restarts=$(docker inspect -f '{{.RestartCount}}' sovereign-backend 2>/dev/null || echo 0)
  honest_running=$(docker ps --filter name='sovereign_map_federated_learning-node-agent-' --filter status=running -q | wc -l)
  byz_running=$(docker ps --filter name='sovereign_map_federated_learning-node-agent-byzantine-' --filter status=running -q | wc -l)
  berr=$(docker compose -f docker-compose.full.yml logs --since=15s backend 2>/dev/null | grep -Eic 'error|exception|failed|traceback' || true)
  am_status=$(docker ps -a --filter name='sovereign-alertmanager' --format '{{.Status}}' | tr ',' ';')

  ts_iso=$(date -Iseconds)
  echo "[$ts_iso] round=$round acc=$acc loss=$loss rounds_total=$rounds_total resets=$RESETS backend_restarts=$backend_restarts honest=$honest_running byz=$byz_running backend_errors_15s=$berr alertmanager='$am_status'" | tee -a "$HOOK_LOG"
  echo "$ts_iso,$round,$acc,$loss,${rounds_total:-0},$backend_restarts,$honest_running,$byz_running,$berr,$am_status" >> "$OUT_DIR/hooks.csv"

  # Phase transition
  if (( phase_idx + 1 < ${#PHASES[@]} )); then
    next_phase="${PHASES[$((phase_idx+1))]}"
    IFS=':' read -r next_round next_honest next_byz <<< "$next_phase"
    if (( round >= next_round )); then
      phase_idx=$((phase_idx+1))
      echo "$(date -Iseconds) phase_shift round>=${next_round} honest=${next_honest} byz=${next_byz}" | tee -a "$PHASE_LOG"
      apply_scale "$next_honest" "$next_byz"
    fi
  fi

  if (( round >= 200 )); then
    echo "[INFO] Reached round 200" | tee -a "$PHASE_LOG"
    break
  fi

  sleep 15

done

# Final artifacts
curl -sS http://localhost:8000/convergence > "$OUT_DIR/convergence-final.json" || true
curl -sS http://localhost:8000/metrics_summary > "$OUT_DIR/metrics-summary-final.json" || true
curl -sS http://localhost:8000/metrics > "$OUT_DIR/backend-metrics-final.prom" || true
curl -sS http://localhost:9090/api/v1/rules > "$OUT_DIR/prometheus-rules.json" || true
curl -sS http://localhost:9090/api/v1/alerts > "$OUT_DIR/prometheus-alerts.json" || true
docker compose -f docker-compose.full.yml logs --tail=1500 backend > "$OUT_DIR/backend-tail.log" || true

echo "{\"resets_detected\": $RESETS, \"output_dir\": \"$OUT_DIR\"}" > "$OUT_DIR/summary.json"

echo "$OUT_DIR" > .last_200round_outdir

echo "[INFO] Done. Artifacts: $OUT_DIR"
