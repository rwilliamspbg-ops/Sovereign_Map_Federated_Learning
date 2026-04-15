#!/usr/bin/env python3
"""Run chaos soak checks when explicitly enabled in CI."""

from __future__ import annotations

import os
import json
import shutil
import subprocess
import sys
import time
import urllib.error
import urllib.request


def _is_strict() -> bool:
    return os.getenv("SOAK_CHAOS_STRICT", "0") == "1"


def _skip_or_fail(reason: str) -> int:
    print('{"status":"skipped","reason":"%s"}' % reason)
    return 1 if _is_strict() else 0


def _emit_progress(
    phase: str,
    state: str,
    timeout_s: float = 0.0,
    metadata: dict | None = None,
) -> None:
    payload = {
        "workflow": "chaos_soak_guard",
        "phase": phase,
        "state": state,
        "timeout_s": max(0.0, float(timeout_s)),
        "ts": int(time.time()),
    }
    if metadata:
        payload["metadata"] = metadata
    print(json.dumps(payload, sort_keys=True))


def _prometheus_available() -> bool:
    try:
        with urllib.request.urlopen(
            "http://localhost:9090/-/ready", timeout=2
        ) as response:
            return response.status == 200
    except (urllib.error.URLError, TimeoutError):
        return False


def _running_node_agents() -> int:
    run = subprocess.run(
        [
            "docker",
            "compose",
            "-f",
            "docker-compose.full.yml",
            "ps",
            "--format",
            "json",
        ],
        text=True,
        capture_output=True,
        timeout=60,
        check=False,
    )
    if run.returncode != 0:
        return 0

    count = 0
    for line in run.stdout.splitlines():
        if "node-agent" in line and '"State":"running"' in line:
            count += 1
    return count


def main() -> int:
    _emit_progress("suite", "started")
    if os.getenv("SOAK_CHAOS_ENABLED", "0") != "1":
        print('{"status":"skipped","reason":"SOAK_CHAOS_ENABLED!=1"}')
        return 0

    if shutil.which("docker") is None:
        return _skip_or_fail("docker not available")

    running_agents = _running_node_agents()
    required_quorum = max(1, int(os.getenv("CHAOS_MIN_CLIENT_QUORUM", "1")))

    if running_agents == 0:
        return _skip_or_fail("no running node-agent containers")

    if running_agents < required_quorum:
        return _skip_or_fail(
            f"insufficient node-agent quorum ({running_agents}<{required_quorum})"
        )

    if not _prometheus_available():
        return _skip_or_fail("prometheus not reachable on localhost:9090")

    _emit_progress("chaos_suite", "started", timeout_s=1200.0)
    run = subprocess.run(
        [sys.executable, "tests/scripts/python/testnet-chaos-suite.py"],
        text=True,
        capture_output=True,
        timeout=1200,
        check=False,
    )
    sys.stdout.write(run.stdout)
    sys.stderr.write(run.stderr)
    _emit_progress(
        "chaos_suite",
        "completed" if run.returncode == 0 else "failed",
        metadata={"returncode": run.returncode},
    )
    return run.returncode


if __name__ == "__main__":
    raise SystemExit(main())
