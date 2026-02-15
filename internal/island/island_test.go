package island

import (
	"testing"
	"time"
)

// TestNewIslandManager tests island mode manager creation
func TestNewIslandManager(t *testing.T) {
	manager := NewIslandManager()
	if manager == nil {
		t.Fatal("Expected non-nil island manager")
	}
	
	if manager.GetMode() != ModeOnline {
		t.Error("Expected initial mode to be Online")
	}
}

// TestModeTransition tests mode transitions between online and island
func TestModeTransition(t *testing.T) {
	manager := NewIslandManager()
	
	// Transition to island mode
	err := manager.TransitionToIslandMode()
	if err != nil {
		t.Fatalf("Failed to transition to island mode: %v", err)
	}
	
	if manager.GetMode() != ModeIsland {
		t.Error("Expected mode to be Island")
	}
	
	// Transition back to online mode
	err = manager.TransitionToOnlineMode()
	if err != nil {
		t.Fatalf("Failed to transition to online mode: %v", err)
	}
	
	if manager.GetMode() != ModeOnline {
		t.Error("Expected mode to be Online")
	}
}

// TestStateSnapshot tests tamper-evident state snapshots
func TestStateSnapshot(t *testing.T) {
	manager := NewIslandManager()
	
	// Create state snapshot
	snapshot, err := manager.CreateStateSnapshot()
	if err != nil {
		t.Fatalf("Failed to create state snapshot: %v", err)
	}
	
	if snapshot == nil {
		t.Fatal("Expected non-nil snapshot")
	}
	
	if snapshot.Hash == "" {
		t.Error("Expected non-empty hash")
	}
	
	if snapshot.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

// TestCachedUpdateStorage tests offline update caching
func TestCachedUpdateStorage(t *testing.T) {
	manager := NewIslandManager()
	
	// Transition to island mode
	err := manager.TransitionToIslandMode()
	if err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	
	// Store updates while offline
	for i := 0; i < 5; i++ {
		update := ModelUpdate{
			NodeID:    "test-node",
			Round:     i,
			Timestamp: time.Now(),
		}
		err := manager.CacheUpdate(update)
		if err != nil {
			t.Errorf("Failed to cache update %d: %v", i, err)
		}
	}
	
	cachedCount := manager.GetCachedUpdateCount()
	if cachedCount != 5 {
		t.Errorf("Expected 5 cached updates, got %d", cachedCount)
	}
}

// TestStateRecovery tests state persistence and recovery
func TestStateRecovery(t *testing.T) {
	manager := NewIslandManager()
	
	// Transition to island mode and create state
	err := manager.TransitionToIslandMode()
	if err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	
	// Persist state
	err = manager.PersistState()
	if err != nil {
		t.Fatalf("Failed to persist state: %v", err)
	}
	
	// Create new manager and recover
	newManager := NewIslandManager()
	err = newManager.RecoverState()
	if err != nil {
		t.Fatalf("Failed to recover state: %v", err)
	}
	
	// Verify recovered state
	if newManager.GetMode() != ModeIsland {
		t.Error("Expected recovered mode to be Island")
	}
}

// TestHashChainIntegrity tests tamper-evident hash chain
func TestHashChainIntegrity(t *testing.T) {
	manager := NewIslandManager()
	
	// Create multiple snapshots
	snapshots := make([]StateSnapshot, 3)
	for i := range snapshots {
		time.Sleep(10 * time.Millisecond)
		snapshot, err := manager.CreateStateSnapshot()
		if err != nil {
			t.Fatalf("Failed to create snapshot: %v", err)
		}
		snapshots[i] = *snapshot
	}
	
	// Verify hash chain
	for i := 1; i < len(snapshots); i++ {
		if snapshots[i].PreviousHash != snapshots[i-1].Hash {
			t.Errorf("Hash chain broken at index %d", i)
		}
	}
}

// TestAutomaticSynchronization tests sync when connectivity restored
func TestAutomaticSynchronization(t *testing.T) {
	manager := NewIslandManager()
	
	// Transition to island and cache updates
	err := manager.TransitionToIslandMode()
	if err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	
	update := ModelUpdate{
		NodeID:    "test-node",
		Timestamp: time.Now(),
	}
	err = manager.CacheUpdate(update)
	if err != nil {
		t.Fatalf("Failed to cache update: %v", err)
	}
	
	// Transition back online (should trigger sync)
	err = manager.TransitionToOnlineMode()
	if err != nil {
		t.Fatalf("Failed to transition online: %v", err)
	}
	
	// Verify cached updates were synced
	time.Sleep(100 * time.Millisecond) // Allow sync to complete
	cachedCount := manager.GetCachedUpdateCount()
	if cachedCount != 0 {
		t.Errorf("Expected 0 cached updates after sync, got %d", cachedCount)
	}
}

// TestConnectivityMonitoring tests connectivity checks
func TestConnectivityMonitoring(t *testing.T) {
	manager := NewIslandManager()
	
	// Initially should be connected
	if !manager.IsConnected() {
		t.Error("Expected initial connection")
	}
	
	// Simulate disconnection
	manager.SimulateDisconnection()
	
	if manager.IsConnected() {
		t.Error("Expected disconnection")
	}
	
	// Should auto-transition to island mode
	time.Sleep(200 * time.Millisecond)
	if manager.GetMode() != ModeIsland {
		t.Error("Expected automatic transition to island mode")
	}
}

// TestCacheSizeLimit tests configurable cache limit
func TestCacheSizeLimit(t *testing.T) {
	manager := NewIslandManagerWithConfig(IslandConfig{
		MaxCachedUpdates: 3,
	})
	
	err := manager.TransitionToIslandMode()
	if err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	
	// Try to cache more than limit
	for i := 0; i < 5; i++ {
		update := ModelUpdate{
			NodeID:    "test-node",
			Round:     i,
			Timestamp: time.Now(),
		}
		err := manager.CacheUpdate(update)
		if i >= 3 && err == nil {
			t.Error("Expected error when exceeding cache limit")
		}
	}
	
	cachedCount := manager.GetCachedUpdateCount()
	if cachedCount > 3 {
		t.Errorf("Expected max 3 cached updates, got %d", cachedCount)
	}
}

// TestStateIntegrityChecks tests integrity verification on recovery
func TestStateIntegrityChecks(t *testing.T) {
	manager := NewIslandManager()
	
	err := manager.TransitionToIslandMode()
	if err != nil {
		t.Fatalf("Failed to transition: %v", err)
	}
	
	err = manager.PersistState()
	if err != nil {
		t.Fatalf("Failed to persist: %v", err)
	}
	
	// Verify integrity
	valid, err := manager.VerifyStateIntegrity()
	if err != nil {
		t.Fatalf("Failed to verify integrity: %v", err)
	}
	
	if !valid {
		t.Error("Expected valid state integrity")
	}
}
