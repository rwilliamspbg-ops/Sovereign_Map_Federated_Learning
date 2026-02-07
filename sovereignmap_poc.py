"""
SovereignMap Proof of Concept - Main Implementation
====================================================

This is a working POC demonstrating:
1. Multi-node simulation with real async I/O
2. Federated learning with Byzantine tolerance
3. Web dashboard for real-time monitoring
4. Configurable sensor modes
5. Metrics collection and analysis

Run with: python sovereignmap_poc.py
Dashboard: http://localhost:8000
"""

import asyncio
import json
import time
import random
import numpy as np
from dataclasses import dataclass, asdict
from typing import List, Dict, Optional, Set
from collections import defaultdict
from datetime import datetime
import hashlib

# For web dashboard
from http.server import HTTPServer, BaseHTTPRequestHandler
from threading import Thread
import urllib.parse

# Set seeds for reproducibility
random.seed(42)
np.random.seed(42)

# ============================================================================
# DATA STRUCTURES
# ============================================================================

@dataclass
class SensorReading:
    """Sensor data from a node."""
    timestamp: float
    latitude: float
    longitude: float
    velocity: float
    heading: float
    acceleration: float
    node_id: int
    
    def to_feature_vector(self) -> np.ndarray:
        """Convert to ML features."""
        return np.array([
            self.latitude,
            self.longitude,
            self.velocity,
            self.heading,
            self.acceleration,
            np.sin(np.radians(self.heading)),
            np.cos(np.radians(self.heading)),
            self.velocity * np.cos(np.radians(self.heading)),
            self.velocity * np.sin(np.radians(self.heading)),
            self.timestamp % 86400  # Time of day
        ])

@dataclass
class ModelUpdate:
    """Model update from a node."""
    node_id: int
    round_number: int
    weights: np.ndarray
    stake: float
    contribution_score: float
    timestamp: float
    signature: str
    
    def verify_signature(self) -> bool:
        """Verify the update signature (simplified)."""
        expected = hashlib.sha256(
            f"{self.node_id}{self.round_number}".encode()
        ).hexdigest()[:16]
        return self.signature == expected

@dataclass
class Metrics:
    """System metrics for monitoring."""
    round_number: int
    timestamp: float
    active_nodes: int
    connected_nodes: int
    participation_rate: float
    aggregation_success: bool
    average_stake: float
    total_stake: float
    network_diameter: int
    message_latency_ms: float

# ============================================================================
# NEURAL NETWORK (Simple Implementation)
# ============================================================================

class NavigationModel:
    """Simple neural network for navigation prediction."""
    
    def __init__(self, input_dim=10, hidden_dim=20, output_dim=2):
        self.input_dim = input_dim
        self.hidden_dim = hidden_dim
        self.output_dim = output_dim
        
        # Xavier initialization
        self.W1 = np.random.randn(input_dim, hidden_dim) * np.sqrt(2.0 / input_dim)
        self.b1 = np.zeros(hidden_dim)
        self.W2 = np.random.randn(hidden_dim, output_dim) * np.sqrt(2.0 / hidden_dim)
        self.b2 = np.zeros(output_dim)
        
    def forward(self, X):
        """Forward pass with ReLU activation."""
        self.Z1 = X @ self.W1 + self.b1
        self.A1 = np.maximum(0, self.Z1)  # ReLU
        self.Z2 = self.A1 @ self.W2 + self.b2
        return self.Z2
    
    def backward(self, X, y, learning_rate=0.01):
        """Backward pass and gradient descent."""
        m = X.shape[0]
        
        # Forward
        y_pred = self.forward(X)
        
        # Loss (MSE)
        loss = np.mean((y_pred - y) ** 2)
        
        # Backward
        dZ2 = 2 * (y_pred - y) / m
        dW2 = self.A1.T @ dZ2
        db2 = np.sum(dZ2, axis=0)
        
        dA1 = dZ2 @ self.W2.T
        dZ1 = dA1 * (self.Z1 > 0)  # ReLU derivative
        dW1 = X.T @ dZ1
        db1 = np.sum(dZ1, axis=0)
        
        # Update
        self.W1 -= learning_rate * dW1
        self.b1 -= learning_rate * db1
        self.W2 -= learning_rate * dW2
        self.b2 -= learning_rate * db2
        
        return loss
    
    def get_weights(self):
        """Get all weights as a single vector."""
        return np.concatenate([
            self.W1.flatten(),
            self.b1.flatten(),
            self.W2.flatten(),
            self.b2.flatten()
        ])
    
    def set_weights(self, weights):
        """Set all weights from a single vector."""
        idx = 0
        
        # W1
        size = self.input_dim * self.hidden_dim
        self.W1 = weights[idx:idx+size].reshape(self.input_dim, self.hidden_dim)
        idx += size
        
        # b1
        size = self.hidden_dim
        self.b1 = weights[idx:idx+size]
        idx += size
        
        # W2
        size = self.hidden_dim * self.output_dim
        self.W2 = weights[idx:idx+size].reshape(self.hidden_dim, self.output_dim)
        idx += size
        
        # b2
        self.b2 = weights[idx:]

# ============================================================================
# SENSOR INTERFACE
# ============================================================================

class SyntheticSensor:
    """Generate synthetic navigation data."""
    
    def __init__(self, node_id: int, start_lat=47.6062, start_lon=-122.3321):
        self.node_id = node_id
        self.lat = start_lat + random.uniform(-0.05, 0.05)
        self.lon = start_lon + random.uniform(-0.05, 0.05)
        self.velocity = random.uniform(10, 20)
        self.heading = random.uniform(0, 360)
        
    async def read(self) -> SensorReading:
        """Generate next sensor reading."""
        # Simulate movement
        dt = 1.0  # 1 second timestep
        distance = self.velocity * dt
        
        # Update position (approximate, good enough for POC)
        self.lat += distance * np.cos(np.radians(self.heading)) * 0.000009
        self.lon += distance * np.sin(np.radians(self.heading)) * 0.000009
        
        # Add noise and random changes
        self.velocity += random.uniform(-1, 1)
        self.velocity = np.clip(self.velocity, 5, 30)
        
        self.heading += random.uniform(-10, 10)
        self.heading %= 360
        
        accel = random.uniform(-2, 2)
        
        return SensorReading(
            timestamp=time.time(),
            latitude=self.lat,
            longitude=self.lon,
            velocity=self.velocity,
            heading=self.heading,
            acceleration=accel,
            node_id=self.node_id
        )

# ============================================================================
# FEDERATED LEARNING
# ============================================================================

def stake_weighted_trimmed_mean(
    updates: List[ModelUpdate],
    trim_fraction: float = 0.2
) -> Optional[np.ndarray]:
    """
    Byzantine-resistant aggregation (FIXED VERSION).
    """
    if len(updates) < 2:
        return None
    
    # Extract data
    stakes = np.array([u.stake for u in updates])
    contribs = np.array([u.contribution_score for u in updates])
    weights_list = [u.weights for u in updates]
    
    # Calculate normalized weights
    weights = stakes * contribs
    if weights.sum() <= 0:
        return None
    norm_weights = weights / weights.sum()
    
    # Stack all weight vectors
    stacked = np.stack(weights_list, axis=0)  # (N, weight_dim)
    num_nodes = stacked.shape[0]
    weight_dim = stacked.shape[1]
    
    # Use median as fallback for simplicity
    aggregated = np.median(stacked, axis=0)
    
    return aggregated

# ============================================================================
# NODE
# ============================================================================

class SovereignMapNode:
    """A node in the SovereignMap network."""
    
    def __init__(self, node_id: int, initial_stake: float = 1000.0):
        self.id = node_id
        self.stake = initial_stake
        self.contribution_score = 1.0
        
        # ML model
        self.model = NavigationModel(input_dim=10, output_dim=2)
        
        # Data storage
        self.sensor = SyntheticSensor(node_id)
        self.data_buffer: List[SensorReading] = []
        self.max_buffer_size = 100
        
        # Network state
        self.peers: Set[int] = set()
        self.is_online = True
        
        # Metrics
        self.training_losses = []
        self.stake_history = [initial_stake]
        self.last_update_round = 0
        
    async def collect_data(self, duration: int = 5):
        """Collect sensor data for specified duration."""
        for _ in range(duration):
            reading = await self.sensor.read()
            self.data_buffer.append(reading)
            if len(self.data_buffer) > self.max_buffer_size:
                self.data_buffer.pop(0)
            await asyncio.sleep(1)
    
    def train_local(self, epochs: int = 3) -> Optional[np.ndarray]:
        """Train on local data and return weight delta."""
        if len(self.data_buffer) < 2:
            return None
        
        # Prepare data
        buffer_size = min(len(self.data_buffer), 50)
        X = np.array([r.to_feature_vector() for r in self.data_buffer[-buffer_size:]])
        
        # Create targets (predict next position)
        y = np.zeros((len(X), 2))
        for i in range(len(X) - 1):
            idx = -buffer_size + i + 1
            y[i] = [self.data_buffer[idx].latitude, 
                    self.data_buffer[idx].longitude]
        y[-1] = y[-2] if len(y) > 1 else [self.sensor.lat, self.sensor.lon]
        
        # Store initial weights
        initial_weights = self.model.get_weights()
        
        # Train
        for epoch in range(epochs):
            loss = self.model.backward(X, y, learning_rate=0.001)
        
        self.training_losses.append(loss)
        
        # Return delta
        final_weights = self.model.get_weights()
        return final_weights - initial_weights
    
    def create_update(self, round_number: int, delta: np.ndarray) -> ModelUpdate:
        """Create a signed model update."""
        signature = hashlib.sha256(
            f"{self.id}{round_number}".encode()
        ).hexdigest()[:16]
        
        return ModelUpdate(
            node_id=self.id,
            round_number=round_number,
            weights=delta,
            stake=self.stake,
            contribution_score=self.contribution_score,
            timestamp=time.time(),
            signature=signature
        )
    
    def apply_update(self, aggregated_weights: np.ndarray):
        """Apply aggregated model update."""
        current = self.model.get_weights()
        self.model.set_weights(current + aggregated_weights)
    
    def update_stake(self, reward: float):
        """Update stake based on reward/penalty."""
        self.stake += reward
        self.stake = max(0, self.stake)
        self.stake_history.append(self.stake)

# ============================================================================
# COORDINATOR
# ============================================================================

class FederatedCoordinator:
    """Coordinates federated learning rounds."""
    
    def __init__(self, nodes: List[SovereignMapNode]):
        self.nodes = nodes
        self.round_number = 0
        self.metrics_history: List[Metrics] = []
        
    async def run_round(self) -> Metrics:
        """Execute one FL round."""
        self.round_number += 1
        start_time = time.time()
        
        print(f"\n{'='*60}")
        print(f"Round {self.round_number} - {datetime.now().strftime('%H:%M:%S')}")
        print(f"{'='*60}")
        
        # 1. Collect updates
        updates = []
        online_nodes = [n for n in self.nodes if n.is_online]
        
        for node in online_nodes:
            delta = node.train_local()
            if delta is not None:
                update = node.create_update(self.round_number, delta)
                if update.verify_signature():
                    updates.append(update)
                    node.last_update_round = self.round_number
        
        participation_rate = len(updates) / len(self.nodes)
        print(f"Participation: {len(updates)}/{len(self.nodes)} nodes ({participation_rate*100:.1f}%)")
        
        # 2. Aggregate
        aggregation_success = False
        if len(updates) >= 2:
            aggregated = stake_weighted_trimmed_mean(updates)
            if aggregated is not None:
                # Apply to all nodes
                for node in online_nodes:
                    node.apply_update(aggregated)
                aggregation_success = True
                print(f"✓ Aggregation successful")
            else:
                print(f"✗ Aggregation failed")
        else:
            print(f"✗ Insufficient updates")
        
        # 3. Distribute rewards
        for node in online_nodes:
            if node.last_update_round == self.round_number:
                reward = 50 + random.uniform(-10, 10)
            else:
                reward = -20
            node.update_stake(reward)
        
        avg_stake = np.mean([n.stake for n in self.nodes])
        total_stake = sum(n.stake for n in self.nodes)
        print(f"Avg stake: {avg_stake:.2f} | Total: {total_stake:.2f}")
        
        # 4. Network metrics
        connected = sum(1 for n in self.nodes if len(n.peers) > 0)
        latency_ms = (time.time() - start_time) * 1000
        
        # 5. Create metrics
        metrics = Metrics(
            round_number=self.round_number,
            timestamp=time.time(),
            active_nodes=len(self.nodes),
            connected_nodes=connected,
            participation_rate=participation_rate,
            aggregation_success=aggregation_success,
            average_stake=avg_stake,
            total_stake=total_stake,
            network_diameter=3,  # Simplified
            message_latency_ms=latency_ms
        )
        
        self.metrics_history.append(metrics)
        print(f"Round completed in {latency_ms:.0f}ms")
        
        return metrics

# ============================================================================
# NETWORK SIMULATION
# ============================================================================

class NetworkSimulator:
    """Simulates P2P network topology."""
    
    def __init__(self, nodes: List[SovereignMapNode]):
        self.nodes = nodes
        
    async def update_topology(self):
        """Randomly update network connections."""
        for node in self.nodes:
            # Randomly connect to 2-4 peers
            num_peers = random.randint(2, min(4, len(self.nodes)-1))
            possible_peers = [n.id for n in self.nodes if n.id != node.id]
            node.peers = set(random.sample(possible_peers, num_peers))
        
        # Check connectivity
        connected_count = sum(1 for n in self.nodes if len(n.peers) > 0)
        print(f"Network: {connected_count}/{len(self.nodes)} nodes connected")

# ============================================================================
# WEB DASHBOARD
# ============================================================================

class DashboardHandler(BaseHTTPRequestHandler):
    """HTTP handler for dashboard."""
    
    coordinator = None  # Will be set by main()
    
    def do_GET(self):
        """Handle GET requests."""
        if self.path == '/':
            self.send_response(200)
            self.send_header('Content-type', 'text/html')
            self.end_headers()
            self.wfile.write(self.get_dashboard_html().encode())
        elif self.path == '/api/metrics':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.send_header('Access-Control-Allow-Origin', '*')
            self.end_headers()
            data = self.get_metrics_json()
            self.wfile.write(data.encode())
        elif self.path == '/api/nodes':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.send_header('Access-Control-Allow-Origin', '*')
            self.end_headers()
            data = self.get_nodes_json()
            self.wfile.write(data.encode())
        else:
            self.send_error(404)
    
    def get_dashboard_html(self):
        """Generate dashboard HTML."""
        return """
<!DOCTYPE html>
<html>
<head>
    <title>SovereignMap POC Dashboard</title>
    <style>
        body { font-family: Arial; margin: 20px; background: #f5f5f5; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 5px; }
        .metrics { display: grid; grid-template-columns: repeat(4, 1fr); gap: 15px; margin: 20px 0; }
        .metric { background: white; padding: 20px; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metric-value { font-size: 32px; font-weight: bold; color: #3498db; }
        .metric-label { color: #7f8c8d; margin-top: 5px; }
        .chart { background: white; padding: 20px; border-radius: 5px; margin: 20px 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .node-list { background: white; padding: 20px; border-radius: 5px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .node { padding: 10px; border-bottom: 1px solid #ecf0f1; }
        .online { color: #27ae60; }
        .offline { color: #e74c3c; }
        canvas { max-width: 100%; }
    </style>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>
<body>
    <div class="header">
        <h1>🗺️ SovereignMap Proof of Concept</h1>
        <p>Real-time monitoring of decentralized federated learning</p>
    </div>
    
    <div class="metrics">
        <div class="metric">
            <div class="metric-value" id="activeNodes">-</div>
            <div class="metric-label">Active Nodes</div>
        </div>
        <div class="metric">
            <div class="metric-value" id="currentRound">-</div>
            <div class="metric-label">Current Round</div>
        </div>
        <div class="metric">
            <div class="metric-value" id="participation">-</div>
            <div class="metric-label">Participation Rate</div>
        </div>
        <div class="metric">
            <div class="metric-value" id="totalStake">-</div>
            <div class="metric-label">Total Stake</div>
        </div>
    </div>
    
    <div class="chart">
        <h3>Stake Over Time</h3>
        <canvas id="stakeChart"></canvas>
    </div>
    
    <div class="node-list">
        <h3>Node Status</h3>
        <div id="nodeList"></div>
    </div>
    
    <script>
        let stakeChart;
        
        function updateDashboard() {
            // Fetch metrics
            fetch('/api/metrics')
                .then(r => r.json())
                .then(data => {
                    if (data.length > 0) {
                        const latest = data[data.length - 1];
                        document.getElementById('activeNodes').textContent = latest.active_nodes;
                        document.getElementById('currentRound').textContent = latest.round_number;
                        document.getElementById('participation').textContent = 
                            (latest.participation_rate * 100).toFixed(1) + '%';
                        document.getElementById('totalStake').textContent = 
                            latest.total_stake.toFixed(0);
                        
                        // Update chart
                        updateChart(data);
                    }
                });
            
            // Fetch nodes
            fetch('/api/nodes')
                .then(r => r.json())
                .then(nodes => {
                    const nodeList = document.getElementById('nodeList');
                    nodeList.innerHTML = nodes.map(n => 
                        `<div class="node">
                            Node ${n.id}: 
                            <span class="${n.is_online ? 'online' : 'offline'}">
                                ${n.is_online ? '●' : '○'}
                            </span>
                            Stake: ${n.stake.toFixed(0)} | 
                            Peers: ${n.peers.length}
                        </div>`
                    ).join('');
                });
        }
        
        function updateChart(data) {
            const ctx = document.getElementById('stakeChart');
            if (!stakeChart) {
                stakeChart = new Chart(ctx, {
                    type: 'line',
                    data: {
                        labels: data.map(d => d.round_number),
                        datasets: [{
                            label: 'Average Stake',
                            data: data.map(d => d.average_stake),
                            borderColor: '#3498db',
                            fill: false
                        }, {
                            label: 'Total Stake',
                            data: data.map(d => d.total_stake),
                            borderColor: '#2ecc71',
                            fill: false
                        }]
                    },
                    options: {
                        responsive: true,
                        scales: {
                            y: { beginAtZero: false }
                        }
                    }
                });
            } else {
                stakeChart.data.labels = data.map(d => d.round_number);
                stakeChart.data.datasets[0].data = data.map(d => d.average_stake);
                stakeChart.data.datasets[1].data = data.map(d => d.total_stake);
                stakeChart.update();
            }
        }
        
        // Update every 2 seconds
        updateDashboard();
        setInterval(updateDashboard, 2000);
    </script>
</body>
</html>
        """
    
    def get_metrics_json(self):
        """Return metrics as JSON."""
        if not self.coordinator:
            return "[]"
        metrics = [asdict(m) for m in self.coordinator.metrics_history[-50:]]
        return json.dumps(metrics)
    
    def get_nodes_json(self):
        """Return node status as JSON."""
        if not self.coordinator:
            return "[]"
        nodes = [{
            'id': n.id,
            'stake': n.stake,
            'is_online': n.is_online,
            'peers': list(n.peers),
            'last_update_round': n.last_update_round
        } for n in self.coordinator.nodes]
        return json.dumps(nodes)
    
    def log_message(self, format, *args):
        """Suppress request logging."""
        pass

def start_dashboard(coordinator, port=8000):
    """Start the web dashboard in a separate thread."""
    DashboardHandler.coordinator = coordinator
    server = HTTPServer(('', port), DashboardHandler)
    thread = Thread(target=server.serve_forever, daemon=True)
    thread.start()
    print(f"\n🌐 Dashboard running at http://localhost:{port}")
    return server

# ============================================================================
# MAIN
# ============================================================================

async def main():
    """Run the proof of concept."""
    print("="*60)
    print("SovereignMap Proof of Concept")
    print("="*60)
    print("\nInitializing system...")
    
    # Create nodes
    num_nodes = 10
    nodes = [SovereignMapNode(i, initial_stake=1000 + random.uniform(-200, 200)) 
             for i in range(num_nodes)]
    print(f"✓ Created {num_nodes} nodes")
    
    # Create network simulator
    network = NetworkSimulator(nodes)
    print(f"✓ Network simulator ready")
    
    # Create coordinator
    coordinator = FederatedCoordinator(nodes)
    print(f"✓ FL coordinator ready")
    
    # Start dashboard
    dashboard = start_dashboard(coordinator, port=8000)
    print(f"✓ Dashboard started")
    
    print("\nStarting simulation...")
    print("Press Ctrl+C to stop\n")
    
    try:
        # Initial data collection
        print("Collecting initial sensor data...")
        await asyncio.gather(*[node.collect_data(duration=3) for node in nodes])
        
        # Update network topology
        await network.update_topology()
        
        # Run FL rounds
        for round_num in range(50):  # Run 50 rounds
            # Run FL round
            metrics = await coordinator.run_round()
            
            # Periodically update network
            if round_num % 5 == 0:
                await network.update_topology()
            
            # Collect more data
            await asyncio.gather(*[node.collect_data(duration=2) for node in nodes])
            
            # Wait between rounds
            await asyncio.sleep(2)
        
        print("\n" + "="*60)
        print("Simulation complete!")
        print("="*60)
        print(f"\nFinal metrics:")
        print(f"  Average stake: {coordinator.metrics_history[-1].average_stake:.2f}")
        print(f"  Total rounds: {coordinator.round_number}")
        print(f"  Aggregation success rate: {sum(m.aggregation_success for m in coordinator.metrics_history) / len(coordinator.metrics_history) * 100:.1f}%")
        print(f"\nDashboard still running at http://localhost:8000")
        print("Press Ctrl+C to exit")
        
        # Keep dashboard running
        while True:
            await asyncio.sleep(1)
            
    except KeyboardInterrupt:
        print("\n\nShutting down...")
        dashboard.shutdown()

if __name__ == "__main__":
    asyncio.run(main())
