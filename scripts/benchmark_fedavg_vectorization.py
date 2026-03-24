#!/usr/bin/env python3
"""Benchmark loop-based vs vectorized FedAvg aggregation.

This script compares the original Python nested-loop aggregation to the
vectorized NumPy implementation now used in src/aggregator.py.
"""

from __future__ import annotations

import argparse
import time
from dataclasses import dataclass
from statistics import median

import numpy as np


@dataclass
class BenchResult:
    clients: int
    loop_ms: float
    vectorized_ms: float
    speedup: float


def mnist_layer_shapes() -> list[tuple[int, ...]]:
    """Return MNISTNet layer shapes that match src/aggregator.py model."""
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


def per_client_bytes() -> int:
    total = 0
    for shape in mnist_layer_shapes():
        total += int(np.prod(shape)) * np.dtype(np.float32).itemsize
    return total


def generate_client_weights(clients: int, seed: int) -> list[list[np.ndarray]]:
    rng = np.random.default_rng(seed)
    shapes = mnist_layer_shapes()
    weights_list: list[list[np.ndarray]] = []
    for _ in range(clients):
        layers = [rng.standard_normal(shape, dtype=np.float32) for shape in shapes]
        weights_list.append(layers)
    return weights_list


def aggregate_loop(weights_list: list[list[np.ndarray]]) -> list[np.ndarray]:
    n = len(weights_list)
    aggregated: list[np.ndarray] = []
    for layer_idx in range(len(weights_list[0])):
        layer_sum = np.zeros_like(weights_list[0][layer_idx])
        for weights in weights_list:
            layer_sum += weights[layer_idx]
        aggregated.append(layer_sum / n)
    return aggregated


def aggregate_vectorized(weights_list: list[list[np.ndarray]]) -> list[np.ndarray]:
    num_layers = len(weights_list[0])
    return [
        np.stack([client[layer_idx] for client in weights_list], axis=0).mean(axis=0)
        for layer_idx in range(num_layers)
    ]


def time_ms(fn, weights_list: list[list[np.ndarray]], runs: int) -> float:
    samples = []
    for _ in range(runs):
        start = time.perf_counter()
        fn(weights_list)
        samples.append((time.perf_counter() - start) * 1000.0)
    return median(samples)


def validate_equal(loop_out: list[np.ndarray], vec_out: list[np.ndarray]) -> None:
    if len(loop_out) != len(vec_out):
        raise RuntimeError("aggregation output layer count mismatch")
    for idx, (a, b) in enumerate(zip(loop_out, vec_out)):
        if not np.allclose(a, b, rtol=1e-5, atol=1e-6):
            raise RuntimeError(f"layer {idx} mismatch between loop and vectorized output")


def run_bench(client_counts: list[int], runs: int, seed: int) -> list[BenchResult]:
    results: list[BenchResult] = []
    budget_bytes = 1_200_000_000  # ~1.2 GB safety cap for test containers.
    client_size = per_client_bytes()
    for clients in client_counts:
        estimated = clients * client_size
        if estimated > budget_bytes:
            print(
                f"Skipping clients={clients}: estimated weights footprint {estimated / (1024**3):.2f} GiB exceeds safety budget"
            )
            continue

        weights = generate_client_weights(clients, seed + clients)

        # Validate correctness once before timing.
        validate_equal(aggregate_loop(weights), aggregate_vectorized(weights))

        loop_ms = time_ms(aggregate_loop, weights, runs)
        vec_ms = time_ms(aggregate_vectorized, weights, runs)
        speedup = loop_ms / vec_ms if vec_ms > 0 else float("inf")

        results.append(
            BenchResult(
                clients=clients,
                loop_ms=loop_ms,
                vectorized_ms=vec_ms,
                speedup=speedup,
            )
        )
    return results


def print_results(results: list[BenchResult]) -> None:
    print("\nFedAvg Aggregation Benchmark (median per scenario)")
    print("| Clients | Loop (ms) | Vectorized (ms) | Speedup |")
    print("|---:|---:|---:|---:|")
    for row in results:
        print(
            f"| {row.clients} | {row.loop_ms:.3f} | {row.vectorized_ms:.3f} | {row.speedup:.2f}x |"
        )


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Benchmark FedAvg vectorization")
    parser.add_argument(
        "--clients",
        type=int,
        nargs="+",
        default=[100, 500, 1000],
        help="Client counts to benchmark",
    )
    parser.add_argument(
        "--runs",
        type=int,
        default=5,
        help="Timing runs per scenario (median reported)",
    )
    parser.add_argument("--seed", type=int, default=42, help="RNG seed")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    results = run_bench(args.clients, args.runs, args.seed)
    print_results(results)


if __name__ == "__main__":
    main()
