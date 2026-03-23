#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="$APP_DIR/build"
EXPORT_DIR="$BUILD_DIR/export"
ARCHIVE_PATH="$BUILD_DIR/SovereignNodeApp.xcarchive"
EXPORT_OPTIONS="$SCRIPT_DIR/config/ExportOptions-AppStore.plist"

SCHEME="${IOS_SCHEME:-SovereignNodeApp}"
WORKSPACE_PATH="${IOS_WORKSPACE:-$APP_DIR/SovereignNodeApp.xcworkspace}"
PROJECT_PATH="${IOS_PROJECT:-$APP_DIR/SovereignNodeApp.xcodeproj}"

bash "$SCRIPT_DIR/validate-store-assets.sh"
mkdir -p "$BUILD_DIR" "$EXPORT_DIR"

if [[ -f "$WORKSPACE_PATH" ]]; then
  echo "[ios] building from workspace"
  xcodebuild -workspace "$WORKSPACE_PATH" -scheme "$SCHEME" -configuration Release -archivePath "$ARCHIVE_PATH" archive
else
  if [[ ! -f "$PROJECT_PATH" ]]; then
    echo "[ios] error: no workspace or project file found"
    exit 1
  fi
  echo "[ios] building from project"
  xcodebuild -project "$PROJECT_PATH" -scheme "$SCHEME" -configuration Release -archivePath "$ARCHIVE_PATH" archive
fi

xcodebuild -exportArchive -archivePath "$ARCHIVE_PATH" -exportPath "$EXPORT_DIR" -exportOptionsPlist "$EXPORT_OPTIONS"

echo "[ios] export completed"
find "$EXPORT_DIR" -type f \( -name "*.ipa" -o -name "*.plist" \) | sort

cat <<'EOT'
[ios] next steps
1) Upload IPA to App Store Connect (Transporter or xcrun altool/notarytool as applicable).
2) Assign build to TestFlight internal/external groups.
3) Promote to App Store after review and compliance checks.
EOT
