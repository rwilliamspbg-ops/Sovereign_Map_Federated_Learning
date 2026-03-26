# Release Candidate Runbook: Phase3D + CDN Secrets and Operations

This document is the operator-facing release-candidate checklist for production promotion.

## Goal

Ensure all required secrets and environment wiring exist before running:

- [deploy/production/deploy-phase3d-production.sh](../deploy/production/deploy-phase3d-production.sh)
- [deploy/production/publish-frontend-cdn.sh](../deploy/production/publish-frontend-cdn.sh)
- [.github/workflows/phase3d-production-deploy.yml](../.github/workflows/phase3d-production-deploy.yml)
- Optional post-production jobs in [.github/workflows/build.yml](../.github/workflows/build.yml)

## Required GitHub Secrets

### AWS and Kubernetes (Phase3D backend)

- `AWS_DEPLOY_ROLE_ARN`
- `AWS_REGION`
- `AWS_ACCOUNT_ID` (optional if role can call STS)
- `EKS_CLUSTER_NAME`
- `K8S_NAMESPACE`
- `PHASE3D_ECR_REPOSITORY`
- `PHASE3D_ENABLE_GPU` (`true` or `false`, optional)

### Frontend CDN publish

- `FRONTEND_S3_BUCKET`
- `CLOUDFRONT_DISTRIBUTION_ID` (optional, recommended)

### Existing platform deployment secrets (already in main deploy flow)

- `DEPLOY_USER`
- `DEPLOY_KEY`
- `DEPLOY_PATH`
- `STAGING_HOST`
- `PROD_HOST`
- `MONGO_PASSWORD`
- `REDIS_PASSWORD`
- `GRAFANA_ADMIN_PASSWORD`

## Required IAM Permissions

The role in `AWS_DEPLOY_ROLE_ARN` should allow:

### ECR

- `ecr:DescribeRepositories`
- `ecr:CreateRepository`
- `ecr:GetAuthorizationToken`
- `ecr:BatchCheckLayerAvailability`
- `ecr:CompleteLayerUpload`
- `ecr:InitiateLayerUpload`
- `ecr:PutImage`
- `ecr:UploadLayerPart`

### EKS/Kubernetes bootstrap

- `eks:DescribeCluster`

### S3 (frontend publish)

- `s3:ListBucket` on target bucket
- `s3:PutObject`
- `s3:DeleteObject`

### CloudFront (optional invalidation)

- `cloudfront:CreateInvalidation`

## Preflight Commands

Run locally in a secure shell before release:

```bash
# 1) Verify shell scripts
bash -n deploy/production/deploy-phase3d-production.sh
bash -n deploy/production/publish-frontend-cdn.sh

# 2) Verify frontend build
cd frontend && npm run build --silent && cd ..

# 3) Verify Phase3D Python syntax
cd packages/training && python -m py_compile phase3d_training.py api.py && cd ../..
```

## Release Sequence

### Path A: Manual workflow

1. Run [.github/workflows/phase3d-production-deploy.yml](../.github/workflows/phase3d-production-deploy.yml)
2. Choose inputs:
   - `image_tag` = commit SHA (recommended)
   - `enable_gpu` = `true` if cluster has GPU nodes
3. Validate deployment health:

```bash
kubectl get pods -n sovereign-map -l app=phase3d-training
kubectl rollout status deployment/phase3d-training -n sovereign-map
kubectl port-forward -n sovereign-map svc/phase3d-training 5001:5001
curl http://localhost:5001/health
```

### Path B: main deploy workflow with optional jobs

1. Trigger or merge into [.github/workflows/build.yml](../.github/workflows/build.yml)
2. Ensure optional jobs are not skipped due to missing secrets
3. Confirm post-production jobs complete:
   - `deploy-phase3d-k8s`
   - `publish-frontend-cdn`

## Frontend CDN Cache Strategy

The publish script enforces:

- Static assets: `public,max-age=31536000,immutable`
- `index.html`: `no-cache,no-store,must-revalidate`

This allows aggressive caching for hashed bundles while ensuring shell HTML refresh.

## Observability Checks (Post-release)

- Backend health endpoint returns 200
- Phase3D start/status endpoints respond
- Frontend loads new build and connects to training API
- No rollout errors in Kubernetes events/logs

Recommended checks:

```bash
kubectl logs -n sovereign-map deployment/phase3d-training --tail=200
kubectl describe deployment phase3d-training -n sovereign-map
```

## Rollback Procedure

### Backend rollback

1. Re-run deploy script with previous known-good tag:

```bash
export IMAGE_TAG=<previous_sha>
./deploy/production/deploy-phase3d-production.sh
```

2. If GPU rollout fails, set:

```bash
export ENABLE_GPU=false
./deploy/production/deploy-phase3d-production.sh
```

### Frontend rollback

1. Re-publish previous dist bundle to S3
2. Invalidate CloudFront

## Sign-off Checklist

- [ ] All required secrets exist in repository settings
- [ ] IAM role can access ECR/EKS/S3/CloudFront
- [ ] `deploy-phase3d-k8s` job completes or intentionally skips
- [ ] `publish-frontend-cdn` job completes or intentionally skips
- [ ] Phase3D health and training endpoints verified
- [ ] Frontend loads and reports expected runtime status
- [ ] Rollback commands tested in staging at least once
