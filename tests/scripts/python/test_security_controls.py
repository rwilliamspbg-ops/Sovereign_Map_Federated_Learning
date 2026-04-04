#!/usr/bin/env python3
"""Security controls contract test for backend request guardrails."""

from __future__ import annotations

import sys
import tempfile
from pathlib import Path
from types import SimpleNamespace

REPO_ROOT = Path(__file__).resolve().parents[3]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

import sovereignmap_production_backend_v2 as backend


def _setup_temp_state(tmpdir: Path) -> None:
    backend.MARKETPLACE_OFFERS_PATH = str(tmpdir / "marketplace_offers.json")
    backend.MARKETPLACE_ROUND_INTENTS_PATH = str(tmpdir / "marketplace_round_intents.json")
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



def _ensure_strategy() -> None:
    if backend.strategy is None:
        backend.strategy = SimpleNamespace(
            round_num=0,
            convergence_history={
                "rounds": [],
                "accuracies": [],
                "losses": [],
                "timestamps": [],
            },
        )



def run() -> int:
    with tempfile.TemporaryDirectory(prefix="security-controls-") as tmp:
        _setup_temp_state(Path(tmp))
        _ensure_strategy()

        backend.JOIN_API_ADMIN_TOKEN = "test-suite-admin-token"
        backend.ALLOW_INSECURE_DEV_ADMIN_TOKEN = False

        # Isolate rate-limit state to keep the test deterministic.
        backend._RATE_LIMIT_WINDOW_SECONDS = 60
        backend._RATE_LIMIT_MAX_REQUESTS = 2
        backend._MUTATION_RATE_LIMIT_MAX_REQUESTS = 2
        backend._rate_limit_bucket.clear()

        client = backend.app.test_client()

        # 1) Protected write endpoint should reject unauthenticated requests.
        unauthorized = client.post(
            "/marketplace/offers",
            json={
                "seller_node_id": "node-x",
                "dataset_fingerprint": "sha256:demo",
                "title": "Demo",
            },
        )
        assert unauthorized.status_code == 401, unauthorized.get_data(as_text=True)

        # 2) Authenticated write request should pass guardrails and reach handler validation.
        authorized = client.post(
            "/marketplace/offers",
            headers={"X-Join-Admin-Token": "test-suite-admin-token"},
            json={
                "seller_node_id": "node-x",
                "dataset_fingerprint": "sha256:demo",
                "title": "Demo",
                "modality": "image",
                "quality_score": 0.9,
                "allowed_tasks": ["classification"],
                "price_per_round": 5,
            },
        )
        assert authorized.status_code == 201, authorized.get_data(as_text=True)

        # 3) API rate limiting should trigger on repeated requests to non-exempt endpoint.
        first = client.get("/ops/trends?limit=5")
        second = client.get("/ops/trends?limit=5")
        third = client.get("/ops/trends?limit=5")
        assert first.status_code == 200
        assert second.status_code == 200
        assert third.status_code == 429, third.get_data(as_text=True)

        # 4) HTTPS enforcement should block non-local insecure requests when local bypass is disabled.
        backend._rate_limit_bucket.clear()
        backend.os.environ["SECURITY_ENFORCE_HTTPS"] = "true"
        backend.os.environ["SECURITY_ALLOW_LOCAL_HTTP"] = "false"
        https_required = client.get("/status", environ_base={"REMOTE_ADDR": "10.10.10.10"})
        assert https_required.status_code == 426, https_required.get_data(as_text=True)

        print('{"status":"ok","security_controls":"validated"}')
    return 0


if __name__ == "__main__":
    raise SystemExit(run())
