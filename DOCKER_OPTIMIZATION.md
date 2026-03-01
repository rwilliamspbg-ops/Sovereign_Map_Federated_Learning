# Docker Optimization Guide - Sovereign Map Federated Learning

## Overview

This guide covers the optimized Docker setup for the Sovereign Map Byzantine-tolerant federated learning system.

## Quick Start

### Development (5 nodes, 2 minutes)
```bash
docker compose -f docker-compose.dev.yml up -d
```

**Access:**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8000
- Grafana: http://localhost:3001 (admin/dev)
- Prometheus: http://localhost:9090

### Production (50 nodes, 5 minutes)
```bash
docker compose -f docker-compose.production.yml up -d --scale node-agent=50
```

### Large-Scale (500+ nodes, 15 minutes)
```bash
docker compose -f docker-compose.large-scale.yml up -d --scale node-agent=500
```

## Architecture Overview

### Multi-Stage Build Benefits

**Backend (Python):**
- **Stage 1 (Builder)**: Compiles all dependencies
- **Stage 2 (Runtime)**: Includes only compiled packages
- **Result**: ~60% smaller image (from 1.8GB → 800MB)

**Frontend (Node.js):**
- **Stage 1 (Dependencies)**: `npm ci --frozen-lockfile` for reproducible builds
- **Stage 2 (Build)**: Builds React application with cached dependencies
- **Stage 3 (Runtime)**: Lightweight nginx serving pre-built assets
- **Result**: ~95% smaller runtime image (from 1.2GB → 60MB)

### Docker Compose Profiles

| Profile | Use Case | Nodes | Memory | Duration |
|---------|----------|-------|--------|----------|
| **dev.yml** | Development & testing | 1 | 2GB | 2 min |
| **production.yml** | Staging/QA | 50 | 4-6GB | 5 min |
| **large-scale.yml** | Production testnet | 500+ | 8-16GB+ | 15 min |

## Environment Variables

### Backend Configuration
```bash
# FL Configuration
NUM_NODES=100                    # Total nodes in cluster
NUM_ROUNDS=100                   # FL rounds to execute
FLASK_PORT=8000                  # Backend API port
FL_SERVER_MODE=true              # Enable FL aggregation

# Database
DATABASE_URI=mongodb://mongo:27017/sovereignmap
MONGO_USER=admin
MONGO_PASSWORD=sovereignmap

# Caching
REDIS_URL=redis://redis:6379/0
REDIS_PASSWORD=sovereignmap

# Monitoring
PROMETHEUS_RETENTION=30d         # Metrics retention
```

### Example .env File
```bash
# Copy from .env.example and customize
cp .env.example .env

# Edit .env with your settings
NUM_NODES=50
PROMETHEUS_RETENTION=60d
GRAFANA_PASSWORD=your_secure_password
```

## Performance Tuning

### For Large-Scale Deployments (1000+ nodes)

**MongoDB Optimization:**
```yaml
mongo:
  command: >
    --wiredTigerCacheSizeGB=2
    --dbPath=/data/db
    --replSet=rs0
    --journalCommitInterval=500
```

**Redis Optimization:**
```yaml
redis:
  command: redis-server --maxmemory 4gb --maxmemory-policy allkeys-lru
```

**Prometheus Optimization:**
```yaml
prometheus:
  command:
    - '--storage.tsdb.max-block-duration=6h'
    - '--query.max-concurrency=10'
    - '--storage.tsdb.retention.time=90d'
```

### Resource Limits

**Development:**
```yaml
backend:
  deploy:
    resources:
      limits:
        cpus: '1'
        memory: 1G
```

**Production (50 nodes):**
```yaml
backend:
  deploy:
    resources:
      limits:
        cpus: '2'
        memory: 2G
```

**Large-Scale (500+ nodes):**
```yaml
backend:
  deploy:
    resources:
      limits:
        cpus: '4'
        memory: 4G
```

## Volume Management

### Local Development
```bash
# Use local driver (default)
docker compose -f docker-compose.dev.yml up -d
```

### Production with External Storage
```yaml
volumes:
  mongo_data:
    driver: local
    driver_opts:
      type: nfs
      o: addr=192.168.1.100,vers=4,soft,timeo=180,bg,tcp,rw
      device: ":/export/sovereignmap/mongo"
```

### Backup Strategy
```bash
# Backup MongoDB
docker exec sovereignmap-mongo mongodump --out /backups/mongo_$(date +%Y%m%d)

# Backup Prometheus data
docker run --rm -v sovereignmap_prometheus_data:/data \
  alpine tar czf /backups/prometheus_$(date +%Y%m%d).tar.gz /data

# Restore MongoDB
docker exec sovereignmap-mongo mongorestore /backups/mongo_latest
```

## Networking

### Internal Network (`sovereignmap`)
- All services communicate via service names
- Automatic DNS resolution within Docker network
- Network policies can restrict traffic

### Port Mapping

| Service | Internal | External | Protocol |
|---------|----------|----------|----------|
| Backend | 8000 | 8000 | HTTP |
| Flower | 8080 | 8080 | gRPC |
| Frontend | 80 | 3000 | HTTP |
| Prometheus | 9090 | 9090 | HTTP |
| Grafana | 3000 | 3001 | HTTP |
| Alertmanager | 9093 | 9093 | HTTP |
| MongoDB | 27017 | 27017 | MongoDB |
| Redis | 6379 | 6379 | Redis |

### Multi-Machine Deployment

**Machine 1 (Backend):**
```bash
docker compose -f docker-compose.production.yml up -d
# Exposes ports 3000, 3001, 8000, 8080, 9090, 9093
```

**Machines 2-N (Node Agents):**
```bash
# Point to backend machine (IP: 192.168.1.100)
BACKEND_URL=http://192.168.1.100:8000 \
  docker compose -f docker-compose.production.yml up -d node-agent
```

## Logging

### Log Levels

**Development:**
```yaml
logging:
  driver: json-file
  options:
    max-size: "5m"
    max-file: "2"
```

**Production:**
```yaml
logging:
  driver: json-file
  options:
    max-size: "50m"
    max-file: "10"
    labels: "app=sovereignmap,env=production"
```

### View Logs
```bash
# Tail backend logs
docker compose logs -f backend

# View logs with timestamps
docker compose logs --timestamps backend

# Last 50 lines
docker compose logs --tail=50 backend

# View specific time range
docker compose logs --since 10m backend
```

### Log Aggregation (ELK Stack Alternative)
```bash
# Deploy with Loki (lightweight)
docker run -d \
  --name loki \
  -p 3100:3100 \
  grafana/loki:latest

# Configure docker-compose to use Loki driver
docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions

# Update docker-compose.yml
logging:
  driver: loki
  options:
    loki-url: "http://localhost:3100/loki/api/v1/push"
    loki-batch-size: "400"
```

## Security Best Practices

### 1. Non-Root User
```dockerfile
# Create non-root user in Dockerfile
RUN useradd -m -u 1001 appuser
USER appuser
```

### 2. Read-Only Filesystems
```yaml
backend:
  read_only: true
  tmpfs:
    - /tmp
    - /app/data  # Writable only this directory
```

### 3. Network Policies
```bash
# Restrict traffic to specific ports
docker network create --opt com.docker.network.driver.mtu=1500 sovereignmap
```

### 4. Secrets Management
```bash
# Use Docker secrets in Swarm mode
echo "sovereignmap_password" | docker secret create mongo_password -

# Or use environment file with strict permissions
echo "MONGO_PASSWORD=secure_pass" > .env
chmod 600 .env
docker compose --env-file .env up -d
```

### 5. Image Scanning
```bash
# Scan images for vulnerabilities
docker scan sovereignmap/backend:latest

# Or use Trivy
trivy image sovereignmap/backend:latest
```

## Troubleshooting

### Container Won't Start
```bash
# Check logs
docker compose logs backend

# Inspect image
docker inspect sovereignmap/backend:latest

# Run with interactive bash
docker run -it sovereignmap/backend:latest /bin/bash
```

### Network Issues
```bash
# Test DNS resolution within network
docker run --network sovereignmap busybox nslookup backend

# Check network connectivity
docker exec sovereignmap-backend curl backend:8000/health
```

### Memory Issues
```bash
# Check container memory usage
docker stats

# View resource limits
docker inspect sovereignmap-backend | grep -A 10 '"MemoryLimit"'

# Increase limits
docker compose up -d --compatibility  # Use compose v3 resource limits
```

### Database Issues
```bash
# Connect to MongoDB directly
docker exec -it sovereignmap-mongo mongosh -u admin -p

# Check Redis
docker exec -it sovereignmap-redis redis-cli ping

# Verify connections
docker logs sovereignmap-backend | grep -i "error\|failed"
```

## Scaling

### Add Nodes During Deployment
```bash
# Scale to 100 nodes
docker compose -f docker-compose.production.yml up -d --scale node-agent=100

# Add 50 more nodes
docker compose -f docker-compose.production.yml up -d --scale node-agent=150
```

### Monitor Scaling
```bash
# Watch convergence in real-time
watch -n 5 'curl -s http://localhost:8000/convergence | jq "{round: .current_round, accuracy: .current_accuracy, nodes: .active_nodes}"'

# Monitor resource usage
docker stats --no-stream
```

## Cleanup

### Stop Services
```bash
# Stop dev environment
docker compose -f docker-compose.dev.yml down

# Stop but keep volumes
docker compose -f docker-compose.dev.yml down --volumes

# Remove everything
docker compose -f docker-compose.dev.yml down -v --remove-orphans
```

### Prune Resources
```bash
# Remove dangling images/volumes/networks
docker system prune -a

# Remove all Sovereign Map resources
docker compose -f docker-compose.production.yml down -v
```

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Build and Push
on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: docker/setup-buildx-action@v2
      
      - uses: docker/build-push-action@v4
        with:
          context: .
          file: ./Dockerfile.backend.optimized
          push: true
          tags: sovereignmap/backend:latest
```

## References

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Dockerfile Best Practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [Docker Performance Tuning](https://docs.docker.com/config/containers/resource_constraints/)

---

**Last Updated**: 2024  
**Status**: Production Ready
