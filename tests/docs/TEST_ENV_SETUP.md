# Test Environment Setup

Use this setup before running the consolidated full validation suite.

## One-command setup

```bash
bash scripts/setup-test-environment.sh
```

## Why this exists

Some Linux environments enforce PEP 668 (externally-managed Python), which can block standard `pip install` writes. The setup script retries with `--break-system-packages` so local test execution remains deterministic.

The full validation runner now executes all Python checks with the same interpreter used to launch the suite (`sys.executable`). This keeps dependency resolution consistent when using a virtual environment.

## What it installs

- Frontend test dependencies via `npm --prefix frontend ci`
- Python dependencies from `requirements.txt`

## Run validation profiles

```bash
npm run test:full:fast
npm run test:full:deep

# direct runner invocation
python tests/scripts/python/run_full_validation_suite.py --profile fast
python tests/scripts/python/run_full_validation_suite.py --profile deep

# preferred explicit venv invocation
./.venv/bin/python tests/scripts/python/run_full_validation_suite.py --profile deep
```

## Run trend and diff checks

```bash
npm run test:trends
npm run test:summary:diff
```

## Backward compatible entrypoint

```bash
npm run test:full
```

`npm run test:full` currently maps to the fast profile.

## Artifacts

Each run emits:

- `test-results/full-validation/full_validation_<timestamp>.json`
- `test-results/full-validation/full_validation_<timestamp>.md`
- `test-results/full-validation/history.jsonl` (trend timeline)

Latest deep-profile run reference (2026-04-04):

- `test-results/full-validation/full_validation_20260404_120632.json`
- `test-results/full-validation/full_validation_20260404_120632.md`
- `test-results/full-validation/history.jsonl`
