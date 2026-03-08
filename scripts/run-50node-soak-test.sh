#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${ROOT_DIR}/test-results"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
OUT_FILE="${OUT_DIR}/50node-soak-${STAMP}.txt"

mkdir -p "${OUT_DIR}"

cd "${ROOT_DIR}"

go run ./testnet/simulator/cmd \
  -nodes 50 \
  -rounds 600 \
  -round-ms 250 \
  -straggler-rate 0.10 \
  -malicious-rate 0.02 | tee "${OUT_FILE}"

echo "soak test result saved to ${OUT_FILE}"
