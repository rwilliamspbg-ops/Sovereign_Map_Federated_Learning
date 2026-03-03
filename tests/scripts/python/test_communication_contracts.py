#!/usr/bin/env python3
"""Dependency-free communication contract smoke test.

Validates source-level contracts for:
- Frontend <-> Backend API compatibility
- Node client <-> Aggregator port alignment
"""

from __future__ import annotations

from pathlib import Path


ROOT = Path(__file__).resolve().parents[3]
BACKEND_FILE = ROOT / "sovereignmap_production_backend_v2.py"
FRONTEND_APP_FILE = ROOT / "frontend" / "src" / "App.jsx"
NODE_CLIENT_FILE = ROOT / "src" / "client.py"


def assert_frontend_contract(app_source: str) -> None:
    required_fragments = [
        "http://localhost:8000",
        "fetch(`${API_BASE}/hud_data`)",
        "fetch(`${API_BASE}/health`)",
        "fetch(`${API_BASE}/metrics_summary`)",
        "fetch(`${API_BASE}/founders`)",
        "fetch(`${API_BASE}/trigger_fl`",
        "fetch(`${API_BASE}/create_enclave`",
    ]

    for fragment in required_fragments:
        assert fragment in app_source, f"Missing frontend contract fragment: {fragment}"


def assert_backend_contract(backend_source: str) -> None:
    required_routes = [
        '@app.route("/health", methods=["GET"])',
        '@app.route("/hud_data", methods=["GET"])',
        '@app.route("/founders", methods=["GET"])',
        '@app.route("/metrics_summary", methods=["GET"])',
        '@app.route("/trigger_fl", methods=["POST"])',
        '@app.route("/create_enclave", methods=["POST"])',
        '@app.route("/status", methods=["GET"])',
    ]

    for route in required_routes:
        assert route in backend_source, f"Missing backend route contract: {route}"

    required_status_contract = [
        '"flower_server_port": 8080',
        '"metrics_api_port": 8000',
    ]
    for fragment in required_status_contract:
        assert (
            fragment in backend_source
        ), f"Missing backend status contract: {fragment}"

    required_health_contract = [
        '"tpm_verified": True',
        '"enclave_status": enclave_status',
    ]
    for fragment in required_health_contract:
        assert (
            fragment in backend_source
        ), f"Missing backend health contract: {fragment}"


def main() -> int:
    backend_source = BACKEND_FILE.read_text(encoding="utf-8")
    assert_backend_contract(backend_source)

    app_source = FRONTEND_APP_FILE.read_text(encoding="utf-8")
    assert_frontend_contract(app_source)

    node_source = NODE_CLIENT_FILE.read_text(encoding="utf-8")
    assert 'server_address: str = "localhost:8080"' in node_source

    print(
        "✅ Communication contracts validated: frontend/backend/node alignment is healthy"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
