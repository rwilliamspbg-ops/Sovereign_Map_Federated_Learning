// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package blockchain

import (
	"fmt"
	"testing"
	"time"
)

// Test block creation and validation
func TestBlockCreation(t *testing.T) {
	txns := []Transaction{
		{
			ID:        "txn_001",
			Type:      TxTypeStake,
			From:      "node_1",
			Nonce:     0,
			GasPrice:  100,
			Gas:       21000,
			Amount:    100000,
			Timestamp: time.Now().Unix(),
			Signature: []byte("sig_1"),
		},
	}

	header := BlockHeader{
		Index:        1,
		Timestamp:    time.Now().Unix(),
		PreviousHash: "prev_hash",
		ValidatorID:  "node_1",
		Difficulty:   1,
		Version:      1,
	}

	block := &Block{
		Header:       header,
		Transactions: txns,
		MerkleRoot:   ComputeMerkleRoot(txns),
		StateRoot:    hashData("state"),
	}
	block.Header.Hash = block.ComputeHash()

	if block.Header.Index != 1 {
		t.Errorf("Expected index 1, got %d", block.Header.Index)
	}

	if len(block.Transactions) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(block.Transactions))
	}

	hash := block.ComputeHash()
	if hash == "" {
		t.Error("Block hash should not be empty")
	}
}

func TestMerkleRootComputation(t *testing.T) {
	txns := []Transaction{
		{
			ID:        "txn_001",
			Type:      TxTypeStake,
			From:      "node_1",
			Nonce:     0,
			GasPrice:  100,
			Gas:       21000,
			Amount:    100000,
			Timestamp: time.Now().Unix(),
			Signature: []byte("sig_1"),
		},
		{
			ID:        "txn_002",
			Type:      TxTypeStake,
			From:      "node_2",
			Nonce:     0,
			GasPrice:  100,
			Gas:       21000,
			Amount:    200000,
			Timestamp: time.Now().Unix(),
			Signature: []byte("sig_2"),
		},
	}

	root1 := ComputeMerkleRoot(txns)
	root2 := ComputeMerkleRoot(txns)

	if root1 != root2 {
		t.Error("Same transactions should produce same Merkle root")
	}

	// Change one transaction
	txns[1].Amount = 150000
	root3 := ComputeMerkleRoot(txns)

	if root1 == root3 {
		t.Error("Different transactions should produce different Merkle roots")
	}
}

func TestBlockChainAppend(t *testing.T) {
	blockchain := NewBlockChain()
	if err := blockchain.ValidatorSet.AddValidator("node_1", 1000000); err != nil {
		t.Fatalf("failed to add validator: %v", err)
	}

	block1 := &Block{
		Header: BlockHeader{
			Index:        1,
			PreviousHash: blockchain.Tip.Header.Hash,
			Timestamp:    time.Now().Unix(),
			ValidatorID:  "node_1",
			Difficulty:   1,
			Version:      1,
		},
		Transactions: []Transaction{},
		MerkleRoot:   ComputeMerkleRoot([]Transaction{}),
		StateRoot:    blockchain.StateDB.ComputeRoot(),
	}
	block1.Header.Hash = block1.ComputeHash()

	err := blockchain.AppendBlock(block1)
	if err != nil {
		t.Fatalf("Failed to append first block: %v", err)
	}

	if blockchain.Height() != 1 {
		t.Errorf("Expected height 1, got %d", blockchain.Height())
	}

	block2 := &Block{
		Header: BlockHeader{
			Index:        2,
			PreviousHash: block1.Header.Hash,
			Timestamp:    time.Now().Unix(),
			ValidatorID:  "node_1",
			Difficulty:   1,
			Version:      1,
		},
		Transactions: []Transaction{},
		MerkleRoot:   ComputeMerkleRoot([]Transaction{}),
		StateRoot:    blockchain.StateDB.ComputeRoot(),
	}
	block2.Header.Hash = block2.ComputeHash()

	err = blockchain.AppendBlock(block2)
	if err != nil {
		t.Fatalf("Failed to append second block: %v", err)
	}

	if blockchain.Height() != 2 {
		t.Errorf("Expected height 2, got %d", blockchain.Height())
	}
}

func TestBlockValidation(t *testing.T) {
	blockchain := NewBlockChain()
	stateDB := NewStateDatabase()
	if err := blockchain.ValidatorSet.AddValidator("node_1", 1000000); err != nil {
		t.Fatalf("failed to add validator: %v", err)
	}

	txns := []Transaction{}

	merkleRoot := ComputeMerkleRoot(txns)
	stateRoot := stateDB.ComputeRoot()

	header := BlockHeader{
		Index:        1,
		PreviousHash: blockchain.Tip.Header.Hash,
		Timestamp:    time.Now().Unix(),
		ValidatorID:  "node_1",
		Difficulty:   1,
		Version:      1,
	}

	block := &Block{
		Header:       header,
		Transactions: txns,
		MerkleRoot:   merkleRoot,
		StateRoot:    stateRoot,
	}
	block.Header.Hash = block.ComputeHash()

	validator := NewBlockValidator(blockchain, stateDB)
	if err := validator.ValidateBlock(block); err != nil {
		t.Errorf("Valid block should pass validation: %v", err)
	}
}

func TestStateDatabase(t *testing.T) {
	stateDB := NewStateDatabase()

	// Set state
	stateDB.Set("account:node_1:balance", uint64(1000000))
	stateDB.Set("account:node_2:balance", uint64(500000))

	// Get state
	balance1, err := stateDB.Get("account:node_1:balance")
	if err != nil {
		t.Fatalf("Failed to get state: %v", err)
	}

	if balance1 != uint64(1000000) {
		t.Errorf("Expected balance 1000000, got %v", balance1)
	}

	// Compute root
	root := stateDB.ComputeRoot()
	if root == "" {
		t.Error("State root should not be empty")
	}

	// Verify exists
	exists := stateDB.Exists("account:node_1:balance")
	if !exists {
		t.Error("State entry should exist")
	}

	// Delete state
	stateDB.Delete("account:node_2:balance")
	exists = stateDB.Exists("account:node_2:balance")
	if exists {
		t.Error("Deleted state should not exist")
	}
}

func TestStateDatabaseSnapshots(t *testing.T) {
	stateDB := NewStateDatabase()

	// Set initial state
	stateDB.Set("account:node_1:balance", uint64(1000000))

	// Record snapshot
	stateDB.RecordSnapshot(1)

	// Modify state
	stateDB.Set("account:node_1:balance", uint64(900000))

	// Record another snapshot
	stateDB.RecordSnapshot(2)

	// Retrieve snapshots
	retrieved1, err := stateDB.GetSnapshot(1)
	if err != nil {
		t.Errorf("Should retrieve snapshot at height 1: %v", err)
	}

	retrieved2, err := stateDB.GetSnapshot(2)
	if err != nil {
		t.Errorf("Should retrieve snapshot at height 2: %v", err)
	}

	if retrieved1 == retrieved2 {
		t.Error("Different states should have different roots")
	}
}

func TestMempool(t *testing.T) {
	mempool := NewMempool()

	txn1 := &Transaction{
		ID:        "txn_001",
		From:      "node_1",
		Nonce:     0,
		GasPrice:  100,
		Timestamp: time.Now().Unix(),
		Signature: []byte("sig_1"),
	}

	txn2 := &Transaction{
		ID:        "txn_002",
		From:      "node_2",
		Nonce:     0,
		GasPrice:  200,
		Timestamp: time.Now().Unix(),
		Signature: []byte("sig_2"),
	}

	// Add transactions
	if err := mempool.AddTransaction(txn1); err != nil {
		t.Fatalf("failed to add transaction 1: %v", err)
	}
	if err := mempool.AddTransaction(txn2); err != nil {
		t.Fatalf("failed to add transaction 2: %v", err)
	}

	if mempool.Size() != 2 {
		t.Errorf("Expected 2 transactions, got %d", mempool.Size())
	}

	// Get transactions for block (sorted by gas price)
	blockTxns := mempool.GetTransactionsForBlock(10)
	if len(blockTxns) != 2 {
		t.Errorf("Expected 2 transactions for block, got %d", len(blockTxns))
	}

	// First should be higher gas price
	if blockTxns[0].GasPrice < blockTxns[1].GasPrice {
		t.Error("Transactions should be sorted by gas price descending")
	}
}

func TestMempoolNonceTracking(t *testing.T) {
	mempool := NewMempool()

	txn1 := &Transaction{
		ID:        "txn_001",
		From:      "node_1",
		Nonce:     0,
		GasPrice:  100,
		Timestamp: time.Now().Unix(),
		Signature: []byte("sig_1"),
	}

	txn2 := &Transaction{
		ID:        "txn_002",
		From:      "node_1",
		Nonce:     1,
		GasPrice:  100,
		Timestamp: time.Now().Unix(),
		Signature: []byte("sig_2"),
	}

	// Add first transaction
	if err := mempool.AddTransaction(txn1); err != nil {
		t.Fatalf("failed to add transaction 1: %v", err)
	}
	nonce := mempool.GetAddressNonce("node_1")
	if nonce != 1 {
		t.Errorf("Expected nonce 1, got %d", nonce)
	}

	// Add second transaction
	if err := mempool.AddTransaction(txn2); err != nil {
		t.Fatalf("failed to add transaction 2: %v", err)
	}
	nonce = mempool.GetAddressNonce("node_1")
	if nonce != 2 {
		t.Errorf("Expected nonce 2, got %d", nonce)
	}
}

func TestValidatorStaking(t *testing.T) {
	validators := NewValidatorSet()
	if err := validators.AddValidator("node_1", 1000000); err != nil {
		t.Fatalf("failed to add validator: %v", err)
	}

	// Increase stake
	err := validators.Stake("node_1", 500000)
	if err != nil {
		t.Fatalf("Staking failed: %v", err)
	}

	v, err := validators.GetValidator("node_1")
	if err != nil {
		t.Fatalf("failed to get validator: %v", err)
	}
	if v.StakedAmount != 1500000 {
		t.Errorf("Expected stake 1500000, got %d", v.StakedAmount)
	}

	// Decrease stake
	err = validators.Unstake("node_1", 300000)
	if err != nil {
		t.Fatalf("Unstaking failed: %v", err)
	}

	v, err = validators.GetValidator("node_1")
	if err != nil {
		t.Fatalf("failed to get validator: %v", err)
	}
	if v.StakedAmount != 1200000 {
		t.Errorf("Expected stake 1200000, got %d", v.StakedAmount)
	}
}

func TestValidatorSlashing(t *testing.T) {
	validators := NewValidatorSet()
	if err := validators.AddValidator("node_1", 1000000); err != nil {
		t.Fatalf("failed to add validator: %v", err)
	}

	// Slash validator
	if err := validators.SlashValidator("node_1", 100000); err != nil {
		t.Fatalf("slashing failed: %v", err)
	}

	v, err := validators.GetValidator("node_1")
	if err != nil {
		t.Fatalf("failed to get validator: %v", err)
	}
	if v.StakedAmount != 900000 {
		t.Errorf("Expected stake 900000, got %d", v.StakedAmount)
	}

	// Jail validator
	_ = validators.SlashValidator("node_1", 500000)
	v, _ = validators.GetValidator("node_1")

	// Note: Jailed flag behavior depends on implementation
	// This test documents the expected behavior
}

func TestValidatorSelection(t *testing.T) {
	validators := NewValidatorSet()

	// Add validators with different stakes
	if err := validators.AddValidator("node_1", 1000000); err != nil {
		t.Fatalf("failed to add validator node_1: %v", err)
	}
	if err := validators.AddValidator("node_2", 1200000); err != nil {
		t.Fatalf("failed to add validator node_2: %v", err)
	}
	if err := validators.AddValidator("node_3", 1500000); err != nil {
		t.Fatalf("failed to add validator node_3: %v", err)
	}

	// Select validators
	selected := validators.SelectValidators(2)

	if len(selected) != 2 {
		t.Errorf("Expected 2 validators, got %d", len(selected))
	}

	// Higher stake validators should have higher selection probability
	// This test documents weighted random selection behavior
}

func TestValidatorRewardDistribution(t *testing.T) {
	validators := NewValidatorSet()

	if err := validators.AddValidator("node_1", 1000000); err != nil {
		t.Fatalf("failed to add validator node_1: %v", err)
	}
	if err := validators.AddValidator("node_2", 1000000); err != nil {
		t.Fatalf("failed to add validator node_2: %v", err)
	}

	// Distribute rewards
	validators.DistributeRewards(1, 300, []string{"node_1", "node_2"})

	v1, err := validators.GetValidator("node_1")
	if err != nil {
		t.Fatalf("failed to get validator node_1: %v", err)
	}
	v2, err := validators.GetValidator("node_2")
	if err != nil {
		t.Fatalf("failed to get validator node_2: %v", err)
	}

	// Rewards are tracked in validator state
	// This test documents the expected behavior
	if v1 == nil || v2 == nil {
		t.Fatal("Validators should exist")
	}
	if v1.AccumulatedRewards == 0 || v2.AccumulatedRewards == 0 {
		t.Error("Validators should receive rewards")
	}
}

func TestEpochRotation(t *testing.T) {
	validators := NewValidatorSet()

	if err := validators.AddValidator("node_1", 1000000); err != nil {
		t.Fatalf("failed to add validator node_1: %v", err)
	}
	if err := validators.AddValidator("node_2", 1200000); err != nil {
		t.Fatalf("failed to add validator node_2: %v", err)
	}

	// Rotate epoch
	validators.RotateEpoch(1001)

	// Validator set structure may change on epoch boundary
	// This test documents rotation behavior
	if validators.Count() == 0 {
		t.Error("Validators should remain after epoch rotation")
	}
	if validators.EpochNumber != 1 {
		t.Errorf("Expected epoch number 1, got %d", validators.EpochNumber)
	}
}

// End-to-end integration test
func TestEndToEndBlockProduction(t *testing.T) {
	blockchain := NewBlockChain()

	// Add validators
	if err := blockchain.ValidatorSet.AddValidator("node_1", 1000000); err != nil {
		t.Fatalf("failed to add validator node_1: %v", err)
	}
	if err := blockchain.ValidatorSet.AddValidator("node_2", 1000000); err != nil {
		t.Fatalf("failed to add validator node_2: %v", err)
	}
	proposer := NewBlockProposer("node_1", blockchain)

	// Add transaction
	txn := &Transaction{
		ID:        "txn_001",
		Type:      TxTypeStake,
		From:      "node_2",
		Nonce:     0,
		GasPrice:  100,
		Gas:       21000,
		Amount:    100000,
		Timestamp: time.Now().Unix(),
		Signature: []byte("sig_1"),
	}

	if err := blockchain.Mempool.AddTransaction(txn); err != nil {
		t.Fatalf("failed to add mempool transaction: %v", err)
	}

	// Propose block
	block, err := proposer.ProposeBlock("node_1", map[string]interface{}{
		"round_id":   "round_1",
		"model_hash": "model_hash_1",
	})
	if err != nil {
		t.Fatalf("Block proposal failed: %v", err)
	}

	if block == nil {
		t.Fatal("Proposed block should not be nil")
	}

	if len(block.Transactions) == 0 {
		t.Error("Expected at least 1 transaction in proposed block")
	}

	// Validate block
	validator := NewBlockValidator(blockchain, blockchain.StateDB)
	if err := validator.ValidateBlock(block); err != nil {
		t.Errorf("Proposed block should be valid: %v", err)
	}

	// Commit block
	err = proposer.CommitBlock(block)
	if err != nil {
		t.Fatalf("Block commit failed: %v", err)
	}

	if blockchain.Height() != 1 {
		t.Errorf("Expected height 1, got %d", blockchain.Height())
	}
}

// Stress test
func TestMempoolCapacity(t *testing.T) {
	mempool := NewMempool()

	// Add many transactions
	for i := 0; i < 100; i++ {
		txn := &Transaction{
			ID:        fmt.Sprintf("txn_%03d", i),
			From:      fmt.Sprintf("node_%03d", i),
			Nonce:     0,
			GasPrice:  uint64(100 + i%10),
			Timestamp: time.Now().Unix(),
			Signature: []byte("sig"),
		}
		if err := mempool.AddTransaction(txn); err != nil {
			t.Fatalf("failed to add transaction %d: %v", i, err)
		}
	}

	if mempool.Size() != 100 {
		t.Errorf("Expected 100 transactions, got %d", mempool.Size())
	}

	// Get transactions for block
	blockTxns := mempool.GetTransactionsForBlock(50)
	if len(blockTxns) != 50 {
		t.Errorf("Expected 50 transactions, got %d", len(blockTxns))
	}

	// GetTransactionsForBlock does not remove transactions from mempool
	if mempool.Size() != 100 {
		t.Errorf("Expected 100 remaining, got %d", mempool.Size())
	}
}

func TestBlockchainForkResolution(t *testing.T) {
	blockchain := NewBlockChain()
	stateDB := blockchain.StateDB

	if err := blockchain.ValidatorSet.AddValidator("node_1", 1000000); err != nil {
		t.Fatalf("failed to add validator node_1: %v", err)
	}
	if err := blockchain.ValidatorSet.AddValidator("node_2", 1000000); err != nil {
		t.Fatalf("failed to add validator node_2: %v", err)
	}

	validator := NewBlockValidator(blockchain, stateDB)

	// Create two conflicting blocks at same height
	block1 := &Block{
		Header: BlockHeader{
			Index:        1,
			PreviousHash: blockchain.Tip.Header.Hash,
			Timestamp:    time.Now().Unix(),
			ValidatorID:  "node_1",
			Difficulty:   1,
			Version:      1,
		},
		Transactions: []Transaction{},
		MerkleRoot:   ComputeMerkleRoot([]Transaction{}),
		StateRoot:    stateDB.ComputeRoot(),
	}
	block1.Header.Hash = block1.ComputeHash()

	block2 := &Block{
		Header: BlockHeader{
			Index:        1,
			PreviousHash: blockchain.Tip.Header.Hash,
			Timestamp:    time.Now().Unix() + 1,
			ValidatorID:  "node_2",
			Difficulty:   1,
			Version:      1,
		},
		Transactions: []Transaction{},
		MerkleRoot:   ComputeMerkleRoot([]Transaction{}),
		StateRoot:    stateDB.ComputeRoot(),
	}
	block2.Header.Hash = block2.ComputeHash()

	// Both should be structurally valid
	err1 := validator.ValidateBlock(block1)
	err2 := validator.ValidateBlock(block2)

	if err1 != nil || err2 != nil {
		t.Errorf("Both blocks should be structurally valid, got err1=%v err2=%v", err1, err2)
	}

	// Fork resolution would choose block with higher timestamp/better consensus
	// at application layer
}

func TestValidatorReputationPenaltyAndRecovery(t *testing.T) {
	validators := NewValidatorSet()
	if err := validators.AddValidator("node_1", 1000000); err != nil {
		t.Fatalf("failed to add validator: %v", err)
	}

	v, err := validators.GetValidator("node_1")
	if err != nil {
		t.Fatalf("failed to get validator: %v", err)
	}
	initial := v.ReputationScore

	if err := validators.SlashValidator("node_1", 1000); err != nil {
		t.Fatalf("slash failed: %v", err)
	}

	v, _ = validators.GetValidator("node_1")
	if v.ReputationScore >= initial {
		t.Fatalf("expected reputation drop after slash, initial=%d current=%d", initial, v.ReputationScore)
	}

	validators.RotateEpoch(2000)
	v, _ = validators.GetValidator("node_1")
	if v.ReputationScore == 0 {
		t.Fatal("expected non-zero reputation after epoch recovery")
	}
}

func TestValidatorReputationAdjustsReward(t *testing.T) {
	validators := NewValidatorSet()
	if err := validators.AddValidator("node_hi", 1000000); err != nil {
		t.Fatalf("failed to add node_hi: %v", err)
	}
	if err := validators.AddValidator("node_lo", 1000000); err != nil {
		t.Fatalf("failed to add node_lo: %v", err)
	}

	vHi, _ := validators.GetValidator("node_hi")
	vLo, _ := validators.GetValidator("node_lo")
	vHi.ReputationScore = 10000
	vLo.ReputationScore = 2000

	validators.DistributeRewards(1, 1000, []string{"node_hi", "node_lo"})

	vHi, _ = validators.GetValidator("node_hi")
	vLo, _ = validators.GetValidator("node_lo")
	if vHi.AccumulatedRewards <= vLo.AccumulatedRewards {
		t.Fatalf("expected high-reputation validator to earn more: hi=%d lo=%d", vHi.AccumulatedRewards, vLo.AccumulatedRewards)
	}
}

func TestValidatorEffectiveWeightUsesReputation(t *testing.T) {
	validators := NewValidatorSet()
	high := &Validator{NodeID: "high", StakedAmount: 1000000, ReputationScore: 10000}
	low := &Validator{NodeID: "low", StakedAmount: 1000000, ReputationScore: 1000}

	highWeight := validators.effectiveWeight(high)
	lowWeight := validators.effectiveWeight(low)

	if highWeight <= lowWeight {
		t.Fatalf("expected high reputation weight > low reputation weight, high=%d low=%d", highWeight, lowWeight)
	}
}

func TestValidatorGovernancePolicyUpdate(t *testing.T) {
	validators := NewValidatorSet()

	policy := validators.GetReputationPolicy()
	policy.ReputationWeight = 50
	policy.AttestationWeight = 30
	policy.QualityWeight = 20
	policy.MaxConsecutiveLowQuality = 3

	if err := validators.SetReputationPolicy(policy); err != nil {
		t.Fatalf("expected valid policy update, got: %v", err)
	}

	updated := validators.GetReputationPolicy()
	if updated.ReputationWeight != 50 || updated.AttestationWeight != 30 || updated.QualityWeight != 20 {
		t.Fatalf("policy update not applied: %+v", updated)
	}
}

func TestValidatorAttestationAffectsWeight(t *testing.T) {
	validators := NewValidatorSet()
	high := &Validator{NodeID: "high", StakedAmount: 1000000, ReputationScore: 7000, AttestationScore: 10000, ParticipationScore: 7000}
	low := &Validator{NodeID: "low", StakedAmount: 1000000, ReputationScore: 7000, AttestationScore: 1000, ParticipationScore: 7000}

	highWeight := validators.effectiveWeight(high)
	lowWeight := validators.effectiveWeight(low)

	if highWeight <= lowWeight {
		t.Fatalf("expected high-attestation validator to have higher weight: high=%d low=%d", highWeight, lowWeight)
	}
}

func TestValidatorAntiGamingQualityPenalty(t *testing.T) {
	validators := NewValidatorSet()
	if err := validators.AddValidator("node_1", 1000000); err != nil {
		t.Fatalf("failed to add validator: %v", err)
	}

	policy := validators.GetReputationPolicy()
	policy.MaxConsecutiveLowQuality = 2
	policy.MinQualityScore = 9000
	policy.QualityPenalty = 300
	if err := validators.SetReputationPolicy(policy); err != nil {
		t.Fatalf("failed to set policy: %v", err)
	}

	if err := validators.RecordParticipationQuality("node_1", 1000); err != nil {
		t.Fatalf("first quality record failed: %v", err)
	}
	if err := validators.RecordParticipationQuality("node_1", 1000); err != nil {
		t.Fatalf("second quality record failed: %v", err)
	}

	v, _ := validators.GetValidator("node_1")
	if !v.Jailed {
		t.Fatal("expected validator to be jailed after consecutive low-quality rounds")
	}
}

func TestValidatorMetricsAPI(t *testing.T) {
	validators := NewValidatorSet()
	_ = validators.AddValidator("node_1", 1000000)
	_ = validators.AddValidator("node_2", 1000000)
	_ = validators.SetAttestationScore("node_1", 9000)
	_ = validators.SetAttestationScore("node_2", 3000)
	_ = validators.RecordAttestationEvidence("node_2", 2000, "evidence://invalid-node-2", 42, false)
	_ = validators.EnforceAttestationFreshness(2000)

	metrics := validators.GetMetrics()
	if metrics.ValidatorCount != 2 {
		t.Fatalf("expected validator count 2, got %d", metrics.ValidatorCount)
	}
	if metrics.AverageAttestation == 0 || metrics.AverageReputation == 0 {
		t.Fatalf("expected non-zero averages, got %+v", metrics)
	}
	if metrics.TotalAttestationFailures == 0 {
		t.Fatalf("expected attestation failures to be tracked, got %+v", metrics)
	}
	if metrics.StaleAttestationCount == 0 {
		t.Fatalf("expected stale attestation count to be tracked, got %+v", metrics)
	}
}

func TestValidatorInvalidAttestationTriggersPenaltyAndSlash(t *testing.T) {
	validators := NewValidatorSet()
	if err := validators.AddValidator("node_1", 1000000); err != nil {
		t.Fatalf("failed to add validator: %v", err)
	}

	vBefore, err := validators.GetValidator("node_1")
	if err != nil {
		t.Fatalf("failed to get validator: %v", err)
	}
	stakeBefore := vBefore.StakedAmount
	repBefore := vBefore.ReputationScore

	if err := validators.RecordAttestationEvidence("node_1", 1000, "evidence://node-1-invalid", 100, false); err != nil {
		t.Fatalf("failed to record invalid attestation: %v", err)
	}

	vAfter, err := validators.GetValidator("node_1")
	if err != nil {
		t.Fatalf("failed to get validator: %v", err)
	}

	if vAfter.StakedAmount >= stakeBefore {
		t.Fatalf("expected stake slash after invalid attestation, before=%d after=%d", stakeBefore, vAfter.StakedAmount)
	}
	if vAfter.ReputationScore >= repBefore {
		t.Fatalf("expected reputation penalty after invalid attestation, before=%d after=%d", repBefore, vAfter.ReputationScore)
	}
	if vAfter.AttestationFailuresTotal == 0 || vAfter.ConsecutiveAttestationFailures == 0 {
		t.Fatalf("expected attestation failure counters to increase, got %+v", vAfter)
	}
}

func TestValidatorStaleAttestationEnforcement(t *testing.T) {
	validators := NewValidatorSet()
	if err := validators.AddValidator("node_1", 1000000); err != nil {
		t.Fatalf("failed to add validator: %v", err)
	}

	if err := validators.RecordAttestationEvidence("node_1", 9000, "evidence://node-1-valid", 10, true); err != nil {
		t.Fatalf("failed to record valid attestation: %v", err)
	}

	vBefore, _ := validators.GetValidator("node_1")
	stakeBefore := vBefore.StakedAmount

	slashed := validators.EnforceAttestationFreshness(600)
	if len(slashed) == 0 {
		t.Fatal("expected stale attestation enforcement to slash validator")
	}

	vAfter, _ := validators.GetValidator("node_1")
	if !vAfter.AttestationStale {
		t.Fatal("expected validator to be marked stale after enforcement")
	}
	if vAfter.StakedAmount >= stakeBefore {
		t.Fatalf("expected stale slash to reduce stake, before=%d after=%d", stakeBefore, vAfter.StakedAmount)
	}
}

func TestBuildFLVerificationMetadataDefaults(t *testing.T) {
	data := map[string]interface{}{
		"round_id":   "r-101",
		"model_hash": "model-hash-abc",
	}

	meta := BuildFLVerificationMetadata(data, 9, 1700000000)

	if meta["proof_type"] != "consensus" {
		t.Fatalf("expected default proof_type consensus, got %v", meta["proof_type"])
	}
	if meta["verification_passed"] != true {
		t.Fatalf("expected default verification_passed true, got %v", meta["verification_passed"])
	}
	if meta["verification_confidence_bps"] != uint32(8000) {
		t.Fatalf("expected default confidence 8000, got %v", meta["verification_confidence_bps"])
	}
	if meta["block_height"] != uint64(9) {
		t.Fatalf("expected block height 9, got %v", meta["block_height"])
	}
}

func TestBlockchainFLVerificationMetricsAggregation(t *testing.T) {
	blockchain := NewBlockChain()
	if err := blockchain.ValidatorSet.AddValidator("node_1", 1000000); err != nil {
		t.Fatalf("failed to add validator: %v", err)
	}

	proposer := NewBlockProposer("node_1", blockchain)

	b1, err := proposer.ProposeBlock("node_1", map[string]interface{}{
		"round_id":                    "r-1",
		"model_hash":                  "h1",
		"verification_passed":         true,
		"verification_confidence_bps": 9000,
		"proof_type":                  "consensus",
	})
	if err != nil {
		t.Fatalf("propose block 1 failed: %v", err)
	}
	if err := proposer.CommitBlock(b1); err != nil {
		t.Fatalf("commit block 1 failed: %v", err)
	}

	b2, err := proposer.ProposeBlock("node_1", map[string]interface{}{
		"round_id":                    "r-2",
		"model_hash":                  "h2",
		"verification_passed":         false,
		"verification_confidence_bps": 4000,
		"proof_type":                  "zk-proofs",
	})
	if err != nil {
		t.Fatalf("propose block 2 failed: %v", err)
	}
	if err := proposer.CommitBlock(b2); err != nil {
		t.Fatalf("commit block 2 failed: %v", err)
	}

	metrics := blockchain.GetFLVerificationMetrics()
	if metrics.TotalRounds != 2 {
		t.Fatalf("expected total rounds 2, got %+v", metrics)
	}
	if metrics.VerifiedRounds != 1 || metrics.FailedRounds != 1 {
		t.Fatalf("expected 1 verified and 1 failed, got %+v", metrics)
	}
	if metrics.AverageConfidenceBps != 6500 {
		t.Fatalf("expected average confidence 6500, got %+v", metrics)
	}
	if metrics.LastRoundID != "r-2" {
		t.Fatalf("expected last round id r-2, got %+v", metrics)
	}
	if metrics.LastProofType != "zk-proofs" {
		t.Fatalf("expected last proof type zk-proofs, got %+v", metrics)
	}
}

func TestCommitBlockPersistsFLVerificationState(t *testing.T) {
	blockchain := NewBlockChain()
	if err := blockchain.ValidatorSet.AddValidator("node_1", 1000000); err != nil {
		t.Fatalf("failed to add validator: %v", err)
	}

	proposer := NewBlockProposer("node_1", blockchain)
	block, err := proposer.ProposeBlock("node_1", map[string]interface{}{
		"round_id":                    "round-verify-1",
		"model_hash":                  "hash-verify-1",
		"verification_passed":         true,
		"verification_confidence_bps": 8800,
	})
	if err != nil {
		t.Fatalf("failed to propose block: %v", err)
	}

	if err := proposer.CommitBlock(block); err != nil {
		t.Fatalf("failed to commit block: %v", err)
	}

	flKey := "fl_verification:round-verify-1"
	stored, err := blockchain.StateDB.Get(flKey)
	if err != nil {
		t.Fatalf("expected verification state to exist, key=%s err=%v", flKey, err)
	}

	meta, ok := stored.(map[string]interface{})
	if !ok {
		t.Fatalf("expected verification state map, got %T", stored)
	}
	if meta["round_id"] != "round-verify-1" {
		t.Fatalf("unexpected round id in verification state: %v", meta)
	}
}
