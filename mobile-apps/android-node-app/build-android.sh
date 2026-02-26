#!/bin/bash
# build-android.sh - Build Android app for release

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "📦 Building Sovereign Node App for Android..."
echo "Project: $PROJECT_DIR"

cd "$PROJECT_DIR"

# Build APK
echo "🏗️  Building APK..."
./gradlew clean build -x test

# Build AAB (for Play Store)
echo "🏗️  Building App Bundle..."
./gradlew bundle release -x test

echo "✅ Android build complete!"
echo ""
echo "Output files:"
echo "- APK: app/build/outputs/apk/release/app-release.apk"
echo "- AAB: app/build/outputs/bundle/release/app-release.aab"
echo ""
echo "Next steps:"
echo "1. APK (direct install):"
echo "   adb install app/build/outputs/apk/release/app-release.apk"
echo ""
echo "2. AAB (Google Play):"
echo "   Upload to Play Console: https://play.google.com/console"
echo ""
echo "3. Debug APK:"
echo "   ./gradlew installDebug  # Install to connected device"
