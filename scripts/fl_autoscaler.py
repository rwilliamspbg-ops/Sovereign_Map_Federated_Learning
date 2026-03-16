#!/usr/bin/env python3
"""Simple FL autoscaler for docker-compose based testnet.

Scales `node-agent` replicas based on FL round throughput and participant count.
Run this on the host with Docker access.
"""

from __future__ import annotations

import json
import os
import subprocess
import time
import urllib.parse
import urllib.request

PROM_URL = os.getenv("PROMETHEUS_URL", "http://localhost:9090")
COMPOSE_FILE = os.getenv("COMPOSE_FILE", "docker-compose.full.yml")
SCALE_TARGET = os.getenv("SCALE_SERVICE", "node-agent")
MIN_REPLICAS = int(os.getenv("MIN_REPLICAS", "5"))
MAX_REPLICAS = int(os.getenv("MAX_REPLICAS", "30"))
LOOP_SECONDS = int(os.getenv("AUTOSCALE_INTERVAL_SECONDS", "30"))
TARGET_ROUNDS_PER_MIN = float(os.getenv("TARGET_ROUNDS_PER_MIN", "1.0"))


def prom_query(expr: str) -> float:
    q = urllib.parse.quote(expr, safe="")
    url = f"{PROM_URL}/api/v1/query?query={q}"
    with urllib.request.urlopen(url, timeout=5) as response:
        payload = json.loads(response.read().decode("utf-8"))
    result = payload.get("data", {}).get("result", [])
    if not result:
        return 0.0
    return float(result[0]["value"][1])


def current_replicas() -> int:
    cmd = [
        "docker",
        "compose",
        "-f",
        COMPOSE_FILE,
        "ps",
        "--services",
        "--filter",
        "status=running",
    ]
    out = subprocess.check_output(cmd, text=True)
    return sum(1 for line in out.splitlines() if line.strip() == SCALE_TARGET)


def scale_to(replicas: int):
    replicas = max(MIN_REPLICAS, min(MAX_REPLICAS, replicas))
    subprocess.check_call(
        [
            "docker",
            "compose",
            "-f",
            COMPOSE_FILE,
            "up",
            "-d",
            "--no-deps",
            "--scale",
            f"{SCALE_TARGET}={replicas}",
            SCALE_TARGET,
        ]
    )


def main():
    print("[autoscaler] starting...")
    while True:
        try:
            rounds_per_min = prom_query("rate(sovereignmap_fl_rounds_total[5m]) * 60")
            active_nodes = int(prom_query("sovereignmap_active_nodes"))
            current = max(active_nodes, MIN_REPLICAS)

            if rounds_per_min < TARGET_ROUNDS_PER_MIN * 0.8 and current < MAX_REPLICAS:
                desired = current + 1
            elif (
                rounds_per_min > TARGET_ROUNDS_PER_MIN * 1.4 and current > MIN_REPLICAS
            ):
                desired = current - 1
            else:
                desired = current

            if desired != current:
                print(
                    f"[autoscaler] rounds/min={rounds_per_min:.3f}, active={active_nodes}, scaling to {desired}"
                )
                scale_to(desired)
            else:
                print(
                    f"[autoscaler] rounds/min={rounds_per_min:.3f}, active={active_nodes}, no scale"
                )
        except Exception as exc:
            print(f"[autoscaler] warning: {exc}")

        time.sleep(LOOP_SECONDS)


if __name__ == "__main__":
    main()
