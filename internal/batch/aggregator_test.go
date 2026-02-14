package batch

import (
	"testing"
	"time"
)

func TestNewAggregator(t *testing.T) {
	tests := []struct {
		name          string
		batchSize     int
		batchTimeout  time.Duration
		expectError   bool
	}{
		{
			name:          "Valid aggregator",
			batchSize:     10,
			batchTimeout:  5 * time.Second,
			expectError:   false,
		},
		{
			name:          "Small batch size",
			batchSize:     1,
			batchTimeout:  1 * time.Second,
			expectError:   false,
		},
		{
			name:          "Large batch size",
			batchSize:     1000,
			batchTimeout:  30 * time.Second,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := NewAggregator(tt.batchSize, tt.batchTimeout)
			if agg == nil {
				t.Fatal("Expected non-nil aggregator")
			}
			if agg.batchSize != tt.batchSize {
				t.Errorf("Expected batch size %d, got %d", tt.batchSize, agg.batchSize)
			}
			if agg.batchTimeout != tt.batchTimeout {
				t.Errorf("Expected timeout %v, got %v", tt.batchTimeout, agg.batchTimeout)
			}
		})
	}
}

func TestAddUpdate(t *testing.T) {
	agg := NewAggregator(5, 10*time.Second)
	if agg == nil {
		t.Fatal("Failed to create aggregator")
	}

	// Test adding single update
	update := ModelUpdate{
		NodeID:    "node1",
		Weights:   []float64{1.0, 2.0, 3.0},
		Timestamp: time.Now(),
	}

	err := agg.AddUpdate(update)
	if err != nil {
		t.Fatalf("Failed to add update: %v", err)
	}

	// Verify update was added
	agg.mu.RLock()
	count := len(agg.pendingUpdates)
	agg.mu.RUnlock()

	if count != 1 {
		t.Errorf("Expected 1 pending update, got %d", count)
	}
}

func TestAddMultipleUpdates(t *testing.T) {
	agg := NewAggregator(10, 10*time.Second)

	// Add multiple updates from different nodes
	nodeIDs := []string{"node1", "node2", "node3"}
	for _, nodeID := range nodeIDs {
		update := ModelUpdate{
			NodeID:    nodeID,
			Weights:   []float64{1.0, 2.0, 3.0},
			Timestamp: time.Now(),
		}
		err := agg.AddUpdate(update)
		if err != nil {
			t.Fatalf("Failed to add update from %s: %v", nodeID, err)
		}
	}

	agg.mu.RLock()
	count := len(agg.pendingUpdates)
	agg.mu.RUnlock()

	if count != 3 {
		t.Errorf("Expected 3 pending updates, got %d", count)
	}
}

func TestAggregate(t *testing.T) {
	agg := NewAggregator(3, 10*time.Second)

	// Add updates with known weights
	updates := []ModelUpdate{
		{
			NodeID:    "node1",
			Weights:   []float64{1.0, 2.0, 3.0},
			Timestamp: time.Now(),
		},
		{
			NodeID:    "node2",
			Weights:   []float64{2.0, 3.0, 4.0},
			Timestamp: time.Now(),
		},
		{
			NodeID:    "node3",
			Weights:   []float64{3.0, 4.0, 5.0},
			Timestamp: time.Now(),
		},
	}

	for _, update := range updates {
		err := agg.AddUpdate(update)
		if err != nil {
			t.Fatalf("Failed to add update: %v", err)
		}
	}

	// Trigger aggregation
	result := agg.Aggregate()
	if result == nil {
		t.Fatal("Expected aggregation result, got nil")
	}

	// Verify averaged weights (should be [2.0, 3.0, 4.0])
	expected := []float64{2.0, 3.0, 4.0}
	if len(result.Weights) != len(expected) {
		t.Fatalf("Expected %d weights, got %d", len(expected), len(result.Weights))
	}

	for i, exp := range expected {
		if result.Weights[i] != exp {
			t.Errorf("Weight[%d]: expected %.1f, got %.1f", i, exp, result.Weights[i])
		}
	}

	// Verify pending updates were cleared
	agg.mu.RLock()
	count := len(agg.pendingUpdates)
	agg.mu.RUnlock()

	if count != 0 {
		t.Errorf("Expected 0 pending updates after aggregation, got %d", count)
	}
}

func TestGetStatus(t *testing.T) {
	agg := NewAggregator(5, 10*time.Second)

	// Add some updates
	for i := 0; i < 3; i++ {
		update := ModelUpdate{
			NodeID:    "node" + string(rune(i)),
			Weights:   []float64{1.0, 2.0, 3.0},
			Timestamp: time.Now(),
		}
		agg.AddUpdate(update)
	}

	status := agg.GetStatus()
	if status == nil {
		t.Fatal("Expected non-nil status")
	}

	// Verify status fields
	if pending, ok := status["pending_updates"].(int); !ok || pending != 3 {
		t.Errorf("Expected pending_updates=3, got %v", status["pending_updates"])
	}

	if batchSize, ok := status["batch_size"].(int); !ok || batchSize != 5 {
		t.Errorf("Expected batch_size=5, got %v", status["batch_size"])
	}
}

func TestConcurrentAddUpdate(t *testing.T) {
	agg := NewAggregator(100, 10*time.Second)
	concurrency := 10
	updatesPerGoroutine := 10

	done := make(chan bool, concurrency)

	// Spawn multiple goroutines adding updates concurrently
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			for j := 0; j < updatesPerGoroutine; j++ {
				update := ModelUpdate{
					NodeID:    "node" + string(rune(id*100+j)),
					Weights:   []float64{1.0, 2.0, 3.0},
					Timestamp: time.Now(),
				}
				agg.AddUpdate(update)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}

	agg.mu.RLock()
	count := len(agg.pendingUpdates)
	agg.mu.RUnlock()

	expectedCount := concurrency * updatesPerGoroutine
	if count != expectedCount {
		t.Errorf("Expected %d pending updates, got %d", expectedCount, count)
	}
}

func TestEmptyAggregation(t *testing.T) {
	agg := NewAggregator(5, 10*time.Second)

	// Try to aggregate with no updates
	result := agg.Aggregate()
	if result != nil {
		t.Error("Expected nil result for empty aggregation")
	}
}

func BenchmarkAddUpdate(b *testing.B) {
	agg := NewAggregator(1000, 10*time.Second)
	update := ModelUpdate{
		NodeID:    "benchNode",
		Weights:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agg.AddUpdate(update)
	}
}

func BenchmarkAggregate(b *testing.B) {
	agg := NewAggregator(100, 10*time.Second)

	// Pre-populate with updates
	for i := 0; i < 100; i++ {
		update := ModelUpdate{
			NodeID:    "node" + string(rune(i)),
			Weights:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			Timestamp: time.Now(),
		}
		agg.AddUpdate(update)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		agg.Aggregate()
		// Repopulate for next iteration
		for j := 0; j < 100; j++ {
			update := ModelUpdate{
				NodeID:    "node" + string(rune(j)),
				Weights:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
				Timestamp: time.Now(),
			}
			agg.AddUpdate(update)
		}
	}
}
