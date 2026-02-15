// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0

package consensus

import (
	"context"
	"testing"
	"time"
)

// TestCoordinatorCreation tests coordinator initialization
func TestCoordinatorCreation(t *testing.T) {
	totalNodes := 10
	timeout := 5 * time.Second
	
	coord := NewCoordinator("node-1", totalNodes, timeout)
	
	if coord == nil {
		t.Fatal("Failed to create coordinator")
	}
	
	if coord.nodeID != "node-1" {
		t.Errorf("Expected nodeID 'node-1', got '%s'", coord.nodeID)
	}
	
	if coord.totalNodes != totalNodes {
		t.Errorf("Expected %d total nodes, got %d", totalNodes, coord.totalNodes)
	}
	
	// Byzantine fault tolerance: quorum = ⌈(2n/3)⌉ + 1
	expectedQuorum := (2 * totalNodes / 3) + 1
	if coord.quorumSize != expectedQuorum {
		t.Errorf("Expected quorum size %d, got %d", expectedQuorum, coord.quorumSize)
	}
}

// TestProposeModel tests model proposal submission
func TestProposeModel(t *testing.T) {
	coord := NewCoordinator("node-1", 10, 5*time.Second)
	ctx := context.Background()
	
	proposal := &ModelProposal{
		Round:      1,
		Weights:    []byte("model-weights-v1"),
		ProposerID: "node-1",
		Proof:      []byte("zk-proof"),
		Timestamp:  time.Now(),
	}
	
	proposalID, err := coord.ProposeModel(ctx, proposal)
	if err != nil {
		t.Fatalf("Failed to propose model: %v", err)
	}
	
	if proposalID == "" {
		t.Error("Expected non-empty proposal ID")
	}
	
	if coord.GetState() != Voting {
		t.Errorf("Expected state Voting, got %v", coord.GetState())
	}
}

// TestVoting tests the voting mechanism
func TestVoting(t *testing.T) {
	coord := NewCoordinator("node-1", 10, 5*time.Second)
	ctx := context.Background()
	
	proposal := &ModelProposal{
		Round:      1,
		Weights:    []byte("model-weights"),
		ProposerID: "node-1",
		Proof:      []byte("proof"),
		Timestamp:  time.Now(),
	}
	
	proposalID, _ := coord.ProposeModel(ctx, proposal)
	
	// Cast votes from multiple nodes
	for i := 1; i <= 7; i++ {
		vote := &Vote{
			NodeID:     "node-" + string(rune('0'+i)),
			ProposalID: proposalID,
			Approve:    true,
			Signature:  []byte("sig"),
			Timestamp:  time.Now(),
		}
		
		err := coord.CastVote(ctx, vote)
		if err != nil {
			t.Errorf("Failed to cast vote: %v", err)
		}
	}
	
	// Check consensus (quorum = 7 for 10 nodes)
	consensus, err := coord.CheckConsensus(proposalID)
	if err != nil {
		t.Fatalf("Failed to check consensus: %v", err)
	}
	
	if !consensus {
		t.Error("Expected consensus to be reached with 7 votes")
	}
}

// TestByzantineQuorum tests BFT quorum requirements
func TestByzantineQuorum(t *testing.T) {
	tests := []struct {
		name       string
		totalNodes int
		expQuorum  int
	}{
		{"3 nodes", 3, 3},   // (2*3/3)+1 = 3
		{"10 nodes", 10, 7}, // (2*10/3)+1 = 7
		{"100 nodes", 100, 67}, // (2*100/3)+1 = 67
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord := NewCoordinator("test-node", tt.totalNodes, 5*time.Second)
			if coord.quorumSize != tt.expQuorum {
				t.Errorf("For %d nodes, expected quorum %d, got %d",
					tt.totalNodes, tt.expQuorum, coord.quorumSize)
			}
		})
	}
}

// TestConsensusFailure tests consensus failure scenarios
func TestConsensusFailure(t *testing.T) {
	coord := NewCoordinator("node-1", 10, 5*time.Second)
	ctx := context.Background()
	
	proposal := &ModelProposal{
		Round:      1,
		Weights:    []byte("weights"),
		ProposerID: "node-1",
		Proof:      []byte("proof"),
		Timestamp:  time.Now(),
	}
	
	proposalID, _ := coord.ProposeModel(ctx, proposal)
	
	// Only 5 votes (insufficient for quorum of 7)
	for i := 1; i <= 5; i++ {
		vote := &Vote{
			NodeID:     "node-" + string(rune('0'+i)),
			ProposalID: proposalID,
			Approve:    true,
			Signature:  []byte("sig"),
			Timestamp:  time.Now(),
		}
		coord.CastVote(ctx, vote)
	}
	
	consensus, _ := coord.CheckConsensus(proposalID)
	if consensus {
		t.Error("Expected consensus to fail with only 5 votes out of required 7")
	}
	
	// Try to commit without consensus
	err := coord.CommitModel(ctx, proposalID)
	if err == nil {
		t.Error("Expected error when committing without consensus")
	}
	
	if coord.GetState() != Aborted {
		t.Errorf("Expected state Aborted, got %v", coord.GetState())
	}
}

// TestCommitModel tests successful model commitment
func TestCommitModel(t *testing.T) {
	coord := NewCoordinator("node-1", 10, 5*time.Second)
	ctx := context.Background()
	
	proposal := &ModelProposal{
		Round:      1,
		Weights:    []byte("weights"),
		ProposerID: "node-1",
		Proof:      []byte("proof"),
		Timestamp:  time.Now(),
	}
	
	proposalID, _ := coord.ProposeModel(ctx, proposal)
	
	// Cast sufficient votes
	for i := 1; i <= 8; i++ {
		vote := &Vote{
			NodeID:     "node-" + string(rune('0'+i)),
			ProposalID: proposalID,
			Approve:    true,
			Signature:  []byte("sig"),
			Timestamp:  time.Now(),
		}
		coord.CastVote(ctx, vote)
	}
	
	err := coord.CommitModel(ctx, proposalID)
	if err != nil {
		t.Errorf("Failed to commit model: %v", err)
	}
	
	if coord.GetState() != Committed {
		t.Errorf("Expected state Committed, got %v", coord.GetState())
	}
}

// TestCoordinatorReset tests coordinator reset functionality
func TestCoordinatorReset(t *testing.T) {
	coord := NewCoordinator("node-1", 10, 5*time.Second)
	ctx := context.Background()
	
	proposal := &ModelProposal{
		Round:      1,
		Weights:    []byte("weights"),
		ProposerID: "node-1",
		Proof:      []byte("proof"),
		Timestamp:  time.Now(),
	}
	
	coord.ProposeModel(ctx, proposal)
	
	// Reset coordinator
	coord.Reset()
	
	if coord.GetState() != Proposing {
		t.Errorf("Expected state Proposing after reset, got %v", coord.GetState())
	}
	
	if len(coord.proposals) != 0 {
		t.Error("Expected proposals to be cleared after reset")
	}
	
	if len(coord.votes) != 0 {
		t.Error("Expected votes to be cleared after reset")
	}
}

// TestDistributedAggregator tests the aggregator with consensus
func TestDistributedAggregator(t *testing.T) {
	coordinator := NewCoordinator("test-node", 10, 5*time.Second)
	aggregator := NewDistributedAggregator("test-node", []string{"peer1", "peer2", "peer3"}, 30*time.Second, coordinator)
	
	if aggregator == nil {
		t.Fatal("Failed to create distributed aggregator")
	}
	
	// Submit a model
	ctx := context.Background()
	modelWeights := []byte("test-model-weights")
	
	err := aggregator.SubmitModel(ctx, "test-node", "model-1", modelWeights)
	if err != nil {
		t.Errorf("Failed to submit model: %v", err)
	}
	
	// Check metrics
	metrics := aggregator.GetMetrics()
	if metrics == nil {
		t.Error("Expected non-nil metrics")
	}
}
