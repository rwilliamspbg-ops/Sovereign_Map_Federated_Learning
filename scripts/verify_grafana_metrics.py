#!/usr/bin/env python3
"""Verify Prometheus/Grafana metric availability for the Sovereign stack.

Checks:
- Prometheus API reachability
- Required scrape jobs are present in activeTargets
- Key metrics are queryable and return a response payload
"""

from __future__ import annotations

import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request
from typing import Dict, List, Tuple

PROM_URL = os.getenv("PROMETHEUS_URL", "http://localhost:9090")
TIMEOUT_SECONDS = float(os.getenv("METRICS_VERIFY_TIMEOUT_SECONDS", "4"))

REQUIRED_JOBS = [
    "sovereign-backend",
    "tpm-metrics",
    "tokenomics-metrics",
    "fl-performance-metrics",
]

# Metric presence checks use expressions that return a scalar/vector even with sparse data.
REQUIRED_QUERIES: Dict[str, str] = {
    "backend_round_gauge": "max(sovereignmap_fl_round) or vector(0)",
    "backend_accuracy": "max(sovereignmap_fl_accuracy) or vector(0)",
    "tpm_trust_score": "max(tpm_node_trust_score) or vector(0)",
    "tokenomics_chain_height": "max(tokenomics_chain_height) or vector(0)",
    "fl_gpu_utilization": "max(fl_gpu_utilization_percent) or vector(0)",
}


def http_get_json(url: str) -> Dict:
    req = urllib.request.Request(url, method="GET")
    with urllib.request.urlopen(req, timeout=TIMEOUT_SECONDS) as resp:
        return json.loads(resp.read().decode("utf-8"))


def prom_query(expr: str) -> Tuple[bool, str]:
    encoded = urllib.parse.quote(expr, safe="")
    url = f"{PROM_URL}/api/v1/query?query={encoded}"
    try:
        payload = http_get_json(url)
    except Exception as exc:
        return False, f"query error: {exc}"

    if payload.get("status") != "success":
        return False, f"query failed status={payload.get('status')}"

    data = payload.get("data", {})
    if "result" not in data:
        return False, "missing result in Prometheus response"

    return True, "ok"


def main() -> int:
    failures: List[str] = []

    # 1) Prometheus availability
    try:
        flags = http_get_json(f"{PROM_URL}/api/v1/status/flags")
        if flags.get("status") != "success":
            failures.append("Prometheus status endpoint returned non-success")
    except urllib.error.URLError as exc:
        print(f"Prometheus unreachable at {PROM_URL}: {exc}")
        return 1
    except Exception as exc:
        print(f"Prometheus status check failed: {exc}")
        return 1

    # 2) Required jobs in active targets
    try:
        targets = http_get_json(f"{PROM_URL}/api/v1/targets")
        active = targets.get("data", {}).get("activeTargets", [])
        jobs = {t.get("labels", {}).get("job") for t in active}
    except Exception as exc:
        print(f"Could not inspect targets: {exc}")
        return 1

    for job in REQUIRED_JOBS:
        if job not in jobs:
            failures.append(f"missing scrape job in activeTargets: {job}")

    # 3) Query verification for key metrics
    for check_name, expr in REQUIRED_QUERIES.items():
        ok, msg = prom_query(expr)
        if not ok:
            failures.append(f"{check_name}: {msg}")

    if failures:
        print("Grafana/metrics verification failed:")
        for item in failures:
            print(f"- {item}")
        return 1

    print("Grafana/metrics verification passed.")
    print(f"Prometheus: {PROM_URL}")
    print("Required jobs and metric queries are available.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
