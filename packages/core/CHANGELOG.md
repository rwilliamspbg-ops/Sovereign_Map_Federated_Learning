# @sovereignmap/core Changelog

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

| Version | TypeScript | Node | Status |
|---------|-----------|------|--------|
| 0.2.0+ | 5.3+ | 18+ | Stable (in development) |
| 0.1.0-alpha | 5.3+ | 18+ | Alpha (superseded) |

---

## API Stability

See [SDK_API_STABILITY.md](../SDK_API_STABILITY.md) for deprecation policy, breaking change procedures, and version support guarantees.
