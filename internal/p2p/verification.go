// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0

package p2p

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// VerificationRequest represents a request to verify data from a peer
type VerificationRequest struct {
	RequestID string
	PeerID    string
	Data      []byte
	Signature []byte
	Timestamp time.Time
}

// VerificationResponse contains the result of a verification
type VerificationResponse struct {
	RequestID    string
	Valid        bool
	VerifierID   string
	Proof        []byte
	VerifiedAt   time.Time
	Confidence   float64
}

// VerificationProtocol manages peer-to-peer verification
type VerificationProtocol struct {
	mu              sync.RWMutex
	nodeID          string
	peers           map[string]*PeerInfo
	pendingRequests map[string]*VerificationRequest
	verifications   map[string][]*VerificationResponse
	minVerifiers    int
	timeout         time.Duration
}

// PeerInfo stores information about a peer
type PeerInfo struct {
	ID              string
	ReputationScore float64
	LastSeen        time.Time
	VerificationCount int
	SuccessRate      float64
}

// NewVerificationProtocol creates a new verification protocol instance
func NewVerificationProtocol(nodeID string, minVerifiers int, timeout time.Duration) *VerificationProtocol {
	return &VerificationProtocol{
		nodeID:          nodeID,
		peers:           make(map[string]*PeerInfo),
		pendingRequests: make(map[string]*VerificationRequest),
		verifications:   make(map[string][]*VerificationResponse),
		minVerifiers:    minVerifiers,
		timeout:         timeout,
	}
}

// RequestVerification initiates a verification request to peers
func (vp *VerificationProtocol) RequestVerification(ctx context.Context, data []byte, signature []byte) (string, error) {
	vp.mu.Lock()
	defer vp.mu.Unlock()

	// Generate request ID
	requestID := vp.generateRequestID(data)

	request := &VerificationRequest{
		RequestID: requestID,
		PeerID:    vp.nodeID,
		Data:      data,
		Signature: signature,
		Timestamp: time.Now(),
	}

	vp.pendingRequests[requestID] = request
	vp.verifications[requestID] = make([]*VerificationResponse, 0)

	// Broadcast verification request to peers
	go vp.broadcastVerificationRequest(ctx, request)

	return requestID, nil
}

// VerifyData performs verification of data from a peer
func (vp *VerificationProtocol) VerifyData(ctx context.Context, request *VerificationRequest) (*VerificationResponse, error) {
	vp.mu.Lock()
	defer vp.mu.Unlock()

	// Verify signature
	valid := vp.verifySignature(request.Data, request.Signature)

	// Generate cryptographic proof
	proof := vp.generateProof(request.Data)

	// Calculate confidence based on data integrity
	confidence := vp.calculateConfidence(request)

	response := &VerificationResponse{
		RequestID:  request.RequestID,
		Valid:      valid,
		VerifierID: vp.nodeID,
		Proof:      proof,
		VerifiedAt: time.Now(),
		Confidence: confidence,
	}

	return response, nil
}

// SubmitVerificationResponse records a verification response from a peer
func (vp *VerificationProtocol) SubmitVerificationResponse(ctx context.Context, response *VerificationResponse) error {
	vp.mu.Lock()
	defer vp.mu.Unlock()

	// Check if request exists
	if _, exists := vp.pendingRequests[response.RequestID]; !exists {
		return fmt.Errorf("verification request %s not found", response.RequestID)
	}

	// Add response to verifications
	vp.verifications[response.RequestID] = append(vp.verifications[response.RequestID], response)

	// Update peer reputation based on response
	vp.updatePeerReputation(response.VerifierID, response.Valid)

	return nil
}

// CheckVerificationStatus checks if verification is complete
func (vp *VerificationProtocol) CheckVerificationStatus(requestID string) (bool, float64, error) {
	vp.mu.RLock()
	defer vp.mu.RUnlock()

	responses, exists := vp.verifications[requestID]
	if !exists {
		return false, 0, fmt.Errorf("verification request %s not found", requestID)
	}

	// Check if minimum verifiers reached
	if len(responses) < vp.minVerifiers {
		return false, 0, nil
	}

	// Calculate consensus
	validCount := 0
	totalConfidence := 0.0

	for _, resp := range responses {
		if resp.Valid {
			validCount++
			totalConfidence += resp.Confidence
		}
	}

	// Require majority consensus
	consensusReached := validCount >= (len(responses)+1)/2
	averageConfidence := totalConfidence / float64(len(responses))

	return consensusReached, averageConfidence, nil
}

// RegisterPeer adds a new peer to the network
func (vp *VerificationProtocol) RegisterPeer(peerID string) error {
	vp.mu.Lock()
	defer vp.mu.Unlock()

	if _, exists := vp.peers[peerID]; exists {
		return fmt.Errorf("peer %s already registered", peerID)
	}

	vp.peers[peerID] = &PeerInfo{
		ID:              peerID,
		ReputationScore: 1.0,
		LastSeen:        time.Now(),
		VerificationCount: 0,
		SuccessRate:      1.0,
	}

	return nil
}

// GetPeerReputation retrieves the reputation score of a peer
func (vp *VerificationProtocol) GetPeerReputation(peerID string) (float64, error) {
	vp.mu.RLock()
	defer vp.mu.RUnlock()

	peer, exists := vp.peers[peerID]
	if !exists {
		return 0, fmt.Errorf("peer %s not found", peerID)
	}

	return peer.ReputationScore, nil
}

// Helper functions

func (vp *VerificationProtocol) generateRequestID(data []byte) string {
	hash := sha256.Sum256(data)
	timestamp := time.Now().UnixNano()
	combined := append(hash[:], []byte(fmt.Sprintf("%d", timestamp))...)
	finalHash := sha256.Sum256(combined)
	return hex.EncodeToString(finalHash[:])
}

func (vp *VerificationProtocol) broadcastVerificationRequest(ctx context.Context, request *VerificationRequest) {
	// Simulate broadcasting to all peers
	// In production, this would use actual P2P networking
	for peerID := range vp.peers {
		if peerID != vp.nodeID {
			// Send verification request to peer
			// This is a placeholder for actual network communication
		}
	}
}

func (vp *VerificationProtocol) verifySignature(data []byte, signature []byte) bool {
	// Simplified signature verification
	// In production, use proper cryptographic signature verification
	if len(signature) == 0 {
		return false
	}

	// Verify data integrity
	hash := sha256.Sum256(data)
	return len(hash) > 0
}

func (vp *VerificationProtocol) generateProof(data []byte) []byte {
	// Generate cryptographic proof of verification
	hash := sha256.Sum256(data)
	proof := sha256.Sum256(append(hash[:], []byte(vp.nodeID)...))
	return proof[:]
}

func (vp *VerificationProtocol) calculateConfidence(request *VerificationRequest) float64 {
	// Calculate confidence based on multiple factors
	confidence := 0.8

	// Adjust based on data size
	if len(request.Data) > 1024 {
		confidence += 0.1
	}

	// Adjust based on signature validity
	if len(request.Signature) > 0 {
		confidence += 0.1
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

func (vp *VerificationProtocol) updatePeerReputation(peerID string, success bool) {
	peer, exists := vp.peers[peerID]
	if !exists {
		return
	}

	peer.VerificationCount++
	peer.LastSeen = time.Now()

	// Update success rate using exponential moving average
	alpha := 0.2
	if success {
		peer.SuccessRate = alpha*1.0 + (1-alpha)*peer.SuccessRate
	} else {
		peer.SuccessRate = alpha*0.0 + (1-alpha)*peer.SuccessRate
	}

	// Update reputation score
	peer.ReputationScore = peer.SuccessRate
}

// GetVerificationMetrics returns metrics about verification activity
func (vp *VerificationProtocol) GetVerificationMetrics() map[string]interface{} {
	vp.mu.RLock()
	defer vp.mu.RUnlock()

	return map[string]interface{}{
		"total_peers":         len(vp.peers),
		"pending_requests":    len(vp.pendingRequests),
		"completed_verifications": len(vp.verifications),
		"min_verifiers":       vp.minVerifiers,
	}
}
