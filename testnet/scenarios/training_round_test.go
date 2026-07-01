package scenarios

import (
	"testing"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/testnet/simulator"
)

func TestTrainingRoundScenarioCompletesRequestedRounds(t *testing.T) {
	result := simulator.Run(simulator.Config{
		NodeCount:         64,
		Rounds:            12,
		RoundDuration:     200 * time.Millisecond,
		StragglerRate:     0,
		MaliciousNodeRate: 0,
		RandomSeed:        1234,
	})

	if result.NodeCount != 64 {
		t.Fatalf("expected 64 nodes, got %d", result.NodeCount)
	}
	if result.RoundsCompleted != result.RoundsRequested {
		t.Fatalf("expected all rounds to complete, got %d/%d", result.RoundsCompleted, result.RoundsRequested)
	}
	if result.StragglerEvents != 0 {
		t.Fatalf("expected no straggler events, got %d", result.StragglerEvents)
	}
	if result.MaliciousNodeEvents != 0 {
		t.Fatalf("expected no malicious events, got %d", result.MaliciousNodeEvents)
	}
	if result.AverageRoundDuration != 200*time.Millisecond {
		t.Fatalf("expected average round duration of 200ms, got %s", result.AverageRoundDuration)
	}
}
