package scheduler

import (
	"sync"
	"time"
)

// AggregationTrigger defines when model aggregation should run.
type AggregationTrigger struct {
	mu              sync.Mutex
	MinUpdates      int
	MaxWait         time.Duration
	pendingUpdates  int
	lastAggregation time.Time
}

// NewAggregationTrigger creates a trigger with sane defaults.
func NewAggregationTrigger(minUpdates int, maxWait time.Duration) *AggregationTrigger {
	if minUpdates <= 0 {
		minUpdates = 1
	}
	return &AggregationTrigger{
		MinUpdates: minUpdates,
		MaxWait:    maxWait,
	}
}

// ObserveUpdate registers one model update and returns whether to aggregate now.
func (t *AggregationTrigger) ObserveUpdate(now time.Time) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if now.IsZero() {
		now = time.Now().UTC()
	}

	t.pendingUpdates++
	if t.pendingUpdates >= t.MinUpdates {
		t.pendingUpdates = 0
		t.lastAggregation = now
		return true
	}

	if t.MaxWait > 0 && t.pendingUpdates > 0 {
		if t.lastAggregation.IsZero() || now.Sub(t.lastAggregation) >= t.MaxWait {
			t.pendingUpdates = 0
			t.lastAggregation = now
			return true
		}
	}

	return false
}

// PendingUpdates reports updates currently buffered for aggregation.
func (t *AggregationTrigger) PendingUpdates() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.pendingUpdates
}

// LastAggregationAt returns the last aggregation timestamp.
func (t *AggregationTrigger) LastAggregationAt() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.lastAggregation
}
