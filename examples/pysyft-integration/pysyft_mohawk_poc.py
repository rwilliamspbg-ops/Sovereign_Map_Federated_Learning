#!/usr/bin/env python3
"""PySyft + Sovereign Mohawk integration proof-of-concept.

Modes:
- mock: No Datasite login required. Simulates client updates and runs Mohawk aggregation.
- datasite: Uses PySyft Datasites (requires credentials + dataset assets).
"""

from __future__ import annotations

import argparse
import importlib
import json
import random
from collections import defaultdict
from typing import Any, Dict, Iterable, List, Optional, Sequence, Tuple

import numpy as np

try:
    sy = importlib.import_module("syft")
except Exception:  # pragma: no cover - optional dependency at runtime
    sy = None

try:
    from mohawk import MohawkNode
except Exception:  # pragma: no cover - optional dependency at runtime
    MohawkNode = None

ParamsDict = Dict[str, np.ndarray]


def _to_serializable_metrics(metrics: Dict[str, float]) -> Dict[str, float]:
    return {k: float(v) for k, v in metrics.items()}


def _flatten_params(params: ParamsDict, ordered_keys: Sequence[str]) -> np.ndarray:
    chunks: List[np.ndarray] = []
    for key in ordered_keys:
        chunks.append(np.ravel(params[key]).astype(np.float32))
    return np.concatenate(chunks)


def _rebuild_params(
    vector: np.ndarray,
    template: ParamsDict,
    ordered_keys: Sequence[str],
) -> ParamsDict:
    restored: ParamsDict = {}
    offset = 0
    for key in ordered_keys:
        shape = template[key].shape
        size = int(np.prod(shape))
        restored[key] = vector[offset : offset + size].reshape(shape)
        offset += size
    return restored


def _mock_local_train(model_params: Optional[ParamsDict], seed: int) -> Tuple[Dict[str, float], ParamsDict]:
    rng = np.random.default_rng(seed)
    if model_params is None:
        model_params = {
            "layer1.weight": rng.normal(0.0, 0.1, size=(4, 4)).astype(np.float32),
            "layer1.bias": rng.normal(0.0, 0.1, size=(4,)).astype(np.float32),
        }
    update = {k: v + rng.normal(0.0, 0.01, size=v.shape).astype(np.float32) for k, v in model_params.items()}
    metrics = {
        "accuracy": float(np.clip(0.80 + rng.normal(0.0, 0.02), 0.0, 1.0)),
        "loss": float(np.clip(0.35 + rng.normal(0.0, 0.03), 0.0, 10.0)),
    }
    return metrics, update


def _average_aggregate(updates: Sequence[ParamsDict]) -> ParamsDict:
    keys = list(updates[0].keys())
    avg: ParamsDict = {}
    for key in keys:
        stacked = np.stack([u[key] for u in updates], axis=0)
        avg[key] = stacked.mean(axis=0)
    return avg


def _extract_vector_from_mohawk_result(result: Dict[str, Any]) -> Optional[np.ndarray]:
    candidates = [
        result.get("aggregated"),
        result.get("gradient"),
        result.get("aggregated_gradient"),
        result.get("result"),
    ]
    for candidate in candidates:
        if candidate is None:
            continue
        return np.asarray(candidate, dtype=np.float32)
    return None


def mohawk_aggregate(
    updates: Sequence[ParamsDict],
    mohawk_node: Optional[Any],
    enable_compression: bool = True,
) -> Tuple[ParamsDict, Dict[str, Any]]:
    ordered_keys = list(updates[0].keys())

    if mohawk_node is None:
        return _average_aggregate(updates), {"engine": "numpy-average", "proof_verified": False}

    flattened = [_flatten_params(p, ordered_keys) for p in updates]
    payload = [
        {"node_id": f"ds_{idx}", "gradient": vec.tolist()}
        for idx, vec in enumerate(flattened)
    ]

    result = mohawk_node.aggregate(payload)
    proof = result.get("proof") if isinstance(result, dict) else None

    verified = False
    if proof is not None and hasattr(mohawk_node, "verify_proof"):
        verified = bool(mohawk_node.verify_proof(proof))

    vector = _extract_vector_from_mohawk_result(result if isinstance(result, dict) else {})
    if vector is None:
        vector = np.mean(np.stack(flattened, axis=0), axis=0)

    if enable_compression and hasattr(mohawk_node, "compress_gradients"):
        compressed = mohawk_node.compress_gradients(vector.tolist(), format="fp16")
        vector = np.asarray(compressed, dtype=np.float32)

    restored = _rebuild_params(vector, updates[0], ordered_keys)
    meta = {
        "engine": "mohawk",
        "proof_verified": verified,
        "proof_present": proof is not None,
    }
    return restored, meta


def init_mohawk(config_path: str) -> Optional[Any]:
    if MohawkNode is None:
        return None
    node = MohawkNode()
    if hasattr(node, "start"):
        node.start(config_path=config_path)
    return node


def run_mock(rounds: int, participants: int, config_path: str) -> Dict[int, List[Dict[str, float]]]:
    fl_metrics: Dict[int, List[Dict[str, float]]] = defaultdict(list)
    model_params: Optional[ParamsDict] = None
    mohawk_node = init_mohawk(config_path)

    print("Running mock mode (no Datasite login required)")
    for epoch in range(rounds):
        epoch_updates: List[ParamsDict] = []
        for pid in range(participants):
            metrics, params = _mock_local_train(model_params, seed=epoch * 100 + pid)
            fl_metrics[epoch].append(_to_serializable_metrics(metrics))
            epoch_updates.append(params)

        model_params, meta = mohawk_aggregate(epoch_updates, mohawk_node)
        mean_acc = float(np.mean([m["accuracy"] for m in fl_metrics[epoch]]))
        mean_loss = float(np.mean([m["loss"] for m in fl_metrics[epoch]]))
        print(
            f"epoch={epoch} participants={participants} engine={meta['engine']} "
            f"proof_verified={meta['proof_verified']} accuracy={mean_acc:.4f} loss={mean_loss:.4f}"
        )

    return fl_metrics


def _build_syft_train_function(data_asset: Any):
    if sy is None:
        raise RuntimeError("PySyft is not installed. Install with: pip install syft>=0.9.5")

    @sy.syft_function_single_use(data=data_asset, model_params=dict)
    def train_one_round(data, model_params=None):
        # Replace with local PyTorch/TF training logic running against private data.
        import numpy as _np

        if model_params is None:
            model_params = {
                "layer1.weight": _np.zeros((4, 4), dtype=_np.float32),
                "layer1.bias": _np.zeros((4,), dtype=_np.float32),
            }

        updated = {k: v + _np.random.normal(0.0, 0.01, size=v.shape).astype(_np.float32) for k, v in model_params.items()}
        metrics = {"accuracy": float(_np.random.uniform(0.80, 0.95)), "loss": float(_np.random.uniform(0.10, 0.35))}
        return metrics, updated

    return train_one_round


def run_datasite(
    rounds: int,
    config_path: str,
    server_url: str,
    email: str,
    password: str,
    dataset_name: str,
    asset_name: str,
) -> Dict[int, List[Dict[str, float]]]:
    if sy is None:
        raise RuntimeError("PySyft is not installed. Install with: pip install syft>=0.9.5")

    client = sy.login(url=server_url, email=email, password=password)
    datasites = [client.datasite]

    model_params: Optional[ParamsDict] = None
    fl_metrics: Dict[int, List[Dict[str, float]]] = defaultdict(list)
    mohawk_node = init_mohawk(config_path)

    print(f"Running datasite mode against {server_url}")
    for epoch in range(rounds):
        epoch_updates: List[ParamsDict] = []
        for ds in datasites:
            data_asset = ds.datasets[dataset_name].assets[asset_name]
            train_fn = _build_syft_train_function(data_asset)

            # Some PySyft deployments require code submission/requesting before execution.
            # This call path works where direct execution is enabled.
            result = ds.code(train_fn)(data=data_asset, model_params=model_params).get()
            metrics, params = result
            params = {k: np.asarray(v, dtype=np.float32) for k, v in params.items()}

            fl_metrics[epoch].append(_to_serializable_metrics(metrics))
            epoch_updates.append(params)

        model_params, meta = mohawk_aggregate(epoch_updates, mohawk_node)
        mean_acc = float(np.mean([m["accuracy"] for m in fl_metrics[epoch]]))
        mean_loss = float(np.mean([m["loss"] for m in fl_metrics[epoch]]))
        print(
            f"epoch={epoch} datasites={len(datasites)} engine={meta['engine']} "
            f"proof_verified={meta['proof_verified']} accuracy={mean_acc:.4f} loss={mean_loss:.4f}"
        )

    return fl_metrics


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="PySyft + Sovereign Mohawk FL PoC")
    parser.add_argument("--mode", choices=["mock", "datasite"], default="mock")
    parser.add_argument("--rounds", type=int, default=5)
    parser.add_argument("--participants", type=int, default=5)
    parser.add_argument("--config", default="capabilities.json")

    parser.add_argument("--server-url", default="http://localhost:8080")
    parser.add_argument("--email", default="")
    parser.add_argument("--password", default="")
    parser.add_argument("--dataset", default="YourDataset")
    parser.add_argument("--asset", default="YourData")
    return parser.parse_args()


def main() -> None:
    args = parse_args()

    random.seed(42)
    np.random.seed(42)

    if args.mode == "mock":
        metrics = run_mock(args.rounds, args.participants, args.config)
    else:
        required = [args.email, args.password]
        if any(not v for v in required):
            raise SystemExit("datasite mode requires --email and --password")
        metrics = run_datasite(
            rounds=args.rounds,
            config_path=args.config,
            server_url=args.server_url,
            email=args.email,
            password=args.password,
            dataset_name=args.dataset,
            asset_name=args.asset,
        )

    print("\\nFinal metrics summary:")
    print(json.dumps(metrics, indent=2))


if __name__ == "__main__":
    main()
