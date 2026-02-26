#!/bin/bash
# build-ios.sh - Build iOS app for release

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE="$PROJECT_DIR/SovereignNodeApp.xcworkspace"
SCHEME="SovereignNodeApp"
CONFIGURATION="Release"

echo "📦 Building Sovereign Node App for iOS..."
echo "Project: $PROJECT_DIR"

# Build for iPhone
echo "🏗️  Building for iPhone..."
xcodebuild \
    -workspace "$WORKSPACE" \
    -scheme "$SCHEME" \
    -configuration "$CONFIGURATION" \
    -sdk iphoneos \
    -arch arm64 \
    build

# Build for Simulator
echo "🏗️  Building for Simulator..."
xcodebuild \
    -workspace "$WORKSPACE" \
    -scheme "$SCHEME" \
    -configuration "$CONFIGURATION" \
    -sdk iphonesimulator \
    -arch arm64 \
    build

echo "✅ iOS build complete!"
echo ""
echo "Next steps:"
echo "1. Open Xcode: open $WORKSPACE"
echo "2. Select target and scheme"
echo "3. Press Cmd+R to run on connected device"
echo "4. Or submit to App Store: Product > Scheme > Edit Scheme > Run > Run"
