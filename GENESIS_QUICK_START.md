# ⚡ Genesis Block Quick Start

**5-Minute Launch Guide**

---

## 🎯 Prerequisites

```bash
# Verify system ready
docker --version          # Must be 24.0+
docker compose version    # Must be 2.20+
ls genesis-launch.sh      # Must exist
```

---

## 🚀 Launch Steps

### 1. Run Launch Script

```bash
./genesis-launch.sh
```

**That's it!** The script automatically:
- ✅ Validates system requirements
- ✅ Launches monitoring stack
- ✅ Creates Genesis block
- ✅ Deploys 20 nodes
- ✅ Starts health monitoring

---

## 📊 Access Dashboards

| Service | URL | Credentials |
|---------|-----|-------------|
| **Grafana** | http://localhost:3000 | admin / admin |
| **Prometheus** | http://localhost:9090 | - |
| **Backend API** | http://localhost:8000 | - |
| **Alertmanager** | http://localhost:9093 | - |

---

## 📈 Key Dashboards

### 🎯 Genesis Launch Overview
```
http://localhost:3000/d/genesis-launch-overview
```
**Primary launch dashboard** - Training progress, node-agent count, accuracy

### 🌐 Network Performance
```
http://localhost:3000/d/network-performance-health
```
**Network health** - Latency, throughput, connections

### 🔒 Consensus & Trust
```
http://localhost:3000/d/consensus-trust-monitoring
```
**Security monitoring** - Trust scores, TPM verification

---

## ✅ Success Checks

**30 seconds after launch:**
```bash
# Check active nodes
curl http://localhost:8000/api/network_status | jq '.active_nodes'
# Expected: >= 20
```

**5 minutes after launch:**
```bash
# Check training rounds
curl http://localhost:8000/api/metrics | jq '.round'
# Expected: >= 5

# Check accuracy
curl http://localhost:8000/api/metrics | jq '.accuracy'
# Expected: > 0.60
```

**30 minutes after launch:**
```bash
# Check convergence
curl http://localhost:8000/api/convergence_status | jq '.accuracy'
# Expected: > 0.85
```

---

## 🔧 Quick Commands

```bash
# View logs
docker compose logs -f

# Check node-agent count
docker ps | grep node-agent | wc -l

# Scale node-agents
docker compose -f docker-compose.production.yml up -d --scale node-agent=50

# Stop everything
docker compose -f docker-compose.production.yml -f docker-compose.monitoring.yml down --remove-orphans

# Restart
docker compose -f docker-compose.production.yml -f docker-compose.monitoring.yml restart
```

---

## 🚨 Troubleshooting

### Issue: Can't access Grafana
```bash
# Check if running
docker ps | grep grafana

# Restart monitoring
docker compose -f docker-compose.monitoring.yml restart
```

### Issue: Low node-agent count
```bash
# Check Docker resources
docker stats

# Increase node-agents manually
docker compose -f docker-compose.production.yml up -d --scale node-agent=30
```

### Issue: Backend errors
```bash
# View backend logs
docker compose logs backend

# Restart backend
docker compose -f docker-compose.production.yml restart backend
```

---

## 📞 Get Help

**Full Documentation**: [GENESIS_LAUNCH_GUIDE.md](GENESIS_LAUNCH_GUIDE.md)

**Common Issues**: See Troubleshooting section in full guide

**Emergency**: Stop all services with `docker compose -f docker-compose.production.yml -f docker-compose.monitoring.yml down --remove-orphans`

---

## 🎉 Success Criteria

After 30 minutes, verify:
- ✅ 20+ nodes online
- ✅ Model accuracy > 85%
- ✅ All trust scores > 75
- ✅ Dashboards showing data
- ✅ No critical alerts

**🚀 Genesis Launch Complete!**

---

*For detailed information, see [GENESIS_LAUNCH_GUIDE.md](GENESIS_LAUNCH_GUIDE.md)*
