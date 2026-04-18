package scheduler

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Metrics tracks scheduler round lifecycle counters.
type Metrics struct {
	mu              sync.RWMutex
	roundsStarted   int
	roundsCompleted int
	timeouts        int
	stragglers      int
	lastRoundID     int
	lastUpdated     time.Time
}

// NewMetrics creates a fresh scheduler metrics collector.
func NewMetrics() *Metrics {
	return &Metrics{lastUpdated: time.Now().UTC()}
}

// RecordRoundStart records a started round.
func (m *Metrics) RecordRoundStart(roundID int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roundsStarted++
	m.lastRoundID = roundID
	m.lastUpdated = time.Now().UTC()
}

// RecordRoundComplete records a completed round.
func (m *Metrics) RecordRoundComplete(roundID int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roundsCompleted++
	m.lastRoundID = roundID
	m.lastUpdated = time.Now().UTC()
}

// RecordTimeout increments timeout count.
func (m *Metrics) RecordTimeout() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timeouts++
	m.lastUpdated = time.Now().UTC()
}

// RecordStraggler increments straggler count.
func (m *Metrics) RecordStraggler() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stragglers++
	m.lastUpdated = time.Now().UTC()
}

// Prometheus renders metrics in text exposition format.
func (m *Metrics) Prometheus(nodeID string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	labels := fmt.Sprintf("node_id=\"%s\"", sanitizeLabel(nodeID))
	var b strings.Builder
	b.WriteString("# HELP sovereign_scheduler_rounds_started_total Total scheduler rounds started\n")
	b.WriteString("# TYPE sovereign_scheduler_rounds_started_total counter\n")
	fmt.Fprintf(&b, "sovereign_scheduler_rounds_started_total{%s} %d\n", labels, m.roundsStarted)
	b.WriteString("# HELP sovereign_scheduler_rounds_completed_total Total scheduler rounds completed\n")
	b.WriteString("# TYPE sovereign_scheduler_rounds_completed_total counter\n")
	fmt.Fprintf(&b, "sovereign_scheduler_rounds_completed_total{%s} %d\n", labels, m.roundsCompleted)
	b.WriteString("# HELP sovereign_scheduler_timeouts_total Total scheduler round timeouts\n")
	b.WriteString("# TYPE sovereign_scheduler_timeouts_total counter\n")
	fmt.Fprintf(&b, "sovereign_scheduler_timeouts_total{%s} %d\n", labels, m.timeouts)
	b.WriteString("# HELP sovereign_scheduler_stragglers_total Total scheduler straggler detections\n")
	b.WriteString("# TYPE sovereign_scheduler_stragglers_total counter\n")
	fmt.Fprintf(&b, "sovereign_scheduler_stragglers_total{%s} %d\n", labels, m.stragglers)
	b.WriteString("# HELP sovereign_scheduler_last_round_id Last observed scheduler round id\n")
	b.WriteString("# TYPE sovereign_scheduler_last_round_id gauge\n")
	fmt.Fprintf(&b, "sovereign_scheduler_last_round_id{%s} %d\n", labels, m.lastRoundID)
	b.WriteString("# HELP sovereign_scheduler_last_update_unix Last scheduler metric update timestamp\n")
	b.WriteString("# TYPE sovereign_scheduler_last_update_unix gauge\n")
	fmt.Fprintf(&b, "sovereign_scheduler_last_update_unix{%s} %d\n", labels, m.lastUpdated.Unix())
	return b.String()
}

func sanitizeLabel(v string) string {
	v = strings.ReplaceAll(v, "\\", "")
	v = strings.ReplaceAll(v, "\"", "")
	return v
}
