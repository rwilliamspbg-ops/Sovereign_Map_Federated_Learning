#!/usr/bin/env python3
"""Local marketplace API contract smoke test for backend v2."""

import json
import tempfile
from pathlib import Path
from types import SimpleNamespace
import sys

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
    with tempfile.TemporaryDirectory(prefix="marketplace-local-") as tmp:
        tmpdir = Path(tmp)
        _configure_temp_marketplace_state(tmpdir)
        _ensure_strategy()
        backend.JOIN_API_ADMIN_TOKEN = "test-suite-admin-token"
        admin_headers = {"X-Join-Admin-Token": backend.JOIN_API_ADMIN_TOKEN}

        client = backend.app.test_client()

        offer_resp = client.post(
            "/marketplace/offers",
            json={
                "seller_node_id": "node-1",
                "dataset_fingerprint": "sha256:test-dataset-1",
                "title": "Vision Edge Pack",
                "modality": "image",
                "quality_score": 0.85,
                "allowed_tasks": ["classification"],
                "price_per_round": 11.5,
                "min_rounds": 1,
                "attestation_status": "verified",
            },
            headers=admin_headers,
        )
        assert offer_resp.status_code == 201, offer_resp.get_data(as_text=True)
        offer = offer_resp.get_json()
        assert offer and offer.get("offer_id")

        intent_resp = client.post(
            "/marketplace/round_intents",
            json={
                "model_owner_id": "hud-operator",
                "task_type": "classification",
                "required_modalities": ["image"],
                "min_quality_score": 0.7,
                "budget_total": 30.0,
            },
            headers=admin_headers,
        )
        assert intent_resp.status_code == 201, intent_resp.get_data(as_text=True)
        intent = intent_resp.get_json()
        intent_id = intent["round_intent_id"]

        list_intents_resp = client.get("/marketplace/round_intents?status=open")
        assert list_intents_resp.status_code == 200
        listed = list_intents_resp.get_json()
        assert listed["count"] >= 1

        match_resp = client.post(
            "/marketplace/match",
            json={"round_intent_id": intent_id, "max_offers": 2},
            headers=admin_headers,
        )
        assert match_resp.status_code == 201, match_resp.get_data(as_text=True)
        contract = match_resp.get_json()
        contract_id = contract["contract_id"]
        assert contract["payout_status"] == "pending"
        assert contract["selected_offers"]
        assert "score_breakdown" in contract["selected_offers"][0]
        assert "selection_diagnostics" in contract

        trigger_resp = client.post("/trigger_fl", headers=admin_headers)
        assert trigger_resp.status_code in (202, 503)
        trigger_payload = trigger_resp.get_json()
        if trigger_payload.get("status") == "accepted":
            assert trigger_payload.get("marketplace_contract")

        release_resp = client.post(
            "/marketplace/escrow/release",
            json={"contract_id": contract_id},
            headers=admin_headers,
        )
        assert release_resp.status_code == 200
        released = release_resp.get_json()
        assert released["payout_status"] == "released"

        dispute_resp = client.post(
            "/marketplace/disputes",
            json={
                "contract_id": contract_id,
                "reporter": "test-suite",
                "reason": "quality concerns",
            },
            headers=admin_headers,
        )
        assert dispute_resp.status_code == 201
        dispute = dispute_resp.get_json()
        assert dispute["status"] == "open"

        dispute_update_resp = client.patch(
            f"/marketplace/disputes/{dispute['dispute_id']}",
            json={"status": "under_review", "actor": "moderator"},
            headers=admin_headers,
        )
        assert dispute_update_resp.status_code == 200

        gov_resp = client.post(
            "/governance/actions",
            json={
                "action_type": "policy_review_requested",
                "actor": "test-suite",
                "source": "tests",
                "payload": {"contract_id": contract_id},
            },
            headers=admin_headers,
        )
        assert gov_resp.status_code == 201

        gov_list_resp = client.get("/governance/actions?limit=10")
        assert gov_list_resp.status_code == 200
        assert gov_list_resp.get_json()["count"] >= 1

        preview_resp = client.post(
            "/marketplace/policy/preview",
            json={
                "round_intent_id": intent_id,
                "max_offers": 3,
                "policy_overrides": {
                    "budget_total": 20.0,
                    "min_quality_score": 0.75,
                },
            },
            headers=admin_headers,
        )
        assert preview_resp.status_code == 200, preview_resp.get_data(as_text=True)
        preview_payload = preview_resp.get_json()
        assert preview_payload["preview"]["has_match"] is True
        assert preview_payload["preview"]["selected_offers"]

        proposal_resp = client.post(
            "/governance/proposals",
            json={
                "title": "Raise quality threshold",
                "proposal_type": "policy_update",
                "description": "Require stronger quality floor for matches",
                "created_by": "test-suite",
            },
            headers=admin_headers,
        )
        assert proposal_resp.status_code == 201, proposal_resp.get_data(as_text=True)
        proposal = proposal_resp.get_json()
        proposal_id = proposal["proposal_id"]

        vote_resp = client.post(
            f"/governance/proposals/{proposal_id}/vote",
            json={
                "voter": "node-1",
                "decision": "yes",
                "weight": 2.0,
                "reason": "supports policy hardening",
            },
            headers=admin_headers,
        )
        assert vote_resp.status_code == 200, vote_resp.get_data(as_text=True)
        voted = vote_resp.get_json()
        assert voted["tally"]["yes"] == 2.0
        assert voted["status"] == "approved"

        list_proposals_resp = client.get("/governance/proposals?limit=10")
        assert list_proposals_resp.status_code == 200
        assert list_proposals_resp.get_json()["count"] >= 1

        get_proposal_resp = client.get(f"/governance/proposals/{proposal_id}")
        assert get_proposal_resp.status_code == 200
        fetched_proposal = get_proposal_resp.get_json()
        assert fetched_proposal["proposal_id"] == proposal_id

        attest_resp = client.post(
            "/attestations/share",
            json={
                "participant_name": "node-growth-1",
                "node_name": "node-growth-1",
                "compute_type": "gpu",
                "compute_capacity": "2x A100",
                "capacity_score": 9.2,
                "attestation_status": "verified",
                "region": "us-east",
                "proof_digest": "sha256:demo-proof",
            },
        )
        assert attest_resp.status_code == 201, attest_resp.get_data(as_text=True)
        attestation = attest_resp.get_json()
        assert attestation["attestation_status"] == "verified"

        attest_feed_resp = client.get("/attestations/feed?limit=10")
        assert attest_feed_resp.status_code == 200
        assert attest_feed_resp.get_json()["count"] >= 1

        sorted_feed_resp = client.get("/attestations/feed?sort_by=reputation&limit=10")
        assert sorted_feed_resp.status_code == 200
        sorted_feed = sorted_feed_resp.get_json()
        assert sorted_feed["attestations"][0].get("reputation_score", 0) >= 0

        invite_request_resp = client.post(
            "/join/request_invite",
            json={
                "participant_name": "new-growth-node",
                "contact_email": "growth@example.org",
                "compute_type": "gpu",
                "region": "us-west",
                "preferred_language": "en",
                "motivation": "Contribute nightly compute",
            },
        )
        assert invite_request_resp.status_code == 201, invite_request_resp.get_data(
            as_text=True
        )
        invite_request = invite_request_resp.get_json()
        assert invite_request["status"] == "pending"

        auth_methods_resp = client.get("/admin/auth/methods")
        assert auth_methods_resp.status_code == 200
        assert "token_admin_enabled" in auth_methods_resp.get_json()

        list_requests_resp = client.get(
            "/join/invite_requests?status=pending&limit=5&offset=0&q=growth&sort_by=participant_name&sort_dir=asc",
            headers=admin_headers,
        )
        assert list_requests_resp.status_code == 200
        listed_requests = list_requests_resp.get_json()
        assert listed_requests["count"] >= 1
        assert listed_requests["total"] >= listed_requests["count"]
        assert listed_requests["sort_by"] == "participant_name"
        assert listed_requests["sort_dir"] == "asc"

        approve_request_resp = client.post(
            f"/join/invite_requests/{invite_request['request_id']}/approve",
            json={"max_uses": 1, "expires_in_hours": 24},
            headers=admin_headers,
        )
        assert approve_request_resp.status_code == 200
        approve_payload = approve_request_resp.get_json()
        invite_id = approve_payload["invite"]["invite_id"]

        list_invites_resp = client.get(
            "/join/invites?include_revoked=true&status=all&limit=5&offset=0&q=growth&sort_by=used&sort_dir=desc",
            headers=admin_headers,
        )
        assert list_invites_resp.status_code == 200
        invites_payload = list_invites_resp.get_json()
        assert invites_payload["count"] >= 1
        assert invites_payload["total"] >= invites_payload["count"]
        assert invites_payload["sort_by"] == "used"
        assert invites_payload["sort_dir"] == "desc"

        revoke_invite_resp = client.post(
            f"/join/invites/{invite_id}/revoke",
            headers=admin_headers,
        )
        assert revoke_invite_resp.status_code == 200
        assert revoke_invite_resp.get_json()["revoked"] is True

        invite_request_reject_resp = client.post(
            "/join/request_invite",
            json={
                "participant_name": "reject-growth-node",
                "contact_email": "reject@example.org",
                "compute_type": "cpu",
                "region": "eu-central",
                "preferred_language": "en",
                "motivation": "can provide fallback compute",
            },
        )
        assert invite_request_reject_resp.status_code == 201
        reject_request_id = invite_request_reject_resp.get_json()["request_id"]

        reject_request_resp = client.post(
            f"/join/invite_requests/{reject_request_id}/reject",
            json={"reason": "insufficient profile detail"},
            headers=admin_headers,
        )
        assert reject_request_resp.status_code == 200
        assert reject_request_resp.get_json()["status"] == "rejected"

        list_registrations_resp = client.get(
            "/join/registrations?status=all&limit=5&offset=0&q=node&sort_by=node_id&sort_dir=asc",
            headers=admin_headers,
        )
        assert list_registrations_resp.status_code == 200
        registrations_payload = list_registrations_resp.get_json()
        assert "total" in registrations_payload
        assert registrations_payload["sort_by"] == "node_id"
        assert registrations_payload["sort_dir"] == "asc"

        expansion_resp = client.get("/network/expansion_summary")
        assert expansion_resp.status_code == 200
        expansion = expansion_resp.get_json()
        assert expansion["total_attestations"] >= 1
        assert "active_nodes" in expansion
        assert "pending_invite_requests" in expansion

        summary_resp = client.get("/metrics_summary")
        assert summary_resp.status_code == 200
        summary = summary_resp.get_json()
        assert "marketplace" in summary
        assert summary["marketplace"]["contracts_total"] >= 1
        assert "governance" in summary
        assert summary["governance"]["proposals_total"] >= 1
        assert "network_expansion" in summary

        contracts_resp = client.get("/marketplace/contracts?payout_status=released")
        assert contracts_resp.status_code == 200
        contracts_payload = contracts_resp.get_json()
        assert contracts_payload["count"] >= 1

        print(json.dumps({"status": "ok", "contracts": contracts_payload["count"]}))
    return 0


if __name__ == "__main__":
    raise SystemExit(run())
