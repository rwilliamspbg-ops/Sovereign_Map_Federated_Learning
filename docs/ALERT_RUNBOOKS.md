# Alert Runbooks

This document defines first-response procedures for SLO and consensus alerts.

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
