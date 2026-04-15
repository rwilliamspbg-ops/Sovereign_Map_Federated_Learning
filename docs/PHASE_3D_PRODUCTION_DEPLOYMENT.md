# Phase 3D Production Deployment (ECR + Kubernetes)

This guide deploys the Phase 3D training backend to production with:

- AWS ECR image push
- Kubernetes rollout (CPU or GPU manifest)
- Health-checked Deployment + Service

## Prerequisites

- AWS CLI configured with deploy permissions
- Docker
- kubectl
- Optional: EKS cluster (script can update kubeconfig automatically)

## Required Environment Variables

```bash
export AWS_REGION=us-east-1
export AWS_ACCOUNT_ID=<optional-account-id>
export ECR_REPOSITORY=sovereign-map-phase3d-training
export EKS_CLUSTER_NAME=<your-eks-cluster>
export K8S_NAMESPACE=sovereign-map
export IMAGE_TAG=$(git rev-parse --short HEAD)
export ENABLE_GPU=true
```

Notes:
- If AWS_ACCOUNT_ID is omitted, the script resolves it from STS.
- If ENABLE_GPU=true, the GPU manifest is used and requests 2 NVIDIA GPUs.

## One-Command Deployment

```bash
chmod +x deploy/production/deploy-phase3d-production.sh
./deploy/production/deploy-phase3d-production.sh
```

What it does:
1. Ensures the ECR repository exists
2. Builds `packages/training/Dockerfile`
3. Pushes image to ECR
4. Applies Kubernetes manifest with the pushed image URI
5. Waits for rollout completion

## Kubernetes Manifests

- CPU manifest: deploy/kubernetes/phase3d-training.yaml
- GPU manifest: deploy/kubernetes/phase3d-training-gpu.yaml

Both manifests expose:
- Deployment: phase3d-training
- Service: phase3d-training (ClusterIP, port 5001)
- Liveness/Readiness probes on /health

## Verify Deployment

```bash
kubectl get pods -n sovereign-map -l app=phase3d-training
kubectl rollout status deployment/phase3d-training -n sovereign-map
kubectl port-forward -n sovereign-map svc/phase3d-training 5001:5001
curl http://localhost:5001/health
```

## GitHub Actions Automation

Manual workflow is available at:

- .github/workflows/phase3d-production-deploy.yml

Required repository secrets:

- AWS_DEPLOY_ROLE_ARN
- AWS_REGION
- AWS_ACCOUNT_ID
- PHASE3D_ECR_REPOSITORY
- EKS_CLUSTER_NAME
- K8S_NAMESPACE

Run it from Actions with optional inputs:
- image_tag
- enable_gpu

## Real Training Defaults

The trainer now defaults to:

- dataset: cifar10
- device: auto (or cuda from frontend real mode)
- multi_gpu: true

Frontend real mode explicitly starts training with:

- dataset=cifar10
- device=cuda
- multi_gpu=true

## Operator Validation Addendum (AV + Chaos)

Use this section when running release-readiness validation for AV runtime and chaos cadence checks.

### Prerequisites

- Docker Compose v2 available (`docker compose version`)
- Python virtual environment with project test dependencies installed
- At least 20 GB free disk before image rebuild-heavy runs
- Local Prometheus endpoint reachable at `http://localhost:9090`
- Evidence directory path set and writable (example: `test-results/release-evidence/<timestamp>`)

### Recommended Validation Sequence

1. Run AV contract and pose checks (ingest, alignment, pose).
2. Bring up runtime stack with backend, node agents, and Prometheus.
3. Run chaos guard with strict cadence and quorum settings.
4. Collect backend/node-agent/prometheus logs and store exit codes in the evidence bundle.

### Expected Alerts and Signals During Chaos Validation

- `ConsensusOpenRoundsBacklog` may spike briefly during restart windows.
- `FLClientParticipationLow` may trigger if active nodes drop below quorum.
- `ChurnBurstDetected` may trigger during intentional node restarts.

These can be expected if they clear automatically after recovery. Treat sustained firing as failure requiring triage.

### Troubleshooting Checks (Operator First Response)

1. Confirm runtime containers are healthy:

```bash
docker compose -f docker-compose.full.yml ps
```

2. Confirm round/activity progression in chaos output and metrics summary:

```bash
tail -n 80 test-results/release-evidence/<timestamp>/chaos_guard_runtime.log
```

3. Confirm node agents are connected and not crash-looping:

```bash
docker compose -f docker-compose.full.yml logs --no-color --tail=200 node-agent
```

4. Confirm backend health and training loop readiness:

```bash
docker compose -f docker-compose.full.yml logs --no-color --tail=200 backend
curl -fsS http://localhost:5001/health
```

5. If cadence gate fails with `active_nodes=0`, verify quorum env values and restart timing, then rerun with preserved evidence output.

Recommended resilience setting for local churn validation when Prometheus active-node metrics are briefly stale:

```bash
export CHAOS_USE_CONTAINER_QUORUM_FALLBACK=1
```

### Runtime Profiles and Provider Policy

The backend now supports explicit runtime performance profiles and provider-aware execution policy publication.

- Profiles: `ultra_latency`, `balanced`, `throughput`
- Profile knobs: batch cadence, precision mode, retry strategy, memory backpressure limits
- Policy view: hardware class, selected provider, optimizer flags, fallback chain

Inspect active runtime configuration:

```bash
curl -fsS http://localhost:8000/runtime/profile | jq .
```

Switch profile at runtime:

```bash
curl -fsS -X POST http://localhost:8000/runtime/profile \
	-H 'Content-Type: application/json' \
	-d '{"profile":"throughput"}' | jq .
```

The same profile/policy snapshot is also exposed through `GET /metrics_summary` under:

- `runtime_profile`
- `provider_execution_policy`
- `memory_pressure`

### Memory Pressure Control-Loop Alerts

New Prometheus control-loop alerts:

- `RuntimeMemoryPressureCritical`: fires when `sovereign_runtime_memory_pressure_level >= 2` for 3m
- `RuntimeBackpressureSustained`: fires when `sovereign_runtime_backpressure_level >= 1` for 10m

These indicate adaptive throttling/offload modes are active and throughput may degrade until pressure clears.

### Current Local Limitation

Tile-stage validation may require OpenCV-enabled build tags in local environments. If OpenCV-tagged modules are excluded, treat tile stage as a tracked gap for release readiness, not a silent pass.
