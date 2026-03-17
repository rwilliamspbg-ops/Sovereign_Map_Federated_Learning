# @sovereignmap/island Changelog

All notable changes to @sovereignmap/island are documented here.

## [0.2.0] - TBD (In Development)

Stable release with comprehensive offline operation testing.

### Added
- **Test Coverage**: 88.4% line coverage with full queue/sync/verify/flush lifecycle testing
- **Documentation**: JSDoc comments for offline state management and tamper detection

### Fixed
- **Chain Integrity Verification**: Now correctly uses canonical entry fields and proper sequence ordering for hash chain validation

### Changed
- SDK version: 0.1.0-alpha.1 → 0.2.0  
- Coverage gates: lines 25%, statements 25%, functions 25%, branches 20% (significantly exceeded)

### Breaking Changes
None - 0.2.0 is the first stable release.

---

## [0.1.0-alpha.1] - 2026-03-17

Initial alpha release of Island Mode for autonomous offline operation.

### Added
- `IslandModeManager` class for offline operation with tamper-evident state recovery
- `IslandModeConfig` interface with enabled, storagePath, and maxOfflineHours settings
- `QueuedUpdate` interface with id, timestamp, update, proof, and sequenceNumber
- `IslandStatus` interface for monitoring offline queue depth and storage usage
- `queueUpdate()` method for buffering updates during network partitions
- `sync()` method for reconciling queued updates after reconnection
- `verifyIntegrity()` method with cryptographic chain validation
- `flush()` method for clearing queue after successful sync
- Level-based persistent storage with ULID generation for unique entry IDs
- Hash-chain construction with tamper detection using SHA-256

### Test Coverage
- 1 integration test covering full lifecycle: queue → verify → sync → flush
- 88.4% line coverage with key paths well tested

### Architecture Notes
- Uses Level for ACID-compliant record storage
- Hash chain: `chainedHash = SHA256(entryHash + previousHash)` 
- Sequence numbers for ordering verification during recovery
- Automatic chain state persistence in database
- Fallback to soft recovery on chain integrity failure

---

## Security Considerations

- Hash chain provides tamper evidence but not cryptographic proof
- Storage is not encrypted; consider disk-level encryption for production
- Level database requires proper cleanup to avoid memory leaks

## Version Support

| Version | TypeScript | Node | Status |
|---------|-----------|------|--------|
| 0.2.0+ | 5.3+ | 18+ | Stable (in development) |
| 0.1.0-alpha | 5.3+ | 18+ | Alpha (superseded) |

---

## API Stability

See [SDK_API_STABILITY.md](../SDK_API_STABILITY.md) for version guarantees.
