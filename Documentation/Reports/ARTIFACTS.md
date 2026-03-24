# Test Artifacts & Results Index

**Quick navigation to all major test results, benchmarks, and artifacts.**

---

## 🔥 Featured Test Results

### 1000-Node NPU Performance Test ⚡

**Status:** ✅ **COMPLETE & VERIFIED**  
**Date:** 2026-03-04  
**Key Achievement:** 4.38x throughput improvement (650 → 2,850 RPS) with 66.9% latency reduction

#### Documentation
- 📊 **[1000-NODE-NPU-TEST-FINAL-SUMMARY.md](/Documentation/Testing/1000-NODE-NPU-TEST-FINAL-SUMMARY.md)** - Complete test report
- 📖 **[1000-NODE-NPU-TEST-GUIDE.md](/Documentation/Testing/1000-NODE-NPU-TEST-GUIDE.md)** - Reproduction guide
- 🐳 **[docker-compose.full.yml](../../docker-compose.full.yml)** - Infrastructure orchestration

#### Test Results
- 📁 **[test-results/1000-node-npu/20260304-103652/](../../test-results/1000-node-npu/20260304-103652/)** - Raw test data
  - [RESULTS.md](../../test-results/1000-node-npu/20260304-103652/RESULTS.md) - Detailed results

#### Execution Scripts
- 🐧 **[tests/scripts/bash/run-1000-node-npu-test.sh](../../tests/scripts/bash/run-1000-node-npu-test.sh)** - Linux/Mac test runner
- 🪟 **[tests/scripts/powershell/run-1000-node-npu-test.ps1](../../tests/scripts/powershell/run-1000-node-npu-test.ps1)** - Windows PowerShell test runner

---

### 5000-Node Kubernetes Stress Test ☸️

**Status:** ✅ **ALL 4 SCENARIOS PASSED**  
**Date:** 2026-03-03  
**Key Achievement:** 86% accuracy maintained with 50% Byzantine nodes at 5000-node scale

#### Documentation
- 📊 **[KUBERNETES_5000_NODE_REPORT.md](/Documentation/Deployment/KUBERNETES_5000_NODE_REPORT.md)** - Complete test report
- 📖 **[SESSION_KUBERNETES_5000_NODE_COMPLETE.md](/Documentation/Reports/SESSION_KUBERNETES_5000_NODE_COMPLETE.md)** - Session analysis

#### Test Results & Artifacts
- 📁 **[test-results/kubernetes-5000-node/](../../test-results/kubernetes-5000-node/)** - Test data directory
  - 📄 [k8s-5000-node-20260303-052718.json](../../test-results/kubernetes-5000-node/k8s-5000-node-20260303-052718.json) - Raw JSON results
  - 📊 **[plots/](../../test-results/kubernetes-5000-node/plots/)** - Visualization charts
    - [master-summary.png](../../test-results/kubernetes-5000-node/plots/master-summary.png) - Overview chart
    - [scenario-1-5000node.png](../../test-results/kubernetes-5000-node/plots/scenario-1-5000node.png) - 5000-node stress test
    - [scenario-2-scaling.png](../../test-results/kubernetes-5000-node/plots/scenario-2-scaling.png) - Linear scaling validation
    - [scenario-3-threshold.png](../../test-results/kubernetes-5000-node/plots/scenario-3-threshold.png) - Byzantine threshold test
    - [scenario-4-intensity.png](../../test-results/kubernetes-5000-node/plots/scenario-4-intensity.png) - Attack intensity analysis

#### Execution Scripts
- 🐍 **[tests/scripts/python/kubernetes-5000-node-test.py](../../tests/scripts/python/kubernetes-5000-node-test.py)** - Main test suite
- 🐍 **[tests/scripts/python/k8s-5000-node-local-test.py](../../tests/scripts/python/k8s-5000-node-local-test.py)** - Local simulator
- 📈 **[generate-k8s-5000-node-plots.py](../../generate-k8s-5000-node-plots.py)** - Visualization generator
- ☸️ **[kubernetes-5000-node-manifests.yaml](../../kubernetes-5000-node-manifests.yaml)** - K8s deployment manifests

---

## 🧪 Additional Test Artifacts

### Byzantine Stress Tests
- 📊 **[BYZANTINE_STRESS_TEST_REPORT.md](/Documentation/Testing/BYZANTINE_STRESS_TEST_REPORT.md)** - Byzantine fault tolerance validation
- 📊 **[BYZANTINE_STRESS_TEST_SUITE_REPORT.md](/Documentation/Testing/BYZANTINE_STRESS_TEST_SUITE_REPORT.md)** - Comprehensive suite
- 📁 **[test-results/byzantine-stress-test/](../../test-results/byzantine-stress-test/)** - Test data
- 📁 **[test-results/byzantine-stress-test-suite/](../../test-results/byzantine-stress-test-suite/)** - Suite data

### GPU Acceleration Tests
- 📊 **[GPU_TESTING_COMPLETE.md](/Documentation/Testing/GPU_TESTING_COMPLETE.md)** - GPU test summary
- 📊 **[GPU_TESTING_RESULTS_REPORT.md](/Documentation/Testing/GPU_TESTING_RESULTS_REPORT.md)** - Detailed results
- 📊 **[GPU_VALIDATION_COMPLETE.md](/Documentation/Testing/GPU_VALIDATION_COMPLETE.md)** - Validation report
- 📊 **[GPU_ACCELERATION_GUIDE.md](/Documentation/Performance/GPU_ACCELERATION_GUIDE.md)** - Setup and usage guide
- 📄 Raw JSON results:
  - [test-results/benchmarks/gpu-benchmark-baseline.json](../../test-results/benchmarks/gpu-benchmark-baseline.json)
  - [test-results/benchmarks/gpu-contention-5nodes.json](../../test-results/benchmarks/gpu-contention-5nodes.json)
  - [test-results/benchmarks/gpu-contention-10nodes.json](../../test-results/benchmarks/gpu-contention-10nodes.json)
  - [test-results/benchmarks/gpu-contention-20nodes.json](../../test-results/benchmarks/gpu-contention-20nodes.json)
  - [test-results/benchmarks/gpu-contention-30nodes.json](../../test-results/benchmarks/gpu-contention-30nodes.json)
  - [test-results/benchmarks/gpu-round-5nodes.json](../../test-results/benchmarks/gpu-round-5nodes.json)
  - [test-results/benchmarks/gpu-round-10nodes.json](../../test-results/benchmarks/gpu-round-10nodes.json)
  - [test-results/benchmarks/gpu-round-20nodes.json](../../test-results/benchmarks/gpu-round-20nodes.json)

### NPU Performance Analysis
- 📊 **[NPU_PERFORMANCE_SCALING_COMPLETE.md](/Documentation/Performance/NPU_PERFORMANCE_SCALING_COMPLETE.md)** - NPU scaling analysis
- 📊 **[NPU_GPU_CPU_PERFORMANCE_ANALYSIS.md](/Documentation/Performance/NPU_GPU_CPU_PERFORMANCE_ANALYSIS.md)** - Comparative analysis
- 📊 **[GRAFANA_NPU_LAPTOP_FINAL.md](/Documentation/Deployment/GRAFANA_NPU_LAPTOP_FINAL.md)** - Grafana NPU monitoring
- 📄 [test-results/benchmarks/npu-gpu-cpu-comparison.json](../../test-results/benchmarks/npu-gpu-cpu-comparison.json) - Raw comparison data

### TPM & Trust Tests
- 📁 **[test-results/tpm-npu-full/](../../test-results/tpm-npu-full/)** - TPM+NPU integrated tests
- 📦 **[test-results/tpm-npu-full-artifacts.tar.gz](../../test-results/tpm-npu-full-artifacts.tar.gz)** - Archived artifacts

### Demo & Simulation Results
- 📁 **[test-results/demo-simulated/](../../test-results/demo-simulated/)** - Simulated demo runs
- 📊 **[COMPLETE_DEMO_DATA_VIEWABLE.md](/Documentation/Reports/COMPLETE_DEMO_DATA_VIEWABLE.md)** - Demo data overview
- 📊 **[DEMO_RESULTS_SUMMARY.txt](/Documentation/Reports/DEMO_RESULTS_SUMMARY.txt)** - Quick summary
- 📊 **[README_DEMO_RESULTS.md](/Documentation/Reports/README_DEMO_RESULTS.md)** - Demo results documentation

### Large-Scale Tests
- 📁 **[test-results/20260227T004204Z_200round_fullscope/](../../test-results/20260227T004204Z_200round_fullscope/)** - 200-round full scope test
- 📁 **[test-results/incremental_scale_test_full_20260228_122530/](../../test-results/incremental_scale_test_full_20260228_122530/)** - Incremental scaling
- 📁 **[test-results/scale_test_5000r_100burst_20260228_123024/](../../test-results/scale_test_5000r_100burst_20260228_123024/)** - 5000-round burst test

---

## 📈 Visualization & Analysis Scripts

- 📈 **[analyze-demo-results.py](../../analyze-demo-results.py)** - Demo results analyzer
- 📈 **[analyze-gpu-results.py](../../analyze-gpu-results.py)** - GPU results analyzer
- 📈 **[tests/scripts/python/generate-byzantine-plots.py](../../tests/scripts/python/generate-byzantine-plots.py)** - Byzantine test plots
- 📈 **[tests/scripts/python/generate-byzantine-test-suite-plots.py](../../tests/scripts/python/generate-byzantine-test-suite-plots.py)** - Suite visualization
- 📈 **[generate-demo-report.py](../../generate-demo-report.py)** - Demo report generator
- 📈 **[generate-k8s-5000-node-plots.py](../../generate-k8s-5000-node-plots.py)** - K8s plots generator

---

## 🧪 Test Execution Scripts

### Quick Start
- 🚀 **[genesis-launch.sh](../../genesis-launch.sh)** - One-command stack launch

### Test Runners
- 🧪 **[tests/scripts/bash/run-test.sh](../../tests/scripts/bash/run-test.sh)** - General test runner (Linux)
- 🧪 **[tests/scripts/powershell/run-test.ps1](../../tests/scripts/powershell/run-test.ps1)** - General test runner (PowerShell)
- 🧪 **[tests/scripts/python/run-test-python.py](../../tests/scripts/python/run-test-python.py)** - Python test orchestrator
- 🧪 **[tests/scripts/powershell/run-full-test.ps1](../../tests/scripts/powershell/run-full-test.ps1)** - Full test suite (PowerShell)

### Specialized Tests
- 🐍 **[simulate-demo.py](../../simulate-demo.py)** - Demo simulator
- 🐍 **[tests/scripts/python/gpu-test-suite.py](../../tests/scripts/python/gpu-test-suite.py)** - GPU test suite
- 🐍 **[tests/scripts/python/byzantine-stress-test-suite.py](../../tests/scripts/python/byzantine-stress-test-suite.py)** - Byzantine stress tests
- 🐍 **[tests/scripts/python/npu-gpu-cpu-benchmark.py](../../tests/scripts/python/npu-gpu-cpu-benchmark.py)** - Hardware acceleration benchmark

### Monitoring & Analysis
- 📊 **[monitor-demo.py](../../monitor-demo.py)** - Real-time demo monitoring
- 📊 **[monitor.bat](../../monitor.bat)** - Windows monitoring script

---

## 🐳 Docker Compose Configurations

- **[docker-compose.full.yml](../../docker-compose.full.yml)** - 1000-node NPU test stack
- **[docker-compose.full.yml](../../docker-compose.full.yml)** - Large-scale deployment
- **[docker-compose.full.yml](../../docker-compose.full.yml)** - Monitoring stack
- **[docker-compose.full.yml](../../docker-compose.full.yml)** - Production configuration
- **[docker-compose.full.yml](../../docker-compose.full.yml)** - Complete stack

---

## 📊 Key Performance Metrics

### 1000-Node NPU Test
- **Throughput:** 2,850 RPS (4.38x improvement)
- **Latency Reduction:** 66.9%
- **Nodes:** 1,000
- **Test Duration:** 22 minutes 22 seconds

### 5000-Node K8s Test
- **Nodes:** 5,000
- **Byzantine Tolerance:** 50% (2,500 malicious nodes)
- **Accuracy:** 86.00%
- **Test Duration:** 2.8 seconds
- **Scenarios Passed:** 4/4 (100%)

### GPU Tests
- **Configurations Tested:** 5, 10, 20, 30 nodes
- **GPU Contention Analysis:** Complete
- **Round Performance:** Validated

---

## 📚 Related Documentation

- 📖 **[TEST_GUIDE.md](/Documentation/Testing/TEST_GUIDE.md)** - Comprehensive testing guide
- 📖 **[TESTING_COMPLETION_INDEX.md](/Documentation/Testing/TESTING_COMPLETION_INDEX.md)** - Testing completion tracker
- 📖 **[ARCHITECTURE.md](/Documentation/Architecture/ARCHITECTURE.md)** - System architecture
- 📖 **[DEPLOYMENT.md](/Documentation/Deployment/DEPLOYMENT.md)** - Deployment guide
- 📖 **[HIL_TESTING.md](/Documentation/Testing/HIL_TESTING.md)** - Hardware-in-the-loop testing

---

## 🔍 Quick Search Tips

**Finding specific test results:**
```bash
# All 1000-node NPU artifacts
find . -path "*/1000-node*" -o -name "*1000*node*npu*"

# All 5000-node K8s artifacts
find . -path "*/kubernetes-5000*" -o -name "*5000*node*"

# All test result directories
ls -la test-results/

# All visualization plots
find test-results/ -name "*.png"
```

---

**Last Updated:** 2026-03-07  
**Artifact Status:** All test artifacts committed and verified in repository
