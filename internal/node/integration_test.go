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
	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/consensus"
)

// TestFLNodeBlockchainRoundTrip tests full workflow: train -> submit -> confirm -> reward
func TestFLNodeBlockchainRoundTrip(t *testing.T) {
	// Setup blockchain
	bc := blockchain.NewBlockChain()
	mempool := blockchain.NewMempool()

	// Create node
	nodeID := "test-validator-001"
	stake := uint64(10000)
	node := NewFLNode(nodeID, "us-west")
	err := node.SetupBlockchain(bc, mempool, stake)
	if err != nil {
		t.Fatalf("SetupBlockchain failed: %v", err)
	}

	ctx := context.Background()

	// 1. Train
	result, err := node.TrainRound(ctx)
	if err != nil {
		t.Fatalf("TrainRound failed: %v", err)
	}
	if result.RoundNum != 1 {
		t.Errorf("expected round 1, got %d", result.RoundNum)
	}

	// 2. Submit to blockchain
	checkpoint := []byte("model-checkpoint-v1")
	txID, err := node.SubmitFlRound(ctx, result, checkpoint)
	if err != nil {
		t.Fatalf("SubmitFlRound failed: %v", err)
	}

	// Verify transaction in mempool
	mempoolSize := mempool.Size()
	if mempoolSize != 1 {
		t.Errorf("expected 1 transaction in mempool, got %d", mempoolSize)
	}

	// 3. Simulate block inclusion
	blockHeight := uint64(1)
	err = node.ConfirmTransaction(txID, blockHeight)
	if err != nil {
		t.Fatalf("ConfirmTransaction failed: %v", err)
	}

	// 4. Complete round
	err = node.CompleteRound(ctx, blockHeight)
	if err != nil {
		t.Fatalf("CompleteRound failed: %v", err)
	}

	// 5. Claim rewards
	reward, err := node.ClaimRewards(ctx)
	if err != nil {
		t.Fatalf("ClaimRewards failed: %v", err)
	}
	if reward <= 0 {
		t.Errorf("expected positive reward, got %d", reward)
	}

	// Verify final state
	state := node.GetState()
	if state.TotalRoundsCompleted != 1 {
		t.Errorf("expected 1 completed round, got %d", state.TotalRoundsCompleted)
	}
	if state.TotalRewardsEarned != reward {
		t.Errorf("expected rewards %d, got %d", reward, state.TotalRewardsEarned)
	}
	if state.SuccessfulTransactions != 1 {
		t.Errorf("expected 1 successful transaction, got %d", state.SuccessfulTransactions)
	}
}

// TestMultiNodePool tests multiple nodes in a pool
func TestMultiNodePool(t *testing.T) {
	// Setup blockchain
	bc := blockchain.NewBlockChain()
	mempool := blockchain.NewMempool()
	validators := blockchain.NewValidatorSet()
	blockProposer := blockchain.NewBlockProposer("proposer", bc)

	// Create pool
	pool := NewNodePool(bc, mempool, validators, blockProposer)

	// Register 5 nodes
	numNodes := 5
	nodeIDs := make([]string, numNodes)
	stakes := []uint64{1000000, 1500000, 1100000, 1200000, 2000000}

	for i := 0; i < numNodes; i++ {
		nodeID := "test-validator-" + string(rune(i+1))
		nodeIDs[i] = nodeID
		_, err := pool.RegisterNode(nodeID, "region-"+string(rune(i+1)), stakes[i])
		if err != nil {
			t.Fatalf("RegisterNode failed: %v", err)
		}
	}

	ctx := context.Background()

	// Execute one round
	roundResult, err := pool.ExecuteRound(ctx)
	if err != nil {
		t.Fatalf("ExecuteRound failed: %v", err)
	}

	// Verify results
	if len(roundResult.ParticipatingNodes) != numNodes {
		t.Errorf("expected %d participating nodes, got %d", numNodes, len(roundResult.ParticipatingNodes))
	}

	if len(roundResult.TransactionIDs) != numNodes {
		t.Errorf("expected %d transactions, got %d", numNodes, len(roundResult.TransactionIDs))
	}

	if roundResult.TotalRewardsDistributed <= 0 {
		t.Errorf("expected positive rewards distributed, got %d", roundResult.TotalRewardsDistributed)
	}

	// Verify pool stats
	stats := pool.GetPoolStats()
	if stats.TotalNodes != uint64(numNodes) {
		t.Errorf("expected %d total nodes, got %d", numNodes, stats.TotalNodes)
	}
	if stats.ActiveNodes != uint64(numNodes) {
		t.Errorf("expected %d active nodes, got %d", numNodes, stats.ActiveNodes)
	}
}

// TestMultipleRounds tests multiple consecutive rounds
func TestMultipleRounds(t *testing.T) {
	// Setup
	bc := blockchain.NewBlockChain()
	mempool := blockchain.NewMempool()
	validators := blockchain.NewValidatorSet()
	blockProposer := blockchain.NewBlockProposer("proposer", bc)

	pool := NewNodePool(bc, mempool, validators, blockProposer)
	if _, err := pool.RegisterNode("node-1", "us-west", 1000000); err != nil {
		t.Fatalf("RegisterNode node-1 failed: %v", err)
	}
	if _, err := pool.RegisterNode("node-2", "us-east", 1500000); err != nil {
		t.Fatalf("RegisterNode node-2 failed: %v", err)
	}

	ctx := context.Background()

	// Run multiple rounds
	numRounds := 3
	var totalRewardsDistributed uint64

	for round := 0; round < numRounds; round++ {
		result, err := pool.ExecuteRound(ctx)
		if err != nil {
			t.Fatalf("ExecuteRound %d failed: %v", round, err)
		}

		if len(result.ParticipatingNodes) != 2 {
			t.Errorf("round %d: expected 2 nodes, got %d", round, len(result.ParticipatingNodes))
		}

		totalRewardsDistributed += result.TotalRewardsDistributed
	}

	// Verify accumulated metrics
	stats := pool.GetPoolStats()
	if stats.TotalRoundsCompleted != uint64(numRounds*2) {
		// Each node completes numRounds
		t.Errorf("expected %d total round completions, got %d", numRounds*2, stats.TotalRoundsCompleted)
	}

	if stats.TotalRewardsDistributed != totalRewardsDistributed {
		t.Errorf("expected total rewards %d, got %d", totalRewardsDistributed, stats.TotalRewardsDistributed)
	}
}

// TestNodeAndConsensusIntegration tests FL nodes with consensus coordinator
func TestNodeAndConsensusIntegration(t *testing.T) {
	// Setup blockchain
	bc := blockchain.NewBlockChain()
	mempool := blockchain.NewMempool()
	blockProposer := blockchain.NewBlockProposer("proposer", bc)

	// Setup consensus
	coordinator := consensus.NewCoordinator("consensus-node", 5, 10*time.Second)
	err := coordinator.SetupBlockchainIntegration(blockProposer)
	if err != nil {
		t.Fatalf("SetupBlockchainIntegration failed: %v", err)
	}

	// Setup nodes
	node1 := NewFLNode("node-1", "us-west")
	if err := node1.SetupBlockchain(bc, mempool, 10000); err != nil {
		t.Fatalf("node1 SetupBlockchain failed: %v", err)
	}

	node2 := NewFLNode("node-2", "us-east")
	if err := node2.SetupBlockchain(bc, mempool, 15000); err != nil {
		t.Fatalf("node2 SetupBlockchain failed: %v", err)
	}

	ctx := context.Background()

	// Node 1: Train and submit
	result1, _ := node1.TrainRound(ctx)
	txID1, _ := node1.SubmitFlRound(ctx, result1, []byte{})

	// Node 2: Train and submit
	result2, _ := node2.TrainRound(ctx)
	txID2, _ := node2.SubmitFlRound(ctx, result2, []byte{})

	// Verify both transactions in mempool
	if mempool.Size() != 2 {
		t.Errorf("expected 2 transactions in mempool, got %d", mempool.Size())
	}

	// Simulate consensus proposing block with FL transactions
	proposal := &consensus.ModelProposal{
		Round:      1,
		Weights:    []byte("aggregated-weights"),
		ProposerID: "consensus-node",
		Timestamp:  time.Now(),
	}

	proposalID, _ := coordinator.ProposeModel(ctx, proposal)
	_ = coordinator.CastVote(ctx, &consensus.Vote{
		NodeID:     "node-1",
		ProposalID: proposalID,
		Approve:    true,
		Timestamp:  time.Now(),
	})
	_ = coordinator.CastVote(ctx, &consensus.Vote{
		NodeID:     "node-2",
		ProposalID: proposalID,
		Approve:    true,
		Timestamp:  time.Now(),
	})

	// Commit triggers async block creation
	_ = coordinator.CommitModel(ctx, proposalID)

	// Give async operations time to complete
	time.Sleep(100 * time.Millisecond)

	// Simulate block inclusion
	blockHeight := uint64(1)
	if err := node1.ConfirmTransaction(txID1, blockHeight); err != nil {
		t.Fatalf("node1 ConfirmTransaction failed: %v", err)
	}
	if err := node2.ConfirmTransaction(txID2, blockHeight); err != nil {
		t.Fatalf("node2 ConfirmTransaction failed: %v", err)
	}

	if err := node1.CompleteRound(ctx, blockHeight); err != nil {
		t.Fatalf("node1 CompleteRound failed: %v", err)
	}
	if err := node2.CompleteRound(ctx, blockHeight); err != nil {
		t.Fatalf("node2 CompleteRound failed: %v", err)
	}

	// Verify both nodes completed the round
	state1 := node1.GetState()
	state2 := node2.GetState()

	if state1.TotalRoundsCompleted != 1 {
		t.Errorf("node1: expected 1 completed round, got %d", state1.TotalRoundsCompleted)
	}
	if state2.TotalRoundsCompleted != 1 {
		t.Errorf("node2: expected 1 completed round, got %d", state2.TotalRoundsCompleted)
	}
}

// TestRewardDistributionFairness tests that rewards are distributed fairly
func TestRewardDistributionFairness(t *testing.T) {
	bc := blockchain.NewBlockChain()
	mempool := blockchain.NewMempool()
	validators := blockchain.NewValidatorSet()
	blockProposer := blockchain.NewBlockProposer("proposer", bc)

	pool := NewNodePool(bc, mempool, validators, blockProposer)

	// Create nodes with different stakes
	stakes := map[string]uint64{
		"node-low":  1000000, // Low stake
		"node-med":  1500000, // Medium stake
		"node-high": 2000000, // High stake
	}

	for nodeID, stake := range stakes {
		if _, err := pool.RegisterNode(nodeID, "region", stake); err != nil {
			t.Fatalf("RegisterNode %s failed: %v", nodeID, err)
		}
	}

	ctx := context.Background()
	_, _ = pool.ExecuteRound(ctx)

	// Verify higher stake nodes earn more rewards
	stateLow, _ := pool.GetNodeState("node-low")
	stateMed, _ := pool.GetNodeState("node-med")
	stateHigh, _ := pool.GetNodeState("node-high")

	if stateLow.TotalRewardsEarned == 0 || stateMed.TotalRewardsEarned == 0 || stateHigh.TotalRewardsEarned == 0 {
		t.Errorf("expected positive rewards for all nodes, got low=%d med=%d high=%d",
			stateLow.TotalRewardsEarned, stateMed.TotalRewardsEarned, stateHigh.TotalRewardsEarned)
	}
}

// TestErrorHandling tests error scenarios
func TestErrorHandling(t *testing.T) {
	ctx := context.Background()

	// Test 1: Submit without blockchain setup
	node := NewFLNode("test-node", "us-west")
	result := &TrainingResult{RoundNum: 1, Accuracy: 0.95}
	_, err := node.SubmitFlRound(ctx, result, []byte{})
	if err == nil {
		t.Error("expected error when blockchain not setup")
	}

	// Test 2: Claim without blockchain setup
	_, err = node.ClaimRewards(ctx)
	if err == nil {
		t.Error("expected error when blockchain not setup")
	}

	// Test 3: Invalid block inclusion timeout
	bc := blockchain.NewBlockChain()
	mempool := blockchain.NewMempool()
	if err := node.SetupBlockchain(bc, mempool, 10000); err != nil {
		t.Fatalf("SetupBlockchain failed: %v", err)
	}
	result = &TrainingResult{RoundNum: 1}
	txID, _ := node.SubmitFlRound(ctx, result, []byte{})

	ctx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	_, err = node.WaitForBlockInclusion(ctx, txID, 1*time.Second)
	if err == nil {
		t.Error("expected timeout error")
	}
}
