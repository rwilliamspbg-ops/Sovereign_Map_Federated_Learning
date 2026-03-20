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

## Quick Architecture Overview

Sovereign Map uses a streaming aggregation model instead of loading full model updates into memory at once.

- Memory efficiency: Mohawk-style chunked processing reduces memory pressure by up to 224x for large update sets.
- Byzantine resilience: selective verification and trust scoring reduce adversarial impact with sublinear validation behavior for high node counts.
- Hardware root of trust: every node contributes attestation and certificate telemetry into the same operational control plane.

```mermaid
flowchart LR
    A[Client Updates] --> B{Traditional FL Aggregator}
    B --> C[Load full model deltas in memory]
    C --> D[High RAM footprint per round]

    A --> E{SovereignMap Mohawk Stream Aggregator}
    E --> F[Chunk updates into streaming windows]
    F --> G[Validate trust and policy per chunk]
    G --> H[Aggregate incrementally]
    H --> I[Low steady-state memory use]
```

## Why Mohawk

Mohawk-style streaming aggregation treats model updates as a continuous stream of chunks rather than a monolithic tensor payload. This allows the coordinator to perform verification, filtering, and merge steps incrementally while retaining bounded working memory. In practice, this is what makes high fan-out node participation feasible on commodity infrastructure: memory usage scales with chunk window size instead of full global update size, while trust and policy checks run inline with aggregation.

## Platform Capability Badges

[![Capability: Federated Learning](https://img.shields.io/badge/Capability-Federated%20Learning-2f9e44?style=flat-square)](sovereignmap_production_backend_v2.py)
[![Capability: Governance Policy](https://img.shields.io/badge/Capability-Governance%20Policy-5f3dc4?style=flat-square)](bridge-policies.json)
[![Capability: Trust and Attestation](https://img.shields.io/badge/Capability-Trust%20and%20Attestation-0b7285?style=flat-square)](tpm_cert_manager.py)
[![Capability: Tokenomics Telemetry](https://img.shields.io/badge/Capability-Tokenomics%20Telemetry-1971c2?style=flat-square)](tokenomics_metrics_exporter.py)
[![Capability: Secure Communication](https://img.shields.io/badge/Capability-Secure%20Communication-087f5b?style=flat-square)](secure_communication.py)
[![Capability: Observability](https://img.shields.io/badge/Capability-Observability-e67700?style=flat-square)](prometheus.yml)
[![Capability: Large Scale Profiles](https://img.shields.io/badge/Capability-Large--Scale%20Deployments-1c7ed6?style=flat-square)](kubernetes-5000-node-manifests.yaml)

## Device Support Badges

[![Web](https://img.shields.io/badge/Device-Web%20Dashboard-0ea5e9?style=flat-square&logo=google-chrome&logoColor=white)](frontend/src/HUD.jsx)
[![Linux](https://img.shields.io/badge/OS-Linux%20Supported-f59e0b?style=flat-square&logo=linux&logoColor=white)](docker-compose.dev.yml)
[![macOS](https://img.shields.io/badge/OS-macOS%20Supported-111827?style=flat-square&logo=apple&logoColor=white)](README.md#quick-start)
[![Windows](https://img.shields.io/badge/OS-Windows%20Supported-2563eb?style=flat-square&logo=windows&logoColor=white)](run-demo-windows.ps1)
[![Android](https://img.shields.io/badge/Mobile-Android%20Supported-16a34a?style=flat-square&logo=android&logoColor=white)](mobile-apps/android-node-app)
[![iOS](https://img.shields.io/badge/Mobile-iOS%20Supported-0f172a?style=flat-square&logo=apple&logoColor=white)](mobile-apps/ios-node-app)
[![Raspberry Pi](https://img.shields.io/badge/Edge-Raspberry%20Pi%20Supported-c026d3?style=flat-square&logo=raspberry-pi&logoColor=white)](README.md#hardware-requirements)
[![NVIDIA Jetson](https://img.shields.io/badge/Edge-NVIDIA%20Jetson%20Supported-15803d?style=flat-square&logo=nvidia&logoColor=white)](README.md#hardware-requirements)
[![x86_64](https://img.shields.io/badge/Arch-x86__64%20Supported-1d4ed8?style=flat-square)](README.md#hardware-requirements)
[![ARM64](https://img.shields.io/badge/Arch-ARM64%20Supported-9333ea?style=flat-square)](README.md#hardware-requirements)
[![Docker](https://img.shields.io/badge/Runtime-Docker%20Compose%20Supported-0284c7?style=flat-square&logo=docker&logoColor=white)](docker-compose.full.yml)

## CI, Security, and Release Badges

### Core Quality Gates

[![Build and Test](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/build.yml)
[![Lint Code Base](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/lint.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/lint.yml)
[![HIL Tests](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/hil-tests.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/hil-tests.yml)
[![Reproducibility Check](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/reproducibility-check.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/reproducibility-check.yml)
[![Observability CI](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/observability-ci.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/observability-ci.yml)
[![OpenCV Go Tests](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/opencv-go-tests.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/opencv-go-tests.yml)

### Security and Governance Gates

[![CodeQL Security Analysis](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/codeql-analysis.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/codeql-analysis.yml)
[![Security Supply Chain](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/security-supply-chain.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/security-supply-chain.yml)
[![Secret Scan](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/secret-scan.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/secret-scan.yml)
[![Governance Check](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/governance-check.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/governance-check.yml)
[![Workflow Action Pin Check](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/workflow-action-pin-check.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/workflow-action-pin-check.yml)
[![Audit Check](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/audit-check.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/audit-check.yml)

### SDK and Release Engineering

[![SDK Security](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-security.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-security.yml)
[![SDK Version](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-version.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-version.yml)
[![SDK Publish](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-publish.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-publish.yml)
[![SDK Provenance](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-provenance.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-provenance.yml)
[![SDK Channels](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-channels.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-channels.yml)
[![Contributor Rankings](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/contributor-rankings.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/contributor-rankings.yml)
[![Docs Link Check](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/docs-link-check.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/docs-link-check.yml)
[![Test Artifacts Review](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/test-artifacts-review.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/test-artifacts-review.yml)

### Deployment and Packaging

[![Build and Deploy](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/deploy.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/deploy.yml)
[![Docker Build](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/docker-build.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/docker-build.yml)
[![Windows Client EXE](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/windows-client-exe.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/windows-client-exe.yml)
[![Phase 3D Production Deploy](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/phase3d-production-deploy.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/phase3d-production-deploy.yml)

## What To Use This Software For

- Running secure, federated ML training across distributed nodes where raw data must stay local.
- Operating Byzantine-resilient model aggregation in adversarial or partially trusted environments.
- Building trust-aware AI infrastructure with policy controls, attestation signals, and auditable telemetry.
- Monitoring real-time FL, tokenomics, and system health through Prometheus and Grafana surfaces.
- Prototyping and scaling from local Docker deployments to large Compose/Kubernetes profiles.

## Hardware Requirements

| Node Class | Minimum (Functional) | Recommended (Sustained) |
| --- | --- | --- |
| Edge CPU Node | Raspberry Pi 4 (4 GB RAM), 4-core ARM CPU, 32 GB storage, Linux, TPM 2.0 device access | Raspberry Pi 5 / x86 mini PC (8-16 GB RAM), NVMe storage, TPM 2.0, stable wired network |
| Edge GPU/NPU Node | Jetson Nano / Intel NPU-capable edge device, 8 GB RAM, CUDA/NPU drivers | NVIDIA Jetson Orin / equivalent, 16+ GB RAM, tuned CUDA/NPU stack |
| Operator / Aggregator | 8 vCPU, 16 GB RAM, SSD, Docker Compose | 16+ vCPU, 32-64 GB RAM, NVMe, GPU optional, isolated monitoring host |
| Monitoring Stack | 2 vCPU, 4 GB RAM for Prometheus + Grafana | 4-8 vCPU, 8-16 GB RAM with longer retention and dashboard concurrency |

Use [hardware_auto_tuner.py](hardware_auto_tuner.py) to auto-profile host capability and choose an acceleration profile before large-scale runs.

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

## Visual Walkthrough

Visual proof for this project should be treated as release evidence, not optional decoration.

Expected screenshot artifacts per release:

- Operations HUD: trust score, node participation, latency wall, and resilience indicators.
- Grafana Operations Overview: gauge deck + trend wall under live load.
- Grafana Tokenomics Overview: mint/bridge/validator/wallet health sections.

Tracked asset locations:

- `docs/screenshots/hud-operations-overview.png`
- `docs/screenshots/grafana-operations-overview.png`
- `docs/screenshots/grafana-tokenomics-overview.png`

Capture workflow and acceptance checklist:

- [docs/screenshots/README.md](docs/screenshots/README.md)

Current status: screenshot paths are defined and release capture workflow is documented; attach rendered PNG/GIF evidence in each tagged release.

## Dual-Plane Runtime Data Flow

```mermaid
sequenceDiagram
    participant Node as Edge Node Client
    participant Flower as Flower Aggregation Plane (:8080)
    participant Mohawk as Mohawk Stream Aggregator
    participant Policy as Trust/Policy Gate
    participant API as Control Plane API (:8000)
    participant Prom as Prometheus
    participant Grafana as Grafana/HUD

    Node->>Flower: Submit model update (FitRes)
    Flower->>Mohawk: Forward update chunks
    Mohawk->>Policy: Validate adapter policy + attestation metadata
    Policy-->>Mohawk: Accept/Reject + reason labels
    Mohawk->>Flower: Incremental aggregate result
    Flower->>API: Publish round metrics + convergence snapshot
    API->>Prom: Expose /metrics and event-derived gauges/counters
    Prom->>Grafana: Scrape telemetry
    API->>Grafana: Serve /health, /ops/health, /hud_data, /metrics_summary
```

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

Live API examples and integration snippets:

- [docs/api/http-examples.md](docs/api/http-examples.md)

OpenAPI/Postman status:

- A full OpenAPI specification is not yet published in-repo.
- Until that lands, use the HTTP example catalog above as the canonical integration quick start.

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
| /ops/health | GET | ops_health | Operational dependency/system snapshot (ports, memory/disk pressure, Prometheus reachability) |
| /ops/events/recent | GET | ops_events_recent | Returns recent operations events for timeline replay |
| /ops/events | GET (SSE) | ops_events_stream | Server-sent event stream for live operations telemetry |

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

### TPM Exporter Functions

| Endpoint | Method | Function | Responsibility |
| --- | --- | --- | --- |
| /metrics | GET | metrics | Prometheus exposition endpoint for TPM/trust metrics |
| /metrics/summary | GET | metrics_summary | Aggregated TPM/trust summary snapshot |
| /health | GET | health | TPM exporter liveness and metadata |
| /event/attestation | POST | event_attestation | Ingest attestation event payloads |
| /event/message | POST | event_message | Ingest trust-related operational messages |

### Endpoint Contract Notes

- Auth boundaries: `/join/registrations` and `/join/revoke/<int:node_id>` are admin-gated and require valid admin authorization headers.
- Auth boundaries: `/verification_policy` supports role-aware updates via `X-API-Role` and optional bearer token wiring.
- Status code behavior: `/create_enclave` may return `202` while provisioning is in progress, then `200` once a stable state transition is reached.
- Status code behavior: `/trigger_fl` may return `202` for accepted async execution and non-2xx when round execution cannot proceed.
- Streaming semantics: `/ops/events` is an SSE endpoint and includes heartbeat events to keep long-lived clients connected.
- Streaming semantics: `/ops/events/recent` should be used to backfill timeline state before attaching to SSE.

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

### Verify stack health (required)

```bash
curl -s http://localhost:8000/status | jq
curl -s http://localhost:8000/health | jq
curl -s http://localhost:8000/ops/health | jq
curl -s http://localhost:8000/training/status | jq
```

Expected checkpoints:

- `/status` returns service identity and ports.
- `/ops/health` reports API, Flower, and Prometheus reachability.
- frontend HUD is reachable at `http://localhost:3000`.
- Grafana is reachable at `http://localhost:3001`.

### First training round (Hello World)

CLI flow:

```bash
# Trigger one global round
curl -s -X POST http://localhost:8000/trigger_fl | jq

# Verify round advanced and metrics updated
curl -s http://localhost:8000/metrics_summary | jq '.federated_learning.current_round, .federated_learning.current_accuracy, .federated_learning.current_loss'
curl -s http://localhost:8000/convergence | jq '.current_round, .current_accuracy, .current_loss'
```

UI flow:

1. Open `http://localhost:3000`.
2. Switch to **Network Operations HUD**.
3. Click **Run Global FL Epoch**.
4. Confirm the live timeline shows a `TRAINING_ROUND` event and round metrics increment.

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

## Contributor First Steps

Before opening a PR, run the same fast checks maintainers use:

```bash
# Discover all available developer targets
make help

# Required baseline
make fmt
make lint
make test

# Recommended reproducibility smoke checks
make smoke
```

For runtime-focused changes (HUD, observability, policy endpoints), include at least one local verification artifact in your PR description:

- `/health` and `/ops/health` output snippet.
- one screenshot from HUD or Grafana.
- command log showing a successful `trigger_fl` round.

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

## Help Wanted: Quick Wins

If you want to contribute quickly, these areas have high impact and low setup friction:

- Test matrix expansion for TPM 2.0 hardware variants and Docker runtimes.
- Apple Silicon (MPS) acceleration optimization and benchmark baselines.
- Additional Grafana panel tuning for high-cardinality node fleets.
- Better synthetic fault workloads for Byzantine and partition simulation paths.

Contribution process and coding standards are in [CONTRIBUTING.md](CONTRIBUTING.md).

## Common Issues

### TPM device access in Docker (`/dev/tpm0`)

- Symptom: trust/attestation metrics stay flat or backend cannot initialize TPM flows.
- Check: container runtime must expose TPM devices and required permissions.
- Typical fix: run with explicit device mapping and appropriate group permissions for TPM access.

### Port conflicts (frontend/backend/observability)

- Symptom: HUD shows backend unreachable or Grafana/Prometheus endpoints fail to bind.
- Check: local services already using frontend/backend ports.
- Typical fix: align Compose port mappings and frontend API base configuration so HUD and backend targets match.

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
