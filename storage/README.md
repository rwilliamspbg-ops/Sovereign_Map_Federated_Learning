# Artifact Storage Layout

This repository uses `storage/` as the artifact store root.

## Paths

- `storage/model_checkpoints/`
  - Model versions, gradient artifacts, and checkpoint metadata.
- `storage/map_tiles/`
  - Map tile encoding, cache, and serving artifacts.
- `storage/ipfs/`
  - IPFS integration helpers and adapters.

## Backend Strategy

Preferred backends:
1. IPFS for content-addressed distribution.
2. Object storage (S3/GCS-compatible) for durable artifact retention.
3. Local filesystem for development and smoke tests.
