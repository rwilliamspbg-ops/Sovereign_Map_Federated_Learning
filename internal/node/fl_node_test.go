// Copyright 2026 Sovereign-Mohawk Core Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package node

import (
	"context"
	"testing"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/blockchain"
)

// TestFLNodeCreation tests basic FL node creation
func TestFLNodeCreation(t *testing.T) {
	nodeID := "test-node-001"
	region := "us-west"

	node := NewFLNode(nodeID, region)

	if node.nodeID != nodeID {
		t.Errorf("expected nodeID %s, got %s", nodeID, node.nodeID)
	}
	if node.region != region {
		t.Errorf("expected region %s, got %s", region, node.region)
	}
	if node.trainingRound != 0 {
		t.Errorf("expected trainingRound 0, got %d", node.trainingRound)
	}
}

// TestFLNodeBlockchainSetup tests blockchain configuration
func TestFLNodeBlockchainSetup(t *testing.T) {
	node := NewFLNode("test-node", "us-west")
	bc := blockchain.NewBlockChain()
	mempool := blockchain.NewMempool()
	initialStake := uint64(10000)

	err := node.SetupBlockchain(bc, mempool, initialStake)
	if err != nil {
		t.Fatalf("SetupBlockchain failed: %v", err)
	}

	if node.nodeStake != initialStake {
		t.Errorf("expected stake %d, got %d", initialStake, node.nodeStake)
	}

	// Test nil blockchain
	node2 := NewFLNode("test-node-2", "us-west")
	err = node2.SetupBlockchain(nil, mempool, initialStake)
	if err == nil {
		t.Error("expected error for nil blockchain")
	}
}

// TestTrainRound tests local training execution
func TestTrainRound(t *testing.T) {
	node := NewFLNode("test-node", "us-west")
	bc := blockchain.NewBlockChain()
	mempool := blockchain.NewMempool()

	if err := node.SetupBlockchain(bc, mempool, 10000); err != nil {
		t.Fatalf("SetupBlockchain failed: %v", err)
	}

	ctx := context.Background()
	result, err := node.TrainRound(ctx)
	if err != nil {
		t.Fatalf("TrainRound failed: %v", err)
	}

	if result.RoundNum != 1 {
		t.Errorf("expected round 1, got %d", result.RoundNum)
	}
	if result.Accuracy <= 0.8 || result.Accuracy >= 1.0 {
		t.Errorf("expected accuracy between 0.85-1.0, got %f", result.Accuracy)
	}
	if result.WeightsHash == "" {
		t.Error("expected non-empty weights hash")
	}
	if result.DataSamples <= 0 {
		t.Errorf("expected positive data samples, got %d", result.DataSamples)
	}
}

// TestSubmitFlRound tests FL transaction submission
func TestSubmitFlRound(t *testing.T) {
	node := NewFLNode("test-node", "us-west")
	bc := blockchain.NewBlockChain()
	mempool := blockchain.NewMempool()

	if err := node.SetupBlockchain(bc, mempool, 10000); err != nil {
		t.Fatalf("SetupBlockchain failed: %v", err)
	}

	ctx := context.Background()
	result, _ := node.TrainRound(ctx)

	// Test submission without blockchain setup
	node2 := NewFLNode("test-node-2", "us-west")
	_, err := node2.SubmitFlRound(ctx, result, []byte{})
	if err == nil {
		t.Error("expected error when blockchain not configured")
	}

	// Test successful submission
	txID, err := node.SubmitFlRound(ctx, result, []byte("checkpoint"))
	if err != nil {
		t.Fatalf("SubmitFlRound failed: %v", err)
	}

	if txID == "" {
		t.Error("expected non-empty transaction ID")
	}

	// Verify transaction is tracked
	state := node.GetState()
	if state.PendingTransactionCount != 1 {
		t.Errorf("expected 1 pending transaction, got %d", state.PendingTransactionCount)
	}
}

// TestTransactionIncrement tests nonce increment for replay protection
func TestTransactionIncrement(t *testing.T) {
	node := NewFLNode("test-node", "us-west")
	bc := blockchain.NewBlockChain()
	mempool := blockchain.NewMempool()

	if err := node.SetupBlockchain(bc, mempool, 10000); err != nil {
		t.Fatalf("SetupBlockchain failed: %v", err)
	}

	ctx := context.Background()

	// Submit multiple transactions
	for i := 0; i < 3; i++ {
		result, _ := node.TrainRound(ctx)
		if _, err := node.SubmitFlRound(ctx, result, []byte{}); err != nil {
			t.Fatalf("SubmitFlRound failed: %v", err)
		}
	}

	// Verify nonce incremented
	if node.transactionNonce != 3 {
		t.Errorf("expected nonce 3, got %d", node.transactionNonce)
	}
}

// TestClaimRewards tests reward claiming
func TestClaimRewards(t *testing.T) {
	node := NewFLNode("test-node", "us-west")
	bc := blockchain.NewBlockChain()
	mempool := blockchain.NewMempool()

	if err := node.SetupBlockchain(bc, mempool, 10000); err != nil {
		t.Fatalf("SetupBlockchain failed: %v", err)
	}

	ctx := context.Background()

	// First claim
	reward1, err := node.ClaimRewards(ctx)
	if err != nil {
		t.Fatalf("ClaimRewards failed: %v", err)
	}

	if reward1 <= 0 {
		t.Errorf("expected positive reward, got %d", reward1)
	}

	// Verify accumulated
	state := node.GetState()
	if state.AccumulatedReward != reward1 {
		t.Errorf("expected accumulated reward %d, got %d", reward1, state.AccumulatedReward)
	}

	// Second claim
	reward2, _ := node.ClaimRewards(ctx)
	state = node.GetState()
	if state.AccumulatedReward != (reward1 + reward2) {
		t.Errorf("expected accumulated reward %d, got %d", reward1+reward2, state.AccumulatedReward)
	}
}

// TestConfirmTransaction tests transaction confirmation
func TestConfirmTransaction(t *testing.T) {
	node := NewFLNode("test-node", "us-west")
	bc := blockchain.NewBlockChain()
	mempool := blockchain.NewMempool()

	node.SetupBlockchain(bc, mempool, 10000)

	ctx := context.Background()
	result, _ := node.TrainRound(ctx)
	txID, _ := node.SubmitFlRound(ctx, result, []byte{})

	// Confirm transaction
	blockHeight := uint64(1)
	err := node.ConfirmTransaction(txID, blockHeight)
	if err != nil {
		t.Fatalf("ConfirmTransaction failed: %v", err)
	}

	// Verify transaction status
	node.mu.RLock()
	flTx := node.pendingTransactions[txID]
	node.mu.RUnlock()

	if flTx.Status != TransactionConfirmed {
		t.Errorf("expected confirmed status, got %v", flTx.Status)
	}
	if flTx.BlockHeight != blockHeight {
		t.Errorf("expected block height %d, got %d", blockHeight, flTx.BlockHeight)
	}
}

// TestCompleteRound tests round completion
func TestCompleteRound(t *testing.T) {
	node := NewFLNode("test-node", "us-west")
	bc := blockchain.NewBlockChain()
	mempool := blockchain.NewMempool()

	node.SetupBlockchain(bc, mempool, 10000)

	ctx := context.Background()

	// Complete multiple rounds
	for i := 0; i < 3; i++ {
		if err := node.CompleteRound(ctx, uint64(i+1)); err != nil {
			t.Fatalf("CompleteRound failed: %v", err)
		}
	}

	state := node.GetState()
	if state.TotalRoundsCompleted != 3 {
		t.Errorf("expected 3 completed rounds, got %d", state.TotalRoundsCompleted)
	}
}

// TestNodePoolCreation tests node pool creation
func TestNodePoolCreation(t *testing.T) {
	bc := blockchain.NewBlockChain()
	mempool := blockchain.NewMempool()
	validators := blockchain.NewValidatorSet()
	blockProposer := blockchain.NewBlockProposer("proposer", bc)

	pool := NewNodePool(bc, mempool, validators, blockProposer)

	if pool.blockchain != bc {
		t.Error("blockchain not set in pool")
	}
	if pool.mempool != mempool {
		t.Error("mempool not set in pool")
	}
}

// TestNodePoolRegisterNode tests node registration in pool
func TestNodePoolRegisterNode(t *testing.T) {
	bc := &blockchain.BlockChain{}
	mempool := blockchain.NewMempool()
	validators := blockchain.NewValidatorSet()
	blockProposer := blockchain.NewBlockProposer("proposer", bc)

	pool := NewNodePool(bc, mempool, validators, blockProposer)

	nodeID := "test-node"
	region := "us-west"
	stake := uint64(1000000)

	node, err := pool.RegisterNode(nodeID, region, stake)
	if err != nil {
		t.Fatalf("RegisterNode failed: %v", err)
	}

	if node.nodeID != nodeID {
		t.Errorf("expected nodeID %s, got %s", nodeID, node.nodeID)
	}

	// Try to register duplicate
	_, err = pool.RegisterNode(nodeID, region, stake)
	if err == nil {
		t.Error("expected error for duplicate node registration")
	}
}

// TestNodePoolGetNodeState tests node state retrieval
func TestNodePoolGetNodeState(t *testing.T) {
	bc := &blockchain.BlockChain{}
	mempool := blockchain.NewMempool()
	validators := blockchain.NewValidatorSet()
	blockProposer := blockchain.NewBlockProposer("proposer", bc)

	pool := NewNodePool(bc, mempool, validators, blockProposer)

	nodeID := "test-node"
	_, err := pool.RegisterNode(nodeID, "us-west", 1000000)
	if err != nil {
		t.Fatalf("RegisterNode failed: %v", err)
	}

	state, err := pool.GetNodeState(nodeID)
	if err != nil {
		t.Fatalf("GetNodeState failed: %v", err)
	}

	if state.NodeID != nodeID {
		t.Errorf("expected nodeID %s, got %s", nodeID, state.NodeID)
	}

	// Test non-existent node
	_, err = pool.GetNodeState("non-existent")
	if err == nil {
		t.Error("expected error for non-existent node")
	}
}

// TestRewardCalculation tests reward distribution logic
func TestRewardCalculation(t *testing.T) {
	distributor := NewRewardDistributor()

	// Base reward
	reward := distributor.CalculateReward(50.0, 50000)
	if reward < distributor.baseRewardPerRound {
		t.Errorf("expected base reward at least %d, got %d", distributor.baseRewardPerRound, reward)
	}

	// Higher accuracy should increase reward
	rewardHigh := distributor.CalculateReward(90.0, 50000)
	if rewardHigh <= reward {
		t.Errorf("expected higher reward for better accuracy, got %d vs %d", rewardHigh, reward)
	}

	// Higher stake should increase reward
	rewardHighStake := distributor.CalculateReward(50.0, 200000)
	rewardLowStake := distributor.CalculateReward(50.0, 50000)
	if rewardHighStake <= rewardLowStake {
		t.Errorf("expected higher reward for higher stake, got %d vs %d", rewardHighStake, rewardLowStake)
	}
}

// TestStakeManager tests stake tracking
func TestStakeManager(t *testing.T) {
	manager := NewStakeManager()

	nodeID := "test-node"
	stake := uint64(10000)

	manager.AddStake(nodeID, stake)

	retrieved := manager.GetNodeStake(nodeID)
	if retrieved != stake {
		t.Errorf("expected stake %d, got %d", stake, retrieved)
	}

	// Add multiple stakes and verify total
	manager.AddStake("node-2", 5000)
	manager.AddStake("node-3", 15000)

	total := manager.GetTotalStake()
	expected := uint64(30000)
	if total != expected {
		t.Errorf("expected total stake %d, got %d", expected, total)
	}
}

// TestWaitForBlockInclusion tests timeout behavior
func TestWaitForBlockInclusion(t *testing.T) {
	node := NewFLNode("test-node", "us-west")
	bc := blockchain.NewBlockChain()
	mempool := blockchain.NewMempool()

	node.SetupBlockchain(bc, mempool, 10000)

	ctx := context.Background()
	result, _ := node.TrainRound(ctx)
	txID, _ := node.SubmitFlRound(ctx, result, []byte{})

	// Test timeout
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	_, err := node.WaitForBlockInclusion(ctx, txID, 1*time.Second)
	if err == nil {
		t.Error("expected timeout error")
	}
}

// TestGetState tests node state reporting
func TestGetState(t *testing.T) {
	node := NewFLNode("test-node", "us-west")
	bc := &blockchain.BlockChain{}
	mempool := blockchain.NewMempool()

	node.SetupBlockchain(bc, mempool, 10000)

	state := node.GetState()

	if state.NodeID != "test-node" {
		t.Errorf("expected nodeID test-node, got %s", state.NodeID)
	}
	if state.NodeStake != 10000 {
		t.Errorf("expected stake 10000, got %d", state.NodeStake)
	}
	if state.TotalRoundsCompleted != 0 {
		t.Errorf("expected 0 rounds, got %d", state.TotalRoundsCompleted)
	}
}
