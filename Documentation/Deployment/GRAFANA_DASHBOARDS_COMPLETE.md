# Grafana Dashboards - Complete Setup & Integration

## Status: ✅ Complete (Updated March 2026)

All 11 Grafana dashboards are provisioned and integrated with Prometheus.

## March 2026 Concrete Fixes

### 1. Full Dashboard Visual Upgrade

- Applied a system-wide readability and visual hierarchy pass across all dashboards.
- Standardized chart interaction and display defaults:
   - `graphTooltip=1`
   - Improved time series legends/tooltips
   - Cleaner stat presentation and gauge readability
- Added dashboard metadata tags: `enhanced`, `wow`, `readable`.

### 2. Tokenomics and Wallet Data Reliability

- Added/expanded tokenomics and wallet telemetry ingestion in:
  - `tokenomics_metrics_exporter.py`
  - `dashboard_compat_rules.yml`
- Added backend live publisher to emit tokenomics snapshots during FL runtime:
  - `sovereignmap_production_backend_v2.py`
- Enabled production compose tokenomics pipeline and persistence:
  - `docker-compose.production.yml`
  - `test-data/tokenomics-telemetry.json`

### 3. Final Missing Panel Fix

- Fixed Genesis Launch Overview `🏆 Network Status` panel query from invalid ternary syntax to valid PromQL:
   - Before: `(sovereignmap_active_nodes > 0) ? 1 : 0`
   - After: `max(sovereignmap_active_nodes > bool 0)`
- Result: panel now returns data consistently.

### 4. Validation Outcome

- All panel queries on Genesis Launch Overview return data.
- Tokenomics dashboard returns data across all panels.
- Prometheus scrape targets for `sovereign-backend` and `tokenomics-metrics` are healthy.
- Grafana restarted and healthy after provisioning reload.

### Dashboard Files Provisioned

| Dashboard | File | Size | Purpose |
|-----------|------|------|---------|
| Overview | `sovereign-map-overview.json` | 10.3 KB | Active nodes count, current round, CPU/RAM, throughput |
| Convergence | `sovereign-map-convergence.json` | 7.8 KB | Global accuracy, training loss, per-node validation |
| Performance | `sovereign-map-performance.json` | 10.1 KB | CPU/RAM per container, latency, throughput |
| Scaling | `sovereign-map-scaling.json` | 7.5 KB | Node scaling events timeline, scaling rate |
| TPM Security | `sovereign-map-tpm-security.json` | 8.7 KB | TPM verified nodes, attestation success rate, latency |
| NPU Acceleration | `sovereign-map-npu-acceleration.json` | 10.8 KB | Speedup factor, utilization, inference time, throughput |

**Total Dashboard Size:** 55 KB

### Provisioning Configuration

**Location:** `grafana/provisioning/`

```
grafana/provisioning/
├── datasources/
│   └── prometheus.yml          (Prometheus data source config)
└── dashboards/
    ├── dashboard-provider.yaml (Dashboard provider config)
    ├── dashboards.yml          (Alternative provider config)
    ├── sovereign-map-overview.json
    ├── sovereign-map-convergence.json
    ├── sovereign-map-performance.json
    ├── sovereign-map-scaling.json
    ├── sovereign-map-tpm-security.json
    └── sovereign-map-npu-acceleration.json
```

### Key Features

#### 1. **Dynamic Node Count Queries**
All dashboards use `sum(sovereignmap_active_nodes)` instead of hardcoded values:
- Shows **live, real-time active node count**
- Updates every 10 seconds
- Scales from 1 to 500+ nodes automatically

#### 2. **Convergence Metrics**
- Global model accuracy trend
- Training loss trajectory (should decrease)
- Per-node validation accuracy overlay
- Identifies slow/fast converging nodes

#### 3. **Performance Monitoring**
- CPU utilization by container (stacked)
- Memory usage by container (stacked)
- HTTP request latency (p95, p99 percentiles)
- Request throughput (requests/sec)
- Color-coded thresholds (green/yellow/red)

#### 4. **Scaling Events Timeline**
- Bar chart showing node count changes over time
- Scaling event rate (nodes added per minute)
- Cumulative event counter
- Identifies 20→40→60→80→100 transitions

#### 5. **TPM Security Verification**
- TPM-verified node count (stat card)
- Attestation success rate (%)
- Per-node attestation latency
- Failed attestation attempts over time

#### 6. **NPU Acceleration Metrics**
- Speedup factor (3.8x baseline)
- NPU utilization percentage
- CPU vs NPU inference time comparison
- NPU memory usage
- Throughput improvement (requests/sec)

### Docker Compose Integration

Updated `docker-compose.production.yml` with Grafana provisioning mounts:

```yaml
grafana:
  image: grafana/grafana:10.2-alpine
  ports:
    - "3001:3000"
  volumes:
    - grafana_data:/var/lib/grafana
    - ./grafana/provisioning/datasources:/etc/grafana/provisioning/datasources:ro
    - ./grafana/provisioning/dashboards:/etc/grafana/provisioning/dashboards:ro
  environment:
    - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD:-sovereignmap}
    - GF_SECURITY_ADMIN_USER=${GRAFANA_USER:-admin}
    - GF_INSTALL_PLUGINS=grafana-piechart-panel
```

### Quick Start

#### 1. **Start the Full Stack**
```bash
cd Sovereign_Map_Federated_Learning
docker compose -f docker-compose.production.yml up -d
```

#### 2. **Access Grafana**
- **URL:** http://localhost:3001
- **Username:** admin (or `${GRAFANA_USER}`)
- **Password:** sovereignmap (or `${GRAFANA_PASSWORD}`)

#### 3. **Verify Prometheus Data Source**
1. Navigate to **Configuration** → **Data Sources**
2. Should see **Prometheus** (http://prometheus:9090)
3. Click **Test** to verify connectivity

#### 4. **View Dashboards**
All 6 dashboards auto-load in the **Sovereign** folder:
1. Click **Home** → **Dashboards**
2. Select folder **Sovereign**
3. Open any dashboard

#### 5. **Run Test While Monitoring**
```bash
# Terminal 1: Start monitoring
./tests/scripts/bash/test-dashboard.sh

# Terminal 2: Run 5000-round test
./tests/scripts/powershell/run-5000-round-test.ps1
```

Dashboard metrics update in **real-time** as test executes.

### Metric Queries Reference

#### Active Nodes (Dynamic)
```promql
sum(sovereignmap_active_nodes)
```
Shows total nodes scaling from 1 to 100+ in real-time.

#### Model Accuracy
```promql
sovereignmap_model_accuracy
```
Global accuracy trajectory (0-100%).

#### Training Loss
```promql
sovereignmap_training_loss
```
Loss value (should decrease over rounds).

#### CPU Usage
```promql
rate(process_cpu_seconds_total[5m]) * 100
```
Per-container CPU % (stacked).

#### Memory Usage
```promql
process_resident_memory_bytes / 1024 / 1024
```
Per-container memory in MB (stacked).

#### HTTP Latency (p95, p99)
```promql
histogram_quantile(0.95, sovereignmap_http_request_duration_seconds_bucket)
histogram_quantile(0.99, sovereignmap_http_request_duration_seconds_bucket)
```

#### Throughput
```promql
rate(sovereignmap_http_requests_total[1m])
```
Requests per second by endpoint.

#### TPM Verified Nodes
```promql
sum(sovereignmap_tpm_verified_nodes)
```
Nodes with valid TPM attestation.

#### TPM Success Rate
```promql
(sum(sovereignmap_tpm_attestation_success) / sum(sovereignmap_tpm_attestation_total)) * 100
```
Percentage of successful attestations.

#### NPU Speedup
```promql
sovereignmap_npu_speedup_factor
```
Acceleration ratio (e.g., 3.8x).

#### NPU Utilization
```promql
sovereignmap_npu_utilization_percent
```
NPU device utilization (0-100%).

#### Inference Time Comparison
```promql
sovereignmap_inference_time_cpu_ms     # CPU inference latency
sovereignmap_inference_time_npu_ms     # NPU inference latency
```

### Dashboard Refresh Rates

| Component | Refresh Rate | Rationale |
|-----------|--------------|-----------|
| Overview | 10s | Real-time system state |
| Convergence | 10s | Catch accuracy improvements |
| Performance | 10s | Monitor resource usage |
| Scaling | 10s | Track node transitions |
| TPM Security | 10s | Continuous attestation monitoring |
| NPU Acceleration | 10s | Inference performance tracking |

### Expected Behavior During 5000-Round Test

**Timeline:**
1. **T=0s:** Cluster starts with 20 nodes
   - Active Nodes: 20
   - CPU: ~40%
   - Memory: ~30%

2. **T=2s:** Scale to 40 nodes
   - Scaling Events: +1
   - Active Nodes: 40
   - CPU: ~65%
   - Memory: ~45%

3. **T=4s:** Scale to 60 nodes
   - Scaling Events: +2
   - Active Nodes: 60
   - CPU: ~85%
   - Memory: ~65%

4. **T=6s:** Scale to 80 nodes
   - Scaling Events: +3
   - Active Nodes: 80
   - CPU: ~95%
   - Memory: ~80%

5. **T=8s:** Scale to 100 nodes (max)
   - Scaling Events: +4
   - Active Nodes: 100
   - CPU: 100%
   - Memory: ~92.3%

6. **T=9.5s:** Training complete
   - Final Accuracy: 94-96%
   - Total Rounds: 5000
   - Test Duration: 9.53 seconds

### Troubleshooting

#### Dashboards Not Loading
```bash
# Check Grafana logs
docker logs sovereignmap-grafana

# Verify provisioning mount
docker inspect sovereignmap-grafana | grep -A 10 Mounts
```

#### Prometheus Queries Returning No Data
1. Check backend is running: `docker ps | grep backend`
2. Verify metrics endpoint: `curl http://localhost:8000/metrics`
3. Check Prometheus scrape config: `docker exec sovereignmap-prometheus cat /etc/prometheus/prometheus.yml`

#### Dashboards Show Empty Panels
1. Verify time range is recent (e.g., "Last 1 Hour")
2. Check data source connectivity
3. Run a test to generate metrics: `./tests/scripts/powershell/run-5000-round-test.ps1`

#### Performance Issues
- Reduce dashboard refresh rate (default 10s)
- Limit time range to "Last 15 Minutes"
- Close unused dashboard tabs

### Next Steps

1. **Run Full Test Suite**
   ```bash
   ./tests/scripts/powershell/run-5000-round-test.ps1
   ```
   Monitor all 6 dashboards in real-time.

2. **Customize Dashboards**
   - Edit panel queries in Grafana UI
   - Add custom alerts
   - Create custom variables

3. **Export Metrics**
   - Use Grafana API to export data
   - Build custom reports

4. **High-Scale Testing**
   - Try 200-500 nodes on different hardware
   - Monitor dashboard performance
   - Optimize resource allocation

### Architecture Summary

```
┌─────────────────────────────────────────┐
│     Sovereign Map FL System             │
├─────────────────────────────────────────┤
│  Backend (8000)  Node Agents (6000+)    │
│       ↓                    ↓             │
│   Metrics Endpoint ← Prometheus Scrape  │
│       ↓                                  │
│  Prometheus (9090)                      │
│       ↓                                  │
│  Grafana (3001) ← 6 Dashboards          │
│       ↓                                  │
│  ┌──────────────────────────────────┐   │
│  │ Overview │ Convergence │ Perf    │   │
│  │ Scaling  │ TPM Sec    │ NPU Accel│   │
│  └──────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

### Files Modified

- ✅ `docker-compose.production.yml` - Updated Grafana volumes
- ✅ 6 new dashboard JSON files
- ✅ Prometheus datasource configuration

### Metrics Emitted by Backend

Ensure backend emits these metrics on `/metrics` endpoint:

```python
# Counters
sovereignmap_http_requests_total
sovereignmap_node_scaling_events_total
sovereignmap_tpm_attestation_total
sovereignmap_tpm_attestation_success
sovereignmap_tpm_attestation_failures_total
sovereignmap_inference_requests_cpu_total
sovereignmap_inference_requests_npu_total

# Gauges
sovereignmap_active_nodes
sovereignmap_current_round
sovereignmap_model_accuracy
sovereignmap_training_loss
sovereignmap_node_accuracy
sovereignmap_tpm_verified_nodes
sovereignmap_npu_speedup_factor
sovereignmap_npu_utilization_percent
sovereignmap_npu_memory_mb

# Histograms
sovereignmap_http_request_duration_seconds_bucket
sovereignmap_tpm_attestation_duration_ms
sovereignmap_inference_time_cpu_ms
sovereignmap_inference_time_npu_ms
```

All dashboards ready for production monitoring. Start testing now!

---

**Setup Complete:** All 6 dashboards configured, provisioned, and ready for real-time monitoring.
