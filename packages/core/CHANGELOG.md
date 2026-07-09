# @sovereignmap/core Changelog

## 0.1.2

### Patch Changes

- f2bf0a4: Bump dependency versions: @libp2p/tcp 11.0.14→11.0.15, libp2p 3.1.7→3.2.0, @types/node 25.5.2→25.6.0, @typescript-eslint/eslint-plugin 8.58.1→8.58.2, @typescript-eslint/parser 8.58.1→8.58.2, @vitest/coverage-v8 4.1.3→4.1.4, vitest 4.1.3→4.1.4, libsodium-wrappers 0.8.2→0.8.3
- Updated dependencies [f2bf0a4]
  - @sovereignmap/consensus@0.1.2
  - @sovereignmap/island@0.1.2
  - @sovereignmap/privacy@0.1.2

## 0.1.1

### Patch Changes

- Finalize 0.1.1 SDK patch release with release-note asset workflow and package artifact alignment.
- Updated dependencies
  - @sovereignmap/privacy@0.1.1
  - @sovereignmap/consensus@0.1.1
  - @sovereignmap/island@0.1.1

## 0.1.0

### Minor Changes

- f569753: **Phase 3 Completion: API Stability Policy & Release Management Infrastructure**

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

### Patch Changes

- 1fdbb45: **Phase 4 Completion: Release Engineering Maturity with Supply-Chain Security**

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

- 3f89923: **Phase 5 Completion: Security & Supply-Chain Attestation**

  ### Security & Vulnerability Scanning

  - Created `sdk-security.yml` workflow with:
    - npm audit vulnerability scanning (with audit-level thresholds)
    - License compliance checking (GPL/AGPL detection)
    - CodeQL static analysis for TypeScript code
    - Scheduled daily scans at 2 AM UTC
    - Automatic reporting to GitHub Security tab
  - Created `security-audit.sh` script for manual vulnerability audit
  - Created `dependency-policy.sh` script for dependency policy enforcement
  - Added npm scripts: `security:audit`, `security:audit:strict`, `security:deps`, `security:check`

  ### SLSA Framework Compliance (Level 2)

  - Created `sdk-provenance.yml` workflow implementing:
    - Build provenance recording in in-toto format
    - SLSA L2 compliance verification
    - Supply-chain provenance artifacts
    - Build environment documentation
    - Provenance upload as CI artifacts
  - Achieves SLSA Build Level 2 with:
    - Authenticated build service (GitHub Actions)
    - Isolated build environments (Ubuntu runners)
    - Reproducible builds with pinned dependencies
    - Complete build audit trails
    - Provenance attestation in in-toto JSON format

  ### Security Documentation & Policy

  - Created comprehensive `SECURITY.md` with:
    - Vulnerability reporting procedure and email contact
    - Security response timeline (24h/5d/30d)
    - Vulnerability disclosure process (private fix → staging → public)
    - CVE and advisory handling procedures
    - Dependency management and update strategy
    - License compliance requirements
    - SLSA framework requirements and verification
    - Best practices for SDK users
    - Security contacts and response SLAs

  ### Workflows & Automation

  - Enhanced `.github/workflows/sdk-security.yml`:
    - npm audit on schedule and on PR/push
    - License compliance scanning
    - CodeQL static analysis
    - Automated security report generation
  - New `.github/workflows/sdk-provenance.yml`:
    - Provenance recording for supply-chain transparency
    - SLSA compliance verification
    - Build artifact documentation

  ### Security Best Practices

  - Documentation covers:
    - Dependency pinning strategy
    - Lock file integrity verification
    - Transitive dependency scanning
    - Update frequency recommendations
    - Vulnerability remediation procedures
    - Supply-chain transparency practices

  ### Foundation for Future Enhancements (Phase 5+)

  - ~SLSA L2 achieved; path to L3 documented~
  - Sigstore/cosign integration ready for future phases
  - Provenance attestation format established
  - Security scanning infrastructure operational
  - Vulnerability remediation processes formalized

- b95e809: Triggering version workflow for Phase 3 enhancements.
- Updated dependencies [f569753]
- Updated dependencies [1fdbb45]
- Updated dependencies [3f89923]
- Updated dependencies [b95e809]
  - @sovereignmap/privacy@0.1.0
  - @sovereignmap/consensus@0.1.0
  - @sovereignmap/island@0.1.0

All notable changes to @sovereignmap/core are documented here.

## [0.2.0] - TBD (In Development)

First stable release with complete API contracts and comprehensive test coverage.

### Added

- **Test Coverage**: 66.56% line coverage with 17 unit and integration tests
- **Lifecycle Testing**: Full end-to-end node initialization covering TPM, privacy, island, network, and consensus initialization paths
- **Network Event Testing**: Tests for disconnect, reconnect, and byzantineDetected event handlers
- **Error Path Testing**: Comprehensive error wrapping and state transition validation
- **Privacy Integration**: Budget exhaustion scenarios and proof generation fallback paths
- **Documentation**: JSDoc comments on all public methods and types

### Fixed

- Network callback handlers now properly tested for all state transitions

### Changed

- SDK version constants: 0.1.0-alpha.1 → 0.2.0
- Coverage gates increased: lines 60%, statements 60%, functions 55%, branches 45%

### Breaking Changes

None - 0.2.0 is the first stable release.

---

## [0.1.0-alpha.1] - 2026-03-17

Initial alpha release of core SDK.

### Added

- `SovereignNode` class for main node client with full lifecycle management
- `NetworkClient` class for libp2p-based P2P networking
- `MessageType` enum for protocol message types
- Error hierarchy: `SovereignMapError`, `NodeInitializationError`, `PrivacyBudgetExceededError`, `ConsensusError`, `NetworkError`, `HardwareAttestationError`, `IslandModeError`
- Constants: `SDK_VERSION`, `PROTOCOL_VERSION`, `SGP001_VERSION`, `MIN_NODE_VERSION`
- `createLogger()` function for structured logging with pino
- Type exports: `NodeConfig`, `NodeStatus`, `NetworkTopology`, `MapUpdate`, `PrivatizedUpdate`, etc.

### Test Coverage

- Initial test suite with 3 baseline tests covering SDK constants, error types, and logger
- Unit tests for NetworkClient (4 tests) covering disconnect, broadcast, and status tracking
- Unit tests for SovereignNode (5 tests) covering submission, consensus, and shutdown flows
- Integration tests (5 tests) covering lifecycle, callbacks, and error scenarios
- Total: 17 tests with 66.56% line coverage

---

## Version Support

| Version     | TypeScript | Node | Status                  |
| ----------- | ---------- | ---- | ----------------------- |
| 0.2.0+      | 5.3+       | 18+  | Stable (in development) |
| 0.1.0-alpha | 5.3+       | 18+  | Alpha (superseded)      |

---

## API Stability

See [SDK_API_STABILITY.md](../SDK_API_STABILITY.md) for deprecation policy, breaking change procedures, and version support guarantees.
