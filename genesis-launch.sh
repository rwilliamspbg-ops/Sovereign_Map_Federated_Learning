#!/bin/bash

##############################################################################
# 🚀 SOVEREIGN MAP - GENESIS BLOCK LAUNCH
# Professional Launch Script for Production Deployment
# Copyright 2026 Sovereign-Mohawk Core Team
##############################################################################

set -euo pipefail

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
LAUNCH_TIME=$(date +%s)
LAUNCH_DATE=$(date '+%Y-%m-%d %H:%M:%S')
LOG_FILE="genesis-launch-${LAUNCH_TIME}.log"
NETWORK_NAME="sovereign-genesis"
MIN_NODES=20
TARGET_NODES=100

##############################################################################
# Logging Functions
##############################################################################

log() {
    echo -e "${GREEN}[$(date '+%H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

log_info() {
    echo -e "${BLUE}ℹ  [INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_warn() {
    echo -e "${YELLOW}⚠  [WARN]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}✗  [ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}✓  [SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

log_header() {
    echo "" | tee -a "$LOG_FILE"
    echo -e "${MAGENTA}$1${NC}" | tee -a "$LOG_FILE"
    echo -e "${MAGENTA}$(echo "$1" | sed 's/./=/g')${NC}" | tee -a "$LOG_FILE"
}

##############################################################################
# Pre-Launch Checks
##############################################################################

pre_launch_checks() {
    log_header "🔍 PRE-LAUNCH VALIDATION"
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker not found. Please install Docker first."
        exit 1
    fi
    log_success "Docker installed: $(docker --version | head -1)"
    
    # Check Docker Compose
    if ! docker compose version >/dev/null 2>&1; then
        log_error "Docker Compose not found. Please install Docker Compose first."
        exit 1
    fi
    log_success "Docker Compose installed: $(docker compose version | head -1)"
    
    # Check required files
    local required_files=(
        "docker-compose.production.yml"
        "prometheus.yml"
        "alertmanager.yml"
        "grafana/dashboards/genesis-launch-overview.json"
    )
    
    for file in "${required_files[@]}"; do
        if [ ! -f "$file" ]; then
            log_error "Required file not found: $file"
            exit 1
        fi
    done
    log_success "All required files present"
    
    # Check available resources
    local total_ram="N/A"
    if command -v free >/dev/null 2>&1; then
        total_ram="$(free -h | awk '/^Mem:/ {print $2}')"
    elif [ -r /proc/meminfo ]; then
        total_ram="$(awk '/MemTotal:/ {printf "%.1fG", $2/1024/1024}' /proc/meminfo)"
    fi

    log_info "System Resources:"
    echo "  - CPU Cores: $(nproc)" | tee -a "$LOG_FILE"
    echo "  - Total RAM: $total_ram" | tee -a "$LOG_FILE"
    echo "  - Disk Space: $(df -h . | awk 'NR==2 {print $4}') available" | tee -a "$LOG_FILE"
    
    # Check ports
    local ports=(8000 8080 9090 9093 3000 3001)
    for port in "${ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            log_warn "Port $port is already in use"
        else
            log_success "Port $port available"
        fi
    done
}

##############################################################################
# Network Initialization
##############################################################################

initialize_network() {
    log_header "🌐 NETWORK INITIALIZATION"

    # Remove any stale containers left from a previous incomplete run
    log_info "Cleaning up any stale containers..."
    if ! docker compose -f docker-compose.production.yml down --remove-orphans 2>&1 | tee -a "$LOG_FILE"; then
        log_warn "Cleanup encountered stale/absent containers; continuing with fresh startup"
    fi

    # Create Docker network
    if docker network inspect "$NETWORK_NAME" >/dev/null 2>&1; then
        log_info "Network '$NETWORK_NAME' already exists"
    else
        docker network create "$NETWORK_NAME" 2>&1 | tee -a "$LOG_FILE"
        log_success "Network '$NETWORK_NAME' created"
    fi
    
    # Pull latest images
    log_info "Pulling latest Docker images..."
    docker compose -f docker-compose.production.yml pull mongo redis prometheus grafana alertmanager 2>&1 | tee -a "$LOG_FILE"
    log_success "Images updated"
}

##############################################################################
# Launch Monitoring Stack
##############################################################################

launch_monitoring() {
    log_header "📊 LAUNCHING MONITORING STACK"
    
    log_info "Starting Prometheus, Grafana, and Alertmanager..."
    docker compose -f docker-compose.production.yml up -d --no-deps prometheus grafana alertmanager 2>&1 | tee -a "$LOG_FILE"
    
    # Wait for services to be ready
    log_info "Waiting for monitoring services to initialize..."
    sleep 10
    
    # Check Prometheus
    if curl -s http://localhost:9090/-/healthy > /dev/null 2>&1; then
        log_success "Prometheus is healthy (http://localhost:9090)"
    else
        log_warn "Prometheus health check failed"
    fi
    
    # Check Grafana
    if curl -s http://localhost:3001/api/health > /dev/null 2>&1; then
        log_success "Grafana is healthy (http://localhost:3001)"
        log_info "Grafana credentials: admin/${GRAFANA_ADMIN_PASSWORD:-CHANGE_ME_GRAFANA}"
    else
        log_warn "Grafana health check failed"
    fi
    
    # Check Alertmanager
    if curl -s http://localhost:9093/-/healthy > /dev/null 2>&1; then
        log_success "Alertmanager is healthy (http://localhost:9093)"
    else
        log_warn "Alertmanager health check failed"
    fi
}

##############################################################################
# Genesis Block Creation
##############################################################################

create_genesis_block() {
    log_header "🎯 CREATING GENESIS BLOCK"
    
    log_info "Initializing Genesis block parameters..."
    
    # Create genesis configuration
    cat > /tmp/genesis-config.json <<EOF
{
  "genesis_time": "$LAUNCH_DATE",
  "chain_id": "sovereign-mainnet",
  "initial_nodes": $MIN_NODES,
  "consensus_mechanism": "BFT",
  "min_trust_score": 75,
  "target_accuracy": 0.85,
  "max_byzantine_tolerance": 0.33
}
EOF
    
    log_success "Genesis configuration created"
    cat /tmp/genesis-config.json | tee -a "$LOG_FILE"
}

##############################################################################
# Launch Node Network
##############################################################################

launch_network() {
    log_header "🚀 LAUNCHING NODE NETWORK"
    
    log_info "Starting Sovereign Map backend..."
    docker compose -f docker-compose.production.yml up -d --build backend 2>&1 | tee -a "$LOG_FILE"
    
    sleep 5
    
    log_info "Deploying initial node set ($MIN_NODES nodes)..."
    docker compose -f docker-compose.production.yml up -d --build --scale node-agent=$MIN_NODES mongo redis backend frontend node-agent 2>&1 | tee -a "$LOG_FILE"
    
    log_success "Initial nodes deployed"
    
    # Wait for network stabilization
    log_info "Waiting for network stabilization (30 seconds)..."
    for i in {30..1}; do
        printf "\r  ⏳ $i seconds remaining..."
        sleep 1
    done
    echo ""
    
    log_success "Network stabilized"
}

##############################################################################
# Health Monitoring
##############################################################################

monitor_health() {
    log_header "💊 HEALTH STATUS CHECK"
    
    # Check backend
    if curl -s http://localhost:8000/health > /dev/null 2>&1; then
        log_success "Backend API is healthy"
    else
        log_error "Backend API is not responding"
        return 1
    fi
    
    # Check node count
    local active_nodes
    active_nodes=$(docker ps --filter "name=node-agent" --format "{{.Names}}" | wc -l)
    log_info "Active nodes: $active_nodes / $MIN_NODES"
    
    if [ "$active_nodes" -ge "$MIN_NODES" ]; then
        log_success "Minimum node count reached"
    else
        log_warn "Node count below minimum threshold"
    fi
    
    # Check FL metrics
    if curl -s http://localhost:8000/convergence > /dev/null 2>&1; then
        log_success "Federated learning metrics available"
    else
        log_warn "FL metrics not yet available"
    fi
}

##############################################################################
# Launch Dashboard
##############################################################################

display_dashboard() {
    log_header "📺 GENESIS LAUNCH DASHBOARD"
    
    echo ""
    echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║                                                              ║${NC}"
    echo -e "${CYAN}║             🚀 SOVEREIGN MAP - GENESIS LAUNCH 🚀             ║${NC}"
    echo -e "${CYAN}║                                                              ║${NC}"
    echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${GREEN}Launch Time:${NC}      $LAUNCH_DATE"
    echo -e "${GREEN}Network ID:${NC}       $NETWORK_NAME"
    echo -e "${GREEN}Initial Nodes:${NC}    $MIN_NODES"
    echo -e "${GREEN}Target Nodes:${NC}     $TARGET_NODES"
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}📊 MONITORING INTERFACES:${NC}"
    echo ""
    echo -e "  ${CYAN}🌐 Grafana Dashboard:${NC}      http://localhost:3001"
    echo -e "     ${MAGENTA}↳${NC} Genesis Overview:      http://localhost:3001/d/genesis-launch-overview"
    echo -e "     ${MAGENTA}↳${NC} Network Performance:   http://localhost:3001/d/network-performance-health"
    echo -e "     ${MAGENTA}↳${NC} Consensus & Trust:     http://localhost:3001/d/consensus-trust-monitoring"
    echo ""
    echo -e "  ${CYAN}📈 Prometheus Metrics:${NC}     http://localhost:9090"
    echo -e "  ${CYAN}🔔 Alert Manager:${NC}          http://localhost:9093"
    echo -e "  ${CYAN}🔌 Backend API:${NC}            http://localhost:8000"
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}🔧 MANAGEMENT COMMANDS:${NC}"
    echo ""
    echo -e "  View logs:         ${GREEN}docker compose logs -f${NC}"
    echo -e "  Scale nodes:       ${GREEN}docker compose -f docker-compose.production.yml up -d --scale node-agent=50${NC}"
    echo -e "  Stop network:      ${GREEN}docker compose -f docker-compose.production.yml down --remove-orphans${NC}"
    echo -e "  Restart services:  ${GREEN}docker compose -f docker-compose.production.yml restart${NC}"
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
}

##############################################################################
# Post-Launch Monitoring
##############################################################################

start_monitoring() {
    log_header "📡 REAL-TIME MONITORING"
    
    echo -e "${YELLOW}Monitoring Genesis launch... (Press Ctrl+C to exit)${NC}"
    echo ""
    
    while true; do
        clear
        display_dashboard
        
        echo -e "${GREEN}📊 Current Metrics:${NC}"
        echo ""
        
        # Get metrics from backend
        if curl -s http://localhost:8000/convergence > /tmp/metrics.json 2>/dev/null; then
            local accuracy
            local loss
            local round
            local nodes
            accuracy=$(jq -r '.current_accuracy // "N/A"' /tmp/metrics.json 2>/dev/null || echo "N/A")
            loss=$(jq -r '.current_loss // "N/A"' /tmp/metrics.json 2>/dev/null || echo "N/A")
            round=$(jq -r '.current_round // "N/A"' /tmp/metrics.json 2>/dev/null || echo "N/A")
            nodes=$(docker ps --filter "name=node-agent" --format "{{.Names}}" | wc -l)
            
            echo -e "  ${CYAN}Round:${NC}          $round"
            echo -e "  ${CYAN}Accuracy:${NC}       $accuracy"
            echo -e "  ${CYAN}Loss:${NC}           $loss"
            echo -e "  ${CYAN}Active Nodes:${NC}   $nodes"
        else
            echo -e "  ${RED}⚠ Metrics not available yet${NC}"
        fi
        
        echo ""
        echo -e "${YELLOW}Last updated: $(date '+%H:%M:%S')${NC}"
        echo -e "${YELLOW}Log file: $LOG_FILE${NC}"
        
        sleep 5
    done
}

##############################################################################
# Main Launch Sequence
##############################################################################

main() {
    clear
    
    echo -e "${MAGENTA}"
    cat << "EOF"
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║     ███████╗ ██████╗ ██╗   ██╗███████╗██████╗ ███████╗██╗    ║
║     ██╔════╝██╔═══██╗██║   ██║██╔════╝██╔══██╗██╔════╝██║    ║
║     ███████╗██║   ██║██║   ██║█████╗  ██████╔╝█████╗  ██║    ║
║     ╚════██║██║   ██║╚██╗ ██╔╝██╔══╝  ██╔══██╗██╔══╝  ██║    ║
║     ███████║╚██████╔╝ ╚████╔╝ ███████╗██║  ██║███████╗██║    ║
║     ╚══════╝ ╚═════╝   ╚═══╝  ╚══════╝╚═╝  ╚═╝╚══════╝╚═╝    ║
║                                                               ║
║                  GENESIS BLOCK LAUNCH SEQUENCE                ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
    
    log "Starting Genesis Block Launch Sequence..."
    log "Log file: $LOG_FILE"
    echo ""
    
    # Execute launch sequence
    pre_launch_checks
    initialize_network
    launch_monitoring
    create_genesis_block
    launch_network
    monitor_health
    
    echo ""
    log_success "🎉 GENESIS BLOCK LAUNCH COMPLETE!"
    echo ""
    
    display_dashboard
    
    # Prompt for monitoring
    echo ""
    read -p "Start real-time monitoring? (y/n) " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        start_monitoring
    else
        log_info "Launch complete. View dashboards at the URLs above."
        log_info "Run './genesis-launch.sh monitor' to start monitoring later."
    fi
}

##############################################################################
# Command Line Interface
##############################################################################

if [ "$1" == "monitor" ]; then
    start_monitoring
elif [ "$1" == "status" ]; then
    monitor_health
elif [ "$1" == "dashboard" ]; then
    display_dashboard
    echo "Press Enter to exit..."
    read
else
    main
fi
