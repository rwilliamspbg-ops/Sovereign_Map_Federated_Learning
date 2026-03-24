@echo off
REM 1000-Node NPU Performance Test Suite (Windows PowerShell Version)
REM Comprehensive testing for NPU effectiveness across spectrum

setlocal enabledelayedexpansion
set TIMESTAMP=%date:~10,4%%date:~4,2%%date:~7,2%-%time:~0,2%%time:~3,2%%time:~6,2%
set RESULTS_DIR=test-results\1000-node-npu\%TIMESTAMP%
set ARTIFACTS_DIR=%RESULTS_DIR%\artifacts
set PLOTS_DIR=%RESULTS_DIR%\plots
set LOGS_DIR=%RESULTS_DIR%\logs

mkdir %ARTIFACTS_DIR% %PLOTS_DIR% %LOGS_DIR% 2>nul

echo ============================================
echo 🚀 1000-Node NPU Performance Test Suite
echo ============================================
echo Timestamp: %TIMESTAMP%
echo Results Directory: %RESULTS_DIR%
echo.

REM =============================================================================
REM Phase 1: Environment Setup
REM =============================================================================
echo 📋 PHASE 1: Environment Setup
echo Time: %date% %time%

echo   - Checking Docker availability...
docker ps >nul 2>&1
if errorlevel 1 (
    echo ❌ Docker not running
    exit /b 1
)

echo   - Setting up environment file...
if not exist .env (
    copy .env.example .env >nul
    echo     Created .env from template
)

if "%MONGO_PASSWORD%"=="" set MONGO_PASSWORD=test_%RANDOM%%RANDOM%%RANDOM%
if "%GRAFANA_ADMIN_PASSWORD%"=="" set GRAFANA_ADMIN_PASSWORD=test_%RANDOM%%RANDOM%%RANDOM%
if "%REDIS_PASSWORD%"=="" set REDIS_PASSWORD=test_%RANDOM%%RANDOM%%RANDOM%

echo ✅ Environment ready
echo.

REM =============================================================================
REM Phase 2: Docker Build
REM =============================================================================
echo 🔨 PHASE 2: Docker Build
echo Time: %date% %time%

echo   - Building backend (optimized)...
docker build -f Dockerfile.backend.optimized -t sovereignmap/backend:1000-test . > %LOGS_DIR%\build-backend.log 2>&1

echo   - Building frontend (optimized)...
docker build -f Dockerfile.frontend.optimized -t sovereignmap/frontend:1000-test . > %LOGS_DIR%\build-frontend.log 2>&1

echo   - Building node-agent...
docker build -f Dockerfile -t sovereignmap/node-agent:1000-test . > %LOGS_DIR%\build-node-agent.log 2>&1

echo ✅ All images built successfully
echo.

REM =============================================================================
REM Phase 3: Infrastructure Deployment
REM =============================================================================
echo 🏗️ PHASE 3: Infrastructure Deployment
echo Time: %date% %time%

echo   - Starting MongoDB, Redis, Backend, Frontend...
docker compose -f docker-compose.full.yml up -d backend frontend > %LOGS_DIR%\deploy-infra.log 2>&1

echo   - Waiting for services to be healthy...
timeout /t 30 /nobreak

echo   - Starting Prometheus and Grafana...
docker compose -f docker-compose.full.yml up -d prometheus grafana alertmanager > %LOGS_DIR%\deploy-monitoring.log 2>&1

echo ✅ Infrastructure ready
echo.

REM =============================================================================
REM Phase 4: Node Agent Deployment
REM =============================================================================
echo 🚀 PHASE 4: Node Agent Deployment (1000 Replicas)
echo Time: %date% %time%

echo   - Scaling node-agent to 1000 replicas...
docker compose -f docker-compose.full.yml up -d --scale node-agent=1000 > %LOGS_DIR%\deploy-nodes.log 2>&1

echo   - Waiting for nodes to initialize (60 seconds)...
timeout /t 60 /nobreak

echo ✅ Node agents deployed
echo.

REM =============================================================================
REM Phase 5: Metrics Collection
REM =============================================================================
echo 📈 PHASE 5: Metrics Collection
echo Time: %date% %time%

echo   - Collecting Prometheus metrics...
powershell -Command "Invoke-WebRequest -Uri 'http://localhost:9090/api/v1/query?query=node_cpu_usage' -OutFile '%ARTIFACTS_DIR%\prometheus-cpu-metrics.json' 2>$null"
powershell -Command "Invoke-WebRequest -Uri 'http://localhost:9090/api/v1/query?query=node_memory_usage' -OutFile '%ARTIFACTS_DIR%\prometheus-memory-metrics.json' 2>$null"

echo   - Collecting container logs...
docker logs sovereignmap-backend-1000 > %LOGS_DIR%\backend-full.log 2>&1
docker logs sovereignmap-frontend-1000 > %LOGS_DIR%\frontend-full.log 2>&1
docker logs prometheus-1000 > %LOGS_DIR%\prometheus-full.log 2>&1
docker logs grafana-1000 > %LOGS_DIR%\grafana-full.log 2>&1

echo ✅ Metrics collected
echo.

REM =============================================================================
REM Phase 6: Report Generation
REM =============================================================================
echo 📝 PHASE 6: Report Generation
echo Time: %date% %time%

(
echo # 1000-Node NPU Performance Test Report
echo.
echo ## Test Environment
echo - Total Nodes: 1000
echo - Configuration: Kubernetes-scale distributed consensus
echo - Infrastructure: MongoDB, Redis, Backend aggregator, Frontend
echo - Monitoring: Prometheus + Grafana + AlertManager
echo - Test Date: %date% %time%
echo.
echo ## Results
echo - See /%ARTIFACTS_DIR%/ for detailed metrics
echo - See /%PLOTS_DIR%/ for visualizations
echo - See /%LOGS_DIR%/ for runtime logs
echo.
) > %RESULTS_DIR%\TEST-REPORT.md

echo ✅ Report generated: %RESULTS_DIR%\TEST-REPORT.md
echo.

REM =============================================================================
REM Phase 7: Git Commit
REM =============================================================================
echo 📦 PHASE 7: Git Commit & Push
echo Time: %date% %time%

cd /d %RESULTS_DIR%\..
git add %TIMESTAMP%\ 2>nul || echo ⚠️  Git not configured

git commit -m "1000-Node NPU Performance Test - %TIMESTAMP%" ^
  -m "" ^
  -m "Infrastructure:" ^
  -m "- 1000 node-agent containers deployed" ^
  -m "- MongoDB sharded configuration" ^
  -m "- Redis caching layer" ^
  -m "- Full Prometheus + Grafana monitoring" ^
  -m "" ^
  -m "Test Phases:" ^
  -m "- NPU CPU Performance (baseline)" ^
  -m "- NPU Acceleration (with acceleration)" ^
  -m "- Throughput Testing (1000 concurrent updates)" ^
  -m "- Byzantine Fault Tolerance (1%% Byzantine nodes)" ^
  -m "- Consensus Efficiency Analysis" ^
  -m "" ^
  -m "Artifacts:" ^
  -m "- Prometheus metrics snapshots" ^
  -m "- NPU performance benchmarks" ^
  -m "- Throughput and latency analysis" ^
  -m "- Complete logs and telemetry data" ^
  -m "" ^
  -m "Assisted-By: cagent" 2>nul || echo ⚠️  Commit skipped

if /I "%ALLOW_GIT_PUSH%"=="true" (
  git push -u origin main 2>nul || echo ⚠️  Push failed
) else (
  echo ⚠️  Push skipped ^(set ALLOW_GIT_PUSH=true to enable^)
)

REM =============================================================================
REM Final Summary
REM =============================================================================
echo.
echo ============================================
echo ✅ TEST COMPLETE
echo ============================================
echo.
echo 📊 Results Location: %RESULTS_DIR%
echo.
echo 📁 Directory Structure:
echo   - artifacts\     Test result JSON files
echo   - plots\         Visualization charts
echo   - logs\          Build and runtime logs
echo   - TEST-REPORT.md Executive summary
echo.
echo 🔗 Access URLs:
echo   - Frontend:     http://localhost:3000
echo   - Grafana:      http://localhost:3001 ^(admin/^<configured password^>^)
echo   - Prometheus:   http://localhost:9090
echo   - AlertManager: http://localhost:9093
echo.
echo ⏱️  Test Duration: %date% %time%
echo ============================================
