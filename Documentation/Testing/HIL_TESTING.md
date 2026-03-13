# Hardware-in-the-Loop (HIL) Testing

This repository includes explicit HIL-oriented tests for:

- TPM 2.0 command flows using `swtpm` (emulated TPM)
- NPU device-selection and CPU/CUDA fallback logic

## What Was Added

- CI workflow: `.github/workflows/hil-tests.yml`
- TPM emulator tests: `tests/hil/test_tpm_emulated.py`
- NPU fallback tests: `tests/hil/test_npu_device_selection.py`

## Local Run: TPM Emulator Tests

Install required tools on Ubuntu:

```bash
sudo apt-get update
sudo apt-get install -y swtpm swtpm-tools tpm2-tools
python -m pip install --upgrade pip
pip install pytest
```

Run:

```bash
pytest -q tests/hil/test_tpm_emulated.py
```

The test suite starts an ephemeral `swtpm` instance and validates:

- TPM capability queries (`tpm2_getcap`)
- Primary key creation (`tpm2_createprimary`)
- PCR read (`tpm2_pcrread`)
- Random byte generation (`tpm2_getrandom`)

## Local Run: NPU Fallback Tests

```bash
python -m pip install --upgrade pip
pip install pytest
pytest -q tests/hil/test_npu_device_selection.py
```

These tests validate that client device selection in `src/client.py`:

- Honors `FORCE_CPU`
- Prefers NPU when available
- Falls back from NPU errors to CUDA
- Falls back to CPU when accelerators are unavailable

## CI Behavior

The `HIL Tests` GitHub Actions workflow runs on push/PR to `main`:

- `tpm-emulated`: installs `swtpm` + `tpm2-tools`, runs TPM emulator tests
- `npu-fallback`: runs NPU fallback logic tests

## Notes

- `swtpm` validates TPM logic and command integration, not physical TPM guarantees.
- For release-grade hardware attestation claims, add a self-hosted runner with a physical TPM and run a scheduled/nightly hardware job.