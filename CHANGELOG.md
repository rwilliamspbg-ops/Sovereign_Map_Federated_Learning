# Sovereign Map SDK - Master Changelog

All notable changes to Sovereign Map SDK packages are documented here.

## [0.2.0] - TBD (In Development)

This is the first stable release of the Sovereign Map SDK with published API contracts and broad test coverage.

### Added
- **FedAvg benchmark compare CI and report artifact**: Added `.github/workflows/fedavg-benchmark-compare.yml` and `scripts/benchmark_fedavg_compare.sh` to generate and upload base-vs-current benchmark comparison markdown.
- **Benchmark compare Make target**: Added `make benchmark-fedavg-compare` to generate `results/metrics/fedavg_benchmark_compare.md` locally.
- **@sovereignmap/core**: 100% line coverage with comprehensive lifecycle, network integration, aggregation, error handling, and signal tests
- **@sovereignmap/privacy**: 100% test coverage for SGP-001 differential privacy engine
- **@sovereignmap/consensus**: 91.13% test coverage for Byzantine fault-tolerant consensus
- **@sovereignmap/island**: 100% line coverage for offline operation with tamper-evident state and chain integrity verification
- **API Contracts Published**: OpenAPI specs for control plane, training service, TPM exporter, and tokenomics exporter under docs/api
- **Integration Collection Published**: Multi-service Postman collection aligned to active Flask routes
- **Documentation**: API Stability Policy defining versioning, deprecation, and breaking change procedures
- **Release Management**: Changesets-based changelog automation and semantic versioning enforcement
- **Testing Infrastructure**: Coverage gates enforced in CI with per-package thresholds (all packages meeting or exceeding 91% line coverage)
- **API Stability Guardrails**: Route-to-spec coverage validator script and API validation GitHub Actions workflow
- **Hosted API Docs**: GitHub Pages deployment workflow for static Swagger UI with multi-spec selector (`/api/swagger-ui.html` when Pages is enabled)
- **Security Hardening**: GitHub Actions SHA pinning across all workflows; SLSA-aligned provenance recording/reporting workflow; CodeQL and supply chain scanning
- **PySyft Integration Demo**: Added a runnable PySyft x Mohawk proof-of-concept package under `examples/pysyft-integration` with script, notebook, compose profile, and operator runbook

### Fixed
- **@sovereignmap/island**: Chain integrity verification uses canonical entry fields with proper sequence ordering; LevelDB lock handling in tests
- **@sovereignmap/core**: Aggregate handler callback now properly typed with null checks; network handler invocation guarded
- **CI/CD Stability**: SDK version workflow resilient with non-blocking create-PR step; Windows client EXE build fixed (PowerShell param ordering)

### Improved
- **Test Quality**: Added branch-coverage tests for error paths, fallback handlers, metrics collection, and Byzantine fault scenarios
- **Workflow Reliability**: Required CI checks are configured and currently passing on main; publish workflows have no known blockers; deterministic builds are reproducible locally
- **Grafana Dashboard UX**: Consolidated to canonical provisioned dashboards, normalized live refresh defaults, and labeled STARRED easy-read dashboards for operators

### Changed

- **Grafana Provisioning**: Removed duplicate/unused dashboard providers and unused mirrored dashboard JSON files; standardized compose home dashboard path to `operations_overview.json`
- **Monitoring Compose Wiring**: Removed stale dashboard mount paths and aligned monitoring stacks to canonical dashboard provisioning

### Breaking Changes
None - 0.2.0 is the first stable release with public APIs.

### Observability Upgrade - 2026-03-20

- Added end-to-end blockchain and bridge telemetry metrics:
   - `tokenomics_chain_height`
   - `tokenomics_bridge_transfers_total`
   - `tokenomics_bridge_routes_active`
   - `tokenomics_fl_verification_ratio`
   - `tokenomics_fl_average_confidence_bps`
- Extended tokenomics backend payload generation to emit bridge throughput, active routes, chain height, and FL verification confidence fields.
- Updated Grafana Operations and Tokenomics dashboards (provisioned + mirrored) with dedicated blockchain and bridge runtime panels.
- Added `make observability-smoke` to validate dashboard PromQL references and dashboard JSON parse integrity in one command.

### Mobile Shield Release - 2026-03-23

- Added hardware-rooted mobile gradient signing flows:
   - iOS Secure Enclave signer and canonical payload adapter.
   - Android StrongBox/Keystore signer and canonical payload adapter.
- Added backend mobile verification route `POST /mobile/verify_gradient` with verification metrics and contract tests.
- Added production store-wrapper packages for Google Play and Apple App Store (build/validate scripts, metadata/config templates, and submission checklists).
- Hardened go-mobile build repeatability with deterministic toolchain wrapper updates and module dependency fixes.
- Updated control-plane OpenAPI spec coverage for mobile verification endpoint to keep API validation CI green.

### Sprint 10 HUD and Documentation Cleanup - 2026-03-23

- Frontend app shell now defaults to Network Operations HUD and removes Browser FL Studio from primary navigation.
- HUD telemetry now includes browser runtime metrics (RTT, downlink, heap usage, load/interactive timing, viewport/device hints).
- Privacy-Utility analysis mode now uses backend convergence metrics as canonical source.
- Observability CI dashboard-query validator allowlist updated to include consensus and aggregation metrics referenced by operations dashboards.

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

### Added - Phase 3 (Coverage Finalization)
- Extended core test suite covering fallback handlers, error paths, Byzantine scenarios, and metric collection
- Island tamper detection tests (previous hash, chained hash verification)
- Network error injection and handler coverage
- Metrics snapshot and historical tracking tests

### Test Coverage Progression
- Phase 1 Start: @sovereignmap/core ~13% lines
- Phase 2 Mid: @sovereignmap/core 43% lines (network + node unit tests)
- Phase 2 Complete: @sovereignmap/core 66.56% lines (lifecycle integration tests)
- Phase 3 Complete: @sovereignmap/core 100% lines; @sovereignmap/island 100% lines
- All package thresholds exceeded or met by 0.1.0 final build

### Achievements
- All SDK packages now exceed 90%+ line coverage
- Comprehensive branch coverage for error handling and edge cases
- Deterministic test environments with proper resource cleanup

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

### Fixes
- Fixed SDK publish lockfile mismatch
