# Capabilities API Schema v1

## Endpoint
- `GET /api/v1/capabilities`

## Purpose
This document defines the versioned response contract for runtime capability discovery. It is intended for clients, tooling, and CI contract validation.

## Contract Version
- Schema version: `v1`
- Protocol version field: `protocol_version` (currently `1.0.0`)

## Top-level fields
- `protocol_version`: string
- `stark_backends`: string[]
- `capabilities`: object
- `bridge_policies`: object
- `capabilities_path`: string
- `bridge_path`: string
- `api`: object
- `observability`: object

## `capabilities` object (selected fields)
- `version`: string
- `runtime`: string
- `byzantine_threshold`: number
- `nodes`: object[]
- `cluster_config`: object
- `performance_targets`: object
- `thinker_clauses`: object

### `capabilities.thinker_clauses`
- `enabled`: boolean
- `preserve_outliers`: boolean
- `minority_retention_min`: number
- `minority_retention_max`: number
- `outlier_distance_zscore_cap`: number
- `manual_review_required_above_zscore`: number
- `escalation_label`: string

## `api` object
- `base_paths`: string[]
- `open_endpoints`: string[]
- `auth_protected_endpoints`: string[]
- `proof_payload`: object
- `hybrid_payload`: object
- `auth`: object

### `api.proof_payload`
- `fields`: string[]
- `supported_encodings`: string[]

### `api.hybrid_payload`
- `fields`: string[]
- `supported_modes`: string[]
- `supported_encodings`: string[]

### `api.auth`
- `default_mode`: string
- `disable_values`: string[]
- `token_headers`: string[]
- `role_header`: string
- `default_token_file`: string
- `roles_enforced_by_default`: boolean
- `default_allowed_roles`: string[]

## `observability` object
- `proof_metrics`: string[]
- `ledger_metrics`: string[]
- `ledger_state`: object

### `observability.ledger_state`
- `entries`: integer
- `capacity`: integer

## Example response (truncated)
```json
{
  "protocol_version": "1.0.0",
  "stark_backends": ["simulated_fri", "winterfell_mock"],
  "api": {
    "base_paths": ["/api", "/api/v1"],
    "open_endpoints": [
      "GET /health",
      "GET /api/v1/status",
      "GET /api/v1/capabilities"
    ],
    "auth_protected_endpoints": [
      "POST /api/v1/proof/verify",
      "POST /api/v1/proof/hybrid/verify",
      "GET /api/v1/ledger"
    ],
    "proof_payload": {
      "fields": ["proof", "encoding", "public_input"],
      "supported_encodings": ["base64", "hex", "raw"]
    },
    "hybrid_payload": {
      "fields": ["mode", "stark_backend", "snark_proof", "stark_proof", "encoding"],
      "supported_modes": ["any", "both", "prefer_snark"],
      "supported_encodings": ["base64", "hex", "raw"]
    }
  },
  "observability": {
    "proof_metrics": [
      "mohawk_proof_verifications_total",
      "mohawk_proof_verification_latency_seconds"
    ],
    "ledger_metrics": [
      "mohawk_ledger_events_total",
      "mohawk_ledger_entries"
    ],
    "ledger_state": {
      "entries": 0,
      "capacity": 1000
    }
  },
  "capabilities": {
    "version": "1.0.0",
    "runtime": "mohawk-proto-v1",
    "byzantine_threshold": 0.55,
    "thinker_clauses": {
      "enabled": true,
      "preserve_outliers": true,
      "minority_retention_min": 0.05,
      "minority_retention_max": 0.2,
      "outlier_distance_zscore_cap": 4.0,
      "manual_review_required_above_zscore": 3.5,
      "escalation_label": "thinker-review"
    }
  }
}
```

## CI contract enforcement
- Test: `TestCapabilitiesContractV1`
- File: `internal/api/handlers_test.go`
- CI workflow: `.github/workflows/build.yml`
