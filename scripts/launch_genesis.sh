#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ -x "${SCRIPT_DIR}/genesis-launch.sh" ]]; then
  exec "${SCRIPT_DIR}/genesis-launch.sh"
fi

echo "genesis-launch.sh not found or not executable in ${SCRIPT_DIR}"
exit 1
