// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package island

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// StateSnapshot represents a tamper-evident snapshot of node state
type StateSnapshot struct {
	Timestamp     time.Time              `json:"timestamp"`
	Round         int                    `json:"round"`
	ModelChecksum string                 `json:"model_checksum"`
	UpdateCount   int                    `json:"update_count"`
	Metadata      map[string]interface{} `json:"metadata"`
	PreviousHash  string                 `json:"previous_hash"`
	Hash          string                 `json:"hash"`
}

// StateManager handles state persistence and recovery
type StateManager struct {
	mu         sync.RWMutex
	snapshots  []StateSnapshot
	maxSnapshots int
	lastSnapshot time.Time
}

// NewStateManager creates a new state manager
func NewStateManager(maxSnapshots int) *StateManager {
	return &StateManager{
		snapshots:   make([]StateSnapshot, 0, maxSnapshots),
		maxSnapshots: maxSnapshots,
		lastSnapshot: time.Now(),
	}
}

// CreateSnapshot creates a tamper-evident state snapshot
func (sm *StateManager) CreateSnapshot(round int, modelChecksum string, updateCount int, metadata map[string]interface{}) (*StateSnapshot, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var previousHash string
	if len(sm.snapshots) > 0 {
		previousHash = sm.snapshots[len(sm.snapshots)-1].Hash
	}

	snapshot := StateSnapshot{
		Timestamp:     time.Now(),
		Round:         round,
		ModelChecksum: modelChecksum,
		UpdateCount:   updateCount,
		Metadata:      metadata,
		PreviousHash:  previousHash,
	}

	// Compute hash for tamper-evidence
	hash, err := sm.computeHash(&snapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to compute hash: %w", err)
	}
	snapshot.Hash = hash

	// Add to snapshot chain
	if len(sm.snapshots) >= sm.maxSnapshots {
		sm.snapshots = sm.snapshots[1:]
	}
	sm.snapshots = append(sm.snapshots, snapshot)
	sm.lastSnapshot = time.Now()

	return &snapshot, nil
}

// GetLatestSnapshot returns the most recent state snapshot
func (sm *StateManager) GetLatestSnapshot() *StateSnapshot {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if len(sm.snapshots) == 0 {
		return nil
	}
	return &sm.snapshots[len(sm.snapshots)-1]
}

// GetSnapshots returns all state snapshots
func (sm *StateManager) GetSnapshots() []StateSnapshot {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	snapshots := make([]StateSnapshot, len(sm.snapshots))
	copy(snapshots, sm.snapshots)
	return snapshots
}

// VerifyChain verifies the integrity of the snapshot chain
func (sm *StateManager) VerifyChain() (bool, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for i, snapshot := range sm.snapshots {
		// Verify hash
		computedHash, err := sm.computeHash(&snapshot)
		if err != nil {
			return false, fmt.Errorf("failed to compute hash for snapshot %d: %w", i, err)
		}

		if computedHash != snapshot.Hash {
			return false, fmt.Errorf("hash mismatch at snapshot %d", i)
		}

		// Verify chain linkage
		if i > 0 {
			if snapshot.PreviousHash != sm.snapshots[i-1].Hash {
				return false, fmt.Errorf("chain broken at snapshot %d", i)
			}
		}
	}

	return true, nil
}

// computeHash computes SHA-256 hash of snapshot for tamper-evidence
func (sm *StateManager) computeHash(snapshot *StateSnapshot) (string, error) {
	// Create a copy without the hash field
	data := map[string]interface{}{
		"timestamp":      snapshot.Timestamp.UnixNano(),
		"round":          snapshot.Round,
		"model_checksum": snapshot.ModelChecksum,
		"update_count":   snapshot.UpdateCount,
		"metadata":       snapshot.Metadata,
		"previous_hash":  snapshot.PreviousHash,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:]), nil
}

// GetTimeSinceLastSnapshot returns duration since last snapshot
func (sm *StateManager) GetTimeSinceLastSnapshot() time.Duration {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return time.Since(sm.lastSnapshot)
}

// ClearSnapshots removes all snapshots (use with caution)
func (sm *StateManager) ClearSnapshots() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.snapshots = make([]StateSnapshot, 0, sm.maxSnapshots)
}
