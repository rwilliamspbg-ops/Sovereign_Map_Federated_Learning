# Sovereign Map SDK - Master Changelog

All notable changes to Sovereign Map SDK packages are documented here.

## [0.2.0] - TBD (In Development)

This is the first stable release of the Sovereign Map SDK with complete API contracts and test coverage infrastructure.

### Added
- **@sovereignmap/core**: 66.56% test coverage with lifecycle and network integration tests
- **@sovereignmap/privacy**: 100% test coverage for SGP-001 differential privacy engine
- **@sovereignmap/consensus**: 91.13% test coverage for Byzantine fault-tolerant consensus
- **@sovereignmap/island**: 88.4% test coverage for offline operation with tamper-evident state
- **Documentation**: API Stability Policy defining versioning, deprecation, and breaking change procedures
- **Release Management**: Changesets-based changelog automation and semantic versioning enforcement
- **Testing Infrastructure**: Coverage gates enforced in CI with per-package thresholds

### Fixed
- **@sovereignmap/island**: Chain integrity verification now uses canonical entry fields and proper sequence ordering

### Breaking Changes
None - 0.2.0 is the first stable release with public APIs.

---

## [0.1.0-alpha.1] - 2026-03-17

Initial alpha release with core SDK packages and functionality.

### Added - Phase 1
- Core SDK scaffolding for @sovereignmap/core, @sovereignmap/privacy, @sovereignmap/consensus, @sovereignmap/island
- Vitest configuration with baseline test coverage
- Coverage gates foundation (10-25% thresholds for Phase 1)
- Package manifests and README files for all 4 SDK packages

### Added - Phase 2
- Deep integration testing for SovereignNode lifecycle (initialization, submission, consensus participation, shutdown)
- NetworkClient testing (disconnect handling, broadcast, status updates)
- Privacy engine integration with budget tracking
- Island mode offline queueing and synchronization
- Hierarchical consensus aggregation testing

### Test Coverage Progression
- Phase 1 Start: @sovereignmap/core ~13% lines
- Phase 2 Mid: @sovereignmap/core 43% lines (network + node unit tests)
- Phase 2 Complete: @sovereignmap/core 66.56% lines (lifecycle integration tests)
- All package thresholds exceeded or met by 0.1.0 final build

---

## Release Process

Each SDK release follows this process:

1. **Create Changeset** (`npm run changeset`)
   - Document features, fixes, and breaking changes
   - Select version bump type (major/minor/patch)

2. **Version Update** (`npm run changeset:version`)
   - Auto-bump package.json versions
   - Generate/update CHANGELOG.md files
   - Create version commit

3. **Build & Test** (`npm run sdk:version`)
   - Run full test suite with coverage gates
   - Build all packages with TypeScript

4. **Publish** (`npm run sdk:publish`)
   - Publish all 4 packages to npm as public
   - Tagging strategy: stable, latest, @next for prereleases

See [SDK_API_STABILITY.md](./SDK_API_STABILITY.md) for detailed versioning policy.
