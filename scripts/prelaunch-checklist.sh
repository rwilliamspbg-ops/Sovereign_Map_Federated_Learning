#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${ROOT_DIR}/test-results"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
OUT_FILE="${OUT_DIR}/prelaunch-checklist-${STAMP}.log"

mkdir -p "${OUT_DIR}"

run_step() {
  local label="$1"
  shift
  echo ""
  echo "==> ${label}" | tee -a "${OUT_FILE}"
  "$@" 2>&1 | tee -a "${OUT_FILE}"
}

echo "Sovereign Map V1.1.0 Pre-Launch Checklist" | tee "${OUT_FILE}"
echo "UTC timestamp: ${STAMP}" | tee -a "${OUT_FILE}"

echo "" | tee -a "${OUT_FILE}"
echo "Host baseline" | tee -a "${OUT_FILE}"
uname -a | tee -a "${OUT_FILE}"

echo "" | tee -a "${OUT_FILE}"
echo "Repository baseline" | tee -a "${OUT_FILE}"
cd "${ROOT_DIR}"
git rev-parse --short HEAD | sed 's/^/commit: /' | tee -a "${OUT_FILE}"
git branch --show-current | sed 's/^/branch: /' | tee -a "${OUT_FILE}"

run_step "Build binaries" bash -lc "cd '${ROOT_DIR}' && go build -o /tmp/sovereign-node-check ./cmd/sovereign-node && go build -o /tmp/metrics-exporter-check ./cmd/metrics-exporter && go build -o /tmp/simulator-check ./testnet/simulator/cmd && rm -f /tmp/sovereign-node-check /tmp/metrics-exporter-check /tmp/simulator-check"

run_step "Docker/project validation" bash -lc "cd '${ROOT_DIR}' && bash validate-docker.sh"

run_step "Secrets validation" bash -lc "cd '${ROOT_DIR}' && MONGO_PASSWORD='A_Strong_Mongo_Password_2026_X' REDIS_PASSWORD='A_Strong_Redis_Password_2026_Y' GRAFANA_ADMIN_PASSWORD='StrongGrafanaPass_2026' bash validate-secrets.sh prod"

run_step "Genesis validation" bash -lc "cd '${ROOT_DIR}' && bash validate-genesis-launch.sh"

run_step "Runtime smoke test" bash -lc "cd '${ROOT_DIR}' && go run ./cmd/sovereign-node start -node-id prelaunch-check-node | head -n 1"

run_step "50-node soak smoke" bash -lc "cd '${ROOT_DIR}' && ./scripts/run-50node-soak-test.sh | tail -n 2"

run_step "Targeted Go tests" bash -lc "cd '${ROOT_DIR}' && go test ./internal/p2p ./internal/monitoring ./testnet/scenarios"

echo "" | tee -a "${OUT_FILE}"
echo "Pre-launch checklist completed successfully." | tee -a "${OUT_FILE}"
echo "Log: ${OUT_FILE}" | tee -a "${OUT_FILE}"
