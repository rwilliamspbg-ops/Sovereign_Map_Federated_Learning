// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package blockchain

import (
	"strings"
	"testing"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/blockchain/vm"
)

func newExecutorForTest() (*SmartContractExecutor, *StateDatabase) {
	stateDB := NewStateDatabase()
	contractVM := vm.NewSmartContractVM()
	return NewSmartContractExecutor(contractVM, stateDB), stateDB
}

func TestSmartContractDeploymentTransaction(t *testing.T) {
	executor, stateDB := newExecutorForTest()

	txn := &Transaction{
		ID:        "deploy_001",
		Type:      TxTypeSmartContract,
		From:      "node_1",
		Nonce:     0,
		Gas:       1000000,
		GasPrice:  100,
		Timestamp: time.Now().Unix(),
		Signature: []byte("sig"),
		Data: map[string]interface{}{
			"deploy":        true,
			"contract_name": "TokenRewards",
			"code":          "token_rewards_bytecode",
		},
	}

	if err := executor.ExecuteContractTransaction(txn); err != nil {
		t.Fatalf("deployment should succeed: %v", err)
	}

	entries := stateDB.GetAll()
	foundDeployment := false
	for key := range entries {
		if strings.HasPrefix(key, "contract_deployment:") {
			foundDeployment = true
			break
		}
	}
	if !foundDeployment {
		t.Fatal("expected deployment record in state")
	}
}

func TestSmartContractCallTransaction(t *testing.T) {
	executor, stateDB := newExecutorForTest()

	deployTxn := &Transaction{
		ID:        "deploy_001",
		Type:      TxTypeSmartContract,
		From:      "node_1",
		Nonce:     0,
		Gas:       1000000,
		GasPrice:  100,
		Timestamp: time.Now().Unix(),
		Signature: []byte("sig"),
		Data: map[string]interface{}{
			"deploy":        true,
			"contract_name": "TokenRewards",
			"code":          "token_rewards_bytecode",
		},
	}
	if err := executor.ExecuteContractTransaction(deployTxn); err != nil {
		t.Fatalf("deployment should succeed: %v", err)
	}

	var contractAddr string
	for _, value := range stateDB.GetAll() {
		deployment, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		addr, ok := deployment["contract_address"].(string)
		if ok && addr != "" {
			contractAddr = addr
			break
		}
	}
	if contractAddr == "" {
		t.Fatal("failed to resolve deployed contract address")
	}

	callTxn := &Transaction{
		ID:        "call_001",
		Type:      TxTypeSmartContract,
		From:      "node_2",
		Nonce:     0,
		Gas:       200000,
		GasPrice:  100,
		Timestamp: time.Now().Unix(),
		Signature: []byte("sig"),
		Data: map[string]interface{}{
			"call":             true,
			"contract_address": contractAddr,
			"function":         "stake",
			"params": map[string]interface{}{
				"amount": uint64(100),
			},
		},
	}

	if err := executor.ExecuteContractTransaction(callTxn); err != nil {
		t.Fatalf("contract call should succeed: %v", err)
	}

	callKey := "contract_call:" + contractAddr + ":" + callTxn.ID
	if !stateDB.Exists(callKey) {
		t.Fatalf("expected call record at key %s", callKey)
	}
}

func TestExecuteContractTransactionRejectsMissingOperation(t *testing.T) {
	executor, _ := newExecutorForTest()

	txn := &Transaction{
		ID:        "bad_001",
		Type:      TxTypeSmartContract,
		From:      "node_1",
		Nonce:     0,
		Gas:       100000,
		GasPrice:  100,
		Timestamp: time.Now().Unix(),
		Signature: []byte("sig"),
		Data: map[string]interface{}{
			"contract_name": "TokenRewards",
		},
	}

	err := executor.ExecuteContractTransaction(txn)
	if err == nil {
		t.Fatal("expected error when neither deploy nor call is specified")
	}
	if !strings.Contains(err.Error(), "must specify 'deploy' or 'call'") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGovernanceProposalAppliesReputationPolicy(t *testing.T) {
	stateDB := NewStateDatabase()
	contractVM := vm.NewSmartContractVM()
	validators := NewValidatorSet()
	executor := NewSmartContractExecutorWithValidators(contractVM, stateDB, validators)

	deployTxn := &Transaction{
		ID:        "deploy_gov_001",
		Type:      TxTypeSmartContract,
		From:      "node_1",
		Nonce:     0,
		Gas:       1000000,
		GasPrice:  1,
		Timestamp: time.Now().Unix(),
		Signature: []byte("sig"),
		Data: map[string]interface{}{
			"deploy":        true,
			"contract_name": "Governance",
			"code":          "governance_code",
		},
	}
	if err := executor.ExecuteContractTransaction(deployTxn); err != nil {
		t.Fatalf("deploy governance contract failed: %v", err)
	}

	var contractAddr string
	for _, value := range stateDB.GetAll() {
		deployment, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		if name, _ := deployment["contract_name"].(string); name == "Governance" {
			if addr, ok := deployment["contract_address"].(string); ok {
				contractAddr = addr
				break
			}
		}
	}
	if contractAddr == "" {
		t.Fatal("expected governance contract address")
	}

	createTxn := &Transaction{
		ID:        "proposal_001",
		Type:      TxTypeSmartContract,
		From:      "node_1",
		Nonce:     0,
		Gas:       300000,
		GasPrice:  1,
		Timestamp: time.Now().Unix(),
		Signature: []byte("sig"),
		Data: map[string]interface{}{
			"call":             true,
			"contract_address": contractAddr,
			"function":         "createProposal",
			"params": map[string]interface{}{
				"title":       "Tune reputation policy",
				"description": "Adjust weighting",
				"action":      "set_reputation_policy",
				"min_votes":   uint64(1),
				"params": map[string]interface{}{
					"reputation_weight":  40,
					"attestation_weight": 40,
					"quality_weight":     20,
				},
			},
		},
	}
	if err := executor.ExecuteContractTransaction(createTxn); err != nil {
		t.Fatalf("create proposal failed: %v", err)
	}

	voteTxn := &Transaction{
		ID:        "vote_001",
		Type:      TxTypeSmartContract,
		From:      "node_2",
		Nonce:     0,
		Gas:       200000,
		GasPrice:  1,
		Timestamp: time.Now().Unix(),
		Signature: []byte("sig"),
		Data: map[string]interface{}{
			"call":             true,
			"contract_address": contractAddr,
			"function":         "vote",
			"params": map[string]interface{}{
				"proposal_id": createTxn.ID,
				"support":     true,
			},
		},
	}
	if err := executor.ExecuteContractTransaction(voteTxn); err != nil {
		t.Fatalf("vote failed: %v", err)
	}

	execTxn := &Transaction{
		ID:        "exec_001",
		Type:      TxTypeSmartContract,
		From:      "node_1",
		Nonce:     0,
		Gas:       300000,
		GasPrice:  1,
		Timestamp: time.Now().Unix(),
		Signature: []byte("sig"),
		Data: map[string]interface{}{
			"call":             true,
			"contract_address": contractAddr,
			"function":         "executeProposal",
			"params": map[string]interface{}{
				"proposal_id": createTxn.ID,
			},
		},
	}
	if err := executor.ExecuteContractTransaction(execTxn); err != nil {
		t.Fatalf("execute proposal failed: %v", err)
	}

	policy := validators.GetReputationPolicy()
	if policy.ReputationWeight != 40 || policy.AttestationWeight != 40 || policy.QualityWeight != 20 {
		t.Fatalf("governance policy not applied: %+v", policy)
	}

	foundAudit := false
	for key := range stateDB.GetAll() {
		if strings.HasPrefix(key, "governance_policy_audit:") {
			foundAudit = true
			break
		}
	}
	if !foundAudit {
		t.Fatal("expected governance policy audit entry in state")
	}
}
