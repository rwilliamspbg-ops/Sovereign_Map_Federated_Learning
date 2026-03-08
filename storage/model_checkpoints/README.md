# Model Checkpoints Artifact Store

This package provides the canonical location for model checkpoint artifacts.

Backends:
- `filesystem` (default local implementation)
- `ipfs` (planned adapter)
- `object-storage` (planned S3/GCS-compatible adapter)

Primary implementation:
- `store.go` exposes `Store.Put(version, payload)` and returns metadata with digest and URI.
