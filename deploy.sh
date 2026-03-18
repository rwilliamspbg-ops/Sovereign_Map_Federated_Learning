#!/usr/bin/env bash
# Sovereign Map Docker Quick Start Launcher
# Usage: bash deploy.sh [dev|prod|full|large-scale]

set -e

PROFILE="${1:-dev}"

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
    echo -e "${YELLOW}Select deployment profile:${NC}"
    echo "  1) dev         - Development (1 node, 2 min)"
    echo "  2) prod        - Production (50 nodes, 5 min)"
    echo "  3) full        - Full stack (5 nodes, 3 min)"
    echo "  4) large-scale - Large-Scale (500 nodes, 15 min)"
    echo "  5) Exit"
}

# Map profile names
case "$PROFILE" in
    dev)
        COMPOSE_FILE="docker-compose.dev.yml"
        DEFAULT_ENV_FILE=".env.dev"
        SCALE="1"
        MEMORY="1-2GB"
        DESC="Development Environment"
        ;;
    prod|production)
        COMPOSE_FILE="docker-compose.production.yml"
        DEFAULT_ENV_FILE=".env.production"
        SCALE="50"
        MEMORY="4-6GB"
        DESC="Production Environment"
        ;;
    full)
        COMPOSE_FILE="docker-compose.full.yml"
        DEFAULT_ENV_FILE=".env.full"
        SCALE="5"
        MEMORY="3-6GB"
        DESC="Full Stack Environment"
        ;;
    large-scale)
        COMPOSE_FILE="docker-compose.large-scale.yml"
        DEFAULT_ENV_FILE=".env.production"
        SCALE="500+"
        MEMORY="8-16GB+"
        DESC="Large-Scale Testnet"
        ;;
    *)
        show_menu
        read -r -p "Choose option (1-5): " choice
        case "$choice" in
            1) PROFILE="dev" && $0 dev ;;
            2) PROFILE="prod" && $0 prod ;;
            3) PROFILE="full" && $0 full ;;
            4) PROFILE="large-scale" && $0 large-scale ;;
            5) exit 0 ;;
            *) echo "Invalid option" && exit 1 ;;
        esac
        exit 0
        ;;
esac

COMPOSE_ENV_FILE="${COMPOSE_ENV_FILE:-$DEFAULT_ENV_FILE}"
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

echo -e "${GREEN}Profile: $DESC${NC}"
echo "  Nodes: $SCALE"
echo "  Memory: $MEMORY"
echo "  Compose: $COMPOSE_FILE"
if [ -n "$COMPOSE_ENV_FILE" ] && [ -f "$COMPOSE_ENV_FILE" ]; then
    echo "  Env file: $COMPOSE_ENV_FILE"
else
    echo "  Env file: none (using compose defaults)"
fi
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

if [ "$PROFILE" = "prod" ] || [ "$PROFILE" = "production" ] || [ "$PROFILE" = "large-scale" ]; then
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
if [ "$PROFILE" = "prod" ] || [ "$PROFILE" = "production" ]; then
    compose_cmd up -d --scale node-agent=50
elif [ "$PROFILE" = "full" ]; then
    compose_cmd up -d --scale node-agent=5
elif [ "$PROFILE" = "large-scale" ]; then
    read -r -p "Enter number of nodes (default 500): " NODES
    NODES=${NODES:-500}
    compose_cmd up -d --scale node-agent="$NODES"
else
    compose_cmd up -d
fi

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
