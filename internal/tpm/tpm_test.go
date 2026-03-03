package tpm

import (
	"fmt"
	"testing"
	"time"
)

func TestNewAttestationManager(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	if manager == nil {
		t.Fatal("expected non-nil attestation manager")
	}
}

func TestGenerateAttestation(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	nodeID := "test-node-001"

	attestation, err := manager.GenerateAttestation(nodeID, []byte("nonce-1"))
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

func TestVerifyAttestation(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	nodeID := "test-node-002"

	attestation, err := manager.GenerateAttestation(nodeID, []byte("nonce-2"))
	if err != nil {
		t.Fatalf("Failed to generate attestation: %v", err)
	}

	valid, err := manager.VerifyAttestation(attestation)
	if err != nil {
		t.Fatalf("Failed to verify attestation: %v", err)
	}

	if !valid {
		t.Error("Expected valid attestation")
	}
}

func TestGetVerifiedQuoteCache(t *testing.T) {
	nodeID := "cache-node"

	q1, err := GetVerifiedQuote(nodeID)
	if err != nil {
		t.Fatalf("first GetVerifiedQuote failed: %v", err)
	}
	q2, err := GetVerifiedQuote(nodeID)
	if err != nil {
		t.Fatalf("second GetVerifiedQuote failed: %v", err)
	}
	if string(q1) != string(q2) {
		t.Fatal("expected cached quote to match initial quote")
	}
}

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
			result, err := VerifyByzantineResilience(tt.totalNodes, tt.faultyNodes)
			if tt.expectSuccess && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expectSuccess {
				t.Errorf("Expected %v, got %v for %d total and %d faulty nodes",
					tt.expectSuccess, result, tt.totalNodes, tt.faultyNodes)
			}
		})
	}
}

func TestPCRIntegrityChecks(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	nodeID := "test-node-004"

	attestation, err := manager.GenerateAttestation(nodeID, []byte("nonce-4"))
	if err != nil {
		t.Fatalf("Failed to generate attestation: %v", err)
	}

	if len(attestation.PCRValues) == 0 {
		t.Fatal("Expected PCR values")
	}

	expectedPCRs := []int{0, 1, 7}
	for _, pcr := range expectedPCRs {
		if _, exists := attestation.PCRValues[pcr]; !exists {
			t.Errorf("Expected PCR %d to be present", pcr)
		}
	}
}

func TestConcurrentAttestations(t *testing.T) {
	manager := NewAttestationManager(100, time.Minute, true)
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			nodeID := fmt.Sprintf("node-%d", id)
			_, err := manager.GenerateAttestation(nodeID, []byte("nonce"))
			if err != nil {
				t.Errorf("Failed concurrent attestation: %v", err)
			}
			done <- true
		}(i)
	}

	for i := 0; i < concurrency; i++ {
		<-done
	}
}

func TestAttestationReportStructure(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	attestation, err := manager.GenerateAttestation("test-node-006", []byte("nonce-6"))
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

func TestVerifyRejectsReplay(t *testing.T) {
	nodeID := "replay-node"
	quote, err := generateTPMQuoteForNode(nodeID)
	if err != nil {
		t.Fatalf("failed to generate quote: %v", err)
	}

	if err := Verify(nodeID, quote); err != nil {
		t.Fatalf("first verify failed unexpectedly: %v", err)
	}

	if err := Verify(nodeID, quote); err == nil {
		t.Fatal("expected replay verify to fail")
	}
}

func TestGenerateAttestationRejectsEmptyNodeID(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	if _, err := manager.GenerateAttestation("", []byte("nonce")); err == nil {
		t.Fatal("expected empty node ID to fail")
	}
}

func TestVerifyAttestationRejectsNil(t *testing.T) {
	manager := NewAttestationManager(10, time.Minute, true)
	if _, err := manager.VerifyAttestation(nil); err == nil {
		t.Fatal("expected nil attestation to fail")
	}
}
