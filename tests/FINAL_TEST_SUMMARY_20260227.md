# Final Test Summary — 2026-02-27

## Scope

This document captures the latest available federated-learning test outputs in this repository and consolidates the key outcome metrics for commit history.

## Latest Captured Run

- Run folder: `test-results/round200_live/20260227_014246_10node_bft`
- Target rounds reached: `200/200`
- Final round: `200`
- Final accuracy: `99.5`
- Final loss: `0.1`
- Samples collected (summary tracker): `9`

Primary source files used:

- `test-results/round200_live/20260227_014246_10node_bft/summary.txt`
- `test-results/round200_live/20260227_014246_10node_bft/metrics_summary_final.json`

## Full-Scope Monitoring Session

- Session folder: `test-results/20260227T004204Z_200round_fullscope`
- Session start (UTC): `2026-02-27T00:44:06Z`
- Stack: `docker-compose.full.yml`
- Node count: `10`
- Config highlights:
  - `NUM_ROUNDS=200`
  - `MIN_FIT_CLIENTS=1`
  - `MIN_AVAILABLE_CLIENTS=1`
  - `ROUND_TIMEOUT_SECONDS=1200`

Key artifacts captured in that folder include:

- `TEST_RUN_SUMMARY.md`
- `live_round_metrics.csv`
- `live_events.log`
- `monitor_stdout.log`
- `backend_status_start.json`
- `convergence_start.json`

## Result Data References

Additional aligned outputs available under:

- `results/test-runs/`
- `results/benchmarks/`
- `results/analysis/`

Notable files from the latest boundary workflow:

- `results/test-runs/bft_20node_200round_50_70_results_20260227T124206Z.json`
- `results/benchmarks/bft_20node_200round_50_70_summary_20260227T124206Z.csv`
- `results/analysis/BFT_20NODE_200ROUND_BOUNDARY_REPORT_20260227T124338Z.md`

## Final Status

- Last run metrics captured and summarized.
- Test-result artifacts are present in repository result directories.
- This summary file provides a stable final checkpoint for version control.