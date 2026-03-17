#!/bin/bash

# SDK Release Notes Generator
# Creates detailed release notes with changelog, SBOMs, and security info

set -e

VERSION="${1:-}"
PREVIOUS_VERSION="${2:-}"
CHANGELOG_FILE="${3:-CHANGELOG.md}"

if [ -z "$VERSION" ]; then
  echo "Usage: generate-release-notes.sh <version> [previous-version] [changelog-file]"
  echo ""
  echo "Examples:"
  echo "  ./generate-release-notes.sh 0.2.0 0.1.0 CHANGELOG.md"
  exit 1
fi

# Determine if this is a prerelease
if [[ "$VERSION" =~ -alpha ]] || [[ "$VERSION" =~ -beta ]]; then
  PRERELEASE=true
  CHANNEL="prerelease"
else
  PRERELEASE=false
  CHANNEL="stable"
fi

echo "📝 Generating Release Notes"
echo "============================"
echo "Version: $VERSION"
echo "Channel: $CHANNEL"
echo "Prerelease: $PRERELEASE"
echo ""

# Extract version-specific changelog section
if grep -q "## \[$VERSION\]" "$CHANGELOG_FILE"; then
  CHANGELOG_SECTION=$(sed -n "/## \[$VERSION\]/,/## \[/p" "$CHANGELOG_FILE" | head -n -1)
else
  CHANGELOG_SECTION="See CHANGELOG.md for details"
fi

# Build release notes
RELEASE_NOTES=$(cat << EOF
# Sovereign Map SDK $VERSION

## Release Summary

This release contains improvements to the Sovereign Map federated learning SDK.

$CHANGELOG_SECTION

## Installation

Install the latest version:

\`\`\`bash
npm install @sovereignmap/core@$VERSION
npm install @sovereignmap/privacy@$VERSION
npm install @sovereignmap/consensus@$VERSION
npm install @sovereignmap/island@$VERSION
\`\`\`

Or install all packages together:

\`\`\`bash
npm install @sovereignmap/{core,privacy,consensus,island}@$VERSION
\`\`\`

## What's New

### Packages
- \`@sovereignmap/core\` - Main SDK client with node lifecycle and privacy integration
- \`@sovereignmap/privacy\` - SGP-001 differential privacy engine
- \`@sovereignmap/consensus\` - Byzantine-tolerant consensus primitives
- \`@sovereignmap/island\` - Offline operation with tamper-evident state

### Release Artifacts

This release includes the following security artifacts:

- **Software Bill of Materials (SBOM)**: CycloneDX format in \`.sbom/\` directory
  - Lists all direct and transitive dependencies
  - Enables vulnerability scanning and supply-chain audits
  
- **Provenance Attestation**: Supply-chain provenance via Sigstore
  - Verifiable build and release information
  - Enables SLSA framework compliance

### API Stability

This release follows [Semantic Versioning](https://semver.org/):
- **Major**: Breaking changes to public APIs (major version bump)
- **Minor**: New features with backward compatibility (minor version bump)
- **Patch**: Bug fixes without API changes (patch version bump)

See [SDK_API_STABILITY.md](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/blob/main/SDK_API_STABILITY.md) for versioning policy and stability guarantees.

## Security

### Reporting Security Issues

If you discover a security vulnerability, please email security@sovereignmap.dev instead of using the issue tracker.

### Verification

Verify the integrity of this release:

1. **Provenance**: Check the Sigstore attestation
2. **SBOM**: Review dependencies in \`.sbom/*.json\` files
3. **Signatures**: Verify GPG signatures on release artifacts (if configured)

### Deprecation Policy

For information on API deprecation and long-term support, see our [API Stability Policy](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/blob/main/SDK_API_STABILITY.md#deprecation-process).

## Performance & Testing

### Test Coverage
- \`@sovereignmap/core\`: 66.56% line coverage
- \`@sovereignmap/privacy\`: 100% line coverage
- \`@sovereignmap/consensus\`: 91.13% line coverage
- \`@sovereignmap/island\`: 88.4% line coverage

### Benchmarks

Run performance benchmarks with:
\`\`\`bash
npm run bench --prefix packages/core
\`\`\`

## Known Issues

See [GitHub Issues](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/issues) for active issues and feature requests.

## Support & Community

- **Documentation**: [SDK Docs](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/tree/main/docs)
- **Issues**: [GitHub Issues](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/issues)
- **Discussions**: [GitHub Discussions](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/discussions)

## Contributors

This release was made possible by contributions from the Sovereign Map community. Thank you!

---

**Release Date**: $(date -u +%Y-%m-%d)  
**Release Channel**: $CHANNEL  
**Node.js Support**: 18+  
**TypeScript Support**: 5.3+
EOF
)

echo "$RELEASE_NOTES"

# Save to file if output directory specified
if [ -n "$4" ]; then
  OUTPUT_FILE="$4/release-notes-$VERSION.md"
  mkdir -p "$(dirname "$OUTPUT_FILE")"
  echo "$RELEASE_NOTES" > "$OUTPUT_FILE"
  echo ""
  echo "✅ Release notes saved to $OUTPUT_FILE"
fi
