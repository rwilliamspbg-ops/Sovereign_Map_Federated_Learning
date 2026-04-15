# AV-Ready v1.0 One-Pass Sprint Plan

## Purpose

Execute a single focused sprint that closes the highest-impact autonomous vehicle mapping gaps in this repository and produces a measurable AV-ready v1.0 baseline.

This sprint is scoped for one pass (10 working days) and prioritizes:

- Deterministic sensor contracts and ingestion
- Time alignment and calibration enforcement
- Real fused pose output in place of fallback paths
- End-to-end map update and tile generation loop
- CI-gated validation and operational SLOs

## Scope Boundary

### In Scope (v1.0)

- Camera + GPS + IMU ingestion as required runtime paths
- LiDAR ingestion path as optional-but-supported input
- Fused localization output (confidence-scored) for map updates
- Deterministic tile generation and cache persistence
- Metrics, alerts, and replay/integration validation

### Out of Scope (Post-v1.0)

- Full autonomous driving stack (planning/control)
- Hardware-specific deployment tuning for all edge SKUs
- Large-scale production fleet rollout policy

## One-Pass Definition of Done

AV v1.0 is complete only when all pass criteria below are met in one branch cycle:

1. Sensor contracts versioned and validated at ingest boundaries
2. Timestamp alignment enforced with out-of-order and stale handling
3. Calibration/transforms required for runtime startup
4. SLAM/fusion path emits non-fallback pose with confidence metadata
5. End-to-end pipeline runs: ingest -> pose -> map update -> tile emit
6. CI includes AV integration lane and passes on mainline merge gate
7. Observability dashboards/alerts cover ingest lag, frame drops, pose quality

## Sprint Backlog in Build Order

## 1) Sensor Contract Hardening (Days 1-2)

### Build

- Canonical contract models for camera, GPS, IMU, LiDAR payloads
- Contract version field and strict validation
- Negative-path tests for malformed payloads

### Where

- `sensors/mobile/phone_client_api.go`
- `sensors/drone/telemetry_ingest.go`
- `packages/core/src/types.ts`
- `tests/scripts/python/` (add API contract tests)

### Accept

- Invalid payloads return deterministic `4xx` with reason code
- Contract tests pass in CI

## 2) Time Alignment and Ordering Control (Days 2-3)

### Build

- Sensor timestamp normalization utility
- Alignment window (camera/GPS/IMU/LiDAR)
- Late/duplicate/future frame policy

### Where

- `sensors/mobile/phone_client_api.go`
- `sensors/drone/telemetry_ingest.go`
- `cmd/metrics-exporter/main.go` (expose alignment lag metrics)

### Accept

- Replay with intentional skew produces stable aligned output
- Alignment lag metric visible and alertable

## 3) Calibration and Transform Registry (Days 3-4)

### Build

- Calibration manifest (intrinsic/extrinsic)
- Transform chain enforcement (sensor -> vehicle -> world)
- Runtime startup guardrail for missing/invalid calibration

### Where

- `sensors/camera/frame_capture.go`
- `sensors/slam/orbslam_bridge.go`
- `storage/map_tiles/tile_encoder.go`
- `config/` (add calibration artifacts)

### Accept

- Startup fails fast on invalid calibration
- Transform consistency tests pass

## 4) Fused Pose Runtime (Replace Fallback) (Days 4-6)

### Build

- Integrate visual features + motion priors + GPS corrections
- Pose confidence score and tracking state contract
- Remove fallback-only behavior for nominal runtime path

### Where

- `sensors/slam/feature_extraction.go`
- `sensors/slam/orbslam_bridge.go`

### Accept

- Nominal operation emits non-fallback tracking state
- Pose drift and confidence metrics exported

## 5) Map Update + Tile Generation E2E (Days 6-7)

### Build

- Map update builder from fused poses and observations
- Deterministic tile encoding and cache write-through
- Provenance metadata (source timestamps, confidence, update id)

### Where

- `storage/map_tiles/tile_encoder.go`
- `storage/map_tiles/tile_cache.go`
- `storage/model_checkpoints/store.go`

### Accept

- Same input replay yields identical tile output hash
- Tile metadata includes lineage fields

## 6) Runtime Integration in Sovereign Node (Days 7-8)

### Build

- Wire sensor/mapping workers into runtime start/join flow
- Add AV service health endpoints and readiness checks

### Where

- `cmd/sovereign-node/main.go`
- `cmd/node-agent/main.go`

### Accept

- Single node can execute ingest -> pose -> tile loop unattended
- Health/readiness reflects each pipeline stage

## 7) Observability and SLO Enforcement (Days 8-9)

### Build

- Metrics: ingest rate, frame drop %, alignment lag, pose confidence, tile latency
- Alerts: sensor outage, high lag, low confidence streak, backlog growth
- Dashboard panels for AV runtime status

### Where

- `prometheus.yml`
- `fl_slo_alerts.yml`
- `fl_detailed_alerts.yml`
- `grafana/provisioning/dashboards/operations_overview.json`

### Accept

- Fault injection causes expected alerts
- Alert clear behavior verified after recovery

## 8) CI Gate and Replay Validation (Days 9-10)

### Build

- Add AV integration lane to CI
- Add deterministic replay fixtures
- Add final gate summary artifact

### Where

- `.github/workflows/opencv-go-tests.yml` (expand scope)
- `.github/workflows/full-validation-pr-gate.yml`
- `tests/scripts/python/`
- `test-results/` (artifact output)

### Accept

- AV lane required for PR merge
- Full one-pass gate report generated and archived

## Daily Execution Plan

### Day 1

- Finalize v1.0 contract schema and test matrix
- Implement first ingest validators

### Day 2

- Complete ingest validators and negative tests
- Start timestamp normalization

### Day 3

- Complete alignment engine and metrics
- Add calibration manifest and loader

### Day 4

- Enforce calibration runtime guardrails
- Begin fused pose integration

### Day 5

- Continue fused pose integration
- Add pose confidence and drift telemetry

### Day 6

- Complete fused pose integration
- Implement map update builder

### Day 7

- Complete deterministic tile path
- Integrate node runtime wiring

### Day 8

- Add health/readiness and observability rules
- Build Grafana panels

### Day 9

- Add/expand CI lanes and replay tests
- Run first full end-to-end gate

### Day 10

- Resolve failures, rerun gates, publish AV v1.0 sprint report

## Risks and Mitigations

- OpenCV runtime variance:
  - Use tagged build lane and pin known-good GoCV/OpenCV matrix
- Sensor clock skew:
  - Enforce timestamp policy and confidence degradation when skew exceeds threshold
- Integration regressions in main runtime:
  - Feature-flag AV pipeline in `cmd/sovereign-node/main.go` during rollout

## Final Deliverables

1. AV v1.0 sprint completion report in `Documentation/Reports/`
2. CI proof artifact bundle under `test-results/`
3. Updated roadmap status in `Documentation/Project/ROADMAP.md`
4. Updated release-readiness summary reflecting AV v1.0 gate status
