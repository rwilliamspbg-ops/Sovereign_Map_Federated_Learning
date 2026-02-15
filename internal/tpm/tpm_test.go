package tpm

import (
	"testing"
	"time"
  	"fmt"
)

// TestNewAttestationManager tests the creation of a new attestation manager
func TestNewAttestationManager(t *testing.T) {
	manager := NewAttestationManager()
	if manager == nil {
		t.Fatal("Expected non-nil attestation manager")
	}
}

// TestGenerateAttestation tests attestation generation
func TestGenerateAttestation(t *testing.T) {
	manager := NewAttestationManager()
	nodeID := "test-node-001"
	
	attestation, err := manager.GenerateAttestation(nodeID)
	if err != nil {
		t.Fatalf("Failed to generate attestation: %v", err)
	}
	
	if attestation.NodeID != nodeID {
		t.Errorf("Expected nodeID %s, got %s", nodeID, attestation.NodeID)
	}
	
	if attestation.Quote == nil {
		t.Error("Expected non-nil quote")
	}
	
	if len(attestation.PCRValues) == 0 {
		t.Error("Expected PCR values")
	}
}

// TestVerifyAttestation tests attestation verification
func TestVerifyAttestation(t *testing.T) {
	manager := NewAttestationManager()
	nodeID := "test-node-002"
	
	// Generate attestation
	attestation, err := manager.GenerateAttestation(nodeID)
	if err != nil {
		t.Fatalf("Failed to generate attestation: %v", err)
	}
	
	// Verify attestation
	valid, err := manager.VerifyAttestation(attestation)
	if err != nil {
		t.Fatalf("Failed to verify attestation: %v", err)
	}
	
	if !valid {
		t.Error("Expected valid attestation")
	}
}

// TestAttestationCaching tests quote caching mechanism
func TestAttestationCaching(t *testing.T) {
	manager := NewAttestationManager()
	nodeID := "test-node-003"
	
	// Generate first attestation
	start1 := time.Now()
	_, err := manager.GenerateAttestation(nodeID)
	if err != nil {
		t.Fatalf("Failed to generate first attestation: %v", err)
	}
	duration1 := time.Since(start1)
	
	// Generate second attestation (should use cache)
	start2 := time.Now()
	_, err = manager.GenerateAttestation(nodeID)
	if err != nil {
		t.Fatalf("Failed to generate second attestation: %v", err)
	}
	duration2 := time.Since(start2)
	
	// Cached version should be significantly faster
	if duration2 > duration1/2 {
		t.Log("Warning: Caching may not be working as expected")
	}
}

// TestByzantineVerification tests Byzantine fault tolerance verification
func TestByzantineVerification(t *testing.T) {
	tests := []struct {
		name          string
		totalNodes    int
		faultyNodes   int
		expectSuccess bool
	}{
		{"No faulty nodes", 10, 0, true},
		{"Within tolerance", 10, 3, true},
		{"At boundary", 10, 4, true},
		{"Exceeds tolerance", 10, 5, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyByzantineTolerance(tt.totalNodes, tt.faultyNodes)
			if result != tt.expectSuccess {
				t.Errorf("Expected %v, got %v for %d total and %d faulty nodes",
					tt.expectSuccess, result, tt.totalNodes, tt.faultyNodes)
			}
		})
	}
}

// TestPCRIntegrityChecks tests Platform Configuration Register validation
func TestPCRIntegrityChecks(t *testing.T) {
	manager := NewAttestationManager()
	nodeID := "test-node-004"
	
	attestation, err := manager.GenerateAttestation(nodeID)
	if err != nil {
		t.Fatalf("Failed to generate attestation: %v", err)
	}
	
	// Verify PCR values are present
	if len(attestation.PCRValues) == 0 {
		t.Fatal("Expected PCR values")
	}
	
	// Check specific PCR indices
	expectedPCRs := []int{0, 1, 2, 7} // Common security-critical PCRs
	for _, pcr := range expectedPCRs {
		if _, exists := attestation.PCRValues[pcr]; !exists {
			t.Errorf("Expected PCR %d to be present", pcr)
		}
	}
}

// TestAttestationCacheTTL tests cache expiration
func TestAttestationCacheTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping TTL test in short mode")
	}
	
	manager := NewAttestationManager()
	nodeID := "test-node-005"
	
	// Generate attestation
	attest1, err := manager.GenerateAttestation(nodeID)
	if err != nil {
		t.Fatalf("Failed to generate attestation: %v", err)
	}
	
	// Wait for cache to expire (assuming short TTL for testing)
	time.Sleep(2 * time.Second)
	
	// Generate new attestation
	attest2, err := manager.GenerateAttestation(nodeID)
	if err != nil {
		t.Fatalf("Failed to generate second attestation: %v", err)
	}
	
	// Timestamps should be different after TTL expiration
	if attest1.Timestamp.Equal(attest2.Timestamp) {
		t.Error("Expected different timestamps after cache expiration")
	}
}

// TestConcurrentAttestations tests thread-safety
func TestConcurrentAttestations(t *testing.T) {
	manager := NewAttestationManager()
	concurrency := 10
	done := make(chan bool, concurrency)
	
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			nodeID := fmt.Sprintf("node-%d", id)
			_, err := manager.GenerateAttestation(nodeID)
			if err != nil {
				t.Errorf("Failed concurrent attestation: %v", err)
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < concurrency; i++ {
		<-done
	}
}

// TestAttestationReportStructure tests report data structure
func TestAttestationReportStructure(t *testing.T) {
	manager := NewAttestationManager()
	attestation, err := manager.GenerateAttestation("test-node-006")
	if err != nil {
		t.Fatalf("Failed to generate attestation: %v", err)
	}
	
	if attestation.NodeID == "" {
		t.Error("Expected non-empty NodeID")
	}
	
	if attestation.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
	
	if attestation.Signature == nil {
		t.Error("Expected non-nil signature")
	}
}
