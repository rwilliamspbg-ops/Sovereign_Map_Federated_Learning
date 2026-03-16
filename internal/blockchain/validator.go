// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package blockchain

import (
	"fmt"
	"time"
)

// BlockValidator validates blocks before acceptance into the chain
type BlockValidator struct {
	blockchain *BlockChain
	stateDB    *StateDatabase
}

// NewBlockValidator creates a new block validator
func NewBlockValidator(bc *BlockChain, stateDB *StateDatabase) *BlockValidator {
	return &BlockValidator{
		blockchain: bc,
		stateDB:    stateDB,
	}
}

// ValidateBlock performs comprehensive validation of a block
// Returns nil if block is valid, error otherwise
func (bv *BlockValidator) ValidateBlock(block *Block) error {
	// 1. Header validation
	if err := bv.validateHeader(block.Header); err != nil {
		return fmt.Errorf("header validation failed: %w", err)
	}

	// 2. Transaction validation
	for i, txn := range block.Transactions {
		if err := bv.validateTransaction(&txn); err != nil {
			return fmt.Errorf("transaction %d validation failed: %w", i, err)
		}
	}

	// 3. Merkle root verification
	expectedMerkle := ComputeMerkleRoot(block.Transactions)
	if block.MerkleRoot != expectedMerkle {
		return fmt.Errorf("merkle root mismatch: expected %s, got %s", expectedMerkle, block.MerkleRoot)
	}

	// 4. Block hash verification
	expectedHash := block.ComputeHash()
	if block.Header.Hash != expectedHash {
		return fmt.Errorf("block hash mismatch: expected %s, got %s", expectedHash, block.Header.Hash)
	}

	// 5. State root verification (execute transactions and verify state)
	if err := bv.validateStateRoot(block); err != nil {
		return fmt.Errorf("state root validation failed: %w", err)
	}

	// 6. Proof validation (ZK proofs, BFT proofs, etc)
	if err := bv.validateProofs(block); err != nil {
		return fmt.Errorf("proof validation failed: %w", err)
	}

	return nil
}

// validateHeader validates block header
func (bv *BlockValidator) validateHeader(header BlockHeader) error {
	// Check index is sequential
	expectedIndex := bv.blockchain.Height() + 1
	if header.Index != expectedIndex {
		return fmt.Errorf("block index mismatch: expected %d, got %d", expectedIndex, header.Index)
	}

	// Check previous hash matches chain tip
	tip := bv.blockchain.Tip
	if header.PreviousHash != tip.Header.Hash {
		return fmt.Errorf("previous hash mismatch: expected %s, got %s", tip.Header.Hash, header.PreviousHash)
	}

	// Check validator exists
	_, err := bv.blockchain.ValidatorSet.GetValidator(header.ValidatorID)
	if err != nil {
		return fmt.Errorf("validator not found: %s", header.ValidatorID)
	}

	// Check timestamp is reasonable (not in future, not too old)
	now := time.Now().Unix()
	if header.Timestamp > now+300 { // 5 minute tolerance
		return fmt.Errorf("block timestamp in future: %d > %d", header.Timestamp, now)
	}
	if header.Timestamp < tip.Header.Timestamp {
		return fmt.Errorf("block timestamp before previous block")
	}

	// Check difficulty matches network rules
	if header.Difficulty != 1 { // Simple POA doesn't require difficulty
		return fmt.Errorf("invalid difficulty: %d", header.Difficulty)
	}

	return nil
}

// validateTransaction validates a transaction
func (bv *BlockValidator) validateTransaction(txn *Transaction) error {
	if err := txn.Validate(); err != nil {
		return err
	}

	// Type-specific validation
	switch txn.Type {
	case TxTypeFlRound:
		return bv.validateFlRoundTxn(txn)
	case TxTypeStake:
		return bv.validateStakeTxn(txn)
	case TxTypeUnstake:
		return bv.validateUnstakeTxn(txn)
	case TxTypeReward:
		return bv.validateRewardTxn(txn)
	case TxTypeSmartContract:
		return bv.validateSmartContractTxn(txn)
	case TxTypeCheckpoint:
		return bv.validateCheckpointTxn(txn)
	default:
		return fmt.Errorf("unknown transaction type: %s", txn.Type)
	}
}

// validateFlRoundTxn validates a federated learning round transaction
func (bv *BlockValidator) validateFlRoundTxn(txn *Transaction) error {
	if txn.Amount != 0 {
		return fmt.Errorf("FL round transaction should have zero amount")
	}

	// Check required data fields
	if txn.Data == nil {
		return fmt.Errorf("missing transaction data")
	}

	if _, ok := txn.Data["round_id"]; !ok {
		return fmt.Errorf("missing round_id in FL transaction")
	}

	if _, ok := txn.Data["model_hash"]; !ok {
		return fmt.Errorf("missing model_hash in FL transaction")
	}

	return nil
}

// validateStakeTxn validates a staking transaction
func (bv *BlockValidator) validateStakeTxn(txn *Transaction) error {
	if txn.Amount == 0 {
		return fmt.Errorf("stake amount must be greater than zero")
	}

	if txn.To != "" {
		return fmt.Errorf("stake transaction should not have To address")
	}

	return nil
}

// validateUnstakeTxn validates an unstaking transaction
func (bv *BlockValidator) validateUnstakeTxn(txn *Transaction) error {
	if txn.Amount == 0 {
		return fmt.Errorf("unstake amount must be greater than zero")
	}

	return nil
}

// validateRewardTxn validates a reward transaction
func (bv *BlockValidator) validateRewardTxn(txn *Transaction) error {
	if txn.Amount == 0 {
		return fmt.Errorf("reward amount must be greater than zero")
	}

	if txn.To == "" {
		return fmt.Errorf("reward transaction must have To address")
	}

	return nil
}

// validateSmartContractTxn validates a smart contract transaction
func (bv *BlockValidator) validateSmartContractTxn(txn *Transaction) error {
	if txn.Gas == 0 {
		return fmt.Errorf("smart contract transaction must specify gas")
	}

	if txn.Data == nil {
		return fmt.Errorf("smart contract transaction missing data")
	}

	return nil
}

// validateCheckpointTxn validates a checkpoint transaction
func (bv *BlockValidator) validateCheckpointTxn(txn *Transaction) error {
	if txn.Data == nil {
		return fmt.Errorf("checkpoint transaction missing data")
	}

	if _, ok := txn.Data["checkpoint_hash"]; !ok {
		return fmt.Errorf("checkpoint transaction missing checkpoint_hash")
	}

	return nil
}

// validateStateRoot verifies the state root by replaying transactions
func (bv *BlockValidator) validateStateRoot(block *Block) error {
	// Clone state to check transitions
	tempState := bv.stateDB.Clone()

	// Execute all transactions
	for i, txn := range block.Transactions {
		if err := bv.executeTransaction(&txn, tempState); err != nil {
			return fmt.Errorf("transaction %d execution failed: %w", i, err)
		}
	}

	// Verify resulting state root
	tempState.IncrementVersion()
	actualRoot := tempState.ComputeRoot()

	if block.StateRoot != actualRoot {
		return fmt.Errorf("state root mismatch: expected %s, got %s", actualRoot, block.StateRoot)
	}

	return nil
}

// executeTransaction applies a transaction to state
func (bv *BlockValidator) executeTransaction(txn *Transaction, state *StateDatabase) error {
	switch txn.Type {
	case TxTypeFlRound:
		// Record FL round in state
		flKey := fmt.Sprintf("fl_round:%s", txn.Data["round_id"])
		state.Set(flKey, txn.Data)

	case TxTypeStake:
		// Update validator stake
		stakeKey := fmt.Sprintf("stake:%s", txn.From)
		currentStake := uint64(0)
		if val, err := state.Get(stakeKey); err == nil {
			currentStake = val.(uint64)
		}
		state.Set(stakeKey, currentStake+txn.Amount)

	case TxTypeUnstake:
		// Decrease validator stake
		stakeKey := fmt.Sprintf("stake:%s", txn.From)
		if val, err := state.Get(stakeKey); err == nil {
			currentStake := val.(uint64)
			if currentStake < txn.Amount {
				return fmt.Errorf("insufficient stake to unstake")
			}
			state.Set(stakeKey, currentStake-txn.Amount)
		}

	case TxTypeReward:
		// Add reward to recipient
		rewardKey := fmt.Sprintf("reward:%s", txn.To)
		currentReward := uint64(0)
		if val, err := state.Get(rewardKey); err == nil {
			currentReward = val.(uint64)
		}
		state.Set(rewardKey, currentReward+txn.Amount)

	case TxTypeSmartContract:
		// Smart contract execution (simplified)
		contractKey := fmt.Sprintf("contract:%s", txn.ID)
		state.Set(contractKey, txn.Data)

	case TxTypeCheckpoint:
		// Record checkpoint
		checkpointKey := fmt.Sprintf("checkpoint:%s", txn.Data["checkpoint_hash"])
		state.Set(checkpointKey, txn.Data)
	}

	return nil
}

// validateProofs validates cryptographic proofs attached to block
func (bv *BlockValidator) validateProofs(block *Block) error {
	if block.ProofData == nil || len(block.ProofData) == 0 {
		// Optional: proofs may not be required for all blocks
		return nil
	}

	// TODO: Implement proof validation
	// - Verify BFT consensus proofs
	// - Verify ZK-SNARK proofs from FL
	// - Verify TPM attestation proofs

	return nil
}

// ValidateBlockSequence validates a sequence of blocks
func (bv *BlockValidator) ValidateBlockSequence(blocks []*Block) error {
	for i, block := range blocks {
		if err := bv.ValidateBlock(block); err != nil {
			return fmt.Errorf("block %d validation failed: %w", i, err)
		}
	}
	return nil
}
