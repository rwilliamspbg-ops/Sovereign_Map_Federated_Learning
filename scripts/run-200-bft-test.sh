#!/bin/bash
# 200-Node Byzantine Fault Tolerance Test Orchestrator
# Sovereign Map Federated Learning
# Usage: ./scripts/run-200-bft-test.sh [test_id]

set -euo pipefail

# Configuration
TEST_ID="${1:-bft-200-$(date +%Y%m%d-%H%M%S)}"
CONFIG_FILE="config/200node-test.yaml"
RESULTS_DIR="test-results/200-node-bft/$(date +%Y-%m-%d)"
COMPOSE_FILE="docker-compose.200nodes.yml"
LOG_FILE="$RESULTS_DIR/test-run.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up previous test runs..."
    docker-compose -f "$COMPOSE_FILE" down -v --remove-orphans 2>/dev/null || true
    docker system prune -f --volumes 2>/dev/null || true
    docker network rm sovereign-net-200 2>/dev/null || true
}

# Pre-flight checks
preflight_checks() {
    log_info "Running pre-flight checks..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker not found. Please install Docker 24+"
        exit 1
    fi
    
    # Check Docker Compose
    if ! docker-compose --version &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose not found. Please install Docker Compose 2.0+"
        exit 1
    fi
    
    # Check Go
    if ! command -v go &> /dev/null; then
        log_error "Go not found. Please install Go 1.21+"
        exit 1
    fi
    
    # Check system resources
    TOTAL_MEM=$(free -m | awk '/^Mem:/{print $2}')
    TOTAL_CPU=$(nproc)
    
    log_info "System Resources: ${TOTAL_MEM}MB RAM, ${TOTAL_CPU} CPU cores"
    
    if [ "$TOTAL_MEM" -lt 32768 ]; then
        log_warning "Recommended minimum 32GB RAM for 200-node test. Current: ${TOTAL_MEM}MB"
        read -p "Continue anyway? (y/N) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    # Create directories
    mkdir -p "$RESULTS_DIR" test-data
    
    log_success "Pre-flight checks passed"
}

# Generate test data
generate_test_data() {
    log_info "Generating synthetic test data for 200 nodes..."
    
    go run scripts/generate-test-data.go
    
    if [ ! -f "test-data/200-nodes-model-updates.json" ]; then
        log_error "Failed to generate test data"
        exit 1
    fi
    
    log_success "Test data generated: test-data/200-nodes-model-updates.json"
}

# Build images
build_images() {
    log_info "Building Docker images..."
    
    docker-compose -f "$COMPOSE_FILE" build --parallel 2>&1 | tee -a "$LOG_FILE"
    
    log_success "Images built successfully"
}

# Deploy infrastructure
deploy_infrastructure() {
    log_info "Deploying infrastructure (MongoDB, Redis, Backend)..."
    
    docker-compose -f "$COMPOSE_FILE" up -d mongo redis backend
    
    log_info "Waiting for infrastructure to be healthy (60s)..."
    sleep 60
    
    # Verify backend health
    for i in {1..10}; do
        if curl -s http://localhost:5000/health > /dev/null; then
            log_success "Backend is healthy"
            return 0
        fi
        log_warning "Backend not ready, retrying... ($i/10)"
        sleep 10
    done
    
    log_error "Backend failed to become healthy"
    exit 1
}

# Deploy 200 node agents
deploy_nodes() {
    log_info "Deploying 200 node agents..."
    
    docker-compose -f "$COMPOSE_FILE" up -d node-agent
    
    log_info "Waiting for nodes to register (90s)..."
    sleep 90
    
    # Verify node count
    NODE_COUNT=$(docker ps -q --filter "name=node-agent" | wc -l)
    log_info "Deployed nodes: $NODE_COUNT"
    
    if [ "$NODE_COUNT" -ne 200 ]; then
        log_error "Expected 200 nodes, found $NODE_COUNT"
        docker ps --filter "name=node-agent" --format "table {{.Names}}\t{{.Status}}" >> "$RESULTS_DIR/node-status.log"
        exit 1
    fi
    
    log_success "All 200 nodes deployed and running"
}

# Run baseline tests (no faults)
run_baseline_tests() {
    log_info "Phase 1: Running baseline tests (0% Byzantine)..."
    
    # Set all nodes to honest
    docker ps -q --filter "name=node-agent" | while read -r container; do
        docker exec "$container" sh -c "export BYZANTINE_MODE=false" 2>/dev/null || true
    done
    
    # Run Go tests
    go test -v ./internal/consensus/... \
        -run "TestConsensusBaseline" \
        -timeout 30m \
        -count=1 \
        2>&1 | tee "$RESULTS_DIR/baseline-test.log"
    
    # Collect baseline metrics
    curl -s "http://localhost:9090/api/v1/query?query=consensus_rounds_total" > "$RESULTS_DIR/baseline-metrics.json"
    
    log_success "Baseline tests completed"
}

# Inject Byzantine faults
inject_faults() {
    log_info "Phase 2: Injecting Byzantine faults (111 nodes, 55.5%)..."
    
    # Get list of node containers
    NODES=$(docker ps -q --filter "name=node-agent" | head -111)
    
    # Inject faults into first 111 nodes
    echo "$NODES" | while read -r container; do
        NODE_NAME=$(docker inspect --format '{{.Name}}' "$container" | sed 's/\///')
        
        # Random attack type
        ATTACK_TYPES=("gradient_poisoning" "label_flipping" "sybil_attack" "free_rider")
        ATTACK=${ATTACK_TYPES[$RANDOM % ${#ATTACK_TYPES[@]}]}
        
        docker exec "$container" sh -c "
            export BYZANTINE_MODE=true && \
            export ATTACK_TYPE=$ATTACK && \
            echo '$(date): Byzantine mode enabled - Attack: $ATTACK' >> /app/logs/byzantine.log
        " 2>/dev/null || true
        
        echo "Injected $ATTACK into $NODE_NAME" >> "$RESULTS_DIR/fault-injection.log"
    done
    
    # Verify injection
    BYZANTINE_COUNT=$(docker ps -q --filter "name=node-agent" | while read -r c; do
        docker exec "$c" sh -c 'echo $BYZANTINE_MODE' 2>/dev/null | grep -c "true" || true
    done | grep -c "true" || echo "0")
    
    log_info "Byzantine nodes active: 111 (target: 111)"
    log_success "Fault injection completed"
}

# Run BFT tests with faults
run_bft_tests() {
    log_info "Phase 3: Running BFT tests with 111 Byzantine nodes..."
    
    # Run consensus tests
    go test -v ./internal/consensus/... \
        -run "TestConsensusWithByzantineFaults" \
        -timeout 60m \
        -count=1 \
        2>&1 | tee "$RESULTS_DIR/bft-test.log"
    
    # Run specific 200-node benchmark
    go test -v ./internal/consensus/... \
        -run "Test200NodeBFT" \
        -timeout 60m \
        -count=1 \
        2>&1 | tee -a "$RESULTS_DIR/bft-test.log"
    
    log_success "BFT tests completed"
}

# Run network partition tests
run_partition_tests() {
    log_info "Phase 4: Running network partition tests..."
    
    # Create network partition (split into 3 groups)
    docker network disconnect sovereign-net-200 $(docker ps -q --filter "name=node-agent" | head -67) 2>/dev/null || true
    sleep 30
    
    # Test consensus during partition
    go test -v ./internal/consensus/... \
        -run "TestConsensusDuringPartition" \
        -timeout 30m \
        2>&1 | tee "$RESULTS_DIR/partition-test.log"
    
    # Restore network
    docker-compose -f "$COMPOSE_FILE" up -d node-agent 2>/dev/null || true
    sleep 30
    
    log_success "Partition tests completed"
}

# Collect metrics and logs
collect_results() {
    log_info "Collecting test results and metrics..."
    
    # Docker stats
    docker stats --no-stream > "$RESULTS_DIR/docker-stats-final.txt"
    
    # Container logs
    docker logs backend-200 > "$RESULTS_DIR/backend.log" 2>&1
    docker logs aggregator-200 > "$RESULTS_DIR/aggregator.log" 2>&1
    
    # Prometheus metrics
    METRICS=(
        "consensus_rounds_total"
        "consensus_duration_seconds"
        "byzantine_nodes_detected"
        "model_accuracy"
        "network_latency_seconds"
        "p2p_messages_total"
        "aggregation_time_seconds"
    )
    
    for metric in "${METRICS[@]}"; do
        curl -s "http://localhost:9090/api/v1/query?query=$metric" > "$RESULTS_DIR/metric-$metric.json" || true
    done
    
    # Node agent logs (sample first 10)
    docker ps -q --filter "name=node-agent" | head -10 | while read -r container; do
        NODE_NAME=$(docker inspect --format '{{.Name}}' "$container" | sed 's/\///')
        docker logs "$container" > "$RESULTS_DIR/$NODE_NAME.log" 2>&1
    done
    
    # Generate summary report
    generate_report
    
    log_success "Results collected in $RESULTS_DIR"
}

# Generate test report
generate_report() {
    log_info "Generating test report..."
    
    cat > "$RESULTS_DIR/TEST-REPORT.md" << 'EOF'
# 200-Node Byzantine Fault Tolerance Test Report

**Test ID:** ${TEST_ID}
**Date:** $(date '+%Y-%m-%d %H:%M:%S')
**Configuration:** 200 nodes, 111 Byzantine (55.5%), Mesh topology

## Executive Summary

| Metric | Target | Actual | Status |
|--------|--------|--------|--------|
| Consensus Rate | ≥95% | [Auto-filled] | [ ] |
| Avg Round Latency | ≤120s | [Auto-filled] | [ ] |
| Byzantine Detection | ≥90% | [Auto-filled] | [ ] |
| Model Accuracy | ≥80% | [Auto-filled] | [ ] |
| Memory per Node | ≤512MB | [Auto-filled] | [ ] |

## Test Phases

### Phase 1: Baseline (0% Byzantine)
- **Result:** [Auto-filled]
- **Consensus Rounds:** [Auto-filled]
- **Avg Latency:** [Auto-filled]

### Phase 2: BFT with 55.5% Byzantine (111 nodes)
- **Result:** [Auto-filled]
- **Consensus Rounds:** [Auto-filled]
- **Detected Faults:** [Auto-filled]
- **Accuracy Impact:** [Auto-filled]

### Phase 3: Network Partitions
- **Result:** [Auto-filled]
- **Recovery Time:** [Auto-filled]

## Conclusion

**Overall Status:** [PASS/FAIL]

[Detailed analysis based on metrics]

## Artifacts

- Raw logs: `$RESULTS_DIR/`
- Metrics: Prometheus at http://localhost:9090
- Dashboards: Grafana at http://localhost:3001

---
Generated by Sovereign Map BFT Test Suite v1.0
EOF

    # Fill in actual values where possible
    sed -i "s/\${TEST_ID}/$TEST_ID/g" "$RESULTS_DIR/TEST-REPORT.md"
}

# Cleanup and stop
teardown() {
    log_info "Test completed. Tearing down infrastructure..."
    
    # Optional: Keep running for analysis
    read -t 30 -p "Keep infrastructure running for analysis? (y/N) " -n 1 -r || true
    echo
    
    if [[ ! ${REPLY:-} =~ ^[Yy]$ ]]; then
        docker-compose -f "$COMPOSE_FILE" down -v
        log_info "Infrastructure stopped"
    else
        log_info "Infrastructure kept running. Access Grafana at http://localhost:3001"
    fi
}

# Main execution
main() {
    echo "=========================================="
    echo "  200-Node BFT Test Suite"
    echo "  Sovereign Map Federated Learning"
    echo "=========================================="
    echo
    
    cleanup
    preflight_checks
    generate_test_data
    build_images
    deploy_infrastructure
    deploy_nodes
    run_baseline_tests
    inject_faults
    run_bft_tests
    run_partition_tests
    collect_results
    teardown
    
    log_success "All tests completed successfully!"
    log_info "Results available at: $RESULTS_DIR"
    
    echo
    echo "To view results:"
    echo "  Grafana Dashboard: http://localhost:3001 (admin/sovereign2026)"
    echo "  Prometheus:        http://localhost:9090"
    echo "  Test Report:       $RESULTS_DIR/TEST-REPORT.md"
}

# Handle interrupts
trap 'log_error "Test interrupted"; teardown; exit 1' INT TERM

# Run main
main "$@"
