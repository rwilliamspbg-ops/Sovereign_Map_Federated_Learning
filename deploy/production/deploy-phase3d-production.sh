#!/usr/bin/env bash
set -euo pipefail

# Production deployment automation for Phase 3D trainer:
# - Build Docker image
# - Push image to AWS ECR
# - Deploy/rollout to Kubernetes (EKS or any configured kubectl target)

AWS_REGION="${AWS_REGION:-us-east-1}"
AWS_ACCOUNT_ID="${AWS_ACCOUNT_ID:-}"
ECR_REPOSITORY="${ECR_REPOSITORY:-sovereign-map-phase3d-training}"
EKS_CLUSTER_NAME="${EKS_CLUSTER_NAME:-}"
K8S_NAMESPACE="${K8S_NAMESPACE:-sovereign-map}"
IMAGE_TAG="${IMAGE_TAG:-$(git rev-parse --short HEAD)}"
ENABLE_GPU="${ENABLE_GPU:-true}"
K8S_MANIFEST_CPU="${K8S_MANIFEST_CPU:-deploy/kubernetes/phase3d-training.yaml}"
K8S_MANIFEST_GPU="${K8S_MANIFEST_GPU:-deploy/kubernetes/phase3d-training-gpu.yaml}"

log() {
  printf "[phase3d-deploy] %s\n" "$*"
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

require_cmd aws
require_cmd docker
require_cmd kubectl
require_cmd sed

if [[ -z "${AWS_ACCOUNT_ID}" ]]; then
  AWS_ACCOUNT_ID="$(aws sts get-caller-identity --query Account --output text)"
fi

if [[ -n "${EKS_CLUSTER_NAME}" ]]; then
  log "Updating kubeconfig for EKS cluster ${EKS_CLUSTER_NAME} in ${AWS_REGION}"
  aws eks update-kubeconfig --name "${EKS_CLUSTER_NAME}" --region "${AWS_REGION}"
fi

ECR_REGISTRY="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
IMAGE_URI="${ECR_REGISTRY}/${ECR_REPOSITORY}:${IMAGE_TAG}"

log "Ensuring ECR repository exists: ${ECR_REPOSITORY}"
if ! aws ecr describe-repositories --repository-names "${ECR_REPOSITORY}" --region "${AWS_REGION}" >/dev/null 2>&1; then
  aws ecr create-repository --repository-name "${ECR_REPOSITORY}" --region "${AWS_REGION}" >/dev/null
fi

log "Logging into ECR: ${ECR_REGISTRY}"
aws ecr get-login-password --region "${AWS_REGION}" | docker login --username AWS --password-stdin "${ECR_REGISTRY}"

log "Building image: ${IMAGE_URI}"
docker build -f packages/training/Dockerfile -t "${IMAGE_URI}" .

log "Pushing image to ECR"
docker push "${IMAGE_URI}"

log "Ensuring namespace exists: ${K8S_NAMESPACE}"
kubectl get namespace "${K8S_NAMESPACE}" >/dev/null 2>&1 || kubectl create namespace "${K8S_NAMESPACE}"

if [[ "${ENABLE_GPU}" == "true" ]]; then
  SOURCE_MANIFEST="${K8S_MANIFEST_GPU}"
  log "Using GPU deployment manifest"
else
  SOURCE_MANIFEST="${K8S_MANIFEST_CPU}"
  log "Using CPU deployment manifest"
fi

TMP_MANIFEST="$(mktemp)"
sed "s|__IMAGE__|${IMAGE_URI}|g" "${SOURCE_MANIFEST}" > "${TMP_MANIFEST}"

log "Applying Kubernetes manifest"
kubectl apply -n "${K8S_NAMESPACE}" -f "${TMP_MANIFEST}"

log "Waiting for rollout"
kubectl rollout status deployment/phase3d-training -n "${K8S_NAMESPACE}" --timeout=300s

rm -f "${TMP_MANIFEST}"

log "Deployment complete"
log "Image: ${IMAGE_URI}"
log "Health check command: kubectl port-forward -n ${K8S_NAMESPACE} svc/phase3d-training 5001:5001"
