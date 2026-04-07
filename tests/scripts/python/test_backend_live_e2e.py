#!/usr/bin/env python3
"""Live end-to-end HTTP tests for backend guardrails and streaming behavior."""

from __future__ import annotations

import json
import multiprocessing as mp
import os
import socket
import sys
import tempfile
import time
from pathlib import Path
from types import SimpleNamespace
from urllib.request import Request, urlopen
from urllib.error import HTTPError

import requests

REPO_ROOT = Path(__file__).resolve().parents[3]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))


def _free_port() -> int:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        sock.bind(("127.0.0.1", 0))
        return int(sock.getsockname()[1])


def _seed_state(tmpdir: Path, backend) -> None:
    backend.MARKETPLACE_OFFERS_PATH = str(tmpdir / "marketplace_offers.json")
    backend.MARKETPLACE_ROUND_INTENTS_PATH = str(
        tmpdir / "marketplace_round_intents.json"
    )
    backend.MARKETPLACE_CONTRACTS_PATH = str(tmpdir / "marketplace_contracts.json")
    backend.MARKETPLACE_DISPUTES_PATH = str(tmpdir / "marketplace_disputes.json")
    backend.GOVERNANCE_ACTION_LOG_PATH = str(tmpdir / "governance_actions.json")
    backend.GOVERNANCE_PROPOSALS_PATH = str(tmpdir / "governance_proposals.json")
    backend.JOIN_INVITES_PATH = str(tmpdir / "join_invites.json")
    backend.JOIN_REGISTRATIONS_PATH = str(tmpdir / "join_registrations.json")
    backend.JOIN_INVITE_REQUESTS_PATH = str(tmpdir / "join_invite_requests.json")
    backend.COMPUTE_ATTESTATIONS_PATH = str(tmpdir / "compute_attestations.json")

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
        backend.COMPUTE_ATTESTATIONS_PATH,
    ]:
        Path(target).parent.mkdir(parents=True, exist_ok=True)
        Path(target).write_text("[]", encoding="utf-8")


def _run_server(port: int, token: str, tmp_dir: str) -> None:
    import sovereignmap_production_backend_v2 as backend

    backend.JOIN_API_ADMIN_TOKEN = token
    backend.strategy = SimpleNamespace(
        round_num=0,
        convergence_history={
            "rounds": [],
            "accuracies": [],
            "losses": [],
            "timestamps": [],
        },
    )

    backend._RATE_LIMIT_WINDOW_SECONDS = 60
    backend._RATE_LIMIT_MAX_REQUESTS = 2
    backend._MUTATION_RATE_LIMIT_MAX_REQUESTS = 2

    os.environ["SECURITY_ENFORCE_HTTPS"] = "true"
    os.environ["SECURITY_ALLOW_LOCAL_HTTP"] = "false"

    _seed_state(Path(tmp_dir), backend)

    backend.app.run(
        host="127.0.0.1", port=port, debug=False, threaded=True, use_reloader=False
    )


def _wait_ready(base_url: str, timeout_s: float = 8.0) -> bool:
    deadline = time.time() + timeout_s
    while time.time() < deadline:
        try:
            req = Request(f"{base_url}/status", method="GET")
            urlopen(req, timeout=0.8)
            return True
        except HTTPError as exc:
            # With HTTPS enforcement enabled, readiness can return 426 over HTTP.
            if exc.code in {200, 426}:
                return True
        except Exception:
            time.sleep(0.15)
    return False


def run() -> int:
    token = "live-e2e-admin-token"
    port = _free_port()
    base = f"http://127.0.0.1:{port}"

    with tempfile.TemporaryDirectory(prefix="backend-live-e2e-") as tmp:
        proc = mp.Process(target=_run_server, args=(port, token, tmp), daemon=True)
        proc.start()

        try:
            assert _wait_ready(base), "backend did not become ready"

            # HTTP over non-HTTPS should be rejected when HTTPS is enforced.
            try:
                urlopen(Request(f"{base}/status", method="GET"), timeout=1.0)
                raise AssertionError("Expected HTTPS enforcement failure")
            except HTTPError as exc:
                assert exc.code == 426

            # Use proxy-header HTTPS simulation to continue test traffic.
            secure_headers = {"X-Forwarded-Proto": "https"}

            status_resp = requests.get(
                f"{base}/status", headers=secure_headers, timeout=2
            )
            assert status_resp.status_code == 200

            # Protected mutating endpoint blocks unauthenticated requests.
            unauth = requests.post(
                f"{base}/simulate/byzantineAttacks",
                headers=secure_headers,
                timeout=2,
            )
            assert unauth.status_code == 401

            # Authenticated mutating endpoint succeeds.
            auth_headers = {
                **secure_headers,
                "X-Join-Admin-Token": token,
                "Content-Type": "application/json",
            }
            auth_ok = requests.post(
                f"{base}/simulate/byzantineAttacks",
                headers=auth_headers,
                timeout=2,
            )
            assert auth_ok.status_code == 200

            # Rate limiting should trigger on third call with configured limit=2.
            r1 = requests.get(
                f"{base}/ops/trends?limit=5", headers=secure_headers, timeout=2
            )
            r2 = requests.get(
                f"{base}/ops/trends?limit=5", headers=secure_headers, timeout=2
            )
            r3 = requests.get(
                f"{base}/ops/trends?limit=5", headers=secure_headers, timeout=2
            )
            assert (
                r1.status_code == 200
                and r2.status_code == 200
                and r3.status_code == 429
            )

            # SSE endpoint should stream data lines when reachable.
            with requests.get(
                f"{base}/ops/events",
                headers=secure_headers,
                stream=True,
                timeout=4,
            ) as stream_resp:
                assert stream_resp.status_code == 200
                found_data_line = False
                for line in stream_resp.iter_lines(decode_unicode=True):
                    if not line:
                        continue
                    if line.startswith("data:"):
                        found_data_line = True
                        break
                assert found_data_line, "No SSE data line received"

            print(json.dumps({"status": "ok", "live_e2e": "validated"}))
            return 0
        finally:
            if proc.is_alive():
                proc.terminate()
                proc.join(timeout=3)


if __name__ == "__main__":
    raise SystemExit(run())
