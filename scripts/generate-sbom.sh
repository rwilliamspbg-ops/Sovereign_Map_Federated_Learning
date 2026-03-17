#!/bin/bash

# SDK SBOM Generator
# Creates CycloneDX Software Bill of Materials for SDK packages

set -e

PACKAGES=("core" "privacy" "consensus" "island")
OUTPUT_DIR="${1:-.sbom}"
VERSION="${2:-}"

if [ ! -d "$OUTPUT_DIR" ]; then
  mkdir -p "$OUTPUT_DIR"
fi

echo "📦 Generating Software Bill of Materials (SBOMs)"
echo "=================================================="
echo "Output directory: $OUTPUT_DIR"
echo ""

for pkg in "${PACKAGES[@]}"; do
  PKG_DIR="packages/$pkg"
  PKG_NAME="@sovereignmap/$pkg"
  OUTPUT_FILE="$OUTPUT_DIR/${pkg}-sbom.json"
  
  if [ ! -f "$PKG_DIR/package.json" ]; then
    echo "⚠️  Skipping $pkg (package.json not found)"
    continue
  fi
  
  echo "🔍 Generating SBOM for $PKG_NAME..."
  
  cd "$PKG_DIR"
  
  # Generate CycloneDX SBOM using cyclonedx-npm
  npx cyclonedx-npm --output-file "../../$OUTPUT_FILE" --exclude dev 2>/dev/null || \
    # Fallback: Create minimal SBOM if cyclonedx-npm fails
    node -e "
      const pkg = require('./package.json');
      const sbom = {
        bomVersion: 1,
        specVersion: '1.4',
        version: 1,
        metadata: {
          timestamp: new Date().toISOString(),
          tools: [{ vendor: 'Sovereign Map', name: 'SDK Release Tools', version: '1.0.0' }],
          component: {
            bom_ref: '${PKG_NAME}@${VERSION || pkg.version}',
            type: 'library',
            name: '${PKG_NAME}',
            version: '${VERSION || pkg.version}',
            description: pkg.description,
            homepage: pkg.repository?.url,
            licenses: [{ license: { name: pkg.license } }]
          }
        },
        components: Object.entries(pkg.dependencies || {}).map(([name, version]) => ({
          bom_ref: name + '@' + version,
          type: 'library',
          name,
          version
        }))
      };
      console.log(JSON.stringify(sbom, null, 2));
    " > "../../$OUTPUT_FILE"
  
  cd - > /dev/null
  
  FILE_SIZE=$(du -h "$OUTPUT_FILE" | cut -f1)
  echo "  ✓ Created $OUTPUT_FILE ($FILE_SIZE)"
done

echo ""
echo "✅ SBOM generation complete"
echo ""
echo "📋 Generated files:"
ls -lh "$OUTPUT_DIR"/*.json 2>/dev/null | awk '{print "  " $9 " (" $5 ")"}'

echo ""
echo "Usage:"
echo "  Attach .sbom/*-sbom.json files to GitHub releases"
echo "  Reference in security audit reports and supply-chain documentation"
