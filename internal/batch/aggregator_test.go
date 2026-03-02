package batch

import (
	"testing"
)

func TestNewAggregator(t *testing.T) {
	cfg := &Config{
		TotalNodes:       25,
		HonestNodes:      25,
		MaliciousNodes:   0,
		RedundancyFactor: 10,
	}

	agg := NewAggregator(cfg)
	if agg == nil {
		t.Fatal("expected non-nil aggregator")
	}
	if agg.Config != cfg {
		t.Fatal("expected aggregator to keep provided config")
	}
}

func TestProcessRoundSuccess(t *testing.T) {
	agg := NewAggregator(&Config{
		TotalNodes:       25,
		HonestNodes:      25,
		MaliciousNodes:   0,
		RedundancyFactor: 10,
	})

	if err := agg.ProcessRound(ModeHonestOnly); err != nil {
		t.Fatalf("expected round to pass, got error: %v", err)
	}
}

func TestProcessRoundLivenessFailure(t *testing.T) {
	agg := NewAggregator(&Config{
		TotalNodes:       3,
		HonestNodes:      1,
		MaliciousNodes:   0,
		RedundancyFactor: 1,
	})

	if err := agg.ProcessRound(ModeHonestOnly); err == nil {
		t.Fatal("expected liveness failure error")
	}
}

func TestProcessRoundSafetyFailure(t *testing.T) {
	agg := NewAggregator(&Config{
		TotalNodes:       10,
		HonestNodes:      5,
		MaliciousNodes:   5,
		RedundancyFactor: 10,
	})

	if err := agg.ProcessRound(ModeByzantineMix); err == nil {
		t.Fatal("expected safety failure error")
	}
}
