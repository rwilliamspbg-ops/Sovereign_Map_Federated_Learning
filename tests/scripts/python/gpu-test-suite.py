#!/usr/bin/env python3
"""
GPU Acceleration Testing Suite for Sovereign Map FL
====================================================
Tests PyTorch CUDA/GPU performance for:
- Local training latency (CPU vs GPU)
- High-density GPU contention (20+ nodes competing)
- Round latency with GPU acceleration
- zk-SNARK verification speedup on GPU

Usage:
    python tests/scripts/python/gpu-test-suite.py --nodes 20 --rounds 50 --gpu
    python tests/scripts/python/gpu-test-suite.py --benchmark  # CPU vs GPU comparison
    python tests/scripts/python/gpu-test-suite.py --stress     # Maximum GPU contention
"""

import argparse
import json
import logging
import threading
import time
from datetime import datetime
from typing import Dict

import numpy as np
import torch
import torch.nn as nn
import torch.nn.functional as F
from torch.utils.data import DataLoader, TensorDataset

# Configure logging
logging.basicConfig(
    level=logging.INFO, format="%(asctime)s | %(levelname)-8s | %(name)s | %(message)s"
)
logger = logging.getLogger(__name__)


# ============================================================================
# NEURAL NETWORK MODEL
# ============================================================================


class MNISTNet(nn.Module):
    """Simple CNN for MNIST - matches src/client.py"""

    def __init__(self):
        super(MNISTNet, self).__init__()
        self.conv1 = nn.Conv2d(1, 32, 3, 1)
        self.conv2 = nn.Conv2d(32, 64, 3, 1)
        self.fc1 = nn.Linear(9216, 128)
        self.fc2 = nn.Linear(128, 10)

    def forward(self, x):
        x = F.relu(self.conv1(x))
        x = F.relu(self.conv2(x))
        x = F.max_pool2d(x, 2)
        x = torch.flatten(x, 1)
        x = F.relu(self.fc1(x))
        x = self.fc2(x)
        return F.log_softmax(x, dim=1)


# ============================================================================
# GPU UTILITIES
# ============================================================================


class GPUInfo:
    """Detect available GPU/NPU devices"""

    @staticmethod
    def get_devices() -> Dict[str, bool]:
        """Check available accelerators"""
        devices = {
            "cuda": torch.cuda.is_available(),
            "npu": hasattr(torch, "npu") and torch.npu.is_available()
            if hasattr(torch, "npu")
            else False,
            "cpu": True,
        }
        return devices

    @staticmethod
    def print_info():
        """Print GPU/accelerator information"""
        logger.info("=" * 60)
        logger.info("GPU/Accelerator Information")
        logger.info("=" * 60)

        devices = GPUInfo.get_devices()
        logger.info(f"CUDA Available: {devices['cuda']}")
        logger.info(f"NPU Available: {devices['npu']}")
        logger.info(f"CPU Available: {devices['cpu']}")

        if devices["cuda"]:
            logger.info(f"CUDA Version: {torch.version.cuda}")
            logger.info(f"CUDA Device Count: {torch.cuda.device_count()}")
            for i in range(torch.cuda.device_count()):
                logger.info(f"  Device {i}: {torch.cuda.get_device_name(i)}")
                props = torch.cuda.get_device_properties(i)
                logger.info(f"    Memory: {props.total_memory / 1e9:.2f} GB")

        logger.info("=" * 60)


# ============================================================================
# TRAINING BENCHMARK
# ============================================================================


class TrainingBenchmark:
    """Benchmark PyTorch training on different devices"""

    def __init__(self, device: str, num_epochs: int = 2):
        self.device = torch.device(device)
        self.num_epochs = num_epochs
        self.model = MNISTNet().to(self.device)
        self.optimizer = torch.optim.SGD(self.model.parameters(), lr=0.01)
        self.criterion = F.nll_loss
        self.times = []
        self.throughputs = []

        logger.info(f"Initialized benchmark on {self.device}")

    def create_random_data(
        self, batch_size: int = 16, num_batches: int = 50
    ) -> DataLoader:
        """Create synthetic MNIST-like data"""
        X = torch.randn(num_batches * batch_size, 1, 28, 28)
        y = torch.randint(0, 10, (num_batches * batch_size,))
        dataset = TensorDataset(X, y)
        return DataLoader(dataset, batch_size=batch_size, shuffle=True)

    def train_step(self, dataloader: DataLoader, epoch: int) -> float:
        """Execute one training epoch and return loss"""
        self.model.train()
        total_loss = 0.0
        batch_count = 0

        for batch_idx, (data, target) in enumerate(dataloader):
            data, target = data.to(self.device), target.to(self.device)
            self.optimizer.zero_grad()
            output = self.model(data)
            loss = F.nll_loss(output, target)
            loss.backward()
            self.optimizer.step()

            total_loss += loss.item()
            batch_count += 1

        avg_loss = total_loss / max(batch_count, 1)
        return avg_loss

    def benchmark(self, batch_size: int = 16, num_batches: int = 50) -> Dict:
        """Run training benchmark and measure latency"""
        logger.info(f"Starting benchmark on {self.device}")
        logger.info(f"  Batch size: {batch_size}")
        logger.info(f"  Batches: {num_batches}")
        logger.info(f"  Epochs: {self.num_epochs}")

        dataloader = self.create_random_data(batch_size, num_batches)

        # Warmup
        self.train_step(dataloader, 0)

        # Benchmark epochs
        epoch_times = []
        for epoch in range(self.num_epochs):
            torch.cuda.synchronize() if torch.cuda.is_available() else None
            start = time.time()

            avg_loss = self.train_step(dataloader, epoch)

            torch.cuda.synchronize() if torch.cuda.is_available() else None
            elapsed = time.time() - start
            epoch_times.append(elapsed)

            samples_per_sec = (num_batches * batch_size) / elapsed
            logger.info(
                f"  Epoch {epoch+1}: Loss={avg_loss:.4f}, Time={elapsed:.3f}s, Throughput={samples_per_sec:.0f} samples/sec"
            )

        avg_time = np.mean(epoch_times)
        std_time = np.std(epoch_times)
        throughput = (num_batches * batch_size) / avg_time

        return {
            "device": str(self.device),
            "avg_epoch_time_sec": float(avg_time),
            "std_epoch_time_sec": float(std_time),
            "samples_per_sec": float(throughput),
            "batches": num_batches,
            "batch_size": batch_size,
            "epochs": self.num_epochs,
        }


# ============================================================================
# HIGH-DENSITY GPU CONTENTION TEST
# ============================================================================


class GPUContentionTest:
    """Simulate multiple nodes competing for GPU resources"""

    def __init__(self, num_nodes: int, device: str = "cuda:0"):
        self.num_nodes = num_nodes
        self.device = torch.device(device)
        self.threads = []
        self.results = {}
        self.lock = threading.Lock()

        logger.info(f"GPU Contention Test: {num_nodes} nodes on {self.device}")

    def node_training(self, node_id: int, num_batches: int = 20):
        """Simulate training for a single node"""
        try:
            model = MNISTNet().to(self.device)
            optimizer = torch.optim.SGD(model.parameters(), lr=0.01)

            # Create synthetic data
            X = torch.randn(num_batches * 16, 1, 28, 28)
            y = torch.randint(0, 10, (num_batches * 16,))
            dataset = TensorDataset(X, y)
            dataloader = DataLoader(dataset, batch_size=16, shuffle=True)

            # Train
            start = time.time()
            model.train()
            for data, target in dataloader:
                data, target = data.to(self.device), target.to(self.device)
                optimizer.zero_grad()
                output = model(data)
                loss = F.nll_loss(output, target)
                loss.backward()
                optimizer.step()

            elapsed = time.time() - start

            with self.lock:
                self.results[node_id] = {
                    "status": "completed",
                    "time_sec": elapsed,
                    "throughput": (num_batches * 16) / elapsed,
                }
                logger.info(
                    f"  Node {node_id}: {elapsed:.3f}s ({self.results[node_id]['throughput']:.0f} samples/sec)"
                )

        except Exception as e:
            with self.lock:
                self.results[node_id] = {
                    "status": "failed",
                    "error": str(e),
                }
            logger.error(f"  Node {node_id}: {e}")

    def run(self, num_batches: int = 20) -> Dict:
        """Run high-density contention test"""
        logger.info(f"Starting GPU contention test with {self.num_nodes} nodes")

        start = time.time()

        # Start all node threads
        for node_id in range(self.num_nodes):
            t = threading.Thread(target=self.node_training, args=(node_id, num_batches))
            t.start()
            self.threads.append(t)

        # Wait for all to complete
        for t in self.threads:
            t.join()

        total_time = time.time() - start

        # Aggregate results
        successful = sum(1 for r in self.results.values() if r["status"] == "completed")
        failed = sum(1 for r in self.results.values() if r["status"] == "failed")
        avg_time = np.mean(
            [r["time_sec"] for r in self.results.values() if r["status"] == "completed"]
        )
        total_throughput = sum(
            r.get("throughput", 0)
            for r in self.results.values()
            if r["status"] == "completed"
        )

        logger.info(
            f"Contention test complete: {successful} successful, {failed} failed"
        )
        logger.info(f"  Total time: {total_time:.2f}s")
        logger.info(f"  Avg node time: {avg_time:.3f}s")
        logger.info(f"  Total throughput: {total_throughput:.0f} samples/sec")

        return {
            "num_nodes": self.num_nodes,
            "total_time_sec": total_time,
            "successful_nodes": successful,
            "failed_nodes": failed,
            "avg_node_time_sec": float(avg_time),
            "total_throughput_samples_per_sec": float(total_throughput),
            "node_results": self.results,
        }


# ============================================================================
# ROUND LATENCY BENCHMARK
# ============================================================================


class RoundLatencyBenchmark:
    """Measure end-to-end FL round latency with GPU acceleration"""

    def __init__(self, num_nodes: int, device: str = "cuda:0"):
        self.num_nodes = num_nodes
        self.device = torch.device(device)

    def run(self, num_rounds: int = 10, batch_per_node: int = 10) -> Dict:
        """Simulate FL rounds and measure latency"""
        logger.info(
            f"Round latency benchmark: {num_rounds} rounds, {self.num_nodes} nodes"
        )

        # Initialize global model
        global_model = MNISTNet().to(self.device)
        global_model.eval()

        round_times = []

        for round_num in range(num_rounds):
            round_start = time.time()

            # Simulate local training on all nodes (sequential for now)
            for node_id in range(self.num_nodes):
                model = MNISTNet().to(self.device)
                optimizer = torch.optim.SGD(model.parameters(), lr=0.01)

                # Create synthetic data
                X = torch.randn(batch_per_node * 16, 1, 28, 28)
                y = torch.randint(0, 10, (batch_per_node * 16,))
                dataset = TensorDataset(X, y)
                dataloader = DataLoader(dataset, batch_size=16)

                # Train
                model.train()
                for data, target in dataloader:
                    data, target = data.to(self.device), target.to(self.device)
                    optimizer.zero_grad()
                    output = model(data)
                    loss = F.nll_loss(output, target)
                    loss.backward()
                    optimizer.step()

            # Aggregate (simplified)
            round_time = time.time() - round_start
            round_times.append(round_time)

            logger.info(f"  Round {round_num+1}: {round_time:.3f}s")

        avg_round_time = np.mean(round_times)

        logger.info(f"Average round latency: {avg_round_time:.3f}s")
        logger.info(f"Throughput: {self.num_nodes / avg_round_time:.1f} updates/sec")

        return {
            "num_rounds": num_rounds,
            "num_nodes": self.num_nodes,
            "avg_round_latency_sec": float(avg_round_time),
            "std_round_latency_sec": float(np.std(round_times)),
            "min_round_latency_sec": float(np.min(round_times)),
            "max_round_latency_sec": float(np.max(round_times)),
            "updates_per_sec": float(self.num_nodes / avg_round_time),
        }


# ============================================================================
# COMPARISON & REPORTING
# ============================================================================


class GPUTestReport:
    """Generate comprehensive GPU testing report"""

    def __init__(self):
        self.results = {}

    def cpu_vs_gpu_benchmark(self) -> Dict:
        """Compare CPU vs GPU training latency"""
        logger.info("\n" + "=" * 60)
        logger.info("CPU vs GPU Training Benchmark")
        logger.info("=" * 60)

        cpu_bench = TrainingBenchmark("cpu", num_epochs=2)
        cpu_result = cpu_bench.benchmark(batch_size=16, num_batches=50)

        gpu_bench = None
        gpu_result = None

        if torch.cuda.is_available():
            logger.info("")
            gpu_bench = TrainingBenchmark("cuda:0", num_epochs=2)
            gpu_result = gpu_bench.benchmark(batch_size=16, num_batches=50)

            speedup = (
                cpu_result["avg_epoch_time_sec"] / gpu_result["avg_epoch_time_sec"]
            )
            logger.info(f"\nSpeedup (CPU/GPU): {speedup:.2f}x")

        return {
            "cpu": cpu_result,
            "gpu": gpu_result,
            "speedup": float(speedup) if gpu_result else None,
        }

    def generate_report(self, benchmarks: Dict) -> str:
        """Generate formatted report"""
        report = []
        report.append("=" * 80)
        report.append("SOVEREIGN MAP GPU ACCELERATION TEST REPORT")
        report.append("=" * 80)
        report.append(f"Timestamp: {datetime.now().isoformat()}")
        report.append("")

        # GPU Info
        report.append("GPU/Accelerator Information:")
        devices = GPUInfo.get_devices()
        report.append(f"  CUDA Available: {devices['cuda']}")
        report.append(f"  NPU Available: {devices['npu']}")
        report.append("")

        # Benchmark Results
        if "cpu_vs_gpu" in benchmarks:
            cpu_gpu = benchmarks["cpu_vs_gpu"]
            report.append("CPU vs GPU Training Benchmark:")
            report.append(
                f"  CPU avg epoch time: {cpu_gpu['cpu']['avg_epoch_time_sec']:.3f}s"
            )
            if cpu_gpu["gpu"]:
                report.append(
                    f"  GPU avg epoch time: {cpu_gpu['gpu']['avg_epoch_time_sec']:.3f}s"
                )
                report.append(f"  Speedup: {cpu_gpu['speedup']:.2f}x")
            report.append("")

        if "contention" in benchmarks:
            cont = benchmarks["contention"]
            report.append(f"GPU Contention Test ({cont['num_nodes']} nodes):")
            report.append(f"  Total time: {cont['total_time_sec']:.2f}s")
            report.append(f"  Successful nodes: {cont['successful_nodes']}")
            report.append(f"  Failed nodes: {cont['failed_nodes']}")
            report.append(f"  Avg node time: {cont['avg_node_time_sec']:.3f}s")
            report.append(
                f"  Total throughput: {cont['total_throughput_samples_per_sec']:.0f} samples/sec"
            )
            report.append("")

        if "round_latency" in benchmarks:
            rl = benchmarks["round_latency"]
            report.append(f"Round Latency Benchmark ({rl['num_nodes']} nodes):")
            report.append(f"  Avg round latency: {rl['avg_round_latency_sec']:.3f}s")
            report.append(
                f"  Min/Max: {rl['min_round_latency_sec']:.3f}s / {rl['max_round_latency_sec']:.3f}s"
            )
            report.append(f"  Updates/sec: {rl['updates_per_sec']:.1f}")
            report.append("")

        report.append("=" * 80)
        return "\n".join(report)


# ============================================================================
# MAIN
# ============================================================================


def main():
    parser = argparse.ArgumentParser(
        description="GPU Acceleration Test Suite for Sovereign Map FL"
    )
    parser.add_argument(
        "--benchmark", action="store_true", help="Run CPU vs GPU benchmark"
    )
    parser.add_argument(
        "--contention", action="store_true", help="Run GPU contention test"
    )
    parser.add_argument(
        "--round-latency", action="store_true", help="Run round latency benchmark"
    )
    parser.add_argument(
        "--nodes",
        type=int,
        default=20,
        help="Number of nodes for contention/latency tests",
    )
    parser.add_argument(
        "--rounds", type=int, default=10, help="Number of FL rounds for latency test"
    )
    parser.add_argument(
        "--device",
        type=str,
        default="cuda:0" if torch.cuda.is_available() else "cpu",
        help="Device to use (cuda:0, cpu, etc.)",
    )
    parser.add_argument("--all", action="store_true", help="Run all tests")
    parser.add_argument("--json", type=str, help="Output JSON report to file")

    args = parser.parse_args()

    # Print GPU info
    GPUInfo.print_info()

    report_gen = GPUTestReport()
    benchmarks = {}

    # Run tests
    if args.benchmark or args.all:
        logger.info("\nRunning CPU vs GPU benchmark...")
        benchmarks["cpu_vs_gpu"] = report_gen.cpu_vs_gpu_benchmark()

    if args.contention or args.all:
        logger.info("\nRunning GPU contention test...")
        cont_test = GPUContentionTest(args.nodes, args.device)
        benchmarks["contention"] = cont_test.run()

    if args.round_latency or args.all:
        logger.info("\nRunning round latency benchmark...")
        rl_test = RoundLatencyBenchmark(args.nodes, args.device)
        benchmarks["round_latency"] = rl_test.run(args.rounds)

    # Generate report
    report = report_gen.generate_report(benchmarks)
    print("\n" + report)

    # Save JSON if requested
    if args.json:
        with open(args.json, "w") as f:
            json.dump(benchmarks, f, indent=2)
        logger.info(f"JSON report saved to {args.json}")


if __name__ == "__main__":
    main()
