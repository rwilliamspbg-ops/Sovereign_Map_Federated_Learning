# Autonomy and Digital Twin Operations Guide

This runbook documents autonomy-related runtime functions, APIs, and validation flow.

## Scope

This guide covers:

- Internal autonomy packages under `internal/autonomy`
- Control-plane autonomy endpoints exposed by `sovereignmap_production_backend_v2.py`
- HUD and C2 operator integration points
- Strict chaos soak execution with CPU node-agent profile

## Internal Autonomy Functions

### Capability and Contract Modeling

File: `internal/autonomy/contracts.go`

- `SensorClass`, `ComputeClass`: canonical source/compute taxonomies.
- `SensorCapability`, `ComputeCapability`: normalized capability envelopes.
- `CapabilityMatrix`: policy-ready matrix for planner/orchestrator input.
- `Validate()`: rejects malformed capability envelopes before planning.

### Confidence-Scored Map and Twin State

File: `internal/autonomy/map_twin.go`

- `LayeredMapStore.UpsertCell`: updates occupancy/confidence cells by coordinate.
- `LayeredMapStore.DecayConfidence`: applies time-based confidence decay.
- `LayeredMapStore.Snapshot`: emits map layer summary for API/HUD use.
- Twin lag helpers: expose staleness as an explicit operational signal.

### Course-Correction Planner

File: `internal/autonomy/planner.go`

- `SelectBestCorrection(options, policy)`: safe action selection under risk/gain constraints.
- Rejects unsafe candidates via policy checks (`MaxAcceptedRisk`, `MinMissionGain`).
- Returns scored output with safety rejection indicator for operator review.

### Prediction and Compute Scheduling

File: `internal/autonomy/predictor_orchestrator.go`

- `PredictLinearMotion`: low-compute baseline short-horizon twin prediction.
- `SelectComputeTarget`: chooses edge/node/backend target based on workload profile.

### Runtime Readiness Gate

File: `internal/autonomy/runtime_readiness.go`

- `ValidateReadiness`: enforces production gates:
  - chaos test coverage
  - rollback plan
  - safety gateway
  - audit trail
  - canary strategy
  - SLO dashboards

## Control-Plane Autonomy Endpoints

Autonomy API routes live in `sovereignmap_production_backend_v2.py` and are covered in `docs/api/openapi.yaml`.

- `GET /autonomy/twin/summary`
  - Digital twin lag, confidence, coverage, risk snapshot.
- `GET /autonomy/planner/insights`
  - Candidate actions, selected action, and rejected actions with reasons.
- `GET /autonomy/sensors/quality`
  - Sensor confidence/freshness/anomaly matrix.
- `GET /autonomy/slo/status`
  - SLO compliance summary for autonomy control loops.

Example:

```bash
curl -s http://localhost:8000/autonomy/twin/summary | jq
curl -s http://localhost:8000/autonomy/planner/insights | jq
curl -s http://localhost:8000/autonomy/sensors/quality | jq
curl -s http://localhost:8000/autonomy/slo/status | jq
```

## HUD and C2 Integration

Frontend surfaces:

- `frontend/src/HUD.jsx`: autonomy KPI strip, safety state, recommendations, timeline signals.
- `frontend/src/C2SwarmHUD.jsx`: C2 overlays and operator assist workflows.

Operator validation flow:

1. Open `http://localhost:3000` for primary HUD.
2. Open `http://localhost:3000/?view=c2` for C2 map/command view.
3. Confirm autonomy summary and planner insights populate from live API data.

## CPU Node-Agent Test Profile

File: `Dockerfile.node-agent`

The node-agent now installs CPU-only PyTorch wheels to reduce build footprint in CI/dev containers:

- `torch==2.1.0+cpu`
- `torchvision==0.16.0+cpu`

This avoids downloading CUDA runtime wheels during routine validation.

## Strict Chaos Soak Workflow

### Prerequisites

- Backend, Prometheus, and node-agent up via `docker-compose.full.yml`
- Admin token configured for authenticated manual FL trigger fallback

```bash
export JOIN_API_ADMIN_TOKEN='replace-with-strong-token'
export ALLOW_INSECURE_DEV_ADMIN_TOKEN='false'
docker compose -f docker-compose.full.yml up -d backend prometheus node-agent
```

### Run Strict Chaos Guard

```bash
JOIN_API_ADMIN_TOKEN="$JOIN_API_ADMIN_TOKEN" \
SOAK_CHAOS_ENABLED=1 \
SOAK_CHAOS_STRICT=1 \
CHAOS_MIN_CLIENT_QUORUM=1 \
python3 tests/scripts/python/test_soak_chaos_guard.py
```

Expected pass condition:

- Suite completes with return code `0`.
- `sovereignmap_fl_round` increases across churn steps.

### Fallback Trigger Authentication

File: `tests/scripts/python/testnet-chaos-suite.py`

The suite now sends `X-Join-Admin-Token` from `CHAOS_ADMIN_TOKEN` or `JOIN_API_ADMIN_TOKEN` when invoking `POST /trigger_fl` fallback, aligning test execution with backend auth policy.

## Security Notes

- Keep `JOIN_API_ADMIN_TOKEN` strong in shared environments.
- Do not rely on `local-dev-admin-token` unless explicitly allowed in isolated local workflows.
- Preserve admin-token handling in CI secret stores and avoid plaintext commits.

## Related Artifacts

- `AUTONOMOUS_MAPPING_DIGITAL_TWIN_PLAN.md`
- `AUTONOMOUS_MAPPING_EXECUTION_REPORT.md`
- `HUD_AUTONOMY_IMPROVEMENT_PLAN.md`
- `docs/api/openapi.yaml`
