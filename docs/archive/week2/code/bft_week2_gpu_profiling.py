#!/usr/bin/env python
"""
WEEK 2 TEST 5: GPU PROFILING
Identify acceleration opportunities | Operation timing | Bandwidth analysis
"""

import numpy as np
import time

# Try to import GPU libraries
try:
    import cupy as cp

    GPU_AVAILABLE = True
except ImportError:
    GPU_AVAILABLE = False
    print("[INFO] CuPy not available, CPU-only profiling")

# ============================================================================
# OPERATION PROFILER
# ============================================================================


class OperationProfiler:
    """Profile individual operations for GPU acceleration potential"""

    @staticmethod
    def profile_aggregation(data_size, batch_size, iterations=100):
        """Profile trimmed mean aggregation"""
        updates = np.random.randn(batch_size, data_size).astype(np.float32)

        # CPU timing
        start = time.time()
        for _ in range(iterations):
            norms = np.linalg.norm(updates, axis=1)
            indices = np.argsort(norms)
            trim_count = max(1, int(batch_size * 0.1))
            kept = indices[trim_count:-trim_count]
            result = np.mean(updates[kept], axis=0)
        cpu_time = time.time() - start

        # GPU timing (if available)
        gpu_time = None
        if GPU_AVAILABLE:
            updates_gpu = cp.array(updates)
            start = time.time()
            for _ in range(iterations):
                norms = cp.linalg.norm(updates_gpu, axis=1)
                indices = cp.argsort(norms)
                trim_count = max(1, int(batch_size * 0.1))
                kept = indices[trim_count:-trim_count]
                result = cp.mean(updates_gpu[kept], axis=0)
                cp.cuda.Stream.null.synchronize()
            gpu_time = time.time() - start

        return cpu_time, gpu_time

    @staticmethod
    def profile_gradients(data_size, batch_size, iterations=100):
        """Profile gradient computation (matrix multiply)"""
        X = np.random.randn(batch_size, data_size).astype(np.float32)
        W = np.random.randn(data_size, 10).astype(np.float32)

        # CPU timing
        start = time.time()
        for _ in range(iterations):
            result = X @ W
        cpu_time = time.time() - start

        # GPU timing
        gpu_time = None
        if GPU_AVAILABLE:
            X_gpu = cp.array(X)
            W_gpu = cp.array(W)
            start = time.time()
            for _ in range(iterations):
                result = X_gpu @ W_gpu
                cp.cuda.Stream.null.synchronize()
            gpu_time = time.time() - start

        return cpu_time, gpu_time

    @staticmethod
    def profile_distance_calc(data_size, batch_size, iterations=100):
        """Profile Byzantine detection (distance calculations)"""
        updates = np.random.randn(batch_size, data_size).astype(np.float32)

        # CPU timing
        start = time.time()
        for _ in range(iterations):
            distances = np.zeros((batch_size, batch_size))
            for i in range(batch_size):
                for j in range(i + 1, batch_size):
                    distances[i, j] = np.linalg.norm(updates[i] - updates[j])
        cpu_time = time.time() - start

        # GPU timing (simplified batch operation)
        gpu_time = None
        if GPU_AVAILABLE:
            updates_gpu = cp.array(updates)
            start = time.time()
            for _ in range(iterations):
                # Batch distance: ||A - B||^2 = ||A||^2 + ||B||^2 - 2*A*B.T
                norms_sq = cp.sum(updates_gpu**2, axis=1, keepdims=True)
                distances = norms_sq + norms_sq.T - 2 * (updates_gpu @ updates_gpu.T)
                cp.cuda.Stream.null.synchronize()
            gpu_time = time.time() - start

        return cpu_time, gpu_time


# ============================================================================
# GPU PROFILING TEST
# ============================================================================


class GPUProfilingTest:
    def __init__(self):
        self.results = []

    def run_all(self):
        """Profile all operations"""
        profiler = OperationProfiler()

        operations = ["aggregation", "gradients", "distance"]
        data_sizes = [50, 1000, 10000]
        batch_sizes = [75, 200, 500]

        total_start = time.time()

        for op in operations:
            for ds in data_sizes:
                for bs in batch_sizes:
                    if op == "aggregation":
                        cpu_t, gpu_t = profiler.profile_aggregation(
                            ds, bs, iterations=100
                        )
                    elif op == "gradients":
                        cpu_t, gpu_t = profiler.profile_gradients(
                            ds, bs, iterations=100
                        )
                    else:
                        # Skip very large distance calc for speed
                        if ds > 1000 or bs > 200:
                            continue
                        cpu_t, gpu_t = profiler.profile_distance_calc(
                            ds, bs, iterations=10
                        )

                    speedup = (cpu_t / gpu_t) if gpu_t and gpu_t > 0 else None

                    self.results.append(
                        {
                            "op": op,
                            "data_size": ds,
                            "batch_size": bs,
                            "cpu_time": cpu_t,
                            "gpu_time": gpu_t,
                            "speedup": speedup,
                        }
                    )

        total_elapsed = time.time() - total_start
        return total_elapsed

    def print_results(self):
        """Print profiling results"""
        print("\n" + "=" * 120)
        print("  GPU PROFILING RESULTS")
        print("=" * 120 + "\n")

        for op in ["aggregation", "gradients", "distance"]:
            print(f"  Operation: {op.upper()}")
            print(f"  " + "-" * 116)
            print(
                f"  {'Data Size':<12} {'Batch Size':<12} {'CPU Time (ms)':<16} {'GPU Time (ms)':<16} {'Speedup':<12}"
            )
            print(f"  {'-'*70}")

            op_results = [r for r in self.results if r["op"] == op]

            for res in op_results:
                cpu_ms = res["cpu_time"] * 1000
                gpu_ms = res["gpu_time"] * 1000 if res["gpu_time"] else None
                speedup = res["speedup"]

                if gpu_ms:
                    speedup_str = f"{speedup:.1f}x"
                else:
                    speedup_str = "N/A (no GPU)"

                print(
                    f"  {res['data_size']:<12} {res['batch_size']:<12} {cpu_ms:<16.3f} "
                    f"{gpu_ms if gpu_ms else 'N/A':<16} {speedup_str:<12}"
                )

            print()


# ============================================================================
# MAIN
# ============================================================================

if __name__ == "__main__":
    print("\n" + "=" * 120)
    print("  WEEK 2 TEST 5: GPU PROFILING")
    print("  Aggregation | Gradients | Distance Calculation")
    print("=" * 120 + "\n")

    print(f"  GPU Available: {GPU_AVAILABLE}\n")

    test = GPUProfilingTest()
    elapsed = test.run_all()
    test.print_results()

    # Summary
    print("\n" + "=" * 120)
    print("  PROFILING SUMMARY")
    print("=" * 120 + "\n")

    speedups = [r["speedup"] for r in test.results if r["speedup"]]
    if speedups:
        avg_speedup = np.mean(speedups)
        max_speedup = max(speedups)
        min_speedup = min(speedups)

        print(f"  Average Speedup: {avg_speedup:.1f}x")
        print(f"  Max Speedup: {max_speedup:.1f}x")
        print(f"  Min Speedup: {min_speedup:.1f}x")
        print(
            f"\n  Recommendation: GPU acceleration {'RECOMMENDED' if avg_speedup > 2 else 'OPTIONAL'}"
        )
    else:
        print(f"  GPU not available - CPU-only profiling")
        print(f"  Recommendation: Evaluate GPU deployment for acceleration")

    print(f"\n  Total profiling time: {elapsed:.1f}s")
    print("\n" + "=" * 120 + "\n")
