#!/usr/bin/env bash
set -euo pipefail

# Publish built frontend assets to S3 + optional CloudFront invalidation.

AWS_REGION="${AWS_REGION:-us-east-1}"
FRONTEND_DIST_DIR="${FRONTEND_DIST_DIR:-frontend/dist}"
FRONTEND_S3_BUCKET="${FRONTEND_S3_BUCKET:-}"
CLOUDFRONT_DISTRIBUTION_ID="${CLOUDFRONT_DISTRIBUTION_ID:-}"

log() {
  printf "[frontend-cdn] %s\n" "$*"
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

require_cmd aws

if [[ -z "${FRONTEND_S3_BUCKET}" ]]; then
  echo "FRONTEND_S3_BUCKET is required" >&2
  exit 1
fi

if [[ ! -d "${FRONTEND_DIST_DIR}" ]]; then
  echo "Frontend dist directory not found: ${FRONTEND_DIST_DIR}" >&2
  exit 1
fi

log "Uploading versioned/static assets with long cache TTL"
aws s3 sync "${FRONTEND_DIST_DIR}/" "s3://${FRONTEND_S3_BUCKET}/" \
  --region "${AWS_REGION}" \
  --delete \
  --exclude "index.html" \
  --cache-control "public,max-age=31536000,immutable"

log "Uploading index.html with no-cache policy"
aws s3 cp "${FRONTEND_DIST_DIR}/index.html" "s3://${FRONTEND_S3_BUCKET}/index.html" \
  --region "${AWS_REGION}" \
  --cache-control "no-cache,no-store,must-revalidate" \
  --content-type "text/html; charset=utf-8"

if [[ -n "${CLOUDFRONT_DISTRIBUTION_ID}" ]]; then
  log "Creating CloudFront invalidation"
  aws cloudfront create-invalidation \
    --distribution-id "${CLOUDFRONT_DISTRIBUTION_ID}" \
    --paths "/*" >/dev/null
  log "CloudFront invalidation requested"
else
  log "CLOUDFRONT_DISTRIBUTION_ID not set; skipping invalidation"
fi

log "Frontend CDN publish complete"
