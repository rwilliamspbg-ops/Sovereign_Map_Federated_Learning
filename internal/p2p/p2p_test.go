package p2p

import (
	"testing"
	"time"
)

// TestNewVerificationManager tests P2P verification manager creation
func TestNewVerificationManager(t *testing.T) {
	manager := NewVerificationManager()
	if manager == nil {
		t.Fatal("Expected non-nil verification manager")
	}
}

// TestChallengeGeneration tests cryptographic challenge generation
func TestChallengeGeneration(t *testing.T) {
	manager := NewVerificationManager()
	nodeID := "test-node-001"
	
	challenge, err := manager.GenerateChallenge(nodeID)
	if err != nil {
		t.Fatalf("Failed to generate challenge: %v", err)
	}
	
	if challenge.Nonce == "" {
		t.Error("Expected non-empty nonce")
	}
	
	if challenge.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
	
	if challenge.TargetNode != nodeID {
		t.Errorf("Expected target node %s, got %s", nodeID, challenge.TargetNode)
	}
}

// TestChallengeResponse tests challenge-response verification
func TestChallengeResponse(t *testing.T) {
	manager := NewVerificationManager()
	nodeID := "test-node-002"
	
	// Generate challenge
	challenge, err := manager.GenerateChallenge(nodeID)
	if err != nil {
		t.Fatalf("Failed to generate challenge: %v", err)
	}
	
	// Create response
	response, err := manager.CreateResponse(challenge)
	if err != nil {
		t.Fatalf("Failed to create response: %v", err)
	}
	
	if response.ChallengeID != challenge.ID {
		t.Error("Response challenge ID mismatch")
	}
	
	if response.Signature == nil {
		t.Error("Expected non-nil signature")
	}
}

// TestSignatureVerification tests signature validation
func TestSignatureVerification(t *testing.T) {
	manager := NewVerificationManager()
	nodeID := "test-node-003"
	
	challenge, err := manager.GenerateChallenge(nodeID)
	if err != nil {
		t.Fatalf("Failed to generate challenge: %v", err)
	}
	
	response, err := manager.CreateResponse(challenge)
	if err != nil {
		t.Fatalf("Failed to create response: %v", err)
	}
	
	// Verify signature
	valid, err := manager.VerifyResponse(response)
	if err != nil {
		t.Fatalf("Failed to verify response: %v", err)
	}
	
	if !valid {
		t.Error("Expected valid response signature")
	}
}

// TestReputationScoring tests distributed reputation system
func TestReputationScoring(t *testing.T) {
	manager := NewVerificationManager()
	nodeID := "test-node-004"
	
	// Initial reputation
	initialRep := manager.GetReputation(nodeID)
	if initialRep != 1.0 {
		t.Errorf("Expected initial reputation 1.0, got %f", initialRep)
	}
	
	// Successful verification should increase reputation
	err := manager.UpdateReputation(nodeID, true)
	if err != nil {
		t.Fatalf("Failed to update reputation: %v", err)
	}
	
	newRep := manager.GetReputation(nodeID)
	if newRep <= initialRep {
		t.Error("Expected reputation to increase")
	}
}

// TestReputationDecay tests time-based reputation decay (λ = 0.1/hour)
func TestReputationDecay(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping reputation decay test in short mode")
	}
	
	manager := NewVerificationManager()
	nodeID := "test-node-005"
	
	// Set high reputation
	manager.SetReputation(nodeID, 1.5)
	initialRep := manager.GetReputation(nodeID)
	
	// Wait for decay
	time.Sleep(2 * time.Second)
	
	decayedRep := manager.GetReputation(nodeID)
	if decayedRep >= initialRep {
		t.Error("Expected reputation to decay over time")
	}
}

// TestNodeBlacklisting tests reputation threshold blacklisting
func TestNodeBlacklisting(t *testing.T) {
	manager := NewVerificationManager()
	nodeID := "test-node-006"
	
	// Multiple failed verifications
	for i := 0; i < 5; i++ {
		err := manager.UpdateReputation(nodeID, false)
		if err != nil {
			t.Fatalf("Failed to update reputation: %v", err)
		}
	}
	
	// Check if blacklisted
	if !manager.IsBlacklisted(nodeID) {
		t.Error("Expected node to be blacklisted after multiple failures")
	}
	
	// Blacklisted nodes should fail verification
	challenge, _ := manager.GenerateChallenge(nodeID)
	response, _ := manager.CreateResponse(challenge)
	valid, err := manager.VerifyResponse(response)
	if valid || err == nil {
		t.Error("Expected verification to fail for blacklisted node")
	}
}

// TestConsensusIntegration tests integration with consensus coordinator
func TestConsensusIntegration(t *testing.T) {
	manager := NewVerificationManager()
	
	// Set up multiple nodes
	nodes := []string{"node-1", "node-2", "node-3", "node-4"}
	for _, node := range nodes {
		manager.RegisterNode(node)
	}
	
	// Verify quorum requirements
	quorum := manager.CalculateQuorum(len(nodes))
	expectedQuorum := (2*len(nodes))/3 + 1
	if quorum != expectedQuorum {
		t.Errorf("Expected quorum %d, got %d", expectedQuorum, quorum)
	}
}

// TestConcurrentVerifications tests thread-safe operations
func TestConcurrentVerifications(t *testing.T) {
	manager := NewVerificationManager()
	concurrency := 10
	done := make(chan bool, concurrency)
	
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			nodeID := "node-" + string(rune('0'+id))
			challenge, err := manager.GenerateChallenge(nodeID)
			if err != nil {
				t.Errorf("Failed to generate challenge: %v", err)
			}
			response, err := manager.CreateResponse(challenge)
			if err != nil {
				t.Errorf("Failed to create response: %v", err)
			}
			_, err = manager.VerifyResponse(response)
			if err != nil {
				t.Errorf("Failed to verify response: %v", err)
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < concurrency; i++ {
		<-done
	}
}

// TestReputationPersistence tests reputation storage and retrieval
func TestReputationPersistence(t *testing.T) {
	manager := NewVerificationManager()
	nodeID := "test-node-007"
	
	// Set reputation
	targetRep := 1.25
	manager.SetReputation(nodeID, targetRep)
	
	// Persist
	err := manager.PersistReputations()
	if err != nil {
		t.Fatalf("Failed to persist reputations: %v", err)
	}
	
	// Create new manager and load
	newManager := NewVerificationManager()
	err = newManager.LoadReputations()
	if err != nil {
		t.Fatalf("Failed to load reputations: %v", err)
	}
	
	// Verify loaded reputation
	loadedRep := newManager.GetReputation(nodeID)
	if loadedRep != targetRep {
		t.Errorf("Expected reputation %f, got %f", targetRep, loadedRep)
	}
}
