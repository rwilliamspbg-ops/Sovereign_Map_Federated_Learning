<!-- markdownlint-disable MD012 -->

# Screenshot Evidence Checklist

Use this checklist for every release or major observability/UI PR.

## Required Assets

- `hud-operations-overview.png`
- `grafana-operations-overview.png`
- `grafana-llm-overview.png`
- `grafana-tokenomics-overview.png`

## Capture Guidance

- Resolution: at least 1920x1080.
- Theme: default project theme (no browser dark-mode overrides).
- Data: capture during live traffic or active FL rounds, not idle empty dashboards.
- Timestamp: include system clock/time range in frame where possible.

## Minimum Content Per Screenshot

### HUD Operations Overview

- Platform state ribbon
- Core ports line including Prometheus
- Memory/CPU KPI cards
- Operations timeline with at least one training event

### Grafana Operations Overview

- Service Availability Score
- API p95 Latency
- Throughput and attestation panels

### Grafana Tokenomics Overview

- Mint/supply overview
- Bridge health section
- Validator/wallet economics section

## Optional GIFs (Recommended)

- `hud-first-round.gif`: click "Run Global FL Epoch" and show round metrics increment.
- `ops-alert-remediation.gif`: click degraded banner links and show remediation panes.

