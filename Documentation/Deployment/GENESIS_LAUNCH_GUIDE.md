# 🚀 Genesis Block Launch Guide

**Sovereign Map Federated Learning v1.0.0**  
**Launch Date**: February 28, 2026  
**Status**: Production Ready

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Pre-Launch Requirements](#pre-launch-requirements)
3. [Launch Procedure](#launch-procedure)
4. [Monitoring & Dashboards](#monitoring--dashboards)
5. [Post-Launch Operations](#post-launch-operations)
6. [Troubleshooting](#troubleshooting)
7. [Emergency Procedures](#emergency-procedures)

---

## 🎯 Overview

The **Genesis Block Launch** marks the official production deployment of Sovereign Map's federated learning network. This guide provides comprehensive instructions for launching and monitoring the network.

### Launch Objectives

- ✅ Deploy minimum 20 nodes for Byzantine fault tolerance
- ✅ Establish trusted network with TPM attestation
- ✅ Initialize federated learning with convergence monitoring
- ✅ Enable real-time monitoring and alerting
- ✅ Achieve 93%+ model accuracy within first 500 rounds

### Architecture Components

```
┌─────────────────────────────────────────────────┐
│            🎯 Genesis Network                   │
├─────────────────────────────────────────────────┤
│                                                 │
│  👥 Node Layer (20-100 nodes)                  │
│  ├─ Federated Learning Clients                 │
│  ├─ TPM Security & Trust                       │
│  └─ P2P Communication                          │
│                                                 │
│  🔄 Aggregation Layer                          │
│  ├─ Backend API (Port 8000)                    │
│  ├─ Flower gRPC Server (Port 8080)             │
│  └─ Consensus Mechanism                        │
│                                                 │
│  📊 Monitoring Stack                           │
│  ├─ Prometheus (Port 9090)                     │
│  ├─ Grafana (Port 3001)                        │
│  └─ Alertmanager (Port 9093)                   │
│                                                 │
└─────────────────────────────────────────────────┘
```

---

## 🔧 Pre-Launch Requirements

### System Requirements

**Minimum Specifications:**
- **CPU**: 8 cores
- **RAM**: 16GB
- **Storage**: 100GB SSD
- **Network**: 1 Gbps stable connection
- **OS**: Linux (Ubuntu 22.04+ recommended)

**Recommended for Production:**
- **CPU**: 16+ cores
- **RAM**: 32GB+
- **Storage**: 500GB NVMe SSD
- **Network**: 10 Gbps dedicated
- **OS**: Ubuntu 24.04 LTS

### Software Dependencies

```bash
# Required
- Docker 24.0+
- Docker Compose 2.20+
- Git 2.40+

# Optional (for development)
- Python 3.11+
- Go 1.21+
- Node.js 20+
```

### Installation

```bash
# Clone repository
git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning.git
cd Sovereign_Map_Federated_Learning

# Verify installation
./genesis-launch.sh
```

---

## 🚀 Launch Procedure

### Step 1: Pre-Launch Validation

Run pre-flight checks to ensure system readiness:

```bash
# Automated validation
./genesis-launch.sh

# Manual checks
docker --version
docker compose version
docker network ls
docker system df
```

**Expected Output:**
```
✓ Docker installed: Docker version 24.0.0
✓ Docker Compose installed: Docker Compose version 2.20.0
✓ All required files present
✓ System Resources: 16 cores, 32GB RAM, 400GB available
✓ All ports available
```

### Step 2: Launch Monitoring Stack

The script automatically launches:
- Prometheus (metrics collection)
- Grafana (visualization)
- Alertmanager (alerting)

```bash
# Monitoring services start automatically
# Manual start if needed:
docker compose -f docker-compose.monitoring.yml up -d
```

**Verification:**
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3001 (admin/admin)
- Alertmanager: http://localhost:9093

### Step 3: Genesis Block Creation

The Genesis block initializes the network with:

```json
{
  "genesis_time": "2026-02-28 00:00:00",
  "chain_id": "sovereign-mainnet",
  "initial_nodes": 20,
  "consensus_mechanism": "BFT",
  "min_trust_score": 75,
  "target_accuracy": 0.85,
  "max_byzantine_tolerance": 0.33
}
```

### Step 4: Network Deployment

Nodes are deployed in stages:

```bash
# Phase 1: Initial 20 nodes (automatic)
# Phase 2: Scale to 50 nodes (if needed)
docker compose -f docker-compose.production.yml up -d --scale node-agent=50

# Phase 3: Scale to 100 nodes (production)
docker compose -f docker-compose.production.yml up -d --scale node-agent=100
```

### Step 5: Health Verification

Monitor system health:

```bash
# Quick status check
./genesis-launch.sh status

# Continuous monitoring
./genesis-launch.sh monitor
```

---

## 📊 Monitoring & Dashboards

### Genesis Launch Overview Dashboard

**URL**: http://localhost:3001/d/genesis-launch-overview

**Key Metrics:**
- 🚀 Genesis Block Round (current training round)
- 👥 Active Network Nodes (connected participants)
- 🎯 Model Accuracy (FL training progress)
- 🔒 Network Security Status (TPM verification)

**Panels:**
1. Genesis Training Progress (accuracy & convergence)
2. Network Activity (nodes & rounds per minute)
3. Training Round Performance (duration histogram)
4. Node Trust Scores (security ratings)
5. Network Node Status (pie chart)
6. TPM Verification Performance (P95/P99 latency)

### Network Performance & Health

**URL**: http://localhost:3001/d/network-performance-health

**Key Metrics:**
- 🟢 Online/Offline Nodes
- ⏱️ Average Network Latency
- 📡 Message Success Rate
- 🌐 Network Latency Distribution (P50/P95/P99)

**Panels:**
1. Network Latency Distribution
2. Peer Connection Rate
3. Network Throughput (bytes sent/received)
4. Message Type Distribution
5. Network Topology Heatmap
6. Node Network Stats (detailed table)

### Consensus & Trust Monitoring

**URL**: http://localhost:3001/d/consensus-trust-monitoring

**Key Metrics:**
- 📊 Federated Learning Metrics
- ⚡ Update Throughput by Node
- 🔒 Trust Scores Over Time
- 💾 Cache Hit Rate

**Panels:**
1. FL Metrics (accuracy, loss, convergence)
2. Update Throughput by Node
3. Trust Scores Over Time
4. Certificate Distribution
5. Signature Verification Rate
6. Certificate Expiration Timeline
7. Node Trust Report (detailed table)

### Dashboard Navigation

```
http://localhost:3001
├─ Genesis Launch Overview     (Main launch dashboard)
├─ Network Performance         (Network health & metrics)
├─ Consensus & Trust          (Security & trust monitoring)
└─ Custom Dashboards          (User-created panels)
```

### Custom Queries

Access Prometheus directly for advanced queries:

```promql
# Model accuracy trend
rate(sovereignmap_fl_accuracy[5m])

# Node participation
count(up{job="sovereign-nodes"} == 1)

# Trust score distribution
histogram_quantile(0.95, tpm_node_trust_score_bucket)

# Network throughput
rate(sovereignmap_network_bytes_total[1m])
```

---

## 🔧 Post-Launch Operations

### Scaling Operations

**Gradual Scaling (Recommended):**
```bash
# Scale to 30 nodes
docker compose -f docker-compose.production.yml up -d --scale node-agent=30

# Wait for stabilization (5 minutes)
sleep 300

# Scale to 50 nodes
docker compose -f docker-compose.production.yml up -d --scale node-agent=50
```

**Immediate Scaling:**
```bash
# Scale directly to target
docker compose -f docker-compose.production.yml up -d --scale node-agent=100
```

### Performance Tuning

**Adjust FL Parameters:**
```bash
# Edit backend configuration
vi sovereignmap_production_backend_v2.py

# Key parameters:
# - ROUND_DURATION: Training round length
# - MIN_CLIENTS: Minimum participating nodes
# - CONVERGENCE_THRESHOLD: Accuracy target
```

**Optimize Network:**
```bash
# Adjust Docker resources
docker update --cpus="4" --memory="8g" <container_id>

# Monitor resource usage
docker stats
```

### Backup & Recovery

**Automated Backups:**
```bash
# Backup metrics data
docker run --rm -v prometheus_data:/data -v $(pwd):/backup \
  alpine tar czf /backup/prometheus-backup-$(date +%Y%m%d).tar.gz /data

# Backup Grafana dashboards
docker run --rm -v grafana_data:/data -v $(pwd):/backup \
  alpine tar czf /backup/grafana-backup-$(date +%Y%m%d).tar.gz /data
```

**Restore from Backup:**
```bash
# Stop services
- docker compose -f docker-compose.production.yml down --remove-orphans

# Restore data
tar xzf prometheus-backup-YYYYMMDD.tar.gz -C /var/lib/docker/volumes/prometheus_data/_data

# Restart services
docker compose up -d
```

---

## 🔍 Troubleshooting

### Common Issues

#### Issue: Nodes Not Connecting

**Symptoms:**
- Low active node-agent count
- Network errors in logs
- Timeouts in Grafana

**Solutions:**
```bash
# Check network connectivity
docker network inspect sovereign-genesis

# Restart networking
docker compose -f docker-compose.production.yml -f docker-compose.monitoring.yml restart

# Check firewall rules
sudo ufw status
sudo ufw allow 8000,8080,9090,3000,9093/tcp
```

#### Issue: Low Model Accuracy

**Symptoms:**
- Accuracy < 70% after 100 rounds
- High loss values
- Convergence rate near zero

**Solutions:**
```bash
# Check data distribution
curl http://localhost:8000/convergence | jq '{current_round, current_accuracy, current_loss}'

# Increase training iterations
# Edit docker-compose.yml:
#   environment:
#     - EPOCHS_PER_ROUND=5  # Increase from 3

# Restart nodes
docker compose -f docker-compose.production.yml restart node-agent
```

#### Issue: TPM Verification Failures

**Symptoms:**
- Trust scores < 75
- Signature verification errors
- Red security status

**Solutions:**
```bash
# Regenerate certificates
./tpm-bootstrap.sh

# Check certificate validity
curl http://localhost:8000/health

# Restart TPM services
docker compose -f docker-compose.production.yml restart backend
```

#### Issue: High Memory Usage

**Symptoms:**
- OOM errors
- Container crashes
- Slow performance

**Solutions:**
```bash
# Reduce node-agent count
docker compose -f docker-compose.production.yml up -d --scale node-agent=20

# Increase swap space
sudo fallocate -l 8G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# Adjust Docker limits
# Edit /etc/docker/daemon.json:
# {
#   "default-ulimits": {
#     "memlock": { "soft": -1, "hard": -1 }
#   }
# }
```

### Debug Commands

```bash
# View all logs
docker compose logs -f

# View specific service logs
docker compose logs -f backend
docker compose logs -f node

# Check container health
docker ps -a
docker inspect <container_id>

# Network diagnostics
docker exec <container_id> netstat -tulpn
docker exec <container_id> ping backend

# Resource usage
docker stats --no-stream
```

---

## 🚨 Emergency Procedures

### Emergency Shutdown

```bash
# Graceful shutdown (recommended)
docker compose -f docker-compose.production.yml -f docker-compose.monitoring.yml down --remove-orphans

# Force shutdown (if unresponsive)
docker compose -f docker-compose.production.yml -f docker-compose.monitoring.yml kill
docker compose -f docker-compose.production.yml -f docker-compose.monitoring.yml rm -f
```

### Network Recovery

```bash
# 1. Stop all services
docker compose -f docker-compose.production.yml -f docker-compose.monitoring.yml down --remove-orphans

# 2. Clean Docker state
docker system prune -af --volumes

# 3. Restart from Genesis
./genesis-launch.sh
```

### Data Corruption Recovery

```bash
# 1. Stop services
docker compose -f docker-compose.production.yml -f docker-compose.monitoring.yml down --remove-orphans

# 2. Remove corrupted volumes
docker volume rm prometheus_data grafana_data

# 3. Restore from backup
tar xzf prometheus-backup-*.tar.gz -C /var/lib/docker/volumes/

# 4. Restart services
docker compose -f docker-compose.production.yml -f docker-compose.monitoring.yml up -d
```

### Contact & Support

**Emergency Contacts:**
- Technical Lead: [contact information]
- DevOps Team: [contact information]
- Security Team: [contact information]

**Documentation:**
- Architecture: [ARCHITECTURE.md](/Documentation/Architecture/ARCHITECTURE.md)
- Deployment: [DEPLOYMENT.md](/Documentation/Deployment/DEPLOYMENT.md)
- Testing: [tests/docs/TEST_GUIDE.md](tests/docs/TEST_GUIDE.md)

---

## 📈 Success Metrics

### Launch Success Criteria

✅ **Network Health:**
- Minimum 20 nodes online
- < 50ms average latency
- > 95% message success rate

✅ **Security:**
- All nodes have trust scores > 75
- Zero signature verification failures
- All certificates valid

✅ **Performance:**
- Model accuracy > 85% within 500 rounds
- Round duration < 10 seconds
- Convergence rate > 0.5%

✅ **Monitoring:**
- All dashboards accessible
- Alerts configured and firing correctly
- Metrics collection at 10s intervals

### Post-Launch Checklist

- [ ] Verify all 20+ nodes online
- [ ] Confirm Grafana dashboards loading
- [ ] Check Prometheus scraping all targets
- [ ] Verify TPM trust scores > 75
- [ ] Monitor first 10 training rounds
- [ ] Confirm accuracy trending upward
- [ ] Test alert notifications
- [ ] Backup initial state
- [ ] Document any issues encountered
- [ ] Celebrate successful launch! 🎉

---

## 🎉 Launch Day Timeline

**T-60 minutes**: Pre-launch validation  
**T-45 minutes**: Start monitoring stack  
**T-30 minutes**: Deploy backend services  
**T-15 minutes**: Deploy initial 20 nodes  
**T-10 minutes**: Verify all systems green  
**T-5 minutes**: Final checks  
**T-0**: 🚀 **GENESIS BLOCK LAUNCH**  
**T+5 minutes**: Monitor first rounds  
**T+30 minutes**: Verify convergence  
**T+60 minutes**: Scale to 50 nodes (optional)
**T+120 minutes**: System stabilized

---

## 📝 Conclusion

The Genesis Block Launch establishes the foundation of Sovereign Map's federated learning network. With comprehensive monitoring, automated health checks, and professional dashboards, the network is ready for production deployment.

**Remember:**
- Monitor continuously during first 24 hours
- Scale gradually based on demand
- Keep backups up to date
- Document all issues and resolutions

**Welcome to the Sovereign Map Genesis Era! 🚀**

---

*Last Updated: February 28, 2026*  
*Version: 1.0.0*  
*Status: Production Ready*
