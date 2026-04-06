#!/usr/bin/env python3
"""Performance budget regression tests for backend/runtime and UI throttle settings."""

from __future__ import annotations

import json
import statistics
import sys
import tempfile
import time
from pathlib import Path
from types import SimpleNamespace

REPO_ROOT = Path(__file__).resolve().parents[3]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

import sovereignmap_production_backend_v2 as backend


MAX_ROUND_DRIFT_MS = 80.0
MAX_TELEMETRY_OVERHEAD_MS = 40.0
MIN_CHART_THROTTLE_MS = 200



def _seed_temp_state(tmpdir: Path) -> None:
    backend.MARKETPLACE_OFFERS_PATH = str(tmpdir / "marketplace_offers.json")
    backend.MARKETPLACE_ROUND_INTENTS_PATH = str(tmpdir / "marketplace_round_intents.json")
    backend.MARKETPLACE_CONTRACTS_PATH = str(tmpdir / "marketplace_contracts.json")
    backend.MARKETPLACE_DISPUTES_PATH = str(tmpdir / "marketplace_disputes.json")
    backend.GOVERNANCE_ACTION_LOG_PATH = str(tmpdir / "governance_actions.json")
    backend.GOVERNANCE_PROPOSALS_PATH = str(tmpdir / "governance_proposals.json")
    backend.JOIN_INVITES_PATH = str(tmpdir / "join_invites.json")
    backend.JOIN_REGISTRATIONS_PATH = str(tmpdir / "join_registrations.json")
    backend.JOIN_INVITE_REQUESTS_PATH = str(tmpdir / "join_invite_requests.json")
    backend.MODEL_REGISTRY_PATH = str(tmpdir / "model_registry.jsonl")

    for target in [
        backend.MARKETPLACE_OFFERS_PATH,
        backend.MARKETPLACE_ROUND_INTENTS_PATH,
        backend.MARKETPLACE_CONTRACTS_PATH,
        backend.MARKETPLACE_DISPUTES_PATH,
        backend.GOVERNANCE_ACTION_LOG_PATH,
        backend.GOVERNANCE_PROPOSALS_PATH,
        backend.JOIN_INVITES_PATH,
        backend.JOIN_REGISTRATIONS_PATH,
        backend.JOIN_INVITE_REQUESTS_PATH,
    ]:
        Path(target).parent.mkdir(parents=True, exist_ok=True)
        Path(target).write_text("[]", encoding="utf-8")



def _benchmark_rounds(rounds: int, with_telemetry: bool, warmup_rounds: int = 5) -> list[float]:
    durations_ms = []

    if with_telemetry:
        publish_tpm = backend.publish_tpm_attestation_events
        publish_tok = backend.publish_tokenomics_event
    else:
        publish_tpm = backend.publish_tpm_attestation_events
        publish_tok = backend.publish_tokenomics_event
        backend.publish_tpm_attestation_events = lambda _count: None
        backend.publish_tokenomics_event = lambda _payload: None

    try:
        # Warm-up rounds absorb one-time startup noise in constrained CI containers.
        for _ in range(warmup_rounds):
            backend.execute_manual_fl_round(reason="perf_budget_probe")

        for _ in range(rounds):
            start = time.perf_counter()
            backend.execute_manual_fl_round(reason="perf_budget_probe")
            durations_ms.append((time.perf_counter() - start) * 1000.0)
    finally:
        if not with_telemetry:
            backend.publish_tpm_attestation_events = publish_tpm
            backend.publish_tokenomics_event = publish_tok

    return durations_ms



def _extract_chart_throttle_ms() -> int:
    path = REPO_ROOT / "frontend" / "src" / "BrowserFLDemo.jsx"
    source = path.read_text(encoding="utf-8")
    marker = "VITE_CHART_THROTTLE_MS',"
    idx = source.find(marker)
    if idx < 0:
        raise AssertionError("VITE_CHART_THROTTLE_MS marker not found")

    tail = source[idx : idx + 80]
    digits = ""
    for ch in tail:
        if ch.isdigit():
            digits += ch
        elif digits:
            break
    if not digits:
        raise AssertionError("Unable to parse chart throttle default")
    return int(digits)



def run() -> int:
    with tempfile.TemporaryDirectory(prefix="perf-thresholds-") as tmp:
        _seed_temp_state(Path(tmp))

        backend.strategy = SimpleNamespace(
            round_num=0,
            convergence_history={
                "rounds": [],
                "accuracies": [],
                "losses": [],
                "timestamps": [],
            },
        )

        # Avoid network waits in performance test.
        backend.os.environ["ALLOW_INSECURE_METRICS_ENDPOINTS"] = "false"

        with_telemetry = _benchmark_rounds(rounds=30, with_telemetry=True)
        no_telemetry = _benchmark_rounds(rounds=30, with_telemetry=False)

        sorted_with = sorted(with_telemetry)
        p10 = sorted_with[max(0, int(len(sorted_with) * 0.10) - 1)]
        p90 = sorted_with[min(len(sorted_with) - 1, int(len(sorted_with) * 0.90))]
        drift_ms = p90 - p10
        avg_overhead_ms = statistics.mean(with_telemetry) - statistics.mean(no_telemetry)

        assert drift_ms <= MAX_ROUND_DRIFT_MS, (
            f"round drift budget exceeded: {drift_ms:.2f}ms > {MAX_ROUND_DRIFT_MS:.2f}ms"
        )
        assert avg_overhead_ms <= MAX_TELEMETRY_OVERHEAD_MS, (
            f"telemetry overhead budget exceeded: {avg_overhead_ms:.2f}ms > {MAX_TELEMETRY_OVERHEAD_MS:.2f}ms"
        )

        chart_throttle_ms = _extract_chart_throttle_ms()
        assert chart_throttle_ms >= MIN_CHART_THROTTLE_MS, (
            f"chart throttle too low: {chart_throttle_ms}ms < {MIN_CHART_THROTTLE_MS}ms"
        )

        print(
            json.dumps(
                {
                    "status": "ok",
                    "drift_ms": round(drift_ms, 3),
                    "telemetry_overhead_ms": round(avg_overhead_ms, 3),
                    "chart_throttle_ms": chart_throttle_ms,
                }
            )
        )
    return 0


if __name__ == "__main__":
    raise SystemExit(run())
