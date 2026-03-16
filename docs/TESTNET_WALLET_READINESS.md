# Testnet Platform Readiness (Wallet and Governance)

This document defines the minimum release gate for a sovereign platform testnet deployment, including wallet, governance, and core consensus safety checks.

## One-Command Readiness

Run:

```bash
make testnet-wallet-readiness
```

This executes:

- Wallet-focused tests in blockchain package.
- Full blockchain, node, and consensus test suites.
- Build checks for wallet-enabled binaries.
- Docker Compose config validation for testnet deployment profiles.
- A markdown readiness report under `test-results/testnet-readiness/`.

Readiness is intentionally broader than wallet plumbing: it is the baseline operational gate for platform integrity before public testnet rollout.

## Manual Equivalent

```bash
go test ./internal/blockchain -run "Wallet|CreateTransfer|ApplyTransactionTransfer" -count=1
go test ./internal/blockchain/... -count=1
go test ./internal/node/... -count=1
go test ./internal/consensus/... -count=1
go build ./internal/blockchain/... ./internal/node/... ./internal/consensus/...
go build ./cmd/node-agent ./cmd/metrics-exporter
docker compose -f docker-compose.dev.yml config >/dev/null
docker compose -f docker-compose.production.yml config >/dev/null
docker compose -f docker-compose.participant.yml config >/dev/null
docker compose -f docker-compose.monitoring.yml config >/dev/null
```

## Exit Criteria

Status is READY only when:

- All mandatory test/build checks pass.
- No readiness FAIL entries are present in the generated report.

Warnings do not fail readiness, but must be reviewed before public testnet launch.
