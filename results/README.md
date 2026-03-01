# 📊 Test Results Directory

Comprehensive test results, benchmarks, and analysis from all validation phases.

> Benchmark interpretation note: these artifacts are historical run outputs captured under specific environments/configurations. They support engineering analysis but are not universal guarantees for every deployment.

---

## Structure

```
results/
├── test-runs/           Raw test output and JSON results
├── benchmarks/          Performance metrics and CSV data
└── analysis/            Comprehensive analysis documents
```

---

## Test Runs (results/test-runs/)

Raw outputs from test executions stored as JSON.

### Files

```
100k_nodes_results.json
  - 100K node scaling test
  - Accuracy by round
  - Timing metrics
  - Recovery analysis

500k_nodes_results.json
  - 500K node stress test
  - Per-round timing
  - Byzantine impact
  - Memory efficiency

10m_nodes_results.json
  - 10M node extreme scale
  - Complete execution data
  - Breakthrough metrics

boundary_test_results.json
  - Byzantine boundary sweep
  - 51-60% Byzantine levels
  - Accuracy floor data
  - Recovery times
```

### Usage

```bash
# Load and analyze results
python -c "
import json
with open('results/test-runs/100k_nodes_results.json') as f:
    data = json.load(f)
    print(f\"Final accuracy: {data['results']['final_accuracy']}\")
    print(f\"Average accuracy: {data['results']['avg_accuracy']}\")
"
```

---

## Benchmarks (results/benchmarks/)

Performance metrics in CSV format for easy analysis.

### Files

#### throughput_analysis.csv
```
Scale,Byzantine%,UpdatesPerSec,Status
100000,40,5000,GOOD
500000,40,50000,EXCELLENT
10000000,40,71428,BREAKTHROUGH
```

#### latency_benchmarks.csv
```
Scale,Byzantine%,AvgLatency(s),MinLatency(s),MaxLatency(s)
100000,0,15,14,16
500000,0,10,8,12
10000000,0,127,117,136
```

#### accuracy_by_scale.csv
```
Scale,Byzantine%,FinalAccuracy,AvgAccuracy,MinAccuracy
100000,50,85.0,80.3,76.0
500000,50,85.8,83.0,80.6
10000000,50,83.9,82.2,81.3
```

#### memory_usage.csv
```
Scale,Byzantine%,PeakMemory(MB),AvgMemory(MB),MemoryTrend
100000,50,4500,3200,stable
500000,50,8200,6000,stable
10000000,50,12000,9000,stable
```

### Generating Benchmark Reports

```bash
# Create summary report
cat results/benchmarks/*.csv | head -20

# Analyze scaling trends
python -c "
import pandas as pd
df = pd.read_csv('results/benchmarks/latency_benchmarks.csv')
print(df.sort_values('Scale'))
"
```

---

## Analysis Documents (results/analysis/)

Comprehensive markdown documents with detailed analysis.

### Key Documents

#### EXTREME_SCALE_10M_RESULTS.md
- 10M node test execution
- Breakthrough findings
- Scaling analysis
- Petabyte-scale feasibility
- **Size:** 10.0 KB

#### STRESS_TEST_500K_RESULTS.md
- 500K node stress validation
- 40-55% Byzantine testing
- Memory efficiency proof
- Performance scaling
- **Size:** 8.2 KB

#### BYZANTINE_BOUNDARY_TEST_RESULTS.md
- 52-55.5% Byzantine boundary
- Empirical cliff analysis
- Recovery time curves
- Production recommendations
- **Size:** 10.5 KB

#### TEST_EXECUTION_SUMMARY.md
- Complete test timeline
- All results summary
- Key discoveries
- Recommendations
- **Size:** 9.3 KB

#### RESEARCH_FINDINGS.md
- Complete research analysis
- Byzantine tolerance bounds
- Scaling validation
- Privacy compliance
- Production deployment guide
- **Size:** 16.7 KB

---

## Summary Statistics

### Total Test Configurations

```
Byzantine Boundary Tests:    18 configurations
500K Stress Tests:           15 configurations
10M Extreme Tests:            6 configurations
Scale Validation Tests:      50+ configurations
──────────────────────────────────────────
Total:                       ~90 configurations
Success Rate:               100%
Total Runtime:              ~14 hours
```

### Nodes Tested

```
100K nodes:      50+ configurations
500K nodes:      15 configurations
10M nodes:        6 configurations
──────────────────────────────
Total Node-Rounds: 30M+
```

### Data Volume

```
JSON Results:       ~5 MB
CSV Benchmarks:     ~100 KB
Analysis Docs:      ~85 KB
──────────────────
Total:              ~5.2 MB
```

---

## Accessing Results

### Command Line

```bash
# View latest results
ls -lh results/test-runs/*.json

# Get summary statistics
grep "final_accuracy" results/test-runs/*.json

# Analyze performance
awk -F',' '$1>100000 {print}' results/benchmarks/latency_benchmarks.csv
```

### Python Analysis

```python
import json
import pandas as pd

# Load test results
with open('results/test-runs/10m_nodes_results.json') as f:
    results = json.load(f)

# Load benchmarks
df = pd.read_csv('results/benchmarks/throughput_analysis.csv')
print(df[df['Scale'] == 10000000])
```

### Data Visualization

```bash
# Create plots from benchmarks
python -c "
import pandas as pd
import matplotlib.pyplot as plt

df = pd.read_csv('results/benchmarks/latency_benchmarks.csv')
plt.plot(df['Scale'], df['AvgLatency(s)'])
plt.xlabel('Scale (nodes)')
plt.ylabel('Latency (seconds)')
plt.title('Latency vs Scale')
plt.savefig('latency_plot.png')
"
```

---

## Benchmark Interpretation

### Good Performance Indicators

✅ Throughput increasing with scale  
✅ Latency predictable and linear  
✅ Accuracy maintained above 80%  
✅ Memory usage stable (no bloat)  
✅ 100% success rate

### Warning Signs

⚠️ Latency increasing exponentially  
⚠️ Throughput decreasing at scale  
⚠️ Accuracy dropping below 75%  
⚠️ Memory growing unbounded  
⚠️ Failures or timeouts

### Regression Detection

Compare new results against baseline:

```bash
# Baseline (from v1.0.0a)
# 500K nodes: 83.6% accuracy, 10s/round

# New test:
# 500K nodes: 82.5% accuracy, 12s/round

# Regression indicators:
# Accuracy down 1.1%
# Latency up 20%
```

---

## Archiving Old Results

```bash
# Archive results older than 30 days
find results/test-runs -mtime +30 -tar -czf archive_$(date +%Y%m%d).tar.gz

# Keep recent results for quick access
# Archive older results for reference
```

---

## Publication-Ready Exports

### Generate Paper-Quality Results

```bash
# Export key findings
python -c "
import json

results = {
    '100K nodes': {'accuracy': '86%', 'latency': '15-20s'},
    '500K nodes': {'accuracy': '83.6%', 'latency': '10s'},
    '10M nodes': {'accuracy': '82-83%', 'latency': '127-154s'}
}

# Format for LaTeX table
for scale, metrics in results.items():
    print(f\"{scale} & {metrics['accuracy']} & {metrics['latency']}\\\\\")
"
```

---

## Result Reproducibility

### Run Same Test Again

```bash
# Original test command
python tests/scale-tests/bft_extreme_scale_10m.py

# Results should be within 5% of baseline
# Due to random number generation

# For exact reproduction, fix random seed
# (modify test file to seed random number generator)
```

### Version Tracking

```
Results from: v1.0.0a release
Test date: February 24, 2026
Hardware: AWS t3.2xlarge
Python: 3.11
NumPy: 1.24.x
```

---

## Continuous Monitoring

### Set Up Automated Testing

```bash
# Run tests nightly
0 2 * * * cd /path && python tests/scale-tests/*.py >> results/nightly.log

# Store results
# Monitor for regressions
# Alert on failures
```

---

**Results Directory**  
**v1.0.0a Release**  
**February 24, 2026**
