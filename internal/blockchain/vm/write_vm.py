content = r'''// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package vm

import (
	"fmt"
	"sync"
	"time"
)

// ContractState holds the mutable state of a deployed contract.
type ContractState struct {
	mu   sync.RWMutex
	Data map[string]interface{}
}

// Get retrieves a value from contract state.
func (cs *ContractState) Get(key string) (interface{}, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	val, exists := cs.Data[key]
	return val, exists
}

// Set stores a value in contract state.
func (cs *ContractState) Set(key string, value interface{}) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.Data[key] = value
}

// IncrementCounter atomically increments a uint64 counter by amount.
func (cs *ContractState) IncrementCounter(key string, amount uint64) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	current := uint64(0)
	if val, exists := cs.Data[key]; exists {
		if num, ok := val.(uint64); ok {
			current = num
		} else {
			return fmt.Errorf("value at %s is not uint64", key)
		}
	}
	cs.Data[key] = current + amount
	return nil
}

// ExecutionContext contains information about a contract execution.
type ExecutionContext struct {
	CallerAddress   string
	ContractAddress string
	TransactionID   string
	Timestamp       int64
	GasLimit        uint64
	GasUsed         uint64
	CallData        map[string]interface{}
	StateDB         ContractState
	Events          []ContractEvent
}

// ContractEvent represents an event emitted during execution.
type ContractEvent struct {
	ContractAddress string                 `json:"contract_address"`
	EventName       string                 `json:"event_name"`
	Parameters      map[string]interface{} `json:"parameters"`
	Timestamp       int64                  `json:"timestamp"`
}

// FunctionDef describes a single contract function in the ABI.
type FunctionDef struct {
	Name     string            `json:"name"`
	Inputs   map[string]string `json:"inputs"`
	Outputs  map[string]string `json:"outputs"`
	Payable  bool              `json:"payable"`
	View     bool              `json:"view"`
	GasUsage uint64            `json:"gas_usage"`
}

// EventDef describes a contract event in the ABI.
type EventDef struct {
	Name       string            `json:"name"`
	Parameters map[string]string `json:"parameters"`
}

// ErrorDef describes a contract error/revert.
type ErrorDef struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Code    uint32 `json:"code"`
}

// ContractABI describes the contract interface.
type ContractABI struct {
	Functions map[string]FunctionDef `json:"functions"`
	Events    map[string]EventDef    `json:"events"`
	Errors    map[string]ErrorDef    `json:"errors"`
}

// Contract represents a deployed smart contract.
type Contract struct {
	Address     string         `json:"address"`
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	CreatorAddr string         `json:"creator_address"`
	CreatedAt   int64          `json:"created_at"`
	Code        []byte         `json:"code"`
	ABI         ContractABI    `json:"abi"`
	State       *ContractState `json:"-"`
}

// ContractInterface defines how contracts interact with the blockchain.
type ContractInterface interface {
	GetAddress() string
	GetName() string
	GetABI() ContractABI
	Execute(context *ExecutionContext, function string) (map[string]interface{}, error)
}

// Verify Contract implements ContractInterface.
var _ ContractInterface = (*Contract)(nil)

func (c *Contract) GetAddress() string  { return c.Address }
func (c *Contract) GetName() string     { return c.Name }
func (c *Contract) GetABI() ContractABI { return c.ABI }

// Execute is not implemented for the generic Contract type.
func (c *Contract) Execute(context *ExecutionContext, function string) (map[string]interface{}, error) {
	return nil, fmt.Errorf("execute not implemented for Contract type")
}

// SmartContractVM executes smart contract code.
type SmartContractVM struct {
	mu           sync.RWMutex
	contracts    map[string]*Contract
	maxCodeSize  int
	maxGasPerTxn uint64
}

// NewSmartContractVM creates a new VM instance.
func NewSmartContractVM() *SmartContractVM {
	return &SmartContractVM{
		contracts:    make(map[string]*Contract),
		maxCodeSize:  1000000,  // 1 MB
		maxGasPerTxn: 10000000, // 10 M gas
	}
}

// DeployContract deploys a new smart contract and returns its address.
func (vm *SmartContractVM) DeployContract(ctx *ExecutionContext, contractName string, code []byte, abi ContractABI) (string, error) {
	if len(code) > vm.maxCodeSize {
		return "", fmt.Errorf("contract code exceeds max size: %d > %d", len(code), vm.maxCodeSize)
	}
	addr := generateContractAddress(ctx.CallerAddress, ctx.TransactionID)
	vm.mu.Lock()
	defer vm.mu.Unlock()
	if _, exists := vm.contracts[addr]; exists {
		return "", fmt.Errorf("contract already exists at address: %s", addr)
	}
	contract := &Contract{
		Address:     addr,
		Name:        contractName,
		Version:     "1.0.0",
		CreatorAddr: ctx.CallerAddress,
		CreatedAt:   time.Now().Unix(),
		Code:        code,
		ABI:         abi,
		State: &ContractState{
			Data: make(map[string]interface{}),
		},
	}
	vm.contracts[addr] = contract
	ctx.GasUsed += 100000 // fixed deployment gas
	return addr, nil
}

// CallContract calls a function on a deployed contract.
func (vm *SmartContractVM) CallContract(ctx *ExecutionContext, contractAddr string, functionName string) (map[string]interface{}, error) {
	vm.mu.RLock()
	contract, exists := vm.contracts[contractAddr]
	vm.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractAddr)
	}
	funcDef, funcExists := contract.ABI.Functions[functionName]
	if !funcExists {
		return nil, fmt.Errorf("function not found: %s in contract %s", functionName, contractAddr)
	}
	if ctx.GasUsed+funcDef.GasUsage > ctx.GasLimit {
		return nil, fmt.Errorf("insufficient gas: used %d + needed %d > limit %d",
			ctx.GasUsed, funcDef.GasUsage, ctx.GasLimit)
	}
	result, err := vm.executeFunction(contract, ctx, functionName, funcDef)
	if err != nil {
		return nil, err
	}
	ctx.GasUsed += funcDef.GasUsage
	return result, nil
}

// executeFunction routes execution to the appropriate built-in handler.
func (vm *SmartContractVM) executeFunction(contract *Contract, ctx *ExecutionContext, functionName string, funcDef FunctionDef) (map[string]interface{}, error) {
	switch contract.Name {
	case "TokenRewards":
		return vm.executeTokenRewardsFunction(contract, ctx, functionName)
	case "ModelRegistry":
		return vm.executeModelRegistryFunction(contract, ctx, functionName)
	case "Governance":
		return vm.executeGovernanceFunction(contract, ctx, functionName)
	default:
		return nil, fmt.Errorf("unknown contract type: %s", contract.Name)
	}
}

// executeTokenRewardsFunction handles TokenRewards contract calls.
func (vm *SmartContractVM) executeTokenRewardsFunction(contract *Contract, ctx *ExecutionContext, functionName string) (map[string]interface{}, error) {
	switch functionName {
	case "stake":
		amount := ctx.CallData["amount"].(uint64)
		key := fmt.Sprintf("stake:%s", ctx.CallerAddress)
		contract.State.IncrementCounter(key, amount)
		ctx.Events = append(ctx.Events, ContractEvent{
			ContractAddress: contract.Address,
			EventName:       "Staked",
			Parameters: map[string]interface{}{
				"staker": ctx.CallerAddress,
				"amount": amount,
			},
			Timestamp: time.Now().Unix(),
		})
		return map[string]interface{}{"success": true}, nil

	case "unstake":
		amount := ctx.CallData["amount"].(uint64)
		key := fmt.Sprintf("stake:%s", ctx.CallerAddress)
		if current, ok := contract.State.Get(key); !ok || current.(uint64) < amount {
			return nil, fmt.Errorf("insufficient stake to unstake")
		}
		contract.State.mu.Lock()
		current := contract.State.Data[key].(uint64)
		contract.State.Data[key] = current - amount
		contract.State.mu.Unlock()
		ctx.Events = append(ctx.Events, ContractEvent{
			ContractAddress: contract.Address,
			EventName:       "Unstaked",
			Parameters: map[string]interface{}{
				"staker": ctx.CallerAddress,
				"amount": amount,
			},
			Timestamp: time.Now().Unix(),
		})
		return map[string]interface{}{"success": true}, nil

	case "distributeRewards":
		nodes := ctx.CallData["nodes"].([]string)
		rewards := ctx.CallData["rewards"].([]uint64)
		if len(nodes) != len(rewards) {
			return nil, fmt.Errorf("nodes and rewards arrays have different lengths")
		}
		total := uint64(0)
		for i, nodeID := range nodes {
			key := fmt.Sprintf("reward:%s", nodeID)
			contract.State.IncrementCounter(key, rewards[i])
			total += rewards[i]
		}
		ctx.Events = append(ctx.Events, ContractEvent{
			ContractAddress: contract.Address,
			EventName:       "RewardsDistributed",
			Parameters: map[string]interface{}{
				"node_count":    len(nodes),
				"total_rewards": total,
			},
			Timestamp: time.Now().Unix(),
		})
		return map[string]interface{}{"success": true}, nil

	case "getStake":
		addr := ctx.CallData["address"].(string)
		key := fmt.Sprintf("stake:%s", addr)
		if val, ok := contract.State.Get(key); ok {
			return map[string]interface{}{"stake": val}, nil
		}
		return map[string]interface{}{"stake": uint64(0)}, nil

	case "getRewards":
		addr := ctx.CallData["address"].(string)
		key := fmt.Sprintf("reward:%s", addr)
		if val, ok := contract.State.Get(key); ok {
			return map[string]interface{}{"rewards": val}, nil
		}
		return map[string]interface{}{"rewards": uint64(0)}, nil

	default:
		return nil, fmt.Errorf("unknown function: %s", functionName)
	}
}

// executeModelRegistryFunction handles ModelRegistry contract calls.
func (vm *SmartContractVM) executeModelRegistryFunction(contract *Contract, ctx *ExecutionContext, functionName string) (map[string]interface{}, error) {
	switch functionName {
	case "registerModel":
		ipfsCID := ctx.CallData["ipfs_cid"].(string)
		accuracy := ctx.CallData["accuracy"].(float64)
		modelEntry := map[string]interface{}{
			"ipfs_cid":  ipfsCID,
			"accuracy":  accuracy,
			"submitter": ctx.CallerAddress,
			"timestamp": time.Now().Unix(),
		}
		key := fmt.Sprintf("model:%s", ipfsCID)
		contract.State.Set(key, modelEntry)
		ctx.Events = append(ctx.Events, ContractEvent{
			ContractAddress: contract.Address,
			EventName:       "ModelRegistered",
			Parameters: map[string]interface{}{
				"ipfs_cid":  ipfsCID,
				"accuracy":  accuracy,
				"submitter": ctx.CallerAddress,
			},
			Timestamp: time.Now().Unix(),
		})
		return map[string]interface{}{"success": true}, nil

	case "getModel":
		ipfsCID := ctx.CallData["ipfs_cid"].(string)
		key := fmt.Sprintf("model:%s", ipfsCID)
		if val, ok := contract.State.Get(key); ok {
			return map[string]interface{}{"model": val}, nil
		}
		return nil, fmt.Errorf("model not found: %s", ipfsCID)

	default:
		return nil, fmt.Errorf("unknown function: %s", functionName)
	}
}

// executeGovernanceFunction handles Governance contract calls.
func (vm *SmartContractVM) executeGovernanceFunction(contract *Contract, ctx *ExecutionContext, functionName string) (map[string]interface{}, error) {
	switch functionName {
	case "createProposal":
		title := ctx.CallData["title"].(string)
		description := ctx.CallData["description"].(string)
		proposalID := ctx.TransactionID
		proposal := map[string]interface{}{
			"id":            proposalID,
			"title":         title,
			"description":   description,
			"creator":       ctx.CallerAddress,
			"created_at":    time.Now().Unix(),
			"for_votes":     uint64(0),
			"against_votes": uint64(0),
			"status":        "active",
		}
		key := fmt.Sprintf("proposal:%s", proposalID)
		contract.State.Set(key, proposal)
		ctx.Events = append(ctx.Events, ContractEvent{
			ContractAddress: contract.Address,
			EventName:       "ProposalCreated",
			Parameters: map[string]interface{}{
				"proposal_id": proposalID,
				"title":       title,
				"creator":     ctx.CallerAddress,
			},
			Timestamp: time.Now().Unix(),
		})
		return map[string]interface{}{"proposal_id": proposalID}, nil

	case "vote":
		proposalID := ctx.CallData["proposal_id"].(string)
		support := ctx.CallData["support"].(bool)
		key := fmt.Sprintf("proposal:%s", proposalID)
		if val, ok := contract.State.Get(key); ok {
			proposal := val.(map[string]interface{})
			voteKey := fmt.Sprintf("vote:%s:%s", proposalID, ctx.CallerAddress)
			contract.State.Set(voteKey, support)
			if support {
				proposal["for_votes"] = proposal["for_votes"].(uint64) + 1
			} else {
				proposal["against_votes"] = proposal["against_votes"].(uint64) + 1
			}
			contract.State.Set(key, proposal)
			ctx.Events = append(ctx.Events, ContractEvent{
				ContractAddress: contract.Address,
				EventName:       "VoteCast",
				Parameters: map[string]interface{}{
					"proposal_id": proposalID,
					"voter":       ctx.CallerAddress,
					"support":     support,
				},
				Timestamp: time.Now().Unix(),
			})
			return map[string]interface{}{"success": true}, nil
		}
		return nil, fmt.Errorf("proposal not found: %s", proposalID)

	case "getProposal":
		proposalID := ctx.CallData["proposal_id"].(string)
		key := fmt.Sprintf("proposal:%s", proposalID)
		if val, ok := contract.State.Get(key); ok {
			return map[string]interface{}{"proposal": val}, nil
		}
		return nil, fmt.Errorf("proposal not found: %s", proposalID)

	default:
		return nil, fmt.Errorf("unknown function: %s", functionName)
	}
}

// GetContract retrieves a contract by address.
func (vm *SmartContractVM) GetContract(address string) (*Contract, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	contract, exists := vm.contracts[address]
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", address)
	}
	return contract, nil
}

// GetAllContracts returns all deployed contracts.
func (vm *SmartContractVM) GetAllContracts() []*Contract {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	contracts := make([]*Contract, 0, len(vm.contracts))
	for _, contract := range vm.contracts {
		contracts = append(contracts, contract)
	}
	return contracts
}

// ExportState exports a contract's state as a key-value map.
func (vm *SmartContractVM) ExportState(address string) (map[string]interface{}, error) {
	contract, err := vm.GetContract(address)
	if err != nil {
		return nil, err
	}
	contract.State.mu.RLock()
	defer contract.State.mu.RUnlock()
	stateCopy := make(map[string]interface{})
	for k, v := range contract.State.Data {
		stateCopy[k] = v
	}
	return stateCopy, nil
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// generateContractAddress creates a deterministic contract address from caller + txID.
func generateContractAddress(creatorAddr string, txnID string) string {
	data := fmt.Sprintf("contract:%s:%s", creatorAddr, txnID)
	hash := hashString(data)
	if len(hash) >= 40 {
		return fmt.Sprintf("0x%s", hash[:40])
	}
	return fmt.Sprintf("0x%040s", hash)
}

// hashString returns a hex string representing a simple hash of data.
func hashString(data string) string {
	h := 0
	for _, c := range data {
		h = h*31 + int(c)
	}
	return fmt.Sprintf("%040x", h)
}
'''

with open('/workspaces/Sovereign_Map_Federated_Learning/internal/blockchain/vm/vm.go', 'w') as f:
    f.write(content)
print('vm.go written successfully')
