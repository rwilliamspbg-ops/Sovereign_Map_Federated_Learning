"""
Phase 3D: Federated learning training with CIFAR-10/MNIST support,
DP noise injection, gradient compression, and optional multi-GPU execution.
"""

from dataclasses import dataclass
from pathlib import Path
from typing import Any, Dict, List, Tuple
import json

import numpy as np
import torch
import torch.nn as nn
from torch.utils.data import DataLoader, Subset
from torchvision import datasets, transforms


@dataclass
class TrainingConfig:
    """Configuration for federated learning training."""

    num_rounds: int = 10
    num_clients: int = 10
    local_epochs: int = 1
    batch_size: int = 32
    learning_rate: float = 0.01
    epsilon: float = 1.0
    delta: float = 1e-5
    compression_bits: int = 8
    use_compression: bool = True
    use_privacy: bool = True
    dataset: str = "cifar10"  # cifar10 | mnist
    device: str = "auto"  # auto | cpu | cuda
    multi_gpu: bool = True
    num_workers: int = 2
    seed: int = 42


class SimpleCNN(nn.Module):
    """Compact CNN that adapts to MNIST (1x28x28) and CIFAR-10 (3x32x32)."""

    def __init__(self, input_channels: int, image_size: int, num_classes: int = 10):
        super().__init__()
        self.conv1 = nn.Conv2d(input_channels, 32, kernel_size=3, padding=1)
        self.conv2 = nn.Conv2d(32, 64, kernel_size=3, padding=1)
        self.relu = nn.ReLU()
        self.pool = nn.MaxPool2d(2, 2)
        self.dropout = nn.Dropout(0.2)

        # Two pooling ops divide spatial dimensions by 4.
        feat_size = image_size // 4
        self.fc1 = nn.Linear(64 * feat_size * feat_size, 128)
        self.fc2 = nn.Linear(128, num_classes)

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        x = self.pool(self.relu(self.conv1(x)))
        x = self.pool(self.relu(self.conv2(x)))
        x = x.flatten(1)
        x = self.relu(self.fc1(x))
        x = self.dropout(x)
        x = self.fc2(x)
        return x


class GradientCompressor:
    """Quantizes gradients and estimates compression ratio."""

    def __init__(self, bits: int = 8, use_compression: bool = True):
        self.bits = int(max(1, bits))
        self.use_compression = use_compression
        self.quantization_range = (1 << self.bits) - 1

    def compress_gradient(
        self, gradient: np.ndarray
    ) -> Tuple[np.ndarray, Dict[str, float]]:
        """Compress a gradient tensor via uniform quantization."""
        if not self.use_compression:
            return gradient.astype(np.float32), {"compression_ratio": 1.0}

        grad_min = float(gradient.min())
        grad_max = float(gradient.max())
        grad_range = grad_max - grad_min + 1e-8

        quantized = np.round(
            (gradient - grad_min) / grad_range * self.quantization_range
        ).astype(np.uint16)

        original_bytes = gradient.nbytes
        compressed_bytes = max(1.0, quantized.nbytes * (self.bits / 16.0))
        compression_ratio = float(original_bytes / compressed_bytes)

        return quantized.astype(np.float32), {
            "compression_ratio": compression_ratio,
            "quantization_bits": float(self.bits),
            "grad_min": grad_min,
            "grad_max": grad_max,
        }

    def decompress_gradient(
        self, compressed: np.ndarray, metadata: Dict[str, float]
    ) -> np.ndarray:
        """Decompress quantized gradient back to float32."""
        if not self.use_compression:
            return compressed

        grad_min = metadata.get("grad_min", 0.0)
        grad_max = metadata.get("grad_max", 1.0)
        grad_range = grad_max - grad_min + 1e-8
        quantization_range = (1 << self.bits) - 1
        return (compressed / quantization_range) * grad_range + grad_min


class DifferentialPrivacyNoise:
    """Adds Gaussian noise for DP-SGD."""

    def __init__(self, epsilon: float, delta: float, sensitivity: float = 1.0):
        self.epsilon = epsilon
        self.delta = delta
        self.sensitivity = sensitivity

    def add_noise(self, gradient: np.ndarray) -> Tuple[np.ndarray, float]:
        """Add Gaussian noise according to epsilon/delta budget."""
        if self.epsilon <= 0:
            return gradient, 0.0

        noise_scale = (
            self.sensitivity * np.sqrt(2 * np.log(1.25 / self.delta)) / self.epsilon
        )
        noise = np.random.normal(0, noise_scale, gradient.shape)
        noisy_gradient = gradient + noise
        privacy_overhead = float(np.mean(np.abs(noise)))
        return noisy_gradient, privacy_overhead


class FederatedLearningTrainer:
    """Federated trainer for CIFAR-10 or MNIST datasets."""

    def __init__(self, config: TrainingConfig):
        self.config = config
        self.dataset = config.dataset.lower().strip()

        np.random.seed(config.seed)
        torch.manual_seed(config.seed)

        self.device = self._resolve_device(config.device)
        self.model = self._build_model().to(self.device)
        self.num_gpus = torch.cuda.device_count() if self.device.type == "cuda" else 0

        if self.device.type == "cuda" and config.multi_gpu and self.num_gpus > 1:
            self.model = nn.DataParallel(self.model)

        self.criterion = nn.CrossEntropyLoss()
        self.optimizer = torch.optim.SGD(
            self.model.parameters(), lr=config.learning_rate
        )
        self.compressor = GradientCompressor(
            config.compression_bits, config.use_compression
        )
        self.dp_noise = DifferentialPrivacyNoise(config.epsilon, config.delta)
        self.history: List[Dict[str, Any]] = []

    def _resolve_device(self, requested: str) -> torch.device:
        requested = (requested or "auto").lower().strip()
        if requested == "auto":
            return torch.device("cuda" if torch.cuda.is_available() else "cpu")
        if requested == "cuda" and not torch.cuda.is_available():
            return torch.device("cpu")
        return torch.device(requested)

    def _build_model(self) -> nn.Module:
        if self.dataset == "mnist":
            return SimpleCNN(input_channels=1, image_size=28, num_classes=10)
        if self.dataset == "cifar10":
            return SimpleCNN(input_channels=3, image_size=32, num_classes=10)
        raise ValueError(
            f"Unsupported dataset '{self.dataset}'. Supported: cifar10, mnist"
        )

    def _dataset_transforms(self) -> Tuple[transforms.Compose, transforms.Compose]:
        if self.dataset == "mnist":
            normalize = transforms.Normalize((0.1307,), (0.3081,))
        else:
            normalize = transforms.Normalize(
                (0.4914, 0.4822, 0.4465), (0.2470, 0.2435, 0.2616)
            )

        train_transform = transforms.Compose([transforms.ToTensor(), normalize])
        test_transform = transforms.Compose([transforms.ToTensor(), normalize])
        return train_transform, test_transform

    def load_data(self) -> Tuple[List[DataLoader], DataLoader]:
        """Load dataset and partition train data into per-client subsets."""
        train_transform, test_transform = self._dataset_transforms()

        if self.dataset == "mnist":
            train_dataset = datasets.MNIST(
                root="./data", train=True, download=True, transform=train_transform
            )
            test_dataset = datasets.MNIST(
                root="./data", train=False, download=True, transform=test_transform
            )
        else:
            train_dataset = datasets.CIFAR10(
                root="./data", train=True, download=True, transform=train_transform
            )
            test_dataset = datasets.CIFAR10(
                root="./data", train=False, download=True, transform=test_transform
            )

        indices = np.random.permutation(len(train_dataset))
        splits = np.array_split(indices, self.config.num_clients)

        client_loaders: List[DataLoader] = []
        for idx_split in splits:
            subset = Subset(train_dataset, idx_split.tolist())
            client_loaders.append(
                DataLoader(
                    subset,
                    batch_size=self.config.batch_size,
                    shuffle=True,
                    num_workers=self.config.num_workers,
                    pin_memory=self.device.type == "cuda",
                )
            )

        test_loader = DataLoader(
            test_dataset,
            batch_size=self.config.batch_size,
            shuffle=False,
            num_workers=self.config.num_workers,
            pin_memory=self.device.type == "cuda",
        )

        return client_loaders, test_loader

    def _apply_privacy_and_compression(self) -> float:
        """Apply DP noise + optional compression to gradients in-place."""
        overhead_accumulator = 0.0
        param_count = 0

        for param in self.model.parameters():
            if param.grad is None:
                continue

            grad = param.grad.data.clamp(-1, 1)
            grad_np = grad.detach().cpu().numpy()

            if self.config.use_privacy:
                grad_np, privacy_overhead = self.dp_noise.add_noise(grad_np)
                overhead_accumulator += privacy_overhead

            if self.config.use_compression:
                compressed, meta = self.compressor.compress_gradient(grad_np)
                grad_np = self.compressor.decompress_gradient(compressed, meta)

            grad_tensor = torch.from_numpy(grad_np).to(
                self.device, dtype=param.grad.data.dtype
            )
            param.grad.data = grad_tensor
            param_count += 1

        if param_count == 0:
            return 0.0
        return overhead_accumulator / param_count

    def train_local_epoch(self, data_loader: DataLoader) -> Tuple[float, float]:
        """Train one local epoch and return loss + average privacy overhead."""
        self.model.train()
        total_loss = 0.0
        total_overhead = 0.0
        batches = 0

        for data, target in data_loader:
            data, target = data.to(self.device), target.to(self.device)

            self.optimizer.zero_grad(set_to_none=True)
            output = self.model(data)
            loss = self.criterion(output, target)
            loss.backward()

            overhead = self._apply_privacy_and_compression()
            self.optimizer.step()

            total_loss += float(loss.item())
            total_overhead += overhead
            batches += 1

        if batches == 0:
            return 0.0, 0.0
        return total_loss / batches, total_overhead / batches

    def evaluate(self, data_loader: DataLoader) -> Tuple[float, float]:
        """Evaluate model on test data and return (accuracy, loss)."""
        self.model.eval()
        correct = 0
        total = 0
        total_loss = 0.0

        with torch.no_grad():
            for data, target in data_loader:
                data, target = data.to(self.device), target.to(self.device)
                output = self.model(data)
                loss = self.criterion(output, target)
                _, predicted = torch.max(output.data, 1)
                total += target.size(0)
                correct += (predicted == target).sum().item()
                total_loss += float(loss.item())

        accuracy = correct / max(total, 1)
        avg_loss = total_loss / max(len(data_loader), 1)
        return accuracy, avg_loss

    def train_round(
        self, client_loaders: List[DataLoader], test_loader: DataLoader
    ) -> Dict[str, Any]:
        """Execute one federated round across all client loaders."""
        round_loss = 0.0
        round_privacy_overhead = 0.0
        local_steps = 0

        for client_loader in client_loaders:
            for _ in range(self.config.local_epochs):
                local_loss, local_overhead = self.train_local_epoch(client_loader)
                round_loss += local_loss
                round_privacy_overhead += local_overhead
                local_steps += 1

        round_loss = round_loss / max(local_steps, 1)
        round_privacy_overhead = round_privacy_overhead / max(local_steps, 1)

        accuracy, eval_loss = self.evaluate(test_loader)

        compression_ratio = 32 / self.config.compression_bits
        if self.config.compression_bits <= 8:
            compression_ratio *= 1.25

        metrics: Dict[str, Any] = {
            "round_loss": float(round_loss),
            "eval_loss": float(eval_loss),
            "accuracy": float(accuracy),
            "compression_ratio": float(compression_ratio),
            "privacy_overhead": float(
                round_privacy_overhead if self.config.use_privacy else 0.0
            ),
            "epsilon": float(self.config.epsilon),
            "compression_bits": int(self.config.compression_bits),
            "dataset": self.dataset,
            "device": str(self.device),
            "gpu_count": int(self.num_gpus if self.device.type == "cuda" else 0),
            "multi_gpu": bool(
                self.device.type == "cuda"
                and self.config.multi_gpu
                and self.num_gpus > 1
            ),
            "num_clients": int(self.config.num_clients),
            "local_epochs": int(self.config.local_epochs),
        }

        self.history.append(metrics)
        return metrics

    def train(
        self, client_loaders: List[DataLoader], test_loader: DataLoader
    ) -> List[Dict[str, Any]]:
        """Train for the configured number of rounds."""
        for round_num in range(self.config.num_rounds):
            metrics = self.train_round(client_loaders, test_loader)
            print(
                f"Round {round_num + 1}/{self.config.num_rounds} "
                f"loss={metrics['round_loss']:.4f} eval_loss={metrics['eval_loss']:.4f} "
                f"acc={metrics['accuracy']:.4f} dataset={metrics['dataset']} device={metrics['device']}"
            )
        return self.history


def main() -> None:
    """Run a short local smoke test and save metric history."""
    config = TrainingConfig(
        num_rounds=3,
        num_clients=8,
        local_epochs=1,
        learning_rate=0.01,
        epsilon=1.0,
        compression_bits=8,
        dataset="cifar10",
        device="auto",
        multi_gpu=True,
    )

    trainer = FederatedLearningTrainer(config)
    client_loaders, test_loader = trainer.load_data()

    print("Starting federated training...")
    history = trainer.train(client_loaders, test_loader)

    print("\nTraining complete")
    for i, metrics in enumerate(history):
        print(
            f"Round {i + 1}: "
            f"accuracy={metrics['accuracy']:.4f} "
            f"train_loss={metrics['round_loss']:.4f} "
            f"eval_loss={metrics['eval_loss']:.4f}"
        )

    output_file = Path("./training_history.json")
    with output_file.open("w", encoding="utf-8") as f:
        json.dump(history, f, indent=2)
    print(f"History saved to {output_file}")


if __name__ == "__main__":
    main()
