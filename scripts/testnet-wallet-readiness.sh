#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
REPORT_DIR="$ROOT_DIR/test-results/testnet-readiness"
REPORT_FILE="$REPORT_DIR/testnet-wallet-readiness-$TIMESTAMP.md"
mkdir -p "$REPORT_DIR"

PASS_COUNT=0
WARN_COUNT=0
FAIL_COUNT=0

log_pass() {
  PASS_COUNT=$((PASS_COUNT + 1))
  echo "[PASS] $1"
  printf -- "- PASS: %s\n" "$1" >> "$REPORT_FILE"
}

log_warn() {
  WARN_COUNT=$((WARN_COUNT + 1))
  echo "[WARN] $1"
  printf -- "- WARN: %s\n" "$1" >> "$REPORT_FILE"
}

log_fail() {
  FAIL_COUNT=$((FAIL_COUNT + 1))
  echo "[FAIL] $1"
  printf -- "- FAIL: %s\n" "$1" >> "$REPORT_FILE"
}

run_check() {
  local name="$1"
  shift
  if "$@"; then
    log_pass "$name"
  else
    log_fail "$name"
  fi
}

echo "# Testnet Wallet Readiness Report" > "$REPORT_FILE"
echo >> "$REPORT_FILE"
echo "- Timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)" >> "$REPORT_FILE"
echo "- Workspace: $ROOT_DIR" >> "$REPORT_FILE"
echo >> "$REPORT_FILE"
echo "## Checks" >> "$REPORT_FILE"

echo "Running testnet wallet readiness checks..."

TOOLCHAIN_GOROOT="/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.7.linux-amd64"
if [[ -d "$TOOLCHAIN_GOROOT" ]]; then
  export GOROOT="$TOOLCHAIN_GOROOT"
else
  export GOROOT="$(go env GOROOT)"
fi

run_check "Go toolchain available" command -v go >/dev/null
run_check "Wallet unit tests" go test ./internal/blockchain -run "Wallet|CreateTransfer|ApplyTransactionTransfer" -count=1
run_check "Blockchain package tests" go test ./internal/blockchain/... -count=1
run_check "Node package tests" go test ./internal/node/... -count=1
run_check "Consensus package tests" go test ./internal/consensus/... -count=1
run_check "Core build (blockchain, node, consensus)" go build ./internal/blockchain/... ./internal/node/... ./internal/consensus/...
run_check "Wallet-enabled binaries build" go build ./cmd/node-agent ./cmd/metrics-exporter

if command -v docker >/dev/null 2>&1; then
  if docker compose version >/dev/null 2>&1; then
    run_check "Compose config: full" docker compose -f docker-compose.full.yml config >/dev/null
  else
    log_warn "Docker is installed but docker compose plugin is unavailable"
  fi
else
  log_warn "Docker is not available; compose checks skipped"
fi

echo >> "$REPORT_FILE"
echo "## Summary" >> "$REPORT_FILE"
echo >> "$REPORT_FILE"
echo "- Pass: $PASS_COUNT" >> "$REPORT_FILE"
echo "- Warn: $WARN_COUNT" >> "$REPORT_FILE"
echo "- Fail: $FAIL_COUNT" >> "$REPORT_FILE"

if [[ $FAIL_COUNT -eq 0 ]]; then
  echo "" >> "$REPORT_FILE"
  echo "Status: READY" >> "$REPORT_FILE"
  echo "Readiness status: READY"
  echo "Report: $REPORT_FILE"
  exit 0
fi

echo "" >> "$REPORT_FILE"
echo "Status: NOT READY" >> "$REPORT_FILE"
echo "Readiness status: NOT READY"
echo "Report: $REPORT_FILE"
exit 1
