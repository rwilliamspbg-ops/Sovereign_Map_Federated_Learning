# Sovereign Map Platform Evolution Roadmap

## Goal

Make the network best-in-class for sovereign intelligence coordination: secure, high-performance, policy-driven, and economically resilient.

## Phase 1: Core Protocol and Trust Hardening (Now)

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

## Phase 4: Performance and Operator UX

- Mempool fairness improvements for anti-front-run ordering.
- Account abstraction style transaction batching.
- Gasless submit/claim flows for FL nodes.

## Phase 5: Platform Economy and Marketplace Rails

- Model and data exchange settlement primitives.
- Marketplace policy controls for access, pricing, and compliance.
- Revenue-sharing logic linked to reputation and contribution quality.

## Phase 6: Interoperability and Ecosystem

- Checkpoint anchoring to external L1s.
- Cross-chain reward settlement adapters.
- Public benchmarking and adversarial resilience scorecards.

## Deliverables for Current Sprint

- Ship Phase 1 reputation feature.
- Add tests and validation commands in CI.
- Publish before/after validator behavior summary in docs.
