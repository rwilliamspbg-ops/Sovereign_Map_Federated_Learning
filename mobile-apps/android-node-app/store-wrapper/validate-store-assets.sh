#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
META_DIR="$SCRIPT_DIR/metadata/en-US"

required=(
  "title.txt"
  "short-description.txt"
  "full-description.txt"
  "release-notes.txt"
  "privacy-policy-url.txt"
)

missing=0
for file in "${required[@]}"; do
  if [[ ! -f "$META_DIR/$file" ]]; then
    echo "[android] missing metadata file: $META_DIR/$file"
    missing=1
  fi
done

if [[ $missing -ne 0 ]]; then
  exit 1
fi

echo "[android] metadata validation passed"
