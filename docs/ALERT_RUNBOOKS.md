# Alert Runbooks

This document defines first-response procedures for SLO and consensus alerts.

## Alert Contract Matrix (Sprint 9)

### Routing and Inhibition Baseline
- Route policy source: [alertmanager.yml](../alertmanager.yml)
- Rule sources: [fl_slo_alerts.yml](../fl_slo_alerts.yml), [fl_detailed_alerts.yml](../fl_detailed_alerts.yml), [tpm_alerts.yml](../tpm_alerts.yml)
- Unit test sources: [fl_slo_alerts.test.yml](../fl_slo_alerts.test.yml), [fl_detailed_alerts.test.yml](../fl_detailed_alerts.test.yml), [tpm_alerts.test.yml](../tpm_alerts.test.yml), [internal/monitoring/alertmanager_config_test.go](../internal/monitoring/alertmanager_config_test.go)

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
| SlowTrustVerification | pending | warning | tpm | default | Service-scope inhibition does not apply without service label | tpm_alerts.test.yml | No |
| SlowMessageVerification | pending | warning | tpm | default | Service-scope inhibition does not apply without service label | tpm_alerts.test.yml | No |
| MostCertificatesNotVerified | pending | warning | tpm | default | Service-scope inhibition does not apply without service label | tpm_alerts.test.yml | No |
| HighRatioOfRevokedCertificates | pending | warning | tpm | default | Service-scope inhibition does not apply without service label | tpm_alerts.test.yml | No |

### Coverage Summary
- Total alerts configured: 31
- Alerts with explicit runbook section in this document: 11
- Alerts with promtool rule unit tests: 31
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
