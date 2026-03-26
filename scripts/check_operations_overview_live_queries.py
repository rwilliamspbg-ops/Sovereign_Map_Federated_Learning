#!/usr/bin/env python3
"""Live Prometheus query smoke checks for lower-half operations dashboard panels."""

from __future__ import annotations

import json
import os
import sys
import urllib.parse
import urllib.request
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DASHBOARD_PATH = ROOT / "grafana" / "provisioning" / "dashboards" / "operations_overview.json"
PROM_URL = os.getenv("PROM_URL", "http://localhost:9090")
WINDOW = os.getenv("WINDOW", "15m")
RATE_INTERVAL = os.getenv("RATE_INTERVAL", "5m")


def prom_query(expr: str) -> int:
    query = expr.replace("$window", WINDOW).replace("$__rate_interval", RATE_INTERVAL)
    params = urllib.parse.urlencode({"query": query})
    url = f"{PROM_URL}/api/v1/query?{params}"
    with urllib.request.urlopen(url, timeout=15) as response:
        payload = json.loads(response.read().decode("utf-8"))
    if payload.get("status") != "success":
        return -1
    return len(payload.get("data", {}).get("result", []))


def main() -> int:
    dashboard = json.loads(DASHBOARD_PATH.read_text(encoding="utf-8"))
    failures: list[str] = []
    total_checks = 0

    for panel in dashboard.get("panels", []):
        if panel.get("type") == "row":
            continue
        grid = panel.get("gridPos") or {}
        if int(grid.get("y", 0)) < 48:
            continue

        title = panel.get("title", f"panel-{panel.get('id', 'unknown')}")
        for target in panel.get("targets", []):
            expr = target.get("expr", "").strip()
            if not expr:
                continue
            total_checks += 1
            try:
                series = prom_query(expr)
            except Exception as exc:  # noqa: BLE001
                failures.append(f"{title}: query error: {exc}")
                continue
            if series <= 0:
                failures.append(f"{title}: no series for expr={expr}")

    if failures:
        print("Operations overview live query smoke failed:")
        for item in failures:
            print(f"- {item}")
        return 1

    print(f"Operations overview live query smoke passed ({total_checks} checks).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
