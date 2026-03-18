# 🔗 Sovereign Map: Complete Blockchain Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     SOVEREIGN MAP v2.0                           │
│              (FL + Blockchain Hybrid System)                     │
└─────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│                    CONSENSUS LAYER (Existing)                    │
│                  Byzantine Fault Tolerant BFT                    │
│         (Voting for FL model aggregation + block commits)        │
└──────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────┐
│                   BLOCKCHAIN LAYER (NEW)                         │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Block Structure:                                                │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ Block Header                                               │ │
│  │  • Index: 1234                                             │ │
│  │  • Hash: 0x1234...                                         │ │
│  │  • PreviousHash: 0x5678...                               │ │
│  │  • ValidatorID: node_1                                     │ │
│  │  • Timestamp: 2026-03-16T10:00:00Z                       │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  Transactions (Ordered by gas price):                           │
│  ┌────────┬──────────────────────────────────────────────────┐ │
│  │ Txn 1  │ Type: fl_round       | From: node_1             │ │
│  │        │ Round: 1234, Model: 0x9abc..., Acc: 0.95        │ │
│  ├────────┼──────────────────────────────────────────────────┤ │
│  │ Txn 2  │ Type: stake           | From: node_2, Amt: 1M   │ │
│  ├────────┼──────────────────────────────────────────────────┤ │
│  │ Txn 3  │ Type: reward          | To: node_1, Amt: 100k   │ │
│  ├────────┼──────────────────────────────────────────────────┤ │
│  │ Txn 4  │ Type: smart_contract  | Call: stake(1M)         │ │
│  └────────┴──────────────────────────────────────────────────┘ │
│                                                                  │
│  State (Merkle Tree):                                            │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ stake:node_1 → 100000000                                   │ │
│  │ reward:node_1 → 500000                                     │ │
│  │ fl_round:1234 → {model_hash, accuracy, timestamp}        │ │
│  │ StateRoot: 0xabcd...                                       │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────┐
│                 VALIDATOR/STAKE LAYER (NEW)                      │
│                                                                  │
│  Validator Set:                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ node_1: stake=100M  | votes=1234 | rewards=500k | ✓     │  │
│  │ node_2: stake=50M   | votes=890  | rewards=250k | ✓     │  │
│  │ node_3: stake=25M   | votes=456  | rewards=125k | 🔒    │  │
│  │ node_4: stake=10M   | votes=123  | rewards=50k  | ✓     │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  Rewards: Proportional to stake × FL contribution              │
│  Slashing: Penalties for Byzantine behavior                    │
│  Jailing: Automatic jail after 3 slashes                      │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────┐
│                    P2P NETWORK LAYER (Existing)                  │
│             (libp2p mesh: TCP + QUIC + NAT traversal)           │
│                                                                  │
│  Broadcasting:                                                   │
│  • Transactions gossip to all nodes                             │
│  • Blocks broadcast to all peers                                │
│  • State sync between validators                                │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────────┐
│                  APPLICATION LAYER (Existing)                    │
│              Federated Learning + Smart Contracts                │
│                                                                  │
│  FL Nodes:                                                       │
│  • Train local models                                           │
│  • Submit to consensus                                          │
│  • Create blockchain txns                                       │
│  • Collect on-chain rewards                                     │
│                                                                  │
│  Smart Contracts:                                               │
│  • Token distribution                                           │
│  • Model registry                                               │
│  • Governance (SGPs)                                            │
│  • Cross-chain bridges                                          │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## Data Flow: FL Round → On-Chain Reward

```
Node Training Phase
──────────────────
  ↓
  ├─ Train local model
  ├─ Compute accuracy
  └─ Generate ZK proof
      ↓

Consensus Phase (Existing)
──────────────────────────
  ├─ Submit to BFT
  ├─ Nodes vote on aggregate
  ├─ Reach consensus (2/3 threshold)
  └─ Aggregate model committed
      ↓

Blockchain Phase (NEW)
─────────────────────
  ├─ Create FL round transaction
  │  ├─ round_id: "round_1234"
  │  ├─ model_hash: "0x9abc..."
  │  ├─ accuracy: 0.954
  │  └─ proof: ZK_SNARK
  │
  ├─ Add to mempool
  │  └─ Queue: [FlTxn, StakeTxn, RewardTxn, ...]
  │
  ├─ Validator selected (weighted by stake)
  │  └─ node_1 (100M stake) chosen
  │
  ├─ Propose block
  │  ├─ Select top 1000 txns by gas price
  │  ├─ Compute Merkle root
  │  ├─ Execute txns → update state
  │  ├─ Compute state root
  │  └─ Hash block
  │
  ├─ Broadcast block to network
  │  ├─ Other validators validate
  │  ├─ Verify header, txns, merkle, state, proofs
  │  └─ All nodes accept block
  │
  └─ Commit block
     ├─ Append to blockchain
     ├─ Record state snapshot
     └─ Distribute block rewards
         ├─ node_1 (proposer): +1M tokens
         └─ Participating nodes: +variable reward
            ↓
            Record as on-chain state:
            reward:node_1 → 500k
            reward:node_2 → 250k
            etc.
```

## Transaction Flow

```
┌─────────────────────────────────────────────────┐
│             TRANSACTION CREATION                │
│  (In FL node, validator, or smart contract)     │
├─────────────────────────────────────────────────┤
│                                                 │
│  Transaction {                                  │
│    • ID: unique identifier                      │
│    • Type: fl_round|stake|reward|contract       │
│    • From: sender address                       │
│    • To: recipient (optional)                   │
│    • Nonce: sequential per address              │
│    • Amount: tokens transferred                 │
│    • Data: type-specific payload                │
│    • Signature: sender's signature              │
│  }                                              │
│                                                 │
└─────────────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────────────┐
│            MEMPOOL (Transaction Pool)           │
│     (Pending txns waiting for block)            │
├─────────────────────────────────────────────────┤
│                                                 │
│  Validation:                                    │
│  ✓ Signature valid?                            │
│  ✓ Nonce sequential?                           │
│  ✓ Amount sufficient?                          │
│                                                 │
│  If valid:                                      │
│  ├─ Add to mempool                             │
│  ├─ Gossip via P2P                             │
│  └─ Broadcast to peers                         │
│                                                 │
│  Mempool state:                                 │
│  [FlTxn1, StakeTxn2, RewardTxn3, ...]          │
│  Sorted by: gas_price DESC, timestamp ASC      │
│                                                 │
└─────────────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────────────┐
│         BLOCK CREATION (By Validator)           │
├─────────────────────────────────────────────────┤
│                                                 │
│ Selected Validator:                             │
│ 1. Take top 1000 txns from mempool              │
│ 2. Add to block.Transactions                    │
│ 3. For each txn:                                │
│    ├─ Execute transaction                      │
│    └─ Update state                             │
│ 4. Compute MerkleRoot(transactions)             │
│ 5. Compute StateRoot(resulting state)           │
│ 6. Build block header                          │
│ 7. Hash block                                   │
│ 8. Sign block                                   │
│                                                 │
│ Result: Fully formed block ready to broadcast   │
│                                                 │
└─────────────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────────────┐
│         BLOCK VALIDATION (By All Nodes)         │
├─────────────────────────────────────────────────┤
│                                                 │
│ 5-Layer Validation:                             │
│                                                 │
│ Layer 1: Header Validation                      │
│ ├─ Index sequential?                           │
│ ├─ Hash matches content?                       │
│ ├─ Previous hash matches tip?                  │
│ ├─ Validator exists?                           │
│ └─ Timestamp reasonable?                       │
│                                                 │
│ Layer 2: Transaction Validation                │
│ ├─ Each txn signature valid?                   │
│ ├─ Nonces sequential per sender?               │
│ └─ Type-specific checks                        │
│                                                 │
│ Layer 3: Merkle Root Verification              │
│ ├─ Compute merkle tree of txns                 │
│ └─ Does it match block.MerkleRoot?             │
│                                                 │
│ Layer 4: State Root Verification               │
│ ├─ Execute all txns in order                   │
│ ├─ Apply state changes                         │
│ └─ Does resulting root match block.StateRoot?  │
│                                                 │
│ Layer 5: Proof Validation                      │
│ ├─ Verify ZK-SNARK proofs                      │
│ ├─ Verify BFT consensus proofs                 │
│ └─ Verify TPM attestations                     │
│                                                 │
│ ✅ If all pass: Accept block                   │
│ ❌ If any fail: Reject block                   │
│                                                 │
└─────────────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────────────┐
│      BLOCK COMMITMENT & STATE UPDATE            │
├─────────────────────────────────────────────────┤
│                                                 │
│ 1. Append block to blockchain                  │
│    └─ blockchain.Blocks = [..., newBlock]      │
│                                                 │
│ 2. Remove committed txns from mempool          │
│    └─ For each txn in block: mempool.Remove()  │
│                                                 │
│ 3. Record state snapshot                       │
│    └─ StateDB.RecordSnapshot(blockHeight)      │
│                                                 │
│ 4. Distribute block rewards                    │
│    └─ validator reward 1M tokens               │
│    └─ participating nodes get FL rewards       │
│                                                 │
│ 5. Update validator metrics                    │
│    ├─ validator.AccumulatedVotes++             │
│    ├─ validator.LastActiveBlock = height       │
│    └─ validator.AccumulatedRewards += reward   │
│                                                 │
└─────────────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────────────┐
│           FINALITY & IMMUTABILITY               │
├─────────────────────────────────────────────────┤
│                                                 │
│ Block becomes "final" when:                     │
│ • Confirmed by 2/3 validators (BFT)            │
│ • State snapshot recorded                      │
│ • Cryptographically linked to chain            │
│                                                 │
│ Immutable because:                              │
│ • Each block contains hash of previous         │
│ • Changing any tx requires recomputing         │
│   all subsequent blocks' hashes                │
│ • Would need 2/3 validator consensus again     │
│                                                 │
└─────────────────────────────────────────────────┘
```

## Component Interaction Diagram

```
┌──────────────────────────────────────────────────────────────────┐
│                         SOVEREIGN MAP BLOCKCHAIN                  │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐           ┌──────────────┐                     │
│  │ Consensus   │──────────▶│ BlockProposer│────┐                │
│  │ (BFT)       │           │              │    │                │
│  └─────────────┘           └──────────────┘    │                │
│                                  ▲              │                │
│  ┌─────────────┐                 │              │                │
│  │  FL Nodes   │──────┐          │              │                │
│  │             │      │          │              │                │
│  └─────────────┘      │          │              │                │
│        ▲              │          │              ▼                │
│        │              ▼          │        ┌──────────────┐       │
│  ┌─────────────┐  ┌─────────────┐        │  BlockChain  │       │
│  │Mempool      │  │Transaction  │───────▶│              │       │
│  │             │  │(gas price   │        │ [Genesis]    │       │
│  │[ordered by  │  │ priortized) │        │ [Block 1]    │       │
│  │ gas price]  │  │             │        │ [Block 2]    │       │
│  └─────────────┘  └─────────────┘        │ ...          │       │
│        ▲                                  └──────────────┘       │
│        │                                         ▲               │
│        │                                         │               │
│  ┌──────────────┐                     ┌──────────────────────┐  │
│  │ P2P Network  │◀────────────────────▶│ BlockValidator       │  │
│  │ (gossip)     │                     │(5-layer validation)  │  │
│  └──────────────┘                     └──────────────────────┘  │
│        ▲                                         ▲               │
│        │                                         │               │
│        └─────────────┬──────────────────────────┘               │
│                      │ State updates                             │
│                      ▼                                            │
│          ┌─────────────────────┐                                │
│          │  StateDatabase      │                                │
│          │                     │                                │
│          │ Key-Value Store:    │                                │
│          │ • stake:node_1      │                                │
│          │ • reward:node_1     │                                │
│          │ • fl_round:1234     │                                │
│          │ • StateRoot: Merkle │                                │
│          │ • Snapshots: [...]  │                                │
│          └─────────────────────┘                                │
│                      ▲                                            │
│                      │                                            │
│          ┌────────────────────────┐                              │
│          │ValidatorSet            │                              │
│          │                        │                              │
│          │ • node_1: 100M stake   │                              │
│          │ • node_2: 50M stake    │                              │
│          │ • node_3: 25M stake    │                              │
│          │ • Total: 175M stake    │                              │
│          │ • Epoch: 5             │                              │
│          │ • Rewards: [...]       │                              │
│          └────────────────────────┘                              │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

## Smart Contract VM (Future)

```
┌──────────────────────────────────────────┐
│     Smart Contract Transaction           │
│  (type: smart_contract)                  │
├──────────────────────────────────────────┤
│ • ContractAddress: 0xabc...              │
│ • Function: "stake"                      │
│ • Parameters: [1000000]                  │
│ • Gas: 100000                            │
│ • GasPrice: 1                            │
└──────────────────────────────────────────┘
           ↓
┌──────────────────────────────────────────┐
│   Smart Contract VM (🔜 Phase 4)        │
│                                          │
│  Bytecode Interpreter / WASM Runtime     │
│                                          │
│  execute(contract, function, params)     │
│    ├─ Load contract bytecode             │
│    ├─ Parse function call                │
│    ├─ Check gas limit                    │
│    ├─ Execute function                   │
│    └─ Update state                       │
└──────────────────────────────────────────┘
           ↓
┌──────────────────────────────────────────┐
│      Contract State Updates               │
│                                          │
│  Example (TokenRewards):                 │
│  nodeStakes[msg.sender] += 1000000       │
│  accumulatedRewards[node_1] += 50000     │
│                                          │
│  Recorded in StateDB:                    │
│  "contract:stake:node_1" → 1000000       │
│  "contract:reward:node_1" → 50000        │
└──────────────────────────────────────────┘
```

## Performance Architecture

```
Transaction Creation
    ↓
    ├─ Signature validation: <1ms
    └─ Nonce check: <1ms

Mempool Operations (10,000 txns)
    ├─ Add transaction: <1ms
    ├─ Sort by gas: <50ms
    └─ Select for block: <10ms

Block Creation (1000 txns)
    ├─ Collect from mempool: <10ms
    ├─ Compute Merkle root: <20ms
    ├─ Execute transactions: <30ms
    ├─ Compute state root: <50ms
    └─ Hash block: <5ms
    Total: ~115ms per block

Block Validation
    ├─ Header validation: <5ms
    ├─ Transaction validation: <10ms
    ├─ Merkle verification: <20ms
    ├─ State execution: <30ms
    ├─ Proof verification: <30ms
    └─ Total per block: ~95ms

Block Commitment
    ├─ Append to chain: <1ms
    ├─ Update mempool: <10ms
    ├─ Record snapshot: <5ms
    ├─ Update validators: <5ms
    └─ Total: ~21ms

End-to-End: 1-2 blocks per second @ 1000 tx/block
            = 1000 transactions per second
```

---

## Next Steps

```
┌─────────────────────────────────────────┐
│  Current State: Foundation Ready        │
│  Status: ✅ Complete & Tested          │
└─────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────┐
│  Phase 1: Consensus Integration         │
│  Effort: 1-2 days                       │
│  Status: 🔜 Ready to Start              │
└─────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────┐
│  Phase 2: FL Node Integration           │
│  Effort: 3-5 days                       │
│  Status: 🔜 Ready to Start              │
└─────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────┐
│  Phase 3: Smart Contracts               │
│  Effort: 2-3 weeks                      │
│  Status: 🔜 Framework Ready             │
└─────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────┐
│  Phase 4: Production Hardening          │
│  Effort: 2-4 weeks                      │
│  Status: 🔜 Ready for Planning          │
└─────────────────────────────────────────┘
           ↓
┌─────────────────────────────────────────┐
│  MAINNET: Full Blockchain Ready         │
│  Estimated: 10-17 weeks total           │
└─────────────────────────────────────────┘
```

---

**See also:**
- [BLOCKCHAIN_QUICK_START.md](BLOCKCHAIN_QUICK_START.md) - Integration guide
- [BLOCKCHAIN_TRANSFORMATION.md](BLOCKCHAIN_TRANSFORMATION.md) - Strategic roadmap
- [internal/blockchain/README.md](internal/blockchain/README.md) - Technical reference
