# Test File Catalog

This document categorizes test-related files discovered in the repository and points to canonical locations.

## 1. Go Unit and Integration Tests

- `internal/batch/aggregator_test.go`
- `internal/consensus/consensus_test.go`
- `internal/consensus/consensus_200_test.go`
- `internal/convergence/detector_test.go`
- `internal/island/island_test.go`
- `internal/p2p/p2p_test.go`
- `internal/tpm/tpm_test.go`

## 2. Canonical Test Suite (`tests/`)

### 2.1 HIL Tests

- `tests/hil/test_npu_device_selection.py`
- `tests/hil/test_tpm_emulated.py`

### 2.2 Scale and Boundary Scenario Tests

- `tests/scale-tests/bft_100m_theoretical.py`
- `tests/scale-tests/bft_20node_200round_boundary.py`

### 2.3 Bash Test Orchestration Scripts

- `tests/scripts/bash/commit-test-results.sh`
- `tests/scripts/bash/run-test.sh`
- `tests/scripts/bash/setup-200node-test.sh`
- `tests/scripts/bash/test-dashboard.sh`
- `tests/scripts/bash/test-incremental-scale.sh`

### 2.4 PowerShell Test Orchestration Scripts

- `tests/scripts/powershell/continuous-load-test.ps1`
- `tests/scripts/powershell/run-5000-round-test.ps1`
- `tests/scripts/powershell/run-full-test.ps1`
- `tests/scripts/powershell/run-test.ps1`
- `tests/scripts/powershell/validate-grafana.ps1`

### 2.5 Python Test Helpers and Contract Tests

- `tests/scripts/python/run-test-python.py`
- `tests/scripts/python/test_communication_contracts.py`

### 2.6 Test Config and Reports

- `tests/config/test-config.env`
- `tests/results/reports/FINAL_TEST_REPORT.md`
- `tests/results/reports/VALIDATION_REPORT.md`
- `tests/FINAL_TEST_SUMMARY_20260227.md`

## 3. Root-Level Test Runners and Scenario Scripts

These are actively used top-level scripts and scenario drivers.

- `tests/scripts/python/byzantine-stress-test-suite.py`
- `tests/scripts/python/gpu-test-suite.py`
- `tests/scripts/python/k8s-5000-node-local-test.py`
- `tests/scripts/python/kubernetes-5000-node-test.py`
- `tests/scripts/bash/run-1000-node-npu-test.sh`
- `tests/scripts/powershell/run-1000-node-npu-test.ps1`
- `tests/scripts/bash/run-test.sh`
- `tests/scripts/powershell/run-test.ps1`
- `tests/scripts/python/run-test-python.py`
- `tests/scripts/powershell/run-full-test.ps1`
- `tests/scripts/powershell/run-5000-round-test.ps1`
- `tests/scripts/bash/test-dashboard.sh`
- `tests/scripts/bash/test-incremental-scale.sh`
- `tests/scripts/bash/setup-200node-test.sh`
- `tests/scripts/bash/commit-test-results.sh`
- `tests/scripts/powershell/continuous-load-test.ps1`
- `phase-4-execute-test.sh`

## 4. Test Analysis and Plot Generation

- `tests/scripts/python/generate-byzantine-test-suite-plots.py`
- `scripts/generate-npu-test-plots.py`

## 5. Archived Legacy Test Assets

These files are historical and kept under `archive/`.

- `archive/legacy/code/bft_stress_test.py`
- `archive/legacy/code/bft_stress_test_500k.py`
- `archive/legacy/code/run_bft_test.py`
- `archive/legacy/code/run_bft_tpm_test.py`
- `archive/legacy/code/run_sovereign_test.py`
- `archive/legacy/code/sovereign_map_test_collector.py`
- `archive/legacy/code/tpm_attestation.py`
- `archive/week2/code/run_week2_tests.sh`

## Notes

- Canonical test assets are organized in `tests/`.
- Root-level scripts are retained for compatibility and operational workflows.
- Legacy files under `archive/` are historical references and should not be used as default entry points for new validation runs.
- Test sources and test documentation are intentionally public and must remain visible in the repository.
- Do not commit runtime secrets into test files; use placeholder values in tracked configs and inject secrets from environment variables at execution time.
