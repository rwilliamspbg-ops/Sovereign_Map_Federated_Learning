#!/usr/bin/env python3
"""Short production-like replay for FL aggregation path metric behavior.

This harness runs mixed client participation rounds, applies the same auto-path
selection policy used by the aggregator, and exports Prometheus metric text so
operators can confirm expected panel trend behavior.
"""

from __future__ import annotations

from collections import Counter
from dataclasses import dataclass
import os
import time

import numpy as np
from prometheus_client import Counter as PromCounter, CollectorRegistry, generate_latest, start_http_server


@dataclass(frozen=True)
class ReplayConfig:
    mode: str
    vectorize_min_clients: int
    vectorize_max_peak_bytes: int
    replay_pattern: tuple[int, ...]
    live_port: int
    round_sleep_seconds: float


def mnist_layer_shapes() -> list[tuple[int, ...]]:
    return [
        (32, 1, 3, 3),
        (32,),
        (64, 32, 3, 3),
        (64,),
        (128, 9216),
        (128,),
        (10, 128),
        (10,),
    ]


def generate_client_weights(clients: int, seed: int) -> list[list[np.ndarray]]:
    rng = np.random.default_rng(seed)
    shapes = mnist_layer_shapes()
    return [[rng.standard_normal(shape, dtype=np.float32) for shape in shapes] for _ in range(clients)]


def estimate_vectorized_peak_bytes(weights_list: list[list[np.ndarray]]) -> int:
    client_count = len(weights_list)
    peak = 0
    for layer in weights_list[0]:
        layer_bytes = int(np.prod(layer.shape)) * layer.dtype.itemsize * client_count
        peak = max(peak, layer_bytes)
    return peak


def should_use_vectorized(weights_list: list[list[np.ndarray]], config: ReplayConfig) -> bool:
    if config.mode == "loop":
        return False
    if config.mode == "vectorized":
        return True
    if len(weights_list) < config.vectorize_min_clients:
        return False
    peak = estimate_vectorized_peak_bytes(weights_list)
    return peak <= config.vectorize_max_peak_bytes


def aggregate_loop(weights_list: list[list[np.ndarray]]) -> list[np.ndarray]:
    n = len(weights_list)
    out: list[np.ndarray] = []
    for layer_idx in range(len(weights_list[0])):
        layer_sum = np.zeros_like(weights_list[0][layer_idx])
        for weights in weights_list:
            layer_sum += weights[layer_idx]
        out.append(layer_sum / n)
    return out


def aggregate_vectorized(weights_list: list[list[np.ndarray]]) -> list[np.ndarray]:
    num_layers = len(weights_list[0])
    return [
        np.stack([client[layer_idx] for client in weights_list], axis=0).mean(axis=0)
        for layer_idx in range(num_layers)
    ]


def parse_config() -> ReplayConfig:
    mode = os.getenv("FL_AGGREGATION_MODE", "auto").strip().lower()
    if mode not in {"auto", "loop", "vectorized"}:
        mode = "auto"

    min_clients = int(os.getenv("FL_AGGREGATION_VECTORIZE_MIN_CLIENTS", "50"))
    max_peak = int(os.getenv("FL_AGGREGATION_VECTORIZE_MAX_PEAK_BYTES", str(2 * 1024 * 1024 * 1024)))

    pattern_raw = os.getenv("FL_REPLAY_PATTERN", "24,28,26,120,130,140,125,24,22,128,136,132")
    pattern = tuple(int(p.strip()) for p in pattern_raw.split(",") if p.strip())
    if not pattern:
        pattern = (24, 120, 24, 120)

    live_port = int(os.getenv("FL_REPLAY_LIVE_PORT", "0"))
    if live_port < 0:
        live_port = 0

    round_sleep_seconds = float(os.getenv("FL_REPLAY_SLEEP_SECONDS", "0"))
    if round_sleep_seconds < 0:
        round_sleep_seconds = 0.0

    return ReplayConfig(
        mode=mode,
        vectorize_min_clients=max(1, min_clients),
        vectorize_max_peak_bytes=max(1, max_peak),
        replay_pattern=pattern,
        live_port=live_port,
        round_sleep_seconds=round_sleep_seconds,
    )


def main() -> None:
    config = parse_config()
    registry = CollectorRegistry()
    aggregation_path_counter = PromCounter(
        "fl_aggregation_path_total",
        "Total aggregation path selections by implementation.",
        ["impl"],
        registry=registry,
    )

    window: list[str] = []
    totals = Counter()

    if config.live_port > 0:
        start_http_server(config.live_port, registry=registry)
        print(f"live_metrics_port= {config.live_port}")

    for idx, clients in enumerate(config.replay_pattern, start=1):
        weights = generate_client_weights(clients, seed=1000 + idx)
        use_vectorized = should_use_vectorized(weights, config)
        impl = "vectorized" if use_vectorized else "loop"
        if use_vectorized:
            aggregate_vectorized(weights)
        else:
            aggregate_loop(weights)
        aggregation_path_counter.labels(impl=impl).inc()
        totals[impl] += 1
        window.append(impl)
        if config.round_sleep_seconds > 0:
            time.sleep(config.round_sleep_seconds)

    last5 = Counter(window[-5:])
    trend = "vectorized_up" if last5["vectorized"] > last5["loop"] else "loop_up_or_flat"

    print("replay_rounds=", len(config.replay_pattern))
    print("mode=", config.mode)
    print("vectorize_min_clients=", config.vectorize_min_clients)
    print("vectorize_max_peak_bytes=", config.vectorize_max_peak_bytes)
    print("path_totals=", dict(totals))
    print("last5_rounds=", dict(last5))
    print("panel_trend_expectation=", trend)
    print("\n# metrics_exposition")
    print(generate_latest(registry).decode("utf-8").strip())


if __name__ == "__main__":
    main()
