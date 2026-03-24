# TPM Attestation Bottleneck Analysis (2026-03-24)

## Executive Summary

Targeted hot-path optimization reduced attestation latency by 80.4% and delivered a 5.11x speedup in the benchmark path.

- Baseline full attestation p50: 48.4 us
- Optimized full attestation p50: 9.8 us
- Baseline p99: 81.6 us
- Optimized p99: 19.5 us

Despite strong single-node gains, concurrent batch throughput remains constrained by thread-pool and cache-lock overheads at small-to-medium batch sizes.

## Top Bottlenecks (Ranked)

| Rank | Bottleneck | Baseline Latency | Budget Share | Severity |
| --- | --- | --- | --- | --- |
| 1 | SDK attest call (FFI/library boundary) | 42.72 us | 82.5% | Critical |
| 2 | HMAC-SHA256 signing | 6.13 us | 11.8% | High |
| 3 | JSON serialization/deserialization | 2.50 us | 4.8% | Medium |

## Latency Distribution (N=500)

| Operation | p50 | p95 | p99 |
| --- | --- | --- | --- |
| Baseline full attestation | 48.4 us | 68.0 us | 81.6 us |
| Optimized full attestation | 9.8 us | 10.2 us | 19.5 us |
| Speedup | 4.94x | 6.67x | 4.19x |

## Cache Performance Delta

- Batch-100 at ~50.5% hit rate: 77.4 us vs 103.5 us uncached (1.34x)
- Batch-500 at ~50.1% hit rate: 75.8 us vs 81.6 us uncached (1.08x)
- Single-node cached throughput: 139,692 att/s vs 22,862 att/s baseline (6.11x)

Observation: cache gains plateau around 50% hit rate in typical FL rounds because unique node IDs reduce key reuse.

## Throughput Under Concurrent Load

| Batch Size | Uncached Throughput | Cached Throughput | Thread Peak |
| --- | --- | --- | --- |
| 50 | 7,855 att/s | 7,095 att/s | 1T |
| 100 | 9,661 att/s | 12,925 att/s | 1T |
| 250 | 9,268 att/s | 11,085 att/s | 16T |
| 500 | 12,255 att/s | 13,201 att/s | 2T |

Key finding: batch throughput (7K to 13K att/s) is still 10x to 17x lower than single-node optimized throughput (139,692 att/s). Runtime overhead (thread management and lock contention) dominates at these batch sizes.

## Root Cause Analysis and Mitigations

### 1) SDK Attest Call (Critical, 82.5% of budget)

Root cause:
- Full Python-to-C boundary crossing and marshalling on each call
- No reusable result path for repeated attestation windows

Outlier behavior:
- Max observed latency spike: 3,110 us (possible GC pause, warmup, or transient runtime jitter)

Mitigations:
1. Extend attestation cache TTL from 5s to 30s for round-scoped reuse.
2. Pre-warm cache at FL round start by issuing batched attest operations.
3. Add spike circuit breaker to bypass/evict entries above 200 us.

### 2) HMAC Signing (High, 11.8% of budget)

Root cause:
- `hmac.new(key, msg, sha256).hexdigest()` recomputed for repeated `(pcr0, nonce)` pairs.

Optimization already validated:
- LRU cache (256 entries) yielded up to 77x hit-path speedup.

Mitigations:
1. Set HMAC cache size to expected active node cardinality (for example 1,000).
2. Use round-scoped nonce instead of per-call nonce to maximize key reuse.
3. Precompute digest pool during round initialization.

### 3) JSON Serialization (Medium, 4.8% of budget)

Root cause:
- Per-call JSON parse allocates a new dictionary and incurs allocator pressure.

Fix already validated:
- Return pre-parsed payload by reference in the hot path.

Mitigations:
1. Consider msgpack/protobuf for compact, lower-overhead payload handling.
2. For ultra-low latency path, return C-level structs and bypass JSON in the fast path.

## Production Rollout Plan

### Immediate

1. Roll out round-scoped nonce strategy.
2. Increase attestation cache TTL to 30s.
3. Track hit rate and p95/p99 impact for one FL workload window.

Expected near-term effect: materially higher cache hit rates and lower tail latency in repeated attestation phases.

### Short-Term

1. Implement and tune latency spike circuit breaker (threshold: 200 us).
2. Introduce msgpack payload path for attestation transport/processing.
3. Add lock-contention instrumentation around cache and thread-pool scheduling.

### Long-Term

1. Evaluate hardware TPM/native driver path to reduce or remove FFI overhead.
2. Profile thread-pool model and queueing under batch pressure; consider bounded worker pools and lock sharding.

## Recommended SLO/Operational Targets

- Attestation latency p95: <= 20 us (optimized path)
- Attestation latency p99: <= 30 us (optimized path)
- Cache hit rate: >= 90% during stable FL rounds with round-scoped nonce
- Spike budget: < 1% requests above 200 us

## Conclusion

The SDK attest call remains the dominant bottleneck and primary optimization lever. Current improvements are substantial, but end-to-end batch throughput is still constrained by concurrency overheads. Round-scoped nonce reuse plus longer-lived cache entries provides the clearest path to high hit rates and additional latency reduction in production.
