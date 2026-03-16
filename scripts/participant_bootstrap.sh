#!/usr/bin/env bash
set -euo pipefail

BACKEND_URL="http://localhost:8000"
PARTICIPANT_NAME="local-participant"
INVITE_CODE=""
ADMIN_TOKEN="${JOIN_API_ADMIN_TOKEN:-}"
START_NODE="true"

usage() {
  cat <<'EOF'
Usage: scripts/participant_bootstrap.sh [options]

Options:
  --backend-url URL         Backend API URL (default: http://localhost:8000)
  --participant-name NAME   Participant display name (default: local-participant)
  --invite-code CODE        Existing invite code from /join/invite
  --admin-token TOKEN       Admin token to auto-create invite code
  --no-start                Generate cert/env bundle only (do not start container)

Examples:
  scripts/participant_bootstrap.sh --participant-name alice --admin-token local-dev-admin-token
  scripts/participant_bootstrap.sh --invite-code <code> --no-start
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --backend-url)
      BACKEND_URL="$2"
      shift 2
      ;;
    --participant-name)
      PARTICIPANT_NAME="$2"
      shift 2
      ;;
    --invite-code)
      INVITE_CODE="$2"
      shift 2
      ;;
    --admin-token)
      ADMIN_TOKEN="$2"
      shift 2
      ;;
    --no-start)
      START_NODE="false"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      usage
      exit 1
      ;;
  esac
done

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required"
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required"
  exit 1
fi

OUT_DIR="participants/${PARTICIPANT_NAME}"
CERT_DIR="${OUT_DIR}/certs"
mkdir -p "$CERT_DIR"

if [[ -z "$INVITE_CODE" ]]; then
  if [[ -z "$ADMIN_TOKEN" ]]; then
    echo "Provide --invite-code or --admin-token"
    exit 1
  fi

  INVITE_RESPONSE=$(curl -sS -X POST "${BACKEND_URL}/join/invite" \
    -H "Content-Type: application/json" \
    -H "X-Join-Admin-Token: ${ADMIN_TOKEN}" \
    -d "{\"participant_name\":\"${PARTICIPANT_NAME}\",\"max_uses\":1,\"expires_in_hours\":24}")

  INVITE_CODE=$(echo "$INVITE_RESPONSE" | jq -r '.invite_code // empty')
  if [[ -z "$INVITE_CODE" ]]; then
    echo "Failed to create invite code"
    echo "$INVITE_RESPONSE"
    exit 1
  fi
fi

REGISTER_RESPONSE_FILE="${OUT_DIR}/join-registration.json"
curl -sS -X POST "${BACKEND_URL}/join/register" \
  -H "Content-Type: application/json" \
  -d "{\"invite_code\":\"${INVITE_CODE}\",\"participant_name\":\"${PARTICIPANT_NAME}\"}" \
  > "$REGISTER_RESPONSE_FILE"

NODE_ID=$(jq -r '.registration.node_id // empty' "$REGISTER_RESPONSE_FILE")
AGG_HOST=$(jq -r '.aggregator.host // "backend"' "$REGISTER_RESPONSE_FILE")
AGG_PORT=$(jq -r '.aggregator.port // 8080' "$REGISTER_RESPONSE_FILE")

if [[ -z "$NODE_ID" ]]; then
  echo "Registration failed"
  cat "$REGISTER_RESPONSE_FILE"
  exit 1
fi

jq -r '.certificates.node_cert_pem' "$REGISTER_RESPONSE_FILE" > "${CERT_DIR}/node-cert.pem"
jq -r '.certificates.node_key_pem' "$REGISTER_RESPONSE_FILE" > "${CERT_DIR}/node-key.pem"
jq -r '.certificates.ca_cert_pem' "$REGISTER_RESPONSE_FILE" > "${CERT_DIR}/ca-cert.pem"

cat > "${OUT_DIR}/.participant.env" <<EOF
NODE_ID=${NODE_ID}
AGGREGATOR_HOST=${AGG_HOST}
AGGREGATOR_PORT=${AGG_PORT}
PARTICIPANT_CERT_DIR=./${CERT_DIR}
LLM_ADAPTER_MODEL_FAMILY=$(jq -r '.llm_policy.model_family' "$REGISTER_RESPONSE_FILE")
LLM_ADAPTER_MODEL_VERSION=$(jq -r '.llm_policy.model_version' "$REGISTER_RESPONSE_FILE")
LLM_ADAPTER_TOKENIZER_HASH=$(jq -r '.llm_policy.tokenizer_hash' "$REGISTER_RESPONSE_FILE")
LLM_ADAPTER_RANK=$(jq -r '.llm_policy.allowed_adapter_ranks[0]' "$REGISTER_RESPONSE_FILE")
LLM_ADAPTER_TARGET_MODULES=$(jq -r '.llm_policy.required_target_modules | join(",")' "$REGISTER_RESPONSE_FILE")
ENABLE_DP=false
LOCAL_EPOCHS=1
BATCH_SIZE=16
MAX_SAMPLES_PER_NODE=120
EOF

echo "Participant bundle created in ${OUT_DIR}"
echo "Node ID: ${NODE_ID}"
echo "Aggregator: ${AGG_HOST}:${AGG_PORT}"

if [[ "$START_NODE" == "true" ]]; then
  docker network create sovereign-network >/dev/null 2>&1 || true
  docker compose -f docker-compose.participant.yml --env-file "${OUT_DIR}/.participant.env" up -d --build
  echo "Participant node started with docker-compose.participant.yml"
else
  echo "Skipping node start (--no-start)"
fi
