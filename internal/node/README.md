# FL Node Integration - Phase 5

This package implements federated learning nodes with full blockchain integration.

## Overview

FL nodes perform local model training and submit results on-chain for rewards. The integration creates a complete flow:

1. **Local Training** → Node trains model locally and computes new weights
2. **Submit Transaction** → Create TxTypeFlRound transaction for on-chain recording
3. **Block Inclusion** → Wait for consensus to include transaction in block
4. **Claim Rewards** → Query blockchain for accumulated rewards and update node state

## Core Components

### FLNode

Represents a single federated learning training participant.

**Main Methods:**
- `TrainRound(ctx)` - Execute local model training
- `SubmitFlRound(ctx, result, checkpoint)` - Submit round results to blockchain
- `WaitForBlockInclusion(ctx, txID, timeout)` - Poll for block inclusion
- `ClaimRewards(ctx)` - Query and claim accumulated rewards
- `CompleteRound(ctx, blockHeight)` - Mark round as finalized
- `GetState()` - Retrieve current node state

**Blockchain Integration:**
- Submits transactions directly to mempool
- Tracks pending transactions locally
- Updates on-chain state when transactions are confirmed
- Accumulates rewards based on round participation

### NodePool

Manages multiple FL nodes and coordinates rounds.

**Main Methods:**
- `RegisterNode(nodeID, region, stake)` - Add node to pool
- `ExecuteRound(ctx)` - Run a full FL round across all nodes
- `GetNodeState(nodeID)` - Query individual node state
- `GetPoolStats()` - Aggregate pool statistics

**Responsibilities:**
- Distribute transactions from all participating nodes
- Wait for block inclusion
- Process rewards for all participants
- Track overall pool metrics

### RewardDistributor

Calculates node rewards based on performance and stake.

**Reward Formula:**
```
Base Reward: 10,000 tokens per round
Accuracy Bonus: (accuracy - 50%) × 100 tokens (for accuracy > 50%)
Stake Multiplier: stake / 100,000 (capped at 2.0x)

Total Reward = (Base + Bonus) × StakeMultiplier
```

### StakeManager

Tracks validator stakes for reward weighting.

## Transaction Types

All FL nodes integrate with six blockchain transaction types:

| Type | Data | Gas | Purpose |
|------|------|-----|---------|
| `TxTypeFlRound` | weights, metrics, checkpoint | 200K | FL training result |
| `TxTypeStake` | amount | 50K | Add to validator stake |
| `TxTypeUnstake` | amount | 50K | Withdraw stake |
| `TxTypeReward` | amount, recipient | 50K | Manual reward |
| `TxTypeSmartContract` | contract, function, args | Variable | Execute contract |
| `TxTypeCheckpoint` | data_hash | 100K | Store milestone |

## Integration with Consensus

FL nodes are tightly integrated with the consensus layer:

1. **Consensus Round Completion**
   ```
   Coordinator.CommitModel()
     → Creates FL round block via blockProposer
     → Includes all pending mempool transactions
     → Distributes rewards to validators
   ```

2. **Node Rewards**
   ```
   Node.SubmitFlRound()
     → Transaction in mempool
     → Included in block
     → Confirmed on-chain
     → Node.ClaimRewards()
   ```

## Example Usage

### Single Node Training

```go
// Setup
node := NewFLNode("node-001", "us-west")
bc := &blockchain.BlockChain{}
mempool := blockchain.NewMempool(1000)
node.SetupBlockchain(bc, mempool, 10000) // 10K initial stake

// Train and submit
ctx := context.Background()
result, _ := node.TrainRound(ctx)
txID, _ := node.SubmitFlRound(ctx, result, checkpoint)

// Wait for inclusion (consensus + block creation)
blockHeight, _ := node.WaitForBlockInclusion(ctx, txID, 10*time.Second)

// Mark complete and claim rewards
node.CompleteRound(ctx, blockHeight)
reward, _ := node.ClaimRewards(ctx)
```

### Multi-Node Pool

```go
// Create pool
pool := NewNodePool(bc, mempool, validators, blockProposer)

// Register nodes
pool.RegisterNode("node-001", "us-west", 10000)
pool.RegisterNode("node-002", "us-east", 15000)
pool.RegisterNode("node-003", "eu-west", 8000)

// Run rounds
for round := 0; round < 100; round++ {
    result, _ := pool.ExecuteRound(ctx)
    fmt.Printf("Round %d: %d nodes participated, %d tokens distributed\n",
        round, len(result.ParticipatingNodes), result.TotalRewardsDistributed)
}

// Check pool stats
stats := pool.GetPoolStats()
fmt.Printf("Total rewards: %d, Total stake: %d\n", 
    stats.TotalRewardsDistributed, stats.TotalStaked)
```

## State Flow

### Node State Lifecycle

```
Created
  ↓
SetupBlockchain
  ↓
TrainRound → SubmitFlRound → WaitForBlockInclusion
  ↓              ↓               ↓
Training      Pending      Confirmed
              Transaction    Transaction
                            ↓
                        CompleteRound
                            ↓
                        ClaimRewards
                            ↓
                        Ready for next round
```

### Transaction Status

```
Transaction Created
  ↓
Added to Mempool
  ↓
Included in Block (consensus)
  ↓
Block Validated (5 layers)
  ↓
Appended to Chain
  ↓
Status: Confirmed
  ↓
Rewards Distributed
```

## Metrics

Each FLNode tracks:
- **Training Performance**: Accuracy, loss, elapsed time per round
- **Transaction Status**: Pending, confirmed, failed counts
- **Rewards**: Accumulated, earned per round
- **Participation**: Total rounds completed
- **Timing**: Average round duration

Access via `node.GetState()`:

```go
state := node.GetState()
fmt.Printf("Node: %s\n", state.NodeID)
fmt.Printf("Rounds: %d\n", state.TotalRoundsCompleted)
fmt.Printf("Rewards: %d tokens\n", state.TotalRewardsEarned)
fmt.Printf("Success Rate: %.1f%%\n", 
    float64(state.SuccessfulTransactions) / 
    float64(state.SuccessfulTransactions + state.FailedTransactions) * 100)
```

## Error Handling

Common error scenarios:

1. **Blockchain Not Configured**
   ```
   "blockchain not configured"
   ```
   Solution: Call `SetupBlockchain()` before any operations

2. **Transaction Timeout**
   ```
   "timeout waiting for block inclusion"
   ```
   Solution: Increase timeout duration or check consensus health

3. **Duplicate Node**
   ```
   "node already registered"
   ```
   Solution: Use unique node IDs

## Testing

Run tests with:

```bash
go test ./internal/node -v
```

Test coverage includes:
- Node creation and configuration (4 tests)
- Training execution (5 tests)
- Transaction submission and tracking (4 tests)
- Reward claiming (2 tests)
- Node pool operations (6 tests)
- Integration scenarios (5 tests)

**Total: 26+ test cases**

## Performance Targets

| Operation | Target | Status |
|-----------|--------|--------|
| TrainRound | <1s | On track |
| SubmitFlRound | <100ms | On track |
| WaitForBlockInclusion | <10s | On track |
| ClaimRewards | <100ms | On track |
| Pool.ExecuteRound (10 nodes) | <5s | On track |

## Next Steps

1. **Integration Testing**: Run full scenarios with consensus
2. **Performance Profiling**: Measure end-to-end latency
3. **Testnet Deployment**: Deploy multiple nodes to testnet
4. **Monitoring**: Setup metrics collection and dashboards

## Architecture Diagram

```
FL Training Nodes
├── Node 1
│   ├── Local Model
│   ├── Training Scheduler
│   └── Blockchain Interface
├── Node 2
│   ├── Local Model
│   ├── Training Scheduler
│   └── Blockchain Interface
└── Node N
    ├── Local Model
    ├── Training Scheduler
    └── Blockchain Interface
        ↓
    Mempool
        ↓
    [Consensus Round]
        ↓
    Blockchain
        ├── Transactions
        ├── State Database
        └── Smart Contracts
            ├── TokenRewards
            ├── ModelRegistry
            └── Governance
```

## References

- [Blockchain Core](../blockchain/README.md) - Block structure and validators
- [Smart Contracts](../blockchain/vm/README.md) - Contract execution
- [Consensus Integration](../../PHASE_3_CONSENSUS_INTEGRATION.md) - Consensus-blockchain bridge
- [Blockchain Progress](../../BLOCKCHAIN_PROGRESS_REPORT.md) - Full project status
