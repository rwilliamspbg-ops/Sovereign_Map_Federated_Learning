# Test Results Directory

**Central repository for all test execution results, benchmarks, and artifacts.**

---

## 📂 Directory Structure

```
test-results/
├── 1000-node-npu/              # 1000-node NPU performance test results
│   ├── 20260304-103652/        # Test run from 2026-03-04
│   └── README.md
├── kubernetes-5000-node/       # 5000-node Kubernetes stress test results
│   ├── k8s-5000-node-20260303-052718.json
│   └── plots/                  # Performance visualization charts
├── byzantine-stress-test/      # Byzantine fault tolerance test results
├── byzantine-stress-test-suite/ # Comprehensive Byzantine test suite results
├── demo-simulated/             # Simulated demo execution results
├── tpm-npu-full/               # TPM + NPU integrated test results
├── round200_live/              # 200-round live test results
└── [timestamped directories]   # Historical test runs
```

---

## 🔥 Featured Test Results

### 1. 1000-Node NPU Performance Test

**Directory:** [`1000-node-npu/20260304-103652/`](1000-node-npu/20260304-103652/)  
**Date:** 2026-03-04  
**Status:** ✅ COMPLETE  

**Key Metrics:**
- **Throughput:** 2,850 RPS (4.38x improvement over CPU baseline)
- **Latency Reduction:** 66.9%
- **Test Duration:** 22 minutes 22 seconds
- **Nodes:** 1,000

**Files:**
- [RESULTS.md](1000-node-npu/20260304-103652/RESULTS.md) - Detailed results and analysis

**Documentation:**
- [📊 Full Report](/Documentation/Testing/1000-NODE-NPU-TEST-FINAL-SUMMARY.md)
- [📖 Test Guide](/Documentation/Testing/1000-NODE-NPU-TEST-GUIDE.md)

### 2. 5000-Node Kubernetes Stress Test

**Directory:** [`kubernetes-5000-node/`](kubernetes-5000-node/)  
**Date:** 2026-03-03  
**Status:** ✅ ALL 4 SCENARIOS PASSED  

**Key Metrics:**
- **Nodes:** 5,000
- **Byzantine Tolerance:** 50% (2,500 malicious nodes)
- **Accuracy:** 86.00%
- **Test Duration:** 2.8 seconds
- **Scenarios:** 4/4 passed

**Files:**
- [k8s-5000-node-20260303-052718.json](kubernetes-5000-node/k8s-5000-node-20260303-052718.json) - Raw test data
- [plots/master-summary.png](kubernetes-5000-node/plots/master-summary.png) - Overview chart
- [plots/scenario-1-5000node.png](kubernetes-5000-node/plots/scenario-1-5000node.png) - 5000-node stress
- [plots/scenario-2-scaling.png](kubernetes-5000-node/plots/scenario-2-scaling.png) - Scaling analysis
- [plots/scenario-3-threshold.png](kubernetes-5000-node/plots/scenario-3-threshold.png) - Byzantine threshold
- [plots/scenario-4-intensity.png](kubernetes-5000-node/plots/scenario-4-intensity.png) - Attack intensity

**Documentation:**
- [📊 Full Report](/Documentation/Deployment/KUBERNETES_5000_NODE_REPORT.md)
- [📖 Session Report](/Documentation/Reports/SESSION_KUBERNETES_5000_NODE_COMPLETE.md)

---

## 📊 Additional Test Results

### Byzantine Stress Tests

**Directories:**
- [`byzantine-stress-test/`](byzantine-stress-test/) - Initial Byzantine stress test
- [`byzantine-stress-test-suite/`](byzantine-stress-test-suite/) - Comprehensive test suite

**Documentation:**
- [📊 Byzantine Stress Test Report](/Documentation/Testing/BYZANTINE_STRESS_TEST_REPORT.md)
- [📊 Byzantine Test Suite Report](/Documentation/Testing/BYZANTINE_STRESS_TEST_SUITE_REPORT.md)

### GPU Acceleration Tests

**Files in root directory:**
- `test-results/benchmarks/gpu-benchmark-baseline.json` - GPU baseline performance
- `test-results/benchmarks/gpu-contention-5nodes.json` through `test-results/benchmarks/gpu-contention-30nodes.json` - Contention analysis
- `test-results/benchmarks/gpu-round-5nodes.json` through `test-results/benchmarks/gpu-round-20nodes.json` - Round performance data

**Documentation:**
- [📊 GPU Testing Complete](/Documentation/Testing/GPU_TESTING_COMPLETE.md)
- [📊 GPU Testing Results Report](/Documentation/Testing/GPU_TESTING_RESULTS_REPORT.md)
- [📊 GPU Validation Complete](/Documentation/Testing/GPU_VALIDATION_COMPLETE.md)

### Demo & Simulation Results

**Directory:** [`demo-simulated/`](demo-simulated/)  

**Documentation:**
- [📊 Complete Demo Data](/Documentation/Reports/COMPLETE_DEMO_DATA_VIEWABLE.md)
- [📊 Demo Results Summary](/Documentation/Reports/DEMO_RESULTS_SUMMARY.txt)
- [📖 Demo Results README](/Documentation/Reports/README_DEMO_RESULTS.md)

### TPM + NPU Integrated Tests

**Directory:** [`tpm-npu-full/`](tpm-npu-full/)  
**Archive:** `tpm-npu-full-artifacts.tar.gz`

**Documentation:**
- [📊 Grafana NPU Laptop Final](/Documentation/Deployment/GRAFANA_NPU_LAPTOP_FINAL.md)

### Large-Scale Historical Tests

**Directories:**
- [`20260227T004204Z_200round_fullscope/`](20260227T004204Z_200round_fullscope/) - 200-round full scope
- [`incremental_scale_test_full_20260228_122530/`](incremental_scale_test_full_20260228_122530/) - Incremental scaling
- [`scale_test_5000r_100burst_20260228_123024/`](scale_test_5000r_100burst_20260228_123024/) - 5000-round burst test
- [`round200_live/`](round200_live/) - 200-round live execution

---

## 🔍 Finding Specific Results

### By Test Type

```bash
# All 1000-node NPU results
ls -la 1000-node-npu/

# All 5000-node K8s results
ls -la kubernetes-5000-node/

# All Byzantine test results
ls -la byzantine-stress-test*/

# All GPU test results (in root)
ls -la ../*.json | grep gpu
```

### By Date

```bash
# List all timestamped test runs
ls -lt | grep "^d"

# Find tests from specific date (e.g., 2026-03-04)
find . -type d -name "*20260304*"
```

### By File Type

```bash
# All JSON result files
find . -name "*.json"

# All PNG visualization charts
find . -name "*.png"

# All markdown reports
find . -name "*.md"
```

---

## 📈 Metrics & Data Files

### Real-Time Metrics (Root Directory)

- `metrics-1771560200-accuracy.json` - Accuracy metrics
- `metrics-1771560200-byzantine.json` - Byzantine metrics
- `metrics-1771560200-clients.json` - Client metrics
- `metrics-1771560200-rounds.json` - Round execution metrics
- `raw_metrics.json` - Raw metric data

---

## 📚 Related Documentation

- **[Complete Artifacts Index](/Documentation/Reports/ARTIFACTS.md)** - Full catalog of all test artifacts
- **[Testing Guide](/Documentation/Testing/TEST_GUIDE.md)** - How to run tests
- **[Test Completion Index](/Documentation/Testing/TESTING_COMPLETION_INDEX.md)** - Testing status tracker
- **[Architecture Documentation](/Documentation/Architecture/ARCHITECTURE.md)** - System architecture

---

## 🚀 Running Tests

### Quick Test Commands

```bash
# Run 1000-node NPU test
./run-1000-node-npu-test.sh  # Linux/Mac
./run-1000-node-npu-test.ps1  # Windows

# Run 5000-node K8s test
python tests/scripts/python/kubernetes-5000-node-test.py
python generate-k8s-5000-node-plots.py

# Run GPU test suite
python tests/scripts/python/gpu-test-suite.py

# Run Byzantine stress test suite
python tests/scripts/python/byzantine-stress-test-suite.py
```

### Generate Visualizations

```bash
# Generate K8s plots
python generate-k8s-5000-node-plots.py

# Generate Byzantine plots
python tests/scripts/python/generate-byzantine-plots.py
python tests/scripts/python/generate-byzantine-test-suite-plots.py

# Analyze demo results
python analyze-demo-results.py

# Analyze GPU results
python analyze-gpu-results.py
```

---

**Last Updated:** 2026-03-07  
**For complete test artifact catalog, see:** [ARTIFACTS.md](/Documentation/Reports/ARTIFACTS.md)

