# Sovereign Map Blockchain Upgrade Roadmap

## Goal
Make the network best-in-class for sovereign federated learning: secure, high-performance, and provably fair.

## Phase 1: Core Protocol Hardening (Now)
- Reputation-aware validator selection.
- Reputation-adjusted reward distribution.
- Reputation penalties on slashing and recovery over time.
- Test coverage for selection, slashing, and reward behavior.

## Phase 2: Security and Trust (Next)
- TPM/TEE attestation score integrated into validator reputation.
- Signed attestation evidence stored in state with periodic re-validation.
- Slashing policy for stale or invalid attestation.

## Phase 3: Privacy and Verifiability
- Zero-knowledge proof hooks for FL update validity.
- Privacy-preserving model quality proofs.
- On-chain verification metadata for submitted rounds.

## Phase 4: Performance and UX
- Mempool fairness improvements for anti-front-run ordering.
- Account abstraction style transaction batching.
- Gasless submit/claim flows for FL nodes.

## Phase 5: Interoperability and Ecosystem
- Checkpoint anchoring to external L1s.
- Cross-chain reward settlement adapters.
- Public benchmarking and adversarial resilience scorecards.

## Deliverables for Current Sprint
- Ship Phase 1 reputation feature.
- Add tests and validation commands in CI.
- Publish before/after validator behavior summary in docs.
