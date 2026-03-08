package simulator

import (
	"fmt"
	"math/rand"
	"time"
)

// Config controls simulation scale and timing for soak tests.
type Config struct {
	NodeCount         int
	Rounds            int
	RoundDuration     time.Duration
	StragglerRate     float64
	MaliciousNodeRate float64
	RandomSeed        int64
}

// Result summarizes simulation outcomes for operator review.
type Result struct {
	NodeCount            int
	RoundsRequested      int
	RoundsCompleted      int
	StragglerEvents      int
	MaliciousNodeEvents  int
	AverageRoundDuration time.Duration
}

// Run executes a deterministic-in-shape, stochastic-in-events training simulation.
func Run(cfg Config) Result {
	if cfg.NodeCount <= 0 {
		cfg.NodeCount = 50
	}
	if cfg.Rounds <= 0 {
		cfg.Rounds = 100
	}
	if cfg.RoundDuration <= 0 {
		cfg.RoundDuration = 250 * time.Millisecond
	}
	if cfg.StragglerRate < 0 {
		cfg.StragglerRate = 0
	}
	if cfg.StragglerRate > 1 {
		cfg.StragglerRate = 1
	}
	if cfg.MaliciousNodeRate < 0 {
		cfg.MaliciousNodeRate = 0
	}
	if cfg.MaliciousNodeRate > 1 {
		cfg.MaliciousNodeRate = 1
	}
	if cfg.RandomSeed == 0 {
		cfg.RandomSeed = time.Now().UnixNano()
	}

	rng := rand.New(rand.NewSource(cfg.RandomSeed))
	result := Result{NodeCount: cfg.NodeCount, RoundsRequested: cfg.Rounds}
	var totalDuration time.Duration

	for i := 0; i < cfg.Rounds; i++ {
		roundDuration := cfg.RoundDuration

		if rng.Float64() < cfg.StragglerRate {
			result.StragglerEvents++
			roundDuration += cfg.RoundDuration / 2
		}

		if rng.Float64() < cfg.MaliciousNodeRate {
			result.MaliciousNodeEvents++
		}

		totalDuration += roundDuration
		result.RoundsCompleted++
	}

	result.AverageRoundDuration = totalDuration / time.Duration(result.RoundsCompleted)
	return result
}

// FormatSummary renders a human-readable summary for CI logs.
func FormatSummary(r Result) string {
	return fmt.Sprintf(
		"nodes=%d rounds=%d/%d avg_round=%s stragglers=%d malicious_events=%d",
		r.NodeCount,
		r.RoundsCompleted,
		r.RoundsRequested,
		r.AverageRoundDuration,
		r.StragglerEvents,
		r.MaliciousNodeEvents,
	)
}
