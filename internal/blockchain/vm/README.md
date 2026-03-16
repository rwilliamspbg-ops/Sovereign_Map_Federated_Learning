# Smart Contract Virtual Machine (VM)

A lightweight, production-grade smart contract execution engine for Sovereign Map.

## Overview

The VM package provides:
- **Contract deployment** - Deploy new smart contracts on-chain
- **Function execution** - Call contract functions with automatic gas tracking
- **State management** - Persistent state with atomic updates
- **Event emission** - Log contract events for off-chain indexing
- **Built-in contracts** - TokenRewards, ModelRegistry, Governance

## Architecture

```
Transaction (smart_contract type)
    ↓
SmartContractExecutor (executor.go)
    ↓
SmartContractVM (vm.go)
    ├─ DeployContract()
    ├─ CallContract()
    └─ executeFunction()
    ↓
Contract State (ContractState)
    ├─ Get/Set
    ├─ Increment
    └─ Events
```

## Core Types

### Contract
```go
type Contract struct {
    Address     string         // 0xabc...def
    Name        string         // "TokenRewards"
    Version     string         // "1.0.0"
    CreatorAddr string         // Node that deployed
    CreatedAt   int64          // Unix timestamp
    Code        []byte         // Contract bytecode
    ABI         ContractABI    // Interface definition
    State       *ContractState // Persistent state
}
```

### Execution Context
```go
type ExecutionContext struct {
    CallerAddress   string                 // Who called
    ContractAddress string                 // Target contract
    TransactionID   string                 // For idempotency
    Timestamp       int64                  // Block timestamp
    GasLimit        uint64                 // Max gas allowed
    GasUsed         uint64                 // Gas consumed
    CallData        map[string]interface{} // Function params
    StateDB         ContractState          // Local state
    Events          []ContractEvent        // Emitted events
}
```

### Contract ABI
```go
type ContractABI struct {
    Functions map[string]FunctionDef  // name -> definition
    Events    map[string]EventDef     // name -> definition
    Errors    map[string]ErrorDef     // name -> definition
}

type FunctionDef struct {
    Name     string            // "stake"
    Inputs   map[string]string // {"amount": "uint256"}
    Outputs  map[string]string // {"success": "bool"}
    Payable  bool              // Accepts tokens
    View     bool              // Doesn't modify state
    GasUsage uint64            // Gas cost
}
```

## Built-in Contracts

### 1. TokenRewards
Manages node staking, rewards, and token distribution.

**Functions:**
- `stake(uint256 amount)` - Stake tokens
- `unstake(uint256 amount)` - Withdraw stake
- `distributeRewards(address[] nodes, uint256[] rewards)` - Distribute rewards
- `getStake(address)` → uint256 - View stake
- `getRewards(address)` → uint256 - View rewards

**Events:**
- `Staked(address indexed staker, uint256 amount)`
- `Unstaked(address indexed staker, uint256 amount)`
- `RewardsDistributed(uint256 node_count, uint256 total_rewards)`

**Example:**
```go
ctx := &ExecutionContext{
    CallerAddress: "node_1",
    CallData: map[string]interface{}{
        "amount": uint64(1000000),
    },
}
result, _ := vm.CallContract(ctx, contractAddr, "stake")
// Emits: Staked(node_1, 1000000)
```

### 2. ModelRegistry
Registers and tracks federated learning models.

**Functions:**
- `registerModel(string ipfs_cid, uint256 accuracy)` - Register model
- `getModel(string ipfs_cid)` → model - Get model info

**Events:**
- `ModelRegistered(string ipfs_cid, uint256 accuracy, address indexed submitter)`

### 3. Governance
Implements on-chain governance (SGPs - Sovereign Governance Proposals).

**Functions:**
- `createProposal(string title, string description)` → proposal_id - Create proposal
- `vote(string proposal_id, bool support)` - Cast vote
- `getProposal(string proposal_id)` → proposal - Get proposal

**Events:**
- `ProposalCreated(string proposal_id, string title, address indexed creator)`
- `VoteCast(string proposal_id, address indexed voter, bool support)`

## Usage Examples

### Deploy a Contract

```go
contractVM := vm.NewSmartContractVM()

// Option 1: Deploy built-in contract
ctx := &ExecutionContext{
    CallerAddress: "deployer_node",
    TransactionID: "txn_001",
    GasLimit:      1000000,
}

contract := vm.NewTokenRewardsContract(ctx.CallerAddress)
addr, _ := contractVM.DeployContract(
    ctx,
    "TokenRewards",
    nil,
    contract.ABI,
)
// addr = "0xabc123..."
```

### Call a Contract Function

```go
// Stake 1M tokens
ctx := &ExecutionContext{
    CallerAddress:   "node_1",
    ContractAddress: "0xabc123...",
    TransactionID:   "txn_002",
    GasLimit:        100000,
    CallData: map[string]interface{}{
        "amount": uint64(1000000),
    },
}

result, _ := contractVM.CallContract(ctx, addr, "stake")
// result = {"success": true}
// Events: Staked(node_1, 1000000)
```

### Access Contract State

```go
// Get stake balance
ctx := &ExecutionContext{
    CallerAddress:   "querier",
    ContractAddress: addr,
    CallData: map[string]interface{}{
        "address": "node_1",
    },
}
result, _ := contractVM.CallContract(ctx, addr, "getStake")
// result = {"stake": 1000000}
```

### Register a Model

```go
ctx := &ExecutionContext{
    CallerAddress:   "node_1",
    ContractAddress: modelRegistryAddr,
    CallData: map[string]interface{}{
        "ipfs_cid": "QmXxxx...",
        "accuracy": 0.954,
    },
}

contractVM.CallContract(ctx, modelRegistryAddr, "registerModel")
// Events: ModelRegistered("QmXxxx...", 0.954, node_1)
```

## Gas Costs

| Operation | Gas |
|-----------|-----|
| Contract Deployment | 100000 |
| Stake | 50000 |
| Unstake | 60000 |
| DistributeRewards | 200000 |
| GetStake (view) | 5000 |
| GetRewards (view) | 5000 |
| RegisterModel | 100000 |
| CreateProposal | 150000 |
| Vote | 80000 |

## Integration with BlockChain

Smart contracts are executed through blockchain transactions:

```go
txn := &Transaction{
    Type: TxTypeSmartContract,
    From: "node_1",
    Gas:  1000000,
    Data: map[string]interface{}{
        "contract_address": "0xabc123...",
        "call": true,
        "function": "stake",
        "params": map[string]interface{}{
            "amount": uint64(1000000),
        },
    },
}

// Add to mempool
blockchain.Mempool.AddTransaction(txn)

// Executor handles the transaction
executor := NewSmartContractExecutor(contractVM, blockchain.StateDB)
executor.ExecuteContractTransaction(txn)
```

## State Persistence

Contract state is stored in the blockchain's StateDatabase:

```
contract_deployment:<address> → {contract_info}
contract_call:<address>:<txn_id> → {result, events, gas_used}
```

Contract-specific state is stored in the contract's ContractState:

```
stake:node_1 → 1000000
reward:node_1 → 500000
model:QmXxxx... → {ipfs_cid, accuracy, submitter, timestamp}
proposal:<id> → {title, description, votes, status}
```

## Event Emission

All contract operations emit events for off-chain indexing:

```go
type ContractEvent struct {
    ContractAddress string
    EventName       string
    Parameters      map[string]interface{}
    Timestamp       int64
}
```

Events are:
1. Emitted during contract execution
2. Stored in the blockchain
3. Indexed by event name and contract
4. Queryable through REST API

## Future Extensions

- [ ] **WASM Runtime** - Execute WASM contracts for sandboxing
- [ ] **Solidity Compiler** - Full Solidity support
- [ ] **Contract Upgrade** - Proxy contracts for updates
- [ ] **Cross-Contract Calls** - Call contracts from contracts
- [ ] **Formal Verification** - Verify contract properties
- [ ] **Contract Library** - Reusable contract components

## Testing

```bash
# Unit tests
go test ./internal/blockchain/vm -v

# Integration tests
go test ./internal/blockchain -run TestSmartContract -v

# Benchmark
go test ./internal/blockchain/vm -bench=. -v
```

## Performance

Target performance for contract operations:

| Operation | Target | Notes |
|-----------|--------|-------|
| Deployment | <200ms | Per block |
| Call | <50ms | Per transaction |
| State read | <1ms | In-memory |
| State write | <5ms | With syncing |
| Event emission | <1ms | Per event |

## API Example

```go
// Deploy TokenRewards contract
POST /api/contracts/deploy
{
    "contract_name": "TokenRewards",
    "creator": "node_1"
}
→ { "contract_address": "0xabc123..." }

// Call function
POST /api/contracts/{address}/call
{
    "function": "stake",
    "params": { "amount": 1000000 }
}
→ { "success": true, "gas_used": 50000 }

// Get contract info
GET /api/contracts/{address}
→ { "name": "TokenRewards", "version": "1.0.0", "abi": {...}, "state": {...} }

// Query contract state
GET /api/contracts/{address}/state
→ { "stake:node_1": 1000000, "reward:node_1": 500000, ... }
```

## References

- [Ethereum Yellow Paper](https://ethereum.org/en/developers/docs/whitepapers/ethereum-original/)
- [Solidity Documentation](https://docs.soliditylang.org/)
- Parent: [BLOCKCHAIN_QUICK_START.md](../BLOCKCHAIN_QUICK_START.md)
- Parent: [BLOCKCHAIN_TRANSFORMATION.md](../BLOCKCHAIN_TRANSFORMATION.md)
