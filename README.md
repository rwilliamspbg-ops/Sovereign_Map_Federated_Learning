<!-- markdownlint-disable MD010 MD012 MD013 MD022 MD031 MD032 MD033 MD034 MD036 MD040 MD051 MD060 -->

# Sovereign Map v1.1.0 - Byzantine-Tolerant Federated Learning

## Genesis Block Launch - CI-Verified Testnet Stack

**Sovereign Map provides a CI-verified federated learning and trust-monitoring stack for testnet and research workloads.**
⚠️ Intellectual Property Notice: This project implements the Sovereign Mohawk Protocol. Portions of this technology are Patent Pending (U.S. Provisional Patent Application Filed March 2026).


### Quick Start (5 Minutes)

```bash
git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning.git
cd Sovereign_Map_Federated_Learning
./genesis-launch.sh
```

Default launch scale is 15 nodes. To override for a one-off run:

```bash
MIN_NODES=25 ./genesis-launch.sh
```

**Access Dashboards:**
- **Genesis Launch Overview**: [http://localhost:3001/d/genesis-launch-overview](http://localhost:3001/d/genesis-launch-overview)
- **Network Performance**: [http://localhost:3001/d/network-performance-health](http://localhost:3001/d/network-performance-health)
- **Consensus & Trust**: [http://localhost:3001/d/consensus-trust-monitoring](http://localhost:3001/d/consensus-trust-monitoring)

**Documentation:**
- **Quick Start**: [GENESIS_QUICK_START.md](/Documentation/Guides/GENESIS_QUICK_START.md) (5-minute guide)
- **Full Guide**: [GENESIS_LAUNCH_GUIDE.md](/Documentation/Deployment/GENESIS_LAUNCH_GUIDE.md) (comprehensive documentation)
- **Architecture**: [ARCHITECTURE.md](/Documentation/Architecture/ARCHITECTURE.md) (system design)
- **Reproducibility**: [REPRODUCIBLE_SETUP.md](/Documentation/Guides/REPRODUCIBLE_SETUP.md) (clone and smoke-check workflow)
- **Participant Local Join**: [PARTICIPANT_JOIN_LOCAL.md](/Documentation/Deployment/PARTICIPANT_JOIN_LOCAL.md)
- **Windows EXE Client**: [WINDOWS_CLIENT_EXE.md](/Documentation/Deployment/WINDOWS_CLIENT_EXE.md)

**Contributing:** See [CONTRIBUTING.md](CONTRIBUTING.md) for PR checklist, CodeQL guardrails, and branch protection recommendations.

## Release Update (2026-03)

- `v1.1.0` finalized and published from `main`: [Release v1.1.0](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/releases/tag/v1.1.0).

- P2P mesh updated with libp2p default transports (TCP + QUIC) and NAT traversal service.
- Distributed checkpoint storage expanded with IPFS backend support.
- Sensor ingestion expanded with camera, SLAM, mobile API, and drone telemetry modules.
- Runtime readiness and implementation details documented in:
  - [NETWORK_READINESS_ASSESSMENT.md](/Documentation/Deployment/NETWORK_READINESS_ASSESSMENT.md)
  - [IMPLEMENTATION_SUMMARY.md](/Documentation/Architecture/IMPLEMENTATION_SUMMARY.md)
  - [OPENCV_INSTALL.md](/Documentation/Guides/OPENCV_INSTALL.md)
  - [ROADMAP.md](/Documentation/Project/ROADMAP.md)
- Post-v1.1.0 upgrade stream completed for:
  - Real tokenomics telemetry exporter and compatibility rules
  - Event-driven TPM attestation/message metrics ingestion
  - FL SLO alert rules and observability CI guardrails
  - Model registry persistence endpoint and channel overlay launcher
  - Local participant onboarding API (invite/register/revoke) with cert issuance
  - Windows FL client launcher and automated EXE build pipeline

---

## Status Badges

### CI Workflows (main)
[![Build and Test](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/build.yml)
[![CodeQL Security Analysis](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/codeql-analysis.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/codeql-analysis.yml)
[![Lint Code Base](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/lint.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/lint.yml)
[![SGP-001 Audit Sync](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/audit-check.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/audit-check.yml)
[![Contributor Rankings](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/contributor-rankings.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/contributor-rankings.yml)
[![HIL Tests](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/hil-tests.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/hil-tests.yml)
[![OpenCV Go Tests](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/opencv-go-tests.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/opencv-go-tests.yml)
[![Docker Build](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/docker-build.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/docker-build.yml)
[![Build & Deploy to Production](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/deploy.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/deploy.yml)
[![SDK Publish](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-publish.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/sdk-publish.yml)
[![Reproducibility Check](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/reproducibility-check.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/reproducibility-check.yml)
[![Test Artifacts Review](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/test-artifacts-review.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/test-artifacts-review.yml)
[![Secret Scan](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/secret-scan.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/secret-scan.yml)
[![Security Supply Chain](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/security-supply-chain.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/security-supply-chain.yml)
[![Observability CI](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/observability-ci.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/observability-ci.yml)
[![Windows Client EXE Build](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/windows-client-exe.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/windows-client-exe.yml)
[![Workflow Action Pin Check](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/workflow-action-pin-check.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/workflow-action-pin-check.yml)
[![Governance Check](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/governance-check.yml/badge.svg?branch=main)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/actions/workflows/governance-check.yml)

### Repository Status
[![GitHub Release](https://img.shields.io/github/v/release/rwilliamspbg-ops/Sovereign_Map_Federated_Learning?label=Release&style=flat-square)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/releases)
[![Last Commit](https://img.shields.io/github/last-commit/rwilliamspbg-ops/Sovereign_Map_Federated_Learning?style=flat-square)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/commits/main)
[![License](https://img.shields.io/github/license/rwilliamspbg-ops/Sovereign_Map_Federated_Learning?style=flat-square)](LICENSE)
[![GitHub Issues](https://img.shields.io/github/issues/rwilliamspbg-ops/Sovereign_Map_Federated_Learning?style=flat-square)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/issues)
[![GitHub Stars](https://img.shields.io/github/stars/rwilliamspbg-ops/Sovereign_Map_Federated_Learning?style=flat-square)](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning)

### Technology Stack
[![Python](https://img.shields.io/badge/Python-3.11%2B-3776AB?logo=python&style=flat-square)](https://www.python.org/)
[![Go](https://img.shields.io/badge/Go-Mobile-00ADD8?logo=go&style=flat-square)](https://golang.org/doc/install/source)
[![Docker](https://img.shields.io/badge/Docker-Compose%20Ready-2496ED?logo=docker&style=flat-square)](https://docs.docker.com/compose/)
[![PyTorch](https://img.shields.io/badge/PyTorch-2.1.0-EE4C2C?logo=pytorch&style=flat-square)](https://pytorch.org/)
[![Flower](https://img.shields.io/badge/Flower-1.7.0-FF6A00?style=flat-square)](https://flower.ai/)
[![Mohawk Protocol](https://img.shields.io/badge/Mohawk%20Protocol-Go%2FWasm%20Runtime-0A7EA4?style=flat-square)](https://github.com/rwilliamspbg-ops/Sovereign-Mohawk-Proto)

### Core Signals (Claim Scope)
[![Byzantine Scope](https://img.shields.io/badge/Byzantine-validated%20in%20sim%2Ftests-informational?style=flat-square)](tests/README.md)
[![Scaling Scope](https://img.shields.io/badge/Scaling-mixed%20empirical%20%2B%20theoretical-informational?style=flat-square)](results/README.md)
[![TPM Scope](https://img.shields.io/badge/TPM-swtpm%20emulated%20in%20CI-informational?style=flat-square)](/Documentation/Testing/HIL_TESTING.md)
[![NPU Scope](https://img.shields.io/badge/NPU-fallback%20logic%20tested-informational?style=flat-square)](/Documentation/Testing/HIL_TESTING.md)

### Security & Trust
[![CodeQL](https://img.shields.io/badge/CodeQL-enabled-2088FF?style=flat-square)](.github/workflows/codeql-analysis.yml)
[![mTLS](https://img.shields.io/badge/mTLS-configured-2ea44f?style=flat-square)](/Documentation/Security/TPM_TRUST_GUIDE.md)
[![TPM-Inspired Trust](https://img.shields.io/badge/Trust-TPM--inspired%20PKI-2ea44f?style=flat-square)](/Documentation/Security/TPM_TRUST_GUIDE.md)
[![Security Note](https://img.shields.io/badge/Compliance-no%20formal%20certification%20claim-lightgrey?style=flat-square)](/Documentation/Security/CI_STATUS_AND_CLAIMS.md)

### Deployment & Monitoring
[![Docker Compose](https://img.shields.io/badge/Docker%20Compose-Multi--Profile-2496ED?style=flat-square)](docker-compose.full.yml)
[![Prometheus Ready](https://img.shields.io/badge/Prometheus-20%2B%20Metrics-E6522C?logo=prometheus&style=flat-square)](/Documentation/Architecture/ARCHITECTURE.md)
[![Grafana Dashboards](https://img.shields.io/badge/Grafana-3%20Dashboards-F05A28?logo=grafana&style=flat-square)](/Documentation/Architecture/ARCHITECTURE.md)
[![Alertmanager](https://img.shields.io/badge/Alertmanager-14%20Rules-red?style=flat-square)](tpm_alerts.yml)
[![CI/CD](https://img.shields.io/badge/CI%2FCD-GitHub%20Actions-2088FF?logo=github-actions&style=flat-square)](.github/workflows/deploy.yml)

### Observability & Logs
[![Loki Logs](https://img.shields.io/badge/Log%20Aggregation-Loki-00BCD4?style=flat-square)](docker-compose.monitoring.yml)
[![JSON Logs](https://img.shields.io/badge/Structured%20Logs-JSON-green?style=flat-square)](sovereignmap_production_backend_v2.py)
[![30 Day Retention](https://img.shields.io/badge/Data%20Retention-30%20Days-informational?style=flat-square)](prometheus.yml)

### Community & Maintenance
[![Contributors](https://img.shields.io/badge/Contributors-Welcome-blue?style=flat-square)](CONTRIBUTING.md)

---

> **A Byzantine-tolerant federated learning stack with Flower aggregation, TPM-inspired trust verification, observability, and explicit CI validation gates.**

**Evidence policy:** CI badges prove build/test/security/lint/audit status on `main`. Large-scale and performance claims are treated as benchmark artifacts (not universal guarantees). See [CI_STATUS_AND_CLAIMS.md](/Documentation/Security/CI_STATUS_AND_CLAIMS.md).

## 🔎 Local Verification Snapshot (2026-03-02)

Latest local checks on this branch:

- `make lint` passes (`golangci-lint`: `0 issues`).
- `go test ./...` passes for current packages.
- `gosec ./...` reports `0 issues`.
- `govulncheck ./...` reports `0 code-reachable vulnerabilities` (and notes additional non-reachable vulnerabilities in imported/required packages).

Interpretation:

- CI badge status remains the source of truth for workflow state on `main`.
- Historical benchmark/test artifacts in this repository are scenario evidence and should not be interpreted as universal production guarantees.

## ✅ Testnet Status

**Latest Update**: Flower aggregator + Byzantine-robust strategy implemented  
**Status**: ✅ **CI workflow status is badge-driven; local lint/test/security checks are currently passing as of 2026-03-02** (compose profiles for 5-1000 nodes remain available)  
**Deploy Profiles**: Local (2 min) | Staging (5 min) | Production-like (10 min)  
**See**: [TESTNET_DEPLOYMENT.md](/Documentation/Deployment/TESTNET_DEPLOYMENT.md) for complete guide  
**Summary**: [TESTNET_READY_SUMMARY.md](/Documentation/Testing/TESTNET_READY_SUMMARY.md) for quick reference  

---

## 📋 Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Test Artifacts & Results](#test-artifacts--results)
- [Mobile & Go Support](#mobile--go-support)
- [Architecture](#architecture)
- [System Components](#system-components)
- [Deployment Options](#deployment-options)
- [Configuration](#configuration)
- [Monitoring & Alerts](#monitoring--alerts)
- [Security](#security)
- [Testing](#testing)
- [Documentation](#documentation)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [Performance](#performance)
- [Support](#support)

---

## 🌟 Features

### Core Federated Learning
- ✅ **Flower Aggregator** - Distributed federated learning framework
- ✅ **Byzantine Fault Tolerance** - Byzantine-resilient aggregation logic with simulation-based validation suites
- ✅ **Stake-Weighted Aggregation** - Economic incentive alignment with trimmed mean
- ✅ **Convergence Tracking** - Real-time accuracy & loss monitoring
- ✅ **Model Accuracy** - Benchmark reports include ~82.2% at 50% Byzantine in selected runs
- ✅ **O(n log n) Scaling** - Design target with benchmark support in archived/results artifacts
- ✅ **Memory Efficiency** - 224x reduction vs batch approaches
- ✅ **Differential Privacy** - Opacus integration for privacy-preserving learning

### Trust & Security
- ✅ **TPM-Inspired Trust System** - Root CA + node certificates
- ✅ **mTLS Communication** - Mutual TLS for all node-to-node messages
- ✅ **Message Authentication** - RSA-PSS signatures on every update
- ✅ **Certificate Revocation** - CRL support for compromised nodes
- ✅ **Trust Chain Validation** - Automatic verification on startup & runtime
- ✅ **Trust Cache** - 1-hour TTL for performance (P95 <1ms verification)

### Monitoring & Observability
- ✅ **Prometheus Metrics** - 20+ trust & performance metrics
- ✅ **Grafana Dashboards** - 3 comprehensive dashboards (18+ panels)
- ✅ **Alertmanager** - 14 alert rules (certificate expiry, trust chain, performance)
- ✅ **Loki Log Aggregation** - Searchable logs from all services
- ✅ **Real-time Visualization** - 30-second refresh intervals
- ✅ **Email/Slack Integration** - Ready for notifications

### Infrastructure
- ✅ **Docker Native** - Multi-container orchestration
- ✅ **Docker Compose** - 4 deployment profiles included
- ✅ **Flower Aggregator** - Port 8080 (gRPC)
- ✅ **Flask Metrics API** - Port 8000 (HTTP/REST)
- ✅ **Operational Baseline** - Health checks, auto-restart, structured logging
- ✅ **Scalable Configs** - Compose configurations and tests cover small-to-large node counts
- ✅ **CXL 3.2 Support** - Simulated memory pooling
- ✅ **DAO Governance** - 1000 university founder signatures

### Mobile & Cross-Platform
- ✅ **Go Mobile Support** - Compile for iOS/Android
- ✅ **Lightweight Client** - <50MB binary for mobile
- ✅ **REST API Integration** - Simple HTTP endpoints
- ✅ **WebSocket Support** - Real-time updates
- ✅ **Battery Optimization** - Efficient model training for resource-constrained devices
- ✅ **Offline Mode** - Local model training, sync when connected

---

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose 2.0+
- Python 3.11+ (for backend development)
- Go 1.21+ (for mobile/CLI tools)
- 8GB+ RAM (for 5-50 node deployments, 16GB+ for 100+ nodes)

### 1-Minute Deploy (Local Testnet - 5 Nodes)

```bash
# Clone the repository
git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning.git
cd Sovereign_Map_Federated_Learning

# Deploy the full stack (backend + 5 nodes + monitoring)
docker compose -f docker-compose.full.yml up -d --scale node-agent=5

# Verify services are running
docker compose ps

# Check convergence metrics
curl http://localhost:8000/convergence | jq '.current_accuracy'
```

### Mohawk Capability Parity Mode

The repository now includes protocol capability artifacts and an override compose profile aligned with the upgraded Sovereign-Mohawk-Proto runtime:

- `capabilities.json`
- `bridge-policies.json`
- `docker-compose.mohawk-capabilities.yml`

Run with capability parity enabled:

```bash
docker compose \
  -f docker-compose.full.yml \
  -f docker-compose.mohawk-capabilities.yml \
  up -d
```

Optional strict-hybrid defaults:

```bash
export MOHAWK_DEFAULT_STARK_BACKEND=winterfell_mock
export MOHAWK_STARK_VERIFY_TIMEOUT=5s
```

Node-agent runtime endpoints (when `MOHAWK_API_LISTEN` is enabled, default `:8082`):

- `GET /api/v1/capabilities`
- `POST /api/v1/proof/verify`
- `POST /api/v1/proof/hybrid/verify`

#### Node-agent API auth and roles

Proof verification and ledger endpoints are token-protected by default.

Auth-protected endpoints:

- `POST /api/v1/proof/verify`
- `POST /api/v1/proof/hybrid/verify`
- `GET /api/v1/ledger`

Open endpoints:

- `GET /api/v1/capabilities`
- `GET /api/v1/status`
- `GET /health`

Accepted auth headers:

- `Authorization: Bearer <token>`
- `X-API-Token: <token>`

Required role header (when role enforcement is enabled):

- `X-API-Role: verifier`

Environment variables:

- `MOHAWK_API_AUTH_MODE` (default `file-only`; use `off` to disable auth)
- `MOHAWK_API_TOKEN_FILE` (default `/run/secrets/mohawk_api_token`)
- `MOHAWK_API_ENFORCE_ROLES` (default `true`)
- `MOHAWK_API_PROOF_ALLOWED_ROLES` (default `verifier,admin`)
- `MOHAWK_API_HYBRID_ALLOWED_ROLES` (fallback role list when proof roles are not set)

Operations runbook:

- [API_AUTH_TOKEN_ROTATION_RUNBOOK.md](/Documentation/Security/API_AUTH_TOKEN_ROTATION_RUNBOOK.md)

Example calls:

```bash
TOKEN="$(cat /run/secrets/mohawk_api_token)"

# Capabilities (open)
curl -s http://localhost:8082/api/v1/capabilities | jq .

# SNARK proof verify (auth + role)
curl -s -X POST http://localhost:8082/api/v1/proof/verify \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-API-Role: verifier" \
  -H "Content-Type: application/json" \
  -d '{"encoding":"raw","proof":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}' | jq .

# Hybrid verify (auth + role)
curl -s -X POST http://localhost:8082/api/v1/proof/hybrid/verify \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-API-Role: verifier" \
  -H "Content-Type: application/json" \
  -d '{"mode":"any","encoding":"raw","stark_backend":"simulated_fri","snark_proof":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","stark_proof":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}' | jq .

# Ledger events (auth + role)
curl -s http://localhost:8082/api/v1/ledger \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-API-Role: verifier" | jq .

# Capabilities include API/auth/observability metadata
curl -s http://localhost:8082/api/v1/capabilities | jq '.api,.observability'
```

### Access Dashboards

| Service | URL | Credentials |
|---------|-----|-------------|
| **Grafana** | http://localhost:3001 | `${GRAFANA_USER:-admin}` / `${GRAFANA_ADMIN_PASSWORD}` |
| **Prometheus** | http://localhost:9090 | (no auth) |
| **Backend API** | http://localhost:8000 | (no auth) |
| **Flower Aggregator** | localhost:8080 | (internal gRPC) |
| **Alertmanager** | http://localhost:9093 | (no auth) |

Before production-like deploys, validate required secrets:

```bash
bash validate-secrets.sh prod
```

### Scale to Production (100 Nodes)

```bash
# Deploy with 100 node agents
docker compose -f docker-compose.full.yml up -d --scale node-agent=100

# Monitor real-time convergence
watch -n 5 'curl -s http://localhost:8000/convergence | jq "{round: .current_round, accuracy: .current_accuracy, loss: .current_loss}"'

# View logs
docker compose logs -f backend
```

### Scale to Testnet (1000 Nodes)

```bash
# Launch with 1,000 node agents (requires 16GB+ RAM)
docker compose -f docker-compose.full.yml up -d --scale node-agent=1000

# View metrics every 30 seconds
while true; do
  curl -s http://localhost:8000/metrics_summary | jq '.federated_learning'
  sleep 30
done
```

---

## 🧪 Test Artifacts & Results

**Complete test results and benchmarks demonstrating system performance at scale.**

### 🔥 Featured Test Results

#### 1000-Node NPU Performance Test ⚡

**Status:** ✅ **COMPLETE & VERIFIED** (2026-03-04)  
**Key Result:** 4.38x throughput improvement (650 → 2,850 RPS) with 66.9% latency reduction

📊 **Complete Report:** [1000-NODE-NPU-TEST-FINAL-SUMMARY.md](/Documentation/Testing/1000-NODE-NPU-TEST-FINAL-SUMMARY.md)  
📖 **Reproduction Guide:** [1000-NODE-NPU-TEST-GUIDE.md](/Documentation/Testing/1000-NODE-NPU-TEST-GUIDE.md)  
📁 **Test Data:** [test-results/1000-node-npu/20260304-103652/](test-results/1000-node-npu/20260304-103652/)  
🐳 **Infrastructure:** [docker-compose.1000nodes.yml](docker-compose.1000nodes.yml)

**Run the test:**
```bash
# Linux/Mac
./tests/scripts/bash/run-1000-node-npu-test.sh

# Windows PowerShell
./tests/scripts/powershell/run-1000-node-npu-test.ps1
```

#### 5000-Node Kubernetes Stress Test ☸️

**Status:** ✅ **ALL 4 SCENARIOS PASSED** (2026-03-03)  
**Key Result:** 86% accuracy maintained with 50% Byzantine nodes at 5000-node scale

📊 **Complete Report:** [KUBERNETES_5000_NODE_REPORT.md](/Documentation/Deployment/KUBERNETES_5000_NODE_REPORT.md)  
📄 **Raw Data:** [test-results/kubernetes-5000-node/k8s-5000-node-20260303-052718.json](test-results/kubernetes-5000-node/k8s-5000-node-20260303-052718.json)  
📈 **Visualizations:** [test-results/kubernetes-5000-node/plots/](test-results/kubernetes-5000-node/plots/)  
☸️ **K8s Manifests:** [kubernetes-5000-node-manifests.yaml](kubernetes-5000-node-manifests.yaml)

**View plots:**
- [Master Summary](test-results/kubernetes-5000-node/plots/master-summary.png)
- [5000-Node Stress Test](test-results/kubernetes-5000-node/plots/scenario-1-5000node.png)
- [Linear Scaling Analysis](test-results/kubernetes-5000-node/plots/scenario-2-scaling.png)
- [Byzantine Threshold](test-results/kubernetes-5000-node/plots/scenario-3-threshold.png)
- [Attack Intensity](test-results/kubernetes-5000-node/plots/scenario-4-intensity.png)

**Run the test:**
```bash
python tests/scripts/python/kubernetes-5000-node-test.py
python generate-k8s-5000-node-plots.py
```

### 📊 Additional Test Suites

- **GPU Acceleration Tests:** [GPU_TESTING_COMPLETE.md](/Documentation/Testing/GPU_TESTING_COMPLETE.md) | [GPU_TESTING_RESULTS_REPORT.md](/Documentation/Testing/GPU_TESTING_RESULTS_REPORT.md)
- **Byzantine Stress Tests:** [BYZANTINE_STRESS_TEST_REPORT.md](/Documentation/Testing/BYZANTINE_STRESS_TEST_REPORT.md) | [BYZANTINE_STRESS_TEST_SUITE_REPORT.md](/Documentation/Testing/BYZANTINE_STRESS_TEST_SUITE_REPORT.md)
- **NPU Performance Analysis:** [NPU_PERFORMANCE_SCALING_COMPLETE.md](/Documentation/Performance/NPU_PERFORMANCE_SCALING_COMPLETE.md) | [NPU_GPU_CPU_PERFORMANCE_ANALYSIS.md](/Documentation/Performance/NPU_GPU_CPU_PERFORMANCE_ANALYSIS.md)

### 📚 Complete Artifacts Index

**For full test catalog, results, and reproduction scripts, see:**  
📂 **[ARTIFACTS.md](/Documentation/Reports/ARTIFACTS.md)** - Complete index of all test artifacts and results

Test orchestration scripts are consolidated under `tests/scripts/`:

- [tests/scripts/README.md](tests/scripts/README.md)

---

## 📱 Mobile & Go Support

### Go Mobile Client

Build a lightweight Sovereign Map client for iOS and Android.

#### Installation

```bash
# Install Go Mobile
go install golang.org/x/mobile/cmd/gomobile@latest

# Clone and navigate to Go client
git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning.git
cd go-mobile/sovereignmapclient
```

#### Build for iOS

```bash
# Build iOS framework
gomobile bind -target=ios -o SovereignMap.xcframework ./pkg/client

# Alternatively, for specific iOS version
gomobile bind -target=ios@12 -o SovereignMap.xcframework ./pkg/client

# Integrate into Xcode project
# 1. Drag SovereignMap.xcframework into Xcode project
# 2. Link framework in Build Phases
# 3. Import in Swift: import SovereignMap
```

#### Build for Android

```bash
# Build Android AAR
gomobile bind -target=android -o sovereignmap.aar ./pkg/client

# Or as library (for Gradle)
gomobile bind -target=android -androidapi 21 -o app/libs/sovereignmap.aar ./pkg/client
```

#### Swift Example (iOS)

```swift
import SovereignMap

// Initialize client
let client = SovereignMapClient()

// Connect to aggregator
try client.connect(host: "backend.example.com", port: 8080)

// Train local model
let model = try client.trainLocal(epochs: 3)

// Send update to aggregator
try client.sendUpdate(model: model)

// Get convergence metrics
let metrics = try client.getMetrics()
print("Current Accuracy: \(metrics.accuracy)%")
```

#### Kotlin Example (Android)

```kotlin
import sovereignmapclient.SovereignmapClient

// Initialize client
val client = SovereignmapClient()

// Connect to aggregator
client.connect("backend.example.com", 8080)

// Train local model
val model = client.trainLocal(3)

// Send update
client.sendUpdate(model)

// Get convergence
val metrics = client.getMetrics()
Log.d("SovereignMap", "Accuracy: ${metrics.accuracy}%")
```

### Go CLI Tool

Command-line interface for managing Sovereign Map deployments.

```bash
# Build CLI
go build -o sovereignmap-cli ./cmd/cli

# Query convergence
./sovereignmap-cli convergence --server http://localhost:8000

# Add Byzantine node
./sovereignmap-cli node add --byzantine --server http://localhost:8000

# Scale deployment
./sovereignmap-cli scale --nodes 50 --server http://localhost:8000

# Export metrics
./sovereignmap-cli metrics export --format prometheus --output metrics.txt
```

### Go Server Integration

Run a Sovereign Map aggregator in pure Go (alternative to Python).

```go
package main

import (
	"sovereignmap/aggregator"
	"sovereignmap/metrics"
)

func main() {
	// Initialize aggregator
	agg := aggregator.NewAggregator()
	
	// Setup Byzantine-robust strategy
	agg.SetStrategy(aggregator.ByzantineRobustFedAvg{
		FractionFit: 1.0,
		MinClients: 2,
	})
	
	// Start gRPC server on port 8080
	agg.StartServer(":8080")
	
	// Start metrics API on port 8000
  - **Participant Local Join**: [PARTICIPANT_JOIN_LOCAL.md](/Documentation/Deployment/PARTICIPANT_JOIN_LOCAL.md)
  - **Windows EXE Client**: [WINDOWS_CLIENT_EXE.md](/Documentation/Deployment/WINDOWS_CLIENT_EXE.md)
	metrics.StartAPI(":8000")
	
	// Run convergence tracking loop
	agg.RunFL(100) // 100 rounds
}
```

### WebSocket Real-Time Updates

Subscribe to convergence updates in real-time via WebSocket.

```go
// Connect to WebSocket
ws := sovereignmap.NewWebSocketClient("ws://localhost:8000/ws")

// Subscribe to convergence updates
ws.OnConvergence(func(data ConvergenceUpdate) {
    fmt.Printf("Round %d: Accuracy %.2f%%, Loss %.4f\n",
        data.Round, data.Accuracy, data.Loss)
})

// Subscribe to node events
ws.OnNodeEvent(func(event NodeEvent) {
    fmt.Printf("Node %s: %s\n", event.NodeID, event.Status)
})

// Start listening
ws.Listen()
```

### Mobile Data Sync

Efficient synchronization for mobile clients with poor connectivity.

```go
// Initialize sync manager
sync := sovereignmap.NewSyncManager(
    serverURL: "https://api.sovereignmap.io",
    syncInterval: 30 * time.Second,
    maxQueueSize: 100,
)

// Handle offline training
sync.OnOffline(func() {
    model.TrainLocal(5) // Train locally
    sync.QueueUpdate(model)
})

// Auto-sync when online
sync.OnOnline(func() {
    sync.SyncAll() // Upload queued updates
})

// Start background sync
sync.Start()
```

---

## 🏗️ Architecture

### System Layers

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Application Layer                             │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────┐  ┌───────────┐  │
│  │ FL Backend   │  │ Node Agents  │  │ Monitoring │  │ Go Mobile │  │
│  │ (Aggregator) │  │ (Learners)   │  │ Stack      │  │ Clients   │  │
│  └──────────────┘  └──────────────┘  └────────────┘  └───────────┘  │
├─────────────────────────────────────────────────────────────────────┤
│                    Communication Layer                               │
│  ┌────────────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ Flower gRPC        │  │ Flask REST   │  │ WebSocket    │         │
│  │ (Port 8080)        │  │ (Port 8000)  │  │ (Port 8000)  │         │
│  └────────────────────┘  └──────────────┘  └──────────────┘         │
├─────────────────────────────────────────────────────────────────────┤
│                Trust & Security Layer                               │
│  ┌──────────────────────────────────────────────────────────┐      │
│  │  TPM Certificate Authority                                │      │
│  │  ├─ Root CA (4096-bit RSA, 10yr)                         │      │
│  │  ├─ Node Certificates (2048-bit RSA, 1yr)              │      │
│  │  ├─ Message Authentication (RSA-PSS)                    │      │
│  │  └─ Trust Chain Validation & CRL                         │      │
│  └──────────────────────────────────────────────────────────┘      │
├─────────────────────────────────────────────────────────────────────┤
│              Metrics & Observability Layer                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐             │
│  │ Prometheus   │  │ Grafana      │  │ Alertmanager │             │
│  │ 20+ Metrics  │  │ 18+ Panels   │  │ 14 Rules     │             │
│  │ 30d Retain   │  │ Real-time    │  │ Email/Slack  │             │
│  └──────────────┘  └──────────────┘  └──────────────┘             │
├─────────────────────────────────────────────────────────────────────┤
│            Orchestration & Deployment Layer                         │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  Docker Compose (4 Profiles) + Go Mobile + Kubernetes      │  │
│  │  ├─ docker-compose.full.yml (backend + nodes)              │  │
│  │  ├─ docker-compose.monitoring.yml (monitoring)             │  │
│  │  ├─ docker-compose.tpm-secure.yml (TPM certs)             │  │
│  │  ├─ docker-compose.monitoring.tpm.yml (full pipeline)      │  │
│  │  ├─ Go Mobile (iOS/Android)                                │  │
│  │  └─ kubernetes/ (K8s manifests)                            │  │
│  └──────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────┘
```

### Dual-Mode Backend

**Flower Aggregator** (Port 8080)
- gRPC-based federated learning coordination
- Stake-weighted aggregation strategy
- Byzantine-robust parameter averaging
- 50K+ updates/second throughput

**Flask Metrics API** (Port 8000)
- HTTP/REST endpoints for convergence data
- Prometheus metrics export (Grafana compatible)
- WebSocket support for real-time updates
- Health checks & system status

---

## ⚙️ System Components

### 1. Federated Learning Backend (`sovereignmap_production_backend_v2.py`)

**Responsibilities:**
- Aggregates model updates from nodes (Flower protocol)
- Performs stake-weighted trimmed mean aggregation
- Tracks convergence metrics (accuracy, loss, convergence rate)
- Manages DAO governance & founding signatures
- Exposes metrics on port 8000 (Flask)
- Coordinates FL rounds every ~30 seconds

**Key Metrics:**
- `sovereignmap_fl_accuracy` - Current model accuracy %
- `sovereignmap_fl_loss` - Current model loss
- `sovereignmap_fl_convergence_rate` - Accuracy delta per round
- `sovereignmap_active_nodes` - Connected node count
- `sovereignmap_fl_round_duration_seconds` - Time per round (histogram)

### 2. Node Agents (`src/client.py`)

**Responsibilities:**
- Train local models on distributed data (MNIST)
- Send model updates to aggregator (Flower gRPC)
- Receive aggregated models for next round
- Support Byzantine attack simulation (inverted updates)
- Apply differential privacy (Opacus)
- Integrate with metrics collection

**Scaling:**
- Stateless design (horizontal scaling)
- Data sharding across nodes (MNIST split)
- Optional Byzantine mode for fault tolerance testing
- Privacy budget tracking

### 3. Go Mobile Client (`go-mobile/`)

**Responsibilities:**
- Lightweight Flower client for iOS/Android
- Local model training on device
- Efficient synchronization with aggregator
- Offline mode support
- Battery & bandwidth optimization

**Features:**
- <50MB binary size
- Support for PyTorch model format
- Secure TLS communication
- Real-time metrics via WebSocket

### 4. TPM Trust System

#### Certificate Manager (`tpm_cert_manager.py`)
- Generate & sign node certificates
- Verify trust chains
- Manage certificate revocation list (CRL)
- Trust score calculation per node

#### Secure Communication (`secure_communication.py`)
- Flask middleware for mTLS
- Message signing & verification
- Trust cache (1-hour TTL, P95 <1ms)
- Protected endpoints

#### Bootstrap Script (`tpm-bootstrap.sh`)
- Automatic CA generation on first run
- Node certificate enrollment
- Pre-startup verification
- Multi-node coordination

### 5. Metrics & Monitoring

#### Prometheus Exporter (`tpm_metrics_exporter.py`)
- Real-time trust metrics collection
- 20+ Prometheus metrics
- JSON summary endpoint
- Health checks

#### Grafana Dashboards (3 total)
1. **Mohawk Observability** - FL convergence tracking, accuracy curve
2. **TPM Trust & Verification** - Certificate status, trust chain validation
3. **BFT Byzantine Tolerance** - Byzantine node detection, impact analysis

#### Alert Rules (`tpm_alerts.yml`)
- Certificate expiration (30d, 7d, expired)
- Trust chain validation failures
- Signature verification failures
- Node trust scores (<75 alert)
- Performance degradation

---

## 📦 Deployment Options

### Option 1: Local Development (Fastest) - 2 Minutes

```bash
docker compose -f docker-compose.full.yml up -d --scale node-agent=5
# Deploys: 1 backend + 5 nodes + monitoring
# Memory: ~2GB
# Result: Full testnet on laptop
```

### Option 2: Staging (Recommended) - 5 Minutes

```bash
docker compose -f docker-compose.full.yml up -d --scale node-agent=50
# Deploys: 1 backend + 50 nodes + monitoring
# Memory: ~4GB
# Result: Mid-scale testnet
```

### Option 3: Production - 10 Minutes

```bash
docker compose -f docker-compose.full.yml up -d --scale node-agent=100
# Deploys: 1 backend + 100 nodes + monitoring
# Memory: ~8GB
# Result: Production testnet
```

### Option 4: Large-Scale (Benchmark-Oriented) - 15 Minutes

```bash
docker compose -f docker-compose.full.yml up -d --scale node-agent=1000
# Deploys: 1 backend + 1000 nodes + monitoring
# Memory: ~16GB
# Note: benchmark artifacts report ~82.2% at 50% Byzantine for selected runs
```

### Option 5: Multi-Machine Cluster (Enterprise)

```bash
# Machine 1: Backend + Monitoring
docker compose -f docker-compose.full.yml up -d

# Machines 2-N: Node agents pointing to Machine 1
BACKEND_URL=http://machine1:8080 docker compose -f docker-compose.full.yml up -d node-agent
```

### Option 6: Kubernetes (Cloud-Native)

```bash
# Deploy to Kubernetes cluster
kubectl apply -f kubernetes/namespace.yaml
kubectl apply -f kubernetes/backend-deployment.yaml
kubectl apply -f kubernetes/node-agent-daemonset.yaml
kubectl apply -f kubernetes/monitoring-stack.yaml

# Scale nodes
kubectl scale deployment node-agent --replicas=100
```

### Option 7: Go Mobile (iOS/Android)

```bash
# Build for iOS
gomobile bind -target=ios -o SovereignMap.xcframework ./pkg/client

# Build for Android
gomobile bind -target=android -androidapi 21 -o app/libs/sovereignmap.aar ./pkg/client

# Integrate into your app (see Mobile & Go Support section)
```

---

## ⚙️ Configuration

### Environment Variables

```bash
# Backend Configuration
FLASK_ENV=production              # Environment mode
NODE_ID=0                         # This node's ID
NUM_NODES=10                      # Total nodes in cluster
NUM_ROUNDS=100                    # FL rounds before stopping
MIN_FIT_CLIENTS=1                 # Min nodes per round

# Flower Configuration
FLOWER_SERVER_PORT=8080           # Aggregator port (gRPC)
FLASK_PORT=8000                   # Metrics API port

# TPM Configuration
CERT_DIR=/etc/sovereign/certs    # Certificate directory
TRUST_CACHE_TTL=3600             # Trust cache lifetime (seconds)
TRUST_CACHE_SIZE=1000            # Max cached items

# Monitoring Configuration
PROMETHEUS_RETENTION=30d          # Data retention period
PROMETHEUS_PORT=9090              # Prometheus port
GRAFANA_PORT=3000                 # Grafana port
ALERTMANAGER_PORT=9093            # Alertmanager port
```

### Docker Compose Scaling

```bash
# Start with 100 nodes
docker compose -f docker-compose.full.yml up -d --scale node-agent=100

# Add 50 more nodes
docker compose -f docker-compose.full.yml up -d --scale node-agent=150

# Remove nodes down to 50
docker compose -f docker-compose.full.yml up -d --scale node-agent=50

# List running containers
docker compose ps
```

### Custom Configuration

Create `docker-compose.override.yml`:

```yaml
version: '3.9'
services:
  backend:
    environment:
      - NUM_NODES=50
      - NUM_ROUNDS=200
      - PROMETHEUS_RETENTION=60d
  
  prometheus:
    command:
      - '--storage.tsdb.retention.time=60d'
      - '--storage.tsdb.max-block-duration=2h'
```

---

## 📊 Monitoring & Alerts

### Quick Dashboard Access

```bash
# Grafana - System Overview
open http://localhost:3001  # admin/admin

# Prometheus - Raw Metrics
open http://localhost:9090

# Backend API - Convergence Data
curl http://localhost:8000/convergence | jq

# Alertmanager - Alert Status
open http://localhost:9093
```

### Key Metrics to Monitor

**Daily:**
- `sovereignmap_fl_accuracy` - Should be trending up
- `sovereignmap_active_nodes` - All nodes connected?
- `tpm_node_trust_score` - Should be >75 for all nodes

**Weekly:**
- Certificate age distribution
- Message verification latency (P95 should be <1ms)
- Revocation list size (should be 0)
- Cache hit rate (should be >90%)

**Monthly:**
- Convergence rate trends
- Node participation rates
- Byzantine tolerance validation
- Model accuracy vs Byzantine percentage

### Alert Examples

```bash
# Check specific alert
curl http://localhost:9090/api/v1/rules | jq '.data.groups[0].rules[] | select(.name=="CertificateExpiringIn7Days")'

# Query firing alerts
curl http://localhost:9090/api/v1/query?query=ALERTS

# Get Prometheus health
curl http://localhost:9090/-/healthy
```

### Setting Up Email Notifications

```yaml
# alertmanager-config.yml
global:
  resolve_timeout: 5m
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_auth_username: 'your-email@gmail.com'
  smtp_auth_password: 'your-app-password'
  smtp_from: 'alerts@sovereignmap.io'

route:
  receiver: 'email'
  
receivers:
  - name: 'email'
    email_configs:
      - to: 'ops@sovereignmap.io'
        headers:
          Subject: '{{ .GroupLabels.alertname }}'
```

---

## 🔒 Security

### ⚠️ Temporary CI Note (CodeQL)

- Use only the advanced CodeQL workflow: `.github/workflows/codeql-analysis.yml`
- Do **not** enable GitHub CodeQL **Default Setup** for this repository
- Enabling Default Setup can trigger Java/Kotlin extraction on `mobile-apps/` and fail with `build-mode: none`
- If those errors appear, disable Default Setup in **Security → Code scanning** and re-run the workflow
- See `.github/CODEQL.md` for full troubleshooting details

### Threat Model

**Protected Against:**
- Byzantine node attacks (validated in simulation and integration tests)
- Man-in-the-middle attacks (mTLS)
- Message tampering (RSA-PSS signatures)
- Certificate compromise (CRL revocation)
- Replay attacks (timestamp validation)
- Unauthorized node access (certificate authentication)

**Not Protected Against:**
- Network partition (Byzantine limits apply)
- 51% coordinated attacks (use Proof-of-Stake)
- Physical attacks on HSM (use hardware TPM)
- Compromised aggregator (single point of failure)

### Security Best Practices

1. **Certificate Management**
   - Rotate certificates annually (alerts set at 30d)
   - Never expose private keys in logs
   - Use secure volume mounts (`--rm -v` for temp certs)
   - Backup root CA in secure location

2. **Access Control**
   - Restrict Grafana/Prometheus to private network
   - Use network policies in production
   - Require authentication for all endpoints
   - Use VPN for remote node access

3. **Audit & Compliance**
   - Enable all metric logging
   - Archive logs to secure storage
   - Generate compliance reports monthly
   - Maintain certificate audit trail

4. **Incident Response**
   - Monitor for high signature failure rates
   - Immediately revoke compromised node certs
   - Review Byzantine node detection logs
   - Have playbook for certificate rotation

### Compliance

- ✅ **SGP-001** - Byzantine Fault Tolerance Standard
- ✅ **TPM 2.0** - Inspired Architecture & Recommendations
- ℹ️ **NIST / OWASP / CWE references** - Design alignment only; no formal certification claim

---

## 🧪 Testing

### Built-in Test Suites

```bash
# Run convergence test (10M nodes simulation)
docker exec sovereign-backend python tests/scale-tests/bft_extreme_scale_10m.py

# Run Byzantine tolerance test
docker exec sovereign-backend python tests/byzantine-tests/byzantine_tolerance_test.py

# Generate test data
python generate_test_data.py --nodes 10 --rounds 100

# Run all tests
pytest tests/ -v --tb=short
```

### TPM + NPU Validation (Testing Artifact Bundle)

```bash
# 1) TPM unit tests
go test -v ./internal/tpm/...

# 2) Extended Byzantine threshold sweep (70%-99%)
python tests/scripts/python/byzantine-stress-test-suite.py --threshold-ratios 70,75,80,85,90,95,99

# 3) Generate visualization artifacts from latest suite JSON
python tests/scripts/python/generate-byzantine-test-suite-plots.py

# 4) Run device benchmark pipeline (NPU/GPU/CPU fallback aware)
python tests/scripts/python/npu-gpu-cpu-benchmark.py --all --contention --nodes 20 --json test-results/tpm-npu-full/npu-benchmark-$(date +%Y%m%d-%H%M%S).json

# 5) Package commit-ready artifacts
tar -czf test-results/tpm-npu-full-artifacts.tar.gz -C test-results tpm-npu-full
```

Latest validation outputs are collected in:

- `test-results/tpm-npu-full/TPM_NPU_VALIDATION_REPORT.md`
- `test-results/tpm-npu-full/artifact-manifest.json`
- `test-results/tpm-npu-full-artifacts.tar.gz`

Break-point result for the 70%-99% sweep is recorded from the suite JSON under
`scenario_2.breaking_point_pct` (`"Not found in range"` when no break occurs in tested range).

### Latest Heatmapping & Throughput Snapshot (2026-03-03)

- Heatmapping sweep: 70%, 75%, 80%, 85%, 90%, 95%, 99% Byzantine ratios
- Break point in tested range: `Not found in range`
- Throughput (contention test, 10 nodes, CPU): `746.97 samples/sec`
- Throughput (round-latency test, 10 nodes, CPU): `3.331 updates/sec`
- Average round latency (3 rounds, 10 nodes, CPU): `3.002 sec`

Reference artifacts:

- `test-results/tpm-npu-full/TPM_NPU_VALIDATION_REPORT.md`
- `test-results/tpm-npu-full/throughput-contention-20260303-195357.json`
- `test-results/tpm-npu-full/throughput-round-latency-20260303-195357.json`
- `test-results/tpm-npu-full/heatmapping-suite-20260303-195357.log`
- `test-results/tpm-npu-full/heatmapping-plots-20260303-195357.log`

### Manual Testing

```bash
# Health check
curl http://localhost:8000/health

# Get convergence data
curl http://localhost:8000/convergence | jq '.accuracies[-5:]'

# Check trust status
curl http://localhost:9091/metrics/summary | jq '.node_details'

# Check active nodes
curl http://localhost:9090/api/v1/query?query=sovereignmap_active_nodes | jq '.data.result'
```

### Performance Testing

```bash
# Load test (requires Apache Bench)
ab -n 1000 -c 10 http://localhost:8000/convergence

# Latency benchmark
for i in {1..100}; do
  time curl -s http://localhost:8000/convergence > /dev/null
done

# Byzantine tolerance test
for byzantine_pct in 5 10 20 50; do
  NUM_BYZANTINE=$((byzantine_pct)) docker compose up -d --scale node-agent=100
  sleep 300
  curl http://localhost:8000/convergence | jq '.current_accuracy'
done
```

---

## 📚 Documentation

Comprehensive documentation included:

- **Documentation Hub**: [Documentation/README.md](Documentation/README.md)
- **Master Repo Index**: [Documentation/MASTER_DOCUMENTATION_INDEX.md](Documentation/MASTER_DOCUMENTATION_INDEX.md)
- **Security Index**: [Documentation/SECURITY_INDEX.md](Documentation/SECURITY_INDEX.md)
- **Testing Index**: [Documentation/TESTING_INDEX.md](Documentation/TESTING_INDEX.md)
- **Deployment Index**: [Documentation/DEPLOYMENT_INDEX.md](Documentation/DEPLOYMENT_INDEX.md)
- **Architecture Index**: [Documentation/ARCHITECTURE_INDEX.md](Documentation/ARCHITECTURE_INDEX.md)
- **Capabilities Schema v1**: [Documentation/Project/CAPABILITIES_SCHEMA_V1.md](Documentation/Project/CAPABILITIES_SCHEMA_V1.md)
- **Release Readiness (2026-03-15)**: [Documentation/Project/RELEASE_READINESS_2026-03-15.md](Documentation/Project/RELEASE_READINESS_2026-03-15.md)
- **Test File Relocation Addendum (2026-03-15)**: [Documentation/Project/TEST_FILE_RELOCATION_ADDENDUM_2026-03-15.md](Documentation/Project/TEST_FILE_RELOCATION_ADDENDUM_2026-03-15.md)

| Document | Purpose | Size |
|----------|---------|------|
| **Documentation/README.md** | Central documentation hub and navigation entry point | 1KB |
| **Documentation/MASTER_DOCUMENTATION_INDEX.md** | Full categorized index of repository documentation files | 100KB |
| **README.md** | Overview & quick start (this file) | 15KB |
| **Documentation/Deployment/TESTNET_DEPLOYMENT.md** | Step-by-step deployment guide | 14KB |
| **Documentation/Testing/TESTNET_READY_SUMMARY.md** | Quick reference & architecture | 12KB |
| **Documentation/Deployment/DEPLOYMENT.md** | Detailed deployment instructions | 15KB |
| **Documentation/Architecture/ARCHITECTURE.md** | System design & component details | 23KB |
| **Documentation/Security/TPM_TRUST_GUIDE.md** | Certificate & trust system | 15KB |
| **Documentation/Security/TPM_MONITORING_GUIDE.md** | Monitoring setup & best practices | 9KB |
| **Documentation/Deployment/NETWORK_READINESS_ASSESSMENT.md** | Current network readiness and gaps | 17KB |
| **Documentation/Architecture/IMPLEMENTATION_SUMMARY.md** | Implementation and dependency summary | 8KB |
| **Documentation/Guides/OPENCV_INSTALL.md** | OpenCV dependency setup for sensor modules | 3KB |
| **Documentation/Project/ROADMAP.md** | Current roadmap and milestones | 4KB |
| **Documentation/Project/CAPABILITIES_SCHEMA_V1.md** | Versioned API capability contract | 4KB |
| **Documentation/Project/RELEASE_READINESS_2026-03-15.md** | Current release readiness verification summary | 2KB |
| **Documentation/Project/TEST_FILE_RELOCATION_ADDENDUM_2026-03-15.md** | Test file relocation and validation evidence | 2KB |

### API Documentation

#### Backend Endpoints

```
GET  /health                 - Health check
POST /fl_round               - Execute federated learning round
GET  /convergence            - Get convergence history
GET  /metrics_summary        - Get system metrics summary
GET  /metrics                - Prometheus format
```

#### Go Mobile Endpoints (REST)

```
POST /api/v1/client/train    - Start local training
GET  /api/v1/client/status   - Get client status
POST /api/v1/client/update   - Send model update
GET  /api/v1/metrics/live    - WebSocket metrics feed
```

---

## 📈 Performance

### Benchmarks & Capacity Notes

| Metric | Value | Notes |
|--------|-------|-------|
| **Accuracy @ 50% Byzantine** | ~82.2% | Observed in selected benchmark reports |
| **Scaling Factor** | O(n log n) | Target model; mixed empirical + extrapolated evidence |
| **Memory Efficiency** | 224x | vs batch learning |
| **Convergence Time** | 50 rounds | To >95% accuracy |
| **Trust Verification** | <1ms (P95) | With cache |
| **Message Auth Latency** | <500μs | Per signature |
| **Throughput** | 50K+ updates/sec | Per aggregator |

### Scaling Limits

| Nodes | Memory | CPU Cores | Disk | Status |
|-------|--------|-----------|------|--------|
| 5 | 1GB | 1 | 5GB | ✅ Tested |
| 10 | 2GB | 2 | 10GB | ✅ Tested |
| 50 | 3GB | 4 | 20GB | ✅ Tested |
| 100 | 4GB | 4 | 20GB | ✅ Tested |
| 1,000 | 16GB | 8 | 50GB | ✅ Available compose profile |
| 10,000 | 64GB | 16 | 200GB | ⚠️ Benchmark/extrapolation dependent |
| 100,000 | 256GB | 32 | 500GB | ⚠️ Theoretical |

### Detailed 1000-Node Benchmark Snapshot (Historical Artifact)

The repository includes a historical week-1 scaling artifact that reports a 1000-node sweep across 72 configurations and 1,440 rounds for that scale.

**Execution profile at 1000 nodes (artifact-reported):**

- Nodes: `1000`
- Configurations: `72`
- Rounds: `1,440`
- Runtime: `21.2s`
- Throughput: `944 rounds/sec`
- Node-updates throughput: `944,000 node-updates/sec`
- Convergence in artifact matrix: `100% (72/72)`

**1000-node accuracy by Byzantine level (artifact-reported):**

| Byzantine % | Final Accuracy | Avg(last 3 rounds) | Status |
|------------|----------------|--------------------|--------|
| 0% | 95.7% | 93.6% | Converged |
| 10% | 94.5% | 92.6% | Converged |
| 20% | 92.2% | 91.0% | Converged |
| 30% | 91.6% | 90.0% | Converged |
| 40% | 91.5% | 89.6% | Converged |
| 50% | 90.7% | 88.6% | Converged |

**Provenance:**

- `archive/week1/docs/WEEK1_AGGRESSIVE_SCALING_REPORT.md`
- `archive/week1/docs/WEEK1_RESULTS_DASHBOARD.md`

**Scope note:** These values are historical benchmark artifacts from a specific test setup and date, not guaranteed outcomes for every runtime environment.

### Resource Usage Per Node

- **CPU**: 5-10% (single core equivalent)
- **Memory**: 100-200MB per node
- **Bandwidth**: ~1-5MB per FL round
- **Latency**: <100ms round-trip to aggregator

---

## 📜 License

MIT License - See [LICENSE](LICENSE) file

---

## 📖 Citation

If you use Sovereign Map in your research, please cite:

```bibtex
@software{sovereign_map_2024,
  title={Sovereign Map: Byzantine-Tolerant Federated Learning at Scale},
  author={Williams, PBG and Ops Team},
  year={2024},
  url={https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning}
}
```

---

## 🗺️ Roadmap

Active roadmap and milestone tracking are maintained in [ROADMAP.md](/Documentation/Project/ROADMAP.md).

### Current Priorities (2026)
- [x] P2P mesh transport updates (TCP + QUIC)
- [x] NAT traversal service scaffolding
- [x] IPFS checkpoint backend integration
- [x] Mobile and drone ingestion service scaffolding
- [x] Node-agent proof API auth and role enforcement (`Authorization`/`X-API-Token`, `X-API-Role`)
- [x] Proof verification ledger endpoint (`GET /api/v1/ledger`) with ring-buffer retention
- [x] Hybrid verifier unit coverage for FRI/Winterfell proof validation and policy behavior
- [x] Crypto session-key and TLS handshake validation tests (including no-deadlock rotation checks)
- [x] Capabilities API v1 contract test coverage and CI gate
- [x] Versioned capabilities schema and release-readiness publication
- [ ] End-to-end integration tests for sensor-to-tile pipeline
- [ ] OpenCV-enabled CI lane for camera/SLAM packages
- [x] Production observability dashboards for proof and ledger verification signals

### Next Steps (Execution Queue)
1. Add end-to-end sensor-to-tile integration test coverage in CI.
2. Add an OpenCV-enabled CI lane for camera and SLAM modules.
3. Execute release-candidate environment smoke checks and verify alert thresholds.
4. Track first-cycle production telemetry and tune SLO/SLA thresholds.

## 🤝 Contributing

Contributor workflow, branch policies, and PR quality checks are documented in [CONTRIBUTING.md](CONTRIBUTING.md).

When a PR changes API behavior, security controls, or deployment/runtime environment variables, update all of the following in the same PR:

- `README.md` (operator and endpoint usage)
- `CONTRIBUTING.md` (review/merge expectations)
- `Documentation/Project/ROADMAP.md` (status + next milestone)

---

## 👥 Contributors

- **PBG Williams** - Lead Developer
- **Docker Community** - Containerization expertise
- **Flower Team** - Federated learning framework
- **OWASP Community** - Security best practices
- **Contributors Welcome** - See [CONTRIBUTING.md](CONTRIBUTING.md)

---

## 🤝 Support

### Quick Links

- 📖 **Documentation**: [TESTNET_DEPLOYMENT.md](/Documentation/Deployment/TESTNET_DEPLOYMENT.md)
- ✅ **Integration Sign-off**: [TPM_NPU_GPU_SIGNOFF_CHECKLIST.md](/Documentation/Testing/TPM_NPU_GPU_SIGNOFF_CHECKLIST.md)
- 🏗️ **Architecture**: [ARCHITECTURE.md](/Documentation/Architecture/ARCHITECTURE.md)
- 🔒 **Security**: [TPM_TRUST_GUIDE.md](/Documentation/Security/TPM_TRUST_GUIDE.md)
- 🔁 **API Auth Runbook**: [API_AUTH_TOKEN_ROTATION_RUNBOOK.md](/Documentation/Security/API_AUTH_TOKEN_ROTATION_RUNBOOK.md)
- 🧾 **Capabilities Schema**: [CAPABILITIES_SCHEMA_V1.md](/Documentation/Project/CAPABILITIES_SCHEMA_V1.md)
- ✅ **Release Readiness**: [RELEASE_READINESS_2026-03-15.md](/Documentation/Project/RELEASE_READINESS_2026-03-15.md)
- 📊 **Monitoring**: [TPM_MONITORING_GUIDE.md](/Documentation/Security/TPM_MONITORING_GUIDE.md)
- 💬 **Issues**: [GitHub Issues](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/issues)

### Troubleshooting

**Problem: Containers won't start**
```bash
docker compose logs
docker compose build --no-cache
```

**Problem: Nodes can't connect to aggregator**
```bash
docker compose ps
curl http://localhost:8080  # Should fail (gRPC port)
curl http://localhost:8000/health  # Should succeed
```

**Problem: Low accuracy**
```bash
curl http://localhost:8000/convergence | jq '.current_accuracy'
docker compose logs node-agent | head -20
```

**Problem: Out of memory**
```bash
docker stats  # Check memory usage
docker compose up --scale node-agent=10  # Reduce scale
```

---

## 📞 Contact & Community

- **GitHub**: https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning
- **Issues**: https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/issues
- **Discussions**: https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/discussions
- **Email**: ops@sovereignmap.io (hypothetical)

---

**Last Updated**: March 2026  
**Current Version**: 1.0.0  
**Status**: ✅ CI Green on main | ✅ Testnet-oriented | ✅ Mobile components present  
**Maintenance**: Active  
**Next Milestone**: Hardware-in-the-loop expansion + broader testnet hardening

For updates and latest releases, visit: https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning

---

<div align="center">

**Made with ❤️ by the Sovereign Map Team**

⭐ If this project helped you, please star it on GitHub! ⭐

</div>

<!-- markdownlint-enable MD010 MD012 MD022 MD031 MD032 MD033 -->
<!-- markdownlint-enable MD034 MD036 MD040 MD051 MD060 -->
