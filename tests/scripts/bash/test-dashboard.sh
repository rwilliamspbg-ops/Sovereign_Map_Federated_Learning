#!/usr/bin/env bash
# Real-time test dashboard for incremental scale test

RESULTS_DIR="${1:-.}"
REFRESH_INTERVAL=5

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

clear_screen() {
    clear
}

print_header() {
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║  Sovereign Map - Incremental Scale Test Dashboard              ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_metrics() {
    local convergence_file="$RESULTS_DIR/convergence.log"
    
    if [ ! -f "$convergence_file" ]; then
        echo "⏳ Waiting for test data..."
        return
    fi
    
    echo -e "${YELLOW}📊 CONVERGENCE METRICS${NC}"
    echo "────────────────────────────────────────────────────────────────"
    
    local line_count=$(wc -l < "$convergence_file")
    local latest=$(tail -1 "$convergence_file")
    local oldest=$(head -1 "$convergence_file")
    
    local latest_round=$(echo "$latest" | jq -r '.round // "N/A"')
    local latest_nodes=$(echo "$latest" | jq -r '.nodes // "N/A"')
    local latest_accuracy=$(echo "$latest" | jq -r '.accuracy // "N/A"')
    local latest_loss=$(echo "$latest" | jq -r '.loss // "N/A"')
    
    local oldest_accuracy=$(echo "$oldest" | jq -r '.accuracy // "N/A"')
    
    echo "📈 Current State:"
    echo "  Round:        $latest_round"
    echo "  Active Nodes: $latest_nodes"
    echo "  Accuracy:     ${latest_accuracy}%"
    echo "  Loss:         $latest_loss"
    echo ""
    
    echo "📉 Progression:"
    echo "  Accuracy Improvement: ${oldest_accuracy}% → ${latest_accuracy}%"
    echo "  Data Points: $line_count rounds"
    echo ""
    
    # Extract convergence events
    local convergence_events=$(grep "Convergence.*Scaling" "$RESULTS_DIR/test.log" 2>/dev/null | wc -l || echo 0)
    echo "🎯 Convergence Events: $convergence_events"
    echo ""
}

print_node_scaling() {
    local convergence_file="$RESULTS_DIR/convergence.log"
    
    if [ ! -f "$convergence_file" ]; then
        return
    fi
    
    echo -e "${YELLOW}📊 NODE SCALING HISTORY${NC}"
    echo "────────────────────────────────────────────────────────────────"
    
    # Group by nodes and show first/last occurrence
    jq -r '.nodes' "$convergence_file" | uniq | while read nodes; do
        local first_round=$(jq -r "select(.nodes == \"$nodes\") | .round" "$convergence_file" | head -1)
        local last_round=$(jq -r "select(.nodes == \"$nodes\") | .round" "$convergence_file" | tail -1)
        printf "  %d nodes:  rounds %s - %s\n" "$nodes" "$first_round" "$last_round"
    done
    echo ""
}

print_system_health() {
    echo -e "${YELLOW}💻 SYSTEM HEALTH${NC}"
    echo "────────────────────────────────────────────────────────────────"
    
    # Backend health
    if curl -s http://localhost:8000/health > /dev/null 2>&1; then
        echo -e "  Backend:      ${GREEN}✓ Running${NC}"
    else
        echo -e "  Backend:      ${RED}✗ Unavailable${NC}"
    fi
    
    # Container health
    local running=$(docker compose -f docker-compose.large-scale.yml ps --services --filter "status=running" 2>/dev/null | wc -l || echo 0)
    echo "  Services:     $running running"
    
    # Node agents
    local node_agents=$(docker compose -f docker-compose.large-scale.yml ps node-agent --no-trunc 2>/dev/null | tail -n +2 | wc -l || echo 0)
    echo "  Node Agents:  $node_agents active"
    
    # Memory usage
    local backend_mem=$(docker stats --no-stream sovereignmap-backend --format "{{.MemUsage}}" 2>/dev/null | cut -d' ' -f1 || echo "N/A")
    echo "  Backend RAM:  $backend_mem"
    
    echo ""
}

print_tpm_status() {
    local tpm_file="$RESULTS_DIR/tpm_attestation.json"
    
    if [ ! -f "$tpm_file" ]; then
        echo -e "${YELLOW}🔐 TPM ATTESTATION${NC}"
        echo "────────────────────────────────────────────────────────────────"
        echo "  Status: ⏳ Collecting..."
        echo ""
        return
    fi
    
    echo -e "${YELLOW}🔐 TPM ATTESTATION${NC}"
    echo "────────────────────────────────────────────────────────────────"
    
    local tpm_enabled=$(jq -r '.tpm_enabled' "$tpm_file" 2>/dev/null || echo "unknown")
    local status=$(jq -r '.status' "$tpm_file" 2>/dev/null || echo "unknown")
    
    echo "  Enabled:  $tpm_enabled"
    echo "  Status:   $status"
    echo ""
}

print_npu_status() {
    local npu_file="$RESULTS_DIR/npu_metrics.json"
    
    if [ ! -f "$npu_file" ]; then
        echo -e "${YELLOW}⚡ NPU ACCELERATION${NC}"
        echo "────────────────────────────────────────────────────────────────"
        echo "  Status: ⏳ Collecting..."
        echo ""
        return
    fi
    
    echo -e "${YELLOW}⚡ NPU ACCELERATION${NC}"
    echo "────────────────────────────────────────────────────────────────"
    
    local npu_enabled=$(jq -r '.npu_enabled' "$npu_file" 2>/dev/null || echo "unknown")
    local hardware=$(jq -r '.hardware_info' "$npu_file" 2>/dev/null | head -1 || echo "N/A")
    
    echo "  Enabled:   $npu_enabled"
    echo "  Hardware:  ${hardware:0:30}..."
    echo ""
}

print_recent_events() {
    local log_file="$RESULTS_DIR/test.log"
    
    if [ ! -f "$log_file" ]; then
        return
    fi
    
    echo -e "${YELLOW}📋 RECENT EVENTS${NC}"
    echo "────────────────────────────────────────────────────────────────"
    
    tail -5 "$log_file" | sed 's/^/  /'
    echo ""
}

print_progress_bar() {
    echo -e "${YELLOW}📊 OVERALL PROGRESS${NC}"
    echo "────────────────────────────────────────────────────────────────"
    
    local convergence_file="$RESULTS_DIR/convergence.log"
    
    if [ ! -f "$convergence_file" ]; then
        echo "  ⏳ Test not started"
        return
    fi
    
    local total_rounds=500
    local current_rounds=$(wc -l < "$convergence_file")
    local percentage=$((current_rounds * 100 / total_rounds))
    
    # Simple progress bar
    local bar_length=40
    local filled=$((percentage * bar_length / 100))
    local empty=$((bar_length - filled))
    
    printf "  ["
    printf "%${filled}s" | tr ' ' '█'
    printf "%${empty}s" | tr ' ' '░'
    printf "] %d%% (%d/%d rounds)\n" "$percentage" "$current_rounds" "$total_rounds"
    echo ""
}

print_commands() {
    echo -e "${BLUE}⌨️  COMMANDS${NC}"
    echo "────────────────────────────────────────────────────────────────"
    echo "  View detailed report:     cat $RESULTS_DIR/TEST_REPORT.md"
    echo "  View all convergence:     jq . $RESULTS_DIR/convergence.log"
    echo "  View test log:            tail -f $RESULTS_DIR/test.log"
    echo "  View docker compose logs: docker compose -f docker-compose.large-scale.yml logs -f backend"
    echo "  Stop test:                Press Ctrl+C"
    echo ""
}

print_footer() {
    echo -e "${BLUE}─────────────────────────────────────────────────────────────────${NC}"
    echo "Last updated: $(date '+%Y-%m-%d %H:%M:%S') | Refresh every ${REFRESH_INTERVAL}s | Press Ctrl+C to exit"
}

# Main dashboard loop
main() {
    # Find the most recent test if directory not specified
    if [ ! -d "$RESULTS_DIR" ] || [ "$RESULTS_DIR" = "." ]; then
        RESULTS_DIR=$(find test-results -maxdepth 1 -type d -name "incremental_scale_test_*" | sort -r | head -1)
        if [ -z "$RESULTS_DIR" ]; then
            echo "❌ No test results found"
            exit 1
        fi
    fi
    
    echo "📊 Monitoring: $RESULTS_DIR"
    echo "Press any key to start or Ctrl+C to exit..."
    read -n 1
    
    while true; do
        clear_screen
        print_header
        print_metrics
        print_node_scaling
        print_system_health
        print_tpm_status
        print_npu_status
        print_progress_bar
        print_recent_events
        print_commands
        print_footer
        
        sleep "$REFRESH_INTERVAL"
    done
}

# Handle Ctrl+C
trap 'clear_screen; echo "Dashboard closed"; exit 0' INT

main "$@"
