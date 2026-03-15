package api

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	proofVerificationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mohawk_proof_verifications_total",
			Help: "Total number of proof verifications by scheme/backend/result.",
		},
		[]string{"scheme", "backend", "result"},
	)

	proofVerificationLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mohawk_proof_verification_latency_seconds",
			Help:    "End-to-end proof verification latency in seconds.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"scheme", "backend"},
	)

	ledgerEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mohawk_ledger_events_total",
			Help: "Total number of ledger events recorded by event type.",
		},
		[]string{"event_type"},
	)

	ledgerEntriesGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "mohawk_ledger_entries",
			Help: "Current number of entries stored in the in-memory proof ledger.",
		},
	)
)

func init() {
	prometheus.MustRegister(
		proofVerificationsTotal,
		proofVerificationLatency,
		ledgerEventsTotal,
		ledgerEntriesGauge,
	)
}

func observeProofVerification(scheme string, backend string, accepted bool, latency time.Duration) {
	result := "rejected"
	if accepted {
		result = "accepted"
	}
	proofVerificationsTotal.WithLabelValues(scheme, backend, result).Inc()
	proofVerificationLatency.WithLabelValues(scheme, backend).Observe(latency.Seconds())
}

func observeLedgerEvent(eventType string, currentEntries int) {
	ledgerEventsTotal.WithLabelValues(eventType).Inc()
	ledgerEntriesGauge.Set(float64(currentEntries))
}
