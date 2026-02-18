#!/usr/bin/env python3
"""Sovereign FL Client with differential privacy"""

import flwr as fl
import torch
import torch.nn as nn
import torch.nn.functional as F
from torchvision import datasets, transforms
from torch.utils.data import DataLoader, Subset
from opacus import PrivacyEngine
import argparse
import logging
import numpy as np

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class MNISTNet(nn.Module):
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

class SovereignClient(fl.client.NumPyClient):
    def __init__(self, node_id, byzantine=False):
        self.node_id = node_id
        self.byzantine = byzantine
        self.device = torch.device("cpu")
        self.model = MNISTNet().to(self.device)
        self.trainloader = self._load_data(node_id)
        self.optimizer = torch.optim.SGD(self.model.parameters(), lr=0.01)

        self.privacy_engine = PrivacyEngine()
        self.model, self.optimizer, self.trainloader = self.privacy_engine.make_private(
            module=self.model,
            optimizer=self.optimizer,
            data_loader=self.trainloader,
            noise_multiplier=1.1,
            max_grad_norm=1.0,
        )

    def _load_data(self, node_id):
        transform = transforms.Compose([
            transforms.ToTensor(),
            transforms.Normalize((0.1307,), (0.3081,))
        ])

        dataset = datasets.MNIST('./data', train=True, download=True, transform=transform)
        samples_per_node = len(dataset) // 200
        start_idx = node_id * samples_per_node
        end_idx = start_idx + samples_per_node

        subset = Subset(dataset, range(start_idx, end_idx))
        return DataLoader(subset, batch_size=32, shuffle=True)

    def get_parameters(self, config):
        return [val.cpu().numpy() for _, val in self.model.state_dict().items()]

    def set_parameters(self, parameters):
        params_dict = zip(self.model.state_dict().keys(), parameters)
        state_dict = {k: torch.tensor(v) for k, v in params_dict}
        self.model.load_state_dict(state_dict, strict=True)

    def fit(self, parameters, config):
        self.set_parameters(parameters)
        self.model.train()

        for epoch in range(3):
            for data, target in self.trainloader:
                data, target = data.to(self.device), target.to(self.device)
                self.optimizer.zero_grad()
                output = self.model(data)
                loss = F.nll_loss(output, target)
                loss.backward()
                self.optimizer.step()

        epsilon = self.privacy_engine.get_epsilon(delta=1e-5)
        logger.info(f"Node {self.node_id}: Training complete, epsilon={epsilon:.2f}")

        updated_params = self.get_parameters(config)

        if self.byzantine:
            updated_params = [-p for p in updated_params]
            logger.warning(f"Node {self.node_id}: Sent BYZANTINE update")

        return updated_params, len(self.trainloader.dataset), {}

    def evaluate(self, parameters, config):
        return 0.0, 0, {}

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--node-id', type=int, required=True)
    parser.add_argument('--aggregator', type=str, required=True)
    parser.add_argument('--byzantine', action='store_true')
    args = parser.parse_args()

    client = SovereignClient(node_id=args.node_id, byzantine=args.byzantine)

    fl.client.start_numpy_client(
        server_address=args.aggregator,
        client=client
    )

if __name__ == "__main__":
    main()
