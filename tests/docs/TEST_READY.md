# 🚀 Incremental Scale Test - Ready to Launch

## What's Been Prepared

A complete end-to-end test orchestration system for the Sovereign Map federated learning platform with:

✅ **Test Orchestration Script** - Full 500-round test with incremental scaling
✅ **Real-Time Dashboard** - Live monitoring of all metrics  
✅ **Configuration Management** - Customizable test parameters
✅ **Automated Commit Helper** - One-command result submission
✅ **Comprehensive Guides** - Complete documentation

---

## 🎯 Test Specification

**Deployment Strategy:**
- Start: 20 nodes
- Scale: +20 nodes every time 93% accuracy is reached
- Target: 100 nodes maximum
- Duration: 500 federated learning rounds

**Advanced Features:**
- ✅ TPM attestation verification
- ✅ NPU/GPU acceleration
- ✅ Real-time convergence monitoring
- ✅ System metrics collection
- ✅ Byzantine tolerance validation

**Expected Timeline:**
- Initialization: ~5 min
- Main test loop: ~42 min (500 rounds × 5s)
- Result collection: ~5 min
- **Total: ~52 minutes**

---

## 📦 Files Created

### 1. **tests/scripts/bash/test-incremental-scale.sh** (13.5 KB)
Main test orchestration script that:
- Deploys initial 20-node cluster
- Monitors convergence every round
- Scales to +20 nodes when 93% accuracy reached
- Collects TPM attestation metrics
- Gathers NPU acceleration data
- Generates comprehensive test report

### 2. **tests/scripts/bash/test-dashboard.sh** (9.7 KB)
Real-time monitoring dashboard showing:
- Current convergence metrics
- Node scaling history
- System health (CPU, memory, services)
- TPM attestation status
- NPU acceleration metrics
- Overall progress bar (0-500 rounds)
- Live event log

### 3. **tests/scripts/bash/commit-test-results.sh** (3.1 KB)
Automated commit preparation:
- Finds latest test results
- Extracts key metrics
- Stages all artifacts
- Creates meaningful commit message
- Shows summary before committing

### 4. **tests/config/test-config.env** (1.5 KB)
Test configuration file with:
- Scaling parameters
- Convergence thresholds
- TPM/NPU settings
- Backend configuration
- Monitoring options
- Privacy settings

### 5. **tests/docs/TEST_GUIDE.md** (12.5 KB)
Complete guide covering:
- Quick start instructions
- Test workflow phases
- Configuration customization
- Output structure explanation
- Monitoring options
- Result analysis
- Troubleshooting

---

## 🚀 Quick Start (4 Steps)

Canonical commands below use `tests/...` paths. Root-level script names remain available as compatibility wrappers.

### Step 1: Navigate to Project
```bash
cd Sovereign_Map_Federated_Learning
chmod +x *.sh
```

### Step 2: Start Test
```bash
bash tests/scripts/bash/test-incremental-scale.sh
# This will:
# - Deploy 20-node cluster
# - Run 500 rounds with monitoring
# - Collect TPM and NPU metrics
# - Generate report in test-results/
# Expected time: ~52 minutes
```

### Step 3: Monitor Progress (In Separate Terminal)
```bash
bash tests/scripts/bash/test-dashboard.sh
# Real-time dashboard updates every 5 seconds
# Shows accuracy, node count, system health
# Live progress towards 500 rounds
```

### Step 4: Commit Results
```bash
bash tests/scripts/bash/commit-test-results.sh
# Stages results
# Creates commit message with metrics
# Asks for confirmation
# Commits to git
```

---

## 📊 Expected Output Structure

After test completes, results in `test-results/incremental_scale_test_YYYYMMDD_HHMMSS/`:

```
├── TEST_REPORT.md              # Executive summary
│   ├── Configuration details
│   ├── Convergence history
│   ├── Scaling events log
│   ├── Final metrics
│   └── Status: PASSED ✅
│
├── convergence.log             # Per-round metrics (JSONL)
│   └── {timestamp, round, nodes, accuracy, loss}
│
├── metrics.jsonl               # System metrics per round
│   └── {timestamp, round, backend_stats, prometheus_metrics}
│
├── tpm_attestation.json        # TPM verification results
│   ├── tpm_enabled: true
│   ├── attestation_results: [node trust scores]
│   └── status: completed
│
├── npu_metrics.json            # GPU/NPU acceleration data
│   ├── npu_enabled: true
│   ├── hardware_info
│   └── compute_metrics
│
└── test.log                    # Complete execution log
    └── All timestamped events
```

---

## 📈 Key Metrics You'll Get

**Convergence Analysis:**
- Accuracy progression: 0% → 93%+ over rounds
- Loss trajectory: High → Low
- Convergence speed per scale level

**Scaling Performance:**
- Time to reach 93% at 20 nodes
- Time to reach 93% at 40 nodes
- Time to reach 93% at 60 nodes
- Time to reach 93% at 80 nodes
- Time to reach 93% at 100 nodes

**System Performance:**
- Backend CPU/memory usage
- Node agent resource consumption
- Database query latency
- Overall throughput

**TPM Verification:**
- Node trust scores (should be >75)
- Certificate validation status
- Message signature verification success rate

**NPU Utilization:**
- GPU/CPU utilization percentage
- Memory usage
- Acceleration speedup factor

---

## 🎓 Understanding the Test

### What Gets Tested

1. **Federated Learning Convergence**
   - Model trains across distributed nodes
   - Aggregates updates every round
   - Tracks accuracy improvement

2. **Byzantine Tolerance**
   - System maintains accuracy with faulty nodes
   - Robust aggregation strategy
   - Convergence under adversarial conditions

3. **Horizontal Scaling**
   - Adding nodes doesn't break training
   - Convergence maintained at each scale
   - Smooth transitions between scales

4. **TPM Security**
   - Trust verification for each node
   - Certificate chain validation
   - Message authentication

5. **NPU Acceleration**
   - GPU/hardware acceleration
   - Performance improvement measurement
   - Resource efficiency

### What Happens Each Round

```
Round N:
  1. Backend collects updates from all active nodes
  2. Performs Byzantine-robust aggregation
  3. Calculates new accuracy/loss
  4. Check: accuracy >= 93%?
     YES → If nodes < 100: SCALE UP
     NO  → Continue training
  5. Log metrics
  6. Wait 5 seconds
  7. Round N+1 starts
```

### Convergence Events Expected

- **Event 1** (~Round 75): Accuracy reaches 93% at 20 nodes → Scale to 40
- **Event 2** (~Round 150): Accuracy recovers to 93% at 40 nodes → Scale to 60
- **Event 3** (~Round 225): Accuracy recovers to 93% at 60 nodes → Scale to 80
- **Event 4** (~Round 300): Accuracy recovers to 93% at 80 nodes → Scale to 100
- **Event 5+**: Continue training at 100 nodes for remaining 200 rounds

---

## 🔍 Live Monitoring During Test

### Real-Time Dashboard
```bash
bash tests/scripts/bash/test-dashboard.sh

# Shows every 5 seconds:
# 📈 Accuracy: 93.5% (↑ from 91.2%)
# 📊 Nodes: 60 (recently scaled from 40)
# 📉 Loss: 0.215 (↓ from 0.312)
# 🎯 Convergence Events: 3
# 💻 Backend RAM: 2.1GB / 4GB
# 🔐 TPM Status: ✓ Verified
# ⚡ NPU Util: 87%
# Progress: [████████░░░░░░░░░░░░░░░░░░] 34%
```

### Alternative: Watch Raw Metrics
```bash
# Terminal 1: Latest metrics
watch -n 5 'tail -1 test-results/*/convergence.log | jq'

# Terminal 2: Docker stats
watch -n 1 'docker stats --no-stream'

# Terminal 3: Live logs
tail -f test-results/*/test.log | grep -E "Round|Convergence|Scaling"
```

---

## 📋 Pre-Test Checklist

Before starting, verify:

- [ ] Docker is running: `docker ps`
- [ ] Sufficient disk space: `df -h` (need ~2GB)
- [ ] Memory available: `free -h` (need ~8-16GB)
- [ ] jq installed: `jq --version`
- [ ] Git configured: `git config --list | head -3`
- [ ] Compose file exists: `ls docker-compose.large-scale.yml`
- [ ] Test scripts executable: `ls -l *.sh`

---

## ⚠️ Important Notes

### During Test
- **Don't interrupt**: Let test run to completion (~52 min)
- **Don't modify**: Files are in use, modifications will be lost
- **Monitor wisely**: Dashboard runs in separate terminal
- **Check resources**: Stop other heavy processes

### Expected Behavior
- First few rounds: Accuracy will be random (~0-10%)
- Convergence phase: Quick rise to ~93% (15-25 rounds)
- Scaling impact: Temporary dip when nodes added, then recovery
- Stabilization: High accuracy maintained at 100 nodes

### If Something Goes Wrong
```bash
# Check test status
tail -100 test-results/*/test.log

# View latest metrics
jq . test-results/*/convergence.log | tail -5

# Check backend health
curl http://localhost:8000/health

# Check containers
docker compose -f docker-compose.large-scale.yml ps

# Restart if needed
docker compose -f docker-compose.large-scale.yml down
bash tests/scripts/bash/test-incremental-scale.sh
```

---

## 💾 Commit & Push

### Automatic (Recommended)
```bash
bash tests/scripts/bash/commit-test-results.sh
# Guides you through the process
```

### Manual
```bash
# Stage everything
git add test-results/incremental_scale_test_*/
git add Sovereign_Map_Federated_Learning/test-*.sh
git add tests/config/test-config.env

# Create commit with proper message
git commit -m "test: incremental scale 20→100 nodes, 500 rounds" \
  -m "" \
  -m "Final accuracy: 94.3%" \
  -m "Convergence events: 4" \
  -m "Status: PASSED ✅" \
  -m "" \
  -m "Test artifacts: test-results/incremental_scale_test_*/" \
  -m "Assisted-By: cagent"

# Push
git push origin HEAD
```

---

## 🎯 Success Criteria

Test passes if:

✅ Reaches 93% accuracy at 20 nodes  
✅ Successfully scales to 40 nodes  
✅ Successfully scales to 60 nodes  
✅ Successfully scales to 80 nodes  
✅ Successfully scales to 100 nodes  
✅ Completes all 500 rounds  
✅ TPM attestation passes  
✅ NPU acceleration detected (if available)  
✅ Results properly formatted in test-results/  

---

## 📞 Command Reference

**Start test**
```bash
bash tests/scripts/bash/test-incremental-scale.sh
```

**Monitor test** (in another terminal)
```bash
bash tests/scripts/bash/test-dashboard.sh
```

**Manual monitoring**
```bash
# Latest convergence
tail -1 test-results/*/convergence.log | jq

# Scale timeline
jq -r '.nodes' test-results/*/convergence.log | uniq

# Accuracy progression
jq -r '.accuracy' test-results/*/convergence.log | tail -20
```

**View report**
```bash
cat test-results/incremental_scale_test_*/TEST_REPORT.md
```

**View all metrics**
```bash
jq . test-results/*/convergence.log
jq . test-results/*/metrics.jsonl
```

**Commit results**
```bash
bash tests/scripts/bash/commit-test-results.sh
```

---

## 🚀 Ready?

Everything is prepared and ready to run:

```bash
cd Sovereign_Map_Federated_Learning
bash tests/scripts/bash/test-incremental-scale.sh
```

The test will:
1. ✅ Deploy 20 nodes
2. ✅ Run 500 training rounds
3. ✅ Scale incrementally to 100 nodes
4. ✅ Verify TPM attestation
5. ✅ Measure NPU acceleration
6. ✅ Generate comprehensive report
7. ✅ Prepare for git commit

**Estimated time: ~52 minutes**

Good luck! 🎉

---

**Questions?** Check tests/docs/TEST_GUIDE.md for detailed documentation.
