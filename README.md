[README.md](https://github.com/user-attachments/files/25153488/README.md)
# SovereignMap Proof of Concept

A working proof of concept demonstrating **decentralized federated learning for autonomous navigation** with Byzantine fault tolerance, mesh networking, and cryptoeconomic incentives.

## 🎯 What This Demonstrates

1. **Federated Learning**: Multiple nodes collaboratively train a navigation model without sharing raw data
2. **Byzantine Tolerance**: System resists up to 20% malicious nodes through stake-weighted trimmed mean aggregation
3. **Mesh Networking**: Peer-to-peer topology with dynamic connections
4. **Economic Incentives**: Stake-based rewards and penalties for honest participation
5. **Real-Time Monitoring**: Web dashboard showing live system metrics

## 🚀 Quick Start

### Option 1: Run Locally (Simplest)

```bash
# Install dependencies (only numpy needed)
pip install numpy

# Run the POC
python sovereignmap_poc.py

# Open dashboard in browser
# http://localhost:8000
```

The system will:
- Create 10 simulated nodes
- Run 50 federated learning rounds
- Display real-time metrics on the dashboard
- Show stake evolution, participation rates, and network health

### Option 2: Run Tests

```bash
# Run comprehensive test suite
python test_poc.py

# Tests cover:
# - Neural network functionality
# - Byzantine resistance
# - Federated learning rounds
# - Performance benchmarks
```

### Option 3: Docker Deployment (Multi-Container)

```bash
# Build and run all services
docker-compose up --build

# Dashboard available at http://localhost:8000

# Scale nodes dynamically
docker-compose up --scale node-1=5
```

## 📊 Dashboard Features

The web dashboard at `http://localhost:8000` shows:

- **Real-time Metrics**:
  - Active nodes count
  - Current FL round number
  - Participation rate
  - Total stake in system

- **Stake Charts**:
  - Average stake over time
  - Total stake evolution
  - Individual node performance

- **Node Status**:
  - Online/offline status
  - Current stake per node
  - Peer connections

## 🏗️ Architecture

```
┌─────────────────────────────────────────────┐
│           SovereignMap System               │
├─────────────────────────────────────────────┤
│                                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │  Node 1  │  │  Node 2  │  │  Node N  │ │
│  │          │  │          │  │          │ │
│  │ - Model  │  │ - Model  │  │ - Model  │ │
│  │ - Data   │  │ - Data   │  │ - Data   │ │
│  │ - Stake  │  │ - Stake  │  │ - Stake  │ │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘ │
│       │             │             │        │
│       └─────────────┼─────────────┘        │
│                     │                      │
│            ┌────────▼────────┐            │
│            │  FL Coordinator  │            │
│            │                  │            │
│            │ - Aggregation    │            │
│            │ - Rewards        │            │
│            │ - Validation     │            │
│            └──────────────────┘            │
│                                             │
└─────────────────────────────────────────────┘
```

## 🔧 Components

### 1. NavigationModel
Simple neural network (10 inputs → 20 hidden → 2 outputs) for position prediction.
- Input: GPS, velocity, heading, time, trigonometric features
- Output: Next position (latitude, longitude)
- Training: Backpropagation with MSE loss

### 2. SovereignMapNode
Each autonomous entity in the network:
- Collects sensor data (synthetic GPS traces)
- Trains local model on private data
- Creates signed model updates
- Participates in federated rounds
- Maintains stake balance

### 3. FederatedCoordinator
Orchestrates the learning process:
- Collects updates from nodes
- Verifies signatures
- Aggregates using Byzantine-tolerant algorithm
- Distributes rewards/penalties
- Tracks system metrics

### 4. Stake-Weighted Aggregation
Byzantine-resistant aggregation (fixed version from bug analysis):
```python
def stake_weighted_trimmed_mean(updates, trim_fraction=0.2):
    # 1. Calculate stake-weighted contributions
    # 2. Sort parameter values
    # 3. Trim outliers (top/bottom 20%)
    # 4. Weight remaining values by stake
    # 5. Return aggregated model
```

Key fix: Correctly maps stakes to sorted values (see BUG_FIXES_COMPARISON.md).

## 📈 Metrics Tracked

**System Health**:
- Network connectivity
- Node online status
- Peer connections per node

**Learning Performance**:
- Participation rate per round
- Aggregation success rate
- Model convergence (training loss)

**Economic Metrics**:
- Average stake across nodes
- Total stake in system
- Stake distribution
- Rewards/penalties per round

**Security**:
- Invalid update rejection rate
- Signature verification failures
- Byzantine attack detection

## 🧪 Testing

The test suite (`test_poc.py`) includes:

**Unit Tests**:
- Neural network forward/backward pass
- Weight serialization
- Sensor data generation
- Feature extraction
- Aggregation correctness
- Byzantine resistance

**Integration Tests**:
- Node lifecycle
- FL round execution
- Multi-round training
- Stake updates

**Performance Tests**:
- 100-node aggregation speed
- Training latency
- Memory usage

Expected output:
```
UNIT TESTS: 7/7 passed
INTEGRATION TESTS: 6/6 passed
PERFORMANCE TESTS: 2/2 passed
✓ ALL TESTS PASSED!
```

## 🎮 Usage Examples

### Basic Simulation

```python
import asyncio
from sovereignmap_poc import SovereignMapNode, FederatedCoordinator

async def run_simulation():
    # Create nodes
    nodes = [SovereignMapNode(i, initial_stake=1000) for i in range(5)]
    
    # Collect initial data
    await asyncio.gather(*[node.collect_data(duration=3) for node in nodes])
    
    # Create coordinator
    coordinator = FederatedCoordinator(nodes)
    
    # Run FL rounds
    for round_num in range(10):
        metrics = await coordinator.run_round()
        print(f"Round {round_num}: {metrics.participation_rate*100:.1f}% participation")

asyncio.run(run_simulation())
```

### Custom Sensor Data

```python
from sovereignmap_poc import SensorReading

# Implement custom sensor
class GPSSensor:
    def __init__(self, device_path):
        self.device = open_gps_device(device_path)
    
    async def read(self):
        lat, lon, vel, heading = self.device.read_position()
        return SensorReading(
            timestamp=time.time(),
            latitude=lat,
            longitude=lon,
            velocity=vel,
            heading=heading,
            acceleration=0.0,
            node_id=self.node_id
        )

# Use in node
node.sensor = GPSSensor("/dev/ttyUSB0")
```

### Byzantine Attack Simulation

```python
# Create normal nodes
nodes = [SovereignMapNode(i) for i in range(8)]

# Create Byzantine node
byzantine = SovereignMapNode(8, initial_stake=500)

# Override training to send bad updates
def malicious_train(self):
    return np.random.randn(242) * 1000  # Random large values

byzantine.train_local = lambda: malicious_train(byzantine)

# Add to network
nodes.append(byzantine)

# Run FL round - should resist the attack
coordinator = FederatedCoordinator(nodes)
metrics = await coordinator.run_round()
# Byzantine update will be trimmed out
```

## 🔒 Security Features

1. **Cryptographic Signatures**
   - Each update signed with node's private key (simplified in POC)
   - Coordinator verifies signatures before aggregation

2. **Stake-Based Voting**
   - Higher stake = more influence in aggregation
   - Creates economic disincentive for attacks

3. **Outlier Removal**
   - Trimmed mean removes extreme values
   - Tolerates up to 20% Byzantine nodes

4. **Contribution Scoring**
   - Tracks historical participation
   - Penalizes unreliable nodes

## 📊 Performance Characteristics

Tested on standard hardware:

| Metric | Value |
|--------|-------|
| Aggregation latency (10 nodes) | ~50ms |
| Aggregation latency (100 nodes) | ~200ms |
| Training time per round | ~100ms |
| Memory per node | ~50MB |
| Dashboard update rate | 2 seconds |

## 🚧 Known Limitations

This is a POC with several simplifications:

1. **No Real P2P Networking**: Uses simulated mesh, not actual libp2p/WebRTC
2. **Simplified Cryptography**: Uses basic signatures, not full Ed25519
3. **No Blockchain Integration**: Staking is in-memory, not on-chain
4. **Synthetic Data**: Uses generated GPS traces, not real sensors
5. **Centralized Coordinator**: Single point of failure (should be decentralized)
6. **No Differential Privacy**: Model updates not protected from gradient inversion

See `POC_DESIGN.md` for production roadmap.

## 🛠️ Configuration

Edit these variables in `sovereignmap_poc.py`:

```python
# Number of nodes
num_nodes = 10

# Initial stake range
initial_stake = 1000 + random.uniform(-200, 200)

# FL rounds to run
num_rounds = 50

# Aggregation parameters
trim_fraction = 0.2  # Remove top/bottom 20%

# Reward structure
base_reward = 50
participation_bonus = 10
non_participation_penalty = -20

# Dashboard port
dashboard_port = 8000
```

## 📚 Related Documentation

- `SOVEREIGNMAP_ANALYSIS.md` - Complete system analysis and design rationale
- `BUG_FIXES_COMPARISON.md` - Detailed explanation of fixed bugs
- `POC_DESIGN.md` - Proof of concept design with production roadmap

## 🤝 Contributing

This is a research POC. To extend it:

1. **Add Real Networking**: Replace simulated mesh with libp2p
2. **Blockchain Integration**: Add smart contracts for staking (see POC_DESIGN.md for Solidity example)
3. **Privacy Enhancement**: Implement differential privacy on gradients
4. **Real Sensors**: Integrate with GPS modules, cameras, LiDAR
5. **Scalability**: Add hierarchical aggregation for >100 nodes

## 📖 Use Cases

This POC demonstrates technology for:

- **Autonomous Vehicles**: Collaborative learning without centralizing driving data
- **Drone Swarms**: Distributed navigation and coordination
- **IoT Sensor Networks**: Federated environmental monitoring
- **Decentralized Mapping**: User-owned spatial data (OpenStreetMap++)
- **Privacy-Preserving ML**: Any domain requiring collaborative learning without data sharing

## 📄 License

MIT License - See LICENSE file

## 🙏 Acknowledgments

Based on research in:
- Federated Learning (McMahan et al. 2017)
- Byzantine-Robust Aggregation (Blanchard et al. 2017)
- Proof of Stake Consensus (Ethereum Foundation)
- CXL Memory Pooling (Intel, Compute Express Link Consortium)

## 📞 Support

For questions or issues:
1. Check the documentation files
2. Review test failures for diagnostic info
3. Examine dashboard metrics for system health

## 🎯 Success Criteria

POC is successful if:
- ✅ All tests pass
- ✅ 10 nodes run simultaneously
- ✅ FL aggregation succeeds >95% of rounds
- ✅ System resists simulated Byzantine attacks
- ✅ Dashboard shows live metrics
- ✅ Stake distribution evolves over time

Run the POC and verify these criteria!
