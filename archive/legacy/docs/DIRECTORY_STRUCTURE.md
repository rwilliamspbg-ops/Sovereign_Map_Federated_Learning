# 📁 Directory Structure Reference

Complete organization and purpose of all directories and key files in the Sovereign Map project.

---

## Root Structure

```
Sovereign_Map_Federated_Learning/
├── README.md                    # Main landing page (dashboard)
├── DIRECTORY_STRUCTURE.md       # This file
├── LICENSE                      # MIT License
├── .gitignore                   # Git ignore patterns
├── .gitattributes               # Git attributes
├── requirements.txt             # Python dependencies
├── Makefile                     # Common tasks
├── VERSION.md                   # Version tracking
│
├── docker/                      # Docker configuration
├── src/                         # Source code
├── tests/                       # Test suite
├── documentation/               # Complete documentation
├── config/                      # Configuration files
├── monitoring/                  # Observability setup
├── scripts/                     # Utility scripts
├── terraform/                   # Infrastructure as Code
│
└── [LEGACY DIRECTORIES]         # To be consolidated
    ├── apps/
    ├── cmd/
    ├── internal/
    ├── frontend/
    ├── packages/
    ├── pkg/
    └── test-data/
```

---

## 📦 Core Directories

### `docker/` - Container Configuration

Contains all Docker-related files for containerization and deployment.

```
docker/
├── Dockerfile.backend           # FL Backend API (Python 3.11)
├── Dockerfile.frontend          # React dashboard (Node 18)
├── Dockerfile.monitoring        # Prometheus/Grafana (Alpine)
├── docker-compose.full.yml      # Production stack (8 services)
├── docker-compose.dev.yml       # Development environment
├── .dockerignore                # Docker ignore patterns
│
├── entrypoints/
│   ├── backend.sh              # Backend startup script
│   ├── frontend.sh             # Frontend startup script
│   └── monitoring.sh           # Monitoring startup script
│
└── [README.md]                  # Docker documentation
```

**Purpose:** Centralized container definitions and orchestration configurations.

**Key Files:**
- `docker-compose.full.yml` - Production 8-service stack
- `Dockerfile.backend` - FL Backend (Python 3.11 optimized)
- `Dockerfile.frontend` - React dashboard
- `Dockerfile.monitoring` - Monitoring/Prometheus stack

---

### `src/` - Source Code

Primary application source code.

```
src/
├── sovereign_federation_backend.py      # Main Flask API + WebSocket
├── fl_metrics_translator.py             # Hilbert curve 3D mapping
├── spatial_threat_analyzer.py           # Gemini AI threat analysis
│
├── node/
│   ├── genesis_node.py                  # Core node implementation
│   ├── island_mode.py                   # Offline-first operations
│   └── tpm_attestation.py               # TPM integration
│
├── consensus/
│   ├── bft_engine.py                    # Byzantine Fault Tolerance
│   ├── aggregator.py                    # Model aggregation
│   └── validator.py                     # Peer verification
│
├── privacy/
│   ├── dp_engine.py                     # Differential privacy
│   ├── sgp001_compliance.py             # SGP-001 standard
│   └── encryption.py                    # Cryptographic ops
│
├── utils/
│   ├── metrics.py                       # Metrics utilities
│   ├── logger.py                        # Logging configuration
│   └── config.py                        # Config management
│
└── __init__.py
```

**Purpose:** All application source code organized by logical domain.

**Key Files:**
- `sovereign_federation_backend.py` - Main API entry point (510 lines)
- `fl_metrics_translator.py` - Metric mapping engine (404 lines)
- `spatial_threat_analyzer.py` - AI threat analysis (481 lines)

---

### `tests/` - Test Suite

Comprehensive test collection for validation and benchmarking.

```
tests/
├── bft_week2_100k_nodes.py                    # 100K node scaling
├── bft_week2_100k_byzantine_boundary.py       # Byzantine tolerance (50%)
├── bft_week2_5000_node_scaling.py             # 5K node validation
├── bft_week2_cascading_failures.py            # Failure propagation
├── bft_week2_failure_modes.py                 # 5 failure mode tests
├── bft_week2_network_partitions.py            # Network split scenarios
├── bft_week2_gpu_profiling.py                 # GPU acceleration
├── bft_week2_mnist_validation.py              # Real MNIST data
├── bft_week2_production_readiness.py          # Production checklist
│
├── unit/
│   ├── test_metrics.py
│   ├── test_consensus.py
│   └── test_privacy.py
│
├── integration/
│   ├── test_end_to_end.py
│   └── test_api.py
│
├── fixtures/
│   └── test_data/                             # Test datasets
│
└── conftest.py                                # Pytest configuration
```

**Purpose:** Validation, benchmarking, and scenario testing.

**Test Results:**
- ✅ 8 comprehensive scenarios (Week 2)
- ✅ 100K node scale validation
- ✅ 50% Byzantine tolerance proven
- ✅ Recovery time metrics logged
- ✅ GPU profiling completed

**Running Tests:**
```bash
# All tests
pytest tests/

# Specific test
python tests/bft_week2_100k_nodes.py

# With coverage
pytest tests/ --cov=src
```

---

### `documentation/` - Complete Documentation

Organized reference documentation for all aspects.

```
documentation/
├── README.md                            # Docs index & overview
├── QUICKSTART.md                        # 5-minute setup guide
├── INSTALLATION.md                      # Detailed installation
├── ARCHITECTURE.md                      # System design & components
├── RESEARCH_FINDINGS.md                 # Byzantine analysis results
├── API_REFERENCE.md                     # 8 backend endpoints
├── DEPLOYMENT.md                        # Production deployment guide
├── MONITORING.md                        # Observability setup
├── TESTING.md                           # Test framework documentation
├── TROUBLESHOOTING.md                   # Common issues & solutions
├── CONTRIBUTING.md                      # Developer guidelines
├── DEVELOPMENT.md                       # Local development setup
│
├── guides/
│   ├── DOCKER_GUIDE.md                  # Docker deep dive
│   ├── KUBERNETES_GUIDE.md              # K8s deployment
│   ├── AWS_DEPLOYMENT.md                # AWS-specific setup
│   ├── SECURITY.md                      # Security best practices
│   └── SCALING.md                       # Horizontal scaling
│
├── research/
│   ├── BYZANTINE_ANALYSIS.md            # Detailed BFT findings
│   ├── SCALING_ANALYSIS.md              # Performance analysis
│   ├── PRIVACY_AUDIT.md                 # SGP-001 compliance
│   └── BENCHMARK_RESULTS.md             # Performance benchmarks
│
└── api/
    ├── ENDPOINTS.md                     # Endpoint documentation
    ├── WEBSOCKET.md                     # WebSocket protocol
    ├── METRICS.md                       # Metrics specification
    └── ERRORS.md                        # Error codes & handling
```

**Purpose:** Single source of truth for all documentation.

**Key Documents:**
- `QUICKSTART.md` - Start here for new users
- `RESEARCH_FINDINGS.md` - Byzantine tolerance results
- `DEPLOYMENT.md` - Production deployment
- `ARCHITECTURE.md` - System design overview

---

### `config/` - Configuration Files

All configuration and policy definitions.

```
config/
├── prometheus.yml                       # Prometheus scrape config
├── alertmanager.yml                     # Alert routing rules
├── grafana-dashboard.json               # Grafana dashboard definition
├── bft_rules.yml                        # BFT consensus rules
│
├── docker/
│   ├── .env.production                  # Production environment
│   ├── .env.staging                     # Staging environment
│   └── .env.development                 # Development environment
│
├── kubernetes/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   └── ingress.yaml
│
├── nginx/
│   ├── nginx.conf                       # Main nginx config
│   ├── ssl-config.conf                  # SSL/TLS settings
│   └── upstream.conf                    # Upstream definitions
│
└── monitoring/
    ├── recording-rules.yml              # Prometheus recording rules
    └── alert-rules.yml                  # Alert rule definitions
```

**Purpose:** Centralized configuration management.

**Key Files:**
- `prometheus.yml` - Metrics collection config
- `alertmanager.yml` - Alert notification rules
- `bft_rules.yml` - Byzantine tolerance parameters
- `grafana-dashboard.json` - 11-panel monitoring dashboard

---

### `monitoring/` - Observability Stack

Monitoring and alerting infrastructure setup.

```
monitoring/
├── prometheus/
│   ├── prometheus.yml
│   ├── recording-rules.yml
│   └── alert-rules.yml
│
├── grafana/
│   ├── provisioning/
│   │   ├── datasources/
│   │   │   └── prometheus.yml
│   │   └── dashboards/
│   │       └── bft-monitoring.json
│   └── dashboards/
│       ├── byzantine-metrics.json       # Byzantine threat dashboard
│       ├── consensus-metrics.json       # Consensus performance
│       ├── node-metrics.json            # Per-node metrics
│       ├── privacy-audit.json           # Privacy budget tracking
│       ├── network-topology.json        # Network visualization
│       ├── alert-status.json            # Alert monitoring
│       ├── performance-trends.json      # Historical trends
│       ├── scaling-analysis.json        # Scaling metrics
│       ├── recovery-analysis.json       # Recovery time tracking
│       ├── threat-heatmap.json          # Byzantine threat map
│       └── custom-queries.json          # Custom metric queries
│
├── alertmanager/
│   ├── alertmanager.yml
│   └── notification-templates/
│       ├── slack.tmpl
│       ├── email.tmpl
│       └── pagerduty.tmpl
│
└── loki/                                # Log aggregation
    └── loki-config.yml
```

**Purpose:** Complete monitoring, alerting, and visualization stack.

**Dashboards (11 Total):**
1. Byzantine Metrics
2. Consensus Performance
3. Node Status
4. Privacy Audit
5. Network Topology
6. Alert Status
7. Performance Trends
8. Scaling Analysis
9. Recovery Analysis
10. Threat Heatmap
11. Custom Queries

---

### `scripts/` - Utility Scripts

Automation and operational scripts.

```
scripts/
├── deploy.sh                            # Production deployment
├── health-check.sh                      # System health verification
├── generate-metrics.sh                  # Metric generation
├── setup-monitoring.sh                  # Monitoring stack setup
├── backup-configs.sh                    # Configuration backup
├── cleanup.sh                           # Cleanup old files
│
├── docker/
│   ├── build-all.sh                     # Build all images
│   ├── push-images.sh                   # Push to registry
│   └── pull-latest.sh                   # Pull latest images
│
├── testing/
│   ├── run-all-tests.sh                 # Run complete test suite
│   ├── run-quick-tests.sh               # Quick validation
│   └── generate-test-report.sh          # Test report generation
│
└── dev/
    ├── setup-dev-env.sh                 # Dev environment setup
    ├── lint.sh                          # Code linting
    └── format.sh                        # Code formatting
```

**Purpose:** Automation of common operational tasks.

**Common Commands:**
```bash
./scripts/deploy.sh production              # Deploy to production
./scripts/health-check.sh                   # Verify system health
./scripts/docker/build-all.sh               # Build all Docker images
./scripts/testing/run-all-tests.sh          # Run complete test suite
```

---

### `terraform/` - Infrastructure as Code

Cloud infrastructure definitions.

```
terraform/
├── main.tf                              # Main configuration
├── variables.tf                         # Variable definitions
├── outputs.tf                           # Output values
├── terraform.tfvars                     # Variable values
├── .terraform/                          # Terraform cache (gitignored)
│
├── aws/
│   ├── main.tf
│   ├── ec2.tf                           # EC2 instance config
│   ├── networking.tf                    # VPC & networking
│   ├── iam.tf                           # IAM roles & policies
│   ├── rds.tf                           # Database setup (if used)
│   └── autoscaling.tf                   # Auto-scaling groups
│
├── kubernetes/
│   ├── cluster.tf                       # EKS cluster
│   ├── nodegroup.tf                     # Node groups
│   └── addons.tf                        # Cluster addons
│
└── modules/
    ├── networking/
    ├── compute/
    └── monitoring/
```

**Purpose:** Infrastructure automation and reproducible deployments.

**Deploy to AWS:**
```bash
cd terraform/aws
terraform init
terraform plan
terraform apply
```

---

## 📚 Legacy Directories (To Be Consolidated)

These directories contain code that should be reorganized into `src/` or `tests/`:

```
apps/                   → Merge into src/
cmd/                    → Merge into scripts/
internal/               → Merge into src/
frontend/               → Merge into docker/
packages/               → Merge into src/
pkg/                    → Merge into src/
test-data/              → Merge into tests/fixtures/
audit_results/          → Archive to docs/results/
test-results/           → Archive to docs/results/
```

---

## 🗂️ File Organization Best Practices

### Naming Conventions

**Python Files:**
```
src/                    # Source code
tests/bft_*.py          # BFT tests (week2 format)
scripts/*.sh            # Bash scripts
config/*.yml            # YAML configs
```

**Docker Files:**
```
docker/Dockerfile.*     # Containerfiles (with service suffix)
docker-compose.*.yml    # Compose files (with environment suffix)
```

**Documentation:**
```
documentation/          # All markdown docs
*.md                    # File naming in UPPERCASE_WITH_UNDERSCORES
guides/                 # How-to guides
research/               # Research findings
```

**Configuration:**
```
config/                 # All config files
.env.*                  # Environment files (gitignored)
*.yml                   # YAML configs
*.yaml                  # YAML configs
```

---

## 🚀 Getting Started with Repository

### 1. Understand the Layout
- Read this file first
- Then explore `src/` and `tests/`
- Review `documentation/QUICKSTART.md`

### 2. Run Tests
```bash
# See what's available
ls tests/bft_week2_*.py

# Run specific test
python tests/bft_week2_100k_nodes.py

# Run all tests
pytest tests/
```

### 3. Deploy Locally
```bash
# Build Docker images
docker compose -f docker/docker-compose.dev.yml build

# Start services
docker compose -f docker/docker-compose.dev.yml up

# View dashboard
open http://localhost:3000  # Grafana
open http://localhost:9090  # Prometheus
```

### 4. Review Documentation
```
documentation/
├── QUICKSTART.md        ← Start here
├── ARCHITECTURE.md      ← Understand design
├── RESEARCH_FINDINGS.md ← See results
└── DEPLOYMENT.md        ← Prepare production
```

---

## 📊 File Statistics

| Category | Count | Location |
|----------|-------|----------|
| Python Files | 8+ | `src/`, `tests/` |
| Docker Files | 6+ | `docker/` |
| Documentation | 15+ | `documentation/` |
| Config Files | 10+ | `config/` |
| Test Files | 8+ | `tests/` |
| Scripts | 10+ | `scripts/` |
| Terraform Files | 15+ | `terraform/` |
| **Total** | **100+** | (Organized) |

---

## 🔄 Quick Navigation

### For New Users
1. Start: `README.md`
2. Setup: `documentation/QUICKSTART.md`
3. Learn: `documentation/ARCHITECTURE.md`

### For Developers
1. Code: `src/`
2. Tests: `tests/`
3. Guide: `documentation/DEVELOPMENT.md`

### For DevOps/Operations
1. Docker: `docker/`
2. Config: `config/`
3. Monitoring: `monitoring/`
4. Guide: `documentation/DEPLOYMENT.md`

### For Researchers
1. Results: `documentation/RESEARCH_FINDINGS.md`
2. Tests: `tests/bft_*.py`
3. Analysis: `documentation/research/`

---

## ✅ Checklist for Clean Repository

- [x] README dashboard created
- [x] Directory structure documented
- [x] Docker files consolidated
- [x] Documentation organized
- [x] Configuration centralized
- [x] Scripts organized
- [ ] Legacy directories consolidated
- [ ] Old markdown files archived
- [ ] .gitignore updated
- [ ] CI/CD workflows configured

---

**Last Updated:** February 2026  
**Version:** 1.0  
**Maintained By:** Sovereign Map Team
