#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${HOME}/.local/bin"

mkdir -p "${BIN_DIR}"

cd "${ROOT_DIR}"
go build -o "${BIN_DIR}/sovereign-node" ./cmd/sovereign-node

echo "installed sovereign-node to ${BIN_DIR}/sovereign-node"
echo "add ${BIN_DIR} to PATH if needed"
