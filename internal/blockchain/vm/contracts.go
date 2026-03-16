// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package vm

import (
	"encoding/json"
)

// NewTokenRewardsContract creates a new TokenRewards contract
func NewTokenRewardsContract(creatorAddr string) *Contract {
	return &Contract{
		Name:        "TokenRewards",
		Version:     "1.0.0",
		CreatorAddr: creatorAddr,
		ABI: ContractABI{
			Functions: map[string]FunctionDef{
				"stake": {
					Name:     "stake",
					Inputs:   map[string]string{"amount": "uint256"},
					Payable:  true,
					GasUsage: 50000,
				},
				"unstake": {
					Name:     "unstake",
					Inputs:   map[string]string{"amount": "uint256"},
					GasUsage: 60000,
				},
				"distributeRewards": {
					Name:     "distributeRewards",
					Inputs:   map[string]string{"nodes": "address[]", "rewards": "uint256[]"},
					GasUsage: 200000,
				},
				"getStake": {
					Name:     "getStake",
					Inputs:   map[string]string{"address": "address"},
					Outputs:  map[string]string{"stake": "uint256"},
					View:     true,
					GasUsage: 5000,
				},
				"getRewards": {
					Name:     "getRewards",
					Inputs:   map[string]string{"address": "address"},
					Outputs:  map[string]string{"rewards": "uint256"},
					View:     true,
					GasUsage: 5000,
				},
			},
			Events: map[string]EventDef{
				"Staked": {
					Name: "Staked",
					Parameters: map[string]string{
						"staker": "indexed address",
						"amount": "uint256",
					},
				},
				"Unstaked": {
					Name: "Unstaked",
					Parameters: map[string]string{
						"staker": "indexed address",
						"amount": "uint256",
					},
				},
				"RewardsDistributed": {
					Name: "RewardsDistributed",
					Parameters: map[string]string{
						"node_count":    "uint256",
						"total_rewards": "uint256",
					},
				},
			},
		},
		State: &ContractState{
			Data: make(map[string]interface{}),
		},
	}
}

// NewModelRegistryContract creates a new ModelRegistry contract
func NewModelRegistryContract(creatorAddr string) *Contract {
	return &Contract{
		Name:        "ModelRegistry",
		Version:     "1.0.0",
		CreatorAddr: creatorAddr,
		ABI: ContractABI{
			Functions: map[string]FunctionDef{
				"registerModel": {
					Name: "registerModel",
					Inputs: map[string]string{
						"ipfs_cid": "string",
						"accuracy": "uint256",
					},
					GasUsage: 100000,
				},
				"getModel": {
					Name: "getModel",
					Inputs: map[string]string{
						"ipfs_cid": "string",
					},
					Outputs: map[string]string{
						"model": "tuple",
					},
					View:     true,
					GasUsage: 10000,
				},
			},
			Events: map[string]EventDef{
				"ModelRegistered": {
					Name: "ModelRegistered",
					Parameters: map[string]string{
						"ipfs_cid":  "string",
						"accuracy":  "uint256",
						"submitter": "indexed address",
					},
				},
			},
		},
		State: &ContractState{
			Data: make(map[string]interface{}),
		},
	}
}

// NewGovernanceContract creates a new Governance contract
func NewGovernanceContract(creatorAddr string) *Contract {
	return &Contract{
		Name:        "Governance",
		Version:     "1.0.0",
		CreatorAddr: creatorAddr,
		ABI: ContractABI{
			Functions: map[string]FunctionDef{
				"createProposal": {
					Name: "createProposal",
					Inputs: map[string]string{
						"title":       "string",
						"description": "string",
						"action":      "string",
						"params":      "map",
					},
					Outputs: map[string]string{
						"proposal_id": "string",
					},
					GasUsage: 150000,
				},
				"vote": {
					Name: "vote",
					Inputs: map[string]string{
						"proposal_id": "string",
						"support":     "bool",
					},
					GasUsage: 80000,
				},
				"getProposal": {
					Name: "getProposal",
					Inputs: map[string]string{
						"proposal_id": "string",
					},
					Outputs: map[string]string{
						"proposal": "tuple",
					},
					View:     true,
					GasUsage: 10000,
				},
				"executeProposal": {
					Name: "executeProposal",
					Inputs: map[string]string{
						"proposal_id": "string",
					},
					Outputs: map[string]string{
						"success": "bool",
						"action":  "string",
					},
					GasUsage: 120000,
				},
			},
			Events: map[string]EventDef{
				"ProposalCreated": {
					Name: "ProposalCreated",
					Parameters: map[string]string{
						"proposal_id": "string",
						"title":       "string",
						"creator":     "indexed address",
					},
				},
				"VoteCast": {
					Name: "VoteCast",
					Parameters: map[string]string{
						"proposal_id": "string",
						"voter":       "indexed address",
						"support":     "bool",
					},
				},
				"ProposalExecuted": {
					Name: "ProposalExecuted",
					Parameters: map[string]string{
						"proposal_id": "string",
						"action":      "string",
						"executor":    "indexed address",
					},
				},
			},
		},
		State: &ContractState{
			Data: make(map[string]interface{}),
		},
	}
}

// ContractJSON returns JSON representation of a contract
func (c *Contract) JSON() ([]byte, error) {
	type ContractJSON struct {
		Address     string                 `json:"address"`
		Name        string                 `json:"name"`
		Version     string                 `json:"version"`
		CreatorAddr string                 `json:"creator_address"`
		CreatedAt   int64                  `json:"created_at"`
		ABI         ContractABI            `json:"abi"`
		State       map[string]interface{} `json:"state"`
	}

	c.State.mu.RLock()
	stateCopy := make(map[string]interface{})
	for k, v := range c.State.Data {
		stateCopy[k] = v
	}
	c.State.mu.RUnlock()

	cj := ContractJSON{
		Address:     c.Address,
		Name:        c.Name,
		Version:     c.Version,
		CreatorAddr: c.CreatorAddr,
		CreatedAt:   c.CreatedAt,
		ABI:         c.ABI,
		State:       stateCopy,
	}

	return json.MarshalIndent(cj, "", "  ")
}
