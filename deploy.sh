#!/usr/bin/env bash
# Sovereign Map Docker Quick Start Launcher
# Usage: bash deploy.sh [nodes]

set -e

NODE_SCALE="${1:-5}"
PROFILE="full"

# Color output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════╗"
echo "║    Sovereign Map Federated Learning - Docker Deploy    ║"
echo "╚════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Function to display menu
show_menu() {
    echo ""
    echo -e "${YELLOW}Sovereign Map standard runtime:${NC}"
    echo "  docker compose -f docker-compose.full.yml up -d --scale node-agent=<nodes>"
}

if ! [[ "$NODE_SCALE" =~ ^[0-9]+$ ]] || [ "$NODE_SCALE" -lt 1 ]; then
    echo -e "${RED}✗ Node scale must be a positive integer${NC}"
    show_menu
    exit 1
fi

COMPOSE_FILE="docker-compose.full.yml"
DEFAULT_ENV_FILE=".env.full"
SCALE="$NODE_SCALE"
MEMORY="3-6GB"
DESC="Full Stack Environment"

COMPOSE_ENV_FILE="${COMPOSE_ENV_FILE:-$DEFAULT_ENV_FILE}"
COMPOSE_ENV_ARGS=()
if [ -n "$COMPOSE_ENV_FILE" ] && [ -f "$COMPOSE_ENV_FILE" ]; then
    COMPOSE_ENV_ARGS=(--env-file "$COMPOSE_ENV_FILE")
fi

# Default launch behavior: auto-detect and prefer hardware acceleration.
# Respect explicit overrides from environment or env files.
if [ -z "${NODE_AGENT_FORCE_CPU:-}" ]; then
    NODE_AGENT_FORCE_CPU="false"
fi
if [ -z "${NPU_ENABLED:-}" ]; then
    NPU_ENABLED="true"
fi
if [ -z "${XPU_ENABLED:-}" ]; then
    XPU_ENABLED="true"
fi
if [ -z "${GPU_ENABLED:-}" ]; then
    GPU_ENABLED="true"
fi

export NODE_AGENT_FORCE_CPU
export NPU_ENABLED
export XPU_ENABLED
export GPU_ENABLED

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

echo -e "${GREEN}Profile: $DESC${NC}"
echo "  Nodes: $SCALE"
echo "  Memory: $MEMORY"
echo "  Compose: $COMPOSE_FILE"
if [ -n "$COMPOSE_ENV_FILE" ] && [ -f "$COMPOSE_ENV_FILE" ]; then
    echo "  Env file: $COMPOSE_ENV_FILE"
else
    echo "  Env file: none (using compose defaults)"
fi
echo "  Node agent FORCE_CPU: $NODE_AGENT_FORCE_CPU"
echo "  NPU_ENABLED: $NPU_ENABLED"
echo "  XPU_ENABLED: $XPU_ENABLED"
echo "  GPU_ENABLED: $GPU_ENABLED"
echo ""

# Pre-flight checks
echo -e "${YELLOW}[1/4] Pre-flight checks...${NC}"

if ! command -v docker &> /dev/null; then
    echo -e "${RED}✗ Docker not found. Please install Docker.${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Docker available${NC}"

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}✗ Docker Compose not found.${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Docker Compose available${NC}"

if [ ! -f "$COMPOSE_FILE" ]; then
    echo -e "${RED}✗ $COMPOSE_FILE not found${NC}"
    exit 1
fi
echo -e "${GREEN}✓ $COMPOSE_FILE found${NC}"

if [ "$PROFILE" = "full" ]; then
    echo -e "${YELLOW}Validating required security secrets for $PROFILE...${NC}"

    MONGO_PASSWORD_VALUE="${MONGO_PASSWORD:-}"
    REDIS_PASSWORD_VALUE="${REDIS_PASSWORD:-}"
    GRAFANA_PASSWORD_VALUE="${GRAFANA_PASSWORD:-${GRAFANA_ADMIN_PASSWORD:-}}"

    if [ -f ".env" ]; then
        if [ -z "$MONGO_PASSWORD_VALUE" ]; then
            MONGO_PASSWORD_VALUE=$(grep -E '^MONGO_PASSWORD=' .env | tail -n1 | cut -d '=' -f2-)
        fi
        if [ -z "$REDIS_PASSWORD_VALUE" ]; then
            REDIS_PASSWORD_VALUE=$(grep -E '^REDIS_PASSWORD=' .env | tail -n1 | cut -d '=' -f2-)
        fi
        if [ -z "$GRAFANA_PASSWORD_VALUE" ]; then
            GRAFANA_PASSWORD_VALUE=$(grep -E '^GRAFANA_PASSWORD=' .env | tail -n1 | cut -d '=' -f2-)
        fi
        if [ -z "$GRAFANA_PASSWORD_VALUE" ]; then
            GRAFANA_PASSWORD_VALUE=$(grep -E '^GRAFANA_ADMIN_PASSWORD=' .env | tail -n1 | cut -d '=' -f2-)
        fi
    fi

    if [ -n "$COMPOSE_ENV_FILE" ] && [ -f "$COMPOSE_ENV_FILE" ]; then
        if [ -z "$MONGO_PASSWORD_VALUE" ]; then
            MONGO_PASSWORD_VALUE=$(grep -E '^MONGO_PASSWORD=' "$COMPOSE_ENV_FILE" | tail -n1 | cut -d '=' -f2-)
        fi
        if [ -z "$REDIS_PASSWORD_VALUE" ]; then
            REDIS_PASSWORD_VALUE=$(grep -E '^REDIS_PASSWORD=' "$COMPOSE_ENV_FILE" | tail -n1 | cut -d '=' -f2-)
        fi
        if [ -z "$GRAFANA_PASSWORD_VALUE" ]; then
            GRAFANA_PASSWORD_VALUE=$(grep -E '^GRAFANA_PASSWORD=' "$COMPOSE_ENV_FILE" | tail -n1 | cut -d '=' -f2-)
        fi
        if [ -z "$GRAFANA_PASSWORD_VALUE" ]; then
            GRAFANA_PASSWORD_VALUE=$(grep -E '^GRAFANA_ADMIN_PASSWORD=' "$COMPOSE_ENV_FILE" | tail -n1 | cut -d '=' -f2-)
        fi
    fi

    case "$MONGO_PASSWORD_VALUE" in
        ""|CHANGE_ME*|dev_only_not_for_production)
            echo -e "${RED}✗ MONGO_PASSWORD is missing or insecure placeholder.${NC}"
            echo "  Set a strong value in environment or .env before production deployment."
            exit 1
            ;;
    esac

    case "$REDIS_PASSWORD_VALUE" in
        ""|CHANGE_ME*|dev_only_not_for_production)
            echo -e "${RED}✗ REDIS_PASSWORD is missing or insecure placeholder.${NC}"
            echo "  Set a strong value in environment or .env before production deployment."
            exit 1
            ;;
    esac

    case "$GRAFANA_PASSWORD_VALUE" in
        ""|CHANGE_ME*|changeme|dev_only_not_for_production)
            echo -e "${RED}✗ GRAFANA_PASSWORD is missing or insecure placeholder.${NC}"
            echo "  Set a strong value in environment or .env before production deployment."
            exit 1
            ;;
    esac

    echo -e "${GREEN}✓ Security secrets validation passed${NC}"
fi

# Build images
echo ""
echo -e "${YELLOW}[2/4] Building Docker images...${NC}"
compose_cmd build

# Start services
echo ""
echo -e "${YELLOW}[3/4] Starting services...${NC}"
compose_cmd up -d --scale node-agent="$SCALE"

# Wait for services
echo ""
echo -e "${YELLOW}[4/4] Waiting for services to be healthy...${NC}"
sleep 10

# Show status
echo ""
compose_cmd ps

BACKEND_API_PORT=$(get_profile_var "BACKEND_API_HOST_PORT" "8000")
FRONTEND_PORT=$(get_profile_var "FRONTEND_HOST_PORT" "3000")
GRAFANA_PORT=$(get_profile_var "GRAFANA_HOST_PORT" "3001")
PROMETHEUS_PORT=$(get_profile_var "PROMETHEUS_HOST_PORT" "9090")

# Display access information
echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║             Sovereign Map is Running!                 ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}Access Points:${NC}"
echo "  Frontend:     ${GREEN}http://localhost:${FRONTEND_PORT}${NC}"
echo "  Backend API:  ${GREEN}http://localhost:${BACKEND_API_PORT}${NC}"
echo "  Grafana:      ${GREEN}http://localhost:${GRAFANA_PORT}${NC}  (credentials from GRAFANA_USER/GRAFANA_PASSWORD)"
echo "  Prometheus:   ${GREEN}http://localhost:${PROMETHEUS_PORT}${NC}"
echo ""

# Check convergence
echo -e "${BLUE}Checking convergence...${NC}"
sleep 5
curl -s "http://localhost:${BACKEND_API_PORT}/convergence" | python -m json.tool 2>/dev/null || echo "Backend warming up..."

echo ""
echo -e "${BLUE}Useful Commands:${NC}"
echo "  View logs:        ${GREEN}docker compose ${COMPOSE_ENV_ARGS[*]} -f $COMPOSE_FILE logs -f backend${NC}"
echo "  Monitor:          ${GREEN}watch -n 5 'curl -s http://localhost:${BACKEND_API_PORT}/convergence | jq'${NC}"
echo "  Scale nodes:      ${GREEN}docker compose ${COMPOSE_ENV_ARGS[*]} -f $COMPOSE_FILE up -d --scale node-agent=100${NC}"
echo "  Stop services:    ${GREEN}docker compose ${COMPOSE_ENV_ARGS[*]} -f $COMPOSE_FILE down${NC}"
echo "  Full cleanup:     ${GREEN}docker compose ${COMPOSE_ENV_ARGS[*]} -f $COMPOSE_FILE down -v${NC}"
echo ""
echo -e "${BLUE}Documentation:${NC}"
echo "  Read DOCKER_OPTIMIZATION.md for detailed setup and troubleshooting"
echo ""
