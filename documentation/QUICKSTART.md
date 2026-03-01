# 🚀 Quick Start Guide

Get Sovereign Map up and running in 5 minutes.

---

## Prerequisites

Ensure you have installed:
- **Docker** & **Docker Compose** (latest versions)
- **Python 3.10+** (for local development)
- **Git** (for cloning the repository)
- **16GB RAM** minimum
- **Linux/macOS/Windows with WSL2**

---

## Option 1: Docker Compose (Recommended) - 3 Minutes

### Step 1: Clone Repository
```bash
git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning
cd Sovereign_Map_Federated_Learning
```

### Step 2: Start Full Stack
```bash
docker compose -f docker/docker-compose.full.yml up -d
```

### Step 3: Verify Health
```bash
# Check all services running
docker compose -f docker/docker-compose.full.yml ps

# Expected output:
# NAME                 STATUS           PORTS
# fl-backend          Up 2 minutes      0.0.0.0:8000->8000/tcp
# prometheus          Up 2 minutes      0.0.0.0:9090->9090/tcp
# grafana             Up 2 minutes      0.0.0.0:3000->3000/tcp
# redis               Up 2 minutes      0.0.0.0:6379->6379/tcp
# alertmanager        Up 2 minutes      0.0.0.0:9093->9093/tcp
# nginx               Up 2 minutes      0.0.0.0:80->80/tcp
```

### Step 4: Access Dashboards
```
Backend API:         http://localhost:8000
Grafana Dashboard:   http://localhost:3000        (admin/admin)
Prometheus:          http://localhost:9090
AlertManager:        http://localhost:9093
```

### Step 5: Test the API
```bash
# Health check
curl http://localhost:8000/health

# Expected response:
# {"status": "healthy", "timestamp": "2026-02-24T..."}

# Get metrics
curl http://localhost:8000/metrics
```

Done! The complete stack is running locally.

---

## Option 2: Local Python Development - 2 Minutes

### Step 1: Clone & Navigate
```bash
git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning
cd Sovereign_Map_Federated_Learning
```

### Step 2: Install Dependencies
```bash
pip install -r requirements.txt
```

### Step 3: Start Backend
```bash
python src/sovereign_federation_backend.py
```

Expected output:
```
Starting Sovereign Federation Backend...
Flask app listening on 0.0.0.0:8000
WebSocket server ready at ws://localhost:8000/ws/metrics
```

### Step 4: Test API
```bash
# In another terminal
curl http://localhost:8000/health
```

Done! Backend API is running locally.

---

## Option 3: Kubernetes Deployment - 5 Minutes

### Step 1: Create Namespace
```bash
kubectl create namespace sovereign-map
```

### Step 2: Apply Manifests
```bash
kubectl apply -f config/kubernetes/ -n sovereign-map
```

### Step 3: Verify Deployment
```bash
kubectl get pods -n sovereign-map

# Wait for all pods to be Ready
kubectl wait --for=condition=ready pod -l app=fl-backend -n sovereign-map --timeout=300s
```

### Step 4: Port Forward
```bash
kubectl port-forward -n sovereign-map svc/fl-backend 8000:8000 &
kubectl port-forward -n sovereign-map svc/grafana 3000:3000 &
```

### Step 5: Access Services
```
Backend:  http://localhost:8000
Grafana:  http://localhost:3000
```

---

## Next Steps

### Run a Test
```bash
# Run 100K node scaling test
python tests/bft_week2_100k_nodes.py

# Run Byzantine boundary test
python tests/bft_week2_100k_byzantine_boundary.py
```

Expected results:
- ✅ 100K nodes scale validated
- ✅ 50% Byzantine tolerance proven
- ✅ Recovery times logged
- ✅ Metrics exported to Prometheus

### Review Results
```bash
# Results saved to:
# - recovery_times_by_byzantine_pct.csv
# - convergence_curves.csv
# - accuracy_floor_measurements.txt
```

### View Monitoring Dashboard
1. Open http://localhost:3000
2. Login with `admin/admin`
3. Navigate to "Dashboards" → "Byzantine Metrics"
4. View real-time metrics

---

## Common Commands

### Docker Compose

```bash
# View logs
docker compose -f docker/docker-compose.full.yml logs -f

# Stop all services
docker compose -f docker/docker-compose.full.yml down

# Rebuild images
docker compose -f docker/docker-compose.full.yml build

# View specific service logs
docker compose -f docker/docker-compose.full.yml logs fl-backend
```

### Python Tests

```bash
# Run all tests
pytest tests/ -v

# Run specific test
python tests/bft_week2_100k_nodes.py

# Run with coverage
pytest tests/ --cov=src --cov-report=html
```

### Kubernetes

```bash
# View pod logs
kubectl logs -n sovereign-map deployment/fl-backend

# Get pod details
kubectl describe pod -n sovereign-map -l app=fl-backend

# Scale deployment
kubectl scale deployment fl-backend --replicas=3 -n sovereign-map

# Delete deployment
kubectl delete -f config/kubernetes/ -n sovereign-map
```

---

## Troubleshooting

### Port Already in Use
```bash
# Find process using port 8000
lsof -i :8000

# Kill process
kill -9 <PID>

# Or use different port
docker run -p 8001:8000 ...
```

### Docker Compose Won't Start
```bash
# Check docker daemon
docker ps

# If not running, start Docker Desktop (macOS/Windows)
# Or restart service (Linux):
sudo systemctl restart docker
```

### Memory Issues
```bash
# Increase Docker memory limit
# Docker Desktop → Settings → Resources → Memory: 8GB+

# Or check system resources
docker stats
```

### API Not Responding
```bash
# Check container status
docker ps -a

# View logs
docker logs fl-backend

# Restart container
docker restart fl-backend
```

---

## Configuration

### Environment Variables

Create `.env` file in root:

```bash
# Backend
BACKEND_PORT=8000
BACKEND_HOST=0.0.0.0

# Threat Analysis
GEMINI_API_KEY=your_key_here  # Optional, for AI threat analysis

# Monitoring
PROMETHEUS_RETENTION=15d
GRAFANA_ADMIN_PASSWORD=admin

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# Logging
LOG_LEVEL=INFO
```

### Docker Compose Override

Create `docker-compose.override.yml` for local customizations:

```yaml
version: '3.8'
services:
  fl-backend:
    environment:
      LOG_LEVEL: DEBUG
    ports:
      - "8001:8000"
    volumes:
      - ./src:/app/src  # Hot reload for development
```

---

## Performance Baselines

Expected performance on standard hardware (16GB RAM):

| Operation | Time | Status |
|-----------|------|--------|
| Docker stack startup | ~30s | ✅ |
| Python backend startup | ~5s | ✅ |
| 100K node test | ~2-3 mins | ✅ |
| Consensus round | <500ms | ✅ |
| Metrics export | ~1s | ✅ |
| Dashboard load | ~2s | ✅ |

---

## Next: Deep Dive Topics

After quickstart, explore:

- **[Architecture](ARCHITECTURE.md)** - System design deep dive
- **[API Reference](API_REFERENCE.md)** - Endpoint documentation
- **[Testing](TESTING.md)** - Test framework guide
- **[Deployment](DEPLOYMENT.md)** - Production deployment
- **[Research](RESEARCH_FINDINGS.md)** - Byzantine analysis results

---

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/issues)
- **Discussions**: [GitHub Discussions](https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning/discussions)
- **Email**: [team@sovereignmap.network](mailto:team@sovereignmap.network)

---

**Ready to dive deeper?** Read [ARCHITECTURE.md](ARCHITECTURE.md) next!
