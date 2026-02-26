# TPM Metrics & Monitoring Pipeline - Complete Guide

## Overview

This document describes the complete TPM trust and verification metrics pipeline, integrating certificate management, trust chain verification, and security metrics into Prometheus and Grafana.

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    TPM Nodes                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐           │
│  │  Node 0     │  │  Node 1     │  │  Node N     │           │
│  │ (Trust mgr) │  │ (Trust auth)│  │ (Trust auth)│           │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘           │
│         │                │                │                   │
│         └────────────────┼────────────────┘                   │
│                          │                                     │
│                   ┌──────▼───────┐                            │
│                   │ tpm-metrics   │ :9091/metrics             │
│                   │ Exporter      │                           │
│                   └──────┬────────┘                           │
│                          │                                     │
│    ┌─────────────────────┘                                   │
│    │                                                          │
│    ▼                                                          │
│ ┌────────────────┐     ┌──────────────┐  ┌─────────────┐    │
│ │ Prometheus     │────▶│ Grafana      │  │ Alertmanager│   │
│ │ :9090          │     │ :3000        │  │ :9093       │   │
│ │                │     │              │  │             │   │
│ │ Alert Rules    │     │ Dashboards   │  │ Email/Slack │   │
│ │ └─tpm_alerts   │     │ └─TPM Trust  │  │ Notifications   │
│ │                │     │ └─Mohawk     │  │             │   │
│ └────────────────┘     └──────────────┘  └─────────────┘    │
│                                                               │
└──────────────────────────────────────────────────────────────┘
```

## Components

### 1. TPM Metrics Exporter (`tpm_metrics_exporter.py`)

Exports TPM trust metrics to Prometheus in real-time.

**Metrics Exported:**

```
Certificate Metrics:
├── tpm_certificates_total              (Gauge) - Total issued certificates
├── tpm_certificates_verified_total     (Gauge) - Verified certificates
├── tpm_certificates_revoked_total      (Gauge) - Revoked certificates
├── tpm_certificate_expiry_seconds      (Gauge) - Seconds until expiry
└── tpm_certificate_age_seconds         (Gauge) - Certificate age

Trust Chain Metrics:
├── tpm_trust_chain_valid               (Gauge) - Trust chain valid (1/0)
├── tpm_trust_verification_duration_seconds (Histogram) - Verification time
├── tpm_trust_verification_failures_total   (Counter) - Verification failures
└── tpm_ca_certificate_valid            (Gauge) - CA cert valid

Message Authentication:
├── tpm_messages_signed_total           (Counter) - Signed messages
├── tpm_messages_verified_total         (Counter) - Verified messages
├── tpm_signature_verification_failures_total (Counter) - Failed signatures
└── tpm_message_verification_duration_seconds (Histogram) - Verification time

Node Trust Status:
├── tpm_node_trust_score                (Gauge) - Trust score (0-100)
├── tpm_node_certificate_valid          (Gauge) - Node cert valid
├── tpm_node_certificate_revoked        (Gauge) - Node cert revoked
├── tpm_crl_size                        (Gauge) - Revocation list size
├── tpm_trust_cache_hits_total          (Counter) - Cache hits
└── tpm_trust_cache_misses_total        (Counter) - Cache misses
```

### 2. Alert Rules (`tpm_alerts.yml`)

Prometheus alert rules for critical trust events.

**Alert Categories:**

- **Certificate Expiration**: Warns 30 days before, critical 7 days before
- **Trust Chain**: Alerts on invalid chains or CA issues
- **Signature Failures**: Detects compromised nodes
- **Revocation**: Tracks revoked certificates
- **Node Trust**: Monitors individual node trust scores
- **Performance**: Tracks verification latency

### 3. Grafana Dashboard (`tpm_trust_dashboard.json`)

Comprehensive dashboard with 18 panels monitoring all aspects of TPM trust.

**Dashboard Sections:**

1. **Summary Stats** (4 panels)
   - Total certificates
   - Verified certificates
   - Revoked certificates
   - CA status

2. **Certificate Health** (2 panels)
   - Certificate expiry timeline
   - Certificate age distribution

3. **Trust Chain Status** (2 panels)
   - Trust chain validity
   - Node trust scores (color-coded)

4. **Message Authentication** (2 panels)
   - Message signing rate
   - Message verification rate

5. **Verification Failures** (2 panels)
   - Signature verification failures
   - Failure rate by source node

6. **Performance Metrics** (2 panels)
   - Trust verification latency (P95)
   - Message verification latency (P95)

7. **Advanced Monitoring** (4 panels)
   - Revoked nodes status
   - Certificate validity heatmap
   - Verification failure count
   - Cache hit rate
   - Verification status summary table

### 4. Docker Compose (`docker-compose.monitoring.tpm.yml`)

Complete monitoring stack with TPM integration.

**Services:**

- **prometheus**: Scrapes metrics from exporter, processes alert rules
- **tpm-metrics**: Exports TPM metrics on port 9091
- **alertmanager**: Handles alert notifications
- **grafana**: Visualizes metrics with dashboards
- **loki**: Optional log aggregation

## Deployment

### Quick Start

```bash
# Create sovereign network
docker network create sovereign-network

# Create external tpm-certs volume (if not exists)
docker volume create tpm-certs

# Deploy complete monitoring stack
docker-compose -f docker-compose.monitoring.tpm.yml up -d

# Verify services
docker-compose -f docker-compose.monitoring.tpm.yml ps
```

### Access Points

```bash
# Prometheus
http://localhost:9090

# TPM Metrics Exporter
http://localhost:9091/metrics

# Grafana
http://localhost:3000 (admin/admin)

# Alertmanager
http://localhost:9093

# Loki
http://localhost:3100
```

## Usage Examples

### 1. View TPM Metrics

```bash
# Raw Prometheus metrics
curl http://localhost:9091/metrics

# Summary JSON
curl http://localhost:9091/metrics/summary

# Health check
curl http://localhost:9091/health
```

### 2. Query Metrics in Prometheus

```
# Certificate expiry alerts
tpm_certificate_expiry_seconds < 604800

# Low trust scores
tpm_node_trust_score < 50

# Message verification failures
rate(tpm_signature_verification_failures_total[5m])

# Trust cache hit rate
tpm_trust_cache_hits_total / (tpm_trust_cache_hits_total + tpm_trust_cache_misses_total)
```

### 3. Set Up Alerts

**Email Alerts:**

```yaml
# alertmanager.yml
global:
  resolve_timeout: 5m

route:
  receiver: 'email'
  group_by: ['alertname', 'cluster', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h

receivers:
  - name: 'email'
    email_configs:
      - to: 'security-team@example.com'
        from: 'alerts@example.com'
        smarthost: 'smtp.example.com:587'
        auth_username: 'alerts@example.com'
        auth_password: 'password'
```

**Slack Alerts:**

```yaml
receivers:
  - name: 'slack'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'
        channel: '#security-alerts'
        title: 'TPM Alert: {{ .GroupLabels.alertname }}'
        text: '{{ .CommonAnnotations.description }}'
```

### 4. Create Custom Dashboards

Add panels to Grafana:

```python
# Python script to programmatically add panels
import requests

grafana_url = "http://localhost:3000"
api_key = "your-api-key"

dashboard = {
    "dashboard": {
        "title": "Custom TPM Metrics",
        "panels": [
            {
                "id": 1,
                "title": "Certificate Validity",
                "type": "stat",
                "targets": [
                    {
                        "expr": "tpm_certificates_total",
                        "refId": "A"
                    }
                ]
            }
        ]
    }
}

response = requests.post(
    f"{grafana_url}/api/dashboards/db",
    headers={"Authorization": f"Bearer {api_key}"},
    json=dashboard
)
```

## Monitoring Guide

### Daily Checks

1. **Certificate Expiry**
   - Check `/trust/status` endpoint
   - Monitor `tpm_certificate_expiry_seconds` in Grafana
   - Set up alerts for <30 days

2. **Trust Chain Validity**
   - Verify `tpm_trust_chain_valid` = 1 for all nodes
   - Check `tpm_ca_certificate_valid` = 1

3. **Message Authentication**
   - Monitor `tpm_signature_verification_failures_total` for spikes
   - Check message verification latency (P95 < 500ms)

4. **Node Trust Scores**
   - All nodes should have trust score > 75
   - Investigate nodes with score < 50

### Weekly Checks

1. **Certificate Age**
   - Review oldest certificates (approaching 1-year mark)
   - Plan rotation schedule

2. **Revocation Status**
   - Check CRL size (should be low)
   - Review audit logs for revocations

3. **Performance Trends**
   - Analyze verification time trends
   - Check cache hit rate (should be > 90%)

### Alert Response

**Certificate Expiring in 30 Days**
- Action: Schedule certificate rotation
- Response time: 1 week

**Certificate Expiring in 7 Days**
- Action: Begin emergency rotation procedures
- Response time: 24 hours

**High Signature Verification Failures**
- Action: Investigate compromised node
- Response time: Immediate
- Steps:
  1. Identify source node
  2. Review recent network activity
  3. Consider certificate revocation
  4. Replace node

**Invalid Trust Chain**
- Action: Quarantine affected nodes
- Response time: Immediate
- Steps:
  1. Isolate from network
  2. Audit certificates
  3. Investigate root cause
  4. Re-provision if necessary

## Metrics Export Formats

### Prometheus Text Format (Default)

```
# HELP tpm_certificates_total Total number of issued certificates
# TYPE tpm_certificates_total gauge
tpm_certificates_total 10

# HELP tpm_node_trust_score Trust score for a node (0-100)
# TYPE tpm_node_trust_score gauge
tpm_node_trust_score{node_id="0"} 95
tpm_node_trust_score{node_id="1"} 85
```

### JSON Summary Format

```json
{
  "node_id": 0,
  "timestamp": "2024-02-26T12:30:45.123Z",
  "certificates": {
    "total": 10,
    "verified": 9,
    "revoked": 1
  },
  "ca": {
    "valid": true,
    "certificate_path": "/etc/sovereign/certs/ca-cert.pem"
  }
}
```

## Troubleshooting

### Problem: Metrics Not Appearing in Prometheus

```bash
# 1. Check exporter is running
curl http://localhost:9091/health

# 2. Check Prometheus config
curl http://localhost:9090/api/v1/query?query=up

# 3. Check scrape endpoint
docker logs sovereign-prometheus | grep "tpm-metrics"
```

### Problem: High Certificate Failure Rate

```bash
# 1. Check certificate validity
docker exec sovereign-tpm-ca python -c "
from tpm_cert_manager import TPMCertificateManager
mgr = TPMCertificateManager('/etc/sovereign/certs')
print(mgr.get_trust_report())
"

# 2. Review revocation list
cat /etc/sovereign/certs/crl.json

# 3. Check logs
docker logs sovereign-backend-secure
```

### Problem: High Verification Latency

```bash
# 1. Check cache hit rate
curl http://localhost:9091/metrics | grep cache

# 2. Review certificate count
curl http://localhost:9091/metrics/summary

# 3. Check CPU/Memory usage
docker stats sovereign-tpm-metrics
```

## Performance Tuning

### Cache Configuration

```python
# Adjust trust cache TTL in secure_communication.py
self.cache_ttl = 3600  # 1 hour (increase for stability)
```

### Batch Verification

```python
# Verify multiple certificates in parallel
from concurrent.futures import ThreadPoolExecutor

with ThreadPoolExecutor(max_workers=4) as executor:
    futures = [
        executor.submit(mgr.verify_node_certificate, i)
        for i in range(num_nodes)
    ]
    results = [f.result() for f in futures]
```

### Metrics Retention

```bash
# Adjust Prometheus retention (default 30d)
# In docker-compose.monitoring.tpm.yml
command:
  - '--storage.tsdb.retention.time=60d'  # Keep 60 days
```

## Integration with Existing Monitoring

### Add TPM Dashboard to Grafana Home

1. Create folder: "Security"
2. Add dashboard: "TPM Trust & Verification Monitoring"
3. Pin to home for quick access

### Link to Other Dashboards

- Link TPM Dashboard → Mohawk Observability (for context)
- Link TPM Dashboard → Byzantine Fault Tolerance (for threat analysis)

### Create Unified Alerts

```yaml
# Combine TPM + BFT alerts
- alert: CompromisedNodeDetected
  expr: |
    (tpm_node_trust_score == 0)
    or
    (bft_byzantine_nodes_detected > 0)
  for: 1m
  labels:
    severity: critical
    component: security
```

## Next Steps

1. **Hardware TPM Integration**
   - Use real TPM for key storage
   - Implement measured boot

2. **Key Rotation Automation**
   - Automatic certificate renewal 30 days before expiry
   - Zero-downtime rotation

3. **Audit Trail**
   - Log all trust operations to Loki
   - Generate compliance reports

4. **Advanced Analytics**
   - Detect anomalies in verification patterns
   - Machine learning-based threat detection

5. **Multi-Cluster Support**
   - Federated Prometheus for multiple deployments
   - Centralized alert management

## References

- Prometheus: https://prometheus.io/docs/
- Grafana: https://grafana.com/docs/grafana/latest/
- Alert Rules: https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/
- TPM Guide: TPM_TRUST_GUIDE.md

---

**Version**: 1.0.0  
**Last Updated**: 2024  
**Status**: Production Ready
