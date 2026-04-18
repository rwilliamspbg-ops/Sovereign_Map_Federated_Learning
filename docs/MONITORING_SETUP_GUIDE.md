# Sovereign Map Monitoring Setup Guide

**Document Version**: 1.0  
**Last Updated**: 2026-03-18  
**Status**: Production Ready  

This guide covers the complete setup of the Sovereign Map monitoring stack (Prometheus, Grafana, Alertmanager).

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Local Development Setup](#local-development-setup)
3. [Kubernetes Production Deployment](#kubernetes-production-deployment)
4. [Dashboard Provisioning](#dashboard-provisioning)
5. [Configuration & Customization](#configuration--customization)
6. [Troubleshooting](#troubleshooting)

---

## Quick Start

### Minimal Setup (5 minutes)

```bash
# 1. Enable monitoring in your Sovereign Map application
npm install @sovereign-map/monitoring

# 2. In your main.ts/main.js
import { MonitoringOrchestrator } from '@sovereign-map/monitoring';

const monitoring = new MonitoringOrchestrator(9090);
await monitoring.initialize();

// Integrate with your modules
monitoring.integratePrivacy(privacyEngine);
monitoring.integrateConsensus(consensusModule);
monitoring.integrateNetwork(partitionRecoveryManager);

# 3. Verify metrics available
curl http://localhost:9090/metrics
# Should return Prometheus-format metrics

# 4. Start Prometheus
docker run -d --name prometheus -p 9090:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus:v2.40.0

# 5. Start Grafana
docker run -d --name grafana -p 3000:3000 grafana/grafana:9.3.0

# 6. Open Grafana
open http://localhost:3000/
# Default: admin / admin
```

---

## Local Development Setup

### Prerequisites

```bash
# Node.js v16+
node --version

# Docker & Docker Compose (for containers)
docker --version
docker-compose --version

# Git
git --version
```

### Step 1: Enable Monitoring in Application

```typescript
// src/index.ts
import { MonitoringOrchestrator } from '@sovereign-map/monitoring';
import { PrivacyEngine } from '@sovereign-map/privacy';
import { ConsensusModule } from '@sovereign-map/consensus';

// Initialize on app startup
const monitoring = new MonitoringOrchestrator(9090);
await monitoring.initialize();

// After creating modules
const privacyEngine = new PrivacyEngine();
const consensusModule = new ConsensusModule();

monitoring.integratePrivacy(privacyEngine);
monitoring.integrateConsensus(consensusModule);

// Metrics now flow to http://localhost:9090/metrics
```

### Step 2: Start Monitoring Stack with Docker Compose

```bash
# Create docker-compose file for monitoring
cat > docker-compose.monitoring.yml << 'EOF'
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:v2.40.0
    container_name: prometheus
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=30d'
    ports:
      - "9090:9090"
    networks:
      - monitoring

  grafana:
    image: grafana/grafana:9.3.0
    container_name: grafana
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin
      GF_INSTALL_PLUGINS: grafana-piechart-panel
    volumes:
      - grafana_data:/var/lib/grafana
    ports:
      - "3000:3000"
    depends_on:
      - prometheus
    networks:
      - monitoring

  alertmanager:
    image: prom/alertmanager:v0.24.0
    container_name: alertmanager
    volumes:
      - ./alertmanager.yml:/etc/alertmanager/alertmanager.yml
      - alertmanager_data:/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
    ports:
      - "9093:9093"
    networks:
      - monitoring

networks:
  monitoring:
    driver: bridge

volumes:
  prometheus_data:
  grafana_data:
  alertmanager_data:
EOF

# Start the stack
docker-compose -f docker-compose.monitoring.yml up -d

# Verify services running
docker-compose -f docker-compose.monitoring.yml ps

# View logs
docker-compose -f docker-compose.monitoring.yml logs -f prometheus
```

### Step 3: Configure Prometheus

Create `prometheus.yml`:

```yaml
global:
  scrape_interval: 30s
  evaluation_interval: 30s

scrape_configs:
  - job_name: 'sovereign-map'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 30s

rule_files:
  - '/etc/prometheus/rules/alert-rules.yml'

alerting:
  alertmanagers:
    - static_configs:
        - targets: ['localhost:9093']
```

### Step 4: Provision Grafana Dashboards

```bash
# Generate dashboard files
npm run provision:dashboards:generate

# This creates:
# - grafana-dashboards-provisioning.yaml
# - grafana-dashboards-configmap.yaml
# - dashboard-*.json (6 dashboard files)

# Deploy to Grafana
npm run provision:dashboards:deploy \
  --url=http://localhost:3000 \
  --admin-password=admin

# Verify dashboards created
open http://localhost:3000/dashboards
```

### Step 5: Verify Full Stack

```bash
# 1. Check Prometheus targets
curl http://localhost:9090/api/v1/targets
# Should show 'sovereign-map' job as UP

# 2. Check metrics available
curl http://localhost:9090/api/v1/query?query=up
# Should return current value

# 3. Open Grafana
open http://localhost:3000/

# 4. Verify dashboards imported
# Home > Dashboards > Sovereign Map folder
# Should see: GPU, Privacy, Consensus, Network, System, SLA dashboards

# 5. View sample metrics
curl http://localhost:9090/metrics | grep -E "^gpu_|^privacy_|^consensus_"
# Should show metric lines
```

### Step 6: Test Alerts (Optional)

```bash
# Trigger a test alert
# 1. Go to Prometheus
open http://localhost:9090/

# 2. Execute query that triggers alert
# privacy_overhead_percent > 20
# (This should fire if overhead actually exceeds 20%)

# 3. Check Alertmanager
open http://localhost:9093

# 4. Verify Slack notification (if configured)
```

---

## Kubernetes Production Deployment

### Prerequisites

```bash
# Kubernetes cluster access
kubectl version
# Should show server version 1.20+

# Helm (optional, for advanced deployments)
helm version

# kubectl access to target cluster
kubectl cluster-info
```

### Step 1: Deploy Monitoring Stack

```bash
# Create monitoring namespace and deploy
kubectl apply -f deploy/kubernetes-monitoring-stack.yaml

# Verify all components running
kubectl get pods -n sovereign-map-monitoring
# Should show: prometheus, grafana, alertmanager pods as Running

# Wait for PVCs to bind
kubectl get pvc -n sovereign-map-monitoring
# Should all be Bound

# Check services created
kubectl get svc -n sovereign-map-monitoring
# Should show prometheus, grafana, alertmanager services
```

### Step 2: Verify Services Are Healthy

```bash
# Check Prometheus logs
kubectl logs -n sovereign-map-monitoring -l app=prometheus

# Check Grafana logs
kubectl logs -n sovereign-map-monitoring -l app=grafana

# Port-forward to verify locally
kubectl port-forward -n sovereign-map-monitoring svc/prometheus 9090:9090 &
curl http://localhost:9090/api/v1/series
# Should return series list

kubectl port-forward -n sovereign-map-monitoring svc/grafana 3000:3000 &
curl -u "$GRAFANA_USER:$GRAFANA_PASSWORD" http://localhost:3000/api/health
# Should return {"status":"ok"}
```

### Step 3: Deploy Dashboards via Kubernetes

```bash
# Generate Kubernetes ConfigMap with dashboards
npm run provision:dashboards:k8s --namespace=sovereign-map-monitoring

# This applies grafana-dashboards-configmap.yaml with all dashboard JSONs
# Grafana will auto-mount and reload within 10 seconds

# Verify ConfigMap created
kubectl get configmap -n sovereign-map-monitoring
# Should show grafana-dashboards

# Check Grafana discovered dashboards
kubectl port-forward -n sovereign-map-monitoring svc/grafana 3000:3000
open http://localhost:3000/dashboards
# Should show all 6 dashboards in "Sovereign Map" folder
```

### Step 4: Configure External Access

```bash
# Option A: LoadBalancer (AWS/GCP/Azure)
kubectl get svc -n sovereign-map-monitoring grafana
# Get EXTERNAL-IP, access via http://EXTERNAL-IP:3000

# Option B: NodePort
kubectl get svc -n sovereign-map-monitoring
# Access via http://NODE-IP:NODE-PORT

# Option C: Ingress (requires ingress controller)
kubectl apply -f deploy/kubernetes-monitoring-stack.yaml
# Ingress will be created if ingress controller available
# Access via monitoring.example.com/grafana

# Option D: Port forwarding (development)
kubectl port-forward -n sovereign-map-monitoring svc/grafana 3000:3000
open http://localhost:3000
```

### Step 5: Configure Alert Routing

```bash
# Update Alertmanager Slack webhook
kubectl set env -n sovereign-map-monitoring deployment/alertmanager \
  SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

# Restart Alertmanager to pick up new config
kubectl rollout restart deployment/alertmanager -n sovereign-map-monitoring

# Verify Alertmanager connected
kubectl logs -n sovereign-map-monitoring -l app=alertmanager | grep -i slack
```

### Step 6: Update Grafana Admin Password

```bash
# Change default password
kubectl get secret -n sovereign-map-monitoring grafana-admin -o jsonpath='{.data.password}' | base64 -d
# Shows current password

# Update secret
kubectl patch secret grafana-admin -n sovereign-map-monitoring \
  -p '{"data":{"password":"'$(echo -n "your-new-password" | base64)'"}}'

# Or via Helm values (if using Helm)
helm upgrade sovereign-map-monitoring . \
  --set grafana.admin.password="your-new-password"
```

### Step 7: Enable Persistent Data

```bash
# Check storage classes available
kubectl get storageclass

# Use different storage class if needed
kubectl patch pvc prometheus-storage -n sovereign-map-monitoring \
  -p '{"spec":{"storageClassName":"fast-ssd"}}'

# For cloud storage (EBS/GCP persistent disk)
# Update kubernetes-monitoring-stack.yaml storage class
```

---

## Dashboard Provisioning

### Via Script (Recommended)

```bash
# Generate all dashboard files
npm run provision:dashboards:generate

# Output:
# - grafana-dashboards-provisioning.yaml (Grafana provisioning config)
# - grafana-dashboards-configmap.yaml (For Kubernetes)
# - dashboard-*.json (6 individual dashboard JSONs)
```

### Via HTTP API

```bash
# Deploy to running Grafana instance
npm run provision:dashboards:deploy \
  --url=http://grafana.example.com \
  --admin-password=your-password

# This:
# 1. Authenticates with Grafana
# 2. Posts each dashboard JSON to /api/dashboards/db
# 3. Creates dashboards with proper UIDs and settings
```

### Via Kubernetes

```bash
# Deploy as ConfigMap for auto-loading
npm run provision:dashboards:k8s \
  --namespace=sovereign-map-monitoring

# Grafana mounts ConfigMap at:
# /etc/grafana/provisioning/dashboards
# Auto-reloads within 10 seconds
```

### Manual Import

```bash
# 1. Open Grafana
open http://localhost:3000/

# 2. Left menu > Dashboards > Import

# 3. Choose import method:
#    - Upload JSON file (dashboard-*.json)
#    - Paste JSON text
#    - Enter dashboard UID

# 4. Configure:
#    - Select Prometheus data source
#    - Choose folder: "Sovereign Map"
#    - Click Import
```

### Verify Dashboards

```bash
# Via API
curl -u "$GRAFANA_USER:$GRAFANA_PASSWORD" http://localhost:3000/api/search

# Via UI
open http://localhost:3000/dashboards

# Check all 6 dashboards present:
# ✅ GPU Acceleration Monitor
# ✅ Privacy Budget Monitor
# ✅ Consensus & Byzantine Detection
# ✅ Network Partition & Recovery
# ✅ System Health
# ✅ SLA Compliance Monitor
```

---

## Configuration & Customization

### Grafana Configuration

```yaml
# grafana.env
GF_SECURITY_ADMIN_PASSWORD=your-password
GF_SECURITY_ADMIN_USER=admin
GF_INSTALL_PLUGINS=grafana-piechart-panel,grafana-worldmap-panel
GF_ALERTING_ENABLED=true
GF_SERVER_ROOT_URL=http://grafana.example.com
GF_EXTERNAL_IMAGE_STORAGE_TYPE=s3  # For cloud storage
```

### Prometheus Configuration

```yaml
# prometheus.yml advanced settings
global:
  scrape_interval: 30s
  scrape_timeout: 10s
  evaluation_interval: 30s
  external_labels:
    cluster: 'sovereign-map'
    environment: 'production'

# Remote storage (optional, for long-term retention)
remote_write:
  - url: http://storage-adapter:9201/write
    write_relabel_configs:
      - source_labels: [__name__]
        regex: 'privacy_.*'
        action: keep  # Only send privacy metrics remotely

# Remote read (optional, for querying long-term storage)
remote_read:
  - url: http://storage-adapter:9201/read
    read_recent: true
```

### Alert Customization

Edit alert rules in:
- `packages/monitoring/src/prometheus-alert-rules.ts` (source)
- Kubernetes ConfigMap: `prometheus-alert-rules` (config)

```yaml
# Add custom alert
- alert: CustomPrivacyThreshold
  expr: privacy_overhead_percent > 15  # Lower threshold
  for: 3m  # Faster trigger
  labels:
    severity: warning
    custom: true
  annotations:
    summary: "Custom privacy threshold exceeded"
```

### Multi-Cluster Setup

```yaml
# prometheus.yml for federation
# Scrape from remote Prometheus instances
scrape_configs:
  - job_name: 'federate-cluster1'
    static_configs:
      - targets: ['cluster1-prometheus:9090']
    metrics_path: '/federate'
    honor_labels: true

  - job_name: 'federate-cluster2'
    static_configs:
      - targets: ['cluster2-prometheus:9090']
    metrics_path: '/federate'
    honor_labels: true

# Then aggregate queries across clusters
# Query all clusters: {job=~"federate-.*"}
```

---

## Troubleshooting

### Strict chaos soak fails with 401 on `/trigger_fl`

Symptoms:

- `tests/scripts/python/test_soak_chaos_guard.py` fails during validation with no FL round progression.
- Backend logs show unauthorized requests for `/trigger_fl`.

Cause:

- Backend admin auth is enabled and manual fallback trigger calls require `JOIN_API_ADMIN_TOKEN`.

Resolution:

```bash
export JOIN_API_ADMIN_TOKEN='replace-with-strong-token'
export ALLOW_INSECURE_DEV_ADMIN_TOKEN='false'

docker compose -f docker-compose.full.yml up -d backend prometheus node-agent

JOIN_API_ADMIN_TOKEN="$JOIN_API_ADMIN_TOKEN" \
SOAK_CHAOS_ENABLED=1 \
SOAK_CHAOS_STRICT=1 \
CHAOS_MIN_CLIENT_QUORUM=1 \
python3 tests/scripts/python/test_soak_chaos_guard.py
```

Expected outcome:

- Suite completes with exit code `0`.
- Output includes `PASSED: FL rounds progressed under controlled churn`.

### Node-agent image build exceeds disk in local/CI runs

Current default:

- `Dockerfile.node-agent` uses CPU-only wheels (`torch==2.1.0+cpu`, `torchvision==0.16.0+cpu`) to avoid CUDA package bloat during standard validation.

If disk pressure persists:

```bash
docker compose -f docker-compose.full.yml down --remove-orphans
docker system prune -af --volumes
docker builder prune -af
df -h /
```

### Dashboards Not Appearing

**Problem**: Dashboards imported but not showing in Grafana

**Solution**:

```bash
# 1. Check datasource configured
curl -u "$GRAFANA_USER:$GRAFANA_PASSWORD" http://localhost:3000/api/datasources
# Should show Prometheus with id

# 2. Verify metrics available in Prometheus
curl http://localhost:9090/api/v1/series?match[]=up
# Should return series

# 3. Reload dashboard
# In Grafana: Dashboard settings > Reload dashboard

# 4. Check browser console for errors
# F12 > Console > Look for red errors
```

### Prometheus Not Scraping Metrics

**Problem**: Prometheus shows "DOWN" for sovereign-map target

**Solution**:

```bash
# 1. Verify Sovereign Map running
curl http://localhost:9090/metrics
# Should return non-empty response

# 2. Check Prometheus config syntax
curl http://localhost:9090/-/reload
# Should return 200

# 3. Check target URL in config
cat prometheus.yml | grep -A2 "sovereign-map"
# Verify targets match your setup

# 4. View Prometheus targets
open http://localhost:9090/targets
# Check the "Label" column for error details
```

### High Disk Usage

**Problem**: Prometheus/Grafana consuming too much disk

**Solution**:

```bash
# For Prometheus
# In prometheus.yml, reduce retention:
global:
  storage:
    tsdb:
      retention:
        time: 7d  # Reduce from 30d
        size: 10Gi  # Set size limit

# For Grafana
# In grafana.ini:
[database]
wal = true  # Use write-ahead log to reduce disk I/O

# Restart to apply
docker-compose -f docker-compose.monitoring.yml restart prometheus grafana
```

### Alerts Not Triggering

**Problem**: Expected alerts don't fire

**Solution**:

```bash
# 1. Verify rule loaded
curl http://localhost:9090/api/v1/rules
# Should see all rules in output

# 2. Manually test alert condition
# In Prometheus: Execute the alert rule query
# e.g., privacy_overhead_percent > 20
# Should return series if condition met

# 3. Check alert evaluation window
# Alerts require "for" duration before firing
# Example: for: 5m means condition must be true for 5 minutes

# 4. Verify Alertmanager connected
curl http://localhost:9093/api/v1/alerts
# Should show active alerts (or empty if none)
```

### Grafana Login Issues

**Problem**: Cannot login to Grafana

**Solution**:

```bash
# 1. Reset admin password
docker exec grafana /bin/bash -c \
  'grafana-cli admin reset-admin-password newpassword'

# 2. Check credential in Kubernetes secret
kubectl get secret grafana-admin -n sovereign-map-monitoring -o yaml
# password field should be base64 encoded

# 3. Reset via environment variable
docker run --rm \
  -e GF_SECURITY_ADMIN_PASSWORD=newpass \
  grafana/grafana:latest
```

---

## Maintenance

### Daily Tasks

```bash
# Check critical alerts
open http://localhost:3000/dashboards/sla-compliance-monitor

# Verify all datasources connected
curl http://localhost:3000/api/datasources

# Check disk usage
df -h /var/lib/prometheus
df -h /var/lib/grafana
```

### Weekly Tasks

```bash
# Backup Prometheus data
tar -czf prometheus-backup-$(date +%Y%m%d).tar.gz /var/lib/prometheus

# Review alert tuning
open http://localhost:9090/alerts

# Check Grafana dashboards for panel errors
# Look for "No data" panels
```

### Monthly Tasks

```bash
# Update Prometheus/Grafana/Alertmanager images
docker pull prom/prometheus:latest
docker-compose -f docker-compose.monitoring.yml pull
docker-compose -f docker-compose.monitoring.yml up -d

# Review retention policy
# Scale PVCs if needed for K8s
# kubectl patch pvc prometheus-storage -n sovereign-map-monitoring \
#   -p '{"spec":{"resources":{"requests":{"storage":"100Gi"}}}}'
```

---

## Next Steps

1. **Integration Testing**: Verify metrics flow with your modules
2. **Alert Customization**: Tune thresholds for your environment
3. **Dashboard Customization**: Add custom panels for metrics you care about
4. **Automation**: Set up alert webhooks for your on-call system
5. **Scaling**: Set up federation for multi-cluster monitoring

See [PHASE_3A_MONITORING_OPERATIONAL_GUIDE.md](../docs/PHASE_3A_MONITORING_OPERATIONAL_GUIDE.md) for detailed ops procedures.
