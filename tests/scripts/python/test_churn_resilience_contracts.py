#!/usr/bin/env python3
"""Targeted churn-resilience contracts for backend reconnect and chaos cadence."""

from __future__ import annotations

import importlib.util
import sys
from pathlib import Path
from types import SimpleNamespace

REPO_ROOT = Path(__file__).resolve().parents[3]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

import sovereignmap_production_backend_v2 as backend


def _load_chaos_module():
    module_path = REPO_ROOT / "tests" / "scripts" / "python" / "testnet-chaos-suite.py"
    spec = importlib.util.spec_from_file_location("testnet_chaos_suite", module_path)
    if spec is None or spec.loader is None:
        raise RuntimeError("unable to load testnet-chaos-suite.py")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def _assert_backend_recoverable_error_detection() -> None:
    class GrpcBridgeClosed(Exception):
        pass

    assert backend._is_recoverable_flower_session_error(GrpcBridgeClosed("closed"))
    assert backend._is_recoverable_flower_session_error(
        RuntimeError("Exception iterating responses")
    )
    assert not backend._is_recoverable_flower_session_error(
        ValueError("invalid static config")
    )


def _assert_backend_strategy_snapshot_restore() -> None:
    original_strategy = backend.strategy
    try:
        backend.strategy = SimpleNamespace(
            round_num=7,
            convergence_history={
                "rounds": [1, 2, 7],
                "accuracies": [70.1, 71.2, 73.9],
                "losses": [2.3, 2.1, 1.8],
                "timestamps": [1.0, 2.0, 3.0],
            },
        )
        round_num, history = backend._snapshot_strategy_state()
        restored = backend._build_strategy_with_snapshot(
            round_num,
            history,
            min_fit_clients=1,
            min_available_clients=1,
        )

        assert round_num == 7
        assert restored.round_num == 7
        assert restored.convergence_history["rounds"][-1] == 7
        assert len(restored.convergence_history["accuracies"]) == 3
    finally:
        backend.strategy = original_strategy


def _assert_chaos_cadence_gate_behaviors() -> None:
    chaos = _load_chaos_module()

    def run_gate(round_series, active_series, baseline_round, quorum_threshold):
        timeline = [0.0]
        round_iter = iter(round_series)
        active_iter = iter(active_series)

        def query_fn(expr: str) -> float:
            if expr == "sovereignmap_fl_round":
                return float(next(round_iter, round_series[-1]))
            if expr == "sovereignmap_active_nodes":
                return float(next(active_iter, active_series[-1]))
            raise AssertionError(f"unexpected query: {expr}")

        def now_fn() -> float:
            return timeline[0]

        def sleep_fn(seconds: float) -> None:
            timeline[0] += float(seconds)

        return chaos.wait_for_round_progress_or_quorum(
            baseline_round=baseline_round,
            quorum_threshold=quorum_threshold,
            timeout_s=1.0,
            poll_interval_s=0.1,
            query_fn=query_fn,
            sleep_fn=sleep_fn,
            now_fn=now_fn,
        )

    # Progress by round advance.
    ok, observed_round, observed_nodes = run_gate(
        round_series=[0.0, 0.0, 1.0],
        active_series=[0.0, 0.0, 0.0],
        baseline_round=0.0,
        quorum_threshold=2.0,
    )
    assert ok and observed_round > 0.0

    # Progress by quorum even before round increment.
    ok, observed_round, observed_nodes = run_gate(
        round_series=[3.0, 3.0, 3.0],
        active_series=[0.0, 1.0, 2.0],
        baseline_round=3.0,
        quorum_threshold=2.0,
    )
    assert ok and observed_nodes >= 2.0

    # No round progress and no quorum should fail.
    ok, observed_round, observed_nodes = run_gate(
        round_series=[4.0, 4.0, 4.0, 4.0],
        active_series=[0.0, 0.0, 0.0, 0.0],
        baseline_round=4.0,
        quorum_threshold=1.0,
    )
    assert not ok


def run() -> int:
    _assert_backend_recoverable_error_detection()
    _assert_backend_strategy_snapshot_restore()
    _assert_chaos_cadence_gate_behaviors()
    print('{"status":"ok","churn_resilience_contracts":"validated"}')
    return 0


if __name__ == "__main__":
    raise SystemExit(run())
