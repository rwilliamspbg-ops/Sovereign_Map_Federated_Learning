# Release Notes — TPM/NPU Validation & Throughput Update (2026-03-03)

## Scope
This release consolidates TPM hardening, GPU/NPU/CPU auto-detection improvements, refreshed heatmapping/throughput artifacts, and documentation synchronization.

## Key Changes
- Hardened TPM quote validation with stricter parsing, freshness checks, node binding, and replay-aware nonce handling.
- Improved attestation error handling and validation guardrails.
- Migrated Flower client startup from deprecated `start_numpy_client` to `start_client`.
- Improved runtime accelerator selection with robust NPU/CUDA probing and CPU fallback safety.
- Regenerated heatmapping and throughput artifacts and refreshed packaged bundle.

## Test/Artifact Snapshot
- Byzantine sweep ratios tested: `70, 75, 80, 85, 90, 95, 99`
- Breaking point in tested range: `Not found in range`
- Contention throughput (10 nodes, CPU): `746.97 samples/sec`
- Round-latency throughput (10 nodes, 3 rounds, CPU): `3.331 updates/sec`
- Average round latency: `3.002 sec`

## Primary Artifacts
- `test-results/tpm-npu-full/TPM_NPU_VALIDATION_REPORT.md`
- `test-results/tpm-npu-full/artifact-manifest.json`
- `test-results/tpm-npu-full/throughput-contention-20260303-195357.json`
- `test-results/tpm-npu-full/throughput-round-latency-20260303-195357.json`
- `test-results/tpm-npu-full/heatmapping-suite-20260303-195357.log`
- `test-results/tpm-npu-full/heatmapping-plots-20260303-195357.log`
- `test-results/tpm-npu-full-artifacts.tar.gz`

## Commit Sequence
- `a32752f` — feat: harden TPM validation and migrate Flower client startup
- `a7fb3d3` — test: add heatmapping and throughput artifact refresh
- `1769301` — docs: add latest heatmapping and throughput test snapshot
- `7b490c5` — docs: add heatmapping and throughput snapshot to index
- `c7107d6` — docs: append TPM NPU testing changelog addendum
