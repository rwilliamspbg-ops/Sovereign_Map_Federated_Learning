package scheduler

import (
	"fmt"
	"sync"
	"time"
)

// RoundStatus tracks lifecycle state of a round.
type RoundStatus string

const (
	RoundRunning    RoundStatus = "running"
	RoundAggregated RoundStatus = "aggregated"
	RoundCompleted  RoundStatus = "completed"
)

// TrainingRound is the operator-visible view of a federated round.
type TrainingRound struct {
	RoundID          int
	Status           RoundStatus
	StartedAt        time.Time
	CompletedAt      time.Time
	ExpectedUpdates  int
	ReceivedUpdates  int
	ReceivedByNodeID map[string]time.Time
}

// RoundManager manages training round lifecycle and aggregation signals.
type RoundManager struct {
	mu      sync.Mutex
	nextID  int
	rounds  map[int]*TrainingRound
	trigger *AggregationTrigger
}

// NewRoundManager creates a round manager.
func NewRoundManager(trigger *AggregationTrigger) *RoundManager {
	return &RoundManager{
		nextID:  1,
		rounds:  make(map[int]*TrainingRound),
		trigger: trigger,
	}
}

// StartRound initializes and returns a new training round.
func (m *RoundManager) StartRound(expectedUpdates int) TrainingRound {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := m.nextID
	m.nextID++
	if expectedUpdates < 0 {
		expectedUpdates = 0
	}

	r := &TrainingRound{
		RoundID:          id,
		Status:           RoundRunning,
		StartedAt:        time.Now().UTC(),
		ExpectedUpdates:  expectedUpdates,
		ReceivedByNodeID: make(map[string]time.Time),
	}
	m.rounds[id] = r
	return *r
}

// RecordUpdate registers one node update and indicates if aggregation should run.
func (m *RoundManager) RecordUpdate(roundID int, nodeID string, at time.Time) (bool, error) {
	m.mu.Lock()
	r, ok := m.rounds[roundID]
	if !ok {
		m.mu.Unlock()
		return false, fmt.Errorf("round %d not found", roundID)
	}
	if r.Status != RoundRunning {
		m.mu.Unlock()
		return false, fmt.Errorf("round %d is not running", roundID)
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}
	if _, exists := r.ReceivedByNodeID[nodeID]; !exists {
		r.ReceivedByNodeID[nodeID] = at
		r.ReceivedUpdates++
	}
	m.mu.Unlock()

	if m.trigger == nil {
		return false, nil
	}
	shouldAggregate := m.trigger.ObserveUpdate(at)
	if shouldAggregate {
		m.mu.Lock()
		r.Status = RoundAggregated
		m.mu.Unlock()
	}
	return shouldAggregate, nil
}

// CompleteRound marks a round as completed.
func (m *RoundManager) CompleteRound(roundID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, ok := m.rounds[roundID]
	if !ok {
		return fmt.Errorf("round %d not found", roundID)
	}
	r.Status = RoundCompleted
	r.CompletedAt = time.Now().UTC()
	return nil
}

// GetRound returns a snapshot of a round.
func (m *RoundManager) GetRound(roundID int) (TrainingRound, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, ok := m.rounds[roundID]
	if !ok {
		return TrainingRound{}, false
	}
	out := *r
	out.ReceivedByNodeID = make(map[string]time.Time, len(r.ReceivedByNodeID))
	for k, v := range r.ReceivedByNodeID {
		out.ReceivedByNodeID[k] = v
	}
	return out, true
}
