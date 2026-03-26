# HTTP API Examples

Concrete request/response commands for local development and smoke verification.

## Base URL

- `http://localhost:8000`

## Health and Runtime Identity

```bash
curl -s http://localhost:8000/status | jq
curl -s http://localhost:8000/health | jq
curl -s http://localhost:8000/ops/health | jq
```

## Trigger and Verify a Training Round

```bash
curl -s -X POST http://localhost:8000/trigger_fl | jq
curl -s http://localhost:8000/training/status | jq
curl -s http://localhost:8000/metrics_summary | jq '.federated_learning'
curl -s http://localhost:8000/convergence | jq '.current_round, .current_accuracy, .current_loss'
```

## HUD and Governance Surfaces

```bash
curl -s http://localhost:8000/hud_data | jq
curl -s http://localhost:8000/founders | jq
curl -s http://localhost:8000/trust_snapshot | jq
```

## Update Verification Policy

```bash
curl -s -X POST http://localhost:8000/verification_policy \
  -H 'Content-Type: application/json' \
  -H 'X-API-Role: admin' \
  -d '{
    "require_proof": true,
    "min_confidence_bps": 7500,
    "reject_on_verification_failure": true,
    "allow_consensus_proof": true,
    "allow_zk_proof": true,
    "allow_tee_proof": true
  }' | jq
```

## Recent Operations Events

```bash
curl -s 'http://localhost:8000/ops/events/recent?limit=20' | jq
```

## Model Registry Snapshot

```bash
curl -s 'http://localhost:8000/model_registry?limit=20' | jq
```

## Local Marketplace Flow (Data Offers for Training Rounds)

Marketplace endpoints return deterministic error payloads:

```json
{
  "error": "no_compatible_offers_found",
  "code": "no_compatible_offers_found",
  "message": "no offers satisfied policy, quality, modality, and budget constraints",
  "details": {
    "offers_evaluated": 3,
    "compatible_offers": 1,
    "budget_rejected": 1,
    "rejection_reasons": {
      "offer_expired": 1,
      "modality_mismatch": 1
    }
  }
}
```

Create a data offer:

```bash
curl -s -X POST http://localhost:8000/marketplace/offers \
  -H 'Content-Type: application/json' \
  -d '{
    "seller_node_id": "node-17",
    "dataset_fingerprint": "sha256:demo-cifar10-node17",
    "title": "CIFAR-10 Retail Cameras",
    "modality": "image",
    "quality_score": 0.82,
    "allowed_tasks": ["classification"],
    "price_per_round": 12.5,
    "min_rounds": 3,
    "attestation_status": "verified"
  }' | jq
```

List active offers:

```bash
curl -s 'http://localhost:8000/marketplace/offers?status=active&limit=20' | jq
```

Create a round intent (buyer demand):

```bash
curl -s -X POST http://localhost:8000/marketplace/round_intents \
  -H 'Content-Type: application/json' \
  -d '{
    "model_owner_id": "aggregator-main",
    "task_type": "classification",
    "required_modalities": ["image"],
    "min_quality_score": 0.7,
    "budget_total": 100
  }' | jq
```

List round intents:

```bash
curl -s 'http://localhost:8000/marketplace/round_intents?status=open' | jq
```

Update an existing round intent:

```bash
curl -s -X PATCH http://localhost:8000/marketplace/round_intents/intent-REPLACE_ME \
  -H 'Content-Type: application/json' \
  -d '{"status": "cancelled"}' | jq
```

Create a local match contract:

```bash
curl -s -X POST http://localhost:8000/marketplace/match \
  -H 'Content-Type: application/json' \
  -d '{"round_intent_id": "intent-REPLACE_ME", "max_offers": 3}' | jq
```

Release local escrow once the round completes:

```bash
curl -s -X POST http://localhost:8000/marketplace/escrow/release \
  -H 'Content-Type: application/json' \
  -d '{"contract_id": "contract-REPLACE_ME"}' | jq
```

Run a training round and inspect marketplace contract binding:

```bash
curl -s -X POST http://localhost:8000/trigger_fl | jq '.marketplace_contract'
curl -s http://localhost:8000/training/status | jq '.marketplace_pending_contract'
curl -s http://localhost:8000/metrics_summary | jq '.marketplace'
```

List contracts:

```bash
curl -s 'http://localhost:8000/marketplace/contracts' | jq
```

## Sprint 2 Trust and Governance Surfaces

Create a dispute for a contract:

```bash
curl -s -X POST http://localhost:8000/marketplace/disputes \
  -H 'Content-Type: application/json' \
  -d '{
    "contract_id": "contract-REPLACE_ME",
    "reporter": "hud-operator",
    "reason": "insufficient evidence of update quality"
  }' | jq
```

List disputes:

```bash
curl -s 'http://localhost:8000/marketplace/disputes?status=open&limit=20' | jq
```

Update dispute status:

```bash
curl -s -X PATCH http://localhost:8000/marketplace/disputes/dispute-REPLACE_ME \
  -H 'Content-Type: application/json' \
  -d '{"status": "under_review", "actor": "moderator"}' | jq
```

Create governance action log entry:

```bash
curl -s -X POST http://localhost:8000/governance/actions \
  -H 'Content-Type: application/json' \
  -d '{
    "action_type": "marketplace_policy_review_requested",
    "actor": "hud-operator",
    "source": "hud",
    "payload": {"contract_id": "contract-REPLACE_ME"}
  }' | jq
```

List governance actions:

```bash
curl -s 'http://localhost:8000/governance/actions?limit=20' | jq
```

## Sprint 3 Policy Preview and Proposal Voting

Run a non-persistent policy simulation before creating a contract:

```bash
curl -s -X POST http://localhost:8000/marketplace/policy/preview \
  -H 'Content-Type: application/json' \
  -d '{
    "round_intent_id": "intent-REPLACE_ME",
    "max_offers": 3,
    "policy_overrides": {
      "budget_total": 45,
      "min_quality_score": 0.8,
      "required_modalities": ["image"]
    }
  }' | jq
```

Create a governance proposal:

```bash
curl -s -X POST http://localhost:8000/governance/proposals \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Raise Marketplace Quality Floor",
    "description": "Increase min quality for image rounds to reduce noisy updates",
    "proposal_type": "policy_update",
    "created_by": "hud-operator",
    "close_threshold_yes_ratio": 0.67
  }' | jq
```

List proposals:

```bash
curl -s 'http://localhost:8000/governance/proposals?status=open&limit=20' | jq
```

Cast a weighted vote:

```bash
curl -s -X POST http://localhost:8000/governance/proposals/proposal-REPLACE_ME/vote \
  -H 'Content-Type: application/json' \
  -d '{
    "voter": "node-42",
    "decision": "yes",
    "weight": 2.5,
    "reason": "improves trust and quality"
  }' | jq
```

Update proposal status manually (moderation/workflow):

```bash
curl -s -X PATCH http://localhost:8000/governance/proposals/proposal-REPLACE_ME \
  -H 'Content-Type: application/json' \
  -d '{"status": "closed", "actor": "moderator"}' | jq
```

## Share Compute Attestations and Expand Network

Share available compute with attestation status (public/local endpoint):

```bash
curl -s -X POST http://localhost:8000/attestations/share \
  -H 'Content-Type: application/json' \
  -d '{
    "participant_name": "community-node-12",
    "node_name": "community-node-12",
    "compute_type": "gpu",
    "compute_capacity": "1x A100 40GB",
    "capacity_score": 8.8,
    "attestation_status": "verified",
    "region": "eu-west",
    "proof_digest": "sha256:attestation-proof-001",
    "notes": "available nightly for FL rounds"
  }' | jq
```

List shared compute attestations:

```bash
curl -s 'http://localhost:8000/attestations/feed?status=verified&sort_by=reputation&limit=20' | jq
```

Request a network join invite (self-service):

```bash
curl -s -X POST http://localhost:8000/join/request_invite \
  -H 'Content-Type: application/json' \
  -d '{
    "participant_name": "community-node-44",
    "contact_email": "community-node-44@example.org",
    "compute_type": "gpu",
    "region": "ap-south",
    "preferred_language": "en",
    "motivation": "Contribute verified weekend compute"
  }' | jq
```

Admin list/approve invite requests:

```bash
curl -s 'http://localhost:8000/join/invite_requests?status=pending&limit=20' \
  -H 'X-Join-Admin-Token: local-dev-admin-token' | jq

curl -s 'http://localhost:8000/join/invite_requests?status=pending&q=gpu&limit=10&offset=0' \
  -H 'X-Join-Admin-Token: local-dev-admin-token' | jq

curl -s 'http://localhost:8000/join/invite_requests?status=all&q=node&limit=10&offset=0&sort_by=participant_name&sort_dir=asc' \
  -H 'X-Join-Admin-Token: local-dev-admin-token' | jq

curl -s -X POST http://localhost:8000/join/invite_requests/join-req-REPLACE_ME/approve \
  -H 'Content-Type: application/json' \
  -H 'X-Join-Admin-Token: local-dev-admin-token' \
  -d '{"max_uses": 1, "expires_in_hours": 24}' | jq

curl -s -X POST http://localhost:8000/join/invite_requests/join-req-REPLACE_ME/reject \
  -H 'Content-Type: application/json' \
  -H 'X-Join-Admin-Token: local-dev-admin-token' \
  -d '{"reason": "insufficient profile detail"}' | jq
```

List invites and revoke one:

```bash
curl -s 'http://localhost:8000/join/invites?include_revoked=true&limit=20' \
  -H 'X-Join-Admin-Token: local-dev-admin-token' | jq

curl -s 'http://localhost:8000/join/invites?status=active&q=community&limit=10&offset=0' \
  -H 'X-Join-Admin-Token: local-dev-admin-token' | jq

curl -s 'http://localhost:8000/join/invites?status=all&include_revoked=true&limit=10&offset=0&sort_by=used&sort_dir=desc' \
  -H 'X-Join-Admin-Token: local-dev-admin-token' | jq

curl -s 'http://localhost:8000/join/registrations?status=all&q=node&limit=10&offset=0&sort_by=registered_at&sort_dir=desc' \
  -H 'X-Join-Admin-Token: local-dev-admin-token' | jq

curl -s 'http://localhost:8000/join/registrations?status=all&limit=10&offset=0&sort_by=node_id&sort_dir=asc' \
  -H 'X-Join-Admin-Token: local-dev-admin-token' | jq

curl -s -X POST http://localhost:8000/join/invites/INVITE-ID-REPLACE_ME/revoke \
  -H 'X-Join-Admin-Token: local-dev-admin-token' | jq
```

Inspect admin auth methods supported by backend:

```bash
curl -s http://localhost:8000/admin/auth/methods | jq
```

Optional wallet-based admin auth (when `ADMIN_WALLET_ALLOWLIST` is configured):

```bash
curl -s 'http://localhost:8000/join/invite_requests?status=pending&limit=20' \
  -H 'X-Admin-Wallet: 0xYOUR_ALLOWLISTED_WALLET' | jq
```

Get network expansion summary:

```bash
curl -s http://localhost:8000/network/expansion_summary | jq
curl -s http://localhost:8000/metrics_summary | jq '.network_expansion'
```

## Tokenomics Exporter

If the tokenomics exporter is running separately:

```bash
curl -s http://localhost:9105/health | jq
curl -s http://localhost:9105/metrics | head -n 40
```
