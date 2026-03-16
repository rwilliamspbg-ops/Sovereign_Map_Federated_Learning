#!/usr/bin/env bash
set -euo pipefail

CHANNEL="${1:-dev}"
ENV_FILE="config/channels/${CHANNEL}.env"
COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.full.yml}"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Unknown channel: $CHANNEL"
  echo "Expected one of: dev, stage, prod"
  exit 1
fi

set -a
source "$ENV_FILE"
set +a

echo "Launching channel=$CHANNEL with $COMPOSE_FILE"
docker compose -f "$COMPOSE_FILE" up -d --build

echo "Done. Active channel config loaded from $ENV_FILE"
