#!/usr/bin/env bash

set -euo pipefail

TARGET_NODES="${1:-25}"
COMPOSE_FILE="${2:-docker-compose.dev.yml}"
SERVICE_NAME="${SERVICE_NAME:-node-agent}"
CERT_DIR="${CERT_DIR:-/etc/sovereign/certs}"

if ! [[ "$TARGET_NODES" =~ ^[0-9]+$ ]] || [ "$TARGET_NODES" -lt 1 ]; then
  echo "Error: target node count must be a positive integer (got: $TARGET_NODES)"
  exit 1
fi

if [ ! -f "$COMPOSE_FILE" ]; then
  echo "Error: compose file not found: $COMPOSE_FILE"
  exit 1
fi

echo "[burst] compose file: $COMPOSE_FILE"
echo "[burst] service: $SERVICE_NAME"
echo "[burst] target nodes: $TARGET_NODES"

echo "[burst] ensuring core services are running..."
docker compose -f "$COMPOSE_FILE" up -d mongo redis backend tpm-metrics

echo "[burst] scaling $SERVICE_NAME to $TARGET_NODES..."
docker compose -f "$COMPOSE_FILE" up -d --scale "$SERVICE_NAME=$TARGET_NODES" "$SERVICE_NAME"

echo "[certs] issuing/verifying certificates for node IDs 0..$((TARGET_NODES - 1))"
docker compose -f "$COMPOSE_FILE" exec -T tpm-metrics sh -lc "python - <<'PY'
import json
from pathlib import Path
from tpm_cert_manager import TPMCertificateManager

target_nodes = int('${TARGET_NODES}')
cert_dir = '${CERT_DIR}'

mgr = TPMCertificateManager(cert_dir)
created = 0
verified = 0

for node_id in range(target_nodes):
    cert_path = Path(cert_dir) / f'node-{node_id}-cert.pem'
    key_path = Path(cert_dir) / f'node-{node_id}-key.pem'
    if not cert_path.exists() or not key_path.exists():
        mgr.generate_node_cert(node_id, f'Node-{node_id}')
        created += 1
    if mgr.verify_node_certificate(node_id):
        verified += 1

report = mgr.get_trust_report()
print(json.dumps({
    'target_nodes': target_nodes,
    'created_new_certs': created,
    'verified_nodes': verified,
    'trust_store_total': report.get('total_nodes', 0),
    'cert_dir': cert_dir,
}, indent=2))
PY"

running_count="$(docker compose -f "$COMPOSE_FILE" ps -q "$SERVICE_NAME" | wc -l | tr -d ' ')"
echo "[done] running $SERVICE_NAME replicas: $running_count"
echo "[done] auto cert issue complete"
