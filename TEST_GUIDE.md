# Incremental Scale Test - Complete Guide

## Overview

This comprehensive test orchestration runs the Sovereign Map federated learning system with:
- **Initial deployment**: 20 nodes
- **Scaling strategy**: +20 nodes every time 93% convergence is reached
- **Target**: 100 nodes maximum
- **Duration**: 500 rounds
- **Advanced features**: TPM attestation, NPU acceleration, continuous monitoring

## Quick Start

### 1. Prepare the Test

```bash
# Navigate to project directory
cd Sovereign_Map_Federated_Learning

# Make scripts executable
chmod +x test-incremental-scale.sh test-dashboard.sh commit-test-results.sh

# Review configuration
cat test-config.env
```

### 2. Start the Test

```bash
# Start test (runs in foreground, logs to test-results/)
bash test-incremental-scale.sh

# Expected output:
# ✓ Pre-flight checks
# ✓ Backend deployed with 20 nodes
# ✓ TPM attestation enabled
# ✓ NPU acceleration configured
# ✓ Test loop started (500 rounds)
# ✓ Results collected
# ✓ Report generated
```

### 3. Monitor Progress (In Another Terminal)

```bash
# Open real-time dashboard
bash test-dashboard.sh

# Dashboard shows:
# - Current convergence metrics
# - Node scaling history
# - System health
# - TPM attestation status
# - NPU acceleration metrics
# - Progress bar
```

### 4. Collect and Commit Results

Once test completes:

```bash
# Create git commit with results
bash commit-test-results.sh

# This will:
# ✓ Find latest test results
# ✓ Extract key metrics
# ✓ Stage all test artifacts
# ✓ Create meaningful commit message
# ✓ Show summary before commit
```

---

## Test Workflow

### Phase 1: Initialization (5 minutes)
```
✓ Pre-flight checks (Docker, jq, compose)
✓ Build and deploy with 20 nodes
✓ Verify backend health
✓ Enable TPM attestation
✓ Enable NPU acceleration
```

### Phase 2: Main Test Loop (Variable Duration)
```
For each round (0-500):
  1. Monitor convergence metrics
  2. Collect system metrics
  3. If accuracy >= 93% AND nodes < 100:
     - Scale up by 20 nodes
     - Wait for health checks
     - Continue training
  4. Log metrics (convergence.log, metrics.jsonl)
  5. Sleep 5 seconds before next iteration
```

### Phase 3: Result Collection (5 minutes)
```
✓ Extract TPM attestation metrics
✓ Collect NPU acceleration data
✓ Generate comprehensive report
✓ Archive all results to test-results/
```

### Phase 4: Commit & Archive (2 minutes)
```
✓ Stage test results and scripts
✓ Create meaningful commit message
✓ Push to git repository
```

---

## Configuration

Edit `test-config.env` to customize:

```bash
# Test Parameters
INITIAL_NODES=20                 # Start with 20 nodes
INCREMENT_NODES=20               # Add 20 nodes per convergence
MAX_NODES=100                    # Cap at 100 nodes
CONVERGENCE_THRESHOLD=93         # Scale when 93% accuracy reached
TOTAL_ROUNDS=500                 # Run 500 FL rounds

# Feature Flags
TPM_ENABLED=true                 # Enable TPM attestation
NPU_ENABLED=true                 # Enable NPU acceleration

# Backend
BACKEND_WORKERS=4                # Number of aggregator workers
BATCH_SIZE=32                    # Training batch size
LEARNING_RATE=0.001              # Model learning rate

# Monitoring
PROMETHEUS_RETENTION=90d         # Keep metrics for 90 days
GRAFANA_DASHBOARD_AUTO_SYNC=true # Auto-sync dashboards
```

---

## Output Structure

After test completion:

```
test-results/
└── incremental_scale_test_YYYYMMDD_HHMMSS/
    ├── TEST_REPORT.md           # Final report (human readable)
    ├── test.log                 # Complete execution log
    ├── convergence.log          # Per-round metrics (JSONL)
    ├── metrics.jsonl            # System metrics per round
    ├── tpm_attestation.json     # TPM trust verification
    └── npu_metrics.json         # NPU acceleration data
```

### Key Files Explained

**TEST_REPORT.md** - Executive summary with:
- Test configuration
- Convergence history (last 20 rounds)
- Scaling events
- Final accuracy and convergence count
- Status: PASSED ✅

**convergence.log** - Line-delimited JSON (JSONL) format:
```json
{"timestamp":"2026-02-28 12:00:00","round":"0","nodes":"20","accuracy":"0.0","loss":"2.302"}
{"timestamp":"2026-02-28 12:00:05","round":"1","nodes":"20","accuracy":"5.3","loss":"2.145"}
...
```

**metrics.jsonl** - System metrics per round:
```json
{"timestamp":"...","round":"0","nodes":"20","backend_stats":"45.2%,512MB","prometheus_metrics":{...}}
```

**tpm_attestation.json** - TPM verification results
**npu_metrics.json** - GPU/NPU acceleration stats

---

## Monitoring During Test

### Option 1: Real-Time Dashboard (Recommended)
```bash
bash test-dashboard.sh

# Updates every 5 seconds, shows:
# - Current accuracy, loss, nodes
# - Convergence progression
# - Scaling events history
# - System health (CPU, memory)
# - TPM and NPU status
# - Overall progress bar
```

### Option 2: Manual Monitoring

In separate terminal:

```bash
# Watch convergence updates
watch -n 5 'tail -1 test-results/*/convergence.log | jq'

# Monitor docker stats
docker stats --no-stream

# View live logs
tail -f test-results/*/test.log

# Check Prometheus
open http://localhost:9090

# Check Grafana
open http://localhost:3001
```

### Option 3: Backend API

```bash
# Get current convergence
curl http://localhost:8000/convergence | jq

# Get active nodes
curl http://localhost:9090/api/v1/query?query=sovereignmap_active_nodes | jq

# Get accuracy trend
curl http://localhost:9090/api/v1/query_range?query=sovereignmap_fl_accuracy | jq
```

---

## Understanding Test Results

### Convergence Progression

Typical accuracy curve (should increase monotonically):
```
Round  0: 2.3%  (random initialization)
Round 10: 15.4%
Round 20: 28.1%
Round 30: 42.7%
Round 40: 58.3%
Round 50: 71.2%
Round 60: 81.5%
Round 70: 87.3%
Round 75: 93.0% 🎯 CONVERGENCE 1 - Scale to 40 nodes
Round 80: 93.5%
...
```

### Node Scaling Pattern

Should see convergence events at:
- Round ~75: 20 → 40 nodes (event 1)
- Round ~150: 40 → 60 nodes (event 2)
- Round ~225: 60 → 80 nodes (event 3)
- Round ~300: 80 → 100 nodes (event 4)
- Round ~375+: Stay at 100 nodes (no more scaling)

### Key Metrics to Check

**Good Sign ✅**
- Accuracy continuously increasing
- Loss continuously decreasing
- Regular convergence events (every 75 rounds)
- All nodes healthy and responsive
- TPM attestation passed
- GPU/NPU showing utilization

**Problem Signs ⚠️**
- Accuracy plateaus or decreases
- Loss increases
- Convergence events skipped
- Node agents crashing
- TPM attestation failures
- High memory usage (>90%)

---

## Scaling Analysis

After test completes, analyze scaling behavior:

```bash
# Count convergence events
grep "Convergence.*Scaling" test-results/*/test.log | wc -l

# View scaling timeline
grep "Convergence.*Scaling" test-results/*/test.log

# Extract scaling pattern
jq -r '.nodes' test-results/*/convergence.log | uniq | paste -sd ' '
# Should show: 20 40 60 80 100

# Measure convergence speed at each scale
jq -r '[.round, .nodes, .accuracy] | @tsv' test-results/*/convergence.log | column -t
```

---

## TPM Attestation Results

After test, check TPM verification:

```bash
# View TPM report
cat test-results/*/tpm_attestation.json | jq

# Check trust scores (should be > 75)
curl http://localhost:9090/api/v1/query?query=tpm_node_trust_score | jq '.data.result'

# Verify certificate chain
curl http://localhost:8000/tpm/trust-status | jq
```

**Expected TPM Output:**
```json
{
  "test_name": "incremental_scale_test_...",
  "tpm_enabled": true,
  "attestation_results": [
    {
      "node_id": "node_1",
      "trust_score": 95.2,
      "cert_valid": true,
      "signature_verified": true
    },
    ...
  ],
  "status": "completed"
}
```

---

## NPU/GPU Acceleration

Check NPU metrics:

```bash
# View NPU report
cat test-results/*/npu_metrics.json | jq

# Check if GPU was used
jq '.hardware_info' test-results/*/npu_metrics.json

# Monitor GPU during test (separate terminal)
watch -n 1 nvidia-smi --query-gpu=index,name,utilization.gpu,utilization.memory,memory.used --format=csv,noheader
```

**Expected NPU Output:**
```json
{
  "npu_enabled": true,
  "hardware_info": "NVIDIA Tesla V100",
  "compute_metrics": {
    "total_rounds": 500,
    "final_nodes": 100,
    "gpu_utilization": "85%"
  }
}
```

---

## Prepare for Commit

### Automatic Commit

```bash
# The script handles everything:
bash commit-test-results.sh

# What it does:
# 1. Finds latest test results
# 2. Extracts metrics (accuracy, convergence events)
# 3. Stages all results and scripts
# 4. Creates commit message with metadata
# 5. Shows summary before committing
# 6. Commits with trailers: Assisted-By: cagent
```

### Manual Commit (Alternative)

```bash
# Stage results
git add test-results/incremental_scale_test_YYYYMMDD_HHMMSS/
git add Sovereign_Map_Federated_Learning/test-config.env
git add Sovereign_Map_Federated_Learning/test-incremental-scale.sh

# Create meaningful message
git commit -m "test: incremental scale test - 20→100 nodes, 500 rounds

Test Configuration:
- Initial nodes: 20
- Increment: 20 nodes per convergence
- Max: 100 nodes
- Rounds: 500
- Convergence threshold: 93%
- TPM attestation: enabled
- NPU acceleration: enabled

Results:
- Final accuracy: 94.3%
- Convergence events: 4
- Status: PASSED ✅

Test artifacts: test-results/incremental_scale_test_*/
- TEST_REPORT.md: Executive summary
- convergence.log: Per-round metrics
- metrics.jsonl: System metrics
- tpm_attestation.json: Trust verification
- npu_metrics.json: GPU acceleration data

Assisted-By: cagent" -m "" -m "Test-ID: $(basename test-results/*/)"

# Push
git push origin HEAD
```

---

## Troubleshooting

### Test won't start
```bash
# Check Docker is running
docker ps

# Verify compose file exists
ls docker-compose.large-scale.yml

# Try building manually
docker compose -f docker-compose.large-scale.yml build

# Check system resources
docker system df
```

### Test hanging
```bash
# Check backend health
curl http://localhost:8000/health

# Check docker logs
docker compose -f docker-compose.large-scale.yml logs backend

# Check node agents
docker compose -f docker-compose.large-scale.yml ps node-agent

# Kill and restart
docker compose -f docker-compose.large-scale.yml restart
```

### Low accuracy/not converging
```bash
# Check if nodes are training
docker compose -f docker-compose.large-scale.yml logs node-agent | tail -20

# Verify MongoDB is working
docker exec sovereignmap-mongo mongosh -u admin -p

# Check Prometheus metrics
curl http://localhost:9090/api/v1/query?query=sovereignmap_active_nodes | jq

# View full convergence history
jq . test-results/*/convergence.log | grep accuracy | tail -10
```

### TPM attestation failing
```bash
# Check certificate manager logs
docker exec sovereignmap-backend python -m tpm_cert_manager 2>&1

# Verify certificates exist
docker exec sovereignmap-backend ls -la /etc/sovereign/certs/ 2>/dev/null || echo "Certs not found"

# Regenerate certs
docker compose -f docker-compose.large-scale.yml exec backend bash tpm-bootstrap.sh
```

---

## Test Timeline Estimate

| Phase | Duration | Status |
|-------|----------|--------|
| Initialization | 5 min | Pre-test setup |
| Main loop (500 rounds at 5s each) | ~42 min | Training |
| Result collection | 5 min | Post-test |
| Commit prep | 2 min | Final |
| **Total** | **~54 minutes** | **Full run** |

*Actual time depends on hardware and convergence speed*

---

## Next Steps After Test

1. **Review Results**
   ```bash
   cat test-results/incremental_scale_test_*/TEST_REPORT.md
   ```

2. **Analyze Metrics**
   ```bash
   jq 'select(.accuracy >= 93)' test-results/*/convergence.log
   ```

3. **Compare with Previous Tests**
   ```bash
   ls -lt test-results/incremental_scale_test_*/
   ```

4. **Archive for Analysis**
   ```bash
   tar czf results_backup_$(date +%Y%m%d).tar.gz test-results/
   ```

5. **Share Results**
   ```bash
   # Create summary for stakeholders
   head -30 test-results/*/TEST_REPORT.md
   ```

---

## Support

For issues or questions:

1. Check logs: `cat test-results/*/test.log`
2. View dashboard: `bash test-dashboard.sh`
3. Check backend: `curl http://localhost:8000/convergence | jq`
4. Review docker: `docker compose -f docker-compose.large-scale.yml ps`

---

**Ready to run your test!**

```bash
cd Sovereign_Map_Federated_Learning
bash test-incremental-scale.sh
```

Good luck! 🚀
