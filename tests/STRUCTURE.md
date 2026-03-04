# Test Asset Structure

This repository keeps executable test assets and test outputs organized under `tests/`:

- `tests/scripts/bash/` — Bash test orchestration scripts
- `tests/scripts/powershell/` — PowerShell test scripts
- `tests/scripts/python/` — Python-based test runners
- `tests/config/` — Test-specific configuration files
- `tests/docs/` — Test process documentation
- `tests/results/reports/` — Test result reports

Compatibility wrappers remain at legacy root paths to avoid breaking existing commands and automation.

See `tests/docs/TEST_FILE_CATALOG.md` for the full categorized test file inventory, including Go tests, top-level wrappers, and archived legacy assets.
