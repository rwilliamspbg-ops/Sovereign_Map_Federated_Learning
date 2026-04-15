#!/usr/bin/env bash
set -euo pipefail

GO_BIN="${GO_BIN:-$(command -v go || true)}"
if [[ -z "$GO_BIN" ]]; then
  echo "go not found on PATH" >&2
  exit 1
fi

if command -v readlink >/dev/null 2>&1; then
  GO_BIN_REAL="$(readlink -f "$GO_BIN" 2>/dev/null || true)"
  if [[ -n "$GO_BIN_REAL" ]]; then
    GO_BIN="$GO_BIN_REAL"
  fi
fi

TOOLROOT="$(dirname "$(dirname "$GO_BIN")")"
DRIVER_VERSION="$($GO_BIN version | awk '{print $3}')"
COMPILE_VERSION="$(GOROOT="$TOOLROOT" GOTOOLCHAIN=local "$GO_BIN" tool compile -V=full | awk '{print $3}')"
LINK_VERSION="$(GOROOT="$TOOLROOT" GOTOOLCHAIN=local "$GO_BIN" tool link -V=full | awk '{print $3}')"
DEFAULT_COMPILE_VERSION="$($GO_BIN tool compile -V=full | awk '{print $3}')"

echo "go binary: $GO_BIN"
echo "toolroot: $TOOLROOT"
echo "driver: $DRIVER_VERSION"
echo "compile(aligned): $COMPILE_VERSION"
echo "link(aligned): $LINK_VERSION"
echo "compile(default env): $DEFAULT_COMPILE_VERSION"

if [[ "$DRIVER_VERSION" != "$COMPILE_VERSION" || "$DRIVER_VERSION" != "$LINK_VERSION" ]]; then
  echo "ERROR: aligned toolchain components are inconsistent" >&2
  exit 1
fi

if [[ "$DEFAULT_COMPILE_VERSION" != "$DRIVER_VERSION" ]]; then
  echo "WARNING: default shell Go environment is mixed ($DEFAULT_COMPILE_VERSION vs $DRIVER_VERSION)" >&2
  echo "Use scripts/go-with-aligned-toolchain.sh for deterministic Go commands." >&2
fi

echo "Go toolchain sanity check passed."
