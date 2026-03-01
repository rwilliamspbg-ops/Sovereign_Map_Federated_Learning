#!/usr/bin/env bash
set -e
exec "$(dirname "$0")/tests/scripts/bash/setup-200node-test.sh" "$@"
