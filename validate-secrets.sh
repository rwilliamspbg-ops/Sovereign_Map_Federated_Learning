#!/usr/bin/env bash
# Validate required secrets for production-like deployments.
# Usage:
#   bash validate-secrets.sh [prod|large-scale] [optional_env_file]

set -euo pipefail

PROFILE="${1:-prod}"
ENV_FILE="${2:-.env}"

case "$PROFILE" in
  prod|production|large-scale)
    ;;
  *)
    echo "Usage: bash validate-secrets.sh [prod|large-scale] [optional_env_file]"
    echo "Invalid profile: $PROFILE"
    exit 2
    ;;
esac

load_from_env_file_if_missing() {
  local key="$1"
  local current_value="$2"

  if [ -n "$current_value" ]; then
    printf '%s' "$current_value"
    return 0
  fi

  if [ -f "$ENV_FILE" ]; then
    local loaded
    loaded=$(grep -E "^${key}=" "$ENV_FILE" | tail -n1 | cut -d '=' -f2- || true)
    loaded="${loaded%\"}"
    loaded="${loaded#\"}"
    loaded="${loaded%\'}"
    loaded="${loaded#\'}"
    printf '%s' "$loaded"
    return 0
  fi

  printf ''
}

is_placeholder_or_weak() {
  local value="$1"

  case "$value" in
    ""|CHANGE_ME*|changeme|dev|dev_only_not_for_production|default|password|admin|sovereignmap|sovereign2026)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

assert_valid_secret() {
  local key="$1"
  local value="$2"
  local min_len="$3"

  if is_placeholder_or_weak "$value"; then
    echo "❌ $key is missing or uses a placeholder/weak value"
    return 1
  fi

  if [ "${#value}" -lt "$min_len" ]; then
    echo "❌ $key is too short (min ${min_len} chars)"
    return 1
  fi

  return 0
}

MONGO_PASSWORD_VALUE=$(load_from_env_file_if_missing "MONGO_PASSWORD" "${MONGO_PASSWORD:-}")
REDIS_PASSWORD_VALUE=$(load_from_env_file_if_missing "REDIS_PASSWORD" "${REDIS_PASSWORD:-}")
GRAFANA_PASSWORD_VALUE=$(load_from_env_file_if_missing "GRAFANA_ADMIN_PASSWORD" "${GRAFANA_ADMIN_PASSWORD:-}")

FAIL=0
assert_valid_secret "MONGO_PASSWORD" "$MONGO_PASSWORD_VALUE" 20 || FAIL=1
assert_valid_secret "REDIS_PASSWORD" "$REDIS_PASSWORD_VALUE" 20 || FAIL=1
assert_valid_secret "GRAFANA_ADMIN_PASSWORD" "$GRAFANA_PASSWORD_VALUE" 16 || FAIL=1

if [ "$FAIL" -ne 0 ]; then
  echo ""
  echo "Secret validation failed for profile: $PROFILE"
  echo "Set strong secrets via environment or $ENV_FILE before deploying."
  exit 1
fi

echo "✅ Secret validation passed for profile: $PROFILE"
