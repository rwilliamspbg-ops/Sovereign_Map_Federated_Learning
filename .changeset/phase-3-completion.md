---
"@sovereignmap/core": minor
"@sovereignmap/privacy": patch
"@sovereignmap/consensus": patch
"@sovereignmap/island": patch
---

**Phase 3 Completion: API Stability Policy & Release Management Infrastructure**

### API Stability & Contracts
- Created comprehensive SDK API Stability Policy (`SDK_API_STABILITY.md`)
  - Defined three stability tiers: Stable (core APIs), Subject to Change (SPIs), and Internal
  - Established 2-minor-version deprecation window for breaking changes
  - Documented versioning scheme with clear guarantees for 0.x and 1.0+ phases
  - Defined public API surface for all 4 packages with explicit stability levels

### Release Management
- Integrated changesets for automated changelog management
- Added npm scripts for version management:
  - `npm run changeset` - Create new changesets
  - `npm run changeset:version` - Bump versions based on changesets
  - `npm run changeset:status` - Check pending version bumps
  - `npm run sdk:version` - Run full test → version → build pipeline
- Created CHANGELOG.md files for root and all 4 packages
- Enhanced GitHub Actions workflows:
  - New `sdk-version.yml` workflow detects changesets and creates version PRs
  - Updated `sdk-publish.yml` workflow to handle automatic releases and GitHub release creation

### Documentation Enhancements
- Master changelog documenting Phase 1 (scaffolding), Phase 2 (coverage), and Phase 3 (stability)
- Per-package changelogs with version history and API stability notes
- Clear deprecation roadmap from 0.2.0 → 1.0.0 with defined breaking change procedures

### Next Phase Readiness
- API contracts locked and versioning infrastructure in place
- Ready for Phase 4 (Release Engineering Maturity) focused on prerelease channels and supply-chain attestation
