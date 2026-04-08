#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TOOLROOT_DEFAULT="/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.7.linux-amd64"
TOOLROOT="${TOOLROOT:-${TOOLROOT_DEFAULT}}"
ARTIFACTS_BASE_DIR="${ROOT_DIR}/artifacts/quantum-kex-rotation"

NODE_AGENT_BASE_URL="${NODE_AGENT_BASE_URL:-http://localhost:8082}"
MOHAWK_API_TOKEN="${MOHAWK_API_TOKEN:-}"
MOHAWK_API_TOKEN_FILE="${MOHAWK_API_TOKEN_FILE:-/run/secrets/mohawk_api_token}"
MOHAWK_API_ROLE="${MOHAWK_API_ROLE:-verifier}"
MOHAWK_NEGATIVE_TEST_ROLE="${MOHAWK_NEGATIVE_TEST_ROLE:-observer}"
DRILL_MODE="${DRILL_MODE:-testnet}"
FALLBACK_STARK_BACKEND="${FALLBACK_STARK_BACKEND:-winterfell_mock}"
ENFORCE_NON_MOCK_BACKEND="${ENFORCE_NON_MOCK_BACKEND:-false}"
PQC_KEM_SUITE="${PQC_KEM_SUITE:-declared-hybrid-target:x25519+mlkem768}"
PQC_SIGNATURE_SUITE="${PQC_SIGNATURE_SUITE:-declared-target:mldsa65}"
DRILL_ID="${DRILL_ID:-kex-rotation-$(date -u +%Y%m%dT%H%M%SZ)}"
ARTIFACT_ROOT="${ARTIFACT_ROOT:-${ARTIFACTS_BASE_DIR}/${DRILL_ID}}"
DRILL_RETENTION_DAYS="${DRILL_RETENTION_DAYS:-2555}"

READINESS_URL="${NODE_AGENT_BASE_URL}/api/v1/readiness"
LEDGER_URL="${NODE_AGENT_BASE_URL}/api/v1/ledger"
RECONCILE_URL="${NODE_AGENT_BASE_URL}/api/v1/ledger/reconcile"
HYBRID_VERIFY_URL="${NODE_AGENT_BASE_URL}/api/v1/proof/hybrid/verify"

mkdir -p "${ARTIFACT_ROOT}"

die() {
	echo "[error] $*" >&2
	exit 1
}

info() {
	echo "[info] $*"
}

require_cmd() {
	command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"
}

is_mock_backend() {
	local backend_name
	backend_name="$(echo "$1" | tr '[:upper:]' '[:lower:]')"
	case "${backend_name}" in
		*mock*|simulated_*|test_*)
			return 0
			;;
		*)
			return 1
			;;
	esac
}

expect_bool_true() {
	local actual="$1"
	local error_message="$2"
	[[ "${actual}" == "true" ]] || die "${error_message}"
}

resolve_api_token() {
	if [[ -n "${MOHAWK_API_TOKEN}" ]]; then
		printf "%s" "${MOHAWK_API_TOKEN}"
		return 0
	fi
	if [[ -f "${MOHAWK_API_TOKEN_FILE}" ]]; then
		tr -d '\n\r' <"${MOHAWK_API_TOKEN_FILE}"
		return 0
	fi
	die "no API token available; set MOHAWK_API_TOKEN or MOHAWK_API_TOKEN_FILE"
}

curl_auth_json() {
	local method="$1"
	local url="$2"
	local output_file="$3"
	local body="${4:-}"

	if [[ -n "${body}" ]]; then
		curl -fsS -X "${method}" "${url}" \
			-H "Authorization: Bearer ${API_TOKEN}" \
			-H "X-API-Role: ${MOHAWK_API_ROLE}" \
			-H "Content-Type: application/json" \
			--data "${body}" \
			>"${output_file}"
	else
		curl -fsS -X "${method}" "${url}" \
			-H "Authorization: Bearer ${API_TOKEN}" \
			-H "X-API-Role: ${MOHAWK_API_ROLE}" \
			>"${output_file}"
	fi
}

json_field() {
	local file_path="$1"
	local expression="$2"
	python3 - "$file_path" "$expression" <<'PY'
import json
import sys

path = sys.argv[1]
expr = sys.argv[2]
with open(path, "r", encoding="utf-8") as f:
	data = json.load(f)

parts = [p for p in expr.split(".") if p]
cur = data
for p in parts:
	if isinstance(cur, dict):
		cur = cur.get(p)
	else:
		cur = None
		break

if isinstance(cur, bool):
	print("true" if cur else "false")
elif cur is None:
	print("")
else:
	print(cur)
PY
}

require_cmd curl
require_cmd python3
require_cmd sha256sum

GO_CMD=(go)
if [[ -x "${TOOLROOT}/bin/go" ]]; then
	GO_CMD=(env GOROOT="${TOOLROOT}" GOTOOLCHAIN=local "${TOOLROOT}/bin/go")
else
	require_cmd go
fi

API_TOKEN="$(resolve_api_token)"
[[ -n "${API_TOKEN}" ]] || die "resolved API token is empty"

if [[ "${DRILL_MODE}" == "production" ]]; then
	ENFORCE_NON_MOCK_BACKEND=true
fi

START_TS="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
info "drill id: ${DRILL_ID}"
info "artifact root: ${ARTIFACT_ROOT}"
info "target node-agent: ${NODE_AGENT_BASE_URL}"
info "drill mode: ${DRILL_MODE}"
info "fallback stark backend: ${FALLBACK_STARK_BACKEND}"
info "enforce non-mock backend: ${ENFORCE_NON_MOCK_BACKEND}"

curl -fsS "${READINESS_URL}" >"${ARTIFACT_ROOT}/readiness_pre.json"
READY_PRE="$(json_field "${ARTIFACT_ROOT}/readiness_pre.json" "ready")"
expect_bool_true "${READY_PRE}" "readiness pre-check failed"

curl_auth_json GET "${LEDGER_URL}" "${ARTIFACT_ROOT}/ledger_pre.json"
LEDGER_PRE_COUNT="$(json_field "${ARTIFACT_ROOT}/ledger_pre.json" "count")"

HYBRID_PAYLOAD_FRI="$(python3 - <<'PY'
import base64
import hashlib
import json

transcript = b"quantum-kex-drill-transcript-v1-20260408"
root = hashlib.sha256(transcript).digest()
stark_proof = root + transcript
payload = {
	"mode": "any",
	"encoding": "base64",
	"snark_proof": base64.b64encode(b"invalid-snark-placeholder").decode("ascii"),
	"stark_proof": base64.b64encode(stark_proof).decode("ascii"),
}
print(json.dumps(payload))
PY
)"

curl_auth_json POST "${HYBRID_VERIFY_URL}" "${ARTIFACT_ROOT}/hybrid_verify_pre_rotation.json" "${HYBRID_PAYLOAD_FRI}"
PRE_ACCEPTED="$(json_field "${ARTIFACT_ROOT}/hybrid_verify_pre_rotation.json" "accepted")"
expect_bool_true "${PRE_ACCEPTED}" "pre-rotation hybrid verify failed"

pushd "${ROOT_DIR}" >/dev/null
"${GO_CMD[@]}" test ./internal/crypto -count=1 -run 'TestRotateSessionKeyNoDeadlock|TestRotateSessionKeyReestablishesSharedSecret|TestHandshakeVerification' -v \
	| tee "${ARTIFACT_ROOT}/crypto_rotation_test.log"
popd >/dev/null

curl_auth_json POST "${HYBRID_VERIFY_URL}" "${ARTIFACT_ROOT}/hybrid_verify_post_rotation.json" "${HYBRID_PAYLOAD_FRI}"
POST_ACCEPTED="$(json_field "${ARTIFACT_ROOT}/hybrid_verify_post_rotation.json" "accepted")"
expect_bool_true "${POST_ACCEPTED}" "post-rotation hybrid verify failed"

# Fallback backend rehearsal to prove resilience against backend-specific failures.
HYBRID_PAYLOAD_WINTERFELL="$(python3 - <<'PY'
import base64
import hashlib
import json

domain_sep = b"winterfell-v1:"
transcript = b"winterfell-fallback-drill-transcript-v1-20260408-extended-window"
root = hashlib.sha256(domain_sep + transcript).digest()
stark_proof = root + transcript
backend = "${FALLBACK_STARK_BACKEND}"
payload = {
	"mode": "any",
	"encoding": "base64",
	"stark_backend": backend,
	"snark_proof": base64.b64encode(b"invalid-snark-placeholder").decode("ascii"),
	"stark_proof": base64.b64encode(stark_proof).decode("ascii"),
}
print(json.dumps(payload))
PY
)"

curl_auth_json POST "${HYBRID_VERIFY_URL}" "${ARTIFACT_ROOT}/hybrid_verify_fallback_backend.json" "${HYBRID_PAYLOAD_WINTERFELL}"
FALLBACK_ACCEPTED="$(json_field "${ARTIFACT_ROOT}/hybrid_verify_fallback_backend.json" "accepted")"
FALLBACK_BACKEND="$(json_field "${ARTIFACT_ROOT}/hybrid_verify_fallback_backend.json" "backend")"
expect_bool_true "${FALLBACK_ACCEPTED}" "fallback backend rehearsal failed"

if [[ "${ENFORCE_NON_MOCK_BACKEND}" == "true" ]] && is_mock_backend "${FALLBACK_BACKEND}"; then
	die "non-mock backend enforcement enabled, but backend '${FALLBACK_BACKEND}' is mock/simulated"
fi

# Negative role-failure check proves policy enforcement (expects 401/403).
NEGATIVE_STATUS="$(curl -sS -o "${ARTIFACT_ROOT}/role_failure_negative_response.txt" -w '%{http_code}' \
	"${LEDGER_URL}" \
	-H "Authorization: Bearer ${API_TOKEN}" \
	-H "X-API-Role: ${MOHAWK_NEGATIVE_TEST_ROLE}")"
case "${NEGATIVE_STATUS}" in
	401|403)
		NEGATIVE_ENFORCED=true
		;;
	*)
		NEGATIVE_ENFORCED=false
		die "negative role-failure test expected 401/403, got ${NEGATIVE_STATUS}"
		;;
esac

cat >"${ARTIFACT_ROOT}/role_failure_negative_test.json" <<EOF
{
  "role": "${MOHAWK_NEGATIVE_TEST_ROLE}",
  "status_code": ${NEGATIVE_STATUS},
  "enforced": ${NEGATIVE_ENFORCED}
}
EOF

curl_auth_json GET "${LEDGER_URL}" "${ARTIFACT_ROOT}/ledger_post.json"
curl_auth_json GET "${RECONCILE_URL}" "${ARTIFACT_ROOT}/ledger_reconcile_post.json"
LEDGER_POST_COUNT="$(json_field "${ARTIFACT_ROOT}/ledger_post.json" "count")"
RECONCILE_HEALTHY="$(json_field "${ARTIFACT_ROOT}/ledger_reconcile_post.json" "healthy")"

expect_bool_true "${RECONCILE_HEALTHY}" "ledger reconciliation unhealthy after drill"

ENTRY_DELTA=0
if [[ -n "${LEDGER_PRE_COUNT}" && -n "${LEDGER_POST_COUNT}" ]]; then
	ENTRY_DELTA=$((LEDGER_POST_COUNT - LEDGER_PRE_COUNT))
fi

if [[ "${ENTRY_DELTA}" -lt 2 ]]; then
	die "expected at least 2 ledger entries added by drill, got ${ENTRY_DELTA}"
fi

END_TS="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
DRILL_OUTCOME="pass"
PQC_PRIMITIVES_WIRED=false
PQC_PRODUCTION_READY=false
if [[ "${ENFORCE_NON_MOCK_BACKEND}" == "true" ]] && ! is_mock_backend "${FALLBACK_BACKEND}"; then
	PQC_PRODUCTION_READY=true
fi

cat >"${ARTIFACT_ROOT}/pqc-readiness-evidence.json" <<EOF
{
	"drill_mode": "${DRILL_MODE}",
	"pqc_kem_suite": "${PQC_KEM_SUITE}",
	"pqc_signature_suite": "${PQC_SIGNATURE_SUITE}",
	"primitives_wired_in_runtime": ${PQC_PRIMITIVES_WIRED},
	"fallback_backend": "${FALLBACK_BACKEND}",
	"enforce_non_mock_backend": ${ENFORCE_NON_MOCK_BACKEND},
	"production_pqc_ready": ${PQC_PRODUCTION_READY},
	"notes": "production_pqc_ready is true only when non-mock backend enforcement is enabled and the runtime backend is non-mock"
}
EOF

cat >"${ARTIFACT_ROOT}/retention-policy.json" <<EOF
{
  "policy_version": "1.0",
  "artifact_classification": "security-compliance-evidence",
  "retention_days": ${DRILL_RETENTION_DAYS},
  "immutability_required": true,
  "recommended_controls": [
    "write-once object lock",
    "tamper-evident checksum verification",
    "cross-region replicated backup"
  ]
}
EOF

cat >"${ARTIFACT_ROOT}/immutability-notice.txt" <<EOF
IMMUTABILITY REQUIRED
- Move this evidence bundle to immutable storage within 24h.
- Enable object lock / WORM retention for at least ${DRILL_RETENTION_DAYS} days.
- Verify checksums.sha256 before and after transfer.
EOF

(
	cd "${ARTIFACT_ROOT}"
	find . -maxdepth 1 -type f ! -name 'checksums.sha256' -printf '%f\n' | sort | while IFS= read -r item; do
		sha256sum "${item}"
	done >checksums.sha256
)

cat >"${ARTIFACT_ROOT}/drill-summary.json" <<EOF
{
  "drill_id": "${DRILL_ID}",
  "network": "genesis-testnet",
  "objective": "public quantum kex rotation drill",
  "started_at": "${START_TS}",
  "completed_at": "${END_TS}",
  "node_agent_base_url": "${NODE_AGENT_BASE_URL}",
  "pre_rotation": {
	"readiness": ${READY_PRE},
	"hybrid_verify_accepted": ${PRE_ACCEPTED},
	"ledger_count": ${LEDGER_PRE_COUNT}
  },
  "post_rotation": {
	"hybrid_verify_accepted": ${POST_ACCEPTED},
	"ledger_count": ${LEDGER_POST_COUNT},
	"ledger_reconcile_healthy": ${RECONCILE_HEALTHY}
  },
  "fallback_backend_rehearsal": {
	"backend": "${FALLBACK_BACKEND}",
	"accepted": ${FALLBACK_ACCEPTED}
  },
  "pqc_readiness": {
	"drill_mode": "${DRILL_MODE}",
	"pqc_kem_suite": "${PQC_KEM_SUITE}",
	"pqc_signature_suite": "${PQC_SIGNATURE_SUITE}",
	"primitives_wired_in_runtime": ${PQC_PRIMITIVES_WIRED},
	"production_pqc_ready": ${PQC_PRODUCTION_READY}
	},
  "negative_role_failure_test": {
	"role": "${MOHAWK_NEGATIVE_TEST_ROLE}",
	"status_code": ${NEGATIVE_STATUS},
	"enforced": ${NEGATIVE_ENFORCED}
  },
  "retention": {
	"retention_days": ${DRILL_RETENTION_DAYS},
	"immutability_required": true,
	"checksum_catalog": "checksums.sha256"
  },
  "delta": {
	"ledger_entries_added": ${ENTRY_DELTA}
  },
  "outcome": "${DRILL_OUTCOME}",
  "evidence_files": [
	"readiness_pre.json",
	"ledger_pre.json",
	"hybrid_verify_pre_rotation.json",
	"crypto_rotation_test.log",
	"hybrid_verify_post_rotation.json",
	"hybrid_verify_fallback_backend.json",
	"pqc-readiness-evidence.json",
	"role_failure_negative_response.txt",
	"role_failure_negative_test.json",
	"ledger_post.json",
	"ledger_reconcile_post.json",
	"retention-policy.json",
	"immutability-notice.txt",
	"checksums.sha256",
	"drill-summary.json",
	"drill-summary.md"
  ]
}
EOF

cat >"${ARTIFACT_ROOT}/drill-summary.md" <<EOF
# Quantum KEX Rotation Drill Summary

- Drill ID: ${DRILL_ID}
- Network: genesis-testnet
- Start (UTC): ${START_TS}
- End (UTC): ${END_TS}
- Node-agent endpoint: ${NODE_AGENT_BASE_URL}

## Results

- Pre-rotation readiness: ${READY_PRE}
- Pre-rotation hybrid verify accepted: ${PRE_ACCEPTED}
- Post-rotation hybrid verify accepted: ${POST_ACCEPTED}
- Fallback backend rehearsal (${FALLBACK_BACKEND}) accepted: ${FALLBACK_ACCEPTED}
- Negative role-failure status (${MOHAWK_NEGATIVE_TEST_ROLE}): ${NEGATIVE_STATUS}
- Drill mode: ${DRILL_MODE}
- Declared PQC KEM suite: ${PQC_KEM_SUITE}
- Declared PQC signature suite: ${PQC_SIGNATURE_SUITE}
- PQC primitives wired in runtime: ${PQC_PRIMITIVES_WIRED}
- Production PQC ready: ${PQC_PRODUCTION_READY}
- Ledger reconcile healthy after drill: ${RECONCILE_HEALTHY}
- Ledger entry count before: ${LEDGER_PRE_COUNT}
- Ledger entry count after: ${LEDGER_POST_COUNT}
- Ledger entries added by drill: ${ENTRY_DELTA}

## Compliance Controls

- Retention days: ${DRILL_RETENTION_DAYS}
- Immutability required: true
- Checksum catalog: ${ARTIFACT_ROOT}/checksums.sha256

## Evidence Bundle

Artifacts are stored in: ${ARTIFACT_ROOT}
EOF

python3 - "${ARTIFACTS_BASE_DIR}" <<'PY'
import json
import os
import pathlib
from datetime import datetime, timezone

base = pathlib.Path(os.sys.argv[1])
summaries = []
for child in sorted(base.iterdir()):
	if not child.is_dir():
		continue
	summary_path = child / "drill-summary.json"
	if not summary_path.exists():
		continue
	try:
		data = json.loads(summary_path.read_text(encoding="utf-8"))
	except Exception:
		continue
	summaries.append(
		{
			"drill_id": data.get("drill_id", child.name),
			"started_at": data.get("started_at", ""),
			"completed_at": data.get("completed_at", ""),
			"outcome": data.get("outcome", "unknown"),
			"fallback_backend": (data.get("fallback_backend_rehearsal") or {}).get("backend", ""),
			"negative_role_status": (data.get("negative_role_failure_test") or {}).get("status_code", ""),
			"path": str(child.relative_to(base.parent)),
		}
	)

def sort_key(entry):
	return entry.get("completed_at", "")

summaries.sort(key=sort_key, reverse=True)
latest = summaries[:12]

index_json = {
	"updated_at": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
	"max_entries": 12,
	"entries": latest,
}
(base / "public-drill-index.json").write_text(json.dumps(index_json, indent=2) + "\n", encoding="utf-8")

lines = [
	"# Quantum KEX Public Drill Index",
	"",
	"Last 12 recorded drills (most recent first).",
	"",
	"| Drill ID | Completed (UTC) | Outcome | Fallback Backend | Negative Role Status | Evidence Path |",
	"|---|---|---|---|---|---|",
]
for item in latest:
	lines.append(
		"| {drill_id} | {completed_at} | {outcome} | {fallback_backend} | {negative_role_status} | {path} |".format(**item)
	)
if not latest:
	lines.append("| (none) | - | - | - | - | - |")

(base / "public-drill-index.md").write_text("\n".join(lines) + "\n", encoding="utf-8")
PY

info "public index updated: ${ARTIFACTS_BASE_DIR}/public-drill-index.md"

info "drill complete"
info "summary: ${ARTIFACT_ROOT}/drill-summary.md"
