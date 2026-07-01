package scenarios

import (
	"testing"
	"time"

	"github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/testnet/simulator"
)

func TestMaliciousNodeScenarioFlagsByzantineEvents(t *testing.T) {
	result := simulator.Run(simulator.Config{
		NodeCount:         32,
		Rounds:            8,
		RoundDuration:     150 * time.Millisecond,
		StragglerRate:     0,
		MaliciousNodeRate: 1,
		RandomSeed:        4321,
	})

	if result.RoundsCompleted != 8 {
		t.Fatalf("expected 8 completed rounds, got %d", result.RoundsCompleted)
	}
	if result.MaliciousNodeEvents != 8 {
		t.Fatalf("expected every round to flag a malicious event, got %d", result.MaliciousNodeEvents)
	}
	if result.StragglerEvents != 0 {
		t.Fatalf("expected no straggler events, got %d", result.StragglerEvents)
	}
	if result.AverageRoundDuration != 150*time.Millisecond {
		t.Fatalf("expected average round duration of 150ms, got %s", result.AverageRoundDuration)
	}
}
