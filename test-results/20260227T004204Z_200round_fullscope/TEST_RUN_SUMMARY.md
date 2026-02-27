# 200-Round Full-Scope Test (In Progress)

- Started (UTC): 2026-02-27T00:44:06Z
- Stack: docker-compose.full.yml
- Nodes: 10
- Config:
  - NUM_ROUNDS=200
  - MIN_FIT_CLIENTS=1
  - MIN_AVAILABLE_CLIENTS=1
  - ROUND_TIMEOUT_SECONDS=1200

## Monitoring Hooks Enabled
- Backend health/status/convergence/metrics
- Prometheus rules + alerts API
- Alertmanager status
- Container restart count + backend error window (15s)

## Key Artifacts
- run config: run_config.env
- baseline status: compose_ps_start.txt, backend_status_start.json, convergence_start.json
- baseline metrics/logs: backend_metrics_start.prom, backend_logs_start.txt
- Prometheus snapshots: prometheus_rules_start.json, prometheus_alerts_start.json
- live detailed data:
  - live_round_metrics.csv
  - live_events.log
  - monitor_stdout.log
  - live_monitor_200round.sh
  - monitor.pid

## Current Snapshot
[2026-02-27T00:43:52Z] round=14 acc=99.5 loss=0.1 rounds_total=14.0 fl_round_gauge=14.0 nodes=10 backend_restarts=1 backend_errs_15s=0 alertmanager='Up 7 minutes'
