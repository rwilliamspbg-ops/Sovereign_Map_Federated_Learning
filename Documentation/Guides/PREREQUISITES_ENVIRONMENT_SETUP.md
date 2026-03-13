# Prerequisites & Environment Setup Guide

**System:** Sovereign Map Federated Learning with Mohawk Runtime
**Date:** 2026-03-01
**Status:** Production-Ready Setup Instructions

---

## Table of Contents

1. [System Requirements](#system-requirements)
2. [Docker Engine Setup](#docker-engine-setup)
3. [Monitoring Stack Installation](#monitoring-stack-installation)
4. [TPM 2.0 Configuration](#tpm-20-configuration)
5. [Verification & Health Checks](#verification--health-checks)
6. [Environment Configuration](#environment-configuration)
7. [Troubleshooting](#troubleshooting)

---

## System Requirements

### Minimum Hardware Specifications

```
CPU:        8+ cores (16+ recommended)
RAM:        16 GB minimum (32 GB recommended)
Storage:    100 GB SSD minimum (500 GB recommended)
Network:    1 Gbps Ethernet
TPM:        2.0 module (discrete or firmware)
```

### Your System (AMD Ryzen AI 7 350)
```
CPU:        31 logical cores (excellent)
RAM:        32 GB (perfect)
Storage:    761 GB free (excellent)
TPM:        Integrated in Ryzen AI (ready)
Network:    Docker bridge networks (ready)
Status:     ✅ EXCEEDS requirements
```

### Operating System Support
- ✅ Windows 11 (with WSL2)
- ✅ Windows Server 2019+
- ✅ Linux (Ubuntu 20.04+, Debian 11+)
- ✅ macOS 12+ (Intel & Apple Silicon)

---

## Docker Engine Setup

### 1. Install Docker Desktop

#### Windows 11 / Windows Server
**Option A: Docker Desktop GUI**
```powershell
# Download from: https://www.docker.com/products/docker-desktop

# Or install via Chocolatey
choco install docker-desktop -y

# After installation, restart system
Restart-Computer -Force
```

**Option B: WSL2 with Docker**
```powershell
# Enable WSL2 if not already enabled
wsl --install
wsl --set-default-version 2

# Install Docker Desktop (includes WSL2 backend)
# Download: https://www.docker.com/products/docker-desktop
```

#### Linux (Ubuntu/Debian)
```bash
# Update package list
sudo apt-get update

# Install Docker
sudo apt-get install -y docker.io docker-compose

# Start Docker daemon
sudo systemctl start docker
sudo systemctl enable docker

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker
```

#### macOS
```bash
# Via Homebrew
brew install --cask docker

# Or download: https://www.docker.com/products/docker-desktop
# Launch from Applications
```

---

### 2. Docker Configuration

#### Verify Installation
```bash
docker --version
docker-compose --version
docker ps

# Expected output:
# Docker version 25.0.0+
# Docker Compose version 2.20.0+
# No containers running (if first time)
```

#### Configure Resource Limits

**Windows (WSL2):**
Create/Edit `%USERPROFILE%\.wslconfig`
```ini
[wsl2]
memory=24GB
processors=24
swap=4GB
localhostForwarding=true
```

Restart WSL:
```powershell
wsl --shutdown
wsl --launch
```

**Linux:**
Edit `/etc/docker/daemon.json`
```json
{
  "storage-driver": "overlay2",
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "insecure-registries": [],
  "registry-mirrors": []
}
```

Restart Docker:
```bash
sudo systemctl restart docker
```

**macOS:**
Docker Desktop > Preferences > Resources
```
CPUs:       24
Memory:     24 GB
Disk:       200 GB
Swap:       4 GB
```

#### Enable GPU Support (if available)

**For NVIDIA GPUs:**
```bash
# Install NVIDIA Container Toolkit
distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
curl -s -L https://nvidia.github.io/nvidia-docker/gpgkey | sudo apt-key add -
curl -s -L https://nvidia.github.io/nvidia-docker/$distribution/nvidia-docker.list | \
  sudo tee /etc/apt/sources.list.d/nvidia-docker.list

sudo apt-get update && sudo apt-get install -y nvidia-container-toolkit
sudo systemctl restart docker
```

**For AMD GPUs (ROCm):**
```bash
# Install ROCm
wget -q0 - https://repo.radeon.com/rocm/rocm.gpg.key | sudo apt-key add -
echo 'deb [arch=amd64] https://repo.radeon.com/rocm/apt/debian focal main' | \
  sudo tee /etc/apt/sources.list.d/rocm.list

sudo apt-get update
sudo apt-get install -y rocm-dkms
```

**Test GPU Access:**
```bash
docker run --rm --gpus all nvidia/cuda:12.0-runtime-ubuntu22.04 nvidia-smi
# Or for AMD:
docker run --rm --device /dev/kfd --device /dev/dri rocm/rocm-terminal rocm-smi
```

---

### 3. Docker Network Configuration

**Create Sovereign Map Network:**
```bash
docker network create sovereignmap \
  --driver bridge \
  --subnet 172.28.0.0/16 \
  --opt com.docker.network.driver.mtu=1500
```

**Verify Network:**
```bash
docker network inspect sovereignmap
# Should show subnet: 172.28.0.0/16
```

---

## Monitoring Stack Installation

### 1. Prometheus Installation

**Using Docker:**
```bash
# Create Prometheus config
mkdir -p monitoring/prometheus

# Download config (or create custom)
docker pull prom/prometheus:v2.48.0

# Run Prometheus
docker run -d \
  --name sovereignmap-prometheus \
  --network sovereignmap \
  -p 9090:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml:ro \
  -v prometheus_data:/prometheus \
  prom/prometheus:v2.48.0 \
  --config.file=/etc/prometheus/prometheus.yml \
  --storage.tsdb.path=/prometheus \
  --storage.tsdb.retention.time=30d
```

**Prometheus Configuration (prometheus.yml):**
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'sovereign-map'
    environment: 'production'

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'sovereign-backend'
    static_configs:
      - targets: ['backend:8000']
    metrics_path: '/metrics'
    scrape_interval: 10s

  - job_name: 'docker'
    static_configs:
      - targets: ['localhost:9323']
    scrape_interval: 30s

  - job_name: 'node'
    static_configs:
      - targets: ['localhost:9100']
    scrape_interval: 30s
```

**Verify Prometheus:**
```bash
# Access web UI
http://localhost:9090

# Check status
curl http://localhost:9090/api/v1/status/config
```

---

### 2. Grafana Installation

**Using Docker:**
```bash
# Create Grafana data directory
mkdir -p monitoring/grafana

# Run Grafana
docker run -d \
  --name sovereignmap-grafana \
  --network sovereignmap \
  -p 3001:3000 \
  -e GF_SECURITY_ADMIN_PASSWORD="${GRAFANA_ADMIN_PASSWORD:-CHANGE_ME_GRAFANA}" \
  -e GF_SECURITY_ADMIN_USER=admin \
  -e GF_INSTALL_PLUGINS=grafana-piechart-panel \
  -v grafana_data:/var/lib/grafana \
  -v $(pwd)/grafana/provisioning:/etc/grafana/provisioning:ro \
  grafana/grafana:10.2-alpine
```

**Grafana Configuration (grafana/provisioning/datasources/prometheus.yaml):**
```yaml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://sovereignmap-prometheus:9090
    isDefault: true
    editable: true
    jsonData:
      timeInterval: 15s
```

**Verify Grafana:**
```bash
# Access web UI
http://localhost:3001
# Username: admin
# Password: use your configured `GRAFANA_ADMIN_PASSWORD`

# Check data source
curl http://localhost:3001/api/datasources
```

---

### 3. Node Exporter (System Metrics)

```bash
# Run Node Exporter for system metrics
docker run -d \
  --name sovereignmap-node-exporter \
  --network sovereignmap \
  -p 9100:9100 \
  --volume /proc:/host/proc:ro \
  --volume /sys:/host/sys:ro \
  --volume /:/rootfs:ro \
  prom/node-exporter:v1.6.1 \
  --path.procfs=/host/proc \
  --path.sysfs=/host/sys \
  --collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)
```

---

### 4. Alertmanager (Alert Management)

```bash
# Create Alertmanager config directory
mkdir -p monitoring/alertmanager

# Run Alertmanager
docker run -d \
  --name sovereignmap-alertmanager \
  --network sovereignmap \
  -p 9093:9093 \
  -v $(pwd)/alertmanager.yml:/etc/alertmanager/alertmanager.yml:ro \
  prom/alertmanager:v0.26.0 \
  --config.file=/etc/alertmanager/alertmanager.yml \
  --storage.path=/alertmanager
```

---

## TPM 2.0 Configuration

### 1. Check TPM Availability

#### Windows
```powershell
# Check TPM status
Get-WmiObject -Namespace "root\cimv2\security\microsofttpm" `
  -Class Win32_Tpm | Select-Object IsActivated_InitializationStatus

# Expected: True
```

#### Linux
```bash
# Check TPM device
ls -la /dev/tpm* /dev/tpmrm*

# Expected: /dev/tpm0, /dev/tpmrm0

# Check TPM version
tpm2_getcap handles-persistent
```

#### macOS
```bash
# macOS typically requires external TPM USB device
system_profiler SPiBridgeDataType | grep -i tpm
```

---

### 2. Enable TPM in BIOS/UEFI

**For AMD Ryzen AI 7 350:**

1. **Restart and enter BIOS** (Delete key during boot)
2. **Navigate to Security > TPM settings**
3. **Enable TPM 2.0:**
   - "Security Chip" or "TPM" = Enabled
   - "TPM 2.0" = Enabled
   - "PTT (Platform Trust Technology)" = Enabled (if available)
4. **Save and exit**

---

### 3. Enable TPM Passthrough in WSL2

**Edit `.wslconfig`:**
```ini
[wsl2]
memory=24GB
processors=24
swap=4GB
localhostForwarding=true
# Enable TPM passthrough
tpm=true
```

**Restart WSL:**
```powershell
wsl --shutdown
wsl --launch
```

---

### 4. Docker TPM Device Access

**Enable Docker container access to TPM:**

Create `docker-compose.override.yml`:
```yaml
version: '3.9'

services:
  backend:
    devices:
      - /dev/tpm0:/dev/tpm0
      - /dev/tpmrm0:/dev/tpmrm0
    cap_add:
      - SYS_ADMIN
```

**Or run with docker run:**
```bash
docker run -d \
  --name sovereignmap-backend \
  --device /dev/tpm0:/dev/tpm0 \
  --device /dev/tpmrm0:/dev/tpmrm0 \
  --cap-add SYS_ADMIN \
  sovereignmap/backend:latest
```

---

### 5. TPM Utilities Installation

**Linux:**
```bash
# Install TPM2 tools
sudo apt-get install -y tpm2-tools tpm2-tss libtpm2-pkcs11-0

# Verify TPM
tpm2_getcap handles-persistent
tpm2_nvlist
```

**Windows (PowerShell):**
```powershell
# TPM is built-in, verify:
Get-Service TPM | Select-Object Status
# Expected: Running
```

---

### 6. Verify TPM in Containers

**Test TPM access:**
```bash
# Create test container
docker run -it \
  --device /dev/tpm0:/dev/tpm0 \
  --device /dev/tpmrm0:/dev/tpmrm0 \
  ubuntu:22.04 \
  bash

# Inside container:
apt-get update && apt-get install -y tpm2-tools
tpm2_getcap handles-persistent
exit
```

---

## Verification & Health Checks

### 1. Docker Health Check

```bash
# Check Docker daemon
docker info

# Expected output includes:
# Containers: X
# Images: Y
# Storage Driver: overlay2
```

**Full verification script:**
```bash
#!/bin/bash
echo "=== Docker Health Check ==="
echo "✓ Docker version: $(docker --version)"
echo "✓ Docker Compose: $(docker-compose --version)"
echo "✓ Containers running: $(docker ps -q | wc -l)"
echo "✓ Docker network: $(docker network ls | grep sovereignmap)"
echo "✓ Disk usage: $(docker system df | tail -1)"
```

---

### 2. Monitoring Stack Health Check

```bash
#!/bin/bash
echo "=== Monitoring Stack Health ==="

# Check Prometheus
echo -n "Prometheus: "
curl -s http://localhost:9090/-/healthy > /dev/null && echo "✓" || echo "✗"

# Check Grafana
echo -n "Grafana: "
curl -s http://localhost:3001/api/health > /dev/null && echo "✓" || echo "✗"

# Check Node Exporter
echo -n "Node Exporter: "
curl -s http://localhost:9100/metrics > /dev/null && echo "✓" || echo "✗"

# Check Alertmanager
echo -n "Alertmanager: "
curl -s http://localhost:9093/-/healthy > /dev/null && echo "✓" || echo "✗"

# Check Prometheus datasource in Grafana
echo -n "Prometheus datasource: "
curl -s http://localhost:3001/api/datasources | grep -q prometheus && echo "✓" || echo "✗"
```

---

### 3. TPM Health Check

```bash
#!/bin/bash
echo "=== TPM 2.0 Health Check ==="

# Linux
if [ -f /dev/tpm0 ]; then
  echo "✓ TPM device found: /dev/tpm0"
  tpm2_getcap handles-persistent 2>/dev/null | grep -q "0x" && echo "✓ TPM accessible" || echo "✗ TPM not accessible"
fi

# Windows (PowerShell)
# Get-Service TPM | Select-Object Status
# Get-WmiObject -Namespace "root\cimv2\security\microsofttpm" -Class Win32_Tpm
```

---

### 4. Network Connectivity Check

```bash
#!/bin/bash
echo "=== Network Health Check ==="

# Check sovereignmap network
docker network inspect sovereignmap > /dev/null 2>&1 && echo "✓ Network 'sovereignmap' exists" || echo "✗ Network missing"

# Check container connectivity
docker run --rm --network sovereignmap busybox ping -c 1 sovereignmap-prometheus 2>/dev/null && \
  echo "✓ Container connectivity OK" || echo "✗ Connectivity issue"
```

---

## Environment Configuration

### 1. Environment Variables

**Create `.env` file:**
```bash
# Docker Configuration
DOCKER_REGISTRY=docker.io
DOCKER_NAMESPACE=sovereignmap

# Grafana Configuration
GRAFANA_USER=admin
GRAFANA_PASSWORD=CHANGE_ME_GRAFANA
GRAFANA_ADMIN_UID=admin
GF_INSTALL_PLUGINS=grafana-piechart-panel

# Prometheus Configuration
PROMETHEUS_RETENTION=30d
PROMETHEUS_SCRAPE_INTERVAL=15s
PROMETHEUS_EVALUATION_INTERVAL=15s

# Backend Configuration
FLASK_ENV=production
NODE_ID=0
NUM_NODES=100
NUM_ROUNDS=200
FLASK_PORT=8000
PYTHONUNBUFFERED=1

# TPM Configuration
TPM_ENABLED=true
TPM_DEVICE=/dev/tpm0
TPM_VERIFY=true

# Monitoring Configuration
ENABLE_MONITORING=true
ENABLE_ALERTING=true
ENABLE_PROMETHEUS=true
ENABLE_GRAFANA=true
```

**Load environment:**
```bash
export $(cat .env | xargs)
```

---

### 2. Docker Compose Override

**Create `docker-compose.override.yml` for local development:**
```yaml
version: '3.9'

services:
  backend:
    devices:
      - /dev/tpm0:/dev/tpm0
      - /dev/tpmrm0:/dev/tpmrm0
    environment:
      - TPM_ENABLED=true
      - DEBUG=true
    ports:
      - "8000:8000"
      - "8080:8080"

  prometheus:
    ports:
      - "9090:9090"

  grafana:
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD:-CHANGE_ME_GRAFANA}

  node-exporter:
    ports:
      - "9100:9100"

  alertmanager:
    ports:
      - "9093:9093"
```

---

## Troubleshooting

### Issue: Docker daemon not running

**Windows:**
```powershell
# Check if Docker Desktop is running
Get-Process | Select-String docker

# If not running, start Docker Desktop
Start-Process "C:\Program Files\Docker\Docker\Docker.exe"

# Wait for daemon to start
Start-Sleep -Seconds 30
docker ps
```

**Linux:**
```bash
# Check Docker service status
sudo systemctl status docker

# Start if not running
sudo systemctl start docker

# Enable on boot
sudo systemctl enable docker
```

---

### Issue: TPM device not found

**Verify BIOS settings:**
1. Restart and enter BIOS
2. Navigate to Security > TPM
3. Ensure "TPM 2.0" is set to "Enabled"
4. Save and exit

**Check device:**
```bash
# Linux
ls -la /dev/tpm* /dev/tpmrm*

# If missing, load TPM driver
sudo modprobe tpm_tis
ls -la /dev/tpm0
```

**Windows:**
```powershell
# Verify TPM service
Get-Service TPM | Select-Object Status

# If stopped, start it
Start-Service TPM
```

---

### Issue: Prometheus not scraping metrics

**Check Prometheus config:**
```bash
# Verify prometheus.yml syntax
docker exec sovereignmap-prometheus \
  promtool check config /etc/prometheus/prometheus.yml
```

**Check target status:**
```bash
# Visit http://localhost:9090/targets
# All targets should show "UP"
```

**Restart Prometheus:**
```bash
docker restart sovereignmap-prometheus
```

---

### Issue: Grafana datasource connection failed

**Check network connectivity:**
```bash
# From Grafana container
docker exec sovereignmap-grafana curl http://sovereignmap-prometheus:9090

# Should return Prometheus UI
```

**Recreate datasource:**
```bash
# Delete and recreate in Grafana UI:
# Configuration > Data Sources > Prometheus
# URL: http://sovereignmap-prometheus:9090
# Save & Test
```

---

### Issue: WSL2 TPM passthrough not working

**Verify WSL2 configuration:**
```powershell
# Check .wslconfig
Get-Content $env:USERPROFILE\.wslconfig

# Should include: tpm=true

# Shutdown and restart WSL
wsl --shutdown
wsl --launch

# Verify TPM in WSL
tpm2_getcap handles-persistent
```

---

## Quick Setup Script

**Complete automated setup (Linux/macOS):**

```bash
#!/bin/bash
set -e

echo "🚀 Sovereign Map Environment Setup"

# Docker
echo "📦 Setting up Docker..."
docker network create sovereignmap 2>/dev/null || true

# Directories
mkdir -p monitoring/{prometheus,grafana,alertmanager}
mkdir -p test-results
mkdir -p test-data

# Prometheus
echo "📊 Starting Prometheus..."
docker run -d --name sovereignmap-prometheus --network sovereignmap -p 9090:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml:ro \
  prom/prometheus:v2.48.0

# Grafana
echo "📈 Starting Grafana..."
docker run -d --name sovereignmap-grafana --network sovereignmap -p 3001:3000 \
  -e GF_SECURITY_ADMIN_PASSWORD="${GRAFANA_ADMIN_PASSWORD:-CHANGE_ME_GRAFANA}" \
  grafana/grafana:10.2-alpine

# Node Exporter
echo "📡 Starting Node Exporter..."
docker run -d --name sovereignmap-node-exporter --network sovereignmap -p 9100:9100 \
  prom/node-exporter:v1.6.1

# Verification
echo ""
echo "✅ Setup Complete!"
echo "📊 Prometheus: http://localhost:9090"
echo "📈 Grafana: http://localhost:3001 (admin/<configured password>)"
echo "📡 Node Exporter: http://localhost:9100"
```

---

## Production Deployment Checklist

- [ ] Docker Engine 25.0+ installed
- [ ] WSL2 enabled (Windows)
- [ ] TPM 2.0 enabled in BIOS
- [ ] Docker network `sovereignmap` created
- [ ] Prometheus container running
- [ ] Grafana container running
- [ ] Node Exporter container running
- [ ] Alertmanager container running
- [ ] TPM devices accessible (/dev/tpm0, /dev/tpmrm0)
- [ ] Environment variables configured (.env)
- [ ] All health checks passing
- [ ] Grafana datasource connected
- [ ] Dashboards loaded
- [ ] Alerts configured

---

## Next Steps

1. **Complete setup** using the Quick Setup Script
2. **Verify all components** using Health Check scripts
3. **Configure dashboards** in Grafana
4. **Start the application** with docker-compose
5. **Run tests** using scripts in `/scripts` directory

---

**Your system (AMD Ryzen AI 7 350, 32GB RAM, 761GB storage) EXCEEDS all requirements.**
**Ready for production federated learning deployment!**
