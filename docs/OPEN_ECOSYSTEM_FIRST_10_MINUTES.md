# Open Ecosystem First 10 Minutes (Local)

This guide runs entirely local.

## Prerequisites

1. Backend API available at `http://localhost:8000`.
2. Frontend available at `http://localhost:3000` (or local Vite port).

## 1. Create Offer

```bash
curl -s -X POST http://localhost:8000/marketplace/offers \
  -H 'Content-Type: application/json' \
  -d '{
    "seller_node_id": "node-quickstart-1",
    "dataset_fingerprint": "sha256:quickstart-local-001",
    "title": "Quickstart Image Pack",
    "modality": "image",
    "quality_score": 0.84,
    "allowed_tasks": ["classification"],
    "price_per_round": 10.0,
    "min_rounds": 1,
    "attestation_status": "verified"
  }' | jq
```

## 2. Create Intent

```bash
curl -s -X POST http://localhost:8000/marketplace/round_intents \
  -H 'Content-Type: application/json' \
  -d '{
    "model_owner_id": "owner-quickstart",
    "task_type": "classification",
    "required_modalities": ["image"],
    "min_quality_score": 0.7,
    "budget_total": 100
  }' | jq
```

Capture `round_intent_id` from the response.

## 3. Match Contract

```bash
curl -s -X POST http://localhost:8000/marketplace/match \
  -H 'Content-Type: application/json' \
  -d '{"round_intent_id": "intent-REPLACE_ME", "max_offers": 3}' | jq
```

Capture `contract_id` from the response.

## 4. Trigger One Training Round

```bash
curl -s -X POST http://localhost:8000/trigger_fl | jq
```

## 5. Release Escrow

```bash
curl -s -X POST http://localhost:8000/marketplace/escrow/release \
  -H 'Content-Type: application/json' \
  -d '{"contract_id": "contract-REPLACE_ME"}' | jq
```

## 6. Inspect Contract Timeline and Metrics

```bash
curl -s http://localhost:8000/marketplace/contracts | jq '.contracts[0].timeline'
curl -s http://localhost:8000/training/status | jq '.marketplace_pending_contract'
curl -s http://localhost:8000/metrics_summary | jq '.marketplace'
```

## Troubleshooting

1. `no_compatible_offers_found`:

- Check `details.rejection_reasons` in response.
- Increase `budget_total` or reduce quality threshold.

1. `round_intent_not_open`:

- Intent was already matched/cancelled/closed.
- Create a new intent or patch status appropriately.

1. `contract_already_released`:

- Escrow for that contract is already released.
