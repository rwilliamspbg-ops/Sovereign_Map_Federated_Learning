#!/usr/bin/env bash
set -e
OUT_DIR="$1"
LOG_CSV="$OUT_DIR/live_round_metrics.csv"
EVENT_LOG="$OUT_DIR/live_events.log"

if [ ! -f "$LOG_CSV" ]; then
  echo "timestamp_utc,current_round,current_accuracy,current_loss,rounds_total,fl_round_gauge,node_running,backend_restarts,backend_error_count_15s,alertmanager_status" > "$LOG_CSV"
fi

while true; do
  TS=$(date -u +%Y-%m-%dT%H:%M:%SZ)

  CJSON=$(curl -sS http://localhost:8000/convergence || echo '{}')
  R=$(python3 -c "import json,sys; d=json.loads(sys.stdin.read() or '{}'); print(d.get('current_round',''))" <<< "$CJSON" 2>/dev/null || true)
  A=$(python3 -c "import json,sys; d=json.loads(sys.stdin.read() or '{}'); print(d.get('current_accuracy',''))" <<< "$CJSON" 2>/dev/null || true)
  L=$(python3 -c "import json,sys; d=json.loads(sys.stdin.read() or '{}'); print(d.get('current_loss',''))" <<< "$CJSON" 2>/dev/null || true)

  METR=$(curl -sS http://localhost:8000/metrics 2>/dev/null | awk '/sovereignmap_fl_rounds_total /{rt=$2} /sovereignmap_fl_round /{rg=$2} END{printf "%s,%s", rt, rg}')
  RT=${METR%,*}
  RG=${METR#*,}

  N=$(docker ps --filter name='sovereign_map_federated_learning-node-agent-' --filter status=running -q | wc -l)
  BR=$(docker inspect -f '{{.RestartCount}}' sovereign-backend 2>/dev/null || echo 0)
  BERR=$(docker compose -f docker-compose.full.yml logs --since=15s backend 2>/dev/null | grep -Eic 'error|exception|failed|traceback' || true)
  AM=$(docker ps -a --filter name='sovereign-alertmanager' --format '{{.Status}}' | tr ',' ';')

  echo "$TS,$R,$A,$L,$RT,$RG,$N,$BR,$BERR,$AM" >> "$LOG_CSV"
  echo "[$TS] round=$R acc=$A loss=$L rounds_total=$RT fl_round_gauge=$RG nodes=$N backend_restarts=$BR backend_errs_15s=$BERR alertmanager='$AM'" >> "$EVENT_LOG"

  CUR_ROUND="$R"
  if [ -z "$CUR_ROUND" ]; then
    CUR_ROUND="$RG"
  fi

  if [ -n "$CUR_ROUND" ] && [ "${CUR_ROUND%.*}" -ge 200 ] 2>/dev/null; then
    echo "[$TS] round_200_reached=true" >> "$EVENT_LOG"
    docker compose -f docker-compose.full.yml ps > "$OUT_DIR/compose_ps_end.txt" || true
    curl -sS http://localhost:8000/convergence > "$OUT_DIR/convergence_end.json" || true
    curl -sS http://localhost:8000/metrics > "$OUT_DIR/backend_metrics_end.prom" || true
    curl -sS http://localhost:9090/api/v1/alerts > "$OUT_DIR/prometheus_alerts_end.json" || true
    docker compose -f docker-compose.full.yml logs --tail=500 backend > "$OUT_DIR/backend_logs_end.txt" || true
    exit 0
  fi

  sleep 15
done
