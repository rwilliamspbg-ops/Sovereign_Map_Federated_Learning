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
DEFAULT_NODE_COUNT=3
MIN_NODES="${MIN_NODES:-$DEFAULT_NODE_COUNT}"
TARGET_NODES="${TARGET_NODES:-$MIN_NODES}"
COMPOSE_FILE="docker-compose.production.yml"
COMPOSE_ENV_FILE="${COMPOSE_ENV_FILE:-.env.production}"
COMPOSE_ENV_ARGS=()
if [ -n "$COMPOSE_ENV_FILE" ] && [ -f "$COMPOSE_ENV_FILE" ]; then
    COMPOSE_ENV_ARGS=(--env-file "$COMPOSE_ENV_FILE")
fi

compose_cmd() {
    docker compose "${COMPOSE_ENV_ARGS[@]}" -f "$COMPOSE_FILE" "$@"
}

get_profile_var() {
    local key="$1"
    local fallback="$2"
    local value=""
    if [ -n "$COMPOSE_ENV_FILE" ] && [ -f "$COMPOSE_ENV_FILE" ]; then
        value=$(grep -E "^${key}=" "$COMPOSE_ENV_FILE" | tail -n1 | cut -d '=' -f2-)
    fi
    echo "${value:-$fallback}"
}

BACKEND_API_PORT=$(get_profile_var "BACKEND_API_HOST_PORT" "8000")
BACKEND_GRPC_PORT=$(get_profile_var "BACKEND_GRPC_HOST_PORT" "8080")
FRONTEND_PORT=$(get_profile_var "FRONTEND_HOST_PORT" "3000")
GRAFANA_PORT=$(get_profile_var "GRAFANA_HOST_PORT" "3001")
PROMETHEUS_PORT=$(get_profile_var "PROMETHEUS_HOST_PORT" "9090")
ALERTMANAGER_PORT=$(get_profile_var "ALERTMANAGER_HOST_PORT" "9093")

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
    local underline
    underline=${1//?/=}
    echo -e "${MAGENTA}${underline}${NC}" | tee -a "$LOG_FILE"
}

get_node_services() {
    compose_cmd config --services | grep '^node-agent' || true
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
    local ports=("$BACKEND_API_PORT" "$BACKEND_GRPC_PORT" "$PROMETHEUS_PORT" "$ALERTMANAGER_PORT" "$FRONTEND_PORT" "$GRAFANA_PORT")
    for port in "${ports[@]}"; do
        if lsof -Pi :"$port" -sTCP:LISTEN -t >/dev/null 2>&1; then
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
    if ! compose_cmd down --remove-orphans 2>&1 | tee -a "$LOG_FILE"; then
        log_warn "Cleanup encountered stale/absent containers; continuing with fresh startup"
    fi

    # Remove any container tracked by this Compose project label.
    # This is more reliable than hard-coded names and catches scaled services.
    local project_name
    local project_container_ids
    project_name=$(compose_cmd config 2>/dev/null | awk '/^name: / {print $2; exit}')
    project_name=${project_name:-sovereign_map_federated_learning}

    log_info "Removing stale project containers for '$project_name'..."
    project_container_ids=$(docker ps -aq --filter "label=com.docker.compose.project=$project_name")
    if [ -n "$project_container_ids" ]; then
        echo "$project_container_ids" | xargs docker rm -f >/dev/null 2>&1 || true
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
    compose_cmd pull mongo redis prometheus grafana alertmanager 2>&1 | tee -a "$LOG_FILE"
    log_success "Images updated"
}

##############################################################################
# Launch Monitoring Stack
##############################################################################

launch_monitoring() {
    log_header "📊 LAUNCHING MONITORING STACK"

    # Ensure Compose metadata is clean for monitoring services
    compose_cmd rm -fsv prometheus grafana alertmanager >/dev/null 2>&1 || true
    
    log_info "Starting Prometheus, Grafana, and Alertmanager..."
    compose_cmd up -d --no-deps --no-recreate prometheus grafana alertmanager 2>&1 | tee -a "$LOG_FILE"
    
    # Wait for services to be ready
    log_info "Waiting for monitoring services to initialize..."
    sleep 10
    
    # Check Prometheus
    if curl -s "http://localhost:${PROMETHEUS_PORT}/-/healthy" > /dev/null 2>&1; then
        log_success "Prometheus is healthy (http://localhost:${PROMETHEUS_PORT})"
    else
        log_warn "Prometheus health check failed"
    fi
    
    # Check Grafana
    if curl -s "http://localhost:${GRAFANA_PORT}/api/health" > /dev/null 2>&1; then
        log_success "Grafana is healthy (http://localhost:${GRAFANA_PORT})"
        log_info "Grafana credentials: admin/${GRAFANA_ADMIN_PASSWORD:-CHANGE_ME_GRAFANA}"
    else
        log_warn "Grafana health check failed"
    fi
    
    # Check Alertmanager
    if curl -s "http://localhost:${ALERTMANAGER_PORT}/-/healthy" > /dev/null 2>&1; then
        log_success "Alertmanager is healthy (http://localhost:${ALERTMANAGER_PORT})"
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

    local available_nodes
    local desired_nodes
    local node_services

    mapfile -t node_services < <(get_node_services)
    available_nodes=${#node_services[@]}
    if [ "$available_nodes" -eq 0 ]; then
        log_error "No node-agent services found in docker-compose.production.yml"
        return 1
    fi

    desired_nodes=$MIN_NODES
    if [ "$desired_nodes" -gt "$available_nodes" ]; then
        log_warn "Requested MIN_NODES=$desired_nodes but only $available_nodes node services are defined; using $available_nodes"
        desired_nodes=$available_nodes
    fi

    # Clear stale compose state for core services before startup
    compose_cmd rm -fsv backend frontend mongo redis "${node_services[@]}" >/dev/null 2>&1 || true
    
    log_info "Starting Sovereign Map backend..."
    compose_cmd up -d --build --no-recreate backend 2>&1 | tee -a "$LOG_FILE"
    
    sleep 5
    
    log_info "Deploying initial node set ($desired_nodes nodes)..."
    compose_cmd up -d --build --no-recreate mongo redis backend frontend "${node_services[@]:0:$desired_nodes}" 2>&1 | tee -a "$LOG_FILE"
    
    log_success "Initial nodes deployed"
    
    # Wait for network stabilization
    log_info "Waiting for network stabilization (30 seconds)..."
    for i in {30..1}; do
        printf "\r  ⏳ %s seconds remaining..." "$i"
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
    if curl -s "http://localhost:${BACKEND_API_PORT}/health" > /dev/null 2>&1; then
        log_success "Backend API is healthy"
    else
        log_error "Backend API is not responding"
        return 1
    fi
    
    # Check node count
    local active_nodes
    active_nodes=$(docker ps --filter "name=node-agent" --format "{{.Names}}" | wc -l)
    log_info "Active nodes: $active_nodes"

    if [ "$active_nodes" -gt 0 ]; then
        log_success "Node agents are running"
    else
        log_warn "No node agents are running"
    fi
    
    # Check FL metrics
    if curl -s "http://localhost:${BACKEND_API_PORT}/convergence" > /dev/null 2>&1; then
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
    echo -e "  ${CYAN}🌐 Grafana Dashboard:${NC}      http://localhost:${GRAFANA_PORT}"
    echo -e "     ${MAGENTA}↳${NC} Genesis Overview:      http://localhost:${GRAFANA_PORT}/d/genesis-launch-overview"
    echo -e "     ${MAGENTA}↳${NC} Network Performance:   http://localhost:${GRAFANA_PORT}/d/network-performance-health"
    echo -e "     ${MAGENTA}↳${NC} Consensus & Trust:     http://localhost:${GRAFANA_PORT}/d/consensus-trust-monitoring"
    echo ""
    echo -e "  ${CYAN}📈 Prometheus Metrics:${NC}     http://localhost:${PROMETHEUS_PORT}"
    echo -e "  ${CYAN}🔔 Alert Manager:${NC}          http://localhost:${ALERTMANAGER_PORT}"
    echo -e "  ${CYAN}🔌 Backend API:${NC}            http://localhost:${BACKEND_API_PORT}"
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}🔧 MANAGEMENT COMMANDS:${NC}"
    echo ""
    echo -e "  View logs:         ${GREEN}docker compose logs -f${NC}"
    echo -e "  Start node agents: ${GREEN}docker compose ${COMPOSE_ENV_ARGS[*]} -f $COMPOSE_FILE up -d node-agent-1 node-agent-2 node-agent-3${NC}"
    echo -e "  Stop network:      ${GREEN}docker compose ${COMPOSE_ENV_ARGS[*]} -f $COMPOSE_FILE down --remove-orphans${NC}"
    echo -e "  Restart services:  ${GREEN}docker compose ${COMPOSE_ENV_ARGS[*]} -f $COMPOSE_FILE restart${NC}"
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
        if curl -s "http://localhost:${BACKEND_API_PORT}/convergence" > /tmp/metrics.json 2>/dev/null; then
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

CMD="${1:-}"

if [ "$CMD" == "monitor" ]; then
    start_monitoring
elif [ "$CMD" == "status" ]; then
    monitor_health
elif [ "$CMD" == "dashboard" ]; then
    display_dashboard
    echo "Press Enter to exit..."
    read -r
else
    main
fi
