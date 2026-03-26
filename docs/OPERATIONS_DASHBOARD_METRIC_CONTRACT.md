# Operations Dashboard Metric Contract

This document defines the contract for the lower-half panels in the Operations Live Overview dashboard.

Dashboard file:
- grafana/provisioning/dashboards/operations_overview.json

## Scope

This contract covers non-row panels with y >= 48 (consensus/churn + trust/governance + TPM internals).

## Panel-to-Metric Mapping

| Panel ID | Panel Title | Query Inputs (primary) |
| --- | --- | --- |
| 51 | Consensus Capacity and Load | sovereignmap_active_nodes, tokenomics_validator_count, sovereignmap_fl_round |
| 52 | Async and Staleness Signals | sovereignmap_fl_rounds_total, tpm_trust_verification_failures_total, sovereign_ops_http_request_p95_ms_5m |
| 53 | Churn Event Totals | sovereign_ops_control_actions_total |
| 61 | Differential Privacy Budget (Epsilon) | tokenomics_fl_verification_ratio, tokenomics_fl_average_confidence_bps |
| 62 | Federated Risk Signals (%) | tokenomics_fl_verification_ratio, tokenomics_stake_concentration_gini, tokenomics_bridge_collateral_ratio_percent |
| 63 | TEE Capacity and CXL Throughput | tpm_ca_certificate_valid, tpm_node_trust_score, tokenomics_bridge_routes_active |
| 64 | Attestation and Thermal Telemetry | tpm_node_attestation_latency_ms, tpm_node_trust_score, tokenomics_validator_count |
| 65 | Governance Economics Signals | tokenomics_token_supply_total, tokenomics_stake_concentration_gini, tokenomics_validator_count, tokenomics_stake_participation_ratio |
| 67 | TPM Cache Hit vs Miss Rate | tpm_certificates_verified_total, tpm_trust_verification_failures_total, tpm_ca_certificate_valid, tpm_certificates_total |
| 68 | TPM Lock Wait p95 (us) | tpm_node_attestation_latency_ms |
| 69 | TPM Nonce Replay and Cleanup | tpm_trust_verification_failures_total, tpm_certificates_verified_total, tpm_certificates_total, tpm_ca_certificate_valid |
| 70 | FL Aggregation Path Usage | sovereignmap_fl_rounds_total |
| 71 | What Changed (Current vs Prior Window) | sovereignmap_fl_rounds_total, tokenomics_bridge_transfers_total, tokenomics_validator_count, sovereign_ops_control_actions_total |

## Validation Paths

Static query compatibility:
- python3 scripts/check_dashboard_queries.py

Live query smoke (requires local compose stack):
- python3 scripts/check_operations_overview_live_queries.py

## Expected Prometheus Jobs

For reliable lower-half panel population, these jobs should be up:
- sovereign-backend
- tpm-metrics
- tokenomics-metrics
- fl-performance-metrics

Check status:
- curl -fsS http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health, lastError: .lastError}'
