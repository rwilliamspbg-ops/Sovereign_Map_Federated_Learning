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

// Package p2p implements peer-to-peer verification protocols
// for coordinatorless federated learning

package p2p

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// PeerInfo represents information about a peer node
type PeerInfo struct {
	ID            string
	Address       string
	PublicKey     []byte
	TPMAttestation []byte
	LastSeen      time.Time
	Reputation    float64
}

// VerificationRequest represents a request to verify model updates
type VerificationRequest struct {
	RequestID     string
	ModelWeights  []byte
	Proof         []byte
	ProposerID    string
	Round         int
	Timestamp     time.Time
}

// VerificationResponse represents a peer's verification result
type VerificationResponse struct {
	RequestID     string
	VerifierID    string
	Valid         bool
	Signature     []byte
	Timestamp     time.Time
	ReasonCode    string
}

// Verifier handles peer-to-peer verification of model updates
type Verifier struct {
	mu               sync.RWMutex
	nodeID           string
	peers            map[string]*PeerInfo
	verifications    map[string][]*VerificationResponse
	minVerifications int
	timeout          time.Duration
}

// NewVerifier creates a new P2P verifier
func NewVerifier(nodeID string, minVerifications int, timeout time.Duration) *Verifier {
	return &Verifier{
		nodeID:           nodeID,
		peers:            make(map[string]*PeerInfo),
		verifications:    make(map[string][]*VerificationResponse),
		minVerifications: minVerifications,
		timeout:          timeout,
	}
}

// RegisterPeer adds a new peer to the verification network
func (v *Verifier) RegisterPeer(peer *PeerInfo) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if peer.ID == "" {
		return fmt.Errorf("peer ID cannot be empty")
	}
	
	// Initialize reputation score
	if peer.Reputation == 0 {
		peer.Reputation = 1.0
	}
	
	peer.LastSeen = time.Now()
	v.peers[peer.ID] = peer
	
	return nil
}

// RemovePeer removes a peer from the verification network
func (v *Verifier) RemovePeer(peerID string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	delete(v.peers, peerID)
}

// RequestVerification broadcasts a verification request to peers
func (v *Verifier) RequestVerification(ctx context.Context, req *VerificationRequest) (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	if req.RequestID == "" {
		req.RequestID = v.generateRequestID(req)
	}
	
	v.verifications[req.RequestID] = make([]*VerificationResponse, 0)
	
	return req.RequestID, nil
}

// SubmitVerification records a verification response from a peer
func (v *Verifier) SubmitVerification(ctx context.Context, resp *VerificationResponse) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	
	// Verify the peer exists
	peer, exists := v.peers[resp.VerifierID]
	if !exists {
		return fmt.Errorf("unknown verifier: %s", resp.VerifierID)
	}
	
	// Check if request exists
	if _, exists := v.verifications[resp.RequestID]; !exists {
		return fmt.Errorf("unknown request: %s", resp.RequestID)
	}
	
	// Record verification
	v.verifications[resp.RequestID] = append(v.verifications[resp.RequestID], resp)
	
	// Update peer reputation based on response
	v.updateReputation(peer, resp.Valid)
	
	return nil
}

// CheckVerificationStatus checks if sufficient verifications have been received
func (v *Verifier) CheckVerificationStatus(requestID string) (bool, float64, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	responses, exists := v.verifications[requestID]
	if !exists {
		return false, 0, fmt.Errorf("request not found: %s", requestID)
	}
	
	if len(responses) < v.minVerifications {
		return false, 0, nil
	}
	
	// Calculate weighted verification score based on peer reputation
	totalWeight := 0.0
	validWeight := 0.0
	
	for _, resp := range responses {
		if peer, exists := v.peers[resp.VerifierID]; exists {
			totalWeight += peer.Reputation
			if resp.Valid {
				validWeight += peer.Reputation
			}
		}
	}
	
	if totalWeight == 0 {
		return false, 0, fmt.Errorf("no valid verifiers")
	}
	
	confidenceScore := validWeight / totalWeight
	
	// Require >66% confidence for Byzantine fault tolerance
	return confidenceScore > 0.66, confidenceScore, nil
}

// GetActivePeers returns list of active peers
func (v *Verifier) GetActivePeers() []*PeerInfo {
	v.mu.RLock()
	defer v.mu.RUnlock()
	
	activeTimeout := 5 * time.Minute
	active := make([]*PeerInfo, 0)
	
	for _, peer := range v.peers {
		if time.Since(peer.LastSeen) < activeTimeout {
			active = append(active, peer)
		}
	}
	
	return active
}

// updateReputation adjusts peer reputation based on verification behavior
func (v *Verifier) updateReputation(peer *PeerInfo, valid bool) {
	if valid {
		peer.Reputation = min(peer.Reputation+0.1, 2.0)
	} else {
		peer.Reputation = max(peer.Reputation-0.2, 0.1)
	}
}

// generateRequestID creates a unique ID for verification requests
func (v *Verifier) generateRequestID(req *VerificationRequest) string {
	data := fmt.Sprintf("%s-%d-%d", req.ProposerID, req.Round, req.Timestamp.Unix())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
