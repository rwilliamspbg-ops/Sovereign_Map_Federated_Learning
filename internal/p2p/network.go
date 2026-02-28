// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0

package p2p

import (
	"sync"
	"time"
)

// Network manages peer-to-peer connections for federated learning
type Network struct {
	mu           sync.RWMutex
	nodeID       string
	peers        map[string]*Peer
	verification *VerificationProtocol
}

// Peer represents a connected peer node
type Peer struct {
	ID          string                 `json:"id"`
	Address     string                 `json:"address"`
	Connected   bool                   `json:"connected"`
	LastSeen    time.Time              `json:"last_seen"`
	Reputation  float64                `json:"reputation"`
	UpdateCount int                    `json:"update_count"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewNetwork creates a new P2P network instance
func NewNetwork(nodeID string, minVerifiers int, timeout time.Duration) *Network {
	return &Network{
		nodeID:       nodeID,
		peers:        make(map[string]*Peer),
		verification: NewVerificationProtocol(nodeID, minVerifiers, timeout),
	}
}

// AddPeer registers a new peer
func (n *Network) AddPeer(id, address string, reputation float64) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.peers[id] = &Peer{
		ID:          id,
		Address:     address,
		Connected:   true,
		LastSeen:    time.Now(),
		Reputation:  reputation,
		UpdateCount: 0,
		Metadata:    make(map[string]interface{}),
	}
}

// RemovePeer removes a peer from the network
func (n *Network) RemovePeer(id string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.peers, id)
}

// UpdatePeerLastSeen updates the last seen timestamp for a peer
func (n *Network) UpdatePeerLastSeen(id string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if peer, exists := n.peers[id]; exists {
		peer.LastSeen = time.Now()
	}
}

// GetPeers returns information about all peers
func (n *Network) GetPeers() []map[string]interface{} {
	n.mu.RLock()
	defer n.mu.RUnlock()

	result := make([]map[string]interface{}, 0, len(n.peers))
	for _, peer := range n.peers {
		result = append(result, map[string]interface{}{
			"id":           peer.ID,
			"address":      peer.Address,
			"connected":    peer.Connected,
			"last_seen":    peer.LastSeen,
			"reputation":   peer.Reputation,
			"update_count": peer.UpdateCount,
		})
	}
	return result
}

// GetActivePeerCount returns the number of currently active peers
func (n *Network) GetActivePeerCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	activeCount := 0
	now := time.Now()
	timeout := 5 * time.Minute

	for _, peer := range n.peers {
		if peer.Connected && now.Sub(peer.LastSeen) < timeout {
			activeCount++
		}
	}
	return activeCount
}

// GetPeer returns information about a specific peer
func (n *Network) GetPeer(id string) (*Peer, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	peer, exists := n.peers[id]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent race conditions
	copy := *peer
	return &copy, true
}

// Broadcast sends a message to all connected peers
func (n *Network) Broadcast(message []byte) error {
	n.mu.RLock()
	peers := make([]*Peer, 0, len(n.peers))
	for _, peer := range n.peers {
		if peer.Connected {
			peers = append(peers, peer)
		}
	}
	n.mu.RUnlock()

	// In production, implement actual network communication
	// For now, this is a placeholder for the networking layer
	for _, peer := range peers {
		_ = peer // Would send message to peer
	}

	return nil
}

// SetPeerConnected updates the connection status of a peer
func (n *Network) SetPeerConnected(id string, connected bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if peer, exists := n.peers[id]; exists {
		peer.Connected = connected
		if connected {
			peer.LastSeen = time.Now()
		}
	}
}

// GetNodeID returns this node's ID
func (n *Network) GetNodeID() string {
	return n.nodeID
}

// GetVerificationProtocol returns the verification protocol instance
func (n *Network) GetVerificationProtocol() *VerificationProtocol {
	return n.verification
}
