# C2 Swarm HUD Quick Start

This guide covers secure operation of the command-and-control swarm HUD.

## Accessing C2 HUD

1. Start the stack (example):

```bash
docker compose -f docker-compose.full.yml up -d --scale node-agent=5
```

1. Open the C2 HUD view:

- Main HUD: `http://localhost:3000`
- C2 HUD: `http://localhost:3000/?view=c2`

## Command Authorization Model

`POST /swarm/command` requires admin authentication (`Authorization: Bearer <token>` or `X-Join-Admin-Token`).

Role policy is evaluated using `X-API-Role`:

- `admin`: full command set
- `commander`: broad tactical control (except isolate-only paths)
- `operator`: limited tactical controls
- `auditor`, `viewer`: read-only

If role policy blocks a command, the API returns `403` with `error=forbidden`.

## Signed Audit Trail

Every accepted command is appended to an HMAC-signed audit chain.

- Audit API: `GET /swarm/audit/recent?limit=100` (admin auth required)
- Signature key: `SWARM_AUDIT_SIGNING_KEY`
- Persisted file: `SWARM_AUDIT_LOG_PATH` (default `./data/swarm_command_audit.jsonl`)

## Performance and Safety Defaults

- Map node cap: `SWARM_MAX_MAP_NODES` (default `1000`)
- Command rate limit: `SWARM_COMMAND_RATE_LIMIT_PER_MIN` (default `120`)
- Command log cap: `SWARM_COMMAND_LOG_MAX` (default `500`)
- Idempotency window: nonce replay suppression over a rolling 10-minute window

## Operator Commands

Quick endpoint checks:

```bash
curl -s http://localhost:8000/swarm/status | jq
curl -s "http://localhost:8000/swarm/map?limit=200&layers=paths,coverage" | jq '.node_count'
curl -s http://localhost:8000/swarm/commands?limit=20 | jq
curl -s http://localhost:8000/autonomy/twin/summary | jq
curl -s http://localhost:8000/autonomy/planner/insights | jq
```

Autonomy sensor and SLO checks:

```bash
curl -s http://localhost:8000/autonomy/sensors/quality | jq
curl -s http://localhost:8000/autonomy/slo/status | jq
```

Submit a command (example):

```bash
curl -s -X POST http://localhost:8000/swarm/command \
  -H "Authorization: Bearer ${JOIN_API_ADMIN_TOKEN}" \
  -H "X-API-Role: commander" \
  -H "Content-Type: application/json" \
  -d '{
    "command":"hold_position",
    "target_scope":"group",
    "target_ids":["squad-a"],
    "client_nonce":"ops-run-001",
    "parameters":{}
  }' | jq
```

Fetch signed audit entries:

```bash
curl -s http://localhost:8000/swarm/audit/recent?limit=20 \
  -H "Authorization: Bearer ${JOIN_API_ADMIN_TOKEN}" | jq
```

Trigger one FL round (admin-auth required):

```bash
curl -s -X POST http://localhost:8000/trigger_fl \
  -H "X-Join-Admin-Token: ${JOIN_API_ADMIN_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{}' | jq
```

## Strict Chaos Soak (CPU-Friendly)

The node-agent image uses CPU-only PyTorch wheels by default to keep local and CI chaos builds smaller.

Run strict churn validation:

```bash
JOIN_API_ADMIN_TOKEN="${JOIN_API_ADMIN_TOKEN}" \
SOAK_CHAOS_ENABLED=1 \
SOAK_CHAOS_STRICT=1 \
CHAOS_MIN_CLIENT_QUORUM=1 \
python3 tests/scripts/python/test_soak_chaos_guard.py
```

Expected success criteria:

- The suite exits `0`.
- Log includes `PASSED: FL rounds progressed under controlled churn`.

## Latency Benchmark

Run the benchmark script to collect map/command p95 and p99 latency under concurrency:

```bash
python3 tests/scripts/python/benchmark_swarm_c2_latency.py \
  --base-url http://localhost:8000 \
  --token "${JOIN_API_ADMIN_TOKEN}" \
  --iterations 200 \
  --concurrency 16
```

The script prints a single JSON summary with endpoint error counts and latency percentiles.
