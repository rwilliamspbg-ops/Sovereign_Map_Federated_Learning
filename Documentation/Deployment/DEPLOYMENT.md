# Sovereign Map - Complete Deployment Guide

This guide provides step-by-step instructions for deploying Sovereign Map in development, staging, and production environments.

**Table of Contents**
- [Prerequisites](#prerequisites)
- [Local Development](#local-development)
- [Production Deployment](#production-deployment)
- [GitHub Actions CI/CD](#github-actions-cicd)
- [Scaling](#scaling)
- [Monitoring & Maintenance](#monitoring--maintenance)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

**Minimum (Development):**
- Docker 20.10+
- Docker Compose 2.0+
- 8GB RAM
- 20GB disk space
- Python 3.11+

**Recommended (Production):**
- Docker 24.0+ LTS
- Docker Compose 2.20+
- 16GB+ RAM
- 100GB+ SSD
- Dedicated Linux machine
- Static IP address

### Software

```bash
# Install Docker (Ubuntu/Debian)
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker --version
docker compose version
```

### Network Setup

```bash
# Create Docker network (required once per machine)
docker network create sovereign-network

# Verify
docker network ls | grep sovereign-network
```

## Local Development

### 1. Clone Repository

```bash
git clone https://github.com/rwilliamspbg-ops/Sovereign_Map_Federated_Learning.git
cd Sovereign_Map_Federated_Learning

# Verify files exist
ls -la docker-compose.*.yml
```

### 2. Quick Start (10 Nodes)

```bash
# Start the complete stack
docker compose -f docker-compose.full.yml up -d

# Wait for services to be ready (30-60 seconds)
docker compose -f docker-compose.full.yml ps

# Verify all services are 'running'
```

### 3. Access Services

```bash
# Backend API
curl http://localhost:8000/health

# Grafana Dashboard
open http://localhost:3000
# Login: admin / admin

# Prometheus Metrics
open http://localhost:9090

# Check convergence
curl http://localhost:8000/convergence | jq
```

### 4. View Logs

```bash
# Backend logs
docker compose -f docker-compose.full.yml logs -f backend

# Node agent logs
docker compose -f docker-compose.full.yml logs -f node-agent

# All services
docker compose -f docker-compose.full.yml logs -f
```

### 5. Scale Nodes (During Development)

```bash
# Start with 50 nodes
docker compose -f docker-compose.full.yml up -d --scale node-agent=50

# Monitor scaling
watch -n 1 'docker compose -f docker-compose.full.yml ps'

# Scale up to 100
docker compose -f docker-compose.full.yml up -d --scale node-agent=100

# Scale down to 25
docker compose -f docker-compose.full.yml up -d --scale node-agent=25
```

### 6. Stop Services

```bash
# Stop all services (containers preserved)
docker compose -f docker-compose.full.yml stop

# Stop and remove containers
docker compose -f docker-compose.full.yml down

# Stop and remove containers + volumes
docker compose -f docker-compose.full.yml down -v

# Remove images too
docker compose -f docker-compose.full.yml down --rmi all
```

## Production Deployment

### 1. Pre-Deployment Checklist

```bash
# Check system resources
free -h                      # RAM
df -h                        # Disk space
nproc                        # CPU cores
uname -a                     # OS info

# Verify Docker installation
docker ps
docker network ls
```

### 2. Configure Production Environment

Create `.env` file:

```bash
# Production Settings
FLASK_ENV=production
NUM_NODES=100
PROMETHEUS_RETENTION=60d
GRAFANA_ADMIN_PASSWORD=your-secure-password
ALERTMANAGER_EMAIL=alerts@example.com

# Security
TRUST_CACHE_TTL=3600
CERT_DIR=/etc/sovereign/certs
```

### 3. Deploy Full Stack

```bash
# Load environment variables
export $(cat .env | xargs)

# Create TPM certificate volume (first time only)
docker volume create tpm-certs

# Start monitoring stack (first)
docker compose -f docker-compose.full.yml up -d

# Start TPM security layer
docker compose -f docker-compose.full.yml up -d

# Start main application (last)
docker compose -f docker-compose.full.yml up -d --scale node-agent=100

# Verify all services running
docker compose -f docker-compose.full.yml ps
```

### 4. Health Checks

```bash
# Backend health
curl http://localhost:8000/health

# TPM metrics
curl http://localhost:9091/health

# Prometheus scrape targets
curl http://localhost:9090/api/v1/targets

# Check all metrics exist
curl http://localhost:9091/metrics | grep tpm_certificates_total
```

### 5. Verify Convergence

```bash
# Check initial accuracy
curl http://localhost:8000/convergence | jq '.current_accuracy'

# Check after 5 minutes
sleep 300
curl http://localhost:8000/convergence | jq '.current_accuracy'

# Should show improvement
```

### 6. Configure Alertmanager

Create `alertmanager.yml`:

```yaml
global:
  resolve_timeout: 5m
  smtp_smarthost: smtp.gmail.com:587
  smtp_auth_username: your-email@gmail.com
  smtp_auth_password: your-app-password
  smtp_from: alerts@example.com

route:
  receiver: security-team
  group_by: ['alertname', 'cluster']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h

receivers:
  - name: security-team
    email_configs:
      - to: security@example.com
        headers:
          Subject: 'TPM Alert: {{ .GroupLabels.alertname }}'
```

Mount in compose override:

```yaml
version: '3.9'
services:
  alertmanager:
    volumes:
      - ./alertmanager.yml:/etc/alertmanager/config.yml
```

### 7. Enable SSL/TLS (Optional)

```bash
# For HTTPS access to Grafana
# Use nginx reverse proxy or AWS ALB

# nginx example:
upstream grafana {
    server 127.0.0.1:3000;
}

server {
    listen 443 ssl http2;
    server_name monitoring.example.com;
    
    ssl_certificate /etc/letsencrypt/live/monitoring.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/monitoring.example.com/privkey.pem;
    
    location / {
        proxy_pass http://grafana;
        proxy_set_header Host $host;
    }
}
```

## GitHub Actions CI/CD

### 1. Generate SSH Keys for Deployment

```bash
# Generate deployment key
ssh-keygen -t rsa -b 4096 -f deploy_key -N ""

# Display private key (for GitHub Secrets)
cat deploy_key

# Display public key (for authorized_keys)
cat deploy_key.pub
```

### 2. Add SSH Key to GitHub

**In GitHub Repository:**
1. Go to Settings → Deploy Keys
2. Click "Add deploy key"
3. Paste public key content (`deploy_key.pub`)
4. Check "Allow write access"
5. Click "Add key"

**Or create a GitHub Personal Access Token:**
1. Go to Settings → Developer settings → Personal access tokens
2. Click "Generate new token"
3. Select `repo` scope
4. Save token

### 3. Create GitHub Secrets

**In GitHub Repository:**
1. Go to Settings → Secrets and variables → Actions
2. Click "New repository secret"

Add these secrets:

```
DEPLOY_KEY          = [contents of deploy_key]
DEPLOY_HOST         = production.example.com
DEPLOY_USER         = ubuntu
DEPLOY_PATH         = /home/ubuntu/sovereign-map
```

### 4. Create GitHub Actions Workflow

Create `.github/workflows/deploy.yml`:

```yaml
name: Build & Deploy

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'
      
      - name: Install dependencies
        run: |
          python -m pip install --upgrade pip
          pip install -r requirements.txt
      
      - name: Run tests
        run: |
          python tests/scale-tests/bft_extreme_scale_10m.py
          python tests/byzantine-tests/byzantine_tolerance_test.py
      
      - name: Build Docker images
        run: |
          docker build -f Dockerfile -t sovereign-map:${{ github.sha }} .
          docker tag sovereign-map:${{ github.sha }} sovereign-map:latest

  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Deploy to production
        uses: appleboy/ssh-action@master
        with:
          host: ${{ secrets.DEPLOY_HOST }}
          username: ${{ secrets.DEPLOY_USER }}
          key: ${{ secrets.DEPLOY_KEY }}
          script: |
            cd ${{ secrets.DEPLOY_PATH }}
            git pull origin main
            docker compose -f docker-compose.full.yml pull
            docker compose -f docker-compose.full.yml up -d --scale node-agent=100
            docker compose -f docker-compose.full.yml ps
      
      - name: Verify deployment
        run: |
          for i in {1..30}; do
            if curl -f http://${{ secrets.DEPLOY_HOST }}:8000/health; then
              echo "Deployment successful"
              exit 0
            fi
            echo "Waiting for service... ($i/30)"
            sleep 10
          done
          exit 1
```

### 5. Run Deployment

**Manual Trigger:**
1. Go to GitHub Actions tab
2. Select "Build & Deploy" workflow
3. Click "Run workflow"
4. Select branch (main)
5. Click "Run workflow"

**Automatic on Push:**
```bash
# Any push to main branch triggers deployment
git push origin main
```

### 6. Monitor Deployment

```bash
# Watch GitHub Actions
# Go to Actions tab → Latest workflow run

# SSH to production and check
ssh ubuntu@production.example.com
cd /home/ubuntu/sovereign-map
docker compose ps
docker compose logs -f backend
```

## Scaling

### Horizontal Scaling (Add More Nodes)

```bash
# Current status
docker compose -f docker-compose.full.yml ps

# Scale up
docker compose -f docker-compose.full.yml up -d --scale node-agent=500

# Watch scaling
watch -n 2 'docker compose -f docker-compose.full.yml ps | tail -20'
```

### Vertical Scaling (Increase Resources)

Create `docker-compose.override.yml`:

```yaml
version: '3.9'
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 8G
        reservations:
          cpus: '2'
          memory: 4G

  prometheus:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 4G
```

Apply:
```bash
docker compose up -d
```

### Multi-Machine Deployment

**Machine 1 (Backend & Monitoring):**
```bash
docker compose -f docker-compose.full.yml up -d
```

**Machines 2-N (Node Agents Only):**
```bash
# Install Docker on each machine
# Create sovereign-network
docker network create sovereign-network

# Modify Docker Compose to point to Machine 1
export BACKEND_URL=http://machine1-ip:8000

# Start only node-agent service
docker compose -f docker-compose.full.yml up -d node-agent
```

## Monitoring & Maintenance

### Daily Tasks

```bash
# Check system health
docker compose ps
docker system df

# Review logs for errors
docker compose logs | grep -i error

# Monitor convergence
curl http://localhost:8000/convergence | jq '.current_accuracy'

# Check certificate expiry
curl http://localhost:9091/metrics/summary | jq '.node_details'
```

### Weekly Tasks

```bash
# Review alert history
curl http://localhost:9090/api/v1/alerts | jq '.data'

# Clean up old logs
docker system prune -a

# Backup metrics database
docker exec sovereign-prometheus tar czf /prometheus-backup.tar.gz /prometheus/
docker cp sovereign-prometheus:/prometheus-backup.tar.gz ./backups/
```

### Monthly Tasks

```bash
# Rotate certificates (if not automated)
# Update dependencies
docker compose pull
docker compose up -d

# Review performance metrics
# Archive logs and metrics
# Generate compliance report
```

### Backup Strategy

```bash
# Backup volumes
docker run --rm \
  -v tpm-certs:/data \
  -v $(pwd)/backups:/backup \
  alpine tar czf /backup/tpm-certs-backup.tar.gz -C /data .

# Backup Prometheus data
docker run --rm \
  -v prometheus-data:/data \
  -v $(pwd)/backups:/backup \
  alpine tar czf /backup/prometheus-backup.tar.gz -C /data .

# Restore
docker run --rm \
  -v tpm-certs:/data \
  -v $(pwd)/backups:/backup \
  alpine tar xzf /backup/tpm-certs-backup.tar.gz -C /data
```

## Troubleshooting

### Problem: Services won't start

```bash
# Check Docker daemon
sudo systemctl status docker

# View system logs
sudo journalctl -xu docker.service -n 50

# Check resource availability
free -h
df -h

# Rebuild images
docker compose build --no-cache
docker compose up -d
```

### Problem: High memory usage

```bash
# Check which container is using memory
docker stats

# Limit container memory in override
services:
  backend:
    deploy:
      resources:
        limits:
          memory: 4G

# Reduce number of nodes
docker compose up -d --scale node-agent=50
```

### Problem: Metrics not appearing

```bash
# Check exporter health
curl http://localhost:9091/health

# Check Prometheus scrape targets
curl http://localhost:9090/api/v1/targets

# View exporter logs
docker logs sovereign-tpm-metrics

# Manually scrape
curl http://localhost:9091/metrics | head -50
```

### Problem: Certificate errors

```bash
# Check TPM volume mounted
docker inspect sovereign-backend | grep -A5 Mounts

# Verify certificates exist
docker exec sovereign-backend ls -la /etc/sovereign/certs/

# Regenerate certificates
docker exec sovereign-tpm-ca python tpm_cert_manager.py

# Restart affected services
docker compose restart backend node-agent
```

## Performance Tuning

### Optimize Prometheus

```yaml
# prometheus.yml
global:
  scrape_interval: 30s          # Increase for less load
  evaluation_interval: 30s
  external_labels:
    cluster: production

# Keep less data
--storage.tsdb.retention.time=15d
```

### Optimize Grafana

```bash
# Use CDN for Grafana assets
# Increase dashboard cache
# Disable unused data sources
# Use query caching
```

### Optimize Backend

```python
# sovereignmap_production_backend_v2.py
# Increase FL round interval (default 30s)
time.sleep(60)  # Every minute instead

# Reduce node count for testing
initialize_system(num_nodes=50)
```

## Post-Deployment Checklist

- [ ] All services running (`docker compose ps`)
- [ ] Backend health check passing (`curl http://localhost:8000/health`)
- [ ] Grafana dashboard accessible
- [ ] Prometheus metrics being collected
- [ ] Alertmanager configured
- [ ] SSL/TLS configured (if required)
- [ ] Backups configured
- [ ] Monitoring alerts set up
- [ ] Documentation reviewed
- [ ] Team trained on operations

## Support

For deployment issues:
1. Check logs: `docker compose logs -f service-name`
2. Review troubleshooting section above
3. Check GitHub Issues
4. Contact support team

---

**Last Updated**: February 2024  
**Deployment Version**: v1.0.0
