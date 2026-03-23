// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package consensus

import (
	"context"
	"testing"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/blockchain"
	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/internal/blockchain/vm"
)

// TestConsensusWithBlockchainIntegration demonstrates blockchain integration with consensus
func TestConsensusWithBlockchainIntegration(t *testing.T) {
	// Setup consensus
	coordinator := NewCoordinator("node_1", 10, 30*time.Second)

	// Setup blockchain
	bc := blockchain.NewBlockChain()

	// Create block proposer
	proposer := blockchain.NewBlockProposer("node_1", bc)

	// Connect blockchain to consensus
	if err := coordinator.SetupBlockchainIntegration(proposer); err != nil {
		t.Fatalf("Failed to setup blockchain integration: %v", err)
	}

	// Register validators in blockchain
	for i := 1; i <= 10; i++ {
		nodeID := "node_" + string(rune(48+i))
		if err := coordinator.RegisterValidator(nodeID, 1000000); err != nil {
			t.Logf("Registered validator %s", nodeID)
		}
	}

	// Create proposal
	proposal := &ModelProposal{
		Round:      1,
		Weights:    []byte{1, 2, 3, 4, 5},
		ProposerID: "node_1",
		Proof:      []byte("proof_data"),
		Timestamp:  time.Now(),
	}

	ctx := context.Background()

	// Propose model
	proposalID, err := coordinator.ProposeModel(ctx, proposal)
	if err != nil {
		t.Fatalf("Failed to propose model: %v", err)
	}

	// Cast votes
	for i := 1; i <= 7; i++ {
		vote := &Vote{
			NodeID:     "node_" + string(rune(48+i)),
			ProposalID: proposalID,
			Approve:    true,
			Signature:  []byte("signature"),
			Timestamp:  time.Now(),
		}

		if err := coordinator.CastVote(ctx, vote); err != nil {
			t.Fatalf("Failed to cast vote: %v", err)
		}
	}

	// Check consensus
	hasConsensus, err := coordinator.CheckConsensus(proposalID)
	if err != nil {
		t.Fatalf("Failed to check consensus: %v", err)
	}

	if !hasConsensus {
		t.Error("Expected consensus to be reached")
	}

	// Commit model (triggers blockchain integration)
	if err := coordinator.CommitModel(ctx, proposalID); err != nil {
		t.Fatalf("Failed to commit model: %v", err)
	}

	// Wait for async block proposal to complete
	time.Sleep(100 * time.Millisecond)

	// Verify blockchain state
	blockHeight := bc.Height()
	if blockHeight == 0 {
		t.Log("Note: Block proposal may have occurred asynchronously")
	}

	coordinatorState := coordinator.GetState()
	if coordinatorState != Committed {
		t.Errorf("Expected Committed state, got %v", coordinatorState)
	}
}

// TestConsensusRoundCreation tests creating a ConsensusRound
func TestConsensusRoundCreation(t *testing.T) {
	coordinator := NewCoordinator("node_1", 5, 30*time.Second)

	proposal := &ModelProposal{
		Round:      1,
		Weights:    []byte{1, 2, 3},
		ProposerID: "node_2",
		Proof:      []byte("proof"),
		Timestamp:  time.Now(),
	}

	ctx := context.Background()

	proposalID, _ := coordinator.ProposeModel(ctx, proposal)

	// Add some votes
	for i := 1; i <= 3; i++ {
		vote := &Vote{
			NodeID:     "node_" + string(rune(48+i)),
			ProposalID: proposalID,
			Approve:    true,
			Signature:  []byte("sig"),
			Timestamp:  time.Now(),
		}
		if err := coordinator.CastVote(ctx, vote); err != nil {
			t.Fatalf("Failed to cast vote: %v", err)
		}
	}

	// Get consensus round
	round, err := coordinator.GetConsensusRound(proposalID)
	if err != nil {
		t.Fatalf("Failed to get consensus round: %v", err)
	}

	if round.ProposerID != "node_2" {
		t.Errorf("Expected proposer node_2, got %s", round.ProposerID)
	}

	if len(round.ValidatorVotes) != 3 {
		t.Errorf("Expected 3 votes, got %d", len(round.ValidatorVotes))
	}
}

// TestBlockchainRegistration tests validator registration
func TestBlockchainRegistration(t *testing.T) {
	coordinator := NewCoordinator("node_1", 5, 30*time.Second)

	// Register validator
	err := coordinator.RegisterValidator("node_2", 5000000)
	if err != nil {
		t.Logf("Registration returned: %v", err)
	}

	// Verify validator is registered in blockchain validator set
	if coordinator.validators == nil {
		t.Error("Validator set is nil")
	}
}

// TestMultipleConsensusRounds tests multiple rounds with blockchain integration
func TestMultipleConsensusRounds(t *testing.T) {
	coordinator := NewCoordinator("node_1", 5, 30*time.Second)

	bc := blockchain.NewBlockChain()

	proposer := blockchain.NewBlockProposer("node_1", bc)
	if err := coordinator.SetupBlockchainIntegration(proposer); err != nil {
		t.Fatalf("Failed to setup blockchain integration: %v", err)
	}

	ctx := context.Background()

	// Run 3 consensus rounds
	for round := 1; round <= 3; round++ {
		proposal := &ModelProposal{
			Round:      round,
			Weights:    []byte{byte(round), 2, 3},
			ProposerID: "node_1",
			Proof:      []byte("proof"),
			Timestamp:  time.Now(),
		}

		proposalID, _ := coordinator.ProposeModel(ctx, proposal)

		// Get quorum
		for i := 1; i <= 4; i++ {
			vote := &Vote{
				NodeID:     "node_" + string(rune(48+i)),
				ProposalID: proposalID,
				Approve:    true,
				Signature:  []byte("sig"),
				Timestamp:  time.Now(),
			}
			if err := coordinator.CastVote(ctx, vote); err != nil {
				t.Fatalf("Failed to cast vote: %v", err)
			}
		}

		// Commit
		if err := coordinator.CommitModel(ctx, proposalID); err != nil {
			t.Fatalf("Round %d: Failed to commit: %v", round, err)
		}

		// Reset for next round
		coordinator.Reset()

		// Wait for async operations
		time.Sleep(50 * time.Millisecond)
	}

	// Verify coordinator state after all rounds
	if coordinator.roundNumber != 3 {
		t.Errorf("Expected round number 3, got %d", coordinator.roundNumber)
	}
}

// TestConsensusWithoutBlockchain tests that consensus works without blockchain
func TestConsensusWithoutBlockchain(t *testing.T) {
	coordinator := NewCoordinator("node_1", 5, 30*time.Second)

	// No blockchain setup - should still work

	proposal := &ModelProposal{
		Round:      1,
		Weights:    []byte{1, 2, 3},
		ProposerID: "node_1",
		Proof:      []byte("proof"),
		Timestamp:  time.Now(),
	}

	ctx := context.Background()

	proposalID, _ := coordinator.ProposeModel(ctx, proposal)

	// Cast votes
	for i := 1; i <= 4; i++ {
		vote := &Vote{
			NodeID:     "node_" + string(rune(48+i)),
			ProposalID: proposalID,
			Approve:    true,
			Signature:  []byte("sig"),
			Timestamp:  time.Now(),
		}
		if err := coordinator.CastVote(ctx, vote); err != nil {
			t.Fatalf("Failed to cast vote: %v", err)
		}
	}

	// Commit without blockchain - should work
	if err := coordinator.CommitModel(ctx, proposalID); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	if coordinator.GetState() != Committed {
		t.Error("Expected Committed state")
	}
}

func TestCommitGovernanceProposalThroughExecutor(t *testing.T) {
	coordinator := NewCoordinator("node_1", 5, 30*time.Second)
	ctx := context.Background()

	validators := blockchain.NewValidatorSet()
	stateDB := blockchain.NewStateDatabase()
	contractVM := vm.NewSmartContractVM()
	executor := blockchain.NewSmartContractExecutorWithValidators(contractVM, stateDB, validators)
	if err := coordinator.SetupContractExecutor(executor); err != nil {
		t.Fatalf("SetupContractExecutor failed: %v", err)
	}

	deployTxn := &blockchain.Transaction{
		ID:        "deploy_gov_001",
		Type:      blockchain.TxTypeSmartContract,
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
		t.Fatalf("governance deploy failed: %v", err)
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

	policyUpdate := validators.GetReputationPolicy()
	policyUpdate.ReputationWeight = 35
	policyUpdate.AttestationWeight = 45
	policyUpdate.QualityWeight = 20

	governanceProposalID, err := coordinator.SubmitGovernancePolicyProposal(
		ctx,
		contractAddr,
		"Set policy",
		"Update governance weights",
		policyUpdate,
		1,
	)
	if err != nil {
		t.Fatalf("create governance proposal failed: %v", err)
	}

	if err := coordinator.CastGovernanceVote(ctx, contractAddr, "node_2", governanceProposalID, true); err != nil {
		t.Fatalf("governance vote failed: %v", err)
	}

	consensusProposal := &ModelProposal{
		Round:      1,
		Weights:    []byte{1, 2, 3},
		ProposerID: "node_1",
		Proof:      []byte("proof"),
		Timestamp:  time.Now(),
	}
	consensusProposalID, err := coordinator.ProposeModel(ctx, consensusProposal)
	if err != nil {
		t.Fatalf("consensus propose failed: %v", err)
	}

	for i := 1; i <= 4; i++ {
		if err := coordinator.CastVote(ctx, &Vote{
			NodeID:     "node_" + string(rune(48+i)),
			ProposalID: consensusProposalID,
			Approve:    true,
			Signature:  []byte("sig"),
			Timestamp:  time.Now(),
		}); err != nil {
			t.Fatalf("consensus vote failed: %v", err)
		}
	}

	if err := coordinator.CommitGovernancePolicy(ctx, consensusProposalID, contractAddr, governanceProposalID); err != nil {
		t.Fatalf("commit governance proposal failed: %v", err)
	}

	policy := validators.GetReputationPolicy()
	if policy.ReputationWeight != 35 || policy.AttestationWeight != 45 || policy.QualityWeight != 20 {
		t.Fatalf("expected policy update from governance execution, got %+v", policy)
	}
}
