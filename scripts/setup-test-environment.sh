#!/usr/bin/env bash
set -euo pipefail

# Reproducible local setup for full validation test suite.
# Supports Debian/Ubuntu PEP 668 environments by using --break-system-packages.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

echo "[setup] Installing frontend test dependencies"
npm --prefix frontend ci

echo "[setup] Installing Python dependencies"
if python -m pip install -r requirements.txt >/tmp/setup-test-pip.log 2>&1; then
  echo "[setup] Python dependencies installed with standard pip flow"
else
  echo "[setup] Standard pip install failed, retrying with --break-system-packages"
  python -m pip install --break-system-packages -r requirements.txt
fi

echo "[setup] Environment ready"
