# WEEK 2 TEST STRUCTURE: REAL DATA, FAILURE MODES & ULTRA-SCALING

## Overview

Week 2 extends Week 1 validation from synthetic to realistic scenarios. We move from pure Byzantine tolerance testing to production-hardened testing with real datasets, failure modes, and ultra-large scale validation.

**Total Tests:** 7
**Estimated Runtime:** 180 seconds (3 minutes)
**Priority Order:** MNIST → Failures → Partitions → Cascading → GPU Profiling → 5000 Nodes

---

## Test 1: MNIST Real Dataset Validation (Priority #1)
**File:** `bft_week2_mnist_validation.py`
**Runtime:** ~40 seconds
**Purpose:** Validate convergence with real image classification data

### Scope
- MNIST dataset (60K training, 10K test images)
- Simulated federated learning (IID and non-IID splits)
- Multi-scale testing (75, 200, 500 nodes)
- Byzantine attacks at 0%, 20%, 50% levels

### Test Matrix
```
Scales:               3 (75, 200, 500 nodes)
Byzantine Levels:     3 (0%, 20%, 50%)
Data Distribution:    2 (IID, non-IID)
Attack Types:         3 (sign-flip, label-flip, poison)
Aggregation Methods:  3 (mean, median, trimmed_mean)
Rounds:               25 (reduced for speed)
Total Configs:        108
```

### Expected Results
- IID convergence: ~95% accuracy
- Non-IID convergence: ~92% accuracy
- Byzantine tolerance: Same as Week 1 (scale-invariant)
- Accuracy loss vs synthetic: <2%

### Key Metrics
- Convergence rate (target: 100%)
- Real model accuracy
- Byzantine tolerance pattern
- Computational overhead
- Memory efficiency

### Output
```
results/week2_mnist_validation_results.txt
├─ Convergence by scale
├─ Accuracy vs Byzantine %
├─ IID vs non-IID comparison
├─ Model accuracy curve
└─ Performance metrics
```

---

## Test 2: Failure Mode Testing (Priority #2)
**File:** `bft_week2_failure_modes.py`
**Runtime:** ~35 seconds
**Purpose:** Test system robustness with node failures

### Failure Scenarios
1. **Random Node Dropout** (1-5% at random times)
   - Tests: Session robustness, message retry
   - Target: >95% convergence despite dropouts

2. **Byzantine Crashes** (Byzantine nodes abort mid-round)
   - Tests: Byzantine resilience when nodes fail
   - Target: >90% convergence

3. **Cascading Failures** (5% failure triggers more)
   - Tests: Failure propagation
   - Target: <20% failure cascade

4. **Network Timeouts** (0.1% of rounds timeout)
   - Tests: Timeout handling
   - Target: >98% message delivery

5. **Slow Nodes** (10% of nodes respond 10x slower)
   - Tests: Stragglers handling
   - Target: <5% delay impact on convergence

### Test Matrix
```
Scales:              3 (75, 200, 500 nodes)
Failure Modes:       5 (dropout, crash, cascade, timeout, straggler)
Failure Rates:       3 (1%, 3%, 5%)
Byzantine Levels:    2 (0%, 20%)
Rounds:              20
Total Configs:       90
```

### Expected Results
- Random dropout (1%): 96% convergence, <1% accuracy loss
- Random dropout (5%): 92% convergence, <2% accuracy loss
- Byzantine crashes: 90% convergence, 1-2% accuracy loss
- Network timeouts: 99% delivery, minimal impact
- Slow nodes: 94% convergence, <3% slowdown

### Output
```
results/week2_failure_modes_results.txt
├─ Dropout tolerance (1%, 3%, 5%)
├─ Byzantine crash scenarios
├─ Timeout impact analysis
├─ Straggler effect measurement
└─ Recovery time metrics
```

---

## Test 3: Network Partitions (Priority #3)
**File:** `bft_week2_network_partitions.py`
**Runtime:** ~30 seconds
**Purpose:** Test Byzantine tolerance under network partitions

### Partition Scenarios
1. **Two-Part Partition** (50% on each side, 5% loss between)
   - Tests: Split-brain mitigation
   - Target: Detect partition, minimal divergence

2. **Minority Partition** (10% isolated, 0.5% loss)
   - Tests: Minority recovery
   - Target: Minority remains responsive, majority continues

3. **Geographic Partition** (3 regions, regional loss 1%, cross-region 10%)
   - Tests: Multi-region resilience
   - Target: Cross-region communication sparse but effective

4. **Cascading Partition** (Initial 5%, grows to 20%)
   - Tests: Partition growth handling
   - Target: <30% accuracy loss during partition

### Test Matrix
```
Scales:              2 (200, 500 nodes)
Partition Types:     4 (binary, minority, geographic, cascading)
Byzantine Levels:    2 (0%, 20%)
Partition Duration:  2 (rounds 10-20, rounds 10-30)
Rounds:              30
Total Configs:       64
```

### Expected Results
- Binary partition (50/50): System detects, continues on majority
- Minority partition: Minority diverges slowly, recovers on merge
- Geographic partition: Cross-region slower but converges
- Cascading partition: Triggers failover, <5% final accuracy loss

### Output
```
results/week2_network_partitions_results.txt
├─ Binary partition detection
├─ Minority node recovery
├─ Geographic partition behavior
├─ Cascading partition growth
└─ Partition healing metrics
```

---

## Test 4: Cascading Failures (Priority #4)
**File:** `bft_week2_cascading_failures.py`
**Runtime:** ~25 seconds
**Purpose:** Test failure propagation and system stability

### Cascading Patterns
1. **Avalanche Cascade** (1 failure triggers 5% more)
   - Initial: 5% nodes fail
   - Each round: New failures if >5% already failed
   - Target: Stop cascade at <25%

2. **Threshold Cascade** (Failures accelerate above 20%)
   - Linear growth up to 20%
   - Exponential growth above 20%
   - Target: Detect threshold, prevent exponential phase

3. **Recovery Cascade** (Failed nodes restart after 5 rounds)
   - Tests: Recovery behavior
   - Target: Reintegration successful, no divergence

4. **Byzantine Amplification** (Byzantine nodes trigger honest node failures)
   - Tests: Byzantine-triggered cascades
   - Target: Limit cascade to 20% despite Byzantine amplification

### Test Matrix
```
Scales:              2 (200, 500 nodes)
Cascade Patterns:    4 (avalanche, threshold, recovery, amplification)
Initial Failure %:   3 (1%, 3%, 5%)
Byzantine Level:     2 (0%, 20%)
Rounds:              40
Total Configs:       48
```

### Expected Results
- Avalanche cascade: Limited to 15-20% additional failures
- Threshold cascade: Detected, mitigation effective
- Recovery cascade: 95% successful reintegration
- Byzantine amplification: Contained, <30% final failures

### Output
```
results/week2_cascading_failures_results.txt
├─ Cascade growth metrics
├─ Avalanche progression
├─ Threshold detection
├─ Recovery success rate
└─ Byzantine amplification limit
```

---

## Test 5: GPU Profiling (Priority #5)
**File:** `bft_week2_gpu_profiling.py`
**Runtime:** ~20 seconds
**Purpose:** Identify GPU acceleration opportunities

### Profiling Targets
1. **Aggregation Operations** (Can be GPU-accelerated)
   - Current: CPU trimmed_mean O(n log n)
   - GPU potential: CuPy array operations
   - Expected speedup: 5-10x on aggregation

2. **Gradient Operations** (Matrix multiply heavy)
   - Current: NumPy matrix ops
   - GPU potential: PyTorch/TensorFlow
   - Expected speedup: 10-50x on large gradients

3. **Byzantine Detection** (Distance computations)
   - Current: CPU norm calculations
   - GPU potential: Batch norm computation
   - Expected speedup: 3-5x

4. **Network Simulation** (Embarrassingly parallel)
   - Current: Sequential delivery checks
   - GPU potential: Parallel delivery simulation
   - Expected speedup: 2-3x

### Profile Matrix
```
Operations:          4 (aggregation, gradients, detection, network)
Data Sizes:          3 (small 50D, medium 1000D, large 10000D)
Batch Sizes:         3 (75 nodes, 200 nodes, 500 nodes)
GPU Availability:    2 (with GPU, CPU-only fallback)
Iterations:          100
Total Profiles:      72
```

### Expected Results
- Aggregation: 5-10x speedup with GPU
- Gradients (large): 10-50x with GPU
- Detection: 3-5x with GPU
- Overall potential: 3-8x system speedup

### Output
```
results/week2_gpu_profiling_results.txt
├─ Operation timing breakdown (CPU vs GPU)
├─ Memory usage comparison
├─ Bandwidth utilization
├─ Speedup analysis
└─ Recommendations for acceleration
```

---

## Test 6: Ultra-Large Scale (Priority #6)
**File:** `bft_week2_5000_node_scaling.py`
**Runtime:** ~25 seconds
**Purpose:** Validate scaling to 5000+ nodes

### Optimization Strategy
1. **Reduce Rounds** (20 → 10) - 50% speedup
2. **Sample Aggregation** (All nodes → Random sample of 500)
   - 10x aggregation speedup
   - Byzantine tolerance maintained if >50% sample honest
3. **Hierarchical Aggregation** (500 → 50 leaders → global)
   - O(1) per level instead of O(n)

### Test Matrix
```
Scales:              2 (2500, 5000 nodes)
Byzantine Levels:    3 (0%, 20%, 50%)
Aggregation:         3 (full, sampled 500, hierarchical)
Rounds:              10
Total Configs:       18
```

### Expected Results
- 2500N with sampling: ~5-8 seconds
- 5000N with sampling: ~10-15 seconds
- Accuracy loss vs 1000N: <2%
- Byzantine tolerance: Maintained >50%

### Output
```
results/week2_5000_node_scaling_results.txt
├─ Execution time by method
├─ Sampling impact on accuracy
├─ Hierarchical aggregation performance
├─ Byzantine tolerance at scale
└─ Recommended production configuration
```

---

## Test 7: Comprehensive Production Readiness (Priority #7)
**File:** `bft_week2_production_readiness.py`
**Runtime:** ~5 seconds (aggregate results)
**Purpose:** Final validation checklist

### Checks
1. **Real Data Convergence** (MNIST: >90% accuracy)
2. **Failure Resilience** (>95% convergence despite 5% failures)
3. **Byzantine Tolerance** (>50% Byzantine level working)
4. **Scale Validation** (1000-5000 nodes supported)
5. **Network Robustness** (Partitions handled)
6. **Performance** (<30 seconds for 1000 nodes)
7. **Memory Efficiency** (<1 GB for 1000 nodes)

### Pass Criteria
```
✓ All tests must pass with confidence >95%
✓ No single failure >10% accuracy loss
✓ Convergence rate >90% in all scenarios
✓ Linear scaling maintained to 5000 nodes
✓ Byzantine tolerance scale-invariant
```

### Output
```
PRODUCTION_READINESS_REPORT.md
├─ Test results summary
├─ Pass/fail on each criterion
├─ Risk assessment
├─ Deployment recommendations
└─ Known limitations
```

---

## Test Execution Order

### Fast Path (40 minutes, priority tests only)
```
1. bft_week2_mnist_validation.py           (40s)
2. bft_week2_failure_modes.py              (35s)
3. bft_week2_production_readiness.py       (5s)
Total: 80 seconds
```

### Full Path (180 minutes, all tests)
```
1. bft_week2_mnist_validation.py           (40s)  → Real data validation
2. bft_week2_failure_modes.py              (35s)  → Failure resilience
3. bft_week2_network_partitions.py         (30s)  → Network robustness
4. bft_week2_cascading_failures.py         (25s)  → Failure propagation
5. bft_week2_gpu_profiling.py              (20s)  → Acceleration potential
6. bft_week2_5000_node_scaling.py          (25s)  → Ultra-scale
7. bft_week2_production_readiness.py       (5s)   → Final report
Total: ~180 seconds (3 minutes)
```

---

## Results Organization

All results stored in:
```
Sovereign_Map_Federated_Learning/
├─ results/
│  ├─ week2_mnist_validation_results.txt
│  ├─ week2_failure_modes_results.txt
│  ├─ week2_network_partitions_results.txt
│  ├─ week2_cascading_failures_results.txt
│  ├─ week2_gpu_profiling_results.txt
│  ├─ week2_5000_node_scaling_results.txt
│  └─ PRODUCTION_READINESS_REPORT.md
├─ plots/
│  ├─ mnist_convergence_curves.png
│  ├─ failure_mode_analysis.png
│  ├─ partition_recovery.png
│  ├─ cascading_failure_growth.png
│  ├─ gpu_speedup_comparison.png
│  └─ scaling_efficiency_5000.png
└─ WEEK2_FINAL_REPORT.md (consolidated)
```

---

## Dependencies

### Required Packages
```
numpy               (already have)
matplotlib          (for plots)
scikit-learn        (for datasets, metrics)
tensorflow/torch    (optional, for GPU testing)
cupy                (optional, for GPU acceleration)
```

### Dataset Requirements
- MNIST: Auto-downloaded on first run (~50 MB)
- Cache: ~/.keras/datasets/mnist.npz

---

## Success Criteria (Week 2)

| Criterion | Target | Weight |
|-----------|--------|--------|
| MNIST convergence | >90% | High |
| Failure resilience (5%) | >90% convergence | High |
| Byzantine tolerance | >50% maintained | High |
| Network partitions | Detected & handled | Medium |
| Cascading failures | Contained <30% | Medium |
| GPU acceleration | 5-10x on aggregation | Low |
| 5000 node scaling | <20 seconds | Medium |

---

## Next Steps After Week 2

### If All Tests Pass ✓
- Production deployment ready
- Recommend 1000-2000 node initial deployment
- Monitor Byzantine tolerance in real-world attacks
- Plan GPU acceleration phase (Week 3)

### If Tests Show Gaps
- Identify bottleneck (Byzantine tolerance, failures, scaling)
- Implement targeted fix
- Re-run relevant subset of tests
- Document lessons learned

### Week 3 Candidates
1. GPU acceleration implementation
2. Byzantine attack sophistication
3. Real multi-region deployment
4. Production monitoring/alerting
5. Machine learning optimization

---

## Key Differences from Week 1

| Aspect | Week 1 | Week 2 |
|--------|--------|--------|
| Data | Synthetic gradients | Real (MNIST) |
| Failures | None | 5 failure modes |
| Network | Simplified | Partitions tested |
| Scale | 75-1000 | 1000-5000 |
| Focus | Linear scaling | Production readiness |
| Time | 37.5s | ~180s |
| Tests | 1 | 7 |

---

## File Dependencies

```
Week 2 tests depend on Week 1:
├─ bft_aggressive_scaling.py (reference)
├─ OptimizedAggregation class (reuse)
├─ FastGradientGenerator (adapt for MNIST)
└─ FastNetworkSim (extend with partitions)

Each Week 2 test is independent:
├─ bft_week2_mnist_validation.py (standalone)
├─ bft_week2_failure_modes.py (standalone)
├─ bft_week2_network_partitions.py (standalone)
├─ bft_week2_cascading_failures.py (standalone)
├─ bft_week2_gpu_profiling.py (standalone)
├─ bft_week2_5000_node_scaling.py (standalone)
└─ bft_week2_production_readiness.py (aggregates all)
```

---

## Troubleshooting

### If MNIST download fails
- Check internet connection
- Manual download: http://yann.lecun.com/exdb/mnist/
- Set cache directory: `KERAS_HOME=/path/to/cache`

### If failure mode test runs too fast/slow
- Adjust `rounds` parameter (default 20)
- Increase/decrease `num_nodes` (default 75, 200, 500)

### If GPU profiling shows no GPU
- Check CUDA installation: `nvidia-smi`
- Install cupy: `pip install cupy-cuda11x`
- Fallback to CPU-only profiling

### If 5000 node test times out
- Reduce rounds: `rounds=5`
- Enable sampling: `use_sampling=True`
- Enable hierarchical: `use_hierarchical=True`

---

## Support & Documentation

- **Architecture:** See Week 1 code structure
- **Byzantine Attacks:** BFT_ATTACK_FEB_2026.md
- **Performance:** WEEK1_AGGRESSIVE_SCALING_REPORT.md
- **Results Format:** Results stored in plaintext + PNG plots

---

## Final Notes

**Week 2 is about confidence building.**

- Week 1 proved: "Can we scale?"
- Week 2 proves: "Can we handle reality?"

Real data, real failures, real partitions. If all tests pass, deployment is low-risk.
