# SovereignMap Proof of Concept - Design Document

## Executive Summary

**Goal**: Build a working proof of concept demonstrating decentralized autonomous navigation with federated learning, mesh networking, and cryptoeconomic incentives.

**Timeline**: 12 weeks (3 months)
**Team Size**: 3-5 engineers
**Budget**: $50K-$150K (depending on scope)

---

## POC Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     SovereignMap POC Stack                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │
│  │   Frontend   │  │   Backend    │  │  Blockchain  │        │
│  │              │  │              │  │              │        │
│  │ - Dashboard  │  │ - Node API   │  │ - Smart      │        │
│  │ - Metrics    │  │ - Aggregator │  │   Contracts  │        │
│  │ - Visualizer │  │ - Mesh Mgr   │  │ - Staking    │        │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘        │
│         │                  │                  │                 │
│         └──────────────────┼──────────────────┘                 │
│                            │                                    │
│         ┌──────────────────┴──────────────────┐                │
│         │                                      │                │
│    ┌────▼─────┐  ┌──────────┐  ┌──────────┐  │                │
│    │  Node 1  │  │  Node 2  │  │  Node N  │  │                │
│    │          │  │          │  │          │  │                │
│    │ - Model  │  │ - Model  │  │ - Model  │  │                │
│    │ - Data   │  │ - Data   │  │ - Data   │  │                │
│    │ - P2P    │  │ - P2P    │  │ - P2P    │  │                │
│    └──────────┘  └──────────┘  └──────────┘  │                │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Implementation Phases

### Phase 1: Core Simulation (Weeks 1-3)
**Goal**: Working simulation with fixed bugs and comprehensive testing

**Deliverables**:
- ✅ Fixed aggregation algorithm with unit tests
- ✅ Simulated mesh network with realistic latency
- ✅ Economic model with configurable parameters
- ✅ Metrics dashboard showing system health

**Tech Stack**:
- Python 3.12
- NumPy, NetworkX
- PyTorch (optional) or custom neural net
- Matplotlib/Plotly for visualization
- Pytest for testing

### Phase 2: Real P2P Network (Weeks 4-6)
**Goal**: Replace simulated mesh with actual peer-to-peer networking

**Deliverables**:
- libp2p-based P2P communication
- Gossip protocol for message propagation
- NAT traversal and discovery
- Real network metrics (latency, bandwidth)

**Tech Stack**:
- libp2p (Python or Go)
- Protocol Buffers for serialization
- Docker for multi-node testing

### Phase 3: Blockchain Integration (Weeks 7-9)
**Goal**: Add on-chain staking and governance

**Deliverables**:
- Smart contracts for staking
- Token distribution mechanism
- Slashing conditions
- On-chain reputation tracking

**Tech Stack**:
- Solidity + Hardhat (Ethereum)
- OR Rust + Anchor (Solana)
- OR CosmWasm (Cosmos)
- Web3.py or ethers.js for integration

### Phase 4: Real-World Data (Weeks 10-12)
**Goal**: Replace mock data with actual sensor inputs

**Deliverables**:
- GPS data ingestion
- Simple camera/LiDAR simulation
- Path prediction model
- Privacy-preserving aggregation

**Tech Stack**:
- OpenCV for image processing
- GPS libraries (gpsd, geopy)
- Differential privacy libraries (Opacus)

---

## Three POC Approaches (Choose One)

### Approach A: Simulation-First (Recommended)
**Best for**: Academic research, algorithm validation, low budget

**Pros**:
- Fast iteration
- Full control over parameters
- Easy to demonstrate Byzantine attacks
- Minimal infrastructure costs

**Cons**:
- Not "real" networking
- Less impressive to investors
- Doesn't validate production feasibility

**Cost**: $10K-$30K
**Time**: 6-8 weeks

### Approach B: Hybrid (Network + Simulation)
**Best for**: Startup MVP, technical demos, balanced approach

**Pros**:
- Real P2P networking validates architecture
- Simulated sensors keep complexity manageable
- Can run on local machines or cloud
- Good middle ground

**Cons**:
- More complex than pure simulation
- Still not fully production-ready
- Network debugging can be time-consuming

**Cost**: $30K-$80K
**Time**: 10-12 weeks

### Approach C: Full Stack (Blockchain + Sensors)
**Best for**: Well-funded projects, production pilot, maximum credibility

**Pros**:
- Closest to production system
- Real economic incentives
- Can integrate with IoT devices
- Impressive for fundraising

**Cons**:
- Highest complexity
- Longest development time
- Most expensive
- Many potential failure points

**Cost**: $80K-$150K
**Time**: 12-16 weeks

---

## Detailed Component Design

### 1. Node Architecture

```python
class SovereignMapNode:
    """
    Each node runs this code independently.
    POC simplifications marked with [POC].
    """
    
    def __init__(self, node_id, config):
        # Identity & Auth
        self.id = node_id
        self.private_key = generate_keypair()  # Ed25519
        self.stake = config.initial_stake
        
        # Networking [POC: Use libp2p or WebRTC]
        self.p2p_host = create_libp2p_host(self.private_key)
        self.peers = set()
        
        # ML Model [POC: Simple linear model]
        self.model = create_navigation_model(
            input_dim=10,  # GPS, velocity, heading, etc.
            output_dim=2   # Next position prediction
        )
        
        # Data [POC: Synthetic or recorded traces]
        self.data_buffer = LocalDataBuffer(max_size=1000)
        self.sensor_interface = SensorInterface(config.sensor_type)
        
        # Blockchain [POC: Local testnet or Ganache]
        self.blockchain = BlockchainClient(
            rpc_url=config.blockchain_rpc,
            contract_address=config.staking_contract
        )
        
        # Metrics
        self.metrics = MetricsCollector()
    
    async def run(self):
        """Main event loop."""
        while True:
            # 1. Collect sensor data
            data = await self.sensor_interface.read()
            self.data_buffer.add(data)
            
            # 2. Train local model
            if self.data_buffer.ready_for_training():
                delta = await self.train_local()
                
                # 3. Broadcast update to peers
                await self.broadcast_update(delta)
            
            # 4. Participate in aggregation
            if self.should_aggregate():
                await self.coordinate_aggregation()
            
            # 5. Update stake based on performance
            await self.update_stake_onchain()
            
            await asyncio.sleep(config.round_duration)
```

### 2. Federated Learning Pipeline

```python
class FederatedLearningCoordinator:
    """
    Coordinates FL rounds across the network.
    POC: Can be centralized initially, then decentralize.
    """
    
    async def run_round(self, round_number):
        print(f"Starting FL Round {round_number}")
        
        # 1. Select participants based on stake
        participants = await self.select_participants(
            min_stake=100,
            max_participants=50
        )
        
        # 2. Request updates from participants
        updates = await self.collect_updates(
            participants,
            timeout=30  # seconds
        )
        
        # 3. Validate updates (check signatures)
        valid_updates = self.validate_updates(updates)
        
        # 4. Aggregate with Byzantine tolerance
        aggregated = stake_weighted_trimmed_mean(
            valid_updates,
            trim_fraction=0.2
        )
        
        # 5. Broadcast aggregated model
        await self.broadcast_model(aggregated)
        
        # 6. Distribute rewards
        await self.distribute_rewards(
            participants,
            valid_updates
        )
        
        # 7. Log metrics
        self.metrics.record_round(
            round_number=round_number,
            participants=len(participants),
            valid_updates=len(valid_updates),
            aggregation_success=aggregated is not None
        )
```

### 3. P2P Networking Layer

```python
# Using libp2p (pseudocode)
class MeshNetwork:
    """
    Handles peer discovery and message routing.
    POC: Start with 5-10 nodes on local network.
    """
    
    def __init__(self, node_id, listen_port):
        self.node_id = node_id
        self.host = None
        self.dht = None  # Kademlia DHT for peer discovery
        self.pubsub = None  # GossipSub for broadcasts
        
    async def start(self):
        # Create libp2p host
        self.host = await new_host(
            listen_addrs=[f"/ip4/0.0.0.0/tcp/{self.listen_port}"]
        )
        
        # Setup DHT for peer discovery
        self.dht = KademliaDHT(self.host)
        await self.dht.bootstrap()
        
        # Setup pubsub for broadcasts
        self.pubsub = GossipSub(self.host)
        await self.pubsub.subscribe("sovereignmap/updates")
        await self.pubsub.subscribe("sovereignmap/aggregation")
        
    async def broadcast_update(self, update):
        """Broadcast model update to all peers."""
        message = serialize_update(update)
        await self.pubsub.publish("sovereignmap/updates", message)
    
    async def find_peers(self, min_peers=5):
        """Discover peers via DHT."""
        peers = await self.dht.find_peers("sovereignmap")
        for peer in peers[:min_peers]:
            await self.host.connect(peer)
```

### 4. Smart Contract Design

```solidity
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

/**
 * @title SovereignMapStaking
 * @notice Manages staking for SovereignMap nodes
 * POC: Simplified version without upgradability
 */
contract SovereignMapStaking {
    struct Node {
        address owner;
        uint256 stake;
        uint256 contributionScore;
        uint256 lastActiveRound;
        bool isActive;
    }
    
    mapping(bytes32 => Node) public nodes;
    mapping(uint256 => bytes32[]) public roundParticipants;
    
    uint256 public currentRound;
    uint256 public minStake = 100 ether;
    uint256 public rewardPool;
    
    event NodeRegistered(bytes32 indexed nodeId, address owner, uint256 stake);
    event RewardDistributed(bytes32 indexed nodeId, uint256 amount);
    event NodeSlashed(bytes32 indexed nodeId, uint256 amount);
    
    /**
     * Register a new node with initial stake.
     */
    function registerNode(bytes32 nodeId) external payable {
        require(msg.value >= minStake, "Insufficient stake");
        require(!nodes[nodeId].isActive, "Node already registered");
        
        nodes[nodeId] = Node({
            owner: msg.sender,
            stake: msg.value,
            contributionScore: 100, // Start at 100%
            lastActiveRound: currentRound,
            isActive: true
        });
        
        emit NodeRegistered(nodeId, msg.sender, msg.value);
    }
    
    /**
     * Record participation in FL round (called by coordinator).
     */
    function recordParticipation(bytes32[] calldata nodeIds) external {
        currentRound++;
        roundParticipants[currentRound] = nodeIds;
        
        for (uint i = 0; i < nodeIds.length; i++) {
            nodes[nodeIds[i]].lastActiveRound = currentRound;
        }
    }
    
    /**
     * Distribute rewards to participants.
     * POC: Simple equal distribution, production would be stake-weighted.
     */
    function distributeRewards(bytes32[] calldata nodeIds, uint256[] calldata amounts) 
        external 
    {
        require(nodeIds.length == amounts.length, "Length mismatch");
        
        for (uint i = 0; i < nodeIds.length; i++) {
            Node storage node = nodes[nodeIds[i]];
            require(node.isActive, "Inactive node");
            
            node.stake += amounts[i];
            emit RewardDistributed(nodeIds[i], amounts[i]);
        }
    }
    
    /**
     * Slash a node for misbehavior.
     */
    function slashNode(bytes32 nodeId, uint256 slashAmount) external {
        Node storage node = nodes[nodeId];
        require(node.isActive, "Inactive node");
        
        uint256 actualSlash = slashAmount > node.stake ? node.stake : slashAmount;
        node.stake -= actualSlash;
        rewardPool += actualSlash;
        
        if (node.stake < minStake) {
            node.isActive = false;
        }
        
        emit NodeSlashed(nodeId, actualSlash);
    }
    
    /**
     * Withdraw stake (requires cooldown period).
     */
    function withdrawStake(bytes32 nodeId) external {
        Node storage node = nodes[nodeId];
        require(msg.sender == node.owner, "Not owner");
        require(currentRound - node.lastActiveRound > 10, "Cooldown period");
        
        uint256 amount = node.stake;
        node.stake = 0;
        node.isActive = false;
        
        payable(msg.sender).transfer(amount);
    }
}
```

### 5. Data Collection Interface

```python
class SensorInterface:
    """
    Abstracts sensor data collection.
    POC: Support multiple modes.
    """
    
    def __init__(self, mode="synthetic"):
        self.mode = mode
        
        if mode == "synthetic":
            self.data_generator = SyntheticDataGenerator()
        elif mode == "recorded":
            self.data_generator = RecordedDataPlayer("data/traces.csv")
        elif mode == "gps":
            self.data_generator = GPSDataCollector()
        elif mode == "camera":
            self.data_generator = CameraDataCollector()
        else:
            raise ValueError(f"Unknown mode: {mode}")
    
    async def read(self):
        """
        Returns:
            SensorReading with fields:
            - timestamp
            - latitude, longitude
            - velocity, heading
            - acceleration
            - [optional] camera_frame
        """
        return await self.data_generator.next_reading()

class SyntheticDataGenerator:
    """Generate realistic synthetic navigation data."""
    
    def __init__(self):
        # Start from a random location
        self.lat = 47.6062 + random.uniform(-0.1, 0.1)  # Seattle area
        self.lon = -122.3321 + random.uniform(-0.1, 0.1)
        self.velocity = 15.0  # m/s
        self.heading = random.uniform(0, 360)  # degrees
        
    async def next_reading(self):
        # Simulate movement
        self.lat += self.velocity * 0.000001 * np.cos(np.radians(self.heading))
        self.lon += self.velocity * 0.000001 * np.sin(np.radians(self.heading))
        
        # Add noise
        self.velocity += random.uniform(-1, 1)
        self.heading += random.uniform(-5, 5)
        
        # Keep velocity in reasonable range
        self.velocity = np.clip(self.velocity, 5, 30)
        
        return SensorReading(
            timestamp=time.time(),
            latitude=self.lat,
            longitude=self.lon,
            velocity=self.velocity,
            heading=self.heading % 360,
            acceleration=random.uniform(-2, 2)
        )
```

---

## Testing Strategy

### Unit Tests
```python
# test_aggregation.py
def test_stake_weighted_aggregation_correctness():
    """Verify stake weighting works correctly."""
    updates = [
        {"delta": {"weights": np.array([10.0])}, "stake": 90},
        {"delta": {"weights": np.array([1.0])}, "stake": 10},
    ]
    result, _ = stake_weighted_trimmed_mean(updates, trim_fraction=0)
    expected = (10.0 * 90 + 1.0 * 10) / 100
    assert np.isclose(result["weights"][0], expected, atol=0.01)

def test_byzantine_resistance():
    """Verify trimmed mean removes outliers."""
    updates = [
        {"delta": {"weights": np.array([5.0])}, "stake": 20},
        {"delta": {"weights": np.array([5.2])}, "stake": 20},
        {"delta": {"weights": np.array([4.8])}, "stake": 20},
        {"delta": {"weights": np.array([100.0])}, "stake": 20},  # Byzantine
        {"delta": {"weights": np.array([-100.0])}, "stake": 20}, # Byzantine
    ]
    result, _ = stake_weighted_trimmed_mean(updates, trim_fraction=0.4)
    # Should be close to 5.0, ignoring the 100 and -100
    assert 4.5 <= result["weights"][0] <= 5.5

# test_networking.py
@pytest.mark.asyncio
async def test_peer_discovery():
    """Test that nodes can discover each other."""
    node1 = await create_test_node(port=4001)
    node2 = await create_test_node(port=4002)
    
    # Wait for discovery
    await asyncio.sleep(5)
    
    assert len(node1.peers) >= 1
    assert len(node2.peers) >= 1

# test_smart_contract.py
def test_stake_registration(smart_contract, accounts):
    """Test node registration with stake."""
    node_id = bytes.fromhex("deadbeef" * 8)
    tx = smart_contract.registerNode(node_id, {"from": accounts[0], "value": "100 ether"})
    
    node = smart_contract.nodes(node_id)
    assert node[0] == accounts[0]  # owner
    assert node[1] == 100 * 10**18  # stake
    assert node[4] == True  # isActive
```

### Integration Tests
```python
# test_integration.py
@pytest.mark.asyncio
@pytest.mark.integration
async def test_full_fl_round():
    """Test a complete federated learning round."""
    # Start 5 nodes
    nodes = [await create_node(i) for i in range(5)]
    
    # Wait for mesh formation
    await asyncio.sleep(10)
    
    # Start FL coordinator
    coordinator = FLCoordinator(nodes)
    
    # Run one round
    result = await coordinator.run_round(round_number=1)
    
    # Verify all nodes participated
    assert result["participants"] == 5
    assert result["valid_updates"] == 5
    assert result["aggregation_success"] == True
    
    # Verify models updated
    for node in nodes:
        assert node.model.last_update_round == 1
```

### Load Tests
```python
# test_load.py
@pytest.mark.load
async def test_100_node_scalability():
    """Test system with 100 nodes."""
    nodes = [await create_node(i) for i in range(100)]
    coordinator = FLCoordinator(nodes)
    
    start_time = time.time()
    result = await coordinator.run_round(round_number=1)
    duration = time.time() - start_time
    
    # Should complete in under 60 seconds
    assert duration < 60
    assert result["participants"] >= 50
    
@pytest.mark.load  
async def test_network_partition_recovery():
    """Test recovery from network partition."""
    nodes = [await create_node(i) for i in range(10)]
    
    # Partition network into two groups
    await partition_network(nodes[:5], nodes[5:])
    
    # Run FL rounds in partitions
    result1 = await run_round(nodes[:5])
    result2 = await run_round(nodes[5:])
    
    # Heal partition
    await heal_network(nodes)
    await asyncio.sleep(10)
    
    # Run round with full network
    result3 = await run_round(nodes)
    
    # Should have all nodes participating again
    assert result3["participants"] == 10
```

---

## Deployment Architecture

### Local Development
```yaml
# docker-compose.yml
version: '3.8'

services:
  # Blockchain (Ganache for testing)
  ganache:
    image: trufflesuite/ganache:latest
    ports:
      - "8545:8545"
    command: --accounts 20 --mnemonic "test test test test test test test test test test test junk"
  
  # Node instances
  node-1:
    build: .
    environment:
      - NODE_ID=1
      - P2P_PORT=4001
      - BLOCKCHAIN_RPC=http://ganache:8545
      - SENSOR_MODE=synthetic
    ports:
      - "4001:4001"
      - "8001:8000"  # API
    depends_on:
      - ganache
  
  node-2:
    build: .
    environment:
      - NODE_ID=2
      - P2P_PORT=4002
      - BLOCKCHAIN_RPC=http://ganache:8545
      - SENSOR_MODE=synthetic
    ports:
      - "4002:4002"
      - "8002:8000"
    depends_on:
      - ganache
  
  node-3:
    build: .
    environment:
      - NODE_ID=3
      - P2P_PORT=4003
      - BLOCKCHAIN_RPC=http://ganache:8545
      - SENSOR_MODE=synthetic
    ports:
      - "4003:4003"
      - "8003:8000"
    depends_on:
      - ganache
  
  # Dashboard
  dashboard:
    build: ./dashboard
    ports:
      - "3000:3000"
    environment:
      - NODE_APIS=http://node-1:8000,http://node-2:8000,http://node-3:8000
```

### Cloud Deployment (AWS)
```
┌─────────────────────────────────────────────────┐
│              AWS VPC (us-west-2)                │
├─────────────────────────────────────────────────┤
│                                                 │
│  ┌──────────────────────────────────────────┐  │
│  │     Application Load Balancer             │  │
│  └────────────┬─────────────────────────────┘  │
│               │                                 │
│  ┌────────────┴─────────────────────────────┐  │
│  │         ECS Fargate (Auto-scaling)        │  │
│  │                                            │  │
│  │  ┌─────┐  ┌─────┐  ┌─────┐  ┌─────┐      │  │
│  │  │Node1│  │Node2│  │Node3│  │NodeN│      │  │
│  │  └─────┘  └─────┘  └─────┘  └─────┘      │  │
│  └────────────────────────────────────────────┘  │
│                                                 │
│  ┌──────────────────────────────────────────┐  │
│  │     RDS PostgreSQL (Metrics Storage)      │  │
│  └──────────────────────────────────────────┘  │
│                                                 │
│  ┌──────────────────────────────────────────┐  │
│  │     ElastiCache Redis (State Cache)       │  │
│  └──────────────────────────────────────────┘  │
│                                                 │
└─────────────────────────────────────────────────┘

External: Ethereum Testnet (Sepolia or Goerli)
```

---

## Metrics & Monitoring

### Key Metrics to Track

**System Health**:
- Network connectivity rate (target: >95%)
- Average peer count per node (target: >5)
- Message propagation latency (target: <2s)

**FL Performance**:
- Aggregation success rate (target: 100%)
- Average participation rate (target: >80%)
- Model convergence speed (rounds to target loss)

**Economic Metrics**:
- Total value locked (stake)
- Stake distribution (Gini coefficient)
- Average reward per node per round
- Slashing events count

**Security Metrics**:
- Byzantine attack detection rate
- Invalid update rejection rate
- Signature verification failures

### Monitoring Dashboard

```python
# dashboard/app.py (Streamlit)
import streamlit as st
import requests

st.title("SovereignMap POC Dashboard")

# Fetch data from nodes
nodes = ["http://localhost:8001", "http://localhost:8002", "http://localhost:8003"]

col1, col2, col3 = st.columns(3)

with col1:
    st.metric("Active Nodes", len([n for n in nodes if is_alive(n)]))
with col2:
    st.metric("Current Round", get_current_round())
with col3:
    st.metric("Total Stake", f"{get_total_stake():.2f} ETH")

# Network topology
st.subheader("Network Topology")
topology = get_network_topology()
st.plotly_chart(create_network_graph(topology))

# FL convergence
st.subheader("Model Convergence")
losses = get_training_losses()
st.line_chart(losses)

# Stake distribution
st.subheader("Stake Distribution")
stakes = get_stake_distribution()
st.bar_chart(stakes)

# Recent events
st.subheader("Recent Events")
events = get_recent_events(limit=10)
st.table(events)
```

---

## Risk Mitigation

### Technical Risks

| Risk | Impact | Probability | Mitigation |
|------|---------|-------------|------------|
| Network partition | High | Medium | Economic penalties, automatic reconnection |
| Byzantine attacks | High | Medium | Trimmed mean, stake weighting |
| Smart contract bugs | Critical | Low | Formal verification, audits, testnet first |
| Scalability issues | Medium | High | Start small (10 nodes), optimize later |
| Data privacy leaks | High | Low | Differential privacy, secure aggregation |

### Non-Technical Risks

| Risk | Impact | Probability | Mitigation |
|------|---------|-------------|------------|
| Regulatory issues | High | Medium | Legal counsel, compliance by design |
| Lack of adoption | High | High | Focus on clear use case, good UX |
| Token economics failure | Medium | Medium | Game theory analysis, simulations |
| Competitor advantage | Low | High | Open source, network effects |

---

## Success Criteria

### Minimum Viable POC
- ✅ 5 nodes running simultaneously
- ✅ FL aggregation completes successfully
- ✅ Network remains connected >90% of time
- ✅ Staking contract deployed on testnet
- ✅ Basic dashboard showing metrics

### Successful POC
- ✅ 10+ nodes on distributed infrastructure
- ✅ Withstands 20% Byzantine nodes
- ✅ Model converges to target accuracy
- ✅ Real GPS data integration
- ✅ Professional dashboard with analytics

### Exceptional POC
- ✅ 50+ nodes on public testnet
- ✅ Real-time privacy guarantees (differential privacy)
- ✅ Sub-second message propagation
- ✅ Economic model validated through simulations
- ✅ Production-ready code quality
- ✅ Research paper quality documentation

---

## Budget Breakdown

### Option A: Simulation-First ($10K-$30K)
- Development (1 engineer × 8 weeks): $16K
- Testing & QA: $2K
- Infrastructure (AWS/DO): $1K
- Documentation: $1K
- **Total: ~$20K**

### Option B: Hybrid ($30K-$80K)
- Development (2 engineers × 10 weeks): $40K
- Smart contract audit: $10K
- Testing & QA: $5K
- Infrastructure: $3K
- Documentation & design: $2K
- **Total: ~$60K**

### Option C: Full Stack ($80K-$150K)
- Development (3 engineers × 12 weeks): $72K
- Smart contract audit: $20K
- Security audit: $15K
- Testing & QA: $10K
- Infrastructure: $8K
- Hardware (test devices): $10K
- Documentation & design: $5K
- Legal consultation: $10K
- **Total: ~$150K**

---

## Next Steps

### Week 1: Setup & Planning
1. Choose POC approach (A, B, or C)
2. Set up development environment
3. Create GitHub repository
4. Write detailed technical specifications
5. Set up CI/CD pipeline

### Week 2-3: Core Implementation
1. Fix and test aggregation algorithm
2. Build simulation framework
3. Implement metrics collection
4. Create basic dashboard

### Week 4+: Extended Features
Follow phase plan based on chosen approach

---

## Conclusion

This POC design provides three paths forward depending on your goals:

1. **Academic/Research**: Choose Approach A for fast validation
2. **Startup MVP**: Choose Approach B for balanced demo
3. **Production Pilot**: Choose Approach C for maximum credibility

All approaches share the same core algorithms (with bugs fixed) but differ in infrastructure complexity and integration depth.

**Recommended Start**: Approach A to validate algorithms, then upgrade to B if successful.
