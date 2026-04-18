#!/usr/bin/env python3
"""Contract checks for swarm C2 HUD endpoints (security + basic performance constraints)."""

from __future__ import annotations

import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[3]
if str(REPO_ROOT) not in sys.path:
    sys.path.insert(0, str(REPO_ROOT))

import sovereignmap_production_backend_v2 as backend


def run() -> int:
    backend.JOIN_API_ADMIN_TOKEN = "swarm-contract-token"
    backend.ALLOW_INSECURE_DEV_ADMIN_TOKEN = False

    with backend.swarm_command_lock:
        backend.swarm_command_log.clear()
        backend.swarm_audit_log.clear()
        backend.swarm_command_nonce_cache.clear()
        backend.swarm_command_timestamps.clear()

    client = backend.app.test_client()

    status = client.get("/swarm/status")
    assert status.status_code == 200, status.get_data(as_text=True)
    status_payload = status.get_json()
    assert "nodes_active" in status_payload

    map_resp = client.get(
        "/swarm/map?limit=4000&layers=paths,risk,coverage,communications"
    )
    assert map_resp.status_code == 200, map_resp.get_data(as_text=True)
    map_payload = map_resp.get_json()
    assert map_payload.get("node_count", 0) <= backend.SWARM_MAX_MAP_NODES
    assert isinstance(map_payload.get("nodes", []), list)

    unauth = client.post("/swarm/command", json={"command": "hold_position"})
    assert unauth.status_code == 401, unauth.get_data(as_text=True)

    invalid = client.post(
        "/swarm/command",
        headers={"X-Join-Admin-Token": "swarm-contract-token"},
        json={
            "command": "reassign_role",
            "target_scope": "node",
            "target_ids": ["node-0001"],
            "parameters": {"role": "bad-role"},
        },
    )
    assert invalid.status_code == 400, invalid.get_data(as_text=True)

    accepted = client.post(
        "/swarm/command",
        headers={"X-Join-Admin-Token": "swarm-contract-token"},
        json={
            "command": "hold_position",
            "target_scope": "group",
            "target_ids": ["squad-a", "squad-b"],
            "client_nonce": "nonce-contract-1",
            "parameters": {},
        },
    )
    assert accepted.status_code == 202, accepted.get_data(as_text=True)
    accepted_payload = accepted.get_json()
    assert accepted_payload.get("status") == "accepted"
    assert accepted_payload.get("role") == "admin"
    assert accepted_payload.get("audit_signature")

    denied_by_role = client.post(
        "/swarm/command",
        headers={
            "X-Join-Admin-Token": "swarm-contract-token",
            "X-API-Role": "operator",
        },
        json={
            "command": "isolate_node",
            "target_scope": "node",
            "target_ids": ["node-0007"],
            "parameters": {},
        },
    )
    assert denied_by_role.status_code == 403, denied_by_role.get_data(as_text=True)
    denied_payload = denied_by_role.get_json()
    assert denied_payload.get("error") == "forbidden"

    duplicate = client.post(
        "/swarm/command",
        headers={"X-Join-Admin-Token": "swarm-contract-token"},
        json={
            "command": "hold_position",
            "target_scope": "group",
            "target_ids": ["squad-a", "squad-b"],
            "client_nonce": "nonce-contract-1",
            "parameters": {},
        },
    )
    assert duplicate.status_code == 200, duplicate.get_data(as_text=True)
    duplicate_payload = duplicate.get_json()
    assert duplicate_payload.get("duplicate") is True

    commands = client.get("/swarm/commands?limit=20")
    assert commands.status_code == 200, commands.get_data(as_text=True)
    command_payload = commands.get_json()
    assert command_payload.get("count", 0) >= 1

    audit_unauth = client.get("/swarm/audit/recent?limit=10")
    assert audit_unauth.status_code == 401, audit_unauth.get_data(as_text=True)

    audit = client.get(
        "/swarm/audit/recent?limit=10",
        headers={"X-Join-Admin-Token": "swarm-contract-token"},
    )
    assert audit.status_code == 200, audit.get_data(as_text=True)
    audit_payload = audit.get_json()
    assert audit_payload.get("count", 0) >= 1
    first_audit = audit_payload.get("audits", [])[0]
    assert first_audit.get("signature")

    print('{"status":"ok","swarm_c2_contracts":"validated"}')
    return 0


if __name__ == "__main__":
    raise SystemExit(run())
