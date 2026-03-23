#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
KEYSTORE_PROPS="$APP_DIR/keystore.properties"

cd "$APP_DIR"

echo "[android] starting release build"

if [[ ! -f "./gradlew" ]]; then
  echo "[android] error: gradlew not found in $APP_DIR"
  exit 1
fi

bash "$SCRIPT_DIR/validate-store-assets.sh"

./gradlew clean :app:bundleRelease :app:assembleRelease -x test

echo "[android] release artifacts generated"
find "$APP_DIR" -type f \( -name "*.aab" -o -name "*-release.apk" \) | sort

if [[ -f "$KEYSTORE_PROPS" ]]; then
  echo "[android] keystore.properties found; sign config can be consumed by app/build.gradle"
else
  echo "[android] warning: keystore.properties not found; artifacts may be unsigned"
fi

cat <<'EOT'
[android] next steps
1) Upload .aab to Play Console Internal testing.
2) Rollout staged to closed/open tracks.
3) Promote to production after ANR/crash checks.
EOT
