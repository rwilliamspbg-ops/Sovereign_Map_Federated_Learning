# 🚀 Sovereign Map SDK - Build With Me

<p align="center">
  <strong>Join the revolution in decentralized, privacy-preserving mapping</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/SDK-v0.2.0--alpha-blue" alt="SDK Version">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go" alt="Go">
  <img src="https://img.shields.io/badge/Python-3.9+-3776AB?logo=python" alt="Python">
  <img src="https://img.shields.io/badge/TypeScript-5.0+-3178C6?logo=typescript" alt="TypeScript">
  <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License">
</p>

---

## 👋 Welcome, Developer!

I'm building **Sovereign Map** - a coordinator-less, privacy-preserving decentralized mapping network powered by federated learning. This SDK is your gateway to integrating with the protocol, running nodes, and contributing to a truly sovereign spatial operating system.

**Why build with Sovereign Map?**

- 🔒 **Privacy-First**: SGP-001 differential privacy (ε = 1.0, δ = 1e-5) enforced at the hardware level
- 🏝️ **Edge Resilient**: Independent Island Mode for autonomous operation without network connectivity
- ⚡ **No Coordinators**: Byzantine fault-tolerant consensus without centralized control
- 💰 **Earn While You Build**: Node operators earn rewards with up to 27x multipliers
- 🌍 **Real Impact**: Help build the alternative to centralized mapping monopolies

---

## 📚 Table of Contents

- [Quick Start](#-quick-start)
- [SDK Architecture](#-sdk-architecture)
- [Installation](#-installation)
- [Core Packages](#-core-packages)
  - [Core Package](#core-package)
  - [Privacy Package](#privacy-package)
  - [Consensus Package](#consensus-package)
  - [Island Package](#island-package)
- [Usage Examples](#-usage-examples)
- [API Reference](#-api-reference)
- [Running a Node](#-running-a-node)
- [Testing](#-testing)
- [Contributing](#-contributing)
- [Community](#-community)
- [Roadmap](#-roadmap)

---

## ⚡ Quick Start

Get up and running in 5 minutes:

### Prerequisites

- **Go 1.22+** or **Python 3.9+** or **Node.js 18+**
- **Docker** and **Docker Compose** (for full stack deployment)
- **16 GB RAM** minimum (32 GB recommended for production nodes)
- **Linux/macOS/Windows** with WSL2

### 1. Clone the Repository

```bash
git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning
cd Sovereign_Map_Federated_Learning
```

### 2. Install Dependencies

**Go:**
```bash
make tidy
```

**Python:**
```bash
pip install -r requirements.txt
```

**TypeScript (Frontend):**
```bash
cd frontend
npm install
```

### 3. Run Your First Node

```bash
# Quick local simulation
make test

# Full deployment with Docker
make deploy

# View logs
make logs
```

### 4. Access the Dashboard

Open your browser to:
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **Grafana Monitoring**: http://localhost:3001 (admin/admin)

---

## 🏗️ SDK Architecture

The Sovereign Map SDK is organized into modular packages that you can integrate independently or as a complete stack:

```
Sovereign_Map_Federated_Learning/
├── packages/
│   ├── core/           # Node management, protocol messages
│   ├── privacy/        # SGP-001 differential privacy
│   ├── consensus/      # Byzantine fault-tolerant aggregation
│   └── island/         # Independent Island Mode
├── internal/
│   ├── batch/          # Model aggregation
│   ├── wasmhost/       # zk-SNARK verification
│   ├── tpm/            # Hardware attestation
│   ├── p2p/            # Peer verification
│   └── config/         # Configuration management
├── pkg/
│   └── protocol/       # Federated learning messages
├── cmd/
│   └── node-agent/     # Main node agent binary
└── frontend/           # React/Next.js dashboard
```

### Key Design Principles

1. **Modularity**: Each package can be used independently
2. **Privacy-by-Default**: All operations respect SGP-001 boundaries
3. **Hardware Acceleration**: NPU/TPM integration where available
4. **Zero-Trust**: Cryptographic verification at every layer
5. **Developer-Friendly**: Clear APIs, comprehensive docs, and examples

---

## 📦 Installation

### Option 1: Go Module (Recommended for Go Projects)

```bash
# Add to your go.mod
go get github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning
```

### Option 2: Python Package (Coming Soon)

```bash
pip install sovereign-map-sdk
```

### Option 3: NPM Package (Coming Soon)

```bash
npm install @sovereign-map/sdk
```

### Option 4: Docker Image

```bash
docker pull ghcr.io/rwilliamspbg-ops/sovereign-map-node:latest
```

---

## 🧩 Core Packages

### Core Package

**Path**: `packages/core/`

The core package provides node lifecycle management, protocol message handling, and network communication.

#### Key Features

- Node registration and discovery
- Protocol message serialization/deserialization
- gRPC and HTTP client/server implementations
- Health checks and status reporting

#### Example Usage

```go
package main

import (
    "github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/pkg/protocol"
    "github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/config"
)

func main() {
    // Load configuration
    cfg := config.LoadConfig()
    
    // Create a model update
    update := &protocol.ModelUpdate{
        NodeID:      cfg.NodeID,
        ModelWeights: []float32{0.1, 0.2, 0.3},
        Timestamp:   time.Now().Unix(),
        Proof:       generateProof(), // zk-SNARK proof
    }
    
    // Send to aggregator
    client := NewAggregatorClient(cfg.AggregatorURL)
    response, err := client.SubmitUpdate(update)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Update accepted: %v", response)
}
```

#### API Reference

**Node Registration**
```go
func RegisterNode(req *protocol.RegistrationRequest) (*protocol.RegistrationResponse, error)
```

**Model Update Submission**
```go
func SubmitModelUpdate(update *protocol.ModelUpdate) error
```

**Fetch Training Task**
```go
func GetTrainingTask() (*protocol.TrainingTask, error)
```

---

### Privacy Package

**Path**: `packages/privacy/`

The privacy package implements the SGP-001 differential privacy standard with hardware-accelerated noise injection.

#### Key Features

- Gaussian differential privacy mechanism
- Privacy budget tracking (ε = 1.0, δ = 1e-5)
- NPU-accelerated noise generation
- Cryptographic audit trails

#### Example Usage

```go
package main

import (
    "github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/packages/privacy"
)

func main() {
    // Initialize privacy engine
    engine := privacy.NewSGP001Engine(privacy.Config{
        Epsilon: 1.0,
        Delta:   1e-5,
        UseNPU:  true, // Hardware acceleration
    })
    
    // Sensitive data (model gradients)
    gradients := []float32{0.5, 0.3, 0.8, 0.2}
    
    // Apply differential privacy
    privatized, err := engine.Privatize(gradients)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Privatized gradients: %v", privatized)
    
    // Check remaining privacy budget
    budget := engine.GetRemainingBudget()
    log.Printf("Remaining budget: ε=%.2f, δ=%.2e", budget.Epsilon, budget.Delta)
}
```

#### Privacy Budget Management

```go
// Start a new privacy session
session := engine.NewSession("training-round-42")

// Perform multiple operations
for i := 0; i < 10; i++ {
    data := generateTrainingData()
    privatized, _ := session.Privatize(data)
    submitUpdate(privatized)
}

// Verify budget not exceeded
if session.BudgetExceeded() {
    log.Fatal("Privacy budget exceeded!")
}
```

---

### Consensus Package

**Path**: `packages/consensus/`

The consensus package provides Byzantine fault-tolerant model aggregation without coordinators.

#### Key Features

- BFT consensus with 55.5% malicious node resilience
- Quorum-based voting: quorum = ⌈(2n/3)⌉ + 1
- Weighted model aggregation
- Cryptographic proof generation
- Performance metrics tracking

#### Example Usage

```go
package main

import (
    "github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/consensus"
)

func main() {
    // Initialize consensus coordinator
    coord := consensus.NewCoordinator(consensus.CoordinatorConfig{
        NodeID:        "node-001",
        QuorumSize:    7,  // Minimum nodes for consensus
        MaxRounds:     100,
        RoundTimeout:  30 * time.Second,
    })
    
    // Start consensus round
    round, err := coord.StartRound([]byte("training-task-data"))
    if err != nil {
        log.Fatal(err)
    }
    
    // Submit votes from nodes
    for _, nodeUpdate := range collectNodeUpdates() {
        vote := &consensus.Vote{
            NodeID:    nodeUpdate.NodeID,
            RoundID:   round.ID,
            DataHash:  nodeUpdate.Hash,
            Signature: nodeUpdate.Signature,
        }
        coord.SubmitVote(vote)
    }
    
    // Wait for consensus
    result := coord.WaitForConsensus(round.ID)
    if result.Status == consensus.StatusCommitted {
        log.Printf("Consensus reached! Aggregate model: %v", result.AggregateModel)
    }
}
```

#### Distributed Aggregation

```go
// Create distributed aggregator
aggregator := consensus.NewDistributedAggregator(consensus.AggregatorConfig{
    WeightingScheme: consensus.WeightByQuality,
    MinParticipants: 10,
})

// Add model updates from nodes
for _, update := range modelUpdates {
    aggregator.AddUpdate(update.NodeID, update.Weights, update.Quality)
}

// Compute aggregate
aggregate, proof := aggregator.Aggregate()
log.Printf("Aggregate model computed with %d participants", len(modelUpdates))
log.Printf("Cryptographic proof: %x", proof)
```

---

### Island Package

**Path**: `packages/island/`

The island package enables autonomous operation when disconnected from the network.

#### Key Features

- Automatic mode switching (online ↔ island)
- Tamper-evident state snapshots
- Cached update storage
- Automatic synchronization on reconnection
- State recovery after restarts

#### Example Usage

```go
package main

import (
    "github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/island"
)

func main() {
    // Initialize island mode manager
    manager := island.NewManager(island.Config{
        NodeID:              "node-001",
        MaxCachedUpdates:    1000,
        ConnectivityCheck:   5 * time.Second,
        StatePersistence:    "/var/lib/sovereign-map/state",
    })
    
    // Start monitoring
    manager.Start()
    
    // Submit updates (works online or offline)
    update := createModelUpdate()
    err := manager.SubmitUpdate(update)
    if err != nil {
        log.Fatal(err)
    }
    
    // Check current mode
    if manager.IsIslandMode() {
        log.Println("Running in Independent Island Mode")
        log.Printf("Cached updates: %d", manager.GetCachedUpdateCount())
    }
    
    // Wait for reconnection
    <-manager.OnlineNotification()
    log.Println("Back online! Synchronizing...")
}
```

#### State Recovery

```go
// Load previous state after restart
state, err := island.LoadState("/var/lib/sovereign-map/state")
if err != nil {
    log.Fatal(err)
}

// Verify integrity
if !state.VerifyIntegrity() {
    log.Fatal("State tampering detected!")
}

// Resume operations
manager.RestoreState(state)
log.Printf("Restored %d cached updates", len(state.CachedUpdates))
```

---

## 💡 Usage Examples

### Example 1: Simple Node Setup

Run a lightweight node for testing:

```bash
# Set environment variables
export NODE_ID="my-node-001"
export AGGREGATOR_URL="http://localhost:8080"
export ENABLE_TPM="false"  # Disable TPM for testing

# Run node agent
go run cmd/node-agent/main.go
```

### Example 2: Production Node with Monitoring

Deploy a full production node with observability:

```bash
# Deploy node + monitoring stack
docker-compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d

# Access dashboards
# Grafana: http://localhost:3001
# Prometheus: http://localhost:9090
# Node API: http://localhost:8080/health
```

### Example 3: Custom Training Integration

Integrate Sovereign Map into your ML training pipeline:

```python
import sovereign_map_sdk as sm

# Initialize client
client = sm.Client(
    node_id="my-custom-node",
    aggregator_url="http://aggregator:8080",
    privacy_config=sm.PrivacyConfig(epsilon=1.0, delta=1e-5)
)

# Connect to network
client.register()

# Training loop
for epoch in range(100):
    # Your training logic
    model = train_model(data)
    
    # Extract gradients
    gradients = model.get_gradients()
    
    # Submit privatized update
    client.submit_update(
        gradients=gradients,
        metadata={"epoch": epoch, "loss": model.loss}
    )
    
    # Fetch global model
    if epoch % 10 == 0:
        global_model = client.fetch_global_model()
        model.update_weights(global_model)

# Graceful shutdown
client.disconnect()
```

### Example 4: Monitoring and Metrics

Access real-time metrics via the SDK:

```go
package main

import (
    "github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/monitoring"
)

func main() {
    // Get metrics collector
    metrics := monitoring.GetCollector()
    
    // Query federated learning metrics
    convergence := metrics.GetConvergenceRate()
    log.Printf("Convergence rate: %.4f", convergence)
    
    // Privacy budget consumption
    budget := metrics.GetPrivacyBudgetRemaining()
    log.Printf("Remaining privacy budget: ε=%.2f", budget)
    
    // Node health
    health := metrics.GetNodeHealth()
    log.Printf("Node health score: %d/100", health.Score)
    log.Printf("Uptime: %s", health.Uptime)
    
    // Network statistics
    netStats := metrics.GetNetworkStats()
    log.Printf("Connected peers: %d", netStats.PeerCount)
    log.Printf("Sync latency: %dms", netStats.AvgLatencyMs)
}
```

---

## 📖 API Reference

### gRPC API

The node exposes a gRPC API for programmatic interaction:

**Endpoint**: `localhost:9000`

**Service Definition**: `proto/sovereign_map.proto`

#### Methods

**RegisterNode**
```protobuf
rpc RegisterNode(RegistrationRequest) returns (RegistrationResponse);
```

**SubmitModelUpdate**
```protobuf
rpc SubmitModelUpdate(ModelUpdate) returns (UpdateResponse);
```

**FetchGlobalModel**
```protobuf
rpc FetchGlobalModel(ModelRequest) returns (AggregateModel);
```

**GetNodeStatus**
```protobuf
rpc GetNodeStatus(StatusRequest) returns (StatusUpdate);
```

### REST API

The backend also provides a REST API:

**Base URL**: `http://localhost:8080/api/v1`

#### Endpoints

**Health Check**
```
GET /health
Response: {"status": "healthy", "uptime": "24h"}
```

**Submit Model Update**
```
POST /updates
Body: {
  "node_id": "node-001",
  "weights": [0.1, 0.2, 0.3],
  "proof": "0x...",
  "timestamp": 1234567890
}
Response: {"accepted": true, "round": 42}
```

**Fetch Global Model**
```
GET /models/latest
Response: {
  "round": 42,
  "weights": [0.15, 0.25, 0.35],
  "participants": 50,
  "timestamp": 1234567890
}
```

**Node Metrics**
```
GET /metrics
Response: {
  "uptime_seconds": 86400,
  "updates_submitted": 100,
  "privacy_budget_remaining": 0.8,
  "reputation_score": 95.5
}
```

---

## 🖥️ Running a Node

### Hardware Requirements

**Minimum (Development)**:
- CPU: 4 cores
- RAM: 8 GB
- Storage: 50 GB
- Network: 50 Mbps

**Recommended (Production Genesis Node)**:
- NPU: 85 TOPS (e.g., Intel® Core™ Ultra with Intel® AI Boost)
- CPU: 8-core ARM/x86_64
- RAM: 32 GB
- Storage: 512 GB NVMe SSD
- Network: 100 Mbps symmetric, <100ms latency
- TPM: 2.0 module for hardware attestation

### Configuration

Create a `.env` file:

```bash
# Node Identity
NODE_ID=my-node-001
NODE_TYPE=genesis  # genesis, validator, relay

# Network
AGGREGATOR_URL=http://aggregator:8080
BOOTSTRAP_PEERS=node1.example.com:9000,node2.example.com:9000

# Privacy
ENABLE_SGP001=true
PRIVACY_EPSILON=1.0
PRIVACY_DELTA=1e-5
USE_NPU=true

# Security
ENABLE_TPM=true
TPM_PATH=/dev/tpm0

# Storage
DATA_DIR=/var/lib/sovereign-map
STATE_PERSISTENCE=true

# Monitoring
ENABLE_METRICS=true
PROMETHEUS_PORT=9090
```

### Docker Deployment

```bash
# Build images
make build-docker

# Deploy full stack
make deploy

# View logs
docker-compose logs -f node-agent

# Stop gracefully
make stop
```

### Kubernetes Deployment

```bash
# Apply manifests
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# Check status
kubectl get pods -n sovereign-map

# View logs
kubectl logs -f -n sovereign-map deployment/node-agent
```

---

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test ./internal/consensus/...
```

### Integration Tests

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
make test-integration

# Cleanup
docker-compose -f docker-compose.test.yml down
```

### 200-Node Simulation

Test at scale with 200 simulated nodes:

```bash
# Deploy 200-node test environment
./setup-200node-test.sh

# Run test scenario
./phase-4-execute-test.sh

# Capture results
./phase-5-capture-results.sh

# View convergence plot
open audit_results/convergence_plot.png
```

### Performance Benchmarks

```bash
# Run benchmarks
make benchmark

# Results will show:
# - Model update latency
# - Privacy overhead
# - Aggregation throughput
# - Consensus latency
# - Memory usage
```

---

## 🤝 Contributing

I welcome contributions from developers of all skill levels! Here's how you can help:

### Areas of Focus

1. **Protocol Development**
   - Byzantine consensus improvements
   - Privacy mechanism optimizations
   - Cross-chain bridge implementations

2. **Hardware Integration**
   - NPU acceleration for more hardware platforms
   - TPM attestation for ARM devices
   - GPU-accelerated SLAM

3. **SDK Improvements**
   - Language bindings (Rust, Java, C++)
   - Mobile SDKs (iOS, Android)
   - Web Assembly support

4. **Testing & Documentation**
   - Unit test coverage
   - Integration test scenarios
   - API documentation
   - Tutorial content

### Contribution Workflow

1. **Fork the repository**
   ```bash
   gh repo fork rwilliamspbg-ops/Sovereign_Map_Federated_Learning
   ```

2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes**
   - Write tests
   - Update documentation
   - Follow code style guidelines

4. **Run tests**
   ```bash
   make test
   make lint
   ```

5. **Commit with semantic messages**
   ```bash
   git commit -m "feat: add weighted model aggregation"
   git commit -m "fix: resolve race condition in consensus"
   git commit -m "docs: update API reference"
   ```

6. **Push and create PR**
   ```bash
   git push origin feature/your-feature-name
   gh pr create
   ```

### Code Style

- **Go**: Follow [Effective Go](https://golang.org/doc/effective_go)
- **Python**: Follow [PEP 8](https://pep8.org/)
- **TypeScript**: Follow [Airbnb Style Guide](https://github.com/airbnb/javascript)
- **Commit Messages**: Follow [Conventional Commits](https://www.conventionalcommits.org/)

### Pull Request Checklist

- [ ] Tests pass locally (`make test`)
- [ ] Code is linted (`make lint`)
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] No security vulnerabilities introduced
- [ ] Privacy guarantees maintained (SGP-001 compliance)

---

## 💬 Community

Join the Sovereign Map community and connect with other developers:

### Communication Channels

- **Discord**: [discord.gg/sovereignmap](https://discord.gg/sovereignmap)
  - `#sdk-development` - SDK questions and discussions
  - `#node-operators` - Node deployment help
  - `#protocol-research` - Research and proposals
  
- **GitHub Discussions**: [github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/discussions](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/discussions)
  - Technical specifications
  - Feature requests
  - Showcase your projects

- **Twitter/X**: [@SovereignMap](https://twitter.com/SovereignMap)
  - Protocol updates
  - Community highlights

- **Email**: architects@sovereignmap.network
  - Partnership inquiries
  - Security disclosures

### Weekly Office Hours

Join live Q&A sessions:
- **When**: Every Tuesday at 5 PM UTC
- **Where**: Discord voice channel
- **Topics**: SDK usage, protocol design, deployment help

### Community Projects

Check out what others are building:
- [Community Projects Gallery](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/discussions/categories/show-and-tell)
- Share your project with `#built-with-sovereign-map`

---

## 🗺️ Roadmap

### Q2 2026 - SDK v0.3.0

- [ ] Python SDK package release
- [ ] TypeScript/JavaScript SDK
- [ ] Mobile SDK (iOS/Android) alpha
- [ ] Expanded hardware support (Qualcomm NPUs, Apple Neural Engine)
- [ ] Enhanced documentation and tutorials

### Q3 2026 - SDK v1.0.0

- [ ] Production-ready stable release
- [ ] Rust SDK
- [ ] WebAssembly support for browser nodes
- [ ] Advanced privacy features (dynamic epsilon allocation)
- [ ] Performance optimizations (sub-100ms model updates)

### Q4 2026 - Ecosystem Expansion

- [ ] Third-party plugin system
- [ ] Cross-chain bridge SDKs
- [ ] Enterprise-grade tooling
- [ ] Certified training programs
- [ ] Mainnet deployment support

---

## 📄 License

This SDK is released under the **MIT License**. See [LICENSE](./LICENSE) for full details.

**Privacy-Critical Components** (SGP-001 implementation) are additionally covered by our [Privacy Compliance Agreement](./PRIVACY_COMPLIANCE.md), which ensures that any modifications maintain the same privacy guarantees.

---

## 🎯 Build With Me

Sovereign Map is more than a protocol - it's a movement toward data sovereignty and truly decentralized systems. Whether you're:

- A developer interested in privacy-preserving ML
- A researcher exploring federated learning
- A robotics engineer building autonomous systems
- An entrepreneur creating location-based services
- A node operator contributing compute resources

**There's a place for you in this ecosystem.**

### Get Started Today

1. ⭐ **Star this repository** to stay updated
2. 📖 **Read the [Technical Spec](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/discussions/3)**
3. 💬 **Join our [Discord](https://discord.gg/sovereignmap)**
4. 🚀 **Deploy your first node** (see [Running a Node](#-running-a-node))
5. 🤝 **Make your first contribution** (see [Contributing](#-contributing))

### Need Help?

- Browse [Discussions](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/discussions)
- Check [Issues](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/issues) for known problems
- Ask in [Discord #sdk-help](https://discord.gg/sovereignmap)
- Email: architects@sovereignmap.network

---

<p align="center">
  <strong>Every node is sovereign. Every map is private. Every contribution matters.</strong>
</p>

<p align="center">
  <em>Let's build the future of spatial computing - together.</em>
</p>

<p align="center">
  Made with 🗺️ by the Sovereign Map community
</p>
