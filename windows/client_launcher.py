#!/usr/bin/env python3
"""
Windows launcher for Sovereign Map FL client.

Features:
- Optional self-serve join bootstrap (invite/register)
- Saves cert bundle and registration JSON locally
- Prints NPU/GPU/CPU acceleration diagnostics
- Connects to Flower aggregator using existing client runtime
"""

from __future__ import annotations

import argparse
import json
import os
import pathlib
import sys
import urllib.error
import urllib.request
from typing import Any, Dict, Optional


def _post_json(url: str, payload: Dict[str, Any], headers: Optional[Dict[str, str]] = None) -> Dict[str, Any]:
    data = json.dumps(payload).encode("utf-8")
    req_headers = {"Content-Type": "application/json"}
    if headers:
        req_headers.update(headers)

    req = urllib.request.Request(url, data=data, headers=req_headers, method="POST")
    with urllib.request.urlopen(req, timeout=20) as response:
        return json.loads(response.read().decode("utf-8"))


def _bootstrap_join(
    backend_url: str,
    participant_name: str,
    invite_code: Optional[str],
    admin_token: Optional[str],
    out_dir: pathlib.Path,
) -> Dict[str, Any]:
    out_dir.mkdir(parents=True, exist_ok=True)
    cert_dir = out_dir / "certs"
    cert_dir.mkdir(parents=True, exist_ok=True)

    code = invite_code
    if not code:
        if not admin_token:
            raise RuntimeError("invite code is required unless admin token is provided")
        invite = _post_json(
            f"{backend_url}/join/invite",
            {
                "participant_name": participant_name,
                "max_uses": 1,
                "expires_in_hours": 24,
            },
            headers={"X-Join-Admin-Token": admin_token},
        )
        code = invite.get("invite_code")
        if not code:
            raise RuntimeError(f"join invite failed: {invite}")

    registration = _post_json(
        f"{backend_url}/join/register",
        {"invite_code": code, "participant_name": participant_name},
    )

    certificates = registration.get("certificates", {})
    (cert_dir / "node-cert.pem").write_text(certificates.get("node_cert_pem", ""), encoding="utf-8")
    (cert_dir / "node-key.pem").write_text(certificates.get("node_key_pem", ""), encoding="utf-8")
    (cert_dir / "ca-cert.pem").write_text(certificates.get("ca_cert_pem", ""), encoding="utf-8")
    (out_dir / "join-registration.json").write_text(
        json.dumps(registration, indent=2), encoding="utf-8"
    )

    return registration


def _apply_llm_policy_env(policy: Dict[str, Any]):
    os.environ["LLM_ADAPTER_MODEL_FAMILY"] = str(policy.get("model_family", "llama-3.1"))
    os.environ["LLM_ADAPTER_MODEL_VERSION"] = str(policy.get("model_version", "8b-instruct"))
    os.environ["LLM_ADAPTER_TOKENIZER_HASH"] = str(policy.get("tokenizer_hash", "local-dev-tokenizer-v1"))

    ranks = policy.get("allowed_adapter_ranks") or [16]
    os.environ["LLM_ADAPTER_RANK"] = str(ranks[0])

    modules = policy.get("required_target_modules") or ["q_proj", "v_proj"]
    os.environ["LLM_ADAPTER_TARGET_MODULES"] = ",".join(str(m) for m in modules)


def _acceleration_report(torch_mod) -> Dict[str, Any]:
    gpu_backend = "rocm" if getattr(getattr(torch_mod, "version", None), "hip", None) else "cuda"
    report: Dict[str, Any] = {
        "torch_version": torch_mod.__version__,
        "cuda_available": bool(torch_mod.cuda.is_available()),
        "gpu_backend": gpu_backend,
        "cuda_device_count": 0,
        "cuda_devices": [],
        "npu_available": False,
        "npu_device": None,
        "xpu_available": False,
        "xpu_device_count": 0,
        "xpu_devices": [],
        "mps_available": False,
        "selected_device": "cpu",
    }

    if report["cuda_available"]:
        count = torch_mod.cuda.device_count()
        report["cuda_device_count"] = count
        for idx in range(count):
            report["cuda_devices"].append(torch_mod.cuda.get_device_name(idx))
        if count > 0:
            report["selected_device"] = "cuda:0"

    if hasattr(torch_mod, "xpu"):
        try:
            if torch_mod.xpu.is_available():
                report["xpu_available"] = True
                count = torch_mod.xpu.device_count()
                report["xpu_device_count"] = count
                for idx in range(count):
                    if hasattr(torch_mod.xpu, "get_device_name"):
                        report["xpu_devices"].append(torch_mod.xpu.get_device_name(idx))
                    else:
                        report["xpu_devices"].append(f"xpu:{idx}")
                if count > 0:
                    report["selected_device"] = "xpu:0"
        except Exception:
            pass

    if hasattr(torch_mod, "npu"):
        try:
            if torch_mod.npu.is_available():
                report["npu_available"] = True
                report["npu_device"] = "npu:0"
                report["selected_device"] = "npu:0"
        except Exception:
            pass

    mps_backend = getattr(getattr(torch_mod, "backends", None), "mps", None)
    if mps_backend is not None:
        try:
            if mps_backend.is_available():
                report["mps_available"] = True
                if report["selected_device"] == "cpu":
                    report["selected_device"] = "mps"
        except Exception:
            pass

    # Probe selected device with a tiny tensor op
    selected = report["selected_device"]
    try:
        tensor = torch_mod.randn((4, 4), device=selected)
        _ = tensor @ tensor
        report["probe_ok"] = True
    except Exception as exc:
        report["probe_ok"] = False
        report["probe_error"] = str(exc)
        report["selected_device"] = "cpu"

    return report


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Sovereign Map Windows FL Client Launcher")
    parser.add_argument("--aggregator", default="localhost:8080", help="Aggregator host:port")
    parser.add_argument("--node-id", type=int, default=1, help="Node ID for this participant")
    parser.add_argument("--byzantine", action="store_true", help="Run as Byzantine test node")

    parser.add_argument("--backend-url", default="http://localhost:8000", help="Backend URL for join bootstrap")
    parser.add_argument("--participant-name", default="windows-client", help="Participant name for registration")
    parser.add_argument("--invite-code", default="", help="Existing invite code")
    parser.add_argument("--admin-token", default="", help="Join admin token (for local testing only)")
    parser.add_argument("--skip-bootstrap", action="store_true", help="Skip join bootstrap and connect directly")

    parser.add_argument(
        "--output-dir",
        default=str(pathlib.Path.home() / "SovereignMapClient"),
        help="Directory to store certs and registration",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()

    print("Sovereign Map Windows Client Launcher")
    print("-------------------------------------")

    registration = None
    if not args.skip_bootstrap:
        try:
            registration = _bootstrap_join(
                backend_url=args.backend_url.rstrip("/"),
                participant_name=args.participant_name,
                invite_code=args.invite_code or None,
                admin_token=args.admin_token or None,
                out_dir=pathlib.Path(args.output_dir) / args.participant_name,
            )
            agg = registration.get("aggregator", {})
            host = agg.get("host")
            port = agg.get("port")
            if host and port:
                args.aggregator = f"{host}:{port}"

            policy = registration.get("llm_policy", {})
            if isinstance(policy, dict):
                _apply_llm_policy_env(policy)

            reg_node = registration.get("registration", {}).get("node_id")
            if isinstance(reg_node, int):
                args.node_id = reg_node

            print(f"Join bootstrap OK: node_id={args.node_id}, aggregator={args.aggregator}")
        except (RuntimeError, urllib.error.URLError, TimeoutError) as exc:
            print(f"Join bootstrap failed: {exc}")
            return 1

    try:
        import flwr as fl
        import torch
        from src.client import SovereignClient
    except Exception as exc:
        print(f"Missing runtime dependency: {exc}")
        print("Install dependencies or build with windows/build_windows_client_exe.ps1")
        return 1

    report = _acceleration_report(torch)
    print("Acceleration diagnostics:")
    print(json.dumps(report, indent=2))

    if report.get("selected_device") == "cpu":
        print("Warning: No NPU/GPU acceleration detected; running on CPU.")

    print(f"Connecting to aggregator at {args.aggregator} as node {args.node_id}...")

    client = SovereignClient(
        node_id=args.node_id,
        byzantine=args.byzantine,
        server_address=args.aggregator,
    )

    try:
        fl.client.start_client(
            server_address=args.aggregator,
            client=client.to_client(),
            grpc_max_message_length=1024 * 1024 * 1024,
        )
    except Exception as exc:
        print(f"Failed to connect to aggregator: {exc}")
        return 1

    return 0


if __name__ == "__main__":
    sys.exit(main())
