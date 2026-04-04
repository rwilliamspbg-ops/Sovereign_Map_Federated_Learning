#!/usr/bin/env python3
"""Negative-path checks for local marketplace API contracts."""

import json
import os
import sys
import tempfile
import time
from pathlib import Path
from types import SimpleNamespace

REPO_ROOT = Path(__file__).resolve().parents[3]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

import sovereignmap_production_backend_v2 as backend


def _configure_temp_marketplace_state(tmpdir: Path) -> None:
    backend.MARKETPLACE_OFFERS_PATH = str(tmpdir / "marketplace_offers.json")
    backend.MARKETPLACE_ROUND_INTENTS_PATH = str(
        tmpdir / "marketplace_round_intents.json"
    )
    backend.MARKETPLACE_CONTRACTS_PATH = str(tmpdir / "marketplace_contracts.json")
    backend.MARKETPLACE_DISPUTES_PATH = str(tmpdir / "marketplace_disputes.json")
    backend.GOVERNANCE_ACTION_LOG_PATH = str(tmpdir / "governance_actions.json")
    backend.GOVERNANCE_PROPOSALS_PATH = str(tmpdir / "governance_proposals.json")
    backend.COMPUTE_ATTESTATIONS_PATH = str(tmpdir / "compute_attestations.json")
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
        backend.COMPUTE_ATTESTATIONS_PATH,
        backend.JOIN_INVITES_PATH,
        backend.JOIN_REGISTRATIONS_PATH,
        backend.JOIN_INVITE_REQUESTS_PATH,
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
    with tempfile.TemporaryDirectory(prefix="marketplace-negative-") as tmp:
        _configure_temp_marketplace_state(Path(tmp))
        _ensure_strategy()
        backend.JOIN_API_ADMIN_TOKEN = "test-suite-admin-token"
        admin_headers = {"X-Join-Admin-Token": backend.JOIN_API_ADMIN_TOKEN}

        client = backend.app.test_client()

        missing_intent_resp = client.post(
            "/marketplace/match",
            json={"max_offers": 1},
            headers=admin_headers,
        )
        assert missing_intent_resp.status_code == 400
        assert missing_intent_resp.get_json().get("code") == "round_intent_id_required"

        expired_offer_resp = client.post(
            "/marketplace/offers",
            json={
                "seller_node_id": "node-expired",
                "dataset_fingerprint": "sha256:expired-offer",
                "title": "Expired Pack",
                "modality": "image",
                "quality_score": 0.95,
                "allowed_tasks": ["classification"],
                "price_per_round": 10,
                "expires_at": int(time.time()) - 10,
            },
            headers=admin_headers,
        )
        assert expired_offer_resp.status_code == 201

        intent_resp = client.post(
            "/marketplace/round_intents",
            json={
                "model_owner_id": "owner-negative",
                "task_type": "classification",
                "required_modalities": ["image"],
                "min_quality_score": 0.8,
                "budget_total": 20,
            },
            headers=admin_headers,
        )
        assert intent_resp.status_code == 201
        intent_id = intent_resp.get_json()["round_intent_id"]

        no_match_resp = client.post(
            "/marketplace/match",
            json={"round_intent_id": intent_id, "max_offers": 2},
            headers=admin_headers,
        )
        assert no_match_resp.status_code == 422
        no_match_payload = no_match_resp.get_json()
        assert no_match_payload.get("code") == "no_compatible_offers_found"
        reasons = (no_match_payload.get("details") or {}).get("rejection_reasons") or {}
        assert "offer_expired" in reasons

        active_offer_resp = client.post(
            "/marketplace/offers",
            json={
                "seller_node_id": "node-active",
                "dataset_fingerprint": "sha256:active-offer",
                "title": "Active Pack",
                "modality": "image",
                "quality_score": 0.91,
                "allowed_tasks": ["classification"],
                "price_per_round": 5,
            },
            headers=admin_headers,
        )
        assert active_offer_resp.status_code == 201

        ok_match_resp = client.post(
            "/marketplace/match",
            json={"round_intent_id": intent_id, "max_offers": 2},
            headers=admin_headers,
        )
        assert ok_match_resp.status_code == 201
        contract_id = ok_match_resp.get_json()["contract_id"]

        invalid_transition_resp = client.patch(
            f"/marketplace/round_intents/{intent_id}",
            json={"status": "cancelled"},
            headers=admin_headers,
        )
        assert invalid_transition_resp.status_code == 409
        assert (
            invalid_transition_resp.get_json().get("code")
            == "invalid_status_transition"
        )

        first_release_resp = client.post(
            "/marketplace/escrow/release",
            json={"contract_id": contract_id},
            headers=admin_headers,
        )
        assert first_release_resp.status_code == 200

        second_release_resp = client.post(
            "/marketplace/escrow/release",
            json={"contract_id": contract_id},
            headers=admin_headers,
        )
        assert second_release_resp.status_code == 409
        assert second_release_resp.get_json().get("code") == "contract_already_released"

        bad_preview_resp = client.post(
            "/marketplace/policy/preview",
            json={"round_intent_id": intent_id, "max_offers": "abc"},
            headers=admin_headers,
        )
        assert bad_preview_resp.status_code == 400
        assert bad_preview_resp.get_json().get("code") == "invalid_max_offers"

        missing_preview_intent_resp = client.post(
            "/marketplace/policy/preview",
            json={"round_intent_id": "intent-does-not-exist", "max_offers": 1},
            headers=admin_headers,
        )
        assert missing_preview_intent_resp.status_code == 404
        assert (
            missing_preview_intent_resp.get_json().get("code")
            == "round_intent_not_found"
        )

        missing_title_proposal_resp = client.post(
            "/governance/proposals",
            json={"proposal_type": "policy_update", "created_by": "test-suite"},
            headers=admin_headers,
        )
        assert missing_title_proposal_resp.status_code == 400
        assert (
            missing_title_proposal_resp.get_json().get("code")
            == "proposal_title_required"
        )

        create_proposal_resp = client.post(
            "/governance/proposals",
            json={
                "title": "Close voting window",
                "proposal_type": "operational",
                "created_by": "test-suite",
            },
            headers=admin_headers,
        )
        assert create_proposal_resp.status_code == 201
        proposal_id = create_proposal_resp.get_json()["proposal_id"]

        invalid_vote_resp = client.post(
            f"/governance/proposals/{proposal_id}/vote",
            json={"voter": "node-1", "decision": "maybe", "weight": 1.0},
            headers=admin_headers,
        )
        assert invalid_vote_resp.status_code == 400
        assert invalid_vote_resp.get_json().get("code") == "invalid_vote_decision"

        close_proposal_resp = client.patch(
            f"/governance/proposals/{proposal_id}",
            json={"status": "closed", "actor": "moderator"},
            headers=admin_headers,
        )
        assert close_proposal_resp.status_code == 200

        vote_closed_resp = client.post(
            f"/governance/proposals/{proposal_id}/vote",
            json={"voter": "node-1", "decision": "yes", "weight": 1.0},
            headers=admin_headers,
        )
        assert vote_closed_resp.status_code == 409
        assert vote_closed_resp.get_json().get("code") == "proposal_not_open"

        missing_participant_resp = client.post(
            "/attestations/share",
            json={
                "compute_type": "gpu",
                "attestation_status": "verified",
            },
        )
        assert missing_participant_resp.status_code == 400
        assert (
            missing_participant_resp.get_json().get("code")
            == "participant_name_required"
        )

        invalid_attestation_status_resp = client.post(
            "/attestations/share",
            json={
                "participant_name": "node-invalid",
                "compute_type": "gpu",
                "attestation_status": "unknown-status",
            },
        )
        assert invalid_attestation_status_resp.status_code == 400
        assert (
            invalid_attestation_status_resp.get_json().get("code")
            == "invalid_attestation_status"
        )

        missing_email_request_resp = client.post(
            "/join/request_invite",
            json={
                "participant_name": "node-without-email",
                "compute_type": "gpu",
            },
        )
        assert missing_email_request_resp.status_code == 400
        assert (
            missing_email_request_resp.get_json().get("code")
            == "valid_contact_email_required"
        )

        missing_compute_type_request_resp = client.post(
            "/join/request_invite",
            json={
                "participant_name": "node-without-compute",
                "contact_email": "node@example.org",
            },
        )
        assert missing_compute_type_request_resp.status_code == 400
        assert (
            missing_compute_type_request_resp.get_json().get("code")
            == "compute_type_required"
        )

        unauthorized_requests_resp = client.get("/join/invite_requests?status=pending")
        assert unauthorized_requests_resp.status_code == 401

        os.environ["ADMIN_WALLET_ALLOWLIST"] = "0xabc"
        os.environ.pop("ADMIN_WALLET_HEADER_AUTH_ENABLED", None)
        wallet_only_disabled_resp = client.get(
            "/join/invite_requests?status=pending",
            headers={"X-Admin-Wallet": "0xabc"},
        )
        assert wallet_only_disabled_resp.status_code == 401

        os.environ["ADMIN_WALLET_HEADER_AUTH_ENABLED"] = "true"
        wallet_only_enabled_resp = client.get(
            "/join/invite_requests?status=pending",
            headers={"X-Admin-Wallet": "0xabc"},
        )
        assert wallet_only_enabled_resp.status_code == 200
        os.environ.pop("ADMIN_WALLET_HEADER_AUTH_ENABLED", None)
        os.environ.pop("ADMIN_WALLET_ALLOWLIST", None)

        admin_headers = {"X-Join-Admin-Token": backend.JOIN_API_ADMIN_TOKEN}

        invalid_admin_invite_limits_resp = client.post(
            "/join/invite",
            json={
                "participant_name": "bad-limits-node",
                "max_uses": "abc",
                "expires_in_hours": "xyz",
            },
            headers=admin_headers,
        )
        assert invalid_admin_invite_limits_resp.status_code == 400
        assert (
            invalid_admin_invite_limits_resp.get_json().get("code")
            == "invalid_invite_limits"
        )

        invalid_sort_resp = client.get(
            "/join/invite_requests?sort_by=unknown_field&sort_dir=sideways",
            headers=admin_headers,
        )
        assert invalid_sort_resp.status_code == 200
        invalid_sort_payload = invalid_sort_resp.get_json()
        assert invalid_sort_payload.get("sort_by") == "created_at"
        assert invalid_sort_payload.get("sort_dir") == "desc"

        missing_request_approve_resp = client.post(
            "/join/invite_requests/join-req-missing/approve",
            json={"max_uses": 1, "expires_in_hours": 24},
            headers=admin_headers,
        )
        assert missing_request_approve_resp.status_code == 404
        assert (
            missing_request_approve_resp.get_json().get("code") == "request_not_found"
        )

        invalid_request_approve_limits_resp = client.post(
            "/join/request_invite",
            json={
                "participant_name": "approve-limits-node",
                "contact_email": "approve-limits@example.org",
                "compute_type": "gpu",
            },
        )
        assert invalid_request_approve_limits_resp.status_code == 201
        pending_request_id = invalid_request_approve_limits_resp.get_json()[
            "request_id"
        ]

        invalid_request_limits_resp = client.post(
            f"/join/invite_requests/{pending_request_id}/approve",
            json={"max_uses": "abc", "expires_in_hours": "xyz"},
            headers=admin_headers,
        )
        assert invalid_request_limits_resp.status_code == 400
        assert (
            invalid_request_limits_resp.get_json().get("code")
            == "invalid_invite_limits"
        )

        missing_request_reject_resp = client.post(
            "/join/invite_requests/join-req-missing/reject",
            json={"reason": "missing request"},
            headers=admin_headers,
        )
        assert missing_request_reject_resp.status_code == 404
        assert missing_request_reject_resp.get_json().get("code") == "request_not_found"

        missing_invite_revoke_resp = client.post(
            "/join/invites/invite-missing/revoke",
            headers=admin_headers,
        )
        assert missing_invite_revoke_resp.status_code == 404
        assert missing_invite_revoke_resp.get_json().get("code") == "invite_not_found"

        print(json.dumps({"status": "ok", "negative_checks": 20}))

    return 0


if __name__ == "__main__":
    raise SystemExit(run())
