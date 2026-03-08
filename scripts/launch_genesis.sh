#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ -x "${ROOT_DIR}/genesis-launch.sh" ]]; then
  exec "${ROOT_DIR}/genesis-launch.sh"
fi

echo "genesis-launch.sh not found or not executable"
exit 1
