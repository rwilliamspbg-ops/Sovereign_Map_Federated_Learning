---
"@sovereignmap/core": patch
"@sovereignmap/privacy": patch
"@sovereignmap/consensus": patch
"@sovereignmap/island": patch
---

**Phase 4 Completion: Release Engineering Maturity with Supply-Chain Security**

### Release Channel Management
- Created `release-channels.sh` script for managing npm dist-tags
- Support for three release channels:
  - `latest` - Recommended production release
  - `next` - Prerelease and beta versions
  - `stable` - Long-term support releases
- Added new `sdk:channels` npm script for channel management
- Created `sdk-channels.yml` GitHub Actions workflow for manual channel updates

### Software Bill of Materials (SBOM)
- Integrated CycloneDX npm tool for SBOM generation
- Created `generate-sbom.sh` script generating JSON SBOMs for all packages
- SBOMs list all direct and transitive dependencies
- Enable vulnerability scanning and supply-chain audits
- Added to GitHub releases as supplementary artifacts

### Release Documentation
- Created comprehensive `SDK_RELEASE_GUIDE.md` covering:
  - Release phases and checklist
  - Version management and channels
  - Supply-chain security practices
  - Hotfix procedures
  - FAQ and troubleshooting

### Automated Release Notes
- Created `generate-release-notes.sh` script with:
  - Automatic changelog extraction
  - Installation instructions for all channels
  - Security artifact documentation
  - API stability and deprecation info
  - Performance and testing summaries

### CI/CD Enhancements
- Enhanced `sdk-publish.yml` to:
  - Automatically generate SBOMs on publish
  - Create detailed release notes
  - Attach SBOM files to GitHub releases
  - Use softprops/action-gh-release for better release management
- Updated npm scripts with `sdk:sbom`, `sdk:channels`, `sdk:release-notes`, `sdk:release`

### Supply-Chain Security Foundation
- SBOM generation ready for vulnerability scanning
- Release channel management enables staged rollout strategy
- GitHub release artifacts include supply-chain documentation
- Workflow foundation for future Sigstore/cosign integration

### Next Phase Readiness (Phase 5)
- Provenance signing via Sigstore integration
- SLSA framework compliance
- Security scanning automation
- Dependency vulnerability detection
