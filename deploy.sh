#!/usr/bin/env bash
# Sovereign Map Docker Quick Start Launcher
# Usage: bash deploy.sh [dev|prod|large-scale]

set -e

PROFILE="${1:-dev}"
PROJECT_NAME="sovereignmap"

# Color output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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
    echo "  3) large-scale - Large-Scale (500 nodes, 15 min)"
    echo "  4) Exit"
}

# Map profile names
case "$PROFILE" in
    dev)
        COMPOSE_FILE="docker-compose.dev.yml"
        SCALE="1"
        MEMORY="1-2GB"
        DESC="Development Environment"
        ;;
    prod|production)
        COMPOSE_FILE="docker-compose.production.yml"
        SCALE="50"
        MEMORY="4-6GB"
        DESC="Production Environment"
        ;;
    large-scale)
        COMPOSE_FILE="docker-compose.large-scale.yml"
        SCALE="500+"
        MEMORY="8-16GB+"
        DESC="Large-Scale Testnet"
        ;;
    *)
        show_menu
        read -p "Choose option (1-4): " choice
        case "$choice" in
            1) PROFILE="dev" && $0 dev ;;
            2) PROFILE="prod" && $0 prod ;;
            3) PROFILE="large-scale" && $0 large-scale ;;
            4) exit 0 ;;
            *) echo "Invalid option" && exit 1 ;;
        esac
        exit 0
        ;;
esac

echo -e "${GREEN}Profile: $DESC${NC}"
echo "  Nodes: $SCALE"
echo "  Memory: $MEMORY"
echo "  Compose: $COMPOSE_FILE"
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

# Build images
echo ""
echo -e "${YELLOW}[2/4] Building Docker images...${NC}"
docker compose -f "$COMPOSE_FILE" build

# Start services
echo ""
echo -e "${YELLOW}[3/4] Starting services...${NC}"
if [ "$PROFILE" = "prod" ] || [ "$PROFILE" = "production" ]; then
    docker compose -f "$COMPOSE_FILE" up -d --scale node-agent=50
elif [ "$PROFILE" = "large-scale" ]; then
    read -p "Enter number of nodes (default 500): " NODES
    NODES=${NODES:-500}
    docker compose -f "$COMPOSE_FILE" up -d --scale node-agent=$NODES
else
    docker compose -f "$COMPOSE_FILE" up -d
fi

# Wait for services
echo ""
echo -e "${YELLOW}[4/4] Waiting for services to be healthy...${NC}"
sleep 10

# Show status
echo ""
docker compose -f "$COMPOSE_FILE" ps

# Display access information
echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║             Sovereign Map is Running!                 ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}Access Points:${NC}"
echo "  Frontend:     ${GREEN}http://localhost:3000${NC}"
echo "  Backend API:  ${GREEN}http://localhost:8000${NC}"
echo "  Grafana:      ${GREEN}http://localhost:3001${NC}  (admin/admin or admin/dev)"
echo "  Prometheus:   ${GREEN}http://localhost:9090${NC}"
echo ""

# Check convergence
echo -e "${BLUE}Checking convergence...${NC}"
sleep 5
curl -s http://localhost:8000/convergence | python -m json.tool 2>/dev/null || echo "Backend warming up..."

echo ""
echo -e "${BLUE}Useful Commands:${NC}"
echo "  View logs:        ${GREEN}docker compose -f $COMPOSE_FILE logs -f backend${NC}"
echo "  Monitor:          ${GREEN}watch -n 5 'curl -s http://localhost:8000/convergence | jq'${NC}"
echo "  Scale nodes:      ${GREEN}docker compose -f $COMPOSE_FILE up -d --scale node-agent=100${NC}"
echo "  Stop services:    ${GREEN}docker compose -f $COMPOSE_FILE down${NC}"
echo "  Full cleanup:     ${GREEN}docker compose -f $COMPOSE_FILE down -v${NC}"
echo ""
echo -e "${BLUE}Documentation:${NC}"
echo "  Read DOCKER_OPTIMIZATION.md for detailed setup and troubleshooting"
echo ""
