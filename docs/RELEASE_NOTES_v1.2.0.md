# Release Notes: v1.2.0 Wallet-Ready Testnet Package

Release date: 2026-03-16
Tag: v1.2.0
Commit: 2110f64

## Highlights

- Introduced a complete wallet-enabled blockchain core under internal/blockchain with transaction signing, transfer support, staking/unstaking, reward flows, and state-backed balances.
- Added smart contract runtime and execution pipeline with governance, rewards, and registry contract support.
- Integrated consensus and federated node flow with blockchain commit path and governance execution hooks.
- Added validator reputation, attestation, and participation-quality policy controls suitable for testnet governance tuning.
- Added testnet wallet readiness automation and runbook for repeatable release gating.

## Major Additions

### Blockchain Core

- Added block, chain, mempool, validator, proposer, state database, and wallet modules.
- Added transfer transaction type and wallet ledger with state-backed account balances.
- Added comprehensive tests across blockchain modules.

### Smart Contracts and Governance

- Added VM contracts and executor flow.
- Added governance proposal creation, voting, and execution path.
- Added consensus-to-governance execution integration for policy updates.

### FL Node and Pool Integration

- Added wallet-aware FL node behavior for signed FL round transactions.
- Added node pool orchestration tied to blockchain/consensus workflows.
- Added integration tests across multi-node round execution.

### Observability and Operations

- Extended metrics exporter with validator and policy metrics.
- Added testnet wallet readiness script and Make target.
- Added runbook docs for readiness checks.

## Validation

The release was validated with package tests and builds before tag/publish:

- internal/blockchain and internal/blockchain/vm tests
- internal/node tests
- internal/consensus tests
- cmd/metrics-exporter build
- wallet testnet readiness gate via scripts/testnet-wallet-readiness.sh

## Change Volume

- 43 files changed
- 14231 insertions
- 2 deletions

## Referenced Docs

- docs/BLOCKCHAIN_UPGRADE_ROADMAP.md
- docs/TESTNET_WALLET_READINESS.md
- BLOCKCHAIN_QUICK_START.md
- BLOCKCHAIN_TRANSFORMATION.md

## Upgrade Guidance

- Use make testnet-wallet-readiness before promoting deployments.
- Use governance policy execution hooks to tune validator weighting for testnet experiments.
- Monitor validator, policy, and governance metrics via the metrics exporter endpoints.
