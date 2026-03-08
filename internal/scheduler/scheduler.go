package scheduler

import "time"

// RoundState captures minimum scheduler state needed to coordinate FL rounds.
type RoundState struct {
	RoundID    int
	StartedAt  time.Time
	Deadline   time.Time
	NodeTarget int
}

// NewRoundState initializes a new round with timeout policy.
func NewRoundState(roundID, nodeTarget int, timeout time.Duration) RoundState {
	start := time.Now().UTC()
	return RoundState{
		RoundID:    roundID,
		StartedAt:  start,
		Deadline:   start.Add(timeout),
		NodeTarget: nodeTarget,
	}
}

// IsExpired reports whether a round has exceeded its deadline.
func (r RoundState) IsExpired(now time.Time) bool {
	return now.After(r.Deadline)
}
