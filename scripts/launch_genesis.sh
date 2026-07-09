#!/usr/bin/env bash
set -euo pipefail


if [[ -x "./genesis-launch.sh" ]]; then
  exec "./genesis-launch.sh"
fi

echo "genesis-launch.sh not found or not executable"
exit 1
