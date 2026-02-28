#!/usr/bin/env bash

set -euo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.dev.yml}"
DURATION_SECONDS="${DURATION_SECONDS:-600}"
INTERVAL_SECONDS="${INTERVAL_SECONDS:-60}"
TARGET_NODES="${TARGET_NODES:-25}"
STRICT_NPU="${STRICT_NPU:-1}"
RESULTS_ROOT="${RESULTS_ROOT:-test-results/demo-audit}"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
OUT_DIR="$RESULTS_ROOT/$TIMESTAMP"

mkdir -p "$OUT_DIR"

log() {
  echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$OUT_DIR/demo.log"
}

json_escape() {
  printf '%s' "$1" | sed 's/"/\\"/g'
}

log "Starting 10-minute auditable Docker demo"
log "Compose file: $COMPOSE_FILE"
log "Target nodes: $TARGET_NODES"
log "Duration: ${DURATION_SECONDS}s (interval ${INTERVAL_SECONDS}s)"
log "Strict NPU enforcement: $STRICT_NPU"

if [ ! -f "$COMPOSE_FILE" ]; then
  log "ERROR: compose file not found: $COMPOSE_FILE"
  exit 1
fi

log "Running static + runtime port audit"
node scripts/audit-compose-ports.mjs \
  --out-dir "$OUT_DIR" \
  --target-file "$COMPOSE_FILE" | tee "$OUT_DIR/port-audit-summary.txt"

log "Bringing up baseline stack"
docker compose -f "$COMPOSE_FILE" up -d mongo redis backend prometheus tpm-metrics grafana

log "Burst-scaling node agents with auto certificate issuance"
bash scripts/burst-scale-with-certs.sh "$TARGET_NODES" "$COMPOSE_FILE" | tee "$OUT_DIR/burst-scale.log"

log "Waiting 30s for stabilization"
sleep 30

NODE_AGENT_CID="$(docker compose -f "$COMPOSE_FILE" ps -q node-agent | head -n 1 || true)"
if [ -z "$NODE_AGENT_CID" ]; then
  log "ERROR: no node-agent containers found"
  exit 1
fi

log "Capturing NPU visibility evidence"
{
  echo "=== node-agent env (NPU/CPU related) ==="
  docker exec -i "$NODE_AGENT_CID" sh -lc 'env | sort | grep -E "NPU|ASCEND|CUDA|FORCE_CPU" || true'
  echo
  echo "=== device nodes ==="
  docker exec -i "$NODE_AGENT_CID" sh -lc 'ls -l /dev/davinci0 /dev/davinci_manager 2>/dev/null || true'
  echo
  echo "=== npu-smi probe ==="
  docker exec -i "$NODE_AGENT_CID" sh -lc 'command -v npu-smi >/dev/null && npu-smi info || echo "npu-smi not installed in container"'
  echo
  echo "=== training device logs ==="
  docker compose -f "$COMPOSE_FILE" logs --no-color --tail=400 node-agent | grep -E 'Using NPU|Device=npu|Device=cpu|Initialized' || true
} | tee "$OUT_DIR/npu-evidence.txt"

DEVICE_NPU_COUNT="$(grep -Eic 'Device=npu|Using NPU' "$OUT_DIR/npu-evidence.txt" || true)"
DEVICE_CPU_COUNT="$(grep -Eic 'Device=cpu' "$OUT_DIR/npu-evidence.txt" || true)"

if [ "$STRICT_NPU" = "1" ]; then
  if [ "$DEVICE_NPU_COUNT" -lt 1 ]; then
    log "ERROR: strict mode failed - no NPU training evidence detected in node-agent logs"
    log "Hint: verify Ascend runtime inside image and ensure training emits device logs"
    exit 2
  fi
  if [ "$DEVICE_CPU_COUNT" -gt 0 ]; then
    log "ERROR: strict mode failed - CPU training fallback detected in node-agent logs"
    exit 3
  fi
fi

STEPS="$((DURATION_SECONDS / INTERVAL_SECONDS))"
if [ "$STEPS" -lt 1 ]; then
  STEPS=1
fi

log "Running auditable checks for $STEPS iterations"

for step in $(seq 1 "$STEPS"); do
  ts="$(date +'%Y-%m-%dT%H:%M:%S%z')"
  log "Iteration $step/$STEPS"

  echo "## $ts" >> "$OUT_DIR/health-checks.md"

  backend_code="$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8000/health || true)"
  prom_code="$(curl -s -o /dev/null -w '%{http_code}' http://localhost:9090/-/healthy || true)"
  grafana_code="$(curl -s -o /dev/null -w '%{http_code}' http://localhost:3001/api/health || true)"
  tpm_code="$(curl -s -o /dev/null -w '%{http_code}' http://localhost:9091/metrics || true)"

  {
    echo "- backend /health: $backend_code"
    echo "- prometheus /-/healthy: $prom_code"
    echo "- grafana /api/health: $grafana_code"
    echo "- tpm-metrics /metrics: $tpm_code"
    echo
  } >> "$OUT_DIR/health-checks.md"

  docker compose -f "$COMPOSE_FILE" ps > "$OUT_DIR/compose-ps-$step.txt"

  docker stats --no-stream --format '{{.Name}},{{.CPUPerc}},{{.MemUsage}}' > "$OUT_DIR/docker-stats-$step.csv" || true

  curl -s http://localhost:9090/api/v1/label/__name__/values > "$OUT_DIR/prom-metrics-$step.json" || true

  if [ "$step" -lt "$STEPS" ]; then
    sleep "$INTERVAL_SECONDS"
  fi
done

log "Capturing final artifacts"
docker compose -f "$COMPOSE_FILE" ps > "$OUT_DIR/compose-ps-final.txt"
docker compose -f "$COMPOSE_FILE" logs --no-color --tail=500 node-agent > "$OUT_DIR/node-agent-final.log" || true
docker compose -f "$COMPOSE_FILE" logs --no-color --tail=500 backend > "$OUT_DIR/backend-final.log" || true

cat > "$OUT_DIR/SUMMARY.md" <<EOF
# 10-Minute Auditable Docker Demo Summary

- Timestamp: $TIMESTAMP
- Compose file: $COMPOSE_FILE
- Target node agents: $TARGET_NODES
- Duration: ${DURATION_SECONDS}s
- Interval: ${INTERVAL_SECONDS}s
- Strict NPU enforcement: $STRICT_NPU

## Audit Outputs

- Port audit JSON: compose-port-audit.json
- Port audit Markdown: compose-port-audit.md
- Health checks: health-checks.md
- NPU evidence: npu-evidence.txt
- Burst scaling log: burst-scale.log
- Per-interval stats: docker-stats-*.csv
- Prometheus metric snapshots: prom-metrics-*.json

## NPU/CPU Evidence

- NPU indicator matches (Device=npu / Using NPU): $DEVICE_NPU_COUNT
- CPU indicator matches (Device=cpu): $DEVICE_CPU_COUNT

If NPU count is 0 and CPU count > 0, workloads are likely still CPU-bound. Check host device mapping and runtime toolchain.
EOF

log "Demo complete. Artifacts: $OUT_DIR"
