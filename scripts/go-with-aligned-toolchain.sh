#!/usr/bin/env bash
set -euo pipefail

if [[ $# -eq 0 ]]; then
  echo "usage: $0 <go-subcommand> [args...]" >&2
  exit 2
fi

GO_BIN="${GO_BIN:-$(command -v go || true)}"
if [[ -z "$GO_BIN" ]]; then
  echo "go not found on PATH" >&2
  exit 1
fi

# Use the same toolchain root that shipped the selected go binary.
if command -v readlink >/dev/null 2>&1; then
  GO_BIN_REAL="$(readlink -f "$GO_BIN" 2>/dev/null || true)"
  if [[ -n "$GO_BIN_REAL" ]]; then
    GO_BIN="$GO_BIN_REAL"
  fi
fi

TOOLROOT="$(dirname "$(dirname "$GO_BIN")")"
exec env GOROOT="$TOOLROOT" GOTOOLCHAIN=local "$GO_BIN" "$@"
