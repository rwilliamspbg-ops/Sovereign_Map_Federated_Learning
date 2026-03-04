#!/usr/bin/env bash
set -euo pipefail

# Parse arguments
NODES=100
DURATION="10m"
USE_NPU=false
TPM_TEST=false
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.production.yml}"

while [[ $# -gt 0 ]]; do
  case $1 in
    --nodes)
      NODES="$2"
      shift 2
      ;;
    --duration)
      DURATION="$2"
      shift 2
      ;;
    --use-npu)
      USE_NPU=true
      shift
      ;;
    --tpm-test)
      TPM_TEST=true
      shift
      ;;
    --compose)
      COMPOSE_FILE="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Convert duration to seconds
convert_duration_to_seconds() {
  local duration=$1
  case $duration in
    *m)
      echo $((${duration%m} * 60))
      ;;
    *h)
      echo $((${duration%h} * 3600))
      ;;
    *s)
      echo ${duration%s}
      ;;
    *)
      echo $duration
      ;;
  esac
}

DURATION_SECONDS=$(convert_duration_to_seconds "$DURATION")
INTERVAL_SECONDS=$((DURATION_SECONDS > 300 ? DURATION_SECONDS / 10 : 30))
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
OUT_DIR="test-results/demo-1000nodes/$TIMESTAMP"

mkdir -p "$OUT_DIR"

log() {
  echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$OUT_DIR/demo.log"
}

log "=========================================="
log "Sovereign Map 1000-Node Federated Learning Demo"
log "=========================================="
log "Nodes: $NODES"
log "Duration: $DURATION ($DURATION_SECONDS seconds)"
log "Interval: ${INTERVAL_SECONDS}s"
log "NPU Acceleration: $USE_NPU"
log "TPM Testing: $TPM_TEST"
log "Compose File: $COMPOSE_FILE"
log "Output Directory: $OUT_DIR"
log "=========================================="

# Verify docker-compose file exists
if [ ! -f "$COMPOSE_FILE" ]; then
  log "ERROR: compose file not found: $COMPOSE_FILE"
  exit 1
fi

# Start monitoring stack
log "Starting Prometheus and Grafana monitoring stack..."
docker compose -f "$COMPOSE_FILE" up -d prometheus grafana alertmanager

sleep 5

# Check if services are healthy
log "Checking monitoring stack health..."
PROM_HEALTH=$(curl -s -o /dev/null -w '%{http_code}' http://localhost:9090/-/healthy || echo "000")
GRAFANA_HEALTH=$(curl -s -o /dev/null -w '%{http_code}' http://localhost:3001/api/health || echo "000")

log "Prometheus health: $PROM_HEALTH"
log "Grafana health: $GRAFANA_HEALTH"

if [ "$PROM_HEALTH" != "200" ] || [ "$GRAFANA_HEALTH" != "200" ]; then
  log "WARNING: Monitoring stack may not be fully healthy, continuing anyway..."
fi

# Start backend, database, and redis
log "Starting backend infrastructure (MongoDB, Redis, Backend)..."
docker compose -f "$COMPOSE_FILE" up -d mongo redis backend

sleep 10

# Verify backend is healthy
log "Checking backend health..."
BACKEND_HEALTH=$(curl -s -o /dev/null -w '%{http_code}' http://localhost:8000/health || echo "000")
log "Backend health: $BACKEND_HEALTH"

# Set environment variables for node agents
export NUM_NODES=$NODES
export NUM_ROUNDS=$((DURATION_SECONDS / 60))  # ~1 round per minute
export USE_NPU=$USE_NPU
export TPM_TEST=$TPM_TEST

log "Deploying $NODES federated learning node agents..."
log "Environment:"
log "  NUM_NODES=$NUM_NODES"
log "  NUM_ROUNDS=$NUM_ROUNDS"
log "  USE_NPU=$USE_NPU"
log "  TPM_TEST=$TPM_TEST"

# Create a scaled docker-compose override for node agents
cat > "$OUT_DIR/docker-compose.override.yml" <<'OVERRIDE_EOF'
version: '3.9'
services:
  node-agent:
    deploy:
      replicas: REPLACE_NODES
OVERRIDE_EOF

sed -i "s/REPLACE_NODES/$NODES/" "$OUT_DIR/docker-compose.override.yml"

# Start node agents with scale (note: docker compose doesn't support scale in v3.9)
# Instead, we'll use docker run in a loop or rely on compose with replicas
log "Starting node agents (this may take a few minutes for $NODES nodes)..."

# For production scale, we create independent containers
for i in $(seq 1 $NODES); do
  if [ $((i % 100)) -eq 0 ]; then
    log "Started $i/$NODES node agents..."
  fi
  
  docker run -d \
    --name "node-agent-$i" \
    --network sovereignmap \
    --env NODE_ID="node-$i" \
    --env BACKEND_URL="http://backend:8000" \
    --env USE_NPU="$USE_NPU" \
    --env TPM_TEST="$TPM_TEST" \
    --env NUM_ROUNDS="$NUM_ROUNDS" \
    --restart unless-stopped \
    --health-cmd='curl -f http://localhost:6000/health || exit 1' \
    --health-interval=30s \
    --health-timeout=10s \
    --health-retries=3 \
    sovereignmap/node-agent:latest > /dev/null 2>&1 &
  
  # Throttle container creation to avoid overwhelming Docker daemon
  if [ $((i % 50)) -eq 0 ]; then
    sleep 2
  fi
done

log "All $NODES node agents deployment initiated."
sleep 30

# Count running node agents
RUNNING_AGENTS=$(docker ps --filter "name=node-agent-" --format "{{.Names}}" | wc -l)
log "Currently running node agents: $RUNNING_AGENTS/$NODES"

# Main monitoring loop
log "Starting monitoring loop for $DURATION_SECONDS seconds..."

STEPS=$((DURATION_SECONDS / INTERVAL_SECONDS))
if [ "$STEPS" -lt 1 ]; then
  STEPS=1
fi

log "Running $STEPS monitoring iterations (every ${INTERVAL_SECONDS}s)"

for step in $(seq 1 "$STEPS"); do
  CURRENT_TIME=$(date +'%Y-%m-%dT%H:%M:%S%z')
  log "Monitoring iteration $step/$STEPS at $CURRENT_TIME"
  
  # Collect metrics
  {
    echo "# Iteration $step/$STEPS at $CURRENT_TIME"
    echo ""
    
    # Docker stats
    echo "## Docker Container Statistics"
    docker stats --no-stream --format '{{.Name}},{{.CPUPerc}},{{.MemUsage}},{{.MemPerc}},{{.NetIO}},{{.BlockIO}}' 2>/dev/null | head -20 || true
    echo ""
    
    # Node agent count
    echo "## Node Agent Status"
    RUNNING=$(docker ps --filter "name=node-agent-" --filter "status=running" --format "{{.Names}}" | wc -l)
    EXITED=$(docker ps -a --filter "name=node-agent-" --filter "status=exited" --format "{{.Names}}" | wc -l)
    echo "Running: $RUNNING, Exited: $EXITED"
    echo ""
    
    # Prometheus metrics query
    echo "## Prometheus Active Time Series"
    curl -s "http://localhost:9090/api/v1/query?query=up" 2>/dev/null | grep -o '"value":\[' | wc -l || echo "Unable to query"
    echo ""
    
  } >> "$OUT_DIR/metrics-iteration-$step.txt"
  
  # Export Prometheus data every 5 iterations
  if [ $((step % 5)) -eq 0 ]; then
    log "Exporting Prometheus metrics snapshot..."
    curl -s "http://localhost:9090/api/v1/query_range?query=increase(container_cpu_usage_seconds_total%5B5m%5D)&start=$(date -d '5 minutes ago' +%s)&end=$(date +%s)&step=60s" > "$OUT_DIR/prometheus-range-$step.json" 2>/dev/null || true
  fi
  
  # Sleep before next iteration (except on last iteration)
  if [ "$step" -lt "$STEPS" ]; then
    sleep "$INTERVAL_SECONDS"
  fi
done

log "Main monitoring loop complete. Collecting final metrics..."

# Final data collection
log "Capturing final state..."

{
  echo "# Final Docker Compose State"
  docker compose -f "$COMPOSE_FILE" ps
  echo ""
  echo "# Final Node Agent Container Status"
  docker ps -a --filter "name=node-agent-" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
  echo ""
  echo "# Container Statistics Summary"
  docker stats --no-stream --format '{{.Name}},{{.CPUPerc}},{{.MemUsage}}' 2>/dev/null || true
} > "$OUT_DIR/final-state.txt"

# Export all available Prometheus metrics
log "Exporting final Prometheus metrics..."
curl -s "http://localhost:9090/api/v1/label/__name__/values" > "$OUT_DIR/prometheus-metrics-available.json" 2>/dev/null || true
curl -s "http://localhost:9090/graph" > "$OUT_DIR/prometheus-graph.html" 2>/dev/null || true

# Capture backend logs
log "Capturing backend logs..."
docker logs sovereignmap-backend > "$OUT_DIR/backend-final.log" 2>&1 || true
docker logs sovereignmap-prometheus > "$OUT_DIR/prometheus-final.log" 2>&1 || true

# Create summary report
cat > "$OUT_DIR/DEMO_REPORT.md" <<REPORT_EOF
# 1000-Node Sovereign Map Federated Learning Demo Report

## Execution Summary
- **Timestamp:** $TIMESTAMP
- **Nodes Deployed:** $NODES
- **Duration:** $DURATION ($DURATION_SECONDS seconds)
- **Monitoring Interval:** ${INTERVAL_SECONDS}s
- **Total Iterations:** $STEPS
- **NPU Acceleration:** $USE_NPU
- **TPM Testing:** $TPM_TEST
- **Compose File:** $COMPOSE_FILE

## System Information
- **Start Time:** $(date)
- **Output Directory:** $OUT_DIR

## Monitoring Stack Status
- **Prometheus:** http://localhost:9090
  - Health Code: $PROM_HEALTH
- **Grafana:** http://localhost:3001 (admin/sovereignmap2026)
  - Health Code: $GRAFANA_HEALTH
  - Dashboards: 7 (Overview, Convergence, Performance, Scaling, Security, NPU, GPU)

## Backend Infrastructure
- **Backend API:** http://localhost:8000
  - Health Code: $BACKEND_HEALTH
- **MongoDB:** sovereignmap-mongo:27017 (internal)
- **Redis:** sovereignmap-redis:6379 (internal)

## Node Agent Deployment
- **Final Running Agents:** $RUNNING_AGENTS
- **Expected Agents:** $NODES
- **Configuration:**
  - Node ID pattern: node-1 through node-$NODES
  - Health Check: Every 30s via HTTP /health endpoint
  - Network: sovereignmap (172.28.0.0/16)
  - Restart Policy: unless-stopped

## Data Files Generated
- Metrics iterations: metrics-iteration-[1-$STEPS].txt
- Prometheus range data: prometheus-range-*.json
- Final state snapshot: final-state.txt
- Available metrics list: prometheus-metrics-available.json
- Backend logs: backend-final.log
- Prometheus logs: prometheus-final.log
- Grafana dashboard screenshots: (generate manually from http://localhost:3001)

## Performance Metrics

### Expected Performance (Baseline)
- CPU Training: 1,047 samples/sec
- GPU Expected: 2.5-3.5x speedup = 2,618-3,665 samples/sec
- NPU Expected: 4.0-6.0x speedup = 4,188-6,282 samples/sec

### Scaling Projections (1000 nodes)
- With CPU only: ~18 seconds per round
- With GPU (2.5x): ~7.2 seconds per round
- With NPU (4.0x): ~4.5 seconds per round

## Troubleshooting

### Low Performance
1. Check node agent logs: `docker logs node-agent-1`
2. Verify backend connectivity: `curl http://localhost:8000/health`
3. Check Docker resource limits: `docker stats`
4. Review NPU/GPU device mapping in container

### High Memory Usage
1. Reduce node count: '--nodes 500'
2. Check for memory leaks: `docker stats --no-stream`
3. Cleanup old containers: `docker system prune`

### Grafana Dashboards Not Loading
1. Verify Prometheus is up: `curl http://localhost:9090/-/healthy`
2. Check provisioning mount: `docker exec sovereignmap-grafana ls /etc/grafana/provisioning/dashboards/`
3. Restart Grafana: `docker restart sovereignmap-grafana`

## Monitoring URLs
- **Grafana:** http://localhost:3001
- **Prometheus:** http://localhost:9090
- **Backend API:** http://localhost:8000
- **Alertmanager:** http://localhost:9093

## Next Steps
1. Open Grafana dashboard at http://localhost:3001
2. Select desired dashboard (Overview, Convergence, Performance, Scaling)
3. Monitor metrics in real-time
4. Compare performance across different hardware (CPU, GPU, NPU)
5. Review logs in $OUT_DIR for detailed metrics

---
Generated: $CURRENT_TIME
REPORT_EOF

log "=========================================="
log "Demo complete!"
log "Results directory: $OUT_DIR"
log "Summary report: $OUT_DIR/DEMO_REPORT.md"
log "=========================================="
log ""
log "Monitoring Endpoints:"
log "  Grafana:      http://localhost:3001"
log "  Prometheus:  http://localhost:9090"
log "  Backend:     http://localhost:8000"
log ""
log "To view real-time metrics, open Grafana and select a dashboard."
log "To cleanup, run: docker compose -f $COMPOSE_FILE down"
