#!/bin/bash
set -e

TEST_ID="${1:-bft-200-$(date +%Y%m%d-%H%M%S)}"
RESULTS_DIR="test-results/200-node-bft/$(date +%Y-%m-%d)"
COMPOSE_FILE="docker-compose.200nodes.yml"

mkdir -p "$RESULTS_DIR"

echo "🚀 200-Node BFT Test: $TEST_ID"
echo "================================================"

# Pre-flight checks
command -v docker >/dev/null 2>&1 || { echo "❌ Docker required"; exit 1; }
command -v go >/dev/null 2>&1 || { echo "❌ Go required"; exit 1; }

# Cleanup
docker-compose -f "$COMPOSE_FILE" down -v 2>/dev/null || true

# Generate test data
echo "📊 Generating test data..."
go run scripts/generate-test-data.go

# Deploy infrastructure
echo "🔧 Deploying infrastructure..."
docker-compose -f "$COMPOSE_FILE" up -d mongo backend aggregator
sleep 30

# Check backend health
if ! curl -s http://localhost:5000/health >/dev/null; then
    echo "❌ Backend not healthy"
    exit 1
fi

# Deploy 200 nodes
echo "🚀 Deploying 200 node agents..."
docker-compose -f "$COMPOSE_FILE" up -d node-agent
sleep 90

# Verify node count
NODE_COUNT=$(docker ps -q --filter "name=node-agent" | wc -l)
if [ "$NODE_COUNT" -ne 200 ]; then
    echo "❌ Expected 200 nodes, found $NODE_COUNT"
    exit 1
fi
echo "✅ All 200 nodes deployed"

# Run baseline test
echo "🧪 Phase 1: Baseline test (0% Byzantine)..."
go test -v ./internal/consensus/... -run "Test200NodeBFT/BaselineConsensus" -timeout 30m | tee "$RESULTS_DIR/baseline.log"

# Inject Byzantine faults (111 nodes)
echo "🔥 Phase 2: Injecting 111 Byzantine nodes..."
docker ps -q --filter "name=node-agent" | head -111 | while read container; do
    docker exec "$container" sh -c "export BYZANTINE_MODE=true" 2>/dev/null || true
done

# Run BFT test
echo "🛡️ Phase 3: BFT test with 55.5% Byzantine..."
go test -v ./internal/consensus/... -run "Test200NodeBFT/ByzantineFaultTolerance" -timeout 60m | tee "$RESULTS_DIR/byzantine.log"

# Collect results
echo "📈 Collecting results..."
docker stats --no-stream > "$RESULTS_DIR/docker-stats.txt" 2>/dev/null || true
curl -s "http://localhost:9090/api/v1/query?query=consensus_rounds_total" > "$RESULTS_DIR/metrics.json" 2>/dev/null || true

# Generate report
cat > "$RESULTS_DIR/TEST-REPORT.md" << EOL
# 200-Node BFT Test Report

**Test ID:** $TEST_ID  
**Date:** $(date)  
**Configuration:** 200 nodes, 111 Byzantine (55.5%)

## Results
- Baseline: [See baseline.log]
- Byzantine: [See byzantine.log]
- Metrics: [See metrics.json]

## Status
[Review logs for PASS/FAIL]
EOL

echo "✅ Test complete. Results in $RESULTS_DIR"

# Cleanup prompt
read -t 10 -p "Teardown infrastructure? (Y/n) " -n 1 -r || true
echo
if [[ ! ${REPLY:-Y} =~ ^[Nn]$ ]]; then
    docker-compose -f "$COMPOSE_FILE" down -v
    echo "🧹 Infrastructure stopped"
else
    echo "🔌 Infrastructure still running"
    echo "   Grafana: http://localhost:3001 (admin/sovereign2026)"
fi
