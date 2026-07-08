#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ -x "$(dirname "${BASH_SOURCE[0]}")/genesis-launch.sh" ]]; then
  exec "$(dirname "${BASH_SOURCE[0]}")/genesis-launch.sh"
fi

echo "genesis-launch.sh not found or not executable"
exit 1
