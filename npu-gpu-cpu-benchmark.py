#!/usr/bin/env python3
"""
NPU/GPU/CPU Acceleration Testing Suite for Sovereign Map FL
============================================================
Tests PyTorch performance across:
- CPU (baseline)
- GPU (CUDA - when available)
- NPU (Huawei Ascend - when available)
- Device detection and fallback hierarchy

NPU Support:
- AMD Ryzen AI 7 350: Has integrated AI accelerator
- PyTorch NPU binding via torch.npu module
- Falls back to GPU (CUDA) if NPU unavailable
- Falls back to CPU if GPU unavailable

Usage:
    python npu-gpu-cpu-benchmark.py --all
    python npu-gpu-cpu-benchmark.py --npu --nodes 20
    python npu-gpu-cpu-benchmark.py --compare-devices
"""

import argparse
import json
import logging
import os
import sys
import threading
import time
from datetime import datetime
from typing import Dict, List, Tuple, Optional

import numpy as np
import torch
import torch.nn as nn
import torch.nn.functional as F
from torch.utils.data import DataLoader, TensorDataset

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s | %(levelname)-8s | %(name)s | %(message)s'
)
logger = logging.getLogger(__name__)


# ============================================================================
# DEVICE DETECTION & MANAGEMENT
# ============================================================================

class DeviceManager:
    """Detect and manage NPU, GPU, and CPU devices"""
    
    @staticmethod
    def detect_devices() -> Dict[str, Dict]:
        """Detect all available accelerators"""
        devices = {}
        
        # CPU - always available
        devices['cpu'] = {
            'available': True,
            'name': 'CPU',
            'memory': 'System RAM',
            'priority': 3,
            'device_count': 1,
        }
        
        # CUDA (NVIDIA GPUs)
        devices['cuda'] = {
            'available': torch.cuda.is_available(),
            'name': 'NVIDIA GPU (CUDA)',
            'priority': 2,
            'device_count': torch.cuda.device_count() if torch.cuda.is_available() else 0,
        }
        if devices['cuda']['available']:
            try:
                devices['cuda']['memory'] = f"{torch.cuda.get_device_properties(0).total_memory / 1e9:.1f} GB"
                devices['cuda']['device_name'] = torch.cuda.get_device_name(0)
            except:
                devices['cuda']['memory'] = 'Unknown'
        
        # NPU (Huawei Ascend / AMD AI accelerator)
        npu_available = hasattr(torch, 'npu') and torch.npu.is_available() if hasattr(torch, 'npu') else False
        devices['npu'] = {
            'available': npu_available,
            'name': 'NPU (Huawei Ascend / AMD AI)',
            'priority': 1,  # Highest priority
            'device_count': torch.npu.device_count() if npu_available else 0,
        }
        if npu_available:
            try:
                devices['npu']['memory'] = f"{torch.npu.get_device_properties(0).total_memory / 1e9:.1f} GB"
                devices['npu']['device_name'] = torch.npu.get_device_name(0)
            except:
                devices['npu']['memory'] = 'Unknown'
        
        return devices
    
    @staticmethod
    def get_best_device() -> Tuple[torch.device, str]:
        """Get best available device following priority: NPU > GPU > CPU"""
        devices = DeviceManager.detect_devices()
        
        # Priority order: NPU > GPU (CUDA) > CPU
        for device_type in ['npu', 'cuda', 'cpu']:
            if devices[device_type]['available']:
                if device_type == 'npu':
                    return torch.device('npu:0'), 'npu'
                elif device_type == 'cuda':
                    return torch.device('cuda:0'), 'cuda'
                else:
                    return torch.device('cpu'), 'cpu'
        
        return torch.device('cpu'), 'cpu'
    
    @staticmethod
    def print_device_info():
        """Print detailed device information"""
        logger.info("="*70)
        logger.info("Device Detection & Availability")
        logger.info("="*70)
        
        devices = DeviceManager.detect_devices()
        
        for device_type, info in devices.items():
            status = "AVAILABLE" if info['available'] else "NOT AVAILABLE"
            logger.info(f"\n{info['name']:<30} [{status}]")
            if info['available']:
                logger.info(f"  Priority: {info['priority']}")
                logger.info(f"  Count: {info['device_count']}")
                if 'device_name' in info:
                    logger.info(f"  Device: {info['device_name']}")
                if 'memory' in info:
                    logger.info(f"  Memory: {info['memory']}")
        
        best_device, device_type = DeviceManager.get_best_device()
        logger.info(f"\nBest Device Selected: {best_device} ({device_type.upper()})")
        logger.info("="*70)


# ============================================================================
# NEURAL NETWORK MODEL
# ============================================================================

class MNISTNet(nn.Module):
    """Simple CNN for MNIST"""
    
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
# DEVICE-SPECIFIC TRAINING BENCHMARK
# ============================================================================

class DeviceTrainingBenchmark:
    """Benchmark training on specific device"""
    
    def __init__(self, device: str, num_epochs: int = 2):
        """
        Args:
            device: 'cpu', 'cuda', or 'npu'
            num_epochs: Number of training epochs
        """
        self.device_type = device
        
        # Map device type to torch.device
        if device == 'npu':
            self.device = torch.device('npu:0')
        elif device == 'cuda':
            self.device = torch.device('cuda:0')
        else:
            self.device = torch.device('cpu')
        
        self.num_epochs = num_epochs
        self.model = MNISTNet().to(self.device)
        self.optimizer = torch.optim.SGD(self.model.parameters(), lr=0.01)
        
        logger.info(f"Initialized benchmark on {self.device} ({device.upper()})")
    
    def create_random_data(self, batch_size: int = 16, num_batches: int = 50) -> DataLoader:
        """Create synthetic MNIST-like data"""
        X = torch.randn(num_batches * batch_size, 1, 28, 28)
        y = torch.randint(0, 10, (num_batches * batch_size,))
        dataset = TensorDataset(X, y)
        return DataLoader(dataset, batch_size=batch_size, shuffle=True)
    
    def train_step(self, dataloader: DataLoader, epoch: int) -> float:
        """Execute one training epoch"""
        self.model.train()
        total_loss = 0.0
        batch_count = 0
        
        for data, target in dataloader:
            try:
                data, target = data.to(self.device), target.to(self.device)
            except Exception as e:
                logger.warning(f"Device transfer error: {e}, falling back to CPU")
                data, target = data.to(torch.device('cpu')), target.to(torch.device('cpu'))
            
            self.optimizer.zero_grad()
            output = self.model(data)
            loss = F.nll_loss(output, target)
            loss.backward()
            self.optimizer.step()
            
            total_loss += loss.item()
            batch_count += 1
        
        return total_loss / max(batch_count, 1)
    
    def benchmark(self, batch_size: int = 16, num_batches: int = 50) -> Dict:
        """Run benchmark on device"""
        logger.info(f"Starting {self.device_type.upper()} benchmark")
        logger.info(f"  Batch size: {batch_size}, Batches: {num_batches}, Epochs: {self.num_epochs}")
        
        dataloader = self.create_random_data(batch_size, num_batches)
        
        # Warmup
        self.train_step(dataloader, 0)
        
        # Actual benchmark
        epoch_times = []
        for epoch in range(self.num_epochs):
            try:
                # Synchronize before timing (device-specific)
                if self.device_type == 'npu' and hasattr(torch.npu, 'synchronize'):
                    torch.npu.synchronize()
                elif self.device_type == 'cuda':
                    torch.cuda.synchronize()
            except:
                pass
            
            start = time.time()
            avg_loss = self.train_step(dataloader, epoch)
            
            try:
                if self.device_type == 'npu' and hasattr(torch.npu, 'synchronize'):
                    torch.npu.synchronize()
                elif self.device_type == 'cuda':
                    torch.cuda.synchronize()
            except:
                pass
            
            elapsed = time.time() - start
            epoch_times.append(elapsed)
            
            samples_per_sec = (num_batches * batch_size) / elapsed
            logger.info(f"  Epoch {epoch+1}: Loss={avg_loss:.4f}, Time={elapsed:.3f}s, Throughput={samples_per_sec:.0f} samples/sec")
        
        avg_time = np.mean(epoch_times)
        std_time = np.std(epoch_times)
        throughput = (num_batches * batch_size) / avg_time
        
        return {
            "device": self.device_type,
            "avg_epoch_time_sec": float(avg_time),
            "std_epoch_time_sec": float(std_time),
            "samples_per_sec": float(throughput),
            "batches": num_batches,
            "batch_size": batch_size,
            "epochs": self.num_epochs,
        }


# ============================================================================
# MULTI-DEVICE COMPARISON
# ============================================================================

class DeviceComparison:
    """Compare performance across CPU, GPU, and NPU"""
    
    def __init__(self):
        self.results = {}
    
    def benchmark_all_devices(self) -> Dict:
        """Benchmark all available devices"""
        logger.info("\n" + "="*70)
        logger.info("MULTI-DEVICE BENCHMARKING")
        logger.info("="*70)
        
        devices_to_test = []
        available_devices = DeviceManager.detect_devices()
        
        # Determine which devices to test
        if available_devices['npu']['available']:
            devices_to_test.append('npu')
        if available_devices['cuda']['available']:
            devices_to_test.append('cuda')
        devices_to_test.append('cpu')  # Always test CPU
        
        results = {}
        
        for device_type in devices_to_test:
            logger.info(f"\nBenchmarking {device_type.upper()}...")
            try:
                benchmark = DeviceTrainingBenchmark(device_type, num_epochs=2)
                device_result = benchmark.benchmark(batch_size=16, num_batches=50)
                results[device_type] = device_result
                logger.info(f"{device_type.upper()} completed: {device_result['avg_epoch_time_sec']:.3f}s/epoch")
            except Exception as e:
                logger.error(f"Failed to benchmark {device_type}: {e}")
                results[device_type] = {'error': str(e)}
        
        return results
    
    def generate_comparison_report(self, results: Dict) -> str:
        """Generate comparison report"""
        report = []
        report.append("\n" + "="*70)
        report.append("NPU / GPU / CPU PERFORMANCE COMPARISON")
        report.append("="*70)
        report.append(f"Timestamp: {datetime.now().isoformat()}\n")
        
        # Extract valid results
        valid_results = {k: v for k, v in results.items() if 'avg_epoch_time_sec' in v}
        
        if not valid_results:
            report.append("No valid benchmark results available.")
            return "\n".join(report)
        
        # Sort by performance
        sorted_results = sorted(valid_results.items(), 
                              key=lambda x: x[1]['avg_epoch_time_sec'])
        
        # Performance table
        report.append(f"{'Device':<15} {'Epoch Time (s)':<20} {'Throughput':<20} {'Speedup vs CPU':<20}")
        report.append("-" * 75)
        
        cpu_time = valid_results.get('cpu', {}).get('avg_epoch_time_sec', 1.0)
        
        for device_type, result in sorted_results:
            epoch_time = result['avg_epoch_time_sec']
            throughput = result['samples_per_sec']
            speedup = cpu_time / epoch_time if epoch_time > 0 else 0
            
            report.append(f"{device_type.upper():<15} {epoch_time:<20.3f} {throughput:<20.0f} {speedup:<20.2f}x")
        
        # Rankings
        report.append("\n" + "="*70)
        report.append("RANKINGS")
        report.append("-"*70)
        
        for rank, (device_type, result) in enumerate(sorted_results, 1):
            status = "BEST" if rank == 1 else "GOOD" if rank == 2 else "BASELINE"
            report.append(f"{rank}. {device_type.upper():<15} {result['avg_epoch_time_sec']:.3f}s/epoch [{status}]")
        
        return "\n".join(report)


# ============================================================================
# NPU-SPECIFIC CONTENTION TEST
# ============================================================================

class NPUContentionTest:
    """Test NPU performance with multiple threads (simulating nodes)"""
    
    def __init__(self, num_nodes: int, device: str = 'npu'):
        self.num_nodes = num_nodes
        self.device = device
        self.threads = []
        self.results = {}
        self.lock = threading.Lock()
    
    def node_training(self, node_id: int, num_batches: int = 20):
        """Simulate training for a single node on NPU"""
        try:
            if self.device == 'npu':
                device_obj = torch.device('npu:0')
            elif self.device == 'cuda':
                device_obj = torch.device('cuda:0')
            else:
                device_obj = torch.device('cpu')
            
            model = MNISTNet().to(device_obj)
            optimizer = torch.optim.SGD(model.parameters(), lr=0.01)
            
            X = torch.randn(num_batches * 16, 1, 28, 28)
            y = torch.randint(0, 10, (num_batches * 16,))
            dataset = TensorDataset(X, y)
            dataloader = DataLoader(dataset, batch_size=16, shuffle=True)
            
            start = time.time()
            model.train()
            for data, target in dataloader:
                data, target = data.to(device_obj), target.to(device_obj)
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
                logger.info(f"  Node {node_id}: {elapsed:.3f}s ({self.results[node_id]['throughput']:.0f} samples/sec)")
        
        except Exception as e:
            with self.lock:
                self.results[node_id] = {
                    "status": "failed",
                    "error": str(e),
                }
            logger.error(f"  Node {node_id}: {e}")
    
    def run(self, num_batches: int = 20) -> Dict:
        """Run contention test"""
        logger.info(f"Starting {self.device.upper()} contention test with {self.num_nodes} nodes")
        
        start = time.time()
        
        # Start all threads
        for node_id in range(self.num_nodes):
            t = threading.Thread(target=self.node_training, args=(node_id, num_batches))
            t.start()
            self.threads.append(t)
        
        # Wait for completion
        for t in self.threads:
            t.join()
        
        total_time = time.time() - start
        
        # Aggregate results
        successful = sum(1 for r in self.results.values() if r["status"] == "completed")
        failed = sum(1 for r in self.results.values() if r["status"] == "failed")
        avg_time = np.mean([r["time_sec"] for r in self.results.values() if r["status"] == "completed"])
        total_throughput = sum(r.get("throughput", 0) for r in self.results.values() if r["status"] == "completed")
        
        logger.info(f"Test complete: {successful} successful, {failed} failed in {total_time:.2f}s")
        
        return {
            "device": self.device,
            "num_nodes": self.num_nodes,
            "total_time_sec": total_time,
            "successful_nodes": successful,
            "failed_nodes": failed,
            "avg_node_time_sec": float(avg_time),
            "total_throughput_samples_per_sec": float(total_throughput),
        }


# ============================================================================
# MAIN
# ============================================================================

def main():
    parser = argparse.ArgumentParser(description="NPU/GPU/CPU Acceleration Benchmark Suite")
    parser.add_argument('--all', action='store_true', help='Run all benchmarks')
    parser.add_argument('--compare-devices', action='store_true', help='Compare all available devices')
    parser.add_argument('--npu', action='store_true', help='Benchmark NPU only')
    parser.add_argument('--gpu', action='store_true', help='Benchmark GPU only')
    parser.add_argument('--cpu', action='store_true', help='Benchmark CPU only')
    parser.add_argument('--contention', action='store_true', help='Run contention test')
    parser.add_argument('--device', type=str, choices=['npu', 'cuda', 'cpu'], 
                       help='Specific device to benchmark')
    parser.add_argument('--nodes', type=int, default=20, help='Number of nodes for contention test')
    parser.add_argument('--json', type=str, help='Output JSON report')
    
    args = parser.parse_args()
    
    # Print device info
    DeviceManager.print_device_info()
    
    results = {}
    
    # Determine what to run
    if args.all or args.compare_devices:
        logger.info("\nRunning device comparison...")
        comparison = DeviceComparison()
        comparison_results = comparison.benchmark_all_devices()
        results['comparison'] = comparison_results
        print(comparison.generate_comparison_report(comparison_results))
    
    if args.contention:
        device = args.device or 'npu' if DeviceManager.detect_devices()['npu']['available'] else 'cuda' if DeviceManager.detect_devices()['cuda']['available'] else 'cpu'
        logger.info(f"\nRunning contention test on {device.upper()}...")
        contention = NPUContentionTest(args.nodes, device)
        results['contention'] = contention.run()
    
    # Save JSON if requested
    if args.json and results:
        with open(args.json, 'w') as f:
            json.dump(results, f, indent=2)
        logger.info(f"Results saved to {args.json}")


if __name__ == "__main__":
    main()
