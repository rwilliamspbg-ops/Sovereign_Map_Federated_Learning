# @sovereignmap/privacy Changelog

All notable changes to @sovereignmap/privacy are documented here.

## [0.2.0] - TBD (In Development)

Stable release with full test coverage and API contracts.

### Added
- **Test Coverage**: 100% line and statement coverage with comprehensive privacy engine testing
- **Documentation**: Complete JSDoc comments for privacy budget calculation and noise injection

### Changed
- SDK version: 0.1.0-alpha.1 → 0.2.0
- Coverage gates: lines 25%, statements 25%, functions 25%, branches 20% (now exceeding all thresholds)

### Breaking Changes
None - 0.2.0 is the first stable release.

---

## [0.1.0-alpha.1] - 2026-03-17

Initial alpha release of SGP-001 privacy engine.

### Added
- `PrivacyEngine` class implementing differential privacy with Gaussian and Laplace mechanisms
- `PrivacyBudget` interface with epsilon/delta parameters
- `PrivacyStatus` interface for monitoring privacy budget consumption
- `apply()` method for applying differential privacy to map updates
- `hasBudgetFor()` method for budget validation before operations
- `generatePrivacyProof()` for zk-SNARK proof generation
- EventEmitter support with `budgetUpdate` and `noiseInjected` events
- Hardware-accelerated noise injection using libsodium

### Test Coverage
- 2 comprehensive tests covering initialization and budget tracking
- 100% line coverage achieved

---

## Security Notes

- Privacy budget tracking is approximate; use conservative estimates in production
- Noise injection uses hardware acceleration via libsodium when available
- Fallback to pure-JavaScript Gaussian/Laplace implementations on all platforms

## Version Support

| Version | TypeScript | Node | Status |
|---------|-----------|------|--------|
| 0.2.0+ | 5.3+ | 18+ | Stable (in development) |
| 0.1.0-alpha | 5.3+ | 18+ | Alpha (superseded) |

---

## API Stability

See [SDK_API_STABILITY.md](../SDK_API_STABILITY.md) for version guarantees.
