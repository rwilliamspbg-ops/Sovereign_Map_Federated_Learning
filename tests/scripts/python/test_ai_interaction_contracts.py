#!/usr/bin/env python3
"""Contract checks for AI interaction summary/history/decision endpoints."""

from __future__ import annotations

import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[3]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

import sovereignmap_production_backend_v2 as backend

VALID_DECISIONS = {"approve", "edit", "reject", "undo"}


def _seed_history() -> None:
    with backend.interaction_review_lock:
        backend.interaction_review_log.clear()
        backend.interaction_review_log.append(
            {
                "review_id": "seed-review-1",
                "ts": 1710000000,
                "action_id": "start-training-10",
                "action_label": "Start 10-round training",
                "action_kind": "control_action",
                "model_route": "planner",
                "decision": "approve",
                "reason": "operator approved",
                "prompt": "start training for 10 rounds",
                "command": "start_training",
                "parameters": {"rounds": 10},
                "reversible": True,
            }
        )


def run() -> int:
    backend.JOIN_API_ADMIN_TOKEN = "ai-interaction-contract-token"
    backend.ALLOW_INSECURE_DEV_ADMIN_TOKEN = False

    client = backend.app.test_client()

    _seed_history()

    summary_resp = client.get("/ai/interaction/summary")
    assert summary_resp.status_code == 200, summary_resp.get_data(as_text=True)
    summary = summary_resp.get_json()
    assert isinstance(summary, dict)
    assert isinstance(summary.get("recommendations"), list)
    assert isinstance(summary.get("quick_actions"), list)
    assert isinstance(summary.get("context"), dict)

    history_resp = client.get("/ai/interaction/history?limit=10")
    assert history_resp.status_code == 200, history_resp.get_data(as_text=True)
    history = history_resp.get_json()
    assert isinstance(history, dict)
    assert isinstance(history.get("decisions"), list)
    assert history.get("count", 0) >= 1

    # Strict mode should reject missing admin auth.
    backend.AI_INTERACTION_DECISION_AUTH_MODE = "admin_required"
    unauth_resp = client.post(
        "/ai/interaction/decision",
        json={"decision": "approve", "action_id": "qa", "action_label": "QA"},
    )
    assert unauth_resp.status_code == 401, unauth_resp.get_data(as_text=True)
    unauth_payload = unauth_resp.get_json()
    assert unauth_payload.get("error") == "unauthorized"
    assert unauth_payload.get("auth_mode") == "admin_required"

    # Invalid decision should still fail validation under authenticated request.
    invalid_resp = client.post(
        "/ai/interaction/decision",
        headers={"X-Join-Admin-Token": "ai-interaction-contract-token"},
        json={"decision": "shipit", "action_id": "qa", "action_label": "QA"},
    )
    assert invalid_resp.status_code == 400, invalid_resp.get_data(as_text=True)
    assert invalid_resp.get_json().get("error") == "invalid_decision"

    # Happy-path decisions should be accepted and return a structured envelope.
    for decision in sorted(VALID_DECISIONS):
        ok_resp = client.post(
            "/ai/interaction/decision",
            headers={"X-Join-Admin-Token": "ai-interaction-contract-token"},
            json={
                "review_id": f"review-{decision}",
                "decision": decision,
                "action_id": f"action-{decision}",
                "action_label": f"Action {decision}",
                "action_kind": "assistant_query",
                "model_route": "planner",
                "reason": f"reason for {decision}",
                "prompt": "test prompt",
            },
        )
        assert ok_resp.status_code == 200, ok_resp.get_data(as_text=True)
        ok_payload = ok_resp.get_json()
        assert ok_payload.get("status") == "recorded"
        decision_payload = ok_payload.get("decision", {})
        assert decision_payload.get("decision") == decision
        assert decision_payload.get("action_id") == f"action-{decision}"
        assert "ts" in decision_payload

    # Local public mode should permit unauthenticated local test-client writes.
    backend.AI_INTERACTION_DECISION_AUTH_MODE = "public_local"
    public_resp = client.post(
        "/ai/interaction/decision",
        json={
            "decision": "approve",
            "action_id": "local-allow",
            "action_label": "Local allow",
        },
        environ_overrides={"REMOTE_ADDR": "127.0.0.1"},
    )
    assert public_resp.status_code == 200, public_resp.get_data(as_text=True)

    print('{"status":"ok","ai_interaction_contracts":"validated"}')
    return 0


if __name__ == "__main__":
    raise SystemExit(run())
