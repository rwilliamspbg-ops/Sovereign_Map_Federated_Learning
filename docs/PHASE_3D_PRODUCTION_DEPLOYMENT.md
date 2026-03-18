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
