// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

// StateEntry represents a key-value state entry
type StateEntry struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Version     uint64      `json:"version"`
	LastUpdated int64       `json:"last_updated"`
}

// StateSnapshot represents state at a block height
type StateSnapshot struct {
	BlockHeight uint64 `json:"block_height"`
	StateRoot   string `json:"state_root"`
	Timestamp   int64  `json:"timestamp"`
}

// StateDatabase manages on-chain state with Merkle tree verification
type StateDatabase struct {
	mu      sync.RWMutex
	state   map[string]StateEntry
	root    string // Merkle root hash
	version uint64
	history []StateSnapshot
}

// NewStateDatabase creates a new state database
func NewStateDatabase() *StateDatabase {
	return &StateDatabase{
		state:   make(map[string]StateEntry),
		root:    computeInitialRoot(),
		version: 0,
		history: make([]StateSnapshot, 0),
	}
}

// Get retrieves a value from state
func (s *StateDatabase) Get(key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.state[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return entry.Value, nil
}

// Set updates a value in state
func (s *StateDatabase) Set(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state[key] = StateEntry{
		Key:         key,
		Value:       value,
		Version:     s.version,
		LastUpdated: time.Now().Unix(),
	}

	// Rebuild Merkle root
	s.root = s.computeMerkleRoot()
	return nil
}

// Exists checks if a key exists in state
func (s *StateDatabase) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.state[key]
	return exists
}

// Delete removes a key from state
func (s *StateDatabase) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.state[key]; !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	delete(s.state, key)
	s.root = s.computeMerkleRoot()
	return nil
}

// ComputeRoot returns the current state root hash
func (s *StateDatabase) ComputeRoot() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.root
}

// computeMerkleRoot computes the Merkle root of all state entries
func (s *StateDatabase) computeMerkleRoot() string {
	if len(s.state) == 0 {
		return hashData("")
	}

	// Sort keys for deterministic hashing
	keys := make([]string, 0, len(s.state))
	for key := range s.state {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Create hash of each entry
	hashes := make([][]byte, len(keys))
	for i, key := range keys {
		entry := s.state[key]
		entryBytes, _ := json.Marshal(entry)
		hash := sha256.Sum256(entryBytes)
		hashes[i] = hash[:]
	}

	// Build Merkle tree
	for len(hashes) > 1 {
		if len(hashes)%2 != 0 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}
		newHashes := make([][]byte, 0)
		for i := 0; i < len(hashes); i += 2 {
			combined := append(hashes[i], hashes[i+1]...)
			hash := sha256.Sum256(combined)
			newHashes = append(newHashes, hash[:])
		}
		hashes = newHashes
	}

	return hex.EncodeToString(hashes[0])
}

// RecordSnapshot records state at a specific block height
func (s *StateDatabase) RecordSnapshot(blockHeight uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot := StateSnapshot{
		BlockHeight: blockHeight,
		StateRoot:   s.root,
		Timestamp:   time.Now().Unix(),
	}
	s.history = append(s.history, snapshot)
}

// GetSnapshot retrieves state root at a specific block height
func (s *StateDatabase) GetSnapshot(blockHeight uint64) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, snapshot := range s.history {
		if snapshot.BlockHeight == blockHeight {
			return snapshot.StateRoot, nil
		}
	}
	return "", fmt.Errorf("snapshot not found for block height %d", blockHeight)
}

// Clone creates a copy of the state database
func (s *StateDatabase) Clone() *StateDatabase {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clone := NewStateDatabase()
	for key, entry := range s.state {
		clone.state[key] = entry
	}
	clone.root = s.root
	clone.version = s.version
	return clone
}

// GetAll returns all state entries
func (s *StateDatabase) GetAll() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]interface{})
	for key, entry := range s.state {
		result[key] = entry.Value
	}
	return result
}

// IncrementVersion increments the state version
func (s *StateDatabase) IncrementVersion() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.version++
}

func computeInitialRoot() string {
	return hashData("")
}
