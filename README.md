<!-- markdownlint-disable MD013 -->

# Sovereign Map v1.2.0 - Wallet-Ready Federated Learning Blockchain

Sovereign Map combines federated learning, a blockchain execution layer, and governance controls for testnet-scale experimentation.

Intellectual Property Notice: This project implements the Sovereign Mohawk Protocol. Portions of this technology are Patent Pending (U.S. Provisional Patent Application Filed March 2026).

## Current Release

- Latest tag: [v1.2.0](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/releases/tag/v1.2.0)
- Release notes: [docs/RELEASE_NOTES_v1.2.0.md](docs/RELEASE_NOTES_v1.2.0.md)
- PR package template: [docs/PR_DESCRIPTION_v1.2.0.md](docs/PR_DESCRIPTION_v1.2.0.md)

## What's New in v1.2.0

- Wallet-enabled blockchain core with signed transactions, transfer/stake/unstake flows, and state-backed balances.
- Smart contract VM and executor support for governance, rewards, and model registry paths.
- Governance proposal and execution wiring in consensus, including policy updates.
- Validator reputation plus attestation plus participation-quality scoring with anti-gaming controls.
- Expanded metrics exporter coverage for validator policy and governance execution outcomes.
- One-command testnet readiness gate for wallet-aware release validation.

## Quick Start

### Option A: Genesis Launch Stack

```bash
git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning.git
cd Sovereign_Map_Federated_Learning
./genesis-launch.sh
```

Default launch scale is 15 nodes. Override for one run:

```bash
MIN_NODES=25 ./genesis-launch.sh
```

### Option B: Docker Compose Testnet Stack

```bash
docker compose -f docker-compose.full.yml up -d --scale node-agent=5
docker compose ps
```

### Pre-Release Readiness Gate (Recommended)

```bash
make testnet-wallet-readiness
```

Readiness details and exit criteria are documented in [docs/TESTNET_WALLET_READINESS.md](docs/TESTNET_WALLET_READINESS.md).

## Architecture Summary

- Federated learning orchestration and node pool management.
- Wallet-aware blockchain transactions for FL rounds and token/state operations.
- Contract execution runtime for governance and protocol policy updates.
- Consensus coordinator with governance commit path.
- TPM-inspired trust and certificate-based secure communication.
- Monitoring stack with Prometheus, Grafana, and alerting rules.

## Validation and Quality Gates

Use these as the baseline local checks before publishing or promoting testnet changes:

```bash
go test ./internal/blockchain/... -count=1
go test ./internal/node/... -count=1
go test ./internal/consensus/... -count=1
go build ./cmd/metrics-exporter
make testnet-wallet-readiness
```

For release-specific guidance, see [docs/RELEASE_NOTES_v1.2.0.md](docs/RELEASE_NOTES_v1.2.0.md).

## Observability

Core monitoring and alerting assets:

- [prometheus.yml](prometheus.yml)
- [alertmanager.yml](alertmanager.yml)
- [fl_slo_alerts.yml](fl_slo_alerts.yml)
- [tpm_alerts.yml](tpm_alerts.yml)
- [docker-compose.monitoring.yml](docker-compose.monitoring.yml)

### Grafana Dashboard Status (March 2026)

- 11 dashboards are provisioned and live under the `Sovereign Map` folder.
- Visual readability pass applied across all dashboards:
  - Cleaner stat cards (value-forward, centered)
  - Better legend and tooltip consistency for time series
  - Unified tags for discoverability (`enhanced`, `wow`, `readable`)
- Tokenomics telemetry path now supports both modes:
  - Explicit tokenomics and wallet telemetry via `tokenomics-metrics`
  - Compatibility fallback series via `dashboard_compat_rules.yml`
- Genesis launch network status panel was fixed to valid PromQL and now returns data.

### Monitoring Deployment Notes

- Production stack file: [docker-compose.production.yml](docker-compose.production.yml)
- Dashboard definitions: [grafana/provisioning/dashboards](grafana/provisioning/dashboards)
- Tokenomics exporter: [tokenomics_metrics_exporter.py](tokenomics_metrics_exporter.py)
- Backend live tokenomics publisher: [sovereignmap_production_backend_v2.py](sovereignmap_production_backend_v2.py)

## Key Docs

- [README.md](README.md)
- [docs/BLOCKCHAIN_UPGRADE_ROADMAP.md](docs/BLOCKCHAIN_UPGRADE_ROADMAP.md)
- [docs/TESTNET_WALLET_READINESS.md](docs/TESTNET_WALLET_READINESS.md)
- [docs/RELEASE_NOTES_v1.2.0.md](docs/RELEASE_NOTES_v1.2.0.md)
- [CONTRIBUTING.md](CONTRIBUTING.md)
- [Documentation](Documentation)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution workflow and PR standards.

