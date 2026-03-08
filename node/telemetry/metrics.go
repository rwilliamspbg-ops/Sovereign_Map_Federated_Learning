package telemetry

// RuntimeMetrics holds key counters for node health and FL progress.
type RuntimeMetrics struct {
	ActivePeers int
	RoundID     int
}
