# TPM + NPU Full Validation Report

- Generated: 2026-03-03T19:42:43.684187+00:00
- TPM Go tests: PASS
- TPM HIL emulation tests: SKIPPED (missing swtpm/tpm2-tools binaries in container)
- NPU benchmark pipeline: PASS (executed on available device set)
- NPU availability in this run: NO
- CUDA availability in this run: NO

## BFT Breaking Point (70%-99%)
- Ratios tested: 70%, 75%, 80%, 85%, 90%, 95%, 99%
- Breaking point: Not found in range
- All test points passed: True
- Accuracy by ratio:
  - 70%: 85.9706%
  - 75%: 85.9691%
  - 80%: 85.9677%
  - 85%: 85.9664%
  - 90%: 85.9651%
  - 95%: 85.9635%
  - 99%: 85.9618%

## Device Benchmark Snapshot
- Source: test-results/tpm-npu-full/npu-benchmark-20260303-194207.json
- Devices benchmarked: cpu
- CPU epoch time: 1.4614200592041016
- CPU throughput (samples/sec): 547.4127681234135

## Captured Plot Artifacts
- combined-summary.png
- scenario-1-1000node.png
- scenario-2-threshold.png
- scenario-3-intensity.png
