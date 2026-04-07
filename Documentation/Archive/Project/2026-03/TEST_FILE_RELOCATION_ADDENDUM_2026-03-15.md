# Test File Relocation Addendum (2026-03-15)

## Scope

This addendum records the repository cleanup that relocated test orchestration
files and benchmark artifacts from the repository root into structured test
folders.

## What Changed

- Test scripts were consolidated under `tests/scripts/`:
  - `tests/scripts/bash/`
  - `tests/scripts/powershell/`
  - `tests/scripts/python/`
  - `tests/scripts/config/`
- Benchmark JSON artifacts were consolidated under `test-results/benchmarks/`.

## Safety and Compatibility Updates

- Updated workflow references in `.github/workflows/test-artifacts-review.yml`.
- Updated top-level operator commands in `README.md`.
- Updated active documentation references for moved scripts and benchmark artifacts.
- Updated `tests/scripts/bash/run-1000-node-npu-test.sh` to resolve
  repository-root paths after relocation.

## Validation Evidence

- Bash syntax checks passed for scripts under `tests/scripts/bash/`.
- Python syntax checks passed for scripts under `tests/scripts/python/`.
- Focused API contract regression test passed:
  - `go test ./internal/api -run TestCapabilitiesContractV1 -count=1`
- Reference audit completed for active (non-archive) paths to ensure old
  root test script paths were removed.

## Result

- Repository root is cleaner and more maintainable.
- Test assets are organized by runtime and purpose.
- Documentation and workflow references are aligned with the new layout.
