# Sovereign Map Federated Learning

Production-grade federated learning platform that combines Byzantine-resilient aggregation, trust verification, governance policy controls, and full-stack observability.

## Badges

### Quality and Delivery

[![Build and Test](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/build.yml?branch=main&label=Build%20and%20Test)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/build.yml)
[![Lint Code Base](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/lint.yml?branch=main&label=Lint)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/lint.yml)
[![HIL Tests](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/hil-tests.yml?branch=main&label=HIL%20Tests)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/hil-tests.yml)
[![Reproducibility Check](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/reproducibility-check.yml?branch=main&label=Reproducibility)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/reproducibility-check.yml)

### Security and Governance

[![CodeQL Security Analysis](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/codeql-analysis.yml?branch=main&label=CodeQL)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/codeql-analysis.yml)
[![Security Supply Chain](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/security-supply-chain.yml?branch=main&label=Supply%20Chain)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/security-supply-chain.yml)
[![Secret Scan](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/secret-scan.yml?branch=main&label=Secret%20Scan)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/secret-scan.yml)
[![Governance Check](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/governance-check.yml?branch=main&label=Governance)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/governance-check.yml)
[![Workflow Action Pin Check](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/workflow-action-pin-check.yml?branch=main&label=Action%20Pinning)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/workflow-action-pin-check.yml)

### Operations and Release

[![Observability CI](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/observability-ci.yml?branch=main&label=Observability)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/observability-ci.yml)
[![Build and Deploy](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/deploy.yml?branch=main&label=Build%20and%20Deploy)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/deploy.yml)
[![Docker Build](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/docker-build.yml?branch=main&label=Docker%20Build)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/docker-build.yml)
[![License](https://img.shields.io/github/license/rwilliamspbg-ops/Sovereign_Map_Federated_Learning)](LICENSE)

## Technical Brief

Sovereign Map Federated Learning is a dual-plane system:

1. Aggregation plane: Flower-based federated coordination with Byzantine-robust aggregation.
2. Control and telemetry plane: Flask APIs for health, HUD data, trust status, policy controls, and metrics publication.

Core characteristics:

- Byzantine-resilient training strategy with convergence tracking.
- Trust and verification APIs for attestation-style workflows.
- Governance and policy interfaces for runtime safety controls.
- Prometheus-compatible metrics exporters and Grafana-ready dashboard provisioning.
- Multi-profile deployment via Docker Compose and Kubernetes manifests.
- Hardware-aware test coverage including NPU, XPU, CUDA/ROCm, MPS, and CPU fallback paths.

## System Layout

- Backend aggregation and APIs: [sovereignmap_production_backend_v2.py](sovereignmap_production_backend_v2.py)
- Tokenomics metrics exporter: [tokenomics_metrics_exporter.py](tokenomics_metrics_exporter.py)
- TPM metrics exporter: [tpm_metrics_exporter.py](tpm_metrics_exporter.py)
- Frontend HUD: [frontend/src/HUD.jsx](frontend/src/HUD.jsx)
- Compose profiles: [docker-compose.dev.yml](docker-compose.dev.yml), [docker-compose.production.yml](docker-compose.production.yml), [docker-compose.full.yml](docker-compose.full.yml)
- Kubernetes scale profile: [kubernetes-5000-node-manifests.yaml](kubernetes-5000-node-manifests.yaml)

## Capability Map

| Domain | Runtime Surfaces | Purpose |
| --- | --- | --- |
| Federated learning | [sovereignmap_production_backend_v2.py](sovereignmap_production_backend_v2.py), [src/client.py](src/client.py) | Round orchestration, aggregation, convergence |
| Trust and attestation | [tpm_cert_manager.py](tpm_cert_manager.py), [tpm_metrics_exporter.py](tpm_metrics_exporter.py), [secure_communication.py](secure_communication.py) | Identity, verification, trust signals |
| Governance and policy | [bridge-policies.json](bridge-policies.json), [capabilities.json](capabilities.json) | Runtime controls and policy surfaces |
| Tokenomics and economics | [tokenomics_metrics_exporter.py](tokenomics_metrics_exporter.py), [tokenomics_metrics_exporter.py](tokenomics_metrics_exporter.py) | Economic telemetry and dashboard inputs |
| Observability | [prometheus.yml](prometheus.yml), [alertmanager.yml](alertmanager.yml), [fl_slo_alerts.yml](fl_slo_alerts.yml) | Metrics collection, alerting, SLO validation |
| Operations | [deploy.sh](deploy.sh), [deploy_demo.sh](deploy_demo.sh), [Makefile](Makefile) | Deployment and repeatable operator workflows |

## Quick Start

### Prerequisites

- Go 1.25+
- Node.js 20+
- npm 10+
- Python 3.11+
- Docker with Compose plugin

### Option A: Genesis bootstrap

```bash
git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning.git
cd Sovereign_Map_Federated_Learning
./genesis-launch.sh
```

### Option B: Local dev stack

```bash
docker compose -f docker-compose.dev.yml up -d
docker compose ps
```

### Option C: Full profile with participant scaling

```bash
docker compose -f docker-compose.full.yml up -d --scale node-agent=5
```

## Build and Validation Commands

```bash
# Go and backend tests
go test ./... -count=1

# Monorepo package build and tests
npm ci
npm run build:libs
npm run test:ci

# Frontend build
npm --prefix frontend ci
npm --prefix frontend run build
```

## API Surfaces (Operational)

Representative endpoints from the backend and exporters:

- GET /health
- GET /status
- GET /hud_data
- POST /trigger_fl
- POST /create_enclave
- POST /event/tokenomics
- GET /metrics

See implementation in [sovereignmap_production_backend_v2.py](sovereignmap_production_backend_v2.py) and [tokenomics_metrics_exporter.py](tokenomics_metrics_exporter.py).

## Deployment Profiles

- Development: [docker-compose.dev.yml](docker-compose.dev.yml)
- Production: [docker-compose.production.yml](docker-compose.production.yml)
- Full topology: [docker-compose.full.yml](docker-compose.full.yml)
- Monitoring stack: [docker-compose.monitoring.yml](docker-compose.monitoring.yml)
- Large-scale variants: [docker-compose.large-scale.yml](docker-compose.large-scale.yml), [docker-compose.1000nodes.yml](docker-compose.1000nodes.yml), [docker-compose.200nodes.yml](docker-compose.200nodes.yml)

## Repository Standards

- Contribution guidelines: [CONTRIBUTING.md](CONTRIBUTING.md)
- Security policy: [SECURITY.md](SECURITY.md)
- Changelog: [CHANGELOG.md](CHANGELOG.md)
- License: [LICENSE](LICENSE)

## Sanity Report

Timestamp: 2026-03-19

### Functional sanity checks completed

- Tokenomics exporter handles directory-valued source path safely and persists payload without IsADirectoryError.
- HUD simulation controls are wired end-to-end (frontend action -> backend endpoint -> HUD counter surface).
- Backend and exporter modules compile successfully with Python syntax checks.

### CI sanity checks completed

Verified green on main after latest changes:

- Build and Test
- Lint Code Base
- HIL Tests
- Reproducibility Check
- Governance Check
- Workflow Action Pin Check
- CodeQL Security Analysis
- Security Supply Chain
- Secret Scan
- Observability CI
- Build and Deploy

### Operational recommendation

For release candidates, run one additional live smoke test using [docker-compose.full.yml](docker-compose.full.yml) and verify /health, /hud_data, and /event/tokenomics before tagging.

Standalone report: [SANITY_REPORT.md](SANITY_REPORT.md)
