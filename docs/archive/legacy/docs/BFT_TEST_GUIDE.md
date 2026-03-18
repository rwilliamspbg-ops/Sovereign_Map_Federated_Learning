# 🔐 BFT STRESS TEST EXECUTION GUIDE

## Quick Start

```bash
cd Sovereign_Map_Federated_Learning

# Run the BFT test
python3 run_bft_test.py
```

---

## What This Does

**Tests Byzantine Fault Tolerance across 12 configurations:**

| Byzantine % | Median | Multi-Krum | Expected Result |
|-------------|--------|-----------|-----------------|
| 0% (Baseline) | ✅ | ✅ | Both converge perfectly |
| 10% | ✅ | ✅ | Both converge, slight degradation |
| 20% | ✅ | ⚠️ | Both converge, Multi-Krum better |
| 30% | ⚠️ | ⚠️ | Borderline, some degradation |
| 40% | ❌ | ❌ | **CONVERGENCE FAILS** |
| 50% | ❌ | ❌ | Complete failure |

---

## What You'll See

**Console Output:**
```
================================================================================
  🚀 STARTING BFT STRESS TEST - 75 NODES | 200 ROUNDS EACH
================================================================================

✅ Backend healthy at http://localhost:8081
📤 Sending test start request...
✅ Test started successfully
   Tests Queued: 12

📊 MONITORING TEST PROGRESS (updating every 15s)

[000m] Tests Completed: 0/12
[001m] Tests Completed: 1/12 | Latest: 0% Byzantine (median) | Final Acc: 91.2% | Converged: ✅
[002m] Tests Completed: 2/12 | Latest: 0% Byzantine (multi_krum) | Final Acc: 91.1% | Converged: ✅
[004m] Tests Completed: 3/12 | Latest: 10% Byzantine (median) | Final Acc: 89.3% | Converged: ✅
...
[45m] Tests Completed: 12/12 | Latest: 50% Byzantine (multi_krum) | Final Acc: 42.1% | Converged: ❌

✅ ALL TESTS COMPLETED!

================================================================================
  ✅ BFT STRESS TEST COMPLETED
================================================================================

Total Duration: 0:45:32
Tests Run: 12

Critical Threshold: 40%

✅ Report saved to BFT_TEST_RESULTS.md
```

---

## Output Files

- **BFT_TEST_RESULTS.md** - Complete markdown report with findings
- **bft_results.json** - Raw JSON data from all 12 tests
- **bft_curves.json** - Convergence curves (200 points each)

---

## Expected Duration

| Phase | Time |
|-------|------|
| Setup | 5m |
| 12 Tests @ ~3m each | 36m |
| Analysis | 5m |
| **Total** | **~45 minutes** |

---

## Key Findings to Look For

### 1. **Convergence Threshold**
Question: At what % do systems fail to converge?

Expected Answer: 40% (theory: 33%)

### 2. **Algorithm Comparison**
Multi-Krum should outperform Median significantly at 20%+

### 3. **Accuracy Degradation**
How much lower accuracy at each Byzantine %?

### 4. **Convergence Speed**
How many rounds to reach 80% accuracy?

---

## Monitoring During Test

### Option 1: Watch Docker Stats
```bash
docker stats --no-stream
# Monitor memory as 75 nodes process 200 rounds
```

### Option 2: Monitor Logs
```bash
docker logs -f sovereign_map_federated_learning-backend-1 | grep "BFT Test"
```

### Option 3: Check Prometheus
```
http://localhost:9090/graph
# Query: bft_test_accuracy{byzantine_pct}
```

---

## If Test Fails

### Backend Unreachable
```bash
docker compose ps backend
docker logs sovereign_map_federated_learning-backend-1
```

### Out of Memory
```bash
docker stats --no-stream
# If memory > 95%, scale down:
docker compose down
docker system prune -a --force
docker compose up -d --scale node-agent=50
```

### Test Stalled
```bash
# Check if containers are still running
docker ps --filter "label=com.docker.compose.project=sovereign_map_federated_learning" | wc -l

# If many exited, restart
docker compose up -d --scale node-agent=75
```

---

## Understanding Results

### Converged = TRUE
- Final accuracy ≥ 80%
- Last 10 rounds show <1% variance
- Convergence rate stabilized

### Converged = FALSE
- Final accuracy < 80%
- Accuracy oscillates or diverges
- Loss increases in final rounds

### Critical Threshold
The Byzantine % where **NO** configurations converge

---

## Test Configurations

Each test has 12 runs (2 hours of distributed learning):

**Config Structure:**
```
Test 1:  0% Byzantine + Median        (200 rounds)
Test 2:  0% Byzantine + Multi-Krum    (200 rounds)
Test 3:  10% Byzantine + Median       (200 rounds)
Test 4:  10% Byzantine + Multi-Krum   (200 rounds)
Test 5:  20% Byzantine + Median       (200 rounds)
Test 6:  20% Byzantine + Multi-Krum   (200 rounds)
Test 7:  30% Byzantine + Median       (200 rounds)
Test 8:  30% Byzantine + Multi-Krum   (200 rounds)
Test 9:  40% Byzantine + Median       (200 rounds)
Test 10: 40% Byzantine + Multi-Krum   (200 rounds)
Test 11: 50% Byzantine + Median       (200 rounds)
Test 12: 50% Byzantine + Multi-Krum   (200 rounds)
```

---

## Academic Value

This test answers the central question:
**"What is the empirical Byzantine fault tolerance of Sovereign Map at scale?"**

Results will show:
- ✅ Theoretical limit (33%) vs empirical limit
- ✅ Multi-Krum effectiveness at scale (75 nodes)
- ✅ Convergence curves under adversarial conditions
- ✅ Practical recommendations for deployment

---

## Next Steps After Test

1. **Review Report**
   ```bash
   cat BFT_TEST_RESULTS.md
   ```

2. **Analyze Curves**
   - Plot accuracy vs round for each config
   - Compare Median vs Multi-Krum visually

3. **Implement Findings**
   - Recommend safe Byzantine tolerance for production
   - Choose best aggregation algorithm

4. **Scale to Production**
   - Use findings to determine safe deployment limits
   - Configure monitoring for Byzantine detection

---

**Status:** ✅ Ready to execute  
**Node Count:** 75 agents  
**Duration:** ~45 minutes  
**Expected Threshold:** 40% Byzantine (empirical vs 33% theory)
