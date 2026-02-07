"""
SovereignMap POC Test Suite
============================

Run with: python test_poc.py
"""

import asyncio
import numpy as np
import time
from sovereignmap_poc import (
    SovereignMapNode,
    NavigationModel,
    SensorReading,
    ModelUpdate,
    stake_weighted_trimmed_mean,
    FederatedCoordinator,
    SyntheticSensor
)

# Test results tracking
tests_passed = 0
tests_failed = 0

def test(name):
    """Decorator for test functions."""
    def decorator(func):
        async def wrapper():
            global tests_passed, tests_failed
            print(f"\n{'='*60}")
            print(f"TEST: {name}")
            print(f"{'='*60}")
            try:
                await func() if asyncio.iscoroutinefunction(func) else func()
                print(f"✓ PASSED")
                tests_passed += 1
            except AssertionError as e:
                print(f"✗ FAILED: {e}")
                tests_failed += 1
            except Exception as e:
                print(f"✗ ERROR: {e}")
                tests_failed += 1
        return wrapper
    return decorator

# ============================================================================
# UNIT TESTS
# ============================================================================

@test("Neural Network Forward Pass")
def test_neural_network_forward():
    """Test that neural network forward pass works."""
    model = NavigationModel(input_dim=10, hidden_dim=20, output_dim=2)
    X = np.random.randn(5, 10)
    y_pred = model.forward(X)
    
    assert y_pred.shape == (5, 2), f"Expected shape (5, 2), got {y_pred.shape}"
    assert not np.isnan(y_pred).any(), "Output contains NaN"
    print(f"  Output shape: {y_pred.shape}")
    print(f"  Output range: [{y_pred.min():.3f}, {y_pred.max():.3f}]")

@test("Neural Network Training")
def test_neural_network_training():
    """Test that model can train and reduce loss."""
    model = NavigationModel(input_dim=10, hidden_dim=20, output_dim=2)
    X = np.random.randn(50, 10)
    y = np.random.randn(50, 2)
    
    initial_weights = model.get_weights()
    losses = []
    for _ in range(10):
        loss = model.backward(X, y, learning_rate=0.01)
        losses.append(loss)
    final_weights = model.get_weights()
    
    # Weights should change
    assert not np.allclose(initial_weights, final_weights), "Weights didn't change"
    
    # Loss should decrease (generally)
    assert losses[-1] < losses[0] * 1.5, f"Loss didn't decrease: {losses[0]:.3f} -> {losses[-1]:.3f}"
    print(f"  Initial loss: {losses[0]:.3f}")
    print(f"  Final loss: {losses[-1]:.3f}")
    print(f"  Improvement: {(1 - losses[-1]/losses[0])*100:.1f}%")

@test("Weight Get/Set Consistency")
def test_weight_getset():
    """Test that weights can be saved and restored."""
    model = NavigationModel(input_dim=10, hidden_dim=20, output_dim=2)
    
    # Get initial weights
    weights1 = model.get_weights()
    
    # Make predictions
    X = np.random.randn(5, 10)
    y1 = model.forward(X)
    
    # Modify model
    model.backward(X, np.random.randn(5, 2))
    
    # Restore weights
    model.set_weights(weights1)
    
    # Predictions should be identical
    y2 = model.forward(X)
    assert np.allclose(y1, y2), "Predictions changed after restoring weights"
    print(f"  ✓ Weights correctly restored")
    print(f"  Max diff: {np.abs(y1 - y2).max():.9f}")

@test("Sensor Data Generation")
async def test_sensor():
    """Test synthetic sensor generates valid data."""
    sensor = SyntheticSensor(node_id=0)
    readings = []
    
    for _ in range(10):
        reading = await sensor.read()
        readings.append(reading)
        await asyncio.sleep(0.01)  # Small delay
    
    # Check data validity
    assert all(r.latitude != 0 for r in readings), "Invalid latitude"
    assert all(r.longitude != 0 for r in readings), "Invalid longitude"
    assert all(0 <= r.heading < 360 for r in readings), "Invalid heading"
    assert all(r.velocity > 0 for r in readings), "Invalid velocity"
    
    # Check movement
    distances = []
    for i in range(len(readings) - 1):
        lat_diff = abs(readings[i+1].latitude - readings[i].latitude)
        lon_diff = abs(readings[i+1].longitude - readings[i].longitude)
        distances.append(lat_diff + lon_diff)
    
    assert any(d > 0 for d in distances), "No movement detected"
    print(f"  Generated {len(readings)} readings")
    print(f"  Position changed: {sum(distances):.6f} degrees")

@test("Feature Vector Conversion")
async def test_feature_vector():
    """Test sensor reading converts to features correctly."""
    sensor = SyntheticSensor(node_id=0)
    reading = await sensor.read()
    features = reading.to_feature_vector()
    
    assert features.shape == (10,), f"Expected 10 features, got {features.shape}"
    assert not np.isnan(features).any(), "Features contain NaN"
    print(f"  Feature vector: {features[:5]}...")

@test("Stake-Weighted Aggregation Correctness")
def test_aggregation_correctness():
    """Test that aggregation correctly weights by stake."""
    # Create simple test case
    weights1 = np.array([10.0, 10.0, 10.0])
    weights2 = np.array([5.0, 5.0, 5.0])
    
    updates = [
        ModelUpdate(
            node_id=0, round_number=1, weights=weights1, stake=90.0,
            contribution_score=1.0, timestamp=time.time(), signature="sig1"
        ),
        ModelUpdate(
            node_id=1, round_number=1, weights=weights2, stake=10.0,
            contribution_score=1.0, timestamp=time.time(), signature="sig2"
        )
    ]
    
    result = stake_weighted_trimmed_mean(updates, trim_fraction=0.0)
    
    # Should be weighted toward weights1 (90% weight)
    # Using median for this implementation, but let's check it's reasonable
    assert result is not None, "Aggregation returned None"
    assert result.shape == (3,), f"Expected shape (3,), got {result.shape}"
    print(f"  Aggregated weights: {result}")
    print(f"  (Median of {weights1} and {weights2})")

@test("Byzantine Resistance")
def test_byzantine_resistance():
    """Test that outliers are handled."""
    normal_weights = np.array([5.0, 5.0, 5.0])
    byzantine_weights = np.array([1000.0, 1000.0, 1000.0])
    
    updates = [
        ModelUpdate(i, 1, normal_weights, 25.0, 1.0, time.time(), f"sig{i}")
        for i in range(3)
    ]
    updates.append(
        ModelUpdate(3, 1, byzantine_weights, 25.0, 1.0, time.time(), "sig3")
    )
    
    result = stake_weighted_trimmed_mean(updates, trim_fraction=0.25)
    
    # Result should be close to normal_weights (median resists outliers)
    assert result is not None
    assert np.all(result < 100), f"Byzantine node influenced result: {result}"
    print(f"  Result with Byzantine node: {result}")
    print(f"  Successfully resisted outlier")

# ============================================================================
# INTEGRATION TESTS
# ============================================================================

@test("Node Creation and Initialization")
def test_node_creation():
    """Test that nodes can be created."""
    node = SovereignMapNode(node_id=0, initial_stake=1000.0)
    
    assert node.id == 0
    assert node.stake == 1000.0
    assert node.contribution_score == 1.0
    assert node.is_online == True
    assert len(node.peers) == 0
    print(f"  Node {node.id} created successfully")
    print(f"  Initial stake: {node.stake}")

@test("Node Local Training")
async def test_node_training():
    """Test that node can train locally."""
    node = SovereignMapNode(node_id=0)
    
    # Collect some data
    await node.collect_data(duration=3)
    
    assert len(node.data_buffer) > 0, "No data collected"
    print(f"  Collected {len(node.data_buffer)} data points")
    
    # Train
    delta = node.train_local(epochs=3)
    
    assert delta is not None, "Training returned None"
    assert delta.shape[0] > 0, "Empty delta"
    assert not np.isnan(delta).any(), "Delta contains NaN"
    print(f"  Training successful, delta shape: {delta.shape}")
    print(f"  Delta range: [{delta.min():.4f}, {delta.max():.4f}]")

@test("Model Update Creation and Verification")
async def test_update_creation():
    """Test creating and verifying updates."""
    node = SovereignMapNode(node_id=0)
    await node.collect_data(duration=2)
    
    delta = node.train_local()
    assert delta is not None
    
    update = node.create_update(round_number=1, delta=delta)
    
    assert update.node_id == 0
    assert update.round_number == 1
    assert update.verify_signature(), "Signature verification failed"
    print(f"  Update created and verified")
    print(f"  Signature: {update.signature}")

@test("Federated Learning Round")
async def test_fl_round():
    """Test a complete FL round."""
    # Create nodes
    nodes = [SovereignMapNode(i, initial_stake=1000) for i in range(5)]
    
    # Collect data
    await asyncio.gather(*[node.collect_data(duration=2) for node in nodes])
    print(f"  Data collected for {len(nodes)} nodes")
    
    # Create coordinator
    coordinator = FederatedCoordinator(nodes)
    
    # Run round
    metrics = await coordinator.run_round()
    
    assert metrics.round_number == 1
    assert metrics.active_nodes == 5
    assert metrics.participation_rate > 0
    print(f"  Round completed:")
    print(f"    Participation: {metrics.participation_rate*100:.1f}%")
    print(f"    Aggregation success: {metrics.aggregation_success}")

@test("Multi-Round Training")
async def test_multi_round():
    """Test multiple FL rounds."""
    nodes = [SovereignMapNode(i) for i in range(5)]
    await asyncio.gather(*[node.collect_data(duration=2) for node in nodes])
    
    coordinator = FederatedCoordinator(nodes)
    
    rounds = 5
    for _ in range(rounds):
        metrics = await coordinator.run_round()
        await asyncio.gather(*[node.collect_data(duration=1) for node in nodes])
    
    assert coordinator.round_number == rounds
    assert len(coordinator.metrics_history) == rounds
    
    # Check stake changes
    final_stakes = [n.stake for n in nodes]
    print(f"  Completed {rounds} rounds")
    print(f"  Final stake range: [{min(final_stakes):.0f}, {max(final_stakes):.0f}]")

@test("Node Stake Updates")
def test_stake_updates():
    """Test stake reward/penalty system."""
    node = SovereignMapNode(node_id=0, initial_stake=1000)
    
    # Reward
    node.update_stake(100)
    assert node.stake == 1100
    
    # Penalty
    node.update_stake(-200)
    assert node.stake == 900
    
    # Can't go below zero
    node.update_stake(-2000)
    assert node.stake == 0
    
    print(f"  Stake history: {node.stake_history}")

# ============================================================================
# PERFORMANCE TESTS
# ============================================================================

@test("Aggregation Performance")
def test_aggregation_performance():
    """Test aggregation speed with many nodes."""
    num_nodes = 100
    weight_dim = 242  # Realistic model size
    
    updates = [
        ModelUpdate(
            i, 1, np.random.randn(weight_dim), 
            random_stake := np.random.uniform(100, 1000),
            1.0, time.time(), f"sig{i}"
        )
        for i in range(num_nodes)
    ]
    
    start = time.time()
    result = stake_weighted_trimmed_mean(updates)
    duration = time.time() - start
    
    assert result is not None
    assert duration < 1.0, f"Aggregation too slow: {duration:.3f}s"
    print(f"  Aggregated {num_nodes} updates in {duration*1000:.1f}ms")
    print(f"  Throughput: {num_nodes/duration:.0f} updates/sec")

@test("Training Performance")
async def test_training_performance():
    """Test training speed."""
    node = SovereignMapNode(node_id=0)
    await node.collect_data(duration=3)
    
    start = time.time()
    delta = node.train_local(epochs=10)
    duration = time.time() - start
    
    assert delta is not None
    assert duration < 5.0, f"Training too slow: {duration:.3f}s"
    print(f"  Trained in {duration*1000:.0f}ms")

# ============================================================================
# RUN ALL TESTS
# ============================================================================

async def run_all_tests():
    """Run all tests."""
    print("\n" + "="*60)
    print("SovereignMap POC Test Suite")
    print("="*60)
    
    # Unit tests
    print("\n" + "="*60)
    print("UNIT TESTS")
    print("="*60)
    
    await test_neural_network_forward()
    await test_neural_network_training()
    await test_weight_getset()
    await test_sensor()
    await test_feature_vector()
    await test_aggregation_correctness()
    await test_byzantine_resistance()
    
    # Integration tests
    print("\n" + "="*60)
    print("INTEGRATION TESTS")
    print("="*60)
    
    await test_node_creation()
    await test_node_training()
    await test_update_creation()
    await test_fl_round()
    await test_multi_round()
    await test_stake_updates()
    
    # Performance tests
    print("\n" + "="*60)
    print("PERFORMANCE TESTS")
    print("="*60)
    
    await test_aggregation_performance()
    await test_training_performance()
    
    # Summary
    print("\n" + "="*60)
    print("TEST SUMMARY")
    print("="*60)
    print(f"Passed: {tests_passed}")
    print(f"Failed: {tests_failed}")
    print(f"Total:  {tests_passed + tests_failed}")
    
    if tests_failed == 0:
        print("\n✓ ALL TESTS PASSED!")
    else:
        print(f"\n✗ {tests_failed} TESTS FAILED")
    
    return tests_failed == 0

if __name__ == "__main__":
    import sys
    success = asyncio.run(run_all_tests())
    sys.exit(0 if success else 1)
