# Feature Artifacts 2026-03-21

This manifest captures deliverables completed in the PySyft + dashboard consolidation feature branch.

## PySyft x Mohawk PoC Artifacts

- [examples/pysyft-integration/README.md](../examples/pysyft-integration/README.md)
- [examples/pysyft-integration/pysyft_mohawk_poc.py](../examples/pysyft-integration/pysyft_mohawk_poc.py)
- [examples/pysyft-integration/pysyft_mohawk_poc.ipynb](../examples/pysyft-integration/pysyft_mohawk_poc.ipynb)
- [examples/pysyft-integration/requirements-pysyft-demo.txt](../examples/pysyft-integration/requirements-pysyft-demo.txt)

Note: legacy PySyft demo compose file was removed during stack consolidation; run the demo via script and requirements listed above.

## Canonical STARRED Live Dashboards

- [grafana/provisioning/dashboards/operations_overview.json](../grafana/provisioning/dashboards/operations_overview.json)
- [grafana/provisioning/dashboards/tokenomics_overview.json](../grafana/provisioning/dashboards/tokenomics_overview.json)
- [grafana/provisioning/dashboards/llm_overview.json](../grafana/provisioning/dashboards/llm_overview.json)
- [grafana/provisioning/dashboards/dashboards.yml](../grafana/provisioning/dashboards/dashboards.yml)

## Compose Profiles Updated

- [docker-compose.full.yml](../docker-compose.full.yml)

Legacy compose profiles referenced in older notes were consolidated; use `docker-compose.full.yml` as the canonical stack entrypoint.

## Validation Snapshot

- Dashboard JSON parse: passed for all canonical dashboards.
- Compose parse: passed for `docker-compose.monitoring.yml`, `docker-compose.dev.yml`, and `docker-compose.full.yml`.
- PySyft mock PoC run: passed (`--mode mock --rounds 2 --participants 3`).

## Removed as Unused

- `grafana/provisioning/dashboards/dashboard-provider.yaml`
- `grafana/dashboards/*.json`
- `monitoring/dashboards/*.json`
- `monitoring/grafana/provisioning/dashboards/fl-dashboard.json`
