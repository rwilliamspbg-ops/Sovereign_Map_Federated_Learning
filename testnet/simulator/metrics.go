package simulator

import (
	"fmt"
	"strings"
)

// Prometheus renders simulation result metrics in text exposition format.
func Prometheus(r Result, scenario string) string {
	s := sanitizeLabel(scenario)
	labels := fmt.Sprintf("scenario=\"%s\"", s)

	var b strings.Builder
	b.WriteString("# HELP sovereign_simulator_nodes Number of simulated nodes\n")
	b.WriteString("# TYPE sovereign_simulator_nodes gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_simulator_nodes{%s} %d\n", labels, r.NodeCount))
	b.WriteString("# HELP sovereign_simulator_rounds_requested_total Requested simulation rounds\n")
	b.WriteString("# TYPE sovereign_simulator_rounds_requested_total counter\n")
	b.WriteString(fmt.Sprintf("sovereign_simulator_rounds_requested_total{%s} %d\n", labels, r.RoundsRequested))
	b.WriteString("# HELP sovereign_simulator_rounds_completed_total Completed simulation rounds\n")
	b.WriteString("# TYPE sovereign_simulator_rounds_completed_total counter\n")
	b.WriteString(fmt.Sprintf("sovereign_simulator_rounds_completed_total{%s} %d\n", labels, r.RoundsCompleted))
	b.WriteString("# HELP sovereign_simulator_straggler_events_total Straggler events in simulation\n")
	b.WriteString("# TYPE sovereign_simulator_straggler_events_total counter\n")
	b.WriteString(fmt.Sprintf("sovereign_simulator_straggler_events_total{%s} %d\n", labels, r.StragglerEvents))
	b.WriteString("# HELP sovereign_simulator_malicious_events_total Malicious events in simulation\n")
	b.WriteString("# TYPE sovereign_simulator_malicious_events_total counter\n")
	b.WriteString(fmt.Sprintf("sovereign_simulator_malicious_events_total{%s} %d\n", labels, r.MaliciousNodeEvents))
	b.WriteString("# HELP sovereign_simulator_average_round_duration_seconds Average simulated round duration\n")
	b.WriteString("# TYPE sovereign_simulator_average_round_duration_seconds gauge\n")
	b.WriteString(fmt.Sprintf("sovereign_simulator_average_round_duration_seconds{%s} %.6f\n", labels, r.AverageRoundDuration.Seconds()))

	return b.String()
}

func sanitizeLabel(v string) string {
	v = strings.ReplaceAll(v, "\\", "")
	v = strings.ReplaceAll(v, "\"", "")
	return v
}
