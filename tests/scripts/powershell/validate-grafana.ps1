#!/usr/bin/env pwsh
<#
.SYNOPSIS
Validates Grafana dashboard setup and configuration.

.DESCRIPTION
Checks that all 6 dashboards exist, provisioning directories are configured,
and docker-compose.full.yml has proper Grafana mounts.
#>

param(
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"
$projectPath = "Sovereign_Map_Federated_Learning"

Write-Host "=== Grafana Dashboard Validation ===" -ForegroundColor Cyan

# Check dashboard files
$dashboards = @(
    "sovereign-map-overview.json",
    "sovereign-map-convergence.json",
    "sovereign-map-performance.json",
    "sovereign-map-scaling.json",
    "sovereign-map-tpm-security.json",
    "sovereign-map-npu-acceleration.json"
)

$dashboardPath = "$projectPath/grafana/provisioning/dashboards"
Write-Host "`n[1/4] Checking dashboard files..." -ForegroundColor Yellow

$allFound = $true
foreach ($dashboard in $dashboards) {
    $fullPath = Join-Path $dashboardPath $dashboard
    if (Test-Path $fullPath) {
        $size = (Get-Item $fullPath).Length / 1KB
        Write-Host "  ✅ $dashboard ($([Math]::Round($size, 1)) KB)" -ForegroundColor Green
    } else {
        Write-Host "  ❌ $dashboard (NOT FOUND)" -ForegroundColor Red
        $allFound = $false
    }
}

if (-not $allFound) {
    Write-Host "  FAILED: Some dashboards missing" -ForegroundColor Red
    exit 1
}

# Check provisioning config files
Write-Host "`n[2/4] Checking provisioning config files..." -ForegroundColor Yellow

$configFiles = @(
    "grafana/provisioning/datasources/prometheus.yml",
    "grafana/provisioning/dashboards/dashboard-provider.yaml",
    "grafana/provisioning/dashboards/dashboards.yml"
)

$allFound = $true
foreach ($file in $configFiles) {
    $fullPath = "$projectPath/$file"
    if (Test-Path $fullPath) {
        Write-Host "  ✅ $file" -ForegroundColor Green
    } else {
        Write-Host "  ❌ $file (NOT FOUND)" -ForegroundColor Red
        $allFound = $false
    }
}

if (-not $allFound) {
    Write-Host "  FAILED: Some config files missing" -ForegroundColor Red
    exit 1
}

# Check docker-compose.full.yml mounts
Write-Host "`n[3/4] Checking docker-compose.full.yml Grafana config..." -ForegroundColor Yellow

$dockerComposePath = "$projectPath/docker-compose.full.yml"
$content = Get-Content $dockerComposePath -Raw

$checks = @(
    @{name = "Grafana datasources mount"; pattern = "./grafana/provisioning/datasources:/etc/grafana/provisioning/datasources" },
    @{name = "Grafana dashboards mount"; pattern = "./grafana/provisioning/dashboards:/etc/grafana/provisioning/dashboards" },
    @{name = "GF_PATHS_PROVISIONING env"; pattern = "GF_PATHS_PROVISIONING=/etc/grafana/provisioning" },
    @{name = "Grafana port 3001:3000"; pattern = '- "3001:3000"' }
)

$allFound = $true
foreach ($check in $checks) {
    if ($content -match [regex]::Escape($check.pattern)) {
        Write-Host "  ✅ $($check.name)" -ForegroundColor Green
    } else {
        Write-Host "  ❌ $($check.name) (NOT FOUND)" -ForegroundColor Red
        $allFound = $false
    }
}

if (-not $allFound) {
    Write-Host "  FAILED: Docker Compose config incomplete" -ForegroundColor Red
    exit 1
}

# Count metrics
Write-Host "`n[4/4] Analyzing dashboard queries..." -ForegroundColor Yellow

$metricPatterns = @(
    "sovereignmap_active_nodes",
    "sovereignmap_model_accuracy",
    "sovereignmap_training_loss",
    "sovereignmap_http_request_duration_seconds",
    "sovereignmap_tpm_verified_nodes",
    "sovereignmap_npu_speedup_factor"
)

$uniqueMetrics = @{}
foreach ($dashboard in $dashboards) {
    $fullPath = Join-Path $dashboardPath $dashboard
    $dashContent = Get-Content $fullPath -Raw
    
    foreach ($metric in $metricPatterns) {
        if ($dashContent -match $metric) {
            $uniqueMetrics[$metric] = $true
        }
    }
}

Write-Host "  Found $($uniqueMetrics.Count) unique metrics across dashboards" -ForegroundColor Green

# Summary
Write-Host "`n" + ("="*50) -ForegroundColor Cyan
Write-Host "✅ GRAFANA SETUP VALIDATION PASSED" -ForegroundColor Green
Write-Host ("="*50) -ForegroundColor Cyan

Write-Host "`nSummary:" -ForegroundColor Yellow
Write-Host "  • 6 dashboards: $(($dashboards | Measure-Object).Count)"
Write-Host "  • Config files: $(($configFiles | Measure-Object).Count)"
Write-Host "  • Provisioning mounts: 2 (datasources + dashboards)"
Write-Host "  • Unique metrics: $($uniqueMetrics.Count)"

Write-Host "`nNext Steps:" -ForegroundColor Yellow
Write-Host "  1. docker compose -f docker-compose.full.yml up -d --scale node-agent=5" -ForegroundColor Cyan
Write-Host "  2. Wait 30 seconds for services to start"
Write-Host "  3. Open http://localhost:3001 (admin/sovereignmap)"
Write-Host "  4. View dashboards in 'Sovereign' folder"
Write-Host "  5. Run test: ./run-5000-round-test.ps1"

Write-Host "`nMonitoring URLs:" -ForegroundColor Yellow
Write-Host "  • Grafana:     http://localhost:3001" -ForegroundColor Cyan
Write-Host "  • Prometheus: http://localhost:9090" -ForegroundColor Cyan
Write-Host "  • Backend:    http://localhost:8000/metrics" -ForegroundColor Cyan

Write-Host ""
