# API Auth Token Rotation and Role Policy Runbook

## Scope
This runbook covers safe rotation of node-agent API auth tokens and role policy updates for proof and ledger endpoints:

- `POST /api/v1/proof/verify`
- `POST /api/v1/proof/hybrid/verify`
- `GET /api/v1/ledger`

## Preconditions
- Access to deployment environment variables and secret files.
- Ability to restart node-agent containers/processes.
- Validation tooling (`curl`, `jq`, and access to Prometheus/Grafana).

## Current Defaults
- `MOHAWK_API_AUTH_MODE=file-only`
- `MOHAWK_API_TOKEN_FILE=/run/secrets/mohawk_api_token`
- `MOHAWK_API_ENFORCE_ROLES=true`
- `MOHAWK_API_PROOF_ALLOWED_ROLES=verifier,admin`

## Token Rotation Procedure
1. Generate a new token.

```bash
openssl rand -hex 32 > /tmp/new_mohawk_api_token
chmod 600 /tmp/new_mohawk_api_token
```

2. Stage the new token in your secret manager or deployment volume.

```bash
# Example for local compose secret mount
cp /tmp/new_mohawk_api_token /run/secrets/mohawk_api_token.next
chmod 600 /run/secrets/mohawk_api_token.next
```

3. Perform a controlled switchover.

```bash
mv /run/secrets/mohawk_api_token.next /run/secrets/mohawk_api_token
chmod 600 /run/secrets/mohawk_api_token
```

4. Restart or roll the node-agent workload so the file-backed token is reloaded.

```bash
docker compose restart node-agent
```

5. Validate old token is rejected and new token is accepted.

```bash
NEW_TOKEN="$(cat /run/secrets/mohawk_api_token)"

# Should fail with 401 (old token)
curl -s -o /dev/null -w "%{http_code}\n" \
  http://localhost:8082/api/v1/ledger \
  -H "Authorization: Bearer old-token" \
  -H "X-API-Role: verifier"

# Should succeed with 200 (new token)
curl -s -o /dev/null -w "%{http_code}\n" \
  http://localhost:8082/api/v1/ledger \
  -H "Authorization: Bearer ${NEW_TOKEN}" \
  -H "X-API-Role: verifier"
```

## Role Policy Update Procedure
1. Choose an explicit role allow-list for proof/ledger operations.

```bash
export MOHAWK_API_PROOF_ALLOWED_ROLES="verifier,admin,ops"
```

2. Apply config change and restart node-agent.

```bash
docker compose up -d node-agent
```

3. Validate role behavior.

```bash
TOKEN="$(cat /run/secrets/mohawk_api_token)"

# Expected 200
curl -s -o /dev/null -w "%{http_code}\n" \
  http://localhost:8082/api/v1/ledger \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-API-Role: ops"

# Expected 403
curl -s -o /dev/null -w "%{http_code}\n" \
  http://localhost:8082/api/v1/ledger \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-API-Role: observer"
```

## Rollback
- Restore previous token file from secure backup.
- Revert role env vars to last known good values.
- Restart node-agent and re-run validation checks.

## Monitoring and Alert Validation
After rotation or role updates, confirm the following Prometheus metrics are moving as expected:

- `mohawk_proof_verifications_total`
- `mohawk_proof_verification_latency_seconds`
- `mohawk_ledger_events_total`
- `mohawk_ledger_entries`

Recommended checks:

```bash
curl -s http://localhost:8082/metrics | grep -E "mohawk_proof_verifications_total|mohawk_ledger_events_total|mohawk_ledger_entries"
```

## Troubleshooting
- `401 missing api token`: client omitted `Authorization` or `X-API-Token`.
- `401 invalid api token`: token value does not match token file.
- `403 missing api role`: `X-API-Role` header not supplied while role enforcement is on.
- `403 role not allowed`: role is not in `MOHAWK_API_PROOF_ALLOWED_ROLES`.
- `500 auth configuration error`: token file path missing/unreadable/empty.
