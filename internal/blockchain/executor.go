// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package blockchain

import (
	"fmt"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/blockchain/vm"
)

// SmartContractExecutor handles execution of smart contract transactions
type SmartContractExecutor struct {
	contractVM *vm.SmartContractVM
	stateDB    *StateDatabase
	validators *ValidatorSet
}

// NewSmartContractExecutor creates a new executor
func NewSmartContractExecutor(contractVM *vm.SmartContractVM, stateDB *StateDatabase) *SmartContractExecutor {
	return &SmartContractExecutor{
		contractVM: contractVM,
		stateDB:    stateDB,
		validators: nil,
	}
}

// NewSmartContractExecutorWithValidators creates an executor that can apply governance actions.
func NewSmartContractExecutorWithValidators(contractVM *vm.SmartContractVM, stateDB *StateDatabase, validators *ValidatorSet) *SmartContractExecutor {
	return &SmartContractExecutor{
		contractVM: contractVM,
		stateDB:    stateDB,
		validators: validators,
	}
}

// ExecuteContractTransaction executes a smart contract transaction
func (sce *SmartContractExecutor) ExecuteContractTransaction(txn *Transaction) error {
	if txn.Type != TxTypeSmartContract {
		return fmt.Errorf("not a smart contract transaction")
	}

	data := txn.Data
	if data == nil {
		return fmt.Errorf("contract transaction missing data")
	}

	// Determine operation: deploy or call
	if _, isDeploy := data["deploy"]; isDeploy {
		return sce.executeDeployment(txn)
	}

	if _, isCall := data["call"]; isCall {
		return sce.executeCall(txn)
	}

	return fmt.Errorf("contract transaction must specify 'deploy' or 'call'")
}

// executeDeployment deploys a new smart contract
func (sce *SmartContractExecutor) executeDeployment(txn *Transaction) error {
	data := txn.Data

	// Extract deployment parameters
	contractName, ok := data["contract_name"].(string)
	if !ok {
		return fmt.Errorf("missing contract_name in deployment")
	}

	code, ok := data["code"].([]byte)
	if !ok {
		// Try string format
		codeStr, ok := data["code"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid code in deployment")
		}
		code = []byte(codeStr)
	}

	// Create execution context
	ctx := &vm.ExecutionContext{
		CallerAddress: txn.From,
		TransactionID: txn.ID,
		Timestamp:     txn.Timestamp,
		GasLimit:      txn.Gas,
		GasUsed:       0,
		CallData:      make(map[string]interface{}),
		StateDB:       vm.ContractState{Data: make(map[string]interface{})},
		Events:        make([]vm.ContractEvent, 0),
	}

	// Get or create ABI
	var contractABI vm.ContractABI
	if abiData, ok := data["abi"]; ok {
		// Parse ABI from data
		if abiMap, ok := abiData.(map[string]interface{}); ok {
			contractABI = parseABIFromMap(abiMap)
		}
	} else {
		// Use default ABI based on contract name
		contractABI = getDefaultABI(contractName)
	}

	// Deploy contract
	addr, err := sce.contractVM.DeployContract(ctx, contractName, code, contractABI)
	if err != nil {
		return fmt.Errorf("contract deployment failed: %w", err)
	}

	// Store deployment in blockchain state
	deploymentKey := fmt.Sprintf("contract_deployment:%s", addr)
	deploymentInfo := map[string]interface{}{
		"contract_address": addr,
		"contract_name":    contractName,
		"creator":          txn.From,
		"transaction_id":   txn.ID,
		"timestamp":        txn.Timestamp,
		"gas_used":         ctx.GasUsed,
	}

	return sce.stateDB.Set(deploymentKey, deploymentInfo)
}

// executeCall calls a contract function
func (sce *SmartContractExecutor) executeCall(txn *Transaction) error {
	data := txn.Data

	// Extract call parameters
	contractAddr, ok := data["contract_address"].(string)
	if !ok {
		return fmt.Errorf("missing contract_address in call")
	}

	functionName, ok := data["function"].(string)
	if !ok {
		return fmt.Errorf("missing function name in call")
	}

	// Create execution context
	ctx := &vm.ExecutionContext{
		CallerAddress:   txn.From,
		ContractAddress: contractAddr,
		TransactionID:   txn.ID,
		Timestamp:       txn.Timestamp,
		GasLimit:        txn.Gas,
		GasUsed:         0,
		CallData:        extractCallData(data),
		StateDB:         vm.ContractState{Data: make(map[string]interface{})},
		Events:          make([]vm.ContractEvent, 0),
	}

	// Execute contract function
	result, err := sce.contractVM.CallContract(ctx, contractAddr, functionName)
	if err != nil {
		return fmt.Errorf("contract call failed: %w", err)
	}

	if err := sce.applyGovernanceAction(contractAddr, functionName, result); err != nil {
		return fmt.Errorf("governance apply failed: %w", err)
	}

	// Store call result in blockchain state
	callKey := fmt.Sprintf("contract_call:%s:%s", contractAddr, txn.ID)
	callInfo := map[string]interface{}{
		"contract_address": contractAddr,
		"function":         functionName,
		"caller":           txn.From,
		"result":           result,
		"timestamp":        txn.Timestamp,
		"gas_used":         ctx.GasUsed,
		"events":           ctx.Events,
	}

	return sce.stateDB.Set(callKey, callInfo)
}

func (sce *SmartContractExecutor) applyGovernanceAction(contractAddr string, functionName string, result map[string]interface{}) error {
	if functionName != "executeProposal" {
		return nil
	}

	contract, err := sce.contractVM.GetContract(contractAddr)
	if err != nil {
		return err
	}
	if contract.Name != "Governance" {
		return nil
	}

	action, _ := result["action"].(string)
	switch action {
	case "set_reputation_policy":
		if sce.validators == nil {
			return nil
		}
		return sce.applyReputationPolicy(result, contractAddr)
	case "set_verification_policy":
		return sce.applyVerificationPolicyToState(result, contractAddr)
	default:
		return nil
	}
}

func (sce *SmartContractExecutor) applyReputationPolicy(result map[string]interface{}, contractAddr string) error {
	params, ok := result["params"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing governance params")
	}

	current := sce.validators.GetReputationPolicy()
	updated := current

	if v, ok := asUint32(params["slash_penalty"]); ok {
		updated.SlashPenalty = v
	}
	if v, ok := asUint32(params["reward_gain"]); ok {
		updated.RewardGain = v
	}
	if v, ok := asUint32(params["epoch_recovery"]); ok {
		updated.EpochRecovery = v
	}
	if v, ok := asUint32(params["reputation_weight"]); ok {
		updated.ReputationWeight = v
	}
	if v, ok := asUint32(params["attestation_weight"]); ok {
		updated.AttestationWeight = v
	}
	if v, ok := asUint32(params["quality_weight"]); ok {
		updated.QualityWeight = v
	}
	if v, ok := asUint64(params["max_attestation_age_blocks"]); ok {
		updated.MaxAttestationAgeBlocks = v
	}
	if v, ok := asUint32(params["invalid_attestation_penalty"]); ok {
		updated.InvalidAttestationPenalty = v
	}
	if v, ok := asUint32(params["invalid_attestation_slash_bps"]); ok {
		updated.InvalidAttestationSlashBps = v
	}
	if v, ok := asUint32(params["stale_attestation_slash_bps"]); ok {
		updated.StaleAttestationSlashBps = v
	}
	if v, ok := asUint32(params["max_consecutive_attestation_failures"]); ok {
		updated.MaxConsecutiveAttestationFailures = v
	}
	if v, ok := asUint32(params["min_quality_score"]); ok {
		updated.MinQualityScore = v
	}
	if v, ok := asUint32(params["quality_penalty"]); ok {
		updated.QualityPenalty = v
	}
	if v, ok := asUint32(params["max_consecutive_low_quality"]); ok {
		updated.MaxConsecutiveLowQuality = v
	}

	if err := sce.validators.SetReputationPolicy(updated); err != nil {
		return err
	}

	proposalID, _ := result["proposal_id"].(string)
	if proposalID == "" {
		proposalID = "unknown"
	}
	ts := time.Now().Unix()
	auditKey := fmt.Sprintf("governance_policy_audit:%s:%d", proposalID, ts)
	auditValue := map[string]interface{}{
		"proposal_id": proposalID,
		"action":      "set_reputation_policy",
		"timestamp":   ts,
		"old_policy":  current,
		"new_policy":  updated,
	}
	return sce.stateDB.Set(auditKey, auditValue)
}

// applyVerificationPolicyToState writes the governance-approved verification policy into
// the state database. BlockProposer.RefreshVerificationPolicyFromState() reads this key
// at the start of each block proposal to sync the active in-memory policy.
func (sce *SmartContractExecutor) applyVerificationPolicyToState(result map[string]interface{}, contractAddr string) error {
	params, ok := result["params"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing governance params for set_verification_policy")
	}

	// Read current in-state policy as baseline (may be zero-value on first call).
	baseline := map[string]interface{}{
		"require_proof":                  false,
		"min_confidence_bps":             float64(6000),
		"reject_on_verification_failure": false,
		"allow_consensus_proof":          true,
		"allow_zk_proof":                 true,
		"allow_tee_proof":                true,
	}
	if stored, err := sce.stateDB.Get("verification_policy:active"); err == nil {
		if m, ok := stored.(map[string]interface{}); ok {
			baseline = m
		}
	}

	// Merge incoming governance params over the baseline.
	updated := make(map[string]interface{}, len(baseline))
	for k, v := range baseline {
		updated[k] = v
	}
	for _, key := range []string{"require_proof", "reject_on_verification_failure", "allow_consensus_proof", "allow_zk_proof", "allow_tee_proof"} {
		if v, exists := params[key]; exists {
			updated[key] = v
		}
	}
	if v, ok := asUint32(params["min_confidence_bps"]); ok {
		updated["min_confidence_bps"] = float64(v)
	}

	if err := sce.stateDB.Set("verification_policy:active", updated); err != nil {
		return err
	}

	proposalID, _ := result["proposal_id"].(string)
	if proposalID == "" {
		proposalID = "unknown"
	}
	ts := time.Now().Unix()
	auditKey := fmt.Sprintf("governance_verification_audit:%s:%d", proposalID, ts)
	return sce.stateDB.Set(auditKey, map[string]interface{}{
		"proposal_id": proposalID,
		"action":      "set_verification_policy",
		"timestamp":   ts,
		"new_policy":  updated,
	})
}

func asUint32(v interface{}) (uint32, bool) {
	switch n := v.(type) {
	case uint32:
		return n, true
	case uint64:
		return uint32(n), true
	case int:
		if n < 0 {
			return 0, false
		}
		return uint32(n), true
	case float64:
		if n < 0 {
			return 0, false
		}
		return uint32(n), true
	default:
		return 0, false
	}
}

func asUint64(v interface{}) (uint64, bool) {
	switch n := v.(type) {
	case uint32:
		return uint64(n), true
	case uint64:
		return n, true
	case int:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	case float64:
		if n < 0 {
			return 0, false
		}
		return uint64(n), true
	default:
		return 0, false
	}
}

// extractCallData extracts function parameters from transaction data
func extractCallData(data map[string]interface{}) map[string]interface{} {
	callData := make(map[string]interface{})

	// Extract parameters (typically under "params" or directly in data)
	if params, ok := data["params"]; ok {
		if paramMap, ok := params.(map[string]interface{}); ok {
			return paramMap
		}
	}

	// Otherwise, copy all data except reserved keys
	reserved := map[string]bool{
		"call":             true,
		"contract_address": true,
		"function":         true,
		"deploy":           true,
		"contract_name":    true,
		"code":             true,
		"abi":              true,
	}

	for k, v := range data {
		if !reserved[k] {
			callData[k] = v
		}
	}

	return callData
}

// parseABIFromMap converts a map to ContractABI
func parseABIFromMap(abiMap map[string]interface{}) vm.ContractABI {
	// Simplified ABI parsing
	abi := vm.ContractABI{
		Functions: make(map[string]vm.FunctionDef),
		Events:    make(map[string]vm.EventDef),
	}

	if funcs, ok := abiMap["functions"].(map[string]interface{}); ok {
		for name, funcData := range funcs {
			if funcMap, ok := funcData.(map[string]interface{}); ok {
				funcDef := vm.FunctionDef{Name: name}
				if inputs, ok := funcMap["inputs"].(map[string]interface{}); ok {
					funcDef.Inputs = toStringMap(inputs)
				}
				if outputs, ok := funcMap["outputs"].(map[string]interface{}); ok {
					funcDef.Outputs = toStringMap(outputs)
				}
				abi.Functions[name] = funcDef
			}
		}
	}

	return abi
}

// toStringMap converts interface{} map to string map
func toStringMap(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		if s, ok := v.(string); ok {
			result[k] = s
		}
	}
	return result
}

// getDefaultABI returns the default ABI for known contract types
func getDefaultABI(contractName string) vm.ContractABI {
	switch contractName {
	case "TokenRewards":
		contract := vm.NewTokenRewardsContract("")
		return contract.ABI
	case "ModelRegistry":
		contract := vm.NewModelRegistryContract("")
		return contract.ABI
	case "Governance":
		contract := vm.NewGovernanceContract("")
		return contract.ABI
	default:
		return vm.ContractABI{
			Functions: make(map[string]vm.FunctionDef),
			Events:    make(map[string]vm.EventDef),
		}
	}
}

// GetContractState retrieves the state of a deployed contract
func (sce *SmartContractExecutor) GetContractState(contractAddr string) (map[string]interface{}, error) {
	return sce.contractVM.ExportState(contractAddr)
}

// GetContractInfo retrieves information about a deployed contract
func (sce *SmartContractExecutor) GetContractInfo(contractAddr string) (map[string]interface{}, error) {
	contract, err := sce.contractVM.GetContract(contractAddr)
	if err != nil {
		return nil, err
	}

	state, err := sce.contractVM.ExportState(contractAddr)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"address":    contract.Address,
		"name":       contract.Name,
		"version":    contract.Version,
		"creator":    contract.CreatorAddr,
		"created_at": contract.CreatedAt,
		"abi":        contract.ABI,
		"state":      state,
	}, nil
}
