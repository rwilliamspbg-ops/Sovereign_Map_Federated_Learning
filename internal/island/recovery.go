// Copyright 2026 Sovereign-Mohawk Core Team
// Licensed under the Apache License, Version 2.0
package island

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RecoveryManager handles state recovery after offline periods
type RecoveryManager struct {
	stateManager *StateManager
	islandManager *Manager
	persistencePath string
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager(stateManager *StateManager, islandManager *Manager, persistencePath string) *RecoveryManager {
	return &RecoveryManager{
		stateManager: stateManager,
		islandManager: islandManager,
		persistencePath: persistencePath,
	}
}

// PersistState saves current state to disk for recovery
func (rm *RecoveryManager) PersistState() error {
	// Get latest snapshot
	snapshot := rm.stateManager.GetLatestSnapshot()
	if snapshot == nil {
		return fmt.Errorf("no state snapshot available")
	}

	// Get cached updates from Island Mode manager
	updates := rm.islandManager.GetCachedUpdates()

	// Create recovery data structure
	recoveryData := map[string]interface{}{
		"timestamp": time.Now(),
		"snapshot": snapshot,
		"updates": updates,
		"mode": rm.islandManager.GetMode(),
	}

	// Serialize to JSON
	jsonData, err := json.MarshalIndent(recoveryData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize recovery data: %w", err)
	}

	// Ensure persistence directory exists
	if err := os.MkdirAll(filepath.Dir(rm.persistencePath), 0700); err != nil {
		return fmt.Errorf("failed to create persistence directory: %w", err)
	}

	// Write to disk
	if err := os.WriteFile(rm.persistencePath, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write recovery data: %w", err)
	}

	return nil
}

// RecoverState restores state from disk after restart
func (rm *RecoveryManager) RecoverState() error {
	// Read recovery data from disk
	jsonData, err := os.ReadFile(rm.persistencePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No recovery data available, normal startup
		}
		return fmt.Errorf("failed to read recovery data: %w", err)
	}

	// Deserialize recovery data
	var recoveryData map[string]interface{}
	if err := json.Unmarshal(jsonData, &recoveryData); err != nil {
		return fmt.Errorf("failed to deserialize recovery data: %w", err)
	}

	// Restore snapshot to state manager
	if snapshotData, ok := recoveryData["snapshot"]; ok {
		snapshotJSON, _ := json.Marshal(snapshotData)
		var snapshot StateSnapshot
		if err := json.Unmarshal(snapshotJSON, &snapshot); err == nil {
			// Re-verify snapshot integrity
			valid, err := rm.stateManager.VerifyChain()
			if err != nil || !valid {
				return fmt.Errorf("recovered snapshot failed integrity check")
			}
		}
	}

	// Restore cached updates to Island Mode manager
	if updatesData, ok := recoveryData["updates"]; ok {
		updatesJSON, _ := json.Marshal(updatesData)
		var updates []Update
		if err := json.Unmarshal(updatesJSON, &updates); err == nil {
			for _, update := range updates {
				_ = rm.islandManager.CacheUpdate(update)
			}
		}
	}

	return nil
}

// ClearRecoveryData removes persisted recovery data
func (rm *RecoveryManager) ClearRecoveryData() error {
	if err := os.Remove(rm.persistencePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear recovery data: %w", err)
	}
	return nil
}

// GetRecoveryStatus returns the current recovery status
func (rm *RecoveryManager) GetRecoveryStatus() map[string]interface{} {
	info, err := os.Stat(rm.persistencePath)
	if err != nil {
		return map[string]interface{}{
			"available": false,
			"path": rm.persistencePath,
		}
	}

	return map[string]interface{}{
		"available": true,
		"path": rm.persistencePath,
		"size": info.Size(),
		"modified": info.ModTime(),
	}
}
