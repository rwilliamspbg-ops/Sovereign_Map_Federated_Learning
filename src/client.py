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
    
    def __init__(self, node_id: int, byzantine: bool = False, server_address: str = "localhost:8080"):
        self.node_id = node_id
        self.byzantine = byzantine
        self.server_address = server_address
        self.device = torch.device("cpu")
        self.model = MNISTNet().to(self.device)
        self.batch_size = int(os.getenv("BATCH_SIZE", "16"))
        self.local_epochs = int(os.getenv("LOCAL_EPOCHS", "1"))
        self.enable_dp = os.getenv("ENABLE_DP", "false").lower() in ("1", "true", "yes")
        self.max_samples_per_node = int(os.getenv("MAX_SAMPLES_PER_NODE", "120"))
        self.trainloader = self._load_data(node_id)
        self.optimizer = torch.optim.SGD(self.model.parameters(), lr=0.01)
        
        # Differential privacy setup
        self.privacy_engine = None
        if self.enable_dp:
            self.privacy_engine = PrivacyEngine()
            try:
                self.model, self.optimizer, self.trainloader = self.privacy_engine.make_private(
                    module=self.model,
                    optimizer=self.optimizer,
                    data_loader=self.trainloader,
                    noise_multiplier=1.1,
                    max_grad_norm=1.0,
                )
                logger.info(f"Node {self.node_id}: Differential privacy enabled")
            except Exception as e:
                logger.warning(f"Node {self.node_id}: Could not enable DP: {e}, continuing without privacy")
                self.privacy_engine = None
        else:
            logger.info(f"Node {self.node_id}: Differential privacy disabled")
        
        logger.info(f"Node {self.node_id}: Initialized (Byzantine={byzantine})")

    def _load_data(self, node_id: int) -> DataLoader:
        """Load MNIST data with node-specific subset."""
        transform = transforms.Compose([
            transforms.ToTensor(),
            transforms.Normalize((0.1307,), (0.3081,))
        ])

        try:
            dataset = datasets.MNIST('./data', train=True, download=True, transform=transform)
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

    def fit(self, parameters: List[np.ndarray], config: Dict) -> Tuple[List[np.ndarray], int, Dict]:
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

        # Byzantine attack: invert parameters
        if self.byzantine:
            updated_params = [-p for p in updated_params]
            logger.warning(f"Node {self.node_id}: Sent BYZANTINE update")

        num_samples = len(self.trainloader.dataset)
        metrics = {
            "byzantine": self.byzantine,
            "avg_loss": float(np.mean(loss_history)) if loss_history else 0.0,
        }
        if epsilon is not None:
            metrics["epsilon"] = float(epsilon)

        logger.info(f"Node {self.node_id}: Training complete | Loss={metrics['avg_loss']:.4f} | Samples={num_samples}")
        return updated_params, num_samples, metrics

    def evaluate(self, parameters: List[np.ndarray], config: Dict) -> Tuple[float, int, Dict]:
        """Evaluate local model (optional)."""
        # For testnet, we skip local evaluation
        return 0.0, len(self.trainloader.dataset), {}

# ============================================================================
# MAIN
# ============================================================================

def main():
    parser = argparse.ArgumentParser(description="Sovereign Map Federated Learning Client")
    parser.add_argument('--node-id', type=int, required=True, help='Unique node ID')
    parser.add_argument('--aggregator', type=str, required=True, help='Aggregator address (host:port)')
    parser.add_argument('--byzantine', action='store_true', help='Run as Byzantine node')
    
    args = parser.parse_args()

    # Validate aggregator address
    if ':' not in args.aggregator:
        logger.error("Aggregator address must be in format 'host:port'")
        sys.exit(1)

    logger.info(f"Connecting to aggregator at {args.aggregator}")
    logger.info(f"Node ID: {args.node_id}")
    logger.info(f"Byzantine: {args.byzantine}")

    client = SovereignClient(
        node_id=args.node_id,
        byzantine=args.byzantine,
        server_address=args.aggregator
    )

    # Connect to Flower server
    try:
        fl.client.start_numpy_client(
            server_address=args.aggregator,
            client=client,
            grpc_max_message_length=1024 * 1024 * 1024,
        )
    except Exception as e:
        logger.error(f"Failed to connect to aggregator: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
