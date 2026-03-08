package scheduler

import (
	"testing"
	"time"
)

func TestJobQueuePriorityOrdering(t *testing.T) {
	q := NewJobQueue(0)
	if err := q.Enqueue(Job{ID: "low", Priority: PriorityLow, CreatedAt: time.Unix(1, 0)}); err != nil {
		t.Fatalf("enqueue low: %v", err)
	}
	if err := q.Enqueue(Job{ID: "high", Priority: PriorityHigh, CreatedAt: time.Unix(2, 0)}); err != nil {
		t.Fatalf("enqueue high: %v", err)
	}
	if err := q.Enqueue(Job{ID: "normal", Priority: PriorityNormal, CreatedAt: time.Unix(0, 0)}); err != nil {
		t.Fatalf("enqueue normal: %v", err)
	}

	first, ok := q.Dequeue()
	if !ok || first.ID != "high" {
		t.Fatalf("expected high first, got %+v", first)
	}
	second, ok := q.Dequeue()
	if !ok || second.ID != "normal" {
		t.Fatalf("expected normal second, got %+v", second)
	}
	third, ok := q.Dequeue()
	if !ok || third.ID != "low" {
		t.Fatalf("expected low third, got %+v", third)
	}
}

func TestRoundManagerAggregationTrigger(t *testing.T) {
	trigger := NewAggregationTrigger(2, 0)
	mgr := NewRoundManager(trigger)
	r := mgr.StartRound(2)

	agg, err := mgr.RecordUpdate(r.RoundID, "node-a", time.Now().UTC())
	if err != nil {
		t.Fatalf("record first update: %v", err)
	}
	if agg {
		t.Fatalf("unexpected aggregation on first update")
	}

	agg, err = mgr.RecordUpdate(r.RoundID, "node-b", time.Now().UTC())
	if err != nil {
		t.Fatalf("record second update: %v", err)
	}
	if !agg {
		t.Fatalf("expected aggregation on second update")
	}

	snap, ok := mgr.GetRound(r.RoundID)
	if !ok {
		t.Fatalf("round not found")
	}
	if snap.Status != RoundAggregated {
		t.Fatalf("expected aggregated status, got %s", snap.Status)
	}
	if snap.ReceivedUpdates != 2 {
		t.Fatalf("expected 2 updates, got %d", snap.ReceivedUpdates)
	}
}
