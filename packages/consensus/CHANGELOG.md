# @sovereignmap/consensus Changelog

All notable changes to @sovereignmap/consensus are documented here.

## [0.2.0] - TBD (In Development)

Stable release with comprehensive test coverage for Byzantine-tolerant consensus.

### Added
- **Test Coverage**: 91.13% line coverage with tests for proposal flow and hierarchical aggregation
- **Documentation**: JSDoc comments covering consensus configuration and voting mechanics

### Changed
- SDK version: 0.1.0-alpha.1 → 0.2.0
- Coverage gates: lines 25%, statements 25%, functions 25%, branches 20% (significantly exceeded)

### Breaking Changes
None - 0.2.0 is the first stable release.

---

## [0.1.0-alpha.1] - 2026-03-17

Initial alpha release of consensus primitives.

### Added
- `ConsensusParticipant` class implementing Byzantine-tolerant consensus with 55.5% fault tolerance
- `ConsensusConfig` interface with nodeId, totalNodes, byzantineRatio, and quorumSize
- Proposal submission and voting mechanisms with quorum validation
- `HierarchicalAggregator` class for weighted consensus aggregation across node hierarchy
- EventEmitter support with proposal and vote events
- Zero-knowledge proof generation for consensus commitments

### Test Coverage
- 2 integration tests covering proposal acceptance and hierarchical weighted averaging
- 91.13% line coverage with most branches covered

### Architecture Notes
- Quorum size: (2N/3) + 1 where N = totalNodes
- Byzantine resilience: Tolerates up to 34.5% (69/200) malicious nodes
- Validator: 134 of 200 nodes required for consensus (55.5% threshold)

---

## Version Support

| Version | TypeScript | Node | Status |
|---------|-----------|------|--------|
| 0.2.0+ | 5.3+ | 18+ | Stable (in development) |
| 0.1.0-alpha | 5.3+ | 18+ | Alpha (superseded) |

---

## API Stability

See [SDK_API_STABILITY.md](../SDK_API_STABILITY.md) for version guarantees.
