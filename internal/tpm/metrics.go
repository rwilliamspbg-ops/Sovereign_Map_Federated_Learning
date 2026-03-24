package tpm

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	tpmCacheEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mohawk_tpm_cache_events_total",
			Help: "Total TPM cache events by cache type and result.",
		},
		[]string{"cache", "result"},
	)

	tpmLockWaitSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mohawk_tpm_lock_wait_seconds",
			Help:    "Time spent waiting to acquire TPM-related locks.",
			Buckets: []float64{0.000001, 0.0000025, 0.000005, 0.00001, 0.000025, 0.00005, 0.0001, 0.00025, 0.0005, 0.001},
		},
		[]string{"target", "mode"},
	)

	tpmNonceCleanupRunsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mohawk_tpm_nonce_cleanup_runs_total",
			Help: "Total number of nonce cleanup runs.",
		},
	)

	tpmNonceCleanupRemovedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mohawk_tpm_nonce_cleanup_removed_total",
			Help: "Total number of expired nonce entries removed by cleanup.",
		},
	)

	tpmPrewarmRequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mohawk_tpm_prewarm_requests_total",
			Help: "Total number of TPM quote prewarm requests.",
		},
	)

	tpmPrewarmWarmedNodesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mohawk_tpm_prewarm_warmed_nodes_total",
			Help: "Total number of node quotes successfully warmed into cache.",
		},
	)

	tpmNonceReplayRejectionsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "mohawk_tpm_nonce_replay_rejections_total",
			Help: "Total number of replayed nonce rejections.",
		},
	)

	quoteCacheHitsAtomic         atomic.Uint64
	quoteCacheMissesAtomic       atomic.Uint64
	attestationCacheHitsAtomic   atomic.Uint64
	attestationCacheMissesAtomic atomic.Uint64
	nonceCleanupRunsAtomic       atomic.Uint64
	nonceCleanupRemovedAtomic    atomic.Uint64
	prewarmRequestsAtomic        atomic.Uint64
	prewarmWarmedAtomic          atomic.Uint64
	nonceReplayRejectsAtomic     atomic.Uint64
)

func init() {
	prometheus.MustRegister(
		tpmCacheEventsTotal,
		tpmLockWaitSeconds,
		tpmNonceCleanupRunsTotal,
		tpmNonceCleanupRemovedTotal,
		tpmPrewarmRequestsTotal,
		tpmPrewarmWarmedNodesTotal,
		tpmNonceReplayRejectionsTotal,
	)
}

type metricsSnapshot struct {
	quoteCacheHits         uint64
	quoteCacheMisses       uint64
	attestationCacheHits   uint64
	attestationCacheMisses uint64
	nonceCleanupRuns       uint64
	nonceCleanupRemoved    uint64
	prewarmRequests        uint64
	prewarmWarmed          uint64
	nonceReplayRejects     uint64
}

func tpmMetricsSnapshot() metricsSnapshot {
	return metricsSnapshot{
		quoteCacheHits:         quoteCacheHitsAtomic.Load(),
		quoteCacheMisses:       quoteCacheMissesAtomic.Load(),
		attestationCacheHits:   attestationCacheHitsAtomic.Load(),
		attestationCacheMisses: attestationCacheMissesAtomic.Load(),
		nonceCleanupRuns:       nonceCleanupRunsAtomic.Load(),
		nonceCleanupRemoved:    nonceCleanupRemovedAtomic.Load(),
		prewarmRequests:        prewarmRequestsAtomic.Load(),
		prewarmWarmed:          prewarmWarmedAtomic.Load(),
		nonceReplayRejects:     nonceReplayRejectsAtomic.Load(),
	}
}

func observeQuoteCacheHit() {
	tpmCacheEventsTotal.WithLabelValues("quote", "hit").Inc()
	quoteCacheHitsAtomic.Add(1)
}

func observeQuoteCacheMiss() {
	tpmCacheEventsTotal.WithLabelValues("quote", "miss").Inc()
	quoteCacheMissesAtomic.Add(1)
}

func observeAttestationCacheHit() {
	tpmCacheEventsTotal.WithLabelValues("attestation", "hit").Inc()
	attestationCacheHitsAtomic.Add(1)
}

func observeAttestationCacheMiss() {
	tpmCacheEventsTotal.WithLabelValues("attestation", "miss").Inc()
	attestationCacheMissesAtomic.Add(1)
}

func observeQuoteCacheLockWait(mode string, wait time.Duration) {
	tpmLockWaitSeconds.WithLabelValues("quote_cache", mode).Observe(wait.Seconds())
}

func observeNonceCacheLockWait(wait time.Duration) {
	tpmLockWaitSeconds.WithLabelValues("nonce_cache", "write").Observe(wait.Seconds())
}

func observeAttestationCacheLockWait(mode string, wait time.Duration) {
	tpmLockWaitSeconds.WithLabelValues("attestation_cache", mode).Observe(wait.Seconds())
}

func observeNonceCleanup(removed int) {
	tpmNonceCleanupRunsTotal.Inc()
	nonceCleanupRunsAtomic.Add(1)
	if removed > 0 {
		tpmNonceCleanupRemovedTotal.Add(float64(removed))
		nonceCleanupRemovedAtomic.Add(uint64(removed))
	}
}

func observePrewarm(requested, warmed int) {
	if requested > 0 {
		tpmPrewarmRequestsTotal.Add(float64(requested))
		prewarmRequestsAtomic.Add(uint64(requested))
	}
	if warmed > 0 {
		tpmPrewarmWarmedNodesTotal.Add(float64(warmed))
		prewarmWarmedAtomic.Add(uint64(warmed))
	}
}

func observeNonceReplayRejected() {
	tpmNonceReplayRejectionsTotal.Inc()
	nonceReplayRejectsAtomic.Add(1)
}
