# WEEK 1 OPTIMIZATION TWEAKS: Recommended Improvements

## Overview

The current implementation is functional but has opportunities for optimization in accuracy, performance, realism, and analysis. Here are all recommended tweaks organized by priority.

---

## 🔴 HIGH PRIORITY TWEAKS

### 1. Adaptive Convergence Threshold (CRITICAL)

**Current Issue:** Convergence check uses fixed `avg_last_5 >= 80.0`
- This is too aggressive for high Byzantine levels
- 80% accuracy is unrealistic for 40-50% Byzantine

**Tweak:**
```python
# Current (bad):
converged = avg_last_5 >= 80.0

# Better:
def check_convergence(self, accuracies, byzantine_pct):
    """Adaptive convergence threshold based on Byzantine level"""
    
    # Theoretical minimum: (1 - byzantine_pct/100) * 100
    # Practical: account for noise, network losses
    
    if byzantine_pct <= 10:
        threshold = 85.0  # Can achieve high accuracy
    elif byzantine_pct <= 20:
        threshold = 80.0  # Slight degradation
    elif byzantine_pct <= 30:
        threshold = 75.0  # Significant degradation
    elif byzantine_pct <= 40:
        threshold = 70.0  # Heavy attacks
    else:
        threshold = 65.0  # Near-critical Byzantine level
    
    avg_last_10 = np.mean(accuracies[-10:])
    return avg_last_10 >= threshold
```

**Impact:** More realistic convergence detection across Byzantine levels

---

### 2. Byzantine Attack Variance (IMPROVES REALISM)

**Current Issue:** Attack magnitudes are fixed
- Sign-flip is always exactly -w (too deterministic)
- Label-flip is always exactly w * -1.5 (too deterministic)

**Tweak:**
```python
# Current (deterministic):
def sign_flip(w):
    return -w

# Better (realistic):
@staticmethod
def sign_flip(w, variance=0.1):
    """Sign-flip with variance to simulate imperfect attacks"""
    noise = np.random.randn(*w.shape) * variance
    return -w + noise  # Slightly imperfect attack

@staticmethod
def label_flip(w, scale=1.5, variance=0.2):
    """Label-flip with configurable intensity"""
    noise = np.random.randn(*w.shape) * variance
    return -w * scale + noise

@staticmethod
def amplification(w, scale=2.5, variance=0.15):
    """Amplification with variance"""
    noise = np.random.randn(*w.shape) * variance
    return w * scale + noise
```

**Impact:** Attacks become less predictable, more realistic

---

### 3. Network Latency Distribution (BETTER MODEL)

**Current Issue:** Latency is uniformly distributed (unrealistic)
- Real networks have bimodal distributions (fast + slow)
- No correlation between nodes

**Tweak:**
```python
# Current (simple):
def deliver(self):
    if random.random() < 0.001:
        self.packet_loss += 1
        return False
    return True

# Better (realistic):
class RealisticNetworkSimulator:
    def __init__(self):
        self.fast_network_prob = 0.9  # 90% of messages fast
        self.slow_network_prob = 0.1  # 10% slow
        self.total_msgs = 0
        self.packet_loss = 0
        self.latency_outliers = 0
    
    def get_latency_ms(self):
        """Bimodal latency distribution"""
        if random.random() < self.fast_network_prob:
            # Fast path: 1-3ms (normal)
            return random.uniform(1, 3) + random.expovariate(1.0)
        else:
            # Slow path: 20-100ms (outliers)
            return random.uniform(20, 100) + random.expovariate(0.1)
    
    def deliver(self):
        """Better network simulation"""
        self.total_msgs += 1
        
        # Packet loss: 0.1%
        if random.random() < 0.001:
            self.packet_loss += 1
            return False
        
        # Latency check with timeout
        latency_ms = self.get_latency_ms()
        if latency_ms > 5000:
            self.latency_outliers += 1
            return False
        
        return True
```

**Impact:** Network model more realistic with outliers

---

### 4. Byzantine Node Correlation (STRONGER ATTACKS)

**Current Issue:** Byzantine attacks are independent
- In reality, Byzantine nodes often coordinate
- Should test correlated vs. independent attacks

**Tweak:**
```python
def run_round(self, round_num, bft_pct, attack_type):
    """Run with Byzantine node coordination"""
    
    nonce = f"r{round_num}"
    verified = 0
    attacked = 0
    
    # Select which nodes will be Byzantine THIS ROUND
    # (stay constant across all updates in the round)
    num_byzantine = int(self.NUM_NODES * bft_pct / 100.0)
    byzantine_nodes = set(np.random.choice(
        self.NUM_NODES,
        size=num_byzantine,
        replace=False
    ))
    
    for node_id in range(self.NUM_NODES):
        w = np.random.randn(100)
        
        # All Byzantine nodes use same attack
        if node_id in byzantine_nodes:
            w = RealisticByzantineAttack.apply(w, attack_type)
            attacked += 1
        
        if self.net.deliver():
            if self.tpm.create_quote(node_id, nonce):
                verified += 1
    
    # ... rest of accuracy calculation
```

**Impact:** Byzantine nodes now coordinate (more dangerous)

---

### 5. Accuracy Model with Resistance (REALISTIC RECOVERY)

**Current Issue:** Accuracy degradation is linear
- No "recovery" even with honest majority
- Should have resilience

**Tweak:**
```python
def calculate_accuracy(self, round_num, attacked, verified, delivery_rate):
    """Improved accuracy model with Byzantine resistance"""
    
    base = 65.0
    honest_pct = 1.0 - (attacked / self.NUM_NODES)
    
    # Improvement: only honest nodes contribute
    improvement = 2.5 * (round_num / self.ROUNDS) * honest_pct
    
    # Byzantine resistance: Krum, median, etc. help
    # If honest > 66%, system can recover faster
    if honest_pct > 2/3:
        byzantine_factor = 0.2  # Light impact
    elif honest_pct > 0.5:
        byzantine_factor = 0.5  # Medium impact
    else:
        byzantine_factor = 1.0  # Heavy impact
    
    attack_impact = (attacked / self.NUM_NODES) * byzantine_factor
    
    # Network and attestation (same as before)
    network_impact = (1.0 - delivery_rate) * 0.3
    boost = (verified / self.NUM_NODES) * 0.2
    
    accuracy = min(99.5, base + improvement - attack_impact - network_impact + boost)
    return accuracy
```

**Impact:** More realistic Byzantine resistance modeling

---

## 🟡 MEDIUM PRIORITY TWEAKS

### 6. Aggregation Method Impact

**Current Issue:** Doesn't test different aggregation methods
- Mean aggregation (vulnerable to Byzantine)
- Median (more robust)
- Krum (Byzantine-robust)
- Multi-Krum (more robust)

**Tweak:**
```python
class AggregationMethods:
    @staticmethod
    def mean(updates):
        """Vulnerable to Byzantine"""
        return np.mean(updates, axis=0)
    
    @staticmethod
    def median(updates):
        """Robust to outliers"""
        return np.median(updates, axis=0)
    
    @staticmethod
    def krum(updates, byzantine_count=None):
        """Byzantine-robust: exclude most distant point"""
        if byzantine_count is None:
            byzantine_count = len(updates) // 10
        
        # Find Krum score for each update
        distances = []
        for i, u in enumerate(updates):
            dist = sum(
                np.linalg.norm(u - u_j)
                for j, u_j in enumerate(updates)
                if i != j
            )
            distances.append(dist)
        
        # Select non-Byzantine updates
        good_indices = sorted(range(len(distances)), key=lambda i: distances[i])
        good_indices = good_indices[:-byzantine_count]
        
        return np.mean([updates[i] for i in good_indices], axis=0)

# Use in test:
for agg_method in ["mean", "median", "krum"]:
    result = self.run_config(agg_method)
```

**Impact:** Tests aggregation robustness

---

### 7. Multi-Round Statistics

**Current Issue:** Only tracks final accuracy
- Doesn't show convergence curve
- Can't see if converged early/late

**Tweak:**
```python
def run_round_with_tracking(self, round_num, bft_pct, attack_type):
    """Track detailed statistics per round"""
    
    result = self.run_round(round_num, bft_pct, attack_type)
    
    return {
        'round': round_num,
        'accuracy': result['acc'],
        'loss': result['loss'],
        'accuracy_improvement': result['acc'] - self.prev_acc if round_num > 1 else 0,
        'convergence_progress': result['acc'] / 80.0,  # % to convergence threshold
        'stability': np.std(self.recent_accs[-5:]) if len(self.recent_accs) > 1 else 0,
    }
```

**Impact:** Better visibility into convergence dynamics

---

### 8. Byzantine Node Persistence

**Current Issue:** Byzantine selection is per-round (can change)
- In reality, Byzantine nodes are fixed
- Attackers don't rotate

**Tweak:**
```python
class Week1BFTDemo:
    def __init__(self, ...):
        # ... other init ...
        self.byzantine_nodes = {}  # Pre-select per config
    
    def select_byzantine_nodes(self, bft_pct):
        """Select and persist Byzantine nodes for this config"""
        num_byzantine = int(self.NUM_NODES * bft_pct / 100.0)
        return set(np.random.choice(
            self.NUM_NODES,
            size=num_byzantine,
            replace=False
        ))
    
    def run_config(self, bft_pct, attack_type):
        """Run full config with persistent Byzantine nodes"""
        byzantine_nodes = self.select_byzantine_nodes(bft_pct)
        
        for round_num in range(1, self.ROUNDS + 1):
            result = self.run_round_with_byzantine(
                round_num, bft_pct, attack_type, byzantine_nodes
            )
```

**Impact:** More realistic Byzantine node model

---

### 9. TPM Overhead Measurement

**Current Issue:** Mock verification (99% success rate)
- Doesn't account for actual TPM latency
- Doesn't measure crypto overhead

**Tweak:**
```python
class TPMNodePoolWithMetrics:
    def __init__(self, num_nodes):
        # ... existing init ...
        self.verification_times_ms = []
        self.signature_times_ms = []
    
    def create_quote(self, node_id, nonce):
        """Measure actual crypto times"""
        import hashlib
        import time as time_module
        
        node = self.nodes[node_id]
        
        # Measure signature time
        start = time_module.time()
        pcr_data = "".join(node['pcrs'].values()).encode()
        quote_data = hashlib.sha256(pcr_data + nonce.encode()).digest()
        
        signature = node['key'].sign(
            quote_data,
            padding.PSS(mgf=padding.MGF1(hashes.SHA256()), salt_length=padding.PSS.MAX_LENGTH),
            hashes.SHA256()
        )
        sig_time_ms = (time_module.time() - start) * 1000
        self.signature_times_ms.append(sig_time_ms)
        
        # Measure verification time
        start = time_module.time()
        try:
            ak_public = node['key'].public_key()
            ak_public.verify(signature, quote_data, padding.PSS(...), hashes.SHA256())
            verified = True
        except:
            verified = False
        verify_time_ms = (time_module.time() - start) * 1000
        self.verification_times_ms.append(verify_time_ms)
        
        self.quotes_created += 1
        if verified:
            self.quotes_verified += 1
        
        return verified
    
    def print_crypto_stats(self):
        """Print TPM overhead"""
        print(f"\nTPM Cryptography Overhead:")
        print(f"  Avg Signature Time: {np.mean(self.signature_times_ms):.2f}ms")
        print(f"  Avg Verification Time: {np.mean(self.verification_times_ms):.2f}ms")
        print(f"  Max Signature Time: {np.max(self.signature_times_ms):.2f}ms")
        print(f"  Max Verification Time: {np.max(self.verification_times_ms):.2f}ms")
```

**Impact:** Quantifies real TPM overhead

---

### 10. Gradient Diversity Metric

**Current Issue:** All gradients are i.i.d. random
- Doesn't model real data distribution
- Can't detect Byzantine via diversity

**Tweak:**
```python
class RealisticGradientGenerator:
    def __init__(self, num_nodes, diversity=0.9):
        """Generate realistic gradients with diversity"""
        self.diversity = diversity
        # Generate common feature across nodes (0.9) + noise (0.1)
        self.base_gradient = np.random.randn(100) * np.sqrt(diversity)
    
    def generate_honest_gradient(self):
        """Honest nodes have similar gradients (same dataset)"""
        noise = np.random.randn(100) * np.sqrt(1 - self.diversity)
        return self.base_gradient + noise
    
    def generate_byzantine_gradient(self, attack_type):
        """Byzantine attack against common gradient"""
        honest_w = self.generate_honest_gradient()
        return RealisticByzantineAttack.apply(honest_w, attack_type)

# Use in test:
gen = RealisticGradientGenerator(num_nodes=75)
for node in nodes:
    if is_byzantine:
        w = gen.generate_byzantine_gradient(attack_type)
    else:
        w = gen.generate_honest_gradient()
```

**Impact:** Gradients now reflect real federated learning patterns

---

## 🟢 LOW PRIORITY TWEAKS

### 11. Heterogeneous Node Performance

**Issue:** All nodes have same computation speed
- Some nodes may be slower (edge devices)
- Affects timeout probability

**Tweak:**
```python
class NodePool:
    def __init__(self):
        self.node_speeds = {}
        for i in range(NUM_NODES):
            # 80% fast, 20% slow
            self.node_speeds[i] = random.choices(
                [1.0, 0.5],  # speed multiplier
                weights=[0.8, 0.2]
            )[0]
    
    def get_latency_for_node(self, node_id):
        """Latency depends on node speed"""
        base_latency = random.uniform(1, 5)
        return base_latency / self.node_speeds[node_id]
```

---

### 12. Node Dropout Simulation

**Issue:** Doesn't test node failures
- Some nodes may be offline
- Should handle gracefully

**Tweak:**
```python
class NetworkWithDropout:
    def __init__(self, dropout_rate=0.01):  # 1% nodes offline
        self.dropout_rate = dropout_rate
    
    def can_node_send(self, node_id):
        """Some nodes may be offline"""
        return random.random() > self.dropout_rate
```

---

### 13. Time-Based Byzantine Attacks

**Issue:** Byzantine attacks are constant
- Real attacks may escalate over time
- Adaptive attacks

**Tweak:**
```python
def get_attack_intensity(self, round_num, attack_type):
    """Attacks may intensify over rounds"""
    
    # Basic: constant
    if round_num < self.ROUNDS / 2:
        return 1.0  # Full intensity
    else:
        return 1.2  # Intensify in later rounds
```

---

### 14. Statistical Significance Testing

**Issue:** Results are stochastic but no confidence intervals
- Need error bars

**Tweak:**
```python
def run_multiple_trials(self, num_trials=3):
    """Run same config multiple times for statistics"""
    results = []
    for trial in range(num_trials):
        result = self.run_config(bft_pct, attack_type)
        results.append(result)
    
    convergence_rate = sum(r['converged'] for r in results) / num_trials
    confidence = 1.96 * np.sqrt(convergence_rate * (1 - convergence_rate) / num_trials)
    
    return convergence_rate, confidence
```

---

### 15. Comparison to Aggregation-Based Defense

**Issue:** No baseline comparison
- How much does Krum/median help?

**Tweak:**
```python
def compare_defenses(self):
    """Compare aggregation defenses"""
    
    configs = [
        ("mean", 0),      # No defense
        ("median", 0),    # Median aggregation
        ("krum", 0),      # Byzantine-robust
        ("multi_krum", 0) # Stronger Krum
    ]
    
    for agg_method, byzantine_pct in configs:
        result = self.run_config_with_aggregation(agg_method, byzantine_pct)
        self.compare_results.append(result)
```

---

## 📊 Summary: Tweak Impact Matrix

| Tweak | Impact | Effort | Priority |
|-------|--------|--------|----------|
| Adaptive Convergence | High | Low | CRITICAL |
| Byzantine Variance | High | Low | HIGH |
| Network Latency | Medium | Medium | HIGH |
| Byzantine Correlation | Medium | Medium | HIGH |
| Accurate Accuracy Model | High | Medium | HIGH |
| Aggregation Methods | High | High | MEDIUM |
| Multi-Round Stats | Medium | Low | MEDIUM |
| Byzantine Persistence | Medium | Low | MEDIUM |
| TPM Overhead | Low | Medium | LOW |
| Gradient Diversity | High | Medium | LOW |
| Node Heterogeneity | Low | Low | LOW |
| Node Dropout | Low | Low | LOW |
| Adaptive Attacks | Low | Medium | LOW |
| Statistical Testing | Low | Medium | LOW |
| Defense Comparison | High | High | LOW |

---

## 🚀 Quick Implementation Guide

### Start with (< 1 hour):
1. Adaptive convergence threshold
2. Byzantine variance
3. Byzantine persistence

### Then add (1-2 hours):
4. Network latency distribution
5. Byzantine correlation
6. Accuracy model with resistance

### Advanced (2-3 hours):
7. Aggregation methods
8. TPM overhead measurement
9. Multi-round statistics

---

**Recommendation:** Implement tweaks 1-6 before moving to Week 2 scalability testing. These will significantly improve realism without massive code changes.
