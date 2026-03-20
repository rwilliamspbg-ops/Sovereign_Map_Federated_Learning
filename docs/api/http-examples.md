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

## Tokenomics Exporter

If the tokenomics exporter is running separately:

```bash
curl -s http://localhost:9105/health | jq
curl -s http://localhost:9105/metrics | head -n 40
```
