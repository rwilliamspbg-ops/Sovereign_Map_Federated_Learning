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

Live runtime scrape proof:

Command pattern used:

```bash
FL_REPLAY_LIVE_PORT=9915 FL_REPLAY_SLEEP_SECONDS=1.0 \
FL_REPLAY_PATTERN=24,120,24,120,24,120,24,120,24,120,24,120 \
./.venv/bin/python scripts/replay_aggregation_path_metrics.py
```

During execution, `/metrics` was scraped twice and counters moved in-flight:

```text
snapshot_a= {'fl_aggregation_path_total{impl="loop"}': 4.0, 'fl_aggregation_path_total{impl="vectorized"}': 3.0}
snapshot_b= {'fl_aggregation_path_total{impl="loop"}': 4.0, 'fl_aggregation_path_total{impl="vectorized"}': 4.0}
delta= {'fl_aggregation_path_total{impl="loop"}': 0.0, 'fl_aggregation_path_total{impl="vectorized"}': 1.0}
```

This confirms live counter movement that aligns with panel increase semantics.

## 2) Alert Rule Test Status

Command:

```bash
docker run --rm --entrypoint /bin/promtool \
  -v "$PWD":/work -w /work prom/prometheus:v2.48.0 \
  test rules tpm_alerts.test.yml
```

Result:

```text
Unit Testing:  tpm_alerts.test.yml
  SUCCESS
```

Notes:

- Local `promtool` binary is not installed, but tests are executable using the pinned Prometheus image.

## 3) Docs Lint Guard Evidence

Command:

```bash
cat > /tmp/markdownlint-docs.json <<'EOF'
{
  "default": true,
  "MD013": false
}
EOF
npx --yes markdownlint-cli2 --config /tmp/markdownlint-docs.json README.md docs/ALERT_RUNBOOKS.md
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
