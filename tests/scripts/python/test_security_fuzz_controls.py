#!/usr/bin/env python3
"""Fuzz/property-style security abuse tests for high-risk endpoints."""

from __future__ import annotations

import base64
import os
import random
import string
import sys
import tempfile
from pathlib import Path
from types import SimpleNamespace

REPO_ROOT = Path(__file__).resolve().parents[3]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

import sovereignmap_production_backend_v2 as backend


def _seed_state(tmpdir: Path) -> None:
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


def _rnd_token(length: int) -> str:
    alphabet = string.ascii_letters + string.digits + "-_:."
    return "".join(random.choice(alphabet) for _ in range(length))


def run() -> int:
    with tempfile.TemporaryDirectory(prefix="security-fuzz-") as tmp:
        _seed_state(Path(tmp))

        backend.JOIN_API_ADMIN_TOKEN = "fuzz-admin-token"
        backend.ALLOW_INSECURE_DEV_ADMIN_TOKEN = False
        backend._RATE_LIMIT_WINDOW_SECONDS = 60
        backend._RATE_LIMIT_MAX_REQUESTS = 100
        backend._MUTATION_RATE_LIMIT_MAX_REQUESTS = 8
        backend._rate_limit_bucket.clear()

        backend.strategy = SimpleNamespace(
            round_num=0,
            convergence_history={
                "rounds": [],
                "accuracies": [],
                "losses": [],
                "timestamps": [],
            },
        )

        os.environ["SECURITY_ENFORCE_HTTPS"] = "false"
        os.environ["MOBILE_REQUIRE_ATTESTATION"] = "true"

        client = backend.app.test_client()

        # 1) Malformed bearer/header variants should not bypass auth.
        for _ in range(30):
            malformed = _rnd_token(random.randint(1, 120))
            resp = client.post(
                "/simulate/networkPartitions",
                headers={"Authorization": malformed},
            )
            assert resp.status_code in {401, 429}

            resp2 = client.post(
                "/simulate/hardwareFaults",
                headers={"Authorization": f"Bearer {malformed}"},
            )
            assert resp2.status_code in {401, 429}

        # 2) Replay-like burst should trigger mutation rate limits.
        backend._rate_limit_bucket.clear()
        statuses = []
        for _ in range(12):
            resp = client.post(
                "/simulate/byzantineAttacks",
                headers={"X-Join-Admin-Token": backend.JOIN_API_ADMIN_TOKEN},
            )
            statuses.append(resp.status_code)
        assert 429 in statuses, "expected rate limit status in burst run"

        # 3) Oversized payload should be rejected by gradient verify endpoint.
        huge = base64.b64encode(b"A" * 2_100_000).decode("ascii")
        oversized = client.post(
            "/mobile/verify_gradient",
            json={
                "node_id": "node-1",
                "round": 1,
                "signer_alias": "fuzz",
                "public_key_pem": "-----BEGIN PUBLIC KEY-----\ninvalid\n-----END PUBLIC KEY-----",
                "gradient_payload_b64": huge,
                "gradient_signature_b64": "ZmFrZQ==",
                "attestation_payload_b64": "e30=",
            },
        )
        assert oversized.status_code == 413

        # 4) Property-style malformed payload fuzzing should fail safely (4xx, never 500).
        for _ in range(50):
            fuzz_payload = {
                "node_id": _rnd_token(random.randint(0, 16)),
                "round": random.choice([None, -1, "x", random.randint(-10, 10)]),
                "signer_alias": _rnd_token(random.randint(0, 12)),
                "public_key_pem": _rnd_token(random.randint(0, 80)),
                "gradient_payload_b64": _rnd_token(random.randint(0, 128)),
                "gradient_signature_b64": _rnd_token(random.randint(0, 128)),
                "attestation_payload_b64": _rnd_token(random.randint(0, 64)),
            }
            resp = client.post("/mobile/verify_gradient", json=fuzz_payload)
            assert resp.status_code in {400, 413, 429}

        print('{"status":"ok","security_fuzz":"validated"}')
    return 0


if __name__ == "__main__":
    raise SystemExit(run())
