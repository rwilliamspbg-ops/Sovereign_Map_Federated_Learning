# OPS Evidence Bundle - 2026-03-24

## Scope

This note captures evidence for the post-upgrade operational checks:

- FL aggregation path panel trend validation.
- Runbook/doc formatting hardening.
- Docs lint CI guard.
- Operator tuning defaults surfaced in quick-start docs.
- Alert-test execution status in current environment.

## 1) Aggregation Path Replay Evidence

Command:

```bash
./.venv/bin/python scripts/replay_aggregation_path_metrics.py
```

Observed output:

```text
replay_rounds= 12
mode= auto
vectorize_min_clients= 50
vectorize_max_peak_bytes= 2147483648
path_totals= {'loop': 5, 'vectorized': 7}
last5_rounds= {'loop': 2, 'vectorized': 3}
panel_trend_expectation= vectorized_up

# metrics_exposition
# HELP fl_aggregation_path_total Total aggregation path selections by implementation.
# TYPE fl_aggregation_path_total counter
fl_aggregation_path_total{impl="loop"} 5.0
fl_aggregation_path_total{impl="vectorized"} 7.0
```

Interpretation:

- Under mixed participation rounds, vectorized path usage exceeded loop usage.
- Expected panel behavior for `sum(increase(fl_aggregation_path_total[$window])) by (impl)` is a higher recent increase for `impl="vectorized"` in this replay pattern.

## 2) Alert Rule Test Status

Command:

```bash
promtool test rules tpm_alerts.test.yml
```

Result in this environment:

```text
promtool-not-available
```

Notes:

- Alert rule definitions and tests are present and cross-referenced in runbooks.
- Execute the command above in an environment that includes Prometheus tooling to collect full pass/fail output.

## 3) Docs Lint Guard Evidence

Command:

```bash
npx --yes markdownlint-cli2 --config .markdownlint-docs.json README.md docs/ALERT_RUNBOOKS.md
```

Observed output:

```text
Summary: 0 error(s)
```

## 4) Operator Defaults Cross-Reference

Defaults verified in `.env.example`:

- `FL_AGGREGATION_MODE=auto`
- `DP_NOISE_MULTIPLIER=1.1`
- `TPM_ATTESTATION_MAX_REPORTS=256`

Runbook and quick-start surfaces updated:

- `docs/ALERT_RUNBOOKS.md`
- `PHASE_3D_QUICK_START.md`

## 5) Dashboard Validation Reference

Panel title in operations dashboard:

- `FL Aggregation Path Usage`

Query semantics:

- `sum(increase(fl_aggregation_path_total[$window])) by (impl)`

The replay evidence above demonstrates the expected relative trend movement for this panel under a mixed client-load pattern.
