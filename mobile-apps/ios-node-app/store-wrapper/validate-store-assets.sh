#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
META_DIR="$SCRIPT_DIR/metadata/en-US"
CFG_DIR="$SCRIPT_DIR/config"

required=(
  "$META_DIR/name.txt"
  "$META_DIR/subtitle.txt"
  "$META_DIR/description.txt"
  "$META_DIR/keywords.txt"
  "$META_DIR/release-notes.txt"
  "$META_DIR/support-url.txt"
  "$CFG_DIR/ExportOptions-AppStore.plist"
  "$CFG_DIR/PrivacyInfo.xcprivacy"
)

missing=0
for file in "${required[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "[ios] missing required file: $file"
    missing=1
  fi
done

if [[ $missing -ne 0 ]]; then
  exit 1
fi

echo "[ios] metadata/config validation passed"
