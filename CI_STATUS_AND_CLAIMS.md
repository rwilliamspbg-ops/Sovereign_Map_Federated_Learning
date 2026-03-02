# CI Status and Claim Boundaries

This document defines what can be claimed from automation results and what still requires additional validation.

## Local Verification Snapshot (2026-03-02)

Latest local command results on this branch:

- `make lint` surfaced `golangci-lint` typecheck failures caused by missing modules:
	- `github.com/tetratelabs/wazero`
	- `github.com/tetratelabs/wazero/api`
- `go test ./...` failed due to:
	- Missing modules (`wazero`, `wazero/api`, `github.com/stretchr/testify/assert`)
	- Test/API drift in `internal/batch`, `internal/island`, `internal/p2p`, and `internal/tpm`

Claim constraint from this snapshot:

- Do not claim that the current local working tree has a fully passing Go lint/test state.
- Keep using CI workflow badges as branch-level evidence for `main`.

## Workflow Badges (main branch)

[![Build and Test](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/build.yml)
[![CodeQL Security Analysis](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/codeql-analysis.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/codeql-analysis.yml)
[![Lint Code Base](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/lint.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/lint.yml)
[![SGP-001 Audit Sync](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/audit-check.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/audit-check.yml)
[![HIL Tests](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/hil-tests.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/hil-tests.yml)
[![Docker Build](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/docker-build.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/docker-build.yml)
[![Build & Deploy to Production](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/deploy.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/deploy.yml)
[![SDK Publish](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-publish.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-publish.yml)

## Strongly Supported by CI

- Code builds and workflow jobs complete for the tracked branch/commit.
- Unit/integration suites wired into CI pass.
- Static analysis/lint/security gates configured in workflows pass.
- TPM emulation tests pass in CI via `swtpm`/`tpm2-tools`.
- NPU fallback logic tests pass (selection/fallback behavior, not accelerator throughput).

## Supported with Qualification

- Large-scale or extreme-node claims are benchmark artifacts and should be treated as scenario-dependent.
- Performance numbers are snapshots from specific environments, not universal guarantees.
- Security guidance aligns with known standards but is not a formal compliance certification.

## Not Proved by GitHub-Hosted CI Alone

- Physical TPM attestation semantics on real hardware.
- Real NPU offload performance/TOPS across target hardware classes.
- Multi-region production resilience, long-duration adversarial campaigns, or exhaustive red-team coverage.

## Claim Wording Policy

Use language like:

- "CI-verified"
- "Observed in benchmark artifacts"
- "Design target"
- "Emulated TPM validation"

Avoid language like:

- "Guaranteed"
- "Fully production-proved"
- "Formally compliant" (unless external certification evidence is provided)
