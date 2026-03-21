# Feature Artifacts 2026-03-21

This manifest captures deliverables completed in the PySyft + dashboard consolidation feature branch.

## PySyft x Mohawk PoC Artifacts

- [examples/pysyft-integration/README.md](../examples/pysyft-integration/README.md)
- [examples/pysyft-integration/pysyft_mohawk_poc.py](../examples/pysyft-integration/pysyft_mohawk_poc.py)
- [examples/pysyft-integration/pysyft_mohawk_poc.ipynb](../examples/pysyft-integration/pysyft_mohawk_poc.ipynb)
- [examples/pysyft-integration/docker-compose.pysyft-demo.yml](../examples/pysyft-integration/docker-compose.pysyft-demo.yml)
- [examples/pysyft-integration/requirements-pysyft-demo.txt](../examples/pysyft-integration/requirements-pysyft-demo.txt)

## Canonical STARRED Live Dashboards

- [grafana/provisioning/dashboards/operations_overview.json](../grafana/provisioning/dashboards/operations_overview.json)
- [grafana/provisioning/dashboards/tokenomics_overview.json](../grafana/provisioning/dashboards/tokenomics_overview.json)
- [grafana/provisioning/dashboards/llm_overview.json](../grafana/provisioning/dashboards/llm_overview.json)
- [grafana/provisioning/dashboards/dashboards.yml](../grafana/provisioning/dashboards/dashboards.yml)

## Compose Profiles Updated

- [docker-compose.1000nodes.yml](../docker-compose.1000nodes.yml)
- [docker-compose.200nodes.yml](../docker-compose.200nodes.yml)
- [docker-compose.dev.yml](../docker-compose.dev.yml)
- [docker-compose.full.yml](../docker-compose.full.yml)
- [docker-compose.large-scale.yml](../docker-compose.large-scale.yml)
- [docker-compose.monitoring.tpm.yml](../docker-compose.monitoring.tpm.yml)
- [docker-compose.monitoring.yml](../docker-compose.monitoring.yml)
- [docker-compose.production.yml](../docker-compose.production.yml)
- [monitoring/docker-compose.monitoring.yml](../monitoring/docker-compose.monitoring.yml)

## Validation Snapshot

- Dashboard JSON parse: passed for all canonical dashboards.
- Compose parse: passed for `docker-compose.monitoring.yml`, `docker-compose.dev.yml`, and `docker-compose.full.yml`.
- PySyft mock PoC run: passed (`--mode mock --rounds 2 --participants 3`).

## Removed as Unused

- `grafana/provisioning/dashboards/dashboard-provider.yaml`
- `grafana/dashboards/*.json`
- `monitoring/dashboards/*.json`
- `monitoring/grafana/provisioning/dashboards/fl-dashboard.json`
