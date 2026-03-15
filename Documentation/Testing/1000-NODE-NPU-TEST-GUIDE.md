# 1000-Node NPU Performance Test - Quick Start Guide

## Overview
Complete automated test suite for validating NPU effectiveness across 1000 federated learning nodes with comprehensive metrics collection, visualization, and clean git commit.

## Prerequisites
- Docker & Docker Compose v2.0+
- 16+ GB RAM available
- 100+ GB disk space for test artifacts
- Git configured for commits/pushes
- Python 3.8+ (for plot generation)
- ~30-45 minutes for full test suite

## Quick Start

### 1. Start the Test (Linux/Mac)
```bash
cd Sovereign_Map_Federated_Learning
bash tests/scripts/bash/run-1000-node-npu-test.sh
```

### 2. Start the Test (Windows)
```powershell
cd Sovereign_Map_Federated_Learning
powershell -ExecutionPolicy Bypass -File tests/scripts/bash/run-1000-node-npu-test.sh
```

## Test Phases

### Phase 1: Environment Setup
- Docker availability check
- `.env` file initialization
- Credentials and passwords setup

### Phase 2: Docker Build (5-10 min)
- **Backend**: Optimized multi-stage Dockerfile
- **Frontend**: Nginx-based with React/Vue SPA
- **Node-Agent**: FL client with NPU support

### Phase 3: Infrastructure Deployment (3-5 min)
- MongoDB (1GB+ cache, sharded config for 1000 nodes)
- Redis (4GB cache, LRU eviction policy)
- Backend aggregator (8 worker threads, NPU-enabled)
- Frontend visualization server

### Phase 4: Monitoring Stack (2-3 min)
- **Prometheus**: High-cardinality metrics (1000 node targets)
- **Grafana**: Real-time dashboards (3001 port)
- **AlertManager**: Alert routing and aggregation

### Phase 5: Node Agent Deployment (10-15 min)
- Scales to **1000 node agent containers**
- Each with health checks and NPU support
- Sequential startup with connection pooling

### Phase 6: NPU Spectrum Testing (15-20 min)

#### Test 1: CPU Baseline
- No NPU acceleration
- Measures pure CPU performance
- Reference point for comparison

#### Test 2: NPU Acceleration
- Full NPU acceleration enabled
- Direct comparison with baseline
- Performance uplift quantification

#### Test 3: Throughput Testing
- 1000 concurrent update requests
- Latency percentiles (p50, p75, p90, p95, p99)
- Saturation point identification

#### Test 4: Byzantine Fault Tolerance
- 1% Byzantine nodes (10 out of 1000)
- Consensus correctness validation
- Malicious node detection

#### Test 5: Consensus Efficiency
- Message count per round
- Round duration analysis
- Network overhead metrics

### Phase 7: Metrics Collection (5-10 min)
- Prometheus query snapshots
- Container log export
- Test result JSON files

### Phase 8: Plot Generation (5 min)
Automatic generation of **3 detailed visualization plots**:

1. **01-npu-performance-analysis.png**
   - Throughput comparison (CPU vs NPU)
   - Latency comparison
   - Resource utilization (CPU, memory, GPU)
   - Summary metrics table

2. **02-throughput-latency-analysis.png**
   - Latency histogram (50 bins)
   - Cumulative Distribution Function (CDF)
   - Percentile breakdown (p50-p99)
   - Comprehensive statistics

3. **03-consensus-efficiency-analysis.png**
   - Messages per consensus round
   - Consensus success rate (pie chart)
   - Round duration timeline
   - Byzantine resilience metrics

### Phase 9: Report Generation (1 min)
- Markdown executive summary
- Test environment documentation
- Results overview
- Recommendations

### Phase 10: Git Commit & Push (1-2 min)
**Automatic clean commit** with:
- All test artifacts
- Visualization plots
- Metrics JSON files
- Complete logs
- Executive report

Commit message includes:
- Detailed infrastructure setup
- All test phases performed
- Key performance metrics
- Complete artifact list
- Timestamp and tracking

## Output Structure

```
test-results/1000-node-npu/
├── TIMESTAMP/
│   ├── artifacts/
│   │   ├── prometheus-*.json          (Prometheus snapshots)
│   │   ├── npu_*.json                 (NPU benchmarks)
│   │   ├── throughput_test.json       (Latency data)
│   │   ├── consensus_efficiency.json  (Consensus metrics)
│   │   └── bft_test_*.json            (Byzantine test results)
│   ├── plots/
│   │   ├── 01-npu-performance-analysis.png
│   │   ├── 02-throughput-latency-analysis.png
│   │   └── 03-consensus-efficiency-analysis.png
│   ├── logs/
│   │   ├── build-*.log                (Docker builds)
│   │   ├── deploy-*.log               (Service deployments)
│   │   ├── test-*.log                 (Individual test logs)
│   │   └── *-full.log                 (Container runtime logs)
│   └── TEST-REPORT.md                 (Executive summary)
└── artifacts.tar.gz                   (Complete compressed package)
```

## Accessing Results

### 1. Live Dashboards (During Test)
```
🌐 Frontend:     http://localhost:3000
📊 Grafana:      http://localhost:3001    (admin/<configured password>)
📈 Prometheus:   http://localhost:9090
🚨 AlertManager: http://localhost:9093
```

### 2. Test Results Files
```bash
# View all results
ls -lh test-results/1000-node-npu/

# View latest test
cd test-results/1000-node-npu/$(ls -t | head -1)

# View plots
open plots/*.png

# View report
cat TEST-REPORT.md

# View metrics
cat artifacts/throughput_test.json | jq .
```

### 3. Git History
```bash
# View commit
git log --oneline -5

# View commit details
git show HEAD

# View changes
git diff HEAD~1
```

## Monitoring During Test

### Terminal Output
```
✅ PHASE 2: Docker Build
  • Building backend (optimized)...
  • Building frontend (optimized)...
  • Building node-agent...
✅ All images built successfully

✅ PHASE 3: Infrastructure Deployment
  • Starting MongoDB, Redis, Backend, Frontend...
  • Waiting for MongoDB to be healthy...
    ✓ MongoDB healthy (attempt 5/60)
  • Waiting for Backend to be healthy...
    ✓ Backend healthy (attempt 12/60)
  ✓ Frontend healthy (attempt 8/60)

✅ PHASE 5: Node Agent Deployment (1000 Replicas)
  • Scaling node-agent to 1000 replicas...
  • Waiting for nodes to initialize...
    [1/12] Active containers: 250 (target: 1000)
    [2/12] Active containers: 500 (target: 1000)
    [3/12] Active containers: 750 (target: 1000)
    [4/12] Active containers: 1000 (target: 1000)
```

### Docker Compose Status
```bash
# Watch container status
docker compose -f docker-compose.1000nodes.yml ps

# View logs
docker logs -f sovereignmap-backend-1000
docker logs -f prometheus-1000
```

## Troubleshooting

### Docker Out of Space
```bash
# Clean up old containers and images
docker system prune -af --volumes
docker image prune -af
```

### Memory Issues
```bash
# Check Docker memory allocation
docker system df

# Increase Docker Desktop memory limit
# Docker Desktop > Settings > Resources > Memory: Set to 16GB+
```

### Build Failures
```bash
# Check build logs
tail -f test-results/1000-node-npu/*/logs/build-*.log

# Rebuild specific image
docker build -f Dockerfile.backend.optimized -t sovereignmap/backend:1000-test .
```

### Node Agent Deployment Issues
```bash
# Check node-agent logs
docker logs $(docker ps --filter "name=node-agent" --format "{{.ID}}" | head -1)

# Check Docker network
docker network inspect sovereignmap-1000
```

## Performance Expectations

### Hardware Requirements
| Resource | Requirement | Recommended |
|----------|-------------|------------|
| CPU Cores | 8+ | 16+ |
| RAM | 16 GB | 32 GB |
| Disk | 100 GB free | 200 GB |
| Network | 1 Gbps | 10 Gbps |

### Typical Performance
| Metric | Baseline | With NPU |
|--------|----------|---------|
| Throughput | 500-800 RPS | 2000-4000 RPS |
| Avg Latency | 50-100ms | 20-40ms |
| CPU Usage | 70-90% | 40-60% |
| Memory Usage | 8-12 GB | 10-14 GB |

### Test Duration
- **Total Time**: 30-45 minutes
- **Infrastructure Setup**: ~10 min
- **Node Deployment**: ~15 min
- **Test Execution**: ~10 min
- **Result Processing**: ~5-10 min

## Advanced Options

### Custom Node Count
Edit `docker-compose.1000nodes.yml`:
```yaml
node-agent:
  deploy:
    replicas: 500  # Change from 1000 to 500
```

### Extended Test Duration
Edit `tests/scripts/bash/run-1000-node-npu-test.sh`:
```bash
NUM_ROUNDS=100  # Increase consensus rounds
```

### Custom Metrics
Add custom Prometheus queries in `/monitoring/prometheus.1000.yml`:
```yaml
- job_name: 'custom-metrics'
  static_configs:
    - targets: ['localhost:8000']
```

## Clean Up

### Stop Running Services
```bash
docker compose -f docker-compose.1000nodes.yml down
```

### Remove All Test Data
```bash
docker compose -f docker-compose.1000nodes.yml down -v
rm -rf test-results/1000-node-npu/
```

## Support & Issues

For issues or questions:
1. Check logs: `test-results/1000-node-npu/*/logs/`
2. Review TEST-REPORT.md
3. Check Prometheus metrics at http://localhost:9090
4. View Grafana dashboards at http://localhost:3001

## Next Steps

After test completion:
1. Review plots in `plots/` directory
2. Analyze metrics in `artifacts/` directory
3. Check report: `TEST-REPORT.md`
4. View git commit: `git log --oneline`
5. Integrate results into analysis pipeline
6. Archive test artifacts

---

**Last Updated**: 2026-03-03
**Test Version**: 1000-Node NPU Performance v1.0
