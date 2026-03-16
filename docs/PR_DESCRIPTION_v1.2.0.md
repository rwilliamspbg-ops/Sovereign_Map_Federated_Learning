## Summary

This PR delivers the v1.2.0 wallet-ready blockchain testnet package. It adds a full blockchain core, wallet transaction lifecycle, smart contract runtime, governance execution path, FL node integration, and a repeatable testnet readiness gate.

## What Changed

### Blockchain and Wallet

- Added internal/blockchain core modules:
  - block and chain primitives
  - mempool
  - state database
  - validator set and proposer
  - wallet and wallet ledger
- Added signed transfer/stake/unstake/reward/FL round transaction support.

### Smart Contracts and Governance

- Added smart contract VM contracts and runtime execution path.
- Added governance proposal, vote, and execute capabilities.
- Added policy execution hooks to apply governance-approved reputation policy updates.

### Consensus and Node Integration

- Integrated consensus commit path with blockchain block proposal/commit and governance execution.
- Added FL node wallet integration for signed FL submissions and wallet-aware rewards.
- Added node pool integration and multi-node round handling.

### Readiness and Operations

- Added testnet wallet readiness script: scripts/testnet-wallet-readiness.sh
- Added Make target: make testnet-wallet-readiness
- Added testnet wallet readiness runbook.
- Added validator/policy metrics exporter support.

## Validation Performed

- go test ./internal/blockchain/... ./internal/blockchain/vm ...
- go test ./internal/node/...
- go test ./internal/consensus/...
- go build ./cmd/metrics-exporter
- readiness gate: make testnet-wallet-readiness

## Release Metadata

- Tag: v1.2.0
- Commit: 2110f64
- Scope: 43 files changed, 14231 insertions, 2 deletions

## Rollout Notes

- Run make testnet-wallet-readiness before each testnet promotion.
- Use governance-driven policy updates to tune validator weighting safely during testnet.
- Monitor validator and governance counters from metrics exporter endpoints.
