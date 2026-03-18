# Phase 3A: Performance Monitoring Dashboard - Operational Guide

**Document Version**: 1.0  
**Last Updated**: 2026-03-18  
**Status**: Ready for Deployment  
**Audience**: Operations Team, DevOps, SRE  

---

## 1. Overview

The Performance Monitoring Dashboard provides real-time visibility into Sovereign Map Federated Learning system health across five critical dimensions:

1. **GPU Acceleration** - Hardware utilization, memory, temperature, latency
2. **Privacy Metrics** - Budget consumption, noise injection, overhead tracking
3. **Consensus & Byzantine** - Round metrics, node participation, Byzantine detection
4. **Network Partition Recovery** - Partition detection, recovery time, attestation validation
5. **System Health** - CPU, memory, network, process uptime

### Key Features

- **Real-time Metrics**: 50+ metrics updated every 10-30 seconds
- **Prometheus Integration**: Standard text-format export for Prometheus scraping
- **5 Pre-built Dashboards**: Grafana JSON configs for rapid deployment
- **15+ Alert Rules**: Automatic escalation for critical conditions
- **SLA Compliance Tracking**: Built-in compliance checks (5 metrics)
- **Health Probes**: Kubernetes-compatible liveness/readiness checks

### Architecture

```
┌─────────────────────────────────────────────────────┐
│         Sovereign Map Modules                        │
│  ┌────────────┐ ┌────────────┐ ┌──────────────┐   │
│  │  Privacy   │ │ Consensus  │ │   Network    │   │
│  │   Engine   │ │  Module    │ │  Partition   │   │
│  └─────┬──────┘ └─────┬──────┘ └──────┬───────┘   │
│        │               │               │            │
│        └───────────────┼───────────────┘            │
│                        ▼                            │
│        ┌──────────────────────────────┐            │
│        │  Monitoring Orchestrator      │            │
│        │  (Integration Hub)            │            │
│        └──────────────┬─────────────────┘            │
└───────────────────────┼──────────────────────────────┘
                        ▼
┌─────────────────────────────────────────────────────┐
│         Metrics Registry (50+ metrics)              │
│  ┌──────┬──────┬────────┬─────┬────────────────┐  │
│  │ GPU  │Privcy│Consens │Netw │ System         │  │
│  │   10 │   9  │    8   │ 7   │   6  metrics   │  │
│  └──────┴──────┴────────┴─────┴────────────────┘  │
└──────────────┬───────────────────────────────────┘
               ▼
┌─────────────────────────────────────────────────────┐
│      Prometheus HTTP Server (Port 9090)             │
│  ┌──────────────┬──────────────┬─────────────────┐│
│  │ /metrics     │ /health      │ /sla            ││
│  │ (Scraping)   │ (Liveness)   │ (Compliance)    ││
│  └──────────────┴──────────────┴─────────────────┘│
└──────────────┬───────────────────────────────────┘
               │
       ┌───────┴────────┬──────────────┐
       ▼                ▼              ▼
   Prometheus      Grafana         Alertmanager
   (Collect)    (Visualize)      (Escalate)
```

---

## 2. Quick Start

### 2.1 Prerequisites

```bash
# Required: Prometheus
curl https://github.com/prometheus/prometheus/releases/download/v2.40.0/prometheus-2.40.0.linux-amd64.tar.gz -L | tar xz
cd prometheus-2.40.0.linux-amd64
./prometheus --config.file=prometheus.yml

# Optional but Recommended: Grafana
docker run -d --name grafana -p 3000:3000 grafana/grafana:latest

# Optional: Alertmanager for escalation
curl https://github.com/prometheus/alertmanager/releases/download/v0.24.0/alertmanager-0.24.0.linux-amd64.tar.gz -L | tar xz
cd alertmanager-0.24.0.linux-amd64
./alertmanager

# Check Node.js and npm (for Sovereign Map)
node --version  # require v16+
npm --version   # require v7+
```

### 2.2 Enable Monitoring in Sovereign Map

```typescript
// In your main initialization file (e.g., src/index.ts)

import { MonitoringOrchestrator } from '@sovereign-map/monitoring';
import { PrivacyEngine } from '@sovereign-map/privacy';
import { ConsensusModule } from '@sovereign-map/consensus';
import { PartitionRecoveryManager } from '@sovereign-map/consensus';

// Initialize monitoring
const monitoring = new MonitoringOrchestrator(9090); // Port 9090 (configurable)
await monitoring.initialize();

// Integrate with modules
const privacyEngine = new PrivacyEngine();
const consensusModule = new ConsensusModule();
const partitionRecovery = new PartitionRecoveryManager();

monitoring.integratePrivacy(privacyEngine);
monitoring.integrateConsensus(consensusModule);
monitoring.integrateNetwork(partitionRecovery);

// Metrics now flow to http://localhost:9090/metrics
```

### 2.3 Configure Prometheus

Add to `prometheus.yml`:

```yaml
global:
  scrape_interval: 30s
  evaluation_interval: 30s
  external_labels:
    cluster: 'sovereign-map-prod'

scrape_configs:
  - job_name: 'sovereign-map'
    static_configs:
      - targets: ['localhost:9090']  # Sovereign Map metrics port
    scrape_interval: 30s
    scrape_timeout: 10s

  - job_name: prometheus
    static_configs:
      - targets: ['localhost:9090']

rule_files:
  - 'prometheus-alert-rules.yml'

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['localhost:9093']  # Alertmanager
```

### 2.4 Import Grafana Dashboards

1. Open Grafana: http://localhost:3000 (admin/admin)
2. Data Sources → Add Prometheus → http://localhost:9090
3. Dashboards → Import → Paste JSON from:
   - `grafana-dashboards.GPU_DASHBOARD`
   - `grafana-dashboards.PRIVACY_DASHBOARD`
   - `grafana-dashboards.CONSENSUS_DASHBOARD`
   - `grafana-dashboards.NETWORK_DASHBOARD`
   - `grafana-dashboards.SYSTEM_DASHBOARD`
   - `grafana-dashboards.SLA_DASHBOARD`

4. Click "Import" for each dashboard

✅ **You now have real-time monitoring!**

---

## 3. Dashboard Walkthroughs

### 3.1 GPU Acceleration Dashboard

**Purpose**: Monitor hardware acceleration health and performance.

**Key Panels**:

| Panel | Metric | Target | Green | Yellow | Red |
|-------|--------|--------|-------|--------|-----|
| GPU Utilization | `gpu_utilization_percent` | "GPU is being used" | >50% | 20-50% | <20% |
| GPU Memory | `gpu_memory_used_bytes` | Memory health | <70% | 70-90% | >90% |
| Noise Latency | `gpu_noise_latency_seconds` | Generation speed | <10ms | 10-50ms | >100ms |
| Active Devices | `gpu_active_devices` | Type of GPU | CUDA/ROCm | WebGPU | CPU |
| Failure Rate | GPU failures | Detection success | <1% | 1-5% | >5% |

**Interpretation Guide**:

- **All Panels Green**: GPU acceleration operating normally
- **Latency Yellow**: GPU may be thermal throttling or memory-pressured
- **Memory Red**: Reduce batch size, consider multi-GPU deployment
- **Failure Rate Rising**: Check GPU drivers, CUDA/ROCm version compatibility
- **Active Devices = CPU**: GPU detection failed, check hardware availability

**Action Items by Scenario**:

| Condition | Root Cause | Action |
|-----------|-----------|--------|
| Util <10%, Latency Normal | GPU under-utilized | Increase batch size for better throughput |
| Util >95%, Latency > 50ms | GPU saturation | Reduce batch size or scale horizontally |
| Memory >90%, Util Flat | Memory leak | Review GPU buffer pool cleanup |
| Failure Rate >5% | Driver issue | Reinstall GPU drivers, verify CUDA compatibility |
| All metrics present | Graceful fallback active | GPU may have failed temporarily; monitor recovery |

**Kubernetes Health Check**:

```bash
# Check GPU pod status
kubectl get pods -l app=sovereign-map -o wide

# Check GPU utilization
kubectl exec -it <pod-name> -- curl http://localhost:9090/metrics | grep gpu_

# Check health probe
kubectl exec -it <pod-name> -- curl http://localhost:9090/health
```

---

### 3.2 Privacy Budget Monitor

**Purpose**: Track epsilon/delta consumption and privacy overhead.

**Key Panels**:

| Panel | Metric | Target | Green | Yellow | Red |
|-------|--------|--------|-------|--------|-----|
| Epsilon Remaining | `privacy_epsilon_remaining` | Budget health | >0.5 | 0.1-0.5 | <0.1 |
| Privacy Overhead | `privacy_overhead_percent` | Perf impact | <12% | 12-20% | >20% |
| Noise Rate | `privacy_noise_injected_total` | Injection rate | Steady | Increasing | Spiking |
| Overhead CPU vs GPU | Ratio | GPU benefits | <12% GPU | >50% | CPU only |

**Interpretation Guide**:

- **Epsilon >0.5**: Healthy budget, many updates remaining
- **Epsilon 0.1-0.5**: Yellow flag, ~10-100 more updates, consider raising epsilon for future rounds
- **Epsilon <0.1**: Critical, deploy with larger epsilon next round
- **Overhead <12% with GPU**: Spec claim confirmed, GPU acceleration working
- **Overhead >20%**: Either CPU baseline or GPU not active, investigate gpu panel

**Privacy Budget Calculation**:

```
Epsilon per update = Total epsilon / Number of updates
Example: ε=1.0 with 100 updates = 0.01 ε per update

Remaining = Initial - Consumed
Monitor for: Consumed rate to avoid exhaustion
```

**Action Items by Scenario**:

| Condition | Root Cause | Action |
|-----------|-----------|--------|
| Epsilon declining slowly | Normal consumption | Continue monitoring, plan for next epoch |
| Epsilon declining rapidly | Frequent updates | Reduce update frequency or batch larger |
| Overhead 12-20% on GPU | Noise generation expensive | Optimize for larger batch sizes |
| Overhead >20% | GPU fallback or CPU | Check gpu-acceleration.ts for errors |
| Noise injection failures | Memory pressure | Reduce noise dimensions or batch size |

**Privacy Compliance Checklist**:

- [ ] Epsilon remaining > 0.1 (not exhausted)
- [ ] Overhead < 12% with GPU (spec requirement)
- [ ] Noise injection success rate > 99%
- [ ] Delta < 1e-5 (default, acceptable risk)
- [ ] No injection failures in last hour

---

### 3.3 Consensus & Byzantine Detection

**Purpose**: Monitor BFT consensus health and Byzantine attack detection.

**Key Panels**:

| Panel | Metric | Target | Green | Yellow | Red |
|-------|--------|--------|-------|--------|-----|
| Round Time | `consensus_round_duration_seconds` | Latency | <10s | 10-30s | >30s |
| Participation | `consensus_participation_rate` | Node sync | >75% | 66-75% | <66% |
| Nodes Online | `consensus_nodes_online` | Cluster health | >95% | 80-95% | <80% |
| Byzantine Count | `byzantine_nodes_detected` | Attack detection | 0 | 1 | ≥2 |

**Interpretation Guide**:

- **Participation >75%**: Healthy consensus, quorum easily reached
- **Participation 66-75%**: Partition risk, watch for network issues
- **Participation <66%**: Byzantine quorum loss, critical
- **1 Byzantine node**: Monitor, may recover, investigate logs
- **≥2 Byzantine nodes**: Potential attack, escalate immediately
- **Round time >30s**: Network latency or Byzantine disputes, investigate

**Byzantine Tolerance Math**:

```
Safe with N nodes and B Byzantine:
- Tolerance: B < N/3
- Example: 30 nodes → tolerance is 9 Byzantine
- Example: 300 nodes → tolerance is 99 Byzantine

Participation requirement:
- Min: 2/3 (66.7%) + 1
- At 100 nodes: need ≥67 online
- At 300 nodes: need ≥201 online
```

**Action Items by Scenario**:

| Condition | Root Cause | Action |
|-----------|-----------|--------|
| Round time increasing gradually | Consensus disputes | Check for Byzantine nodes |
| Round time sudden spike | Byzantine attack | Review Byzantine panel; isolate suspected nodes |
| Participation dropping | Network partition | Check network connectivity; may need partition recovery |
| 1 Byzantine detected | Faulty node or attack | Monitor for consistency; may auto-recover |
| ≥2 Byzantine detected | Coordinated attack | CRITICAL: Page SRE, review logs, quarantine attackers |

**Consensus Health Checklist**:

- [ ] Participation rate > 66.7% (quorum requirement)
- [ ] Round time < 10s (normal, <30s acceptable)
- [ ] Byzantine count < 3 (for 100-node cluster)
- [ ] No recent timeout events
- [ ] All online nodes reporting rounds

---

### 3.4 Network Partition & Recovery

**Purpose**: Track network partition detection and safe recovery.

**Key Panels**:

| Panel | Metric | Target | Green | Yellow | Red |
|-------|--------|--------|-------|--------|-----|
| Partitions Detected | `network_partitions_detected_total` | Event count | 0 | 1-5/day | >5/day |
| Isolated Nodes | `network_partition_isolated_nodes` | Current size | 0 | 1-10 | >10 |
| Recovery Time | `network_partition_recovery_seconds` | Speed | <50s | 50-120s | >120s |
| Success Rate | `recovery_success / total` | Reliability | >99% | 90-99% | <90% |

**Interpretation Guide**:

- **Partitions 0**: No recent partitions (healthy)
- **Partitions 1-5/day**: Normal for WAN deployments
- **Partitions >5/day**: Unstable network, check ISP/routing
- **Isolated 0**: No active partition
- **Isolated 1-10**: Recovery in progress
- **Isolated >10**: Large partition, recovery delayed
- **Recovery <50s**: Excellent design
- **Recovery 50-120s**: Normal for consensus sync
- **Recovery >120s**: Check clock synchronization, Byzantine disputes

**Recovery State Machine**:

```
HEALTHY
   ▼
PARTITION_DETECTED (heartbeat + consensus + timestamp all agree)
   ▼
ISOLATED (node aware of partition, stops consensus)
   ▼
SYNCHRONIZING (sync blocks with majority)
   ▼
RECOVERING (validate attestation chain)
   ▼
HEALTHY (rejoin cluster)
```

**Action Items by Scenario**:

| Condition | Root Cause | Action |
|-----------|-----------|--------|
| Isolated=0, Recovery=0 | Normal state | OK, no action |
| Isolated>0, Recovery increasing | Slow recovery | Check clock sync (NTP), network latency |
| Isolated>0, Recovery stuck >120s | Byzantine dispute | Review Byzantine nodes; may need manual failover |
| Success rate <90% | Repeated failures | Check attestation chain validation, NTP drift |
| Partitions >5/day | Network instability | Investigate ISP reliability, BGP flapping |
| Isolated>cluster_size/3 | Partition recovery impossible | CRITICAL: Manual intervention needed |

**Network Partition Checklist**:

- [ ] Isolated nodes = 0 (not currently partitioned)
- [ ] Recovery time < 2 minutes when partitions occur
- [ ] Success rate > 99%
- [ ] NTP synchronized across all nodes (clock skew < 100ms)
- [ ] No stuck Byzantine recovery states

---

### 3.5 System Health & Resource Usage

**Purpose**: Monitor infrastructure resource consumption.

**Key Panels**:

| Panel | Metric | Target | Green | Yellow | Red |
|-------|--------|--------|-------|--------|-----|
| Memory Usage | `system_memory_bytes` | Heap | <70% | 70-90% | >90% |
| CPU Usage | `system_cpu_usage_percent` | Utilization | 20-60% | 60-90% | >90% |
| Network Latency | `system_network_latency_ms` | P99 | <50ms | 50-100ms | >100ms |
| Process Uptime | `system_uptime_seconds` | Duration | Days | Hours | Minutes |
| Restarts | `system_restarts` | Stability | 0 | 1-3/day | >3/day |

**Interpretation Guide**:

- **Memory <70%**: Healthy, room for growth
- **Memory 70-90%**: Monitor, verify no leaks, plan upgrade
- **Memory >90%**: Risk of OOM kill, reduce load or scale out
- **CPU 20-60%**: Optimal, headroom available
- **CPU 60-90%**: Acceptable, monitor threshold
- **CPU >90%**: Potential CPU bottleneck, profile and optimize
- **Latency <50ms**: Excellent network
- **Latency 50-100ms**: Expected WAN latency
- **Latency >100ms**: Network issue or saturation
- **Uptime days**: Healthy, expected for production
- **Uptime hours**: Frequent restarts, check logs
- **Restarts 0**: Perfect stability
- **Restarts >3/day**: Multiple crashes, investigate cause

**Resource Scaling Guidance**:

```
Single Node Requirements (per 100M samples/sec):
- Memory: 2GB base + 1GB per 100M samples = ~3-4GB typical
- CPU: 1 core base + 1 core per 200M samples = ~1-2 cores typical
- GPU: 1× GPU (6-12GB VRAM typical)
- Network: 100Mbps base + 50Mbps per 1000 consensus rounds/min

Scaling (10-node cluster):
- Memory: 3-4GB per node = 30-40GB total
- CPU: 1-2 cores per node × 10 = 10-20 cores total
- GPU: 1 GPU per node × 10 = 10 GPUs
- Network: 1Gbps cluster interconnect recommended
```

**Action Items by Scenario**:

| Condition | Root Cause | Action |
|-----------|-----------|--------|
| Memory rising linearly | Memory leak | Profile with node --inspect, fix leak |
| Memory spike then drop | GC pause | Normal; if >1s pause, tune heap |
| CPU consistently 90%+ | CPU-bound bottleneck | Profile with node --prof, optimize hot path |
| Latency increasing | Network congestion | Check bandwidth, consider QoS tuning |
| Uptime < 1 hour | Frequent crash | Check logs for OOM, unhandled exceptions |
| Restarts >3/day | Stability issue | Enable core dumps: `ulimit -c unlimited` |

**Resource Health Checklist**:

- [ ] Memory usage stable, no leaking trend
- [ ] CPU 20-80% range (not too idle, not saturated)
- [ ] Network latency < 100ms p99
- [ ] Uptime > 7 days (no unexpected restarts)
- [ ] System logs clean (no OOM, core dumps, uncaught exceptions)

---

### 3.6 SLA Compliance Monitor

**Purpose**: Real-time SLA metric tracking against defined targets.

**SLA Thresholds**:

| Metric | Target | Consequence |
|--------|--------|-------------|
| Privacy Overhead | <12% (with GPU) | Spec violation if exceeded |
| GPU Detection Success | >99% | No GPU acceleration benefit |
| Consensus Participation | >66.7% | Quorum loss risk |
| Network Latency | <100ms p99 | Consensus delays |
| Byzantine Tolerance | <33% (N/3 rule) | Attack risk |

**Color Codes**:

- 🟢 **Green**: All 5 SLAs passing, production healthy
- 🟡 **Yellow**: 1-2 SLAs at risk, monitor closely
- 🔴 **Red**: ≥3 SLAs failing, escalate immediately

**Interpretation Guide**:

- **All Green**: Healthy production state, no action needed
- **1 Yellow**: Minor issue, investigate root cause
- **2 Yellow**: Multiple issues, start incident response
- **Any Red**: Critical failure, page SRE immediately

**Action Items by Scenario**:

| Condition | Root Cause | Action |
|-----------|-----------|--------|
| All Green | System healthy | Continue normal ops |
| Privacy Yellow | CPU overhead high | Enable GPU acceleration check drivers |
| GPU Yellow | Hardware unavailable | Verify GPU presence, driver version |
| Participation Yellow | Node connectivity | Check network, restart failing nodes |
| Latency Yellow | Network congestion | Monitor BWutilization, adjust QoS |
| Byzantine Red | Attack detected | Page security, isolate compromised nodes |
| Multiple Red | Cascading failure | Initiate disaster recovery procedures |

**SLA Escalation Path**:

```
Green (No action)
   ▼
Yellow (Team lead notified, 15min window to respond)
   ▼
Red (Page on-call engineer immediately)
   ▼
Multiple Red (Page SRE manager, initiate incident command)
   ▼
Persistent Red (Executive escalation, declare SEV-1 incident)
```

---

## 4. Alert Rules Reference

### 4.1 Critical Alerts (Immediate Page)

```
🔴 ByzantineNodesDetected
   - Trigger: Byzantine count ≥ 1
   - Action: Review logs, investigate attack vector
   - Runbook: docs/BYZANTINE_INCIDENT_RESPONSE.md

🔴 PartitionPersists
   - Trigger: Partition isolation > 5 minutes
   - Action: Check network connectivity, may indicate ISP outage
   - Runbook: docs/NETWORK_INCIDENT_RESPONSE.md

🔴 LowConsensusParticipation
   - Trigger: Participation < 66.7%
   - Action: Investigate node failures, restart unresponsive nodes
   - Runbook: docs/CONSENSUS_INCIDENT_RESPONSE.md

🔴 GPUMemoryExhaustion
   - Trigger: GPU memory > 90%
   - Action: Reduce batch size, enable memory swapping
   - Runbook: docs/GPU_INCIDENT_RESPONSE.md

🔴 PrivacyBudgetExhausted
   - Trigger: Epsilon remaining < 0.1
   - Action: Deploy with larger epsilon or pause updates
   - Runbook: docs/PRIVACY_INCIDENT_RESPONSE.md

🔴 ProcessDown
   - Trigger: Pod/process not responding 1+ minute
   - Action: Check logs, restart process
   - Runbook: docs/PROCESS_RECOVERY.md
```

### 4.2 Warning Alerts (Team Lead Notification)

```
🟡 HighPrivacyOverhead
   - Trigger: Overhead > 20%
   - Duration: 5 minutes
   - Investigation: Check GPU detection, CPU % usage

🟡 GPUDetectionFailure
   - Trigger: >10% failure rate
   - Duration: 5 minutes
   - Investigation: Check GPU drivers, CUDA version

🟡 HighCPUUsage
   - Trigger: CPU > 90%
   - Duration: 5 minutes
   - Investigation: Profile bottleneck, consider scaling

🟡 HighNetworkLatency
   - Trigger: Latency p99 > 100ms
   - Duration: 5 minutes
   - Investigation: Check ISP MTR, BGP routing

🟡 HighGPUTemperature
   - Trigger: Temperature > 85°C
   - Duration: 3 minutes
   - Investigation: Check cooling, thermal throttle
```

### 4.3 Best Practice Deployments

**Recommended Alert Configuration**:

```yaml
# In alertmanager.yml

global:
  resolve_timeout: 5m
  slack_api_url: '${SLACK_WEBHOOK_URL}'

route:
  receiver: 'default'
  group_by: ['alertname', 'cluster']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h

  routes:
    - match:
        severity: critical
      receiver: 'pagerduty'
      repeat_interval: 5m
      
    - match:
        severity: warning
      receiver: 'slack'
      repeat_interval: 1h

receivers:
  - name: 'default'
    slack_configs:
      - channel: '#monitoring'
        title: '{{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'

  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: '${PAGERDUTY_SERVICE_KEY}'
        description: '{{ .Alerts.Firing | len }} critical alerts'

  - name: 'slack'
    slack_configs:
      - channel: '#warnings'
```

---

## 5. Troubleshooting Guide

### Issue: Metrics Not Appearing

**Symptoms**: `/metrics` endpoint returns empty, Prometheus scrape shows 0 series

**Diagnosis**:

```bash
# 1. Check Sovereign Map started
curl http://localhost:9090/health
# Expected: {"status": "healthy"}

# 2. Verify metrics endpoint
curl http://localhost:9090/metrics
# Expected: output with "# HELP" lines and metric values

# 3. Check metric format
curl http://localhost:9090/metrics | head -20
# Should see: gauge, counter, histogram metrics

# 4. Verify Prometheus config
cat prometheus.yml | grep -A5 "job_name.*sovereign-map"
```

**Solutions**:

| Problem | Fix |
|---------|-----|
| Port 9090 unreachable | Check firewall rules, verify listening port |
| Empty /metrics output | Verify integration calls in code |
| Scrape target down | Check PrivacyEngine/Consensus/Network initialization |
| Wrong metrics format | Update metrics-registry.ts export format |

### Issue: High False-Positive Alerts

**Symptoms**: Byzantine alerts triggering on normal operations, constant Yellow warnings

**Diagnosis**:

```bash
# Check alert rule timing
# Each rule has:
# - expr: condition to trigger alert
# - for: duration before alerting (prevents flapping)

# Example: "for: 5m" means condition must be true for 5 minutes

# Review recent alerts
curl http://prometheus:9090/api/v1/rules | jq '.data.groups[].rules[] | {name, state}'
```

**Solutions**:

| Problem | Fix |
|---------|-----|
| Alerts flapping every 1-2 min | Increase "for:" duration from 5m to 15m |
| Byzantine detect on recoveries | Tune detection sensitivity, increase Byzantine threshold |
| Memory Yellow on GC pauses | Increase threshold from 90% to 95% |
| Network latency Yellow on bursts | Increase threshold from 100ms to 150ms |

**Recommended Tuning**:

```yaml
# In prometheus-alert-rules.yml

# Reduce sensitivity for bursty metrics
HighMemoryUsage:
  expr: system_memory_bytes > threshold
  for: 10m  # increased from 5m to reduce false positives
  
HighNetworkLatency:
  expr: system_network_latency_ms > 100
  for: 10m  # wait for sustained latency, not bursts

# Increase sensitivity for critical metrics
ByzantineNodesDetected:
  expr: byzantine_nodes_currently_detected >= 1
  for: 30s   # reduced from 1m, we need immediate alerting
```

### Issue: Prometheus Disk Space Growing

**Symptoms**: `/var/lib/prometheus` growing 1GB+ per day, disk full warnings

**Diagnosis**:

```bash
# Check storage usage
du -sh /var/lib/prometheus
# Check data retention
ls -lah /var/lib/prometheus/wal/

# Count time series
curl http://prometheus:9090/api/v1/series | jq '.data | length'
```

**Solutions**:

| Problem | Fix |
|---------|-----|
| Too many cardinality | Review metric labels, avoid unbounded values |
| Retention too long | Set `storage.tsdb.retention.time: 30d` in prometheus.yml |
| WAL growth | Increase `storage.tsdb.wal-segment-size: 256MB` |
| Old snapshots | `rm -rf /var/lib/prometheus/snapshots/*` |

**Recommended Settings**:

```yaml
# prometheus.yml

storage:
  tsdb:
    retention:
      time: 30d  # keep 30 days of metrics
      size: 50GB  # or size-based limit
    path: /var/lib/prometheus
    wal-segment-size: 256MB

# Reduce cardinality (high-cardinality labels are expensive)
# Example: Don't use nodeId as label if you have 10K nodes
# Instead: deviceId or instance (1 label per unique value)
```

### Issue: Grafana Dashboards Not Updating

**Symptoms**: Dashboard panels show "No data" or frozen timestamps

**Diagnosis**:

```bash
# Check Prometheus connectivity in Grafana
curl -u admin:admin http://localhost:3000/api/datasources
# Should show prometheus datasource as "ok"

# Test Prometheus directly
curl http://prometheus:9090/api/v1/query?query=up
# Should return current timestamp, not old data
```

**Solutions**:

| Problem | Fix |
|---------|-----|
| "No data" panels | Check query syntax in panel edit |
| Frozen timestamps | Restart Prometheus scraper |
| 404 on datasource | Re-add Prometheus datasource in Grafana |
| Slow dashboard loads | Increase query timeout, reduce time range |

---

## 6. Maintenance Procedures

### 6.1 Regular Maintenance Schedule

**Daily**:
- [ ] Review critical alerts before standup (5 min)
- [ ] Check SLA compliance dashboard (1 min)
- [ ] Verify no memory leak trend (memory should be stable)

**Weekly**:
- [ ] Review alert thresholds against actual metrics (15 min)
- [ ] Check Prometheus disk usage (5 min)
- [ ] Validate Grafana dashboard accuracy (10 min)
- [ ] Test alert escalation paths (Slack/PagerDuty delivery) (5 min)

**Monthly**:
- [ ] Review metric retention vs disk space (30 min)
- [ ] Tune alert rules based on false positive rate (1 hour)
- [ ] Update runbook documentation (30 min)
- [ ] Capacity planning: project growth rate (1 hour)

**Quarterly**:
- [ ] Performance baseline review (2 hours)
- [ ] SLA target validation against real-world data (2 hours)
- [ ] Disaster recovery drill (metrics from backup) (4 hours)

### 6.2 Backup & Recovery

**Backup Prometheus Data**:

```bash
# Backup Prometheus volumes every 24 hours
tar -czf prometheus-backup-$(date +%Y%m%d).tar.gz \
  /var/lib/prometheus/

# Upload to S3
aws s3 cp prometheus-backup-*.tar.gz s3://backups/prometheus/

# Keep 30-day rolling window
aws s3 ls s3://backups/prometheus/ | sort | tail -30
```

**Restore Prometheus**:

```bash
# Stop Prometheus
systemctl stop prometheus

# Restore from backup
tar -xzf prometheus-backup-YYYYMMDD.tar.gz -C /var/lib/prometheus/

# Start Prometheus
systemctl start prometheus

# Verify metrics are available
curl http://localhost:9090/api/v1/rules | jq '.data.groups | length'
```

### 6.3 Metric Migration (Adding Metrics)

When adding new metrics to Sovereign Map:

```bash
# 1. Update MetricsRegistry with new metric definition
#    (See src/metrics-registry.ts)

# 2. Update alert rules (if needed)
#    (See src/prometheus-alert-rules.ts)

# 3. Deploy code to staging
npm run deploy:staging

# 4. Verify new metrics appear
curl http://staging:9090/metrics | grep new_metric

# 5. Add to Grafana dashboards
#    (Edit JSON in grafana-dashboards.ts)

# 6. Deploy to production with metric backfill
npm run deploy:production

# 7. Monitor acceptance of new metrics (5-10 min)
# 8. Announce to team via Slack
```

---

## 7. FAQ & Common Issues

### Q: Why is GPU utilization low but privacy overhead still high?

**A**: GPU may be initialized but not fully utilized. Causes:
- Small batch size (< 1000 samples): CPU faster for startup overhead
- Low update frequency: GPU sits idle between updates
- Solution: Increase batch size to 10K+, use GPU for privacy budgets > 0.5

### Q: My Byzantine detection is triggering on every consensus round. False positive?

**A**: Likely causes:
- Nodes recovering from transient network issues
- NTP clock skew causing Byzantine detection
- Solution: Tune detection sensitivity, ensure NTP sync < 100ms

### Q: Network partition recovery is taking > 2 minutes. Is this normal?

**A**: Recovery time depends on:
- Cluster size (larger clusters = slower consensus)
- Byzantine nodes present (need majority agreement)
- Network latency (consensus round time × number of rounds)
- Normal: 20-50s for 10 nodes, 1-2 min for 100+ nodes

### Q: How do I know if privacy epsilon is sufficient?

**A**: Rule of thumb:
- ε = 1.0: ~100 updates per round before retraining
- ε = 0.5: ~50 updates per round
- ε = 0.1: ~10 updates per round
- Monitor "Epsilon Remaining" panel; if declining > 0.01/update, increase batch size

### Q: What's the difference between /health and /sla endpoints?

**A**:
- `/health`: Simple liveness probe (is process running?), used by Kubernetes
- `/sla`: Complex compliance check (all 5 SLAs passing?), used for deployment decisions

### Q: Can I modify alert thresholds after deployment?

**A**: Yes, alert rules are hot-reloadable:

```bash
# Edit prometheus-alert-rules.yml
vi prometheus-alert-rules.yml

# Reload rules (no restart needed)
systemctl reload prometheus

# Verify rules loaded
curl http://localhost:9090/api/v1/rules | jq '.data.groups[0].rules[] | .name'
```

---

## 8. Performance & Scaling

### Monitor Performance Impact

```bash
# Measure monitoring overhead (should be <1% CPU, <50MB memory)
ps aux | grep prometheus
ps aux | grep grafana

# Metric query latency
time curl http://localhost:9090/api/v1/query?query=up
# Expect: < 100ms
```

### Scaling Monitoring for Large Clusters

| Scale | Prometheus | Grafana | Disk/Day |
|-------|-----------|---------|----------|
| 10 nodes | 1 instance, 2GB RAM | 1 instance | 1GB |
| 100 nodes | 1 instance, 4GB RAM | 1 instance | 5GB |
| 1000 nodes | 2-3 instances, 8GB RAM | 1 HA pair | 20GB |
| 10K nodes | Thanos (federated) | 3-way HA | 100GB |

**For 1000+ nodes**:

```yaml
# Use Thanos for long-term storage + federation
# Install thanos-sidecar with Prometheus
# Set up Thanos querier for cross-cluster queries
```

---

## 9. References

- [Prometheus Docs](https://prometheus.io/docs/)
- [Grafana Alerting](https://grafana.com/docs/grafana/latest/alerting/)
- [SLO/SLI Best Practices](https://sre.google/sre-book/service-level-objectives/)
- Sovereign Map Docs: `docs/MONITORING.md`
- Alert Runbooks: `docs/INCIDENT_RESPONSE/`

---

**Document End**

**Next Steps**:
1. Import Grafana dashboards from `grafana-dashboards.ts`
2. Configure alert rules in `prometheus-alert-rules.yml`
3. Set up Alertmanager escalation paths
4. Run monitoring validation tests
5. Schedule ops team training

