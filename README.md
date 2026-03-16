<!-- markdownlint-disable MD013 -->

# Sovereign Map

[![Release](https://img.shields.io/github/v/release/rwilliamspbg-ops/Sovereign_Map_Federated_Learning?display_name=tag)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/releases)
[![Build and Test](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/build.yml)
[![Lint Code Base](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/lint.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/lint.yml)
[![Reproducibility Check](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/reproducibility-check.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/reproducibility-check.yml)
[![Governance Check](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/governance-check.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/governance-check.yml)
[![CodeQL Security Analysis](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/codeql-analysis.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/codeql-analysis.yml)
[![Docs Link Check](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/docs-link-check.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/docs-link-check.yml)

Sovereign Map is a sovereign intelligence coordination platform that combines federated learning, blockchain governance, trust attestation, and observability into one deployable stack.

Intellectual Property Notice: this repository implements parts of the Sovereign Mohawk Protocol. Portions are patent pending (U.S. Provisional Patent Application filed March 2026).

## Current Status (March 2026)

- Current release: [v1.2.0](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/releases/tag/v1.2.0)
- Default branch: `main`
- CI baseline: npm-based SDK build/test workflows are aligned with repository lockfiles
- Local baseline validation commands:
  - `npm ci`
  - `npm run build:libs`
  - `npm run test:ci`
  - `npm --prefix frontend run build`

## What Is Included

- Intelligence layer: federated training/inference orchestration and model lifecycle control
- Trust layer: TPM-style attestation, proof verification, secure transport, trust status snapshots
- Governance layer: proposal execution, policy updates, validator reputation and enforcement hooks
- Economic layer: wallet-aware blockchain state, staking flows, reward path integration
- Operations layer: Prometheus metrics, Grafana dashboards, alerting rules, reproducibility and governance CI gates

## Quick Start

### Prerequisites

- Go 1.25+
- Node.js 20+
- npm 10+
- Python 3.12+
- Docker + Docker Compose plugin

### Option A: Genesis Launch

```bash
git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning.git
cd Sovereign_Map_Federated_Learning
./genesis-launch.sh
```

Scale override for one run:

```bash
MIN_NODES=25 ./genesis-launch.sh
```

### Option B: Compose Stack

```bash
docker compose -f docker-compose.full.yml up -d --scale node-agent=5
docker compose ps
```

## Required Validation Before Promotion

```bash
go test ./internal/blockchain/... -count=1
go test ./internal/node/... -count=1
go test ./internal/consensus/... -count=1
go build ./cmd/metrics-exporter
npm ci
npm run build:libs
npm run test:ci
npm --prefix frontend ci
npm --prefix frontend run build
make testnet-wallet-readiness
```

## Operations and Monitoring

- Monitoring stack: [docker-compose.monitoring.yml](docker-compose.monitoring.yml)
- Prometheus config: [prometheus.yml](prometheus.yml)
- Alertmanager config: [alertmanager.yml](alertmanager.yml)
- FL SLO alerts: [fl_slo_alerts.yml](fl_slo_alerts.yml)
- TPM alerts: [tpm_alerts.yml](tpm_alerts.yml)
- Dashboard assets: [grafana/provisioning/dashboards](grafana/provisioning/dashboards)

## Documentation Map

- Main docs hub: [Documentation/README.md](Documentation/README.md)
- Project roadmap: [Documentation/Project/ROADMAP.md](Documentation/Project/ROADMAP.md)
- Latest readiness summary: [Documentation/Project/RELEASE_READINESS_2026-03-16.md](Documentation/Project/RELEASE_READINESS_2026-03-16.md)
- Release notes: [docs/RELEASE_NOTES_v1.2.0.md](docs/RELEASE_NOTES_v1.2.0.md)
- Wallet readiness gate: [docs/TESTNET_WALLET_READINESS.md](docs/TESTNET_WALLET_READINESS.md)
- Contribution standards: [CONTRIBUTING.md](CONTRIBUTING.md)

## Contributing

All changes that affect runtime behavior, security posture, CI gates, or deployment defaults must include documentation updates in the same PR. See [CONTRIBUTING.md](CONTRIBUTING.md).
