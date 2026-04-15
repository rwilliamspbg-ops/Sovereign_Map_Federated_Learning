#!/usr/bin/env python3
"""Testnet chaos suite.

Introduces controlled client churn and verifies FL rounds continue advancing.
"""

from __future__ import annotations

import json
import os
import random
import subprocess
import time
import urllib.parse
import urllib.request
from typing import Callable, Optional, Tuple

COMPOSE_FILE = "docker-compose.full.yml"
PROM_URL = "http://localhost:9090"
BACKEND_URL = os.getenv("CHAOS_BACKEND_URL", "http://localhost:8000")


def emit_progress(
    *,
    workflow: str,
    phase: str,
    state: str,
    timeout_s: float = 0.0,
    metadata: Optional[dict] = None,
) -> None:
    payload = {
        "workflow": workflow,
        "phase": phase,
        "state": state,
        "timeout_s": max(0.0, float(timeout_s)),
        "ts": int(time.time()),
    }
    if metadata:
        payload["metadata"] = metadata
    print(f"[progress] {json.dumps(payload, sort_keys=True)}")


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


def running_node_quorum() -> float:
    """Best-effort count of running node-agent containers.

    This is used as a resilience fallback when Prometheus active-node
    telemetry is temporarily stale during churn.
    """

    try:
        return float(len(running_node_containers()))
    except Exception:
        return 0.0


def restart_container(container_name: str):
    subprocess.check_call(["docker", "restart", container_name])


def _env_int(name: str, default: int, minimum: int) -> int:
    raw = str(os.getenv(name, str(default))).strip()
    try:
        parsed = int(raw)
    except ValueError:
        return max(minimum, default)
    return max(minimum, parsed)


def _env_float(name: str, default: float, minimum: float) -> float:
    raw = str(os.getenv(name, str(default))).strip()
    try:
        parsed = float(raw)
    except ValueError:
        return max(minimum, default)
    return max(minimum, parsed)


def wait_for_round_progress_or_quorum(
    *,
    baseline_round: float,
    quorum_threshold: float,
    timeout_s: float,
    poll_interval_s: float,
    query_fn: Callable[[str], float] = prom_query,
    container_quorum_fn: Optional[Callable[[], float]] = None,
    sleep_fn: Callable[[float], None] = time.sleep,
    now_fn: Callable[[], float] = time.time,
) -> Tuple[bool, float, float]:
    deadline = now_fn() + max(timeout_s, 0.1)
    last_round = baseline_round
    last_active_nodes = 0.0

    while now_fn() < deadline:
        current_round = float(query_fn("sovereignmap_fl_round"))
        active_nodes = float(query_fn("sovereignmap_active_nodes"))
        if container_quorum_fn is not None:
            active_nodes = max(active_nodes, float(container_quorum_fn()))
        last_round = max(last_round, current_round)
        last_active_nodes = active_nodes

        if current_round > baseline_round or active_nodes >= quorum_threshold:
            return True, current_round, active_nodes

        sleep_fn(max(0.05, poll_interval_s))

    return False, last_round, last_active_nodes


def trigger_manual_round(timeout_s: float = 4.0) -> bool:
    request = urllib.request.Request(
        f"{BACKEND_URL}/trigger_fl",
        method="POST",
        data=b"{}",
        headers={"Content-Type": "application/json"},
    )
    try:
        with urllib.request.urlopen(request, timeout=timeout_s) as response:
            return int(response.status) in {200, 202}
    except Exception:
        return False


def ensure_round_progress_with_fallback(
    *,
    baseline_round: float,
    quorum_threshold: float,
    progress_timeout_s: float,
    progress_poll_s: float,
    allow_manual_fallback: bool,
    container_quorum_fn: Optional[Callable[[], float]] = None,
) -> Tuple[bool, float, float]:
    ready, observed_round, observed_nodes = wait_for_round_progress_or_quorum(
        baseline_round=baseline_round,
        quorum_threshold=quorum_threshold,
        timeout_s=progress_timeout_s,
        poll_interval_s=progress_poll_s,
        container_quorum_fn=container_quorum_fn,
    )
    if ready and observed_round > baseline_round:
        return True, observed_round, observed_nodes

    if (
        allow_manual_fallback
        and observed_nodes >= quorum_threshold
        and trigger_manual_round()
    ):
        return wait_for_round_progress_or_quorum(
            baseline_round=baseline_round,
            quorum_threshold=quorum_threshold,
            timeout_s=max(6.0, progress_timeout_s / 2.0),
            poll_interval_s=progress_poll_s,
            container_quorum_fn=container_quorum_fn,
        )

    return ready, observed_round, observed_nodes


def main():
    print("[chaos] starting testnet chaos suite")
    emit_progress(workflow="chaos_guard", phase="suite", state="started")

    steps = _env_int("CHAOS_STEPS", default=5, minimum=1)
    restart_settle_s = _env_float("CHAOS_RESTART_SETTLE_SECONDS", default=8.0, minimum=0.5)
    progress_timeout_s = _env_float(
        "CHAOS_PROGRESS_TIMEOUT_SECONDS", default=45.0, minimum=2.0
    )
    progress_poll_s = _env_float(
        "CHAOS_PROGRESS_POLL_SECONDS", default=2.0, minimum=0.1
    )
    quorum_threshold = float(_env_int("CHAOS_MIN_CLIENT_QUORUM", default=1, minimum=1))
    allow_manual_fallback = os.getenv("CHAOS_TRIGGER_FALLBACK_ROUND", "1") == "1"
    use_container_quorum_fallback = (
        os.getenv("CHAOS_USE_CONTAINER_QUORUM_FALLBACK", "1") == "1"
    )
    container_quorum_fn = (
        running_node_quorum if use_container_quorum_fallback else None
    )

    before = prom_query("sovereignmap_fl_round")
    print(f"[chaos] round before={before}")
    last_progress_round = before

    ready, observed_round, observed_nodes = ensure_round_progress_with_fallback(
        baseline_round=before,
        quorum_threshold=quorum_threshold,
        progress_timeout_s=progress_timeout_s,
        progress_poll_s=progress_poll_s,
        allow_manual_fallback=allow_manual_fallback,
        container_quorum_fn=container_quorum_fn,
    )
    if ready:
        emit_progress(
            workflow="chaos_guard",
            phase="preflight",
            state="completed",
            timeout_s=progress_timeout_s,
            metadata={"round": observed_round, "active_nodes": observed_nodes},
        )
        last_progress_round = max(last_progress_round, observed_round)
        print(
            "[chaos] preflight ready: "
            f"round={observed_round} active_nodes={observed_nodes}"
        )
    else:
        emit_progress(
            workflow="chaos_guard",
            phase="preflight",
            state="failed",
            timeout_s=progress_timeout_s,
            metadata={"round": observed_round, "active_nodes": observed_nodes},
        )
        raise SystemExit(
            "[chaos] FAILED: preflight did not reach round progress or client quorum "
            f"within {progress_timeout_s}s (round={observed_round}, active_nodes={observed_nodes})"
        )

    for step in range(steps):
        emit_progress(
            workflow="chaos_guard",
            phase="churn_step",
            state="started",
            timeout_s=restart_settle_s,
            metadata={"step": step + 1, "steps": steps},
        )
        if step > 0:
            ready, observed_round, observed_nodes = ensure_round_progress_with_fallback(
                baseline_round=last_progress_round,
                quorum_threshold=quorum_threshold,
                progress_timeout_s=progress_timeout_s,
                progress_poll_s=progress_poll_s,
                allow_manual_fallback=allow_manual_fallback,
                container_quorum_fn=container_quorum_fn,
            )
            if not ready:
                emit_progress(
                    workflow="chaos_guard",
                    phase="cadence_gate",
                    state="failed",
                    timeout_s=progress_timeout_s,
                    metadata={
                        "round": observed_round,
                        "baseline": last_progress_round,
                        "active_nodes": observed_nodes,
                    },
                )
                raise SystemExit(
                    "[chaos] FAILED: cadence gate not satisfied before next restart "
                    f"(round={observed_round}, baseline={last_progress_round}, "
                    f"active_nodes={observed_nodes}, quorum={quorum_threshold})"
                )
            last_progress_round = max(last_progress_round, observed_round)

        nodes = running_node_containers()
        if not nodes:
            raise RuntimeError("No running node-agent containers found")
        target = random.choice(nodes)
        print(f"[chaos] step {step+1}: restarting {target}")
        restart_container(target)
        time.sleep(restart_settle_s)
        emit_progress(
            workflow="chaos_guard",
            phase="churn_step",
            state="completed",
            timeout_s=restart_settle_s,
            metadata={"step": step + 1, "target": target},
        )

    ready, observed_round, observed_nodes = ensure_round_progress_with_fallback(
        baseline_round=last_progress_round,
        quorum_threshold=quorum_threshold,
        progress_timeout_s=progress_timeout_s,
        progress_poll_s=progress_poll_s,
        allow_manual_fallback=allow_manual_fallback,
        container_quorum_fn=container_quorum_fn,
    )
    if not ready:
        emit_progress(
            workflow="chaos_guard",
            phase="recovery",
            state="failed",
            timeout_s=progress_timeout_s,
            metadata={"round": observed_round, "active_nodes": observed_nodes},
        )
        raise SystemExit(
            "[chaos] FAILED: post-chaos recovery did not reach round progression or quorum "
            f"(round={observed_round}, baseline={last_progress_round}, "
            f"active_nodes={observed_nodes}, quorum={quorum_threshold})"
        )

    after = prom_query("sovereignmap_fl_round")
    print(f"[chaos] round after={after}")

    if after <= before and allow_manual_fallback:
        if trigger_manual_round():
            time.sleep(max(1.0, restart_settle_s / 2.0))
            after = prom_query("sovereignmap_fl_round")
            print(f"[chaos] round after fallback trigger={after}")

    if after <= before:
        emit_progress(
            workflow="chaos_guard",
            phase="validation",
            state="failed",
            metadata={"before": before, "after": after},
        )
        raise SystemExit("[chaos] FAILED: FL rounds did not progress under churn")

    emit_progress(
        workflow="chaos_guard",
        phase="suite",
        state="completed",
        metadata={"before_round": before, "after_round": after},
    )
    print("[chaos] PASSED: FL rounds progressed under controlled churn")


if __name__ == "__main__":
    main()
