# Alert Runbooks

This document defines first-response procedures for SLO and consensus alerts.

## Alert Contract Matrix (Sprint 9)

### Routing and Inhibition Baseline

- Route policy source: [alertmanager.yml](../alertmanager.yml)
- Rule sources: [fl_slo_alerts.yml](../fl_slo_alerts.yml), [fl_detailed_alerts.yml](../fl_detailed_alerts.yml), [tpm_alerts.yml](../tpm_alerts.yml), [marketplace_alerts.yml](../marketplace_alerts.yml)
- Unit test sources: [fl_slo_alerts.test.yml](../fl_slo_alerts.test.yml), [fl_detailed_alerts.test.yml](../fl_detailed_alerts.test.yml), [tpm_alerts.test.yml](../tpm_alerts.test.yml), [marketplace_alerts.test.yml](../marketplace_alerts.test.yml), [internal/monitoring/alertmanager_config_test.go](../internal/monitoring/alertmanager_config_test.go)

Inhibition semantics:

- ConsensusStatusEndpointDown (critical) suppresses consensus warning alerts.
- Critical alerts suppress warning alerts within the same service.

### FL SLO and Consensus Alerts

| Alert | Owner | Severity | Service | Receiver | Inhibition | Test Coverage | Runbook |
| --- | --- | --- | --- | --- | --- | --- | --- |
| FLRoundStalled | platform | critical | federated-learning | fl-critical | Critical suppresses warning in federated-learning | fl_slo_alerts.test.yml | Yes |
| FLAccuracyDegraded | platform | warning | federated-learning | default | Suppressed by federated-learning critical alerts | fl_slo_alerts.test.yml | Yes |
| FLClientParticipationLow | platform | warning | federated-learning | default | Suppressed by federated-learning critical alerts | fl_slo_alerts.test.yml | Yes |
| FLRoundDurationHighP95 | platform | warning | federated-learning | default | Suppressed by federated-learning critical alerts | fl_slo_alerts.test.yml | Yes |
| BridgeCollateralRatioLow | platform | warning | tokenomics | tokenomics-warning | Suppressed by tokenomics critical alerts | fl_slo_alerts.test.yml | Yes |
| ConsensusStatusEndpointDown | platform | critical | consensus | consensus-critical | Source alert for consensus warning suppression | fl_slo_alerts.test.yml | Yes |
| ConsensusCapacityDrift | platform | warning | consensus | consensus-warning | Suppressed by ConsensusStatusEndpointDown and any consensus critical | fl_slo_alerts.test.yml | Yes |
| ConsensusOpenRoundsBacklog | platform | warning | consensus | consensus-warning | Suppressed by ConsensusStatusEndpointDown and any consensus critical | fl_slo_alerts.test.yml | Yes |
| AsyncStalenessHigh | platform | warning | consensus | consensus-warning | Suppressed by ConsensusStatusEndpointDown and any consensus critical | fl_slo_alerts.test.yml | Yes |
| ChurnBurstDetected | platform | warning | consensus | consensus-warning | Suppressed by ConsensusStatusEndpointDown and any consensus critical | fl_slo_alerts.test.yml | Yes |
| StaleDropBurstDetected | platform | warning | consensus | consensus-warning | Suppressed by ConsensusStatusEndpointDown and any consensus critical | fl_slo_alerts.test.yml | Yes |

### FL Detailed Performance Alerts

| Alert | Owner | Severity | Service Label | Receiver | Inhibition | Test Coverage | Runbook |
| --- | --- | --- | --- | --- | --- | --- | --- |
| HighPrivacyOverhead | pending | warning | none | default | Service-scope inhibition does not apply without service label | fl_detailed_alerts.test.yml | No |
| ByzantineNodesDetected | pending | critical | none | default | Acts as source only when service label exists | fl_detailed_alerts.test.yml | No |
| GPUMemoryExhaustion | pending | critical | none | default | Acts as source only when service label exists | fl_detailed_alerts.test.yml | No |
| NetworkPartitionActive | pending | error | none | default | Not targeted by critical-over-warning inhibition | fl_detailed_alerts.test.yml | No |

### TPM Trust Alerts

| Alert | Owner | Severity | Component | Receiver | Inhibition | Test Coverage | Runbook |
| --- | --- | --- | --- | --- | --- | --- | --- |
| CertificateExpiringIn30Days | pending | warning | tpm | default | Service-scope inhibition does not apply without service label | tpm_alerts.test.yml | No |
| CertificateExpiringIn7Days | pending | critical | tpm | default | Acts as source only when service label exists | tpm_alerts.test.yml | No |
| CertificateExpired | pending | critical | tpm | default | Acts as source only when service label exists | tpm_alerts.test.yml | No |
| CertificateVerificationFailed | pending | warning | tpm | default | Service-scope inhibition does not apply without service label | tpm_alerts.test.yml | No |
| InvalidTrustChain | pending | critical | tpm | default | Acts as source only when service label exists | tpm_alerts.test.yml | No |
| CAInvalid | pending | critical | tpm | default | Acts as source only when service label exists | tpm_alerts.test.yml | No |
| SignatureVerificationFailures | pending | warning | tpm | default | Service-scope inhibition does not apply without service label | tpm_alerts.test.yml | No |
| HighSignatureFailureRate | pending | critical | tpm | default | Acts as source only when service label exists | tpm_alerts.test.yml | No |
| CertificateRevoked | pending | critical | tpm | default | Acts as source only when service label exists | tpm_alerts.test.yml | No |
| HighCRLSize | pending | warning | tpm | default | Service-scope inhibition does not apply without service label | tpm_alerts.test.yml | No |
| LowNodeTrustScore | pending | warning | tpm | default | Service-scope inhibition does not apply without service label | tpm_alerts.test.yml | No |
| ZeroNodeTrustScore | pending | critical | tpm | default | Acts as source only when service label exists | tpm_alerts.test.yml | No |
| SlowTrustVerification | pending | warning | tpm | default | Service-scope inhibition does not apply without service label | tpm_alerts.test.yml | Yes |
| SlowMessageVerification | pending | warning | tpm | default | Service-scope inhibition does not apply without service label | tpm_alerts.test.yml | Yes |
| TPMCacheMissSurge | pending | warning | tpm | default | Service-scope inhibition does not apply without service label | tpm_alerts.test.yml | Yes |
| TPMLockWaitP95Regression | pending | warning | tpm | default | Service-scope inhibition does not apply without service label | tpm_alerts.test.yml | Yes |
| TPMNonceReplayRejectionSpike | pending | warning | tpm | default | Service-scope inhibition does not apply without service label | tpm_alerts.test.yml | Yes |
| MostCertificatesNotVerified | pending | warning | tpm | default | Service-scope inhibition does not apply without service label | tpm_alerts.test.yml | No |
| HighRatioOfRevokedCertificates | pending | warning | tpm | default | Service-scope inhibition does not apply without service label | tpm_alerts.test.yml | No |

### Coverage Summary

- Total alerts configured: 36
- Alerts with explicit runbook section in this document: 18
- Alerts with promtool rule unit tests: 36
- Alertmanager routing and inhibition policy tests: covered by internal/monitoring/alertmanager_config_test.go

## FLRoundStalled

- Check round progression metric: `sovereignmap_fl_rounds_total` and backend scheduler logs.
- Validate backend and aggregator health endpoints.
- Restart stalled workers only after confirming no active commit is in progress.

## FLAccuracyDegraded

- Confirm recent data distribution shifts and node participation quality.
- Compare `sovereignmap_fl_accuracy` against loss and trust-score trends.
- Roll back to last known good model checkpoint if sustained degradation persists.

## FLClientParticipationLow

- Inspect active peer count and network partition indicators.
- Validate join flows and certificate status for recently added nodes.
- Trigger controlled rejoin procedure for disconnected clients.

## FLRoundDurationHighP95

- Check GPU/NPU utilization and queue depths.
- Inspect async staleness and open-round backlog for coupled bottlenecks.
- Scale worker capacity or reduce round payload size where needed.

## BridgeCollateralRatioLow

- Verify current collateral and outstanding bridge obligations.
- Pause new high-risk bridge transfers until ratio recovers.
- Coordinate treasury rebalance per governance policy.

## ConsensusStatusEndpointDown

- Check node-agent API health and `metrics-exporter` reachability.
- Validate connectivity to `/api/v1/consensus/status` from exporter runtime.
- Restore API path availability before addressing downstream warning alerts.

## ConsensusCapacityDrift

- Compare `sovereign_consensus_active_nodes` against `sovereign_consensus_quorum_size`.
- Investigate churn causes: network instability, revoked certs, or crash loops.
- Prioritize node recovery to re-establish quorum safety margin.

## ConsensusOpenRoundsBacklog

- Inspect commit latency and proposal throughput.
- Verify no persistent failures in consensus vote collection.
- Reduce input pressure or scale participants until backlog clears.

## AsyncStalenessHigh

- Review `sovereign_aggregation_staleness_avg_seconds` trend with network lag.
- Investigate stragglers and stale update sources.
- Tighten stale window only after confirming normal network conditions.

## ChurnBurstDetected

- Validate whether churn burst aligns with planned maintenance.
- If unplanned, inspect trust/revocation and node runtime health.
- Apply admission control if burst causes quorum instability.

## StaleDropBurstDetected

- Correlate stale drop spikes with transport latency and node saturation.
- Confirm async mode thresholds are appropriate for current load.
- Mitigate by reducing source lag and improving peer synchronization.

## SlowTrustVerification

- Confirm current p95 trust verification against `tpm_trust_verification_duration_seconds_bucket` and identify if regression is sustained for at least 5 minutes.
- Check cache efficiency using `tpm_trust_cache_hits_total` and `tpm_trust_cache_misses_total`; if hit rate is below expected round baseline, apply round-scoped nonce and extend attestation TTL.
- Inspect SDK/FFI path latency and runtime spikes; trigger circuit-breaker handling for outliers above 200 us and capture traces for postmortem.

## SlowMessageVerification

- Validate p95 message verification latency from `tpm_message_verification_duration_seconds_bucket` and correlate with signature verification failure rates.
- Review HMAC cache cardinality versus active node count; increase cache size when repeated `(pcr0, nonce)` pairs are evicted too aggressively.
- If latency remains elevated, reduce per-call serialization overhead (prefer pre-parsed payload path) and assess lock contention in thread pool and cache critical sections.

## TPMCacheMissSurge

- Confirm miss ratio with `mohawk_tpm_cache_events_total{cache=~"quote|attestation",result=~"hit|miss"}` over 5m and verify the condition persists for at least 10m.
- Check round nonce mode (`TPM_ROUND_SCOPED_NONCE`) and attestation TTL (`TPM_ATTESTATION_CACHE_TTL`); enable round-scoped nonce and raise TTL when misses remain elevated.
- Review cache cardinality pressure by comparing active node count to configured attestation report capacity (`TPM_ATTESTATION_MAX_REPORTS`).

## TPMLockWaitP95Regression

- Inspect `mohawk_tpm_lock_wait_seconds_bucket` by `{target,mode}` to isolate contention hotspot (`quote_cache`, `attestation_cache`, or `nonce_cache`).
- Correlate with request surge and cache miss signals; lock-wait regressions paired with high misses usually indicate ineffective cache reuse.
- Mitigate by restoring cache hit rate first (nonce/TTL tuning) before scaling worker concurrency.

## TPMNonceReplayRejectionSpike

- Verify replay rate trend from `mohawk_tpm_nonce_replay_rejections_total` and determine whether it is expected (duplicate retries) or anomalous (nonce generation collision/replay attack).
- Correlate with client retry storms and transport retransmissions; high replay without failure spikes usually indicates duplicate delivery.
- If anomalous, rotate nonce derivation context for the affected round and audit ingress paths for duplicate submissions.

## MarketplaceEscrowStalled

- Confirm `sovereign_marketplace_escrow_locked` is non-zero and verify `increase(sovereign_marketplace_payout_total[30m]) == 0` in Prometheus UI.
- Inspect pending contracts via `/marketplace/contracts?payout_status=pending` and verify no active disputes are blocking payout release.
- Triage release path by checking `/marketplace/escrow/release` API logs and recent governance actions for moderation holds.

## MarketplaceEscrowHighWatermark

- Validate current locked amount against expected round budget and contract volume.
- Inspect for stale contracts that remain pending after round completion and release in controlled batches.
- If sustained, tighten intent budget limits or increase release cadence to keep escrow within policy bounds.
