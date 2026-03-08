# First 24-Hour Testnet Monitoring Runbook (V1.1.0)

Scope: First 24 hours after Genesis launch
Owner: Testnet Operations

## T-30 to T-0 (Pre-Launch)

1. Confirm branch and commit on launch host.
- `git rev-parse --short HEAD`
- `git branch --show-current`

2. Run pre-launch checklist.
- `./scripts/prelaunch-checklist.sh`
- Save produced log path from script output.

3. Confirm required ports are free.
- `lsof -i :8000 -i :8080 -i :9090 -i :3000 -i :9093`

## T+0 to T+15m (Genesis Bring-Up)

1. Start Genesis launch.
- `./genesis-launch.sh`

2. Verify containers and critical services are healthy.
- `docker compose ps`
- `docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'`

3. Verify runtime startup line for node health.
- `go run ./cmd/sovereign-node start -node-id launch-check-node | head -n 1`
- Expect non-zero `dialed` and `gossip_fanout`.

## T+15m to T+1h (Stabilization)

1. Verify metrics exporter endpoints.
- `go run ./cmd/metrics-exporter -listen :9108 &`
- `curl -sf http://localhost:9108/metrics/scheduler | head`
- `curl -sf http://localhost:9108/metrics/simulator | head`

2. Confirm Prometheus target visibility.
- Open Prometheus and verify `scheduler-exporter` and `simulator-exporter` targets are up.

3. Execute soak smoke.
- `./scripts/run-50node-soak-test.sh`
- Archive result file from `test-results/`.

## T+1h to T+6h (Early Reliability Window)

Run every 30 minutes:
1. Check service health.
- `docker compose ps`

2. Check for container restarts.
- `docker ps --format '{{.Names}} {{.RunningFor}} {{.Status}}'`

3. Check round completion drift from metrics.
- `curl -sf http://localhost:9108/metrics/scheduler | grep -E 'rounds_completed|timeouts|stragglers|last_update'`

4. Check simulator telemetry path (if enabled).
- `curl -sf http://localhost:9108/metrics/simulator | grep -E 'rounds_completed|straggler_events|malicious_events'`

Escalate immediately if:
- `dialed=0` or `gossip_fanout=0`
- Scheduler timeouts spike continuously
- Any critical service enters restart loop

## T+6h to T+24h (Sustainment Window)

Run every 60 minutes:
1. Health and restart checks.
- `docker compose ps`
- `docker ps --format '{{.Names}} {{.Status}}'`

2. Resource pressure checks.
- `free -h`
- `df -h`
- `top -b -n1 | head -20`

3. Data quality checks.
- Confirm new soak/simulation outputs are being generated as expected.

4. Security and secret posture checks.
- `bash validate-secrets.sh prod`

## Incident Response Steps

If peer connectivity degrades:
1. Restart metrics exporter and node runtime checks.
2. Validate bootstrap/seed files in `network/bootstrap/`.
3. Re-run startup smoke command and capture output.

If round progress stalls:
1. Inspect scheduler metrics endpoint.
2. Check Prometheus alerts and service health.
3. If unresolved in 15 minutes, initiate rollback or pause joins.

## Exit Criteria at T+24h

1. No critical service restart loops.
2. Stable non-zero peer dial/gossip fanout.
3. No sustained scheduler timeout storm.
4. Monitoring stack stable and collecting data.
5. At least one successful soak execution archived after launch.
