#!/bin/bash

# SDK Release Channel Manager
# Manages npm dist-tags for release channels (latest, next, stable)

set -e

PACKAGES=("@sovereignmap/core" "@sovereignmap/privacy" "@sovereignmap/consensus" "@sovereignmap/island")
VERSION="${1:-}"
CHANNEL="${2:-latest}"  # latest, next, stable
REGISTRY="${3:-npm}"    # npm or local

if [ -z "$VERSION" ]; then
  echo "Usage: release-channels.sh <version> [channel] [registry]"
  echo ""
  echo "Channels:"
  echo "  latest  - Stable production release (recommended for most users)"
  echo "  next    - Next major version or prerelease candidate"
  echo "  stable  - Long-term support release"
  echo ""
  echo "Examples:"
  echo "  ./release-channels.sh 0.2.0 latest"
  echo "  ./release-channels.sh 0.3.0-beta.1 next"
  echo "  ./release-channels.sh 1.0.0 stable"
  exit 1
fi

echo "📦 SDK Release Channel Manager"
echo "================================"
echo "Version: $VERSION"
echo "Channel: $CHANNEL"
echo ""

case "$CHANNEL" in
  latest)
    echo "🚀 Publishing to 'latest' channel (recommended for production)"
    for pkg in "${PACKAGES[@]}"; do
      echo "  Tag $pkg@$VERSION as latest..."
      npm dist-tag add "$pkg@$VERSION" latest --registry "https://registry.npmjs.org" || true
      # Remove from prerelease channels if applicable
      npm dist-tag rm "$pkg" next --registry "https://registry.npmjs.org" 2>/dev/null || true
    done
    ;;
  
  next)
    echo "⚡ Publishing to 'next' channel (prerelease/beta)"
    for pkg in "${PACKAGES[@]}"; do
      echo "  Tag $pkg@$VERSION as next..."
      npm dist-tag add "$pkg@$VERSION" next --registry "https://registry.npmjs.org" || true
      # Keep latest tag separate
      if [[ ! "$VERSION" =~ -alpha ]]; then
        npm dist-tag add "$pkg@$VERSION" latest --registry "https://registry.npmjs.org" || true
      fi
    done
    ;;
  
  stable)
    echo "🔒 Publishing to 'stable' channel (LTS support)"
    for pkg in "${PACKAGES[@]}"; do
      echo "  Tag $pkg@$VERSION as stable..."
      npm dist-tag add "$pkg@$VERSION" stable --registry "https://registry.npmjs.org" || true
      # Mark as latest if it's a newer stable version
      npm dist-tag add "$pkg@$VERSION" latest --registry "https://registry.npmjs.org" || true
    done
    ;;
  
  *)
    echo "❌ Unknown channel: $CHANNEL"
    echo "Allowed channels: latest, next, stable"
    exit 1
    ;;
esac

echo ""
echo "✅ Release channel tags updated"
echo ""
echo "📋 Current tags for @sovereignmap/core:"
npm dist-tag ls @sovereignmap/core --registry "https://registry.npmjs.org" | head -5 || echo "  (not yet published)"
