#!/usr/bin/env bash
set -e
exec "$(dirname "$0")/tests/scripts/bash/commit-test-results.sh" "$@"
