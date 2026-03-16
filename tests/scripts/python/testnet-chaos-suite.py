#!/usr/bin/env python3
"""Testnet chaos suite.

Introduces controlled client churn and verifies FL rounds continue advancing.
"""

from __future__ import annotations

import json
import random
import subprocess
import time
import urllib.parse
import urllib.request

COMPOSE_FILE = "docker-compose.full.yml"
PROM_URL = "http://localhost:9090"


def prom_query(expr: str) -> float:
    q = urllib.parse.quote(expr, safe="")
    with urllib.request.urlopen(
        f"{PROM_URL}/api/v1/query?query={q}", timeout=5
    ) as response:
        payload = json.loads(response.read().decode("utf-8"))
    result = payload.get("data", {}).get("result", [])
    if not result:
        return 0.0
    return float(result[0]["value"][1])


def running_node_containers() -> list[str]:
    out = subprocess.check_output(
        [
            "docker",
            "compose",
            "-f",
            COMPOSE_FILE,
            "ps",
            "--format",
            "json",
        ],
        text=True,
    )
    containers = []
    for line in out.splitlines():
        if not line.strip():
            continue
        row = json.loads(line)
        name = row.get("Name", "")
        if "node-agent" in name and row.get("State") == "running":
            containers.append(name)
    return containers


def restart_container(container_name: str):
    subprocess.check_call(["docker", "restart", container_name])


def main():
    print("[chaos] starting testnet chaos suite")
    before = prom_query("sovereignmap_fl_round")
    print(f"[chaos] round before={before}")

    for step in range(5):
        nodes = running_node_containers()
        if not nodes:
            raise RuntimeError("No running node-agent containers found")
        target = random.choice(nodes)
        print(f"[chaos] step {step+1}: restarting {target}")
        restart_container(target)
        time.sleep(8)

    after = prom_query("sovereignmap_fl_round")
    print(f"[chaos] round after={after}")

    if after <= before:
        raise SystemExit("[chaos] FAILED: FL rounds did not progress under churn")

    print("[chaos] PASSED: FL rounds progressed under controlled churn")


if __name__ == "__main__":
    main()
