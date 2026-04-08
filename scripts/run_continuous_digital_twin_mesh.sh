#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8000}"
TPM_URL="${TPM_URL:-http://localhost:9091/event/attestation}"
ADMIN_TOKEN="${JOIN_API_ADMIN_TOKEN:-local-dev-admin-token}"

NODE_LIDAR="${NODE_LIDAR:-lidar-node-continuous}"
NODE_GPS="${NODE_GPS:-gps-node-continuous}"
NODE_IMAGE="${NODE_IMAGE:-image-node-continuous}"

log() {
  printf '[%s] %s\n' "$(date -u +'%Y-%m-%dT%H:%M:%SZ')" "$*"
}

post_json() {
  local url="$1"
  local payload="$2"
  curl -sS -o /dev/null -X POST "$url" -H 'Content-Type: application/json' -d "$payload"
}

post_json_admin() {
  local url="$1"
  local payload="$2"
  curl -sS -o /dev/null -X POST "$url" \
    -H 'Content-Type: application/json' \
    -H "X-Join-Admin-Token: ${ADMIN_TOKEN}" \
    -d "$payload"
}

ensure_continuous_training() {
  post_json_admin "${BASE_URL}/training/start" '{"rounds":0}' || true
}

ensure_offer() {
  local seller="$1"
  local modality="$2"
  local title="$3"
  local schema="$4"
  local count="$5"
  local qscore="$6"

  local offer_list
  offer_list="$(curl -sS "${BASE_URL}/marketplace/offers?seller_node_id=${seller}&limit=10")"
  local existing_id
  existing_id="$(jq -r '.offers[0].offer_id // empty' <<<"${offer_list}")"

  if [[ -n "${existing_id}" ]]; then
    post_json_admin "${BASE_URL}/marketplace/offers/${existing_id}" "{\"status\":\"active\",\"quality_score\":${qscore},\"sample_count\":${count}}" || true
  else
    post_json_admin "${BASE_URL}/marketplace/offers" "{\"seller_node_id\":\"${seller}\",\"dataset_fingerprint\":\"fp-${modality}-continuous\",\"title\":\"${title}\",\"description\":\"Continuous ${modality} stream for digital twin mesh\",\"modality\":\"${modality}\",\"label_schema\":\"${schema}\",\"sample_count\":${count},\"quality_score\":${qscore},\"privacy_profile\":\"dp-ready\",\"allowed_tasks\":[\"digital_twin\",\"mapping\"],\"attestation_status\":\"verified\",\"price_per_round\":2.0,\"min_rounds\":1,\"status\":\"active\"}" || true
  fi
}

emit_attestation() {
  local participant="$1"
  local node_name="$2"
  local ctype="$3"
  local capacity="$4"
  local score="$5"

  local ts
  ts="$(date +%s)"
  post_json "${BASE_URL}/attestations/share" "{\"participant_name\":\"${participant}\",\"node_name\":\"${node_name}\",\"compute_type\":\"${ctype}\",\"compute_capacity\":\"${capacity}\",\"capacity_score\":${score},\"attestation_status\":\"verified\",\"region\":\"zone-a\",\"proof_digest\":\"sha256:${ctype}-${ts}\",\"notes\":\"continuous ${ctype} contribution\"}" || true
}

emit_tpm_event() {
  local node_id="$1"
  local latency="$2"
  post_json "${TPM_URL}" "{\"node_id\":${node_id},\"result\":\"success\",\"latency_ms\":${latency},\"trust_score\":0.99}" || true
}

emit_mobile_verify_noise() {
  post_json "${BASE_URL}/mobile/verify_gradient" '{"gradient_payload_b64":"aGVsbG8=","signature_b64":"aW52YWxpZA==","public_key_pem":"bad"}' || true
}

main() {
  log "Starting continuous digital twin mesh generator"
  ensure_continuous_training
  ensure_offer "${NODE_LIDAR}" "lidar" "Lidar Continuous Stream" "3d_points" 12000 0.95
  ensure_offer "${NODE_GPS}" "gps" "GPS Continuous Stream" "lat_lon_time" 15000 0.91
  ensure_offer "${NODE_IMAGE}" "image" "Image Continuous Stream" "rgb_frame" 18000 0.93

  while true; do
    emit_attestation "${NODE_LIDAR}" "Lidar Edge Continuous" "lidar" "32ch-lidar-stream" "0.92"
    emit_attestation "${NODE_GPS}" "GPS Edge Continuous" "gps" "rtk-gps-10hz" "0.88"
    emit_attestation "${NODE_IMAGE}" "Image Edge Continuous" "image" "4k-frames-5fps" "0.90"

    emit_tpm_event 1 32
    emit_tpm_event 2 41
    emit_tpm_event 3 37

    emit_mobile_verify_noise

    curl -sS -o /dev/null "${BASE_URL}/trust_snapshot" || true
    curl -sS -o /dev/null "${BASE_URL}/ops/events/recent?limit=5" || true

    delay="$(( (RANDOM % 6) + 5 ))"
    log "Tick complete; next update in ${delay}s"
    sleep "${delay}"
  done
}

main
