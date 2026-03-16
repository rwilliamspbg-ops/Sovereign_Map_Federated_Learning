# Blockchain Implementation Guide

Welcome to Sovereign Map's blockchain layer! This guide explains the structure, components, and integration points.

## Overview

This blockchain implementation is designed to:
- **Extend** Sovereign Map's Byzantine Fault Tolerant Federated Learning with on-chain state management
- **Integrate** seamlessly with existing consensus, P2P, and FL components
- **Support** smart contracts, staking, and on-chain rewards
- **Scale** to 5000+ nodes with sub-second block times

## Directory Structure

```
internal/blockchain/
├── block.go          # Block structure, transactions, BlockChain type
├── state.go          # StateDatabase with Merkle trees
├── mempool.go        # Transaction mempool for pending transactions
├── validators.go     # ValidatorSet for staking and rewards
├── validator.go      # BlockValidator for validation rules
└── README.md         # This file
```

## Core Components

### 1. Block & Transaction (`block.go`)

**Block Structure:**
```go
type Block struct {
    Header       BlockHeader      // Metadata (index, hash, timestamp)
    Transactions []Transaction    // Ordered transactions
    MerkleRoot   string          // Merkle root of transactions
    StateRoot    string          // Merkle root of resulting state
    ProofData    map[string][]byte // BFT proofs, attestations, etc
}
```

**Transaction Types:**
- `fl_round` - Federated learning round submission
- `stake` - Node staking commitment
- `unstake` - Stake withdrawal
- `reward` - Reward distribution
- `smart_contract` - Contract deployment/invocation
- `checkpoint` - Distributed checkpoint commit (IPFS)

**Usage:**
```go
// Create transaction
txn := &blockchain.Transaction{
    ID:    "txn_001",
    Type:  blockchain.TxTypeFlRound,
    From:  "node_1",
    Nonce: 0,
    Data: map[string]interface{}{
        "round_id":   "round_1",
        "model_hash": "0x1234...",
        "accuracy":   0.95,
    },
}

// Add to blockchain mempool
blockchain.Mempool.AddTransaction(txn)

// Create block from pending transactions
selectedTxns := blockchain.Mempool.GetTransactionsForBlock(1000)
block := &Block{
    Header: BlockHeader{
        Index:        1,
        PreviousHash: genesisBlock.Header.Hash,
        ValidatorID:  "validator_1",
        Timestamp:    time.Now().Unix(),
    },
    Transactions: selectedTxns,
}

// Compute hashes
block.MerkleRoot = ComputeMerkleRoot(block.Transactions)
block.StateRoot = blockchain.StateDB.ComputeRoot()
block.Header.Hash = block.ComputeHash()

// Add to chain
blockchain.AppendBlock(block)
```

### 2. State Database (`state.go`)

Manages on-chain state with Merkle tree verification for every block.

**Key Features:**
- Deterministic Merkle root computation
- State snapshots at each block height
- Thread-safe concurrent access

**Usage:**
```go
// Set state
blockchain.StateDB.Set("stake:node_1", uint64(100000000))
blockchain.StateDB.Set("reward:node_1", uint64(500000))

// Get state
stake, _ := blockchain.StateDB.Get("stake:node_1")

// Compute root
root := blockchain.StateDB.ComputeRoot()

// Record snapshot
blockchain.StateDB.RecordSnapshot(blockHeight)

// Replay attack prevention
snapshot, _ := blockchain.StateDB.GetSnapshot(5)
```

### 3. Transaction Mempool (`mempool.go`)

Manages pending transactions awaiting block inclusion.

**Features:**
- Priority by gas price
- Nonce-based replay protection
- Configurable max size and transactions per block

**Usage:**
```go
// Add transaction
err := blockchain.Mempool.AddTransaction(txn)

// Get transactions for next block
selectedTxns := blockchain.Mempool.GetTransactionsForBlock(1000)

// Remove transaction (when included in block)
blockchain.Mempool.RemoveTransaction(txn.ID)

// Check mempool size
if blockchain.Mempool.Size() > 50000 {
    // Mempool getting full
}
```

### 4. Validator Set (`validators.go`)

Manages consensus validators with staking, rewards, and slashing.

**Key Features:**
- Weighted random validator selection by stake
- Epoch-based rotation
- Slash penalties for misbehavior
- Jail mechanism for repeat offenders

**Usage:**
```go
// Add validator
blockchain.ValidatorSet.AddValidator("node_1", 10000000) // 10M tokens

// Stake more tokens
blockchain.ValidatorSet.Stake("node_1", 5000000)

// Select validators for next block
nextValidators := blockchain.ValidatorSet.SelectValidators(1)

// Distribute block rewards
blockchain.ValidatorSet.DistributeRewards(blockHeight, 1000000, []string{"node_1"})

// Slash validator for misbehavior
blockchain.ValidatorSet.SlashValidator("node_1", 100000)

// Rotate epoch
blockchain.ValidatorSet.RotateEpoch(blockHeight)
```

### 5. Block Validator (`validator.go`)

Comprehensive validation engine for blocks.

**Validates:**
1. Block header (index, hash, previous hash, timestamp)
2. Individual transactions (type, signature, nonce)
3. Merkle root of transactions
4. State root from transaction execution
5. Cryptographic proofs (BFT, ZK-SNARK, TPM)

**Usage:**
```go
validator := NewBlockValidator(blockchain, blockchain.StateDB)

// Validate single block
if err := validator.ValidateBlock(block); err != nil {
    log.Fatal("Invalid block:", err)
}

// Validate sequence
if err := validator.ValidateBlockSequence(blocks); err != nil {
    log.Fatal("Invalid block sequence:", err)
}
```

## Integration with Existing Systems

### Consensus Module (`internal/consensus/`)

**Current State:** Uses BFT voting among nodes to commit FL model updates

**Integration Points:**
```go
// In consensus/coordinator.go, after consensus is reached:

func (c *Coordinator) CommitBlock(round *ConsensusRound) error {
    // Build block from FL round
    block := &Block{
        Header: BlockHeader{
            Index:        c.blockchain.Height() + 1,
            PreviousHash: c.blockchain.Tip.Header.Hash,
            ValidatorID:  c.nodeID,
            Timestamp:    time.Now().Unix(),
        },
        Transactions: []Transaction{
            {
                Type:      TxTypeFlRound,
                From:      c.nodeID,
                Data: map[string]interface{}{
                    "round_id":   round.ID,
                    "model_hash": hashModel(round.AggregateModel),
                },
            },
        },
    }
    
    // Compute state root
    block.StateRoot = c.blockchain.StateDB.ComputeRoot()
    block.MerkleRoot = ComputeMerkleRoot(block.Transactions)
    block.Header.Hash = block.ComputeHash()
    
    // Append to blockchain
    return c.blockchain.AppendBlock(block)
}
```

### P2P Network (`internal/p2p/`)

**Current State:** Uses libp2p for mesh networking

**New Responsibilities:**
- Broadcast pending transactions to mempool
- Sync blocks with new peers
- Validate block signatures and proofs

```go
// In p2p message handler:
case "transaction":
    txn := &blockchain.Transaction{}
    json.Unmarshal(msg.Data, txn)
    blockchain.Mempool.AddTransaction(txn)
    
case "block":
    block := &Block{}
    json.Unmarshal(msg.Data, block)
    validator.ValidateBlock(block)
    blockchain.AppendBlock(block)
```

### Federated Learning Module (`packages/core/`)

**Current State:** Submits model updates to consensus

**New Flow:**
1. Node trains local model (FL round)
2. Submits to consensus (existing)
3. **NEW:** Creates blockchain transaction with model hash
4. **NEW:** Receives on-chain reward if consensus achieved

```go
// In node training flow:
func (n *Node) ExecuteFlRound(round *FLRound) error {
    // ... existing training code ...
    
    // NEW: Create blockchain transaction
    txn := &blockchain.Transaction{
        ID:        uuid.New().String(),
        Type:      blockchain.TxTypeFlRound,
        From:      n.nodeID,
        Nonce:     n.txnNonce++,
        Timestamp: time.Now().Unix(),
        Data: map[string]interface{}{
            "round_id":       round.ID,
            "model_hash":     hashModel(n.localModel),
            "accuracy":       round.AccuracyMetric,
            "convergence":    round.ConvergenceTime,
            "proof":          n.zkProof,
        },
    }
    
    // Add to mempool
    n.blockchain.Mempool.AddTransaction(txn)
    
    // NEW: Wait for block inclusion and collect reward
    if blockIncluded := n.waitForTransactionInBlock(txn.ID, 30*time.Second); blockIncluded {
        reward := n.calculateFlReward(round)
        n.blockchain.StateDB.Set(
            fmt.Sprintf("reward:%s", n.nodeID),
            reward,
        )
    }
    
    return nil
}
```

## Mining & Block Production

### Proof-of-Authority (Recommended for v1)

Selected validator produces next block without computational work:

```go
// In block proposer:
selectedValidator := blockchain.ValidatorSet.SelectValidators(1)[0]

if selectedValidator.NodeID == myNodeID {
    // Propose next block
    block := buildProposedBlock()
    blockchain.AppendBlock(block)
    
    // Distribute rewards
    blockchain.ValidatorSet.DistributeRewards(
        block.Header.Index,
        1000000, // Base reward
        []string{selectedValidator.NodeID},
    )
}
```

### Future: Proof-of-Stake

Weighted random selection ensures fair participation:

```go
// Higher stake = higher chance of block production
selectedValidator := blockchain.ValidatorSet.SelectValidators(1)[0]

// Reward proportional to stake
baseReward := uint64(1000000)
stakerReward := baseReward * selectedValidator.StakedAmount / vs.TotalStake
```

## Smart Contracts

The blockchain supports smart contract execution through a lightweight VM:

```go
// Deploy contract
txn := &blockchain.Transaction{
    Type: blockchain.TxTypeSmartContract,
    Data: map[string]interface{}{
        "bytecode":   contractBytecode,
        "abi":        contractABI,
    },
}

// Call contract function
callTxn := &blockchain.Transaction{
    Type: blockchain.TxTypeSmartContract,
    Data: map[string]interface{}{
        "contract_address": "0xabc...",
        "function":         "stake",
        "params":           []interface{}{1000000},
    },
}
```

**Example: TokenRewardsContract**
```solidity
contract TokenRewards {
    mapping(address => uint256) public nodeStakes;
    
    function stake(uint256 amount) external {
        nodeStakes[msg.sender] += amount;
    }
    
    function distributeFlRewards(address[] calldata nodes, uint256[] calldata rewards) external {
        for (uint i = 0; i < nodes.length; i++) {
            nodeStakes[nodes[i]] += rewards[i];
        }
    }
}
```

## Testing

```bash
# Run blockchain tests
go test ./internal/blockchain/...

# Test block validation
go test ./internal/blockchain/ -run TestValidateBlock -v

# Test mempool
go test ./internal/blockchain/ -run TestMempool -v

# Test consensus integration
go test ./internal/consensus/ -run TestBlockCommit -v
```

## Performance Targets

| Metric | Target | Current |
|--------|--------|---------|
| Block time | 1-2 sec | — |
| Transactions per block | 1000+ | — |
| Transaction throughput | 1000+ tx/sec | — |
| State root verification | <50ms | — |
| Block validation | <100ms | — |
| Validator set size | 100+ | — |
| Network size | 5000+ nodes | — |

## Roadmap

- [ ] **Phase 1:** Core blockchain (blocks, state, mempool, validators)
- [ ] **Phase 2:** Integration with consensus module
- [ ] **Phase 3:** Smart contract VM implementation
- [ ] **Phase 4:** FL→blockchain transaction bridging
- [ ] **Phase 5:** Full-featured Proof-of-Stake
- [ ] **Phase 6:** Cross-chain bridges and scaling

## References

- Existing consensus: `internal/consensus/`
- Existing P2P: `internal/p2p/`
- Existing FL: `packages/core/`
- Tokenomics: `tokenomics_metrics_exporter.py`
- ZK proofs: `internal/hybrid/verifier.go`, `internal/api/`

## Questions?

Refer to [BLOCKCHAIN_TRANSFORMATION.md](../BLOCKCHAIN_TRANSFORMATION.md) for strategic planning and detailed architecture.
