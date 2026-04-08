<!-- markdownlint-disable MD022 MD032 -->

# Quantum KEX Rotation Drill Runbook (Genesis Testnet)

## Purpose
This runbook defines a public, repeatable post-quantum key exchange (KEX) rotation drill on Genesis Testnet.

The drill demonstrates three things in one auditable sequence:
1. Transport security controls execute a key rotation path successfully.
2. Proof verification and ledger writes remain healthy before and after rotation.
3. Operators can publish a concrete evidence bundle for community and investor review.

## Scope
- Environment: Genesis Testnet
- Runtime surface: node-agent auth-gated proof and ledger endpoints
- Evidence source: local artifact bundle produced by the drill script

## Controls Exercised
- Session key rotation pathway in crypto transport tests:
  - TestRotateSessionKeyNoDeadlock
  - TestRotateSessionKeyReestablishesSharedSecret
  - TestHandshakeVerification
- Hybrid verification continuity:
  - POST /api/v1/proof/hybrid/verify
- Fallback backend rehearsal:
  - POST /api/v1/proof/hybrid/verify with `stark_backend=winterfell_mock`
- Ledger integrity continuity:
  - GET /api/v1/ledger
  - GET /api/v1/ledger/reconcile
- Role policy enforcement negative-path:
  - GET /api/v1/ledger with unauthorized role (expects 401/403)

## Preconditions
- Genesis Testnet stack reachable, with node-agent endpoint available.
- API token available via one of:
  - MOHAWK_API_TOKEN environment variable
  - MOHAWK_API_TOKEN_FILE (default /run/secrets/mohawk_api_token)
- Tooling: curl, go, python3

## Execution
### Recommended one-command path

```bash
make quantum-kex-rotation-drill
```

### Strict production guardrail path

```bash
ENFORCE_NON_MOCK_BACKEND=true \
FALLBACK_STARK_BACKEND=external_cmd \
MOHAWK_STARK_VERIFY_CMD="/usr/local/bin/real-stark-verifier" \
bash scripts/quantum-kex-rotation-drill.sh
```

### Direct script path

```bash
NODE_AGENT_BASE_URL=http://localhost:8082 \
MOHAWK_API_TOKEN_FILE=/run/secrets/mohawk_api_token \
bash scripts/quantum-kex-rotation-drill.sh
```

### Optional explicit token override

```bash
NODE_AGENT_BASE_URL=http://localhost:8082 \
MOHAWK_API_TOKEN="<redacted-token>" \
bash scripts/quantum-kex-rotation-drill.sh
```

## Artifact Output
Default output directory:

```text
artifacts/quantum-kex-rotation/<drill-id>/
```

Expected files:
- readiness_pre.json
- ledger_pre.json
- hybrid_verify_pre_rotation.json
- crypto_rotation_test.log
- hybrid_verify_post_rotation.json
- hybrid_verify_fallback_backend.json
- role_failure_negative_response.txt
- role_failure_negative_test.json
- ledger_post.json
- ledger_reconcile_post.json
- retention-policy.json
- immutability-notice.txt
- checksums.sha256
- drill-summary.json
- drill-summary.md

Generated cross-run index files:
- artifacts/quantum-kex-rotation/public-drill-index.json
- artifacts/quantum-kex-rotation/public-drill-index.md

PQC readiness evidence file:
- pqc-readiness-evidence.json

## Success Criteria
- Readiness pre-check is true.
- Hybrid verification accepted both before and after rotation tests.
- Fallback backend rehearsal is accepted.
- Unauthorized role negative test returns 401/403.
- Ledger reconciliation healthy after the drill.
- Ledger entry count increases by at least 2 during the drill.
- All artifact files present and readable.
- For production-mode claims, `production_pqc_ready` must be `true` and non-mock backend enforcement must be enabled.

## Public Disclosure Template
Use the generated drill summary values and publish a concise statement:

```text
Genesis Testnet Quantum KEX Rotation Drill complete.

- Drill ID: <drill-id>
- Window (UTC): <start> -> <end>
- Pre/Post hybrid verification accepted: true/true
- Ledger reconcile healthy after drill: true
- Ledger entries added: <n>

Evidence bundle path: artifacts/quantum-kex-rotation/<drill-id>/
```

## Operational Notes
- This drill is non-destructive and uses existing verification endpoints.
- If auth fails, recheck token source and X-API-Role permissions.
- If reconciliation fails, stop public messaging and open incident triage before rerun.
- `winterfell_mock` remains valid for testnet rehearsal only; production posture requires strict non-mock backend enforcement.

## Retention and Compliance
- Canonical policy: [QUANTUM_KEX_DRILL_RETENTION_POLICY.md](QUANTUM_KEX_DRILL_RETENTION_POLICY.md)
- Treat each bundle as immutable security evidence for at least 2555 days.
- Validate `checksums.sha256` before and after archival transfer.

## Suggested Cadence Through 2027 Epoch Deadline
- Public drill cadence: monthly
- Add an additional ad hoc drill after any transport/auth policy change
- Keep the last 12 drill summaries in release and governance reporting channels
