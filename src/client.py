#!/usr/bin/env python3
"""
Sovereign Map Federated Learning Client
========================================
Flower-based client with:
- Byzantine node support (can send inverted updates)
- Differential privacy (Opacus)
- MNIST training
- Stake tracking (simulated)
"""

import argparse
import logging
import os
import sys
from typing import Dict, List, Tuple

import flwr as fl
import numpy as np
import torch
import torch.nn as nn
import torch.nn.functional as F
from opacus import PrivacyEngine
from torch.utils.data import DataLoader, Subset
from torchvision import datasets, transforms

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# ============================================================================
# NEURAL NETWORK MODEL
# ============================================================================


class MNISTNet(nn.Module):
    """Simple CNN for MNIST classification."""

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
# SOVEREIGN FL CLIENT
# ============================================================================


class SovereignClient(fl.client.NumPyClient):
    """Federated learning client with Byzantine and privacy support."""

    def __init__(
        self,
        node_id: int,
        byzantine: bool = False,
        server_address: str = "localhost:8080",
    ):
        self.node_id = node_id
        self.byzantine = byzantine
        self.server_address = server_address
        self.device = self._select_device()
        self.model = self._initialize_model_on_device()
        self.batch_size = int(os.getenv("BATCH_SIZE", "16"))
        self.local_epochs = int(os.getenv("LOCAL_EPOCHS", "1"))
        self.enable_dp = os.getenv("ENABLE_DP", "false").lower() in ("1", "true", "yes")
        self.max_samples_per_node = int(os.getenv("MAX_SAMPLES_PER_NODE", "120"))
        self.llm_model_family = os.getenv("LLM_ADAPTER_MODEL_FAMILY", "llama-3.1")
        self.llm_model_version = os.getenv("LLM_ADAPTER_MODEL_VERSION", "8b-instruct")
        self.llm_tokenizer_hash = os.getenv(
            "LLM_ADAPTER_TOKENIZER_HASH", "local-dev-tokenizer-v1"
        )
        self.llm_adapter_rank = int(os.getenv("LLM_ADAPTER_RANK", "16"))
        self.llm_target_modules = os.getenv(
            "LLM_ADAPTER_TARGET_MODULES", "q_proj,v_proj"
        )
        self.trainloader = self._load_data(node_id)
        self.optimizer = torch.optim.SGD(self.model.parameters(), lr=0.01)

        # Differential privacy setup
        self.privacy_engine = None
        if self.enable_dp:
            self.privacy_engine = PrivacyEngine()
            try:
                (
                    self.model,
                    self.optimizer,
                    self.trainloader,
                ) = self.privacy_engine.make_private(
                    module=self.model,
                    optimizer=self.optimizer,
                    data_loader=self.trainloader,
                    noise_multiplier=1.1,
                    max_grad_norm=1.0,
                )
                logger.info(f"Node {self.node_id}: Differential privacy enabled")
            except Exception as e:
                logger.warning(
                    f"Node {self.node_id}: Could not enable DP: {e}, continuing without privacy"
                )
                self.privacy_engine = None
        else:
            logger.info(f"Node {self.node_id}: Differential privacy disabled")

        logger.info(
            f"Node {self.node_id}: Initialized (Byzantine={byzantine}, Device={self.device})"
        )

    def _initialize_model_on_device(self) -> nn.Module:
        """Initialize model on selected device with safe CPU fallback."""
        model = MNISTNet()
        try:
            model = model.to(self.device)
        except Exception as e:
            logger.warning(
                f"Node {self.node_id}: Could not initialize model on {self.device} ({e}), falling back to CPU"
            )
            self.device = torch.device("cpu")
            model = model.to(self.device)
        return model

    def _env_enabled(self, name: str, default: str = "true") -> bool:
        """Parse boolean environment flags consistently."""
        return os.getenv(name, default).lower() in ("1", "true", "yes")

    def _first_visible_device_index(self, *env_names: str, default: int = 0) -> int:
        """Return the first visible device index from the configured env vars."""
        for env_name in env_names:
            raw_value = os.getenv(env_name)
            if raw_value is None:
                continue

            device_index_raw = str(raw_value).split(",")[0].strip()
            if not device_index_raw:
                continue

            try:
                return int(device_index_raw)
            except ValueError:
                logger.warning(
                    f"Node {self.node_id}: Invalid {env_name}={raw_value}, using default device {default}"
                )
                return default

        return default

    def _try_accelerator(
        self,
        *,
        backend_name: str,
        backend_attr: str,
        visible_env_names: Tuple[str, ...],
        label: str,
    ) -> object | None:
        """Try a specific accelerator backend and return a selected device on success."""
        backend = getattr(torch, backend_attr, None)
        if backend is None:
            return None

        try:
            if not backend.is_available():
                return None

            device_index = self._first_visible_device_index(*visible_env_names)
            selected_device = torch.device(f"{backend_name}:{device_index}")

            if hasattr(backend, "set_device"):
                backend.set_device(selected_device)

            if self._probe_device(selected_device):
                logger.info(
                    f"Node {self.node_id}: Using {label} device {selected_device}"
                )
                return selected_device
        except Exception as e:
            logger.warning(
                f"Node {self.node_id}: {label} requested but unavailable ({e}), falling back"
            )

        return None

    def _probe_device(self, device: torch.device) -> bool:
        """Verify device is actually usable by running a tiny allocation/op."""
        if not hasattr(torch, "zeros"):
            logger.debug(
                f"Node {self.node_id}: Skipping device probe for {device} (torch.zeros unavailable)"
            )
            return True

        try:
            probe = torch.zeros((1, 1), device=device)
            _ = probe + 1
            if device.type == "cuda":
                torch.cuda.synchronize()
            elif (
                device.type == "xpu"
                and hasattr(torch, "xpu")
                and hasattr(torch.xpu, "synchronize")
            ):
                torch.xpu.synchronize()
            elif (
                device.type == "npu"
                and hasattr(torch, "npu")
                and hasattr(torch.npu, "synchronize")
            ):
                torch.npu.synchronize()
            return True
        except Exception as e:
            logger.warning(
                f"Node {self.node_id}: Device probe failed for {device} ({e})"
            )
            return False

    def _fallback_to_cpu(self, reason: str) -> None:
        """Move model to CPU after runtime accelerator failures."""
        if str(self.device) == "cpu":
            return
        logger.warning(f"Node {self.node_id}: Falling back to CPU ({reason})")
        self.device = torch.device("cpu")
        self.model = self.model.to(self.device)

    def _select_device(self) -> torch.device:
        """Select training device with NPU/XPU/CUDA/MPS/CPU fallback."""
        force_cpu = self._env_enabled("FORCE_CPU", "false")
        if force_cpu:
            logger.info(f"Node {self.node_id}: FORCE_CPU enabled")
            return torch.device("cpu")

        npu_enabled = self._env_enabled("NPU_ENABLED", "true")
        if npu_enabled:
            npu_device = self._try_accelerator(
                backend_name="npu",
                backend_attr="npu",
                visible_env_names=("ASCEND_RT_VISIBLE_DEVICES",),
                label="NPU",
            )
            if npu_device is not None:
                return npu_device

        xpu_enabled = self._env_enabled("XPU_ENABLED", "true")
        if xpu_enabled:
            xpu_device = self._try_accelerator(
                backend_name="xpu",
                backend_attr="xpu",
                visible_env_names=("XPU_VISIBLE_DEVICES", "ZE_AFFINITY_MASK"),
                label="XPU",
            )
            if xpu_device is not None:
                return xpu_device

        gpu_enabled = self._env_enabled("GPU_ENABLED", "true")
        if gpu_enabled and hasattr(torch, "cuda") and torch.cuda.is_available():
            gpu_device = self._try_accelerator(
                backend_name="cuda",
                backend_attr="cuda",
                visible_env_names=(
                    "CUDA_VISIBLE_DEVICES",
                    "HIP_VISIBLE_DEVICES",
                    "ROCR_VISIBLE_DEVICES",
                ),
                label=(
                    "ROCm GPU"
                    if getattr(getattr(torch, "version", None), "hip", None)
                    else "CUDA GPU"
                ),
            )
            if gpu_device is not None:
                return gpu_device

        mps_backend = getattr(getattr(torch, "backends", None), "mps", None)
        mps_enabled = self._env_enabled("MPS_ENABLED", "true")
        if mps_enabled and mps_backend is not None:
            try:
                if mps_backend.is_available():
                    mps_device = torch.device("mps")
                    if self._probe_device(mps_device):
                        logger.info(
                            f"Node {self.node_id}: Using MPS device {mps_device}"
                        )
                        return mps_device
            except Exception as e:
                logger.warning(
                    f"Node {self.node_id}: MPS requested but unavailable ({e}), falling back"
                )

        logger.info(f"Node {self.node_id}: Using CPU device")
        return torch.device("cpu")

    def _load_data(self, node_id: int) -> DataLoader:
        """Load MNIST data with node-specific subset."""
        transform = transforms.Compose(
            [transforms.ToTensor(), transforms.Normalize((0.1307,), (0.3081,))]
        )

        try:
            dataset = datasets.MNIST(
                "./data", train=True, download=True, transform=transform
            )
        except Exception as e:
            logger.warning(f"Could not download MNIST, using random data: {e}")
            # Fallback: generate random data
            return self._create_random_dataloader()

        samples_per_node = max(len(dataset) // 200, 10)
        samples_per_node = min(samples_per_node, self.max_samples_per_node)
        start_idx = (node_id % (len(dataset) // samples_per_node)) * samples_per_node
        end_idx = min(start_idx + samples_per_node, len(dataset))

        subset = Subset(dataset, range(start_idx, end_idx))
        logger.info(f"Node {self.node_id}: Loaded {len(subset)} training samples")
        return DataLoader(subset, batch_size=self.batch_size, shuffle=True)

    def _create_random_dataloader(self) -> DataLoader:
        """Create random data for testing."""
        from torch.utils.data import TensorDataset

        X = torch.randn(100, 1, 28, 28)
        y = torch.randint(0, 10, (100,))
        dataset = TensorDataset(X, y)
        return DataLoader(dataset, batch_size=self.batch_size, shuffle=True)

    def get_parameters(self, config: Dict) -> List[np.ndarray]:
        """Extract model parameters as numpy arrays."""
        return [val.cpu().numpy() for _, val in self.model.state_dict().items()]

    def set_parameters(self, parameters: List[np.ndarray]) -> None:
        """Set model parameters from numpy arrays."""
        params_dict = zip(self.model.state_dict().keys(), parameters)
        state_dict = {k: torch.tensor(v) for k, v in params_dict}
        self.model.load_state_dict(state_dict, strict=False)

    def fit(
        self, parameters: List[np.ndarray], config: Dict
    ) -> Tuple[List[np.ndarray], int, Dict]:
        """Train local model."""
        self.set_parameters(parameters)
        self.model.train()

        epochs = int(config.get("local_epochs", self.local_epochs))
        loss_history = []

        for epoch in range(epochs):
            epoch_loss = 0.0
            batch_count = 0

            try:
                for data, target in self.trainloader:
                    try:
                        data, target = data.to(self.device), target.to(self.device)
                    except Exception as e:
                        self._fallback_to_cpu(f"batch transfer failed: {e}")
                        data, target = data.to(self.device), target.to(self.device)

                    self.optimizer.zero_grad()
                    output = self.model(data)
                    loss = F.nll_loss(output, target)
                    loss.backward()
                    self.optimizer.step()

                    epoch_loss += loss.item()
                    batch_count += 1
            except Exception as e:
                logger.error(f"Node {self.node_id}: Training error: {e}")
                break

            avg_loss = epoch_loss / max(batch_count, 1)
            loss_history.append(avg_loss)

        # Get privacy metrics if available
        epsilon = None
        if self.privacy_engine is not None:
            try:
                epsilon = self.privacy_engine.get_epsilon(delta=1e-5)
            except Exception as e:
                logger.debug(f"Node {self.node_id}: Could not get epsilon: {e}")

        updated_params = self.get_parameters(config)
        update_l2_norm = float(
            np.sqrt(sum(float(np.square(param).sum()) for param in updated_params))
        )

        # Byzantine attack: invert parameters
        if self.byzantine:
            updated_params = [-p for p in updated_params]
            logger.warning(f"Node {self.node_id}: Sent BYZANTINE update")

        num_samples = len(self.trainloader.dataset)
        metrics = {
            "byzantine": self.byzantine,
            "avg_loss": float(np.mean(loss_history)) if loss_history else 0.0,
            "device": str(self.device),
            "llm_model_family": self.llm_model_family,
            "llm_model_version": self.llm_model_version,
            "llm_tokenizer_hash": self.llm_tokenizer_hash,
            "llm_adapter_rank": self.llm_adapter_rank,
            "llm_target_modules": self.llm_target_modules,
            "llm_reported_update_l2_norm": update_l2_norm,
        }
        if epsilon is not None:
            metrics["epsilon"] = float(epsilon)

        logger.info(
            f"Node {self.node_id}: Training complete | Loss={metrics['avg_loss']:.4f} | Samples={num_samples}"
        )
        return updated_params, num_samples, metrics

    def evaluate(
        self, parameters: List[np.ndarray], config: Dict
    ) -> Tuple[float, int, Dict]:
        """Evaluate local model (optional)."""
        # For testnet, we skip local evaluation
        return 0.0, len(self.trainloader.dataset), {}


# ============================================================================
# MAIN
# ============================================================================


def main():
    parser = argparse.ArgumentParser(
        description="Sovereign Map Federated Learning Client"
    )
    parser.add_argument("--node-id", type=int, required=True, help="Unique node ID")
    parser.add_argument(
        "--aggregator", type=str, required=True, help="Aggregator address (host:port)"
    )
    parser.add_argument(
        "--byzantine", action="store_true", help="Run as Byzantine node"
    )

    args = parser.parse_args()

    # Validate aggregator address
    if ":" not in args.aggregator:
        logger.error("Aggregator address must be in format 'host:port'")
        sys.exit(1)

    logger.info(f"Connecting to aggregator at {args.aggregator}")
    logger.info(f"Node ID: {args.node_id}")
    logger.info(f"Byzantine: {args.byzantine}")

    client = SovereignClient(
        node_id=args.node_id, byzantine=args.byzantine, server_address=args.aggregator
    )

    # Connect to Flower server
    try:
        fl.client.start_client(
            server_address=args.aggregator,
            client=client.to_client(),
            grpc_max_message_length=1024 * 1024 * 1024,
        )
    except Exception as e:
        logger.error(f"Failed to connect to aggregator: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
