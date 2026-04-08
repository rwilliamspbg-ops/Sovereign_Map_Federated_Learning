#!/usr/bin/env bash
set -euo pipefail

PREFIX="${1:-}"

if [[ -n "$PREFIX" ]]; then
  CI_CMD=(npm --prefix "$PREFIX" ci)
  LOCKFIX_CMD=(npm --prefix "$PREFIX" install --package-lock-only --no-audit --no-fund)
else
  CI_CMD=(npm ci)
  LOCKFIX_CMD=(npm install --package-lock-only --no-audit --no-fund)
fi

if "${CI_CMD[@]}"; then
  exit 0
fi

echo "npm ci failed; refreshing lockfile metadata and retrying..."
"${LOCKFIX_CMD[@]}"
"${CI_CMD[@]}"
