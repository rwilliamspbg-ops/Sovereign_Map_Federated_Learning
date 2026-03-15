# Test Scripts

Root-level test and benchmark scripts were consolidated into this folder tree:

- tests/scripts/bash
- tests/scripts/powershell
- tests/scripts/python
- tests/scripts/config

## Why this exists

- Keeps repository root focused on runtime/build entrypoints.
- Groups test orchestration by shell/runtime.
- Reduces path sprawl and improves discoverability for CI and contributors.

## Common entrypoints

- NPU 1000-node test (bash): `tests/scripts/bash/run-1000-node-npu-test.sh`
- NPU 1000-node test (PowerShell): `tests/scripts/powershell/run-1000-node-npu-test.ps1`
- General bash test runner: `tests/scripts/bash/run-test.sh`
- General PowerShell test runner: `tests/scripts/powershell/run-test.ps1`
- Python orchestration runner: `tests/scripts/python/run-test-python.py`

## Python suites moved here

- `tests/scripts/python/gpu-test-suite.py`
- `tests/scripts/python/byzantine-stress-test-suite.py`
- `tests/scripts/python/generate-byzantine-plots.py`
- `tests/scripts/python/generate-byzantine-test-suite-plots.py`
- `tests/scripts/python/kubernetes-5000-node-test.py`
- `tests/scripts/python/k8s-5000-node-local-test.py`
- `tests/scripts/python/npu-gpu-cpu-benchmark.py`

## Config and artifacts

- Test config: `tests/scripts/config/test-config.env`
- Benchmark JSON artifacts: `test-results/benchmarks/`
