## Security fixes applied (automated patch)

Summary:

- Replaced an unsafe `sed` via `subprocess(..., shell=True)` with a safe Python file edit in `docs/archive/legacy/code/run_sovereign_test.py`.
- Replaced `subprocess(..., shell=True)` usage in `tests/scripts/python/run_full_validation_suite.py` with `shlex.split()` + list-args execution to avoid shell injection risks.
- Added a Bandit workflow in `.github/workflows/security-python.yml` and a matching pre-commit hook in `.pre-commit-config.yaml` for the highest-risk Python paths.
- Added `bandit` and `pre-commit` to `requirements-ci.txt` so the security checks can run consistently in CI.

Files modified:

- `docs/archive/legacy/code/run_sovereign_test.py` — removed `shell=True` sed invocation and perform safe in-file replacement using Python `re`.
- `tests/scripts/python/run_full_validation_suite.py` — added `import shlex` and switched to `shlex.split()` + `shell=False` for running commands.
- `.github/workflows/security-python.yml` — new CI job that runs Bandit on the core backend, SDK, training, and validation scripts.
- `.pre-commit-config.yaml` — new local hook for targeted Bandit scanning before commits.
- `requirements-ci.txt` — pinned the Python security tooling used by CI and developers.

Checklist (next steps you should perform):

- [ ] Review the modified files and confirm the replacement behavior is acceptable for your workflows.
- [ ] Run the full validation suite locally: `python3 tests/scripts/python/run_full_validation_suite.py` and inspect logs in `test-results/full-validation/`.
- [ ] Audit other flagged files for `shell=True` and unsafe subprocess usage (search for `shell=True`, `subprocess.run(`, `check_output(`).
- [ ] Search for `verify=False` and update HTTPS calls to enable TLS verification.
- [ ] Run the existing `Secret Scan` workflow and remove any committed secrets from history if found.
- [ ] Keep the new Bandit workflow and pre-commit hook in sync with any new Python security-sensitive paths.
- [ ] Rotate any secrets or tokens if you discover they were exposed.

How I tested the changes:

- Static updates only; the changes avoid shell usage. Please run the suite and CI to verify behavior in your environment.

If you want, I can open a PR with these changes and add a `detect-secrets` pre-commit hook and a minimal GitHub Actions workflow to run `bandit` and secret scanning.
