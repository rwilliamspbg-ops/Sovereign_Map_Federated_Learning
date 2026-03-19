<!-- markdownlint-disable MD013 -->

# Sovereign Map Federated Learning

Production-grade federated learning platform that combines Byzantine-resilient aggregation, trust verification, governance policy controls, tokenomics telemetry, and full-stack observability.

## Live Project Pulse

[![Release](https://img.shields.io/github/v/release/rwilliamspbg-ops/Sovereign_Map_Federated_Learning?display_name=tag&style=flat-square&logo=github)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/releases)
[![Last Commit](https://img.shields.io/github/last-commit/rwilliamspbg-ops/Sovereign_Map_Federated_Learning?style=flat-square&logo=git)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/commits/main)
[![Repo Size](https://img.shields.io/github/repo-size/rwilliamspbg-ops/Sovereign_Map_Federated_Learning?style=flat-square&logo=files)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning)
[![Contributors](https://img.shields.io/github/contributors/rwilliamspbg-ops/Sovereign_Map_Federated_Learning?style=flat-square&logo=github)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/graphs/contributors)
[![Issues](https://img.shields.io/github/issues/rwilliamspbg-ops/Sovereign_Map_Federated_Learning?style=flat-square&logo=github)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/issues)
[![Pull Requests](https://img.shields.io/github/issues-pr/rwilliamspbg-ops/Sovereign_Map_Federated_Learning?style=flat-square&logo=github)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/pulls)
[![Stars](https://img.shields.io/github/stars/rwilliamspbg-ops/Sovereign_Map_Federated_Learning?style=flat-square&logo=github)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/stargazers)
[![Forks](https://img.shields.io/github/forks/rwilliamspbg-ops/Sovereign_Map_Federated_Learning?style=flat-square&logo=github)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/network/members)
[![License](https://img.shields.io/github/license/rwilliamspbg-ops/Sovereign_Map_Federated_Learning?style=flat-square)](LICENSE)

## Platform Capability Badges

[![Capability: Federated Learning](https://img.shields.io/badge/Capability-Federated%20Learning-2f9e44?style=flat-square)](sovereignmap_production_backend_v2.py)
[![Capability: Governance Policy](https://img.shields.io/badge/Capability-Governance%20Policy-5f3dc4?style=flat-square)](bridge-policies.json)
[![Capability: Trust and Attestation](https://img.shields.io/badge/Capability-Trust%20and%20Attestation-0b7285?style=flat-square)](tpm_cert_manager.py)
[![Capability: Tokenomics Telemetry](https://img.shields.io/badge/Capability-Tokenomics%20Telemetry-1971c2?style=flat-square)](tokenomics_metrics_exporter.py)
[![Capability: Secure Communication](https://img.shields.io/badge/Capability-Secure%20Communication-087f5b?style=flat-square)](secure_communication.py)
[![Capability: Observability](https://img.shields.io/badge/Capability-Observability-e67700?style=flat-square)](prometheus.yml)
[![Capability: Large Scale Profiles](https://img.shields.io/badge/Capability-Large--Scale%20Deployments-1c7ed6?style=flat-square)](kubernetes-5000-node-manifests.yaml)

## CI, Security, and Release Badges

### Core Quality Gates

[![Build and Test](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/build.yml?branch=main&style=flat-square&label=Build%20and%20Test)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/build.yml)
[![Lint Code Base](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/lint.yml?branch=main&style=flat-square&label=Lint%20Code%20Base)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/lint.yml)
[![HIL Tests](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/hil-tests.yml?branch=main&style=flat-square&label=HIL%20Tests)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/hil-tests.yml)
[![Reproducibility Check](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/reproducibility-check.yml?branch=main&style=flat-square&label=Reproducibility)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/reproducibility-check.yml)
[![Observability CI](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/observability-ci.yml?branch=main&style=flat-square&label=Observability%20CI)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/observability-ci.yml)
[![OpenCV Go Tests](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/opencv-go-tests.yml?branch=main&style=flat-square&label=OpenCV%20Go%20Tests)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/opencv-go-tests.yml)

### Security and Governance Gates

[![CodeQL Security Analysis](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/codeql-analysis.yml?branch=main&style=flat-square&label=CodeQL)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/codeql-analysis.yml)
[![Security Supply Chain](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/security-supply-chain.yml?branch=main&style=flat-square&label=Security%20Supply%20Chain)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/security-supply-chain.yml)
[![Secret Scan](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/secret-scan.yml?branch=main&style=flat-square&label=Secret%20Scan)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/secret-scan.yml)
[![Governance Check](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/governance-check.yml?branch=main&style=flat-square&label=Governance%20Check)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/governance-check.yml)
[![Workflow Action Pin Check](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/workflow-action-pin-check.yml?branch=main&style=flat-square&label=Action%20Pin%20Check)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/workflow-action-pin-check.yml)
[![Audit Check](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/audit-check.yml?branch=main&style=flat-square&label=Audit%20Check)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/audit-check.yml)

### SDK and Release Engineering

[![SDK Security](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/sdk-security.yml?branch=main&style=flat-square&label=SDK%20Security)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-security.yml)
[![SDK Version](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/sdk-version.yml?branch=main&style=flat-square&label=SDK%20Version)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-version.yml)
[![SDK Publish](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/sdk-publish.yml?branch=main&style=flat-square&label=SDK%20Publish)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-publish.yml)
[![SDK Provenance](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/sdk-provenance.yml?branch=main&style=flat-square&label=SDK%20Provenance)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-provenance.yml)
[![SDK Channels](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/sdk-channels.yml?branch=main&style=flat-square&label=SDK%20Channels)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-channels.yml)
[![Contributor Rankings](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/contributor-rankings.yml?branch=main&style=flat-square&label=Contributor%20Rankings)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/contributor-rankings.yml)
[![Docs Link Check](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/docs-link-check.yml?branch=main&style=flat-square&label=Docs%20Link%20Check)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/docs-link-check.yml)
[![Test Artifacts Review](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/test-artifacts-review.yml?branch=main&style=flat-square&label=Test%20Artifacts%20Review)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/test-artifacts-review.yml)

### Deployment and Packaging

[![Build and Deploy](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/deploy.yml?branch=main&style=flat-square&label=Build%20and%20Deploy)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/deploy.yml)
[![Docker Build](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/docker-build.yml?branch=main&style=flat-square&label=Docker%20Build)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/docker-build.yml)
[![Windows Client EXE](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/windows-client-exe.yml?branch=main&style=flat-square&label=Windows%20Client%20EXE)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/windows-client-exe.yml)
[![Phase3D Production Deploy](https://img.shields.io/github/actions/workflow/status/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/phase3d-production-deploy.yml?branch=main&style=flat-square&label=Phase3D%20Production%20Deploy)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/phase3d-production-deploy.yml)

## Technical Brief

Sovereign Map Federated Learning is a dual-plane runtime:

1. Aggregation plane: Flower-based federated coordination with Byzantine-robust strategy logic and convergence tracking.
2. Control and telemetry plane: Flask services for health, HUD, trust/policy operations, join lifecycle, and metrics publication.

Core characteristics:

- Byzantine-resilient training strategy with runtime convergence history.
- Trust and verification APIs for attestation-style governance workflows.
- Policy and join-management endpoints for operator-controlled enrollment.
- Prometheus-compatible metrics exporters for operational and tokenomics surfaces.
- Multi-profile deployment via Docker Compose and Kubernetes manifests.
- Hardware-aware tests spanning NPU, XPU, CUDA/ROCm, MPS, and CPU fallbacks.

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

## Detailed Functions Reference

### Backend API Functions

| Endpoint | Method | Function | Responsibility |
| --- | --- | --- | --- |
| /health | GET | health | Service health, enclave status, HUD telemetry snapshot |
| /status | GET | status | Aggregator runtime identity and core port map |
| /chat | POST | chat_query | HUD assistant query handling for operator prompts |
| /hud_data | GET | hud_data | HUD metrics including audit accuracy and simulation counters |
| /founders | GET | get_founders | Founding-signature identity list for governance views |
| /trigger_fl | POST | trigger_fl_round | Manual FL round simulation and convergence updates |
| /create_enclave | POST | create_enclave | Enclave state transition workflow |
| /convergence | GET | get_convergence | Convergence history arrays for charting |
| /metrics_summary | GET | metrics_summary | Aggregated metrics summary across runtime domains |
| /model_registry | GET | model_registry_recent | Recent persisted model metadata and round snapshots |
| /simulate/<simulation_type> | POST | trigger_hud_simulation | Records HUD simulation events by scenario type |

### Trust, Policy, and Join Lifecycle Functions

| Endpoint | Method | Function | Responsibility |
| --- | --- | --- | --- |
| /trust_snapshot | GET | trust_snapshot | Current trust mode, policy state, and policy history |
| /verification_policy | POST | update_verification_policy | Runtime policy update surface |
| /llm_policy | GET | llm_policy_view | Exposes active LLM adapter validation policy |
| /join/policy | GET | join_policy_view | Join bootstrap policy and onboarding constraints |
| /join/invite | POST | create_join_invite | Issue join invites with bounded TTL and permissions |
| /join/register | POST | register_join_participant | Register participant certificates and join metadata |
| /join/registrations | GET | list_join_registrations | Admin listing of registered participants |
| /join/revoke/<int:node_id> | POST | revoke_join_participant | Administrative revocation of participant certificate |

### Training Control Functions

| Endpoint | Method | Function | Responsibility |
| --- | --- | --- | --- |
| /training/start | POST | start_training | Trigger training start signal for HUD/ops flows |
| /training/stop | POST | stop_training | Trigger training halt signal |
| /training/status | GET | training_status | Current mocked training progress and metrics view |

### Tokenomics Exporter Functions

| Endpoint | Method | Function | Responsibility |
| --- | --- | --- | --- |
| /metrics | GET | metrics | Prometheus exposition endpoint for tokenomics gauges |
| /health | GET | health | Tokenomics exporter liveness and source-file metadata |
| /event/tokenomics | POST | event_tokenomics | Ingest tokenomics events and persist canonical payload |

### Key Internal Runtime Functions

| Function | File | Responsibility |
| --- | --- | --- |
| validate_llm_adapter_update | [sovereignmap_production_backend_v2.py](sovereignmap_production_backend_v2.py) | Policy validation gate for incoming client updates |
| build_tokenomics_payload | [sovereignmap_production_backend_v2.py](sovereignmap_production_backend_v2.py) | Constructs tokenomics publication payload from FL state |
| publish_tokenomics_event | [sovereignmap_production_backend_v2.py](sovereignmap_production_backend_v2.py) | Sends tokenomics telemetry to exporter endpoint |
| publish_tpm_attestation_events | [sovereignmap_production_backend_v2.py](sovereignmap_production_backend_v2.py) | Emits attestation events for trust metrics pipeline |
| run_flower_server | [sovereignmap_production_backend_v2.py](sovereignmap_production_backend_v2.py) | Starts and configures Flower aggregation server |
| run_flask_metrics | [sovereignmap_production_backend_v2.py](sovereignmap_production_backend_v2.py) | Starts Flask API plane for control and telemetry |
| create_app | [tokenomics_metrics_exporter.py](tokenomics_metrics_exporter.py) | Constructs exporter app and endpoint bindings |

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
