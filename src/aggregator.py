#!/usr/bin/env python3
"""Sovereign FL Aggregator with Multi-Krum Byzantine tolerance"""

import flwr as fl
from typing import List, Dict
import torch
import torch.nn as nn
import torch.nn.functional as F
import numpy as np
from prometheus_client import start_http_server, Counter, Gauge
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

ROUND_COUNTER = Counter('fl_rounds_total', 'Total FL rounds completed')
ACCURACY_GAUGE = Gauge('fl_accuracy', 'Current model accuracy')
CLIENT_GAUGE = Gauge('fl_connected_clients', 'Number of connected clients')
BYZANTINE_COUNTER = Counter('fl_byzantine_detected', 'Byzantine nodes detected')

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

class MultiKrumStrategy(fl.server.strategy.Strategy):
    def __init__(self, num_clients=200, num_byzantine=20):
        self.num_clients = num_clients
        self.num_byzantine = num_byzantine
        self.model = MNISTNet()

    def initialize_parameters(self, client_manager):
        return [val.cpu().numpy() for _, val in self.model.state_dict().items()]

    def configure_fit(self, server_round, parameters, client_manager):
        sample_size = min(self.num_clients, client_manager.num_available())
        clients = client_manager.sample(num_clients=sample_size, min_num_clients=sample_size)
        CLIENT_GAUGE.set(len(clients))
        return [(client, fl.common.FitIns(parameters, {})) for client in clients]

    def aggregate_fit(self, server_round, results, failures):
        if not results:
            return None, {}

        weights_list = [fl.common.parameters_to_ndarrays(res.parameters) for _, res in results]
        selected_indices = self._multi_krum_select(weights_list)

        byzantine_count = len(weights_list) - len(selected_indices)
        BYZANTINE_COUNTER.inc(byzantine_count)

        aggregated = self._aggregate_weights([weights_list[i] for i in selected_indices])
        parameters_aggregated = fl.common.ndarrays_to_parameters(aggregated)

        ROUND_COUNTER.inc()
        logger.info(f"Round {server_round}: Selected {len(selected_indices)}/{len(results)} updates")

        return parameters_aggregated, {}

    def _multi_krum_select(self, weights_list):
        n = len(weights_list)
        m = self.num_byzantine

        flattened = [np.concatenate([w.flatten() for w in weights]) for weights in weights_list]
        distances = np.zeros((n, n))

        for i in range(n):
            for j in range(i + 1, n):
                dist = np.linalg.norm(flattened[i] - flattened[j])
                distances[i, j] = dist
                distances[j, i] = dist

        scores = []
        num_neighbors = n - m - 2
        for i in range(n):
            sorted_dists = np.sort(distances[i])
            score = np.sum(sorted_dists[1:num_neighbors+1])
            scores.append((score, i))

        scores.sort()
        num_selection = min(n - 2*m + 2, n)
        return [idx for _, idx in scores[:num_selection]]

    def _aggregate_weights(self, weights_list):
        n = len(weights_list)
        aggregated = []
        for layer_idx in range(len(weights_list[0])):
            layer_sum = np.zeros_like(weights_list[0][layer_idx])
            for weights in weights_list:
                layer_sum += weights[layer_idx]
            aggregated.append(layer_sum / n)
        return aggregated

    def configure_evaluate(self, server_round, parameters, client_manager):
        return []

    def aggregate_evaluate(self, server_round, results, failures):
        return None, {}

def main():
    start_http_server(9090)
    logger.info("Starting Sovereign FL Aggregator on port 8080")

    strategy = MultiKrumStrategy(num_clients=200, num_byzantine=20)

    fl.server.start_server(
        server_address="0.0.0.0:8080",
        config=fl.server.ServerConfig(num_rounds=30),
        strategy=strategy
    )

if __name__ == "__main__":
    main()
